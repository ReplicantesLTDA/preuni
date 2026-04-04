package handler

import (
	"encoding/json"
	"net/http"

	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/svc/user/repository"
)

// OnboardingRequest is the optional body for PATCH /students/me/onboarding.
type OnboardingRequest struct {
	EnrolledTrackIDs []string `json:"enrolled_track_ids,omitempty"`
}

// OnboardingHandler handles PATCH /students/me/onboarding.
// Sets onboarding_completed = true; idempotent.
type OnboardingHandler struct {
	studentRepo *repository.StudentRepository
}

// NewOnboardingHandler constructs an OnboardingHandler.
func NewOnboardingHandler(studentRepo *repository.StudentRepository) *OnboardingHandler {
	return &OnboardingHandler{studentRepo: studentRepo}
}

func (h *OnboardingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req OnboardingRequest
	json.NewDecoder(r.Body).Decode(&req) //nolint:errcheck — body is optional

	userID := pkgmw.UserIDFromContext(r.Context())

	if err := h.studentRepo.SetOnboardingCompleted(r.Context(), userID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	// TODO: create track_enrollments rows for req.EnrolledTrackIDs once content-svc is wired

	student, err := h.studentRepo.FindByID(r.Context(), userID)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	pkgmw.JSON(w, http.StatusOK, toStudentResponse(student))
}
