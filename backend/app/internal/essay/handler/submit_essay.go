package handler

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/preuni/pkg/errors"
	pkgmw "github.com/preuni/pkg/middleware"

	"github.com/preuni/app/internal/essay/repository"
)

// SubmitEssayHandler handles POST /v1/essays.
type SubmitEssayHandler struct {
	repo *repository.Repository
}

// NewSubmitEssayHandler constructs a SubmitEssayHandler.
func NewSubmitEssayHandler(repo *repository.Repository) *SubmitEssayHandler {
	return &SubmitEssayHandler{repo: repo}
}

type submitEssayRequest struct {
	PromptThemeTitle   string `json:"prompt_theme_title"`
	PromptThemeContext string `json:"prompt_theme_context"`
	EssayText          string `json:"essay_text"`
}

type submitEssayResponse struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	SubmittedAt string `json:"submitted_at"`
}

func (h *SubmitEssayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := pkgmw.UserIDFromContext(r.Context())

	var req submitEssayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgmw.ErrorResponse(w, apperrors.Validation("body", "invalid JSON body"))
		return
	}
	if req.PromptThemeTitle == "" || req.PromptThemeContext == "" || req.EssayText == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("body", "prompt_theme_title, prompt_theme_context, and essay_text are required"))
		return
	}

	submission, err := h.repo.Submit(r.Context(), userID, req.PromptThemeTitle, req.PromptThemeContext, req.EssayText)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	pkgmw.JSON(w, http.StatusAccepted, submitEssayResponse{
		ID:          submission.ID,
		Status:      submission.Status,
		SubmittedAt: submission.SubmittedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
