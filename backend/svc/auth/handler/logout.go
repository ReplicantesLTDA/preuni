package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	apperrors "github.com/preuni/pkg/errors"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/svc/auth/repository"
)

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// LogoutRequest is the body for POST /auth/logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutHandler handles POST /auth/logout.
// Revokes the specific refresh token so the session cannot be renewed.
type LogoutHandler struct {
	refreshRepo *repository.RefreshTokenRepository
}

// NewLogoutHandler constructs a LogoutHandler.
func NewLogoutHandler(refreshRepo *repository.RefreshTokenRepository) *LogoutHandler {
	return &LogoutHandler{refreshRepo: refreshRepo}
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("refresh_token", "refresh_token is required"))
		return
	}

	hashHex := hashToken(req.RefreshToken)
	token, err := h.refreshRepo.FindByHash(r.Context(), hashHex)
	if err != nil {
		// Treat unknown token as already revoked — still 204
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err = h.refreshRepo.Revoke(r.Context(), token.ID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
