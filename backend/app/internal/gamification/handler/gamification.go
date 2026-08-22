package handler

import (
	"net/http"
	"time"

	pkgmw "github.com/preuni/pkg/middleware"

	"github.com/preuni/app/internal/gamification/repository"
)

// Handler bundles the ranking + medals HTTP endpoints (contracts/public-api.md).
type Handler struct {
	repo *repository.Repository
}

// NewHandler constructs a Handler.
func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

type rankingEntryResponse struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	WeeklyScore int    `json:"weekly_score"`
	LeagueTier  string `json:"league_tier"`
	RankInTier  *int   `json:"rank_in_tier,omitempty"`
}

// WeeklyLeaderboard handles GET /v1/ranking/weekly?tier=bronze.
func (h *Handler) WeeklyLeaderboard(w http.ResponseWriter, r *http.Request) {
	tier := r.URL.Query().Get("tier")
	if tier == "" {
		tier = "bronze"
	}

	entries, err := h.repo.WeeklyLeaderboard(r.Context(), tier, time.Now())
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	resp := make([]rankingEntryResponse, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, rankingEntryResponse{
			UserID:      e.UserID,
			DisplayName: e.DisplayName,
			WeeklyScore: e.WeeklyScore,
			LeagueTier:  e.LeagueTier,
			RankInTier:  e.RankInTier,
		})
	}
	pkgmw.JSON(w, http.StatusOK, resp)
}

// MyRanking handles GET /v1/ranking/me.
func (h *Handler) MyRanking(w http.ResponseWriter, r *http.Request) {
	userID := pkgmw.UserIDFromContext(r.Context())

	entry, err := h.repo.MyRanking(r.Context(), userID, time.Now())
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	pkgmw.JSON(w, http.StatusOK, rankingEntryResponse{
		UserID:      entry.UserID,
		DisplayName: entry.DisplayName,
		WeeklyScore: entry.WeeklyScore,
		LeagueTier:  entry.LeagueTier,
		RankInTier:  entry.RankInTier,
	})
}

type medalResponse struct {
	Type     string `json:"type"`
	EarnedAt string `json:"earned_at"`
}

// MyMedals handles GET /v1/medals/me.
func (h *Handler) MyMedals(w http.ResponseWriter, r *http.Request) {
	userID := pkgmw.UserIDFromContext(r.Context())

	medals, err := h.repo.ListMedals(r.Context(), userID)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	resp := make([]medalResponse, 0, len(medals))
	for _, m := range medals {
		resp = append(resp, medalResponse{Type: m.Type, EarnedAt: m.EarnedAt.Format(time.RFC3339)})
	}
	pkgmw.JSON(w, http.StatusOK, resp)
}
