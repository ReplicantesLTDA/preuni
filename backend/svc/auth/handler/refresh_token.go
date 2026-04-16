package handler

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/preuni/pkg/errors"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/svc/auth/domain"
	"github.com/preuni/svc/auth/repository"
)

// RefreshTokenRequest is the JSON body for POST /auth/refresh.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenHandler handles POST /auth/refresh.
// Validates the refresh token, rotates it, and issues a new access token.
type RefreshTokenHandler struct {
	credRepo    *repository.CredentialsRepository
	refreshRepo *repository.RefreshTokenRepository
	jwtSvc      *domain.JWTService
}

// NewRefreshTokenHandler constructs a RefreshTokenHandler.
func NewRefreshTokenHandler(
	credRepo *repository.CredentialsRepository,
	refreshRepo *repository.RefreshTokenRepository,
	jwtSvc *domain.JWTService,
) *RefreshTokenHandler {
	return &RefreshTokenHandler{credRepo: credRepo, refreshRepo: refreshRepo, jwtSvc: jwtSvc}
}

func (h *RefreshTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("refresh_token", "refresh_token is required"))
		return
	}

	hashHex := hashToken(req.RefreshToken)
	stored, err := h.refreshRepo.FindByHash(r.Context(), hashHex)
	if err != nil {
		pkgmw.ErrorResponse(w, apperrors.Unauthorized("refresh token is invalid or expired"))
		return
	}

	creds, err := h.credRepo.FindByID(r.Context(), stored.CredentialID)
	if err != nil {
		pkgmw.ErrorResponse(w, apperrors.Unauthorized("account not found"))
		return
	}

	newAccess, err := h.jwtSvc.IssueAccessToken(creds.ID, creds.Email)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	newRaw, newHash, newExpiry, err := h.jwtSvc.IssueRefreshToken()
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	if err = h.refreshRepo.Store(r.Context(), creds.ID, newHash, newExpiry); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	// Rotate: revoke the old token after the new one is stored.
	if err = h.refreshRepo.Revoke(r.Context(), stored.ID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	pkgmw.JSON(w, http.StatusOK, AuthResponse{
		StudentID:    creds.ID,
		AccessToken:  newAccess,
		RefreshToken: newRaw,
		ExpiresIn:    h.jwtSvc.ExpiresIn(),
	})
}
