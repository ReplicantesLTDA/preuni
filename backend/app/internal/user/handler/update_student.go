package handler

import (
	"encoding/json"
	"net/http"

	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/app/internal/user/domain"
	"github.com/preuni/app/internal/user/repository"
)

// UpdateStudentRequest is the JSON body for PATCH /students/me.
type UpdateStudentRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Username    *string `json:"username,omitempty"`
}

// StudentResponse is returned from GET /students/me and PATCH /students/me.
type StudentResponse struct {
	ID                  string  `json:"id"`
	DisplayName         string  `json:"display_name"`
	Username            string  `json:"username"`
	Email               string  `json:"email"`
	AvatarURL           *string `json:"avatar_url"`
	XPTotal             int64   `json:"xp_total"`
	StreakCount         int     `json:"streak_count"`
	ReadinessScore      float64 `json:"readiness_score"`
	OnboardingCompleted bool    `json:"onboarding_completed"`
}

func toStudentResponse(s *repository.Student) StudentResponse {
	return StudentResponse{
		ID:                  s.ID,
		DisplayName:         s.DisplayName,
		Username:            s.Username,
		Email:               s.Email,
		AvatarURL:           s.AvatarURL,
		XPTotal:             s.XPTotal,
		StreakCount:         s.StreakCount,
		ReadinessScore:      s.ReadinessScore,
		OnboardingCompleted: s.OnboardingCompleted,
	}
}

// UpdateStudentHandler handles PATCH /students/me.
type UpdateStudentHandler struct {
	studentRepo *repository.StudentRepository
}

// NewUpdateStudentHandler constructs an UpdateStudentHandler.
func NewUpdateStudentHandler(studentRepo *repository.StudentRepository) *UpdateStudentHandler {
	return &UpdateStudentHandler{studentRepo: studentRepo}
}

func (h *UpdateStudentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req UpdateStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	if req.Username != nil {
		if err := domain.ValidateUsername(*req.Username); err != nil {
			pkgmw.ErrorResponse(w, err)
			return
		}
	}

	userID := pkgmw.UserIDFromContext(r.Context())
	patch := &repository.StudentPatch{
		DisplayName: req.DisplayName,
		Username:    req.Username,
	}

	updated, err := h.studentRepo.Update(r.Context(), userID, patch)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	pkgmw.JSON(w, http.StatusOK, toStudentResponse(updated))
}
