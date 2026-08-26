package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	apperrors "github.com/preuni/pkg/errors"
	"github.com/preuni/pkg/logger"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/app/internal/auth/domain"
	"github.com/preuni/app/internal/auth/ports"
	"github.com/preuni/app/internal/auth/repository"
)

// GoogleLoginRequest is the JSON body for POST /auth/google.
type GoogleLoginRequest struct {
	IDToken string `json:"id_token"`
}

// GoogleLoginHandler handles POST /auth/google: sign-in or sign-up via a
// Google-issued OpenID Connect ID token (obtained client-side, e.g. via
// expo-auth-session -- the backend never sees a Google client secret).
type GoogleLoginHandler struct {
	verifier           domain.GoogleIDTokenVerifier
	clientID           string
	credRepo           *repository.CredentialsRepository
	oauthRepo          *repository.OAuthIdentityRepository
	refreshRepo        *repository.RefreshTokenRepository
	jwtSvc             *domain.JWTService
	studentProvisioner ports.StudentProvisioner
	log                *logger.Logger
}

// NewGoogleLoginHandler constructs a GoogleLoginHandler. If clientID is
// empty, Google sign-in is not configured and every request is rejected --
// this lets the feature ship dormant until real OAuth credentials exist.
func NewGoogleLoginHandler(
	verifier domain.GoogleIDTokenVerifier,
	clientID string,
	credRepo *repository.CredentialsRepository,
	oauthRepo *repository.OAuthIdentityRepository,
	refreshRepo *repository.RefreshTokenRepository,
	jwtSvc *domain.JWTService,
	studentProvisioner ports.StudentProvisioner,
	log *logger.Logger,
) *GoogleLoginHandler {
	return &GoogleLoginHandler{
		verifier:           verifier,
		clientID:           clientID,
		credRepo:           credRepo,
		oauthRepo:          oauthRepo,
		refreshRepo:        refreshRepo,
		jwtSvc:              jwtSvc,
		studentProvisioner: studentProvisioner,
		log:                log,
	}
}

func (h *GoogleLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.clientID == "" {
		pkgmw.ErrorResponse(w, apperrors.Internal(nil))
		return
	}

	var req GoogleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgmw.ErrorResponse(w, apperrors.Validation("body", "invalid JSON"))
		return
	}
	if req.IDToken == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("id_token", "id_token is required"))
		return
	}

	claims, err := h.verifier.Verify(r.Context(), req.IDToken, h.clientID)
	if err != nil {
		pkgmw.ErrorResponse(w, apperrors.Unauthorized("invalid Google ID token"))
		return
	}
	if !claims.EmailVerified {
		pkgmw.ErrorResponse(w, apperrors.Forbidden("Google account email is not verified"))
		return
	}

	credentialID, err := h.oauthRepo.FindCredentialID(r.Context(), "google", claims.Subject)
	switch {
	case apperrors.IsNotFound(err):
		credentialID, err = h.signUp(r.Context(), claims)
		if err != nil {
			pkgmw.ErrorResponse(w, err)
			return
		}
	case err != nil:
		pkgmw.ErrorResponse(w, err)
		return
	}

	creds, err := h.credRepo.FindByID(r.Context(), credentialID)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	accessToken, err := h.jwtSvc.IssueAccessToken(creds.ID, creds.Email)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	rawRefresh, refreshHash, refreshExpiry, err := h.jwtSvc.IssueRefreshToken()
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	if err = h.refreshRepo.Store(r.Context(), creds.ID, refreshHash, refreshExpiry); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	pkgmw.JSON(w, http.StatusOK, AuthResponse{
		StudentID:    creds.ID,
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    h.jwtSvc.ExpiresIn(),
	})
}

// signUp links an existing credential found by email to this Google
// identity, or creates a brand new OAuth-only credential + student profile
// if the email has never been seen before. Returns the linked credential ID.
func (h *GoogleLoginHandler) signUp(ctx context.Context, claims domain.GoogleClaims) (string, error) {
	existing, err := h.credRepo.FindByEmail(ctx, claims.Email)
	if err == nil {
		if linkErr := h.oauthRepo.Link(ctx, existing.ID, "google", claims.Subject); linkErr != nil {
			return "", linkErr
		}
		return existing.ID, nil
	}
	if !apperrors.IsNotFound(err) {
		return "", err
	}

	creds, err := domain.NewOAuthCredentials(claims.Email)
	if err != nil {
		return "", err
	}
	studentID := uuid.New().String()
	creds.ID = studentID

	if err := h.credRepo.Create(ctx, creds); err != nil {
		return "", err
	}
	if err := h.oauthRepo.Link(ctx, studentID, "google", claims.Subject); err != nil {
		return "", err
	}

	if err := h.studentProvisioner.CreateStudent(ctx, ports.CreateStudentRequest{
		StudentID:   studentID,
		DisplayName: claims.Email,
		Email:       claims.Email,
	}); err != nil {
		h.log.Error("failed to create student profile after google signup",
			logger.String("student_id", studentID), logger.Err(err))
	}

	return studentID, nil
}
