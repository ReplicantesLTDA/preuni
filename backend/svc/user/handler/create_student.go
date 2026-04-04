package handler

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/preuni/pkg/errors"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/svc/user/repository"
)

// CreateStudentRequest is the JSON body for POST /internal/students.
// Called by auth-svc after a successful registration.
type CreateStudentRequest struct {
	StudentID   string `json:"student_id"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Email       string `json:"email"`
}

// CreateStudentHandler handles POST /internal/students.
// This endpoint is protected by InternalAuth middleware and is not exposed via NGINX.
type CreateStudentHandler struct {
	studentRepo *repository.StudentRepository
}

// NewCreateStudentHandler constructs a CreateStudentHandler.
func NewCreateStudentHandler(studentRepo *repository.StudentRepository) *CreateStudentHandler {
	return &CreateStudentHandler{studentRepo: studentRepo}
}

func (h *CreateStudentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req CreateStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgmw.ErrorResponse(w, apperrors.Validation("body", "invalid JSON"))
		return
	}
	if req.StudentID == "" || req.DisplayName == "" || req.Email == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("body", "student_id, display_name and email are required"))
		return
	}
	if req.Username == "" {
		req.Username = generateDefaultUsername(req.StudentID)
	}

	student := &repository.Student{
		ID:          req.StudentID,
		DisplayName: req.DisplayName,
		Username:    req.Username,
		Email:       req.Email,
	}

	if err := h.studentRepo.Create(r.Context(), student); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	pkgmw.JSON(w, http.StatusCreated, map[string]string{"id": req.StudentID})
}

// generateDefaultUsername creates a unique username from a UUID fragment when
// the user did not choose one during registration.
func generateDefaultUsername(id string) string {
	if len(id) >= 8 {
		return "user" + id[:8]
	}
	return "user" + id
}
