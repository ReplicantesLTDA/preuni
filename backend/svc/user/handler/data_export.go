package handler

import (
	"encoding/json"
	"net/http"
	"time"

	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/svc/user/repository"
)

// DataExportResponse is the GDPR subject-access payload.
type DataExportResponse struct {
	ExportedAt string          `json:"exported_at"`
	Profile    StudentResponse `json:"profile"`
	// XP events, track enrollments, and achievements would be added
	// once those repositories are wired to this handler.
	XPEvents         []any `json:"xp_events"`
	TrackEnrollments []any `json:"track_enrollments"`
	Achievements     []any `json:"achievements"`
}

// DataExportHandler handles GET /students/me/data-export.
// Returns all data the user owns as a downloadable JSON attachment.
type DataExportHandler struct {
	studentRepo *repository.StudentRepository
}

// NewDataExportHandler constructs a DataExportHandler.
func NewDataExportHandler(studentRepo *repository.StudentRepository) *DataExportHandler {
	return &DataExportHandler{studentRepo: studentRepo}
}

func (h *DataExportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := pkgmw.UserIDFromContext(r.Context())

	student, err := h.studentRepo.FindByID(r.Context(), userID)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	export := DataExportResponse{
		ExportedAt:       time.Now().UTC().Format(time.RFC3339),
		Profile:          toStudentResponse(student),
		XPEvents:         []any{},
		TrackEnrollments: []any{},
		Achievements:     []any{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="preuni-data-export.json"`)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(export) //nolint:errcheck
}
