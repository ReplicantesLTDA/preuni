package handler

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	apperrors "github.com/preuni/pkg/errors"
	pkgmw "github.com/preuni/pkg/middleware"
)

// GetStreakHandler handles GET /v1/streaks/me.
type GetStreakHandler struct {
	db *pgxpool.Pool
}

// NewGetStreakHandler constructs a GetStreakHandler.
func NewGetStreakHandler(db *pgxpool.Pool) *GetStreakHandler {
	return &GetStreakHandler{db: db}
}

type streakResponse struct {
	CurrentStreak int     `json:"current_streak"`
	LongestStreak int     `json:"longest_streak"`
	LastActiveDay *string `json:"last_active_day"`
}

func (h *GetStreakHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := pkgmw.UserIDFromContext(r.Context())

	row := h.db.QueryRow(r.Context(), `
		SELECT streak_count, longest_streak, streak_last_active_date
		FROM users.students
		WHERE id = $1
	`, userID)

	var current, longest int
	var lastActive *time.Time
	if err := row.Scan(&current, &longest, &lastActive); err != nil {
		pkgmw.ErrorResponse(w, apperrors.NotFound("student"))
		return
	}

	var lastActiveStr *string
	if lastActive != nil {
		s := lastActive.Format("2006-01-02")
		lastActiveStr = &s
	}

	pkgmw.JSON(w, http.StatusOK, streakResponse{
		CurrentStreak: current,
		LongestStreak: longest,
		LastActiveDay: lastActiveStr,
	})
}
