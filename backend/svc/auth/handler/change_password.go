package handler

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/preuni/pkg/errors"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/svc/auth/domain"
	"github.com/preuni/svc/auth/repository"
)

// ChangePasswordRequest is the JSON body for POST /auth/password/change.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePasswordHandler handles POST /auth/password/change.
type ChangePasswordHandler struct {
	credRepo    *repository.CredentialsRepository
	refreshRepo *repository.RefreshTokenRepository
}

// NewChangePasswordHandler constructs a ChangePasswordHandler.
func NewChangePasswordHandler(credRepo *repository.CredentialsRepository, refreshRepo *repository.RefreshTokenRepository) *ChangePasswordHandler {
	return &ChangePasswordHandler{credRepo: credRepo, refreshRepo: refreshRepo}
}

func (h *ChangePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgmw.ErrorResponse(w, apperrors.Validation("body", "invalid JSON"))
		return
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("body", "current_password and new_password are required"))
		return
	}

	userID := pkgmw.UserIDFromContext(r.Context())
	creds, err := h.credRepo.FindByID(r.Context(), userID)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	if err = creds.CheckPassword(req.CurrentPassword); err != nil {
		pkgmw.ErrorResponse(w, apperrors.Validation("current_password", "Current password is incorrect"))
		return
	}

	if err = domain.ValidatePassword(req.NewPassword); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	newHash, err := domain.HashPassword(req.NewPassword)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	if err = h.credRepo.UpdatePasswordHash(r.Context(), userID, newHash); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	// Revoke all refresh tokens so the user must re-authenticate on other devices.
	if err = h.refreshRepo.RevokeAllForCredential(r.Context(), userID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
