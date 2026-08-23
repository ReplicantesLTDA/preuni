package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestIntegration_Onboarding_MarksCompleted covers PATCH /v1/students/me/onboarding.
func TestIntegration_Onboarding_MarksCompleted(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	req := httptest.NewRequest(http.MethodPatch, "/v1/students/me/onboarding", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200, body=%s", w.Code, w.Body.String())
	}

	var onboardingCompleted bool
	if err := pool.QueryRow(ctx, `SELECT onboarding_completed FROM users.students WHERE id = $1`, studentID).Scan(&onboardingCompleted); err != nil {
		t.Fatal(err)
	}
	if !onboardingCompleted {
		t.Fatal("expected onboarding_completed = true")
	}

	// Idempotent: calling it again must not error.
	req2 := httptest.NewRequest(http.MethodPatch, "/v1/students/me/onboarding", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second call: got %d, want 200, body=%s", w2.Code, w2.Body.String())
	}
}

// TestIntegration_DeleteStudent_Anonymizes covers DELETE /v1/students/me.
func TestIntegration_DeleteStudent_Anonymizes(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	req := httptest.NewRequest(http.MethodDelete, "/v1/students/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("got %d, want 204, body=%s", w.Code, w.Body.String())
	}

	var displayName string
	var xpTotal int
	if err := pool.QueryRow(ctx, `SELECT display_name, xp_total FROM users.students WHERE id = $1`, studentID).Scan(&displayName, &xpTotal); err != nil {
		t.Fatal(err)
	}
	if displayName != "Deleted User" {
		t.Fatalf("expected display_name to be anonymized, got %q", displayName)
	}
	if xpTotal != 0 {
		t.Fatalf("expected xp_total reset to 0, got %d", xpTotal)
	}
}

// TestIntegration_DataExport_ReturnsProfile covers GET
// /v1/students/me/data-export, which had 0% coverage.
func TestIntegration_DataExport_ReturnsProfile(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	req := httptest.NewRequest(http.MethodGet, "/v1/students/me/data-export", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200, body=%s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); cd == "" {
		t.Fatalf("expected a Content-Disposition attachment header")
	}

	var export struct {
		ExportedAt string `json:"exported_at"`
		Profile    struct {
			ID string `json:"id"`
		} `json:"profile"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &export); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if export.Profile.ID != studentID {
		t.Fatalf("expected profile.id %q, got %q", studentID, export.Profile.ID)
	}
	if export.ExportedAt == "" {
		t.Fatal("expected a non-empty exported_at timestamp")
	}
}
