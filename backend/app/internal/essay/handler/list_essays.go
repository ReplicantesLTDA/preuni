package handler

import (
	"net/http"

	pkgmw "github.com/preuni/pkg/middleware"

	"github.com/preuni/app/internal/essay/repository"
)

// ListEssaysHandler handles GET /v1/essays.
type ListEssaysHandler struct {
	repo *repository.Repository
}

// NewListEssaysHandler constructs a ListEssaysHandler.
func NewListEssaysHandler(repo *repository.Repository) *ListEssaysHandler {
	return &ListEssaysHandler{repo: repo}
}

type essaySummary struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	SubmittedAt string `json:"submitted_at"`
}

func (h *ListEssaysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := pkgmw.UserIDFromContext(r.Context())

	views, err := h.repo.ListByUser(r.Context(), userID)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	summaries := make([]essaySummary, 0, len(views))
	for _, v := range views {
		summaries = append(summaries, essaySummary{
			ID:          v.ID,
			Status:      v.Status,
			SubmittedAt: v.SubmittedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	pkgmw.JSON(w, http.StatusOK, summaries)
}
