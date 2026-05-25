package handler

import (
	"net/http"

	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/app/internal/user/repository"
)

// DeleteStudentHandler handles DELETE /students/me.
// Anonymizes the student row in place (GDPR: preserves referential integrity).
type DeleteStudentHandler struct {
	studentRepo *repository.StudentRepository
}

// NewDeleteStudentHandler constructs a DeleteStudentHandler.
func NewDeleteStudentHandler(studentRepo *repository.StudentRepository) *DeleteStudentHandler {
	return &DeleteStudentHandler{studentRepo: studentRepo}
}

func (h *DeleteStudentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := pkgmw.UserIDFromContext(r.Context())
	if err := h.studentRepo.Anonymize(r.Context(), userID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
