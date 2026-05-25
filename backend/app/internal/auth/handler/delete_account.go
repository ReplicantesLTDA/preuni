package handler

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/preuni/pkg/errors"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/app/internal/auth/repository"
)

// DeleteAccountRequest is the optional body for DELETE /auth/account.
type DeleteAccountRequest struct {
	Confirmation string `json:"confirmation"` // must equal "DELETE"
}

// DeleteAccountHandler handles DELETE /auth/account.
// GDPR: Anonymizes credentials and revokes all sessions. The student row in
// users.students is anonymized via the user domain's own DELETE /v1/students/me
// path (clients call that explicitly during the delete flow).
type DeleteAccountHandler struct {
	credRepo    *repository.CredentialsRepository
	refreshRepo *repository.RefreshTokenRepository
}

// NewDeleteAccountHandler constructs a DeleteAccountHandler.
func NewDeleteAccountHandler(credRepo *repository.CredentialsRepository, refreshRepo *repository.RefreshTokenRepository) *DeleteAccountHandler {
	return &DeleteAccountHandler{credRepo: credRepo, refreshRepo: refreshRepo}
}

func (h *DeleteAccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req DeleteAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
		if req.Confirmation != "DELETE" {
			pkgmw.ErrorResponse(w, apperrors.Validation("confirmation", `type "DELETE" to confirm`))
			return
		}
	}

	userID := pkgmw.UserIDFromContext(r.Context())

	if err := h.refreshRepo.RevokeAllForCredential(r.Context(), userID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	if err := h.credRepo.Anonymize(r.Context(), userID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
