package integration_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestIntegration_UpdateStudent_InvalidJSONIsRejected covers
// UpdateStudentHandler's json.Decode-fails branch, previously untested
// (only its validation and canceled-context branches were covered).
func TestIntegration_UpdateStudent_InvalidJSONIsRejected(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	req := httptest.NewRequest(http.MethodPatch, "/v1/students/me", bytes.NewReader([]byte("{not valid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// Like RegisterHandler, this maps a raw json.Decode error straight to
	// pkgmw.ErrorResponse, which defaults untyped errors to 500 (not the
	// apperrors.Validation 422 used elsewhere) -- documenting current
	// behavior, not asserting a "should be" status.
	if w.Code == http.StatusOK {
		t.Fatalf("update student with invalid JSON: expected an error status, got 200 body=%s", w.Body.String())
	}
}

// TestIntegration_SubmitEssay_InvalidJSONIsRejected covers
// SubmitEssayHandler's json.Decode-fails branch, previously untested
// (only its missing-field validation branch was covered).
func TestIntegration_SubmitEssay_InvalidJSONIsRejected(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	req := httptest.NewRequest(http.MethodPost, "/v1/essays", bytes.NewReader([]byte("{not valid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusBadRequest {
		t.Fatalf("submit essay with invalid JSON: got %d body=%s", w.Code, w.Body.String())
	}
}
