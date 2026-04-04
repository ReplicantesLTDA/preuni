package handler

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/preuni/pkg/errors"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/svc/auth/repository"
)

// DeleteAccountRequest is the optional body for DELETE /auth/account.
type DeleteAccountRequest struct {
	Confirmation string `json:"confirmation"` // must equal "DELETE"
}

// DeleteAccountHandler handles DELETE /auth/account.
// GDPR: Anonymizes credentials and revokes all sessions.
// The user-svc student row is anonymized via its own DELETE /students/me endpoint.
type DeleteAccountHandler struct {
	credRepo    *repository.CredentialsRepository
	refreshRepo *repository.RefreshTokenRepository
	userSvcURL  string
}

// NewDeleteAccountHandler constructs a DeleteAccountHandler.
func NewDeleteAccountHandler(credRepo *repository.CredentialsRepository, refreshRepo *repository.RefreshTokenRepository, userSvcURL string) *DeleteAccountHandler {
	return &DeleteAccountHandler{credRepo: credRepo, refreshRepo: refreshRepo, userSvcURL: userSvcURL}
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

	// 1. Revoke all refresh tokens
	if err := h.refreshRepo.RevokeAllForCredential(r.Context(), userID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	// 2. Anonymize credentials (GDPR erasure)
	if err := h.credRepo.Anonymize(r.Context(), userID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	// 3. Call user-svc to anonymize the student row
	// (best-effort via internal endpoint; actual implementation will use http.Client)

	w.WriteHeader(http.StatusNoContent)
}
