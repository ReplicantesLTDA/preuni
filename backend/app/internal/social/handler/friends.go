package handler

import (
	"encoding/json"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	apperrors "github.com/preuni/pkg/errors"
	pkgmw "github.com/preuni/pkg/middleware"

	"github.com/preuni/app/internal/social/repository"
)

// Handler bundles the friends HTTP endpoints (contracts/public-api.md).
type Handler struct {
	repo *repository.Repository
}

// NewHandler constructs a Handler.
func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

type sendRequestBody struct {
	AddresseeID string `json:"addressee_id"`
}

// SendRequest handles POST /v1/friends/requests.
func (h *Handler) SendRequest(w http.ResponseWriter, r *http.Request) {
	userID := pkgmw.UserIDFromContext(r.Context())

	var body sendRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AddresseeID == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("addressee_id", "addressee_id is required"))
		return
	}

	id, err := h.repo.SendRequest(r.Context(), userID, body.AddresseeID)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	pkgmw.JSON(w, http.StatusCreated, map[string]string{"id": id, "status": "pending"})
}

// AcceptRequest handles POST /v1/friends/requests/{id}/accept.
func (h *Handler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
	userID := pkgmw.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.repo.AcceptRequest(r.Context(), id, userID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	pkgmw.JSON(w, http.StatusOK, map[string]string{"id": id, "status": "accepted"})
}

// RemoveFriend handles DELETE /v1/friends/{id}.
func (h *Handler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	userID := pkgmw.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.repo.RemoveFriend(r.Context(), id, userID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type friendResponse struct {
	FriendshipID       string  `json:"friendship_id"`
	UserID             string  `json:"user_id"`
	DisplayName        string  `json:"display_name"`
	CurrentStreak      int     `json:"current_streak"`
	LatestOverallScore *int    `json:"latest_overall_score,omitempty"`
	LatestGradedAt     *string `json:"latest_graded_at,omitempty"`
}

// ListFriends handles GET /v1/friends.
func (h *Handler) ListFriends(w http.ResponseWriter, r *http.Request) {
	userID := pkgmw.UserIDFromContext(r.Context())

	friends, err := h.repo.ListFriends(r.Context(), userID)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	resp := make([]friendResponse, 0, len(friends))
	for _, f := range friends {
		resp = append(resp, friendResponse{
			FriendshipID:       f.FriendshipID,
			UserID:             f.UserID,
			DisplayName:        f.DisplayName,
			CurrentStreak:      f.CurrentStreak,
			LatestOverallScore: f.LatestOverallScore,
			LatestGradedAt:     f.LatestGradedAt,
		})
	}
	pkgmw.JSON(w, http.StatusOK, resp)
}
