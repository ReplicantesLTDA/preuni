package handler

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/preuni/pkg/errors"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/app/internal/auth/domain"
	"github.com/preuni/app/internal/auth/repository"
)

// LoginRequest is the JSON body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginHandler handles POST /auth/login.
type LoginHandler struct {
	credRepo    *repository.CredentialsRepository
	refreshRepo *repository.RefreshTokenRepository
	jwtSvc      *domain.JWTService
}

// NewLoginHandler constructs a LoginHandler.
func NewLoginHandler(
	credRepo *repository.CredentialsRepository,
	refreshRepo *repository.RefreshTokenRepository,
	jwtSvc *domain.JWTService,
) *LoginHandler {
	return &LoginHandler{credRepo: credRepo, refreshRepo: refreshRepo, jwtSvc: jwtSvc}
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgmw.ErrorResponse(w, apperrors.Validation("body", "invalid JSON"))
		return
	}
	if req.Email == "" || req.Password == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("body", "email and password are required"))
		return
	}

	creds, err := h.credRepo.FindByEmail(r.Context(), req.Email)
	if err != nil {
		// Return a generic error to prevent email enumeration.
		pkgmw.ErrorResponse(w, apperrors.Unauthorized("invalid credentials"))
		return
	}

	if !creds.EmailVerified {
		pkgmw.ErrorResponse(w, apperrors.Forbidden("email address is not yet verified"))
		return
	}

	if err = creds.CheckPassword(req.Password); err != nil {
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
