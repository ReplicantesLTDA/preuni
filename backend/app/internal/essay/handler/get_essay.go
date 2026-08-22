package handler

import (
	"encoding/json"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	pkgmw "github.com/preuni/pkg/middleware"

	"github.com/preuni/app/internal/essay/repository"
)

// GetEssayHandler handles GET /v1/essays/{id}.
type GetEssayHandler struct {
	repo *repository.Repository
}

// NewGetEssayHandler constructs a GetEssayHandler.
func NewGetEssayHandler(repo *repository.Repository) *GetEssayHandler {
	return &GetEssayHandler{repo: repo}
}

type competency struct {
	Competency        int    `json:"competency"`
	Score             int    `json:"score"`
	JustificationPtBr string `json:"justification_pt_br"`
	Excerpt           string `json:"excerpt"`
}

type gradeResponse struct {
	OverallScore int          `json:"overall_score"`
	Competencies []competency `json:"competencies"`
	GradedAt     string       `json:"graded_at"`
}

type essayResponse struct {
	ID          string         `json:"id"`
	Status      string         `json:"status"`
	SubmittedAt string         `json:"submitted_at"`
	Grade       *gradeResponse `json:"grade,omitempty"`
}

func (h *GetEssayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := pkgmw.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	view, err := h.repo.GetByID(r.Context(), id, userID)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	resp := essayResponse{
		ID:          view.ID,
		Status:      view.Status,
		SubmittedAt: view.SubmittedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if view.Grade != nil {
		var comps []competency
		if err := json.Unmarshal(view.Grade.Competencies, &comps); err == nil {
			resp.Grade = &gradeResponse{
				OverallScore: view.Grade.OverallScore,
				Competencies: comps,
				GradedAt:     view.Grade.GradedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		}
	}

	pkgmw.JSON(w, http.StatusOK, resp)
}
