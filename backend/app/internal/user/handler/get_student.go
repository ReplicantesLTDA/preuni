package handler

import (
	"net/http"

	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/app/internal/user/repository"
)

// GetStudentHandler handles GET /students/me.
type GetStudentHandler struct {
	studentRepo *repository.StudentRepository
}

// NewGetStudentHandler constructs a GetStudentHandler.
func NewGetStudentHandler(studentRepo *repository.StudentRepository) *GetStudentHandler {
	return &GetStudentHandler{studentRepo: studentRepo}
}

func (h *GetStudentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := pkgmw.UserIDFromContext(r.Context())
	student, err := h.studentRepo.FindByID(r.Context(), userID)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	pkgmw.JSON(w, http.StatusOK, toStudentResponse(student))
}
