package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestIntegration_AvatarUpload_ReturnsPresignedURL covers PUT
// /v1/students/me/avatar (AvatarHandler.ServeUpload), previously 0%
// coverage -- a pure stub-URL builder with no DB call.
func TestIntegration_AvatarUpload_ReturnsPresignedURL(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	req := httptest.NewRequest(http.MethodPut, "/v1/students/me/avatar", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200, body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		UploadURL string `json:"upload_url"`
		ObjectKey string `json:"object_key"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if resp.UploadURL == "" || resp.ObjectKey == "" {
		t.Fatalf("expected non-empty upload_url and object_key, got %+v", resp)
	}
}

// TestIntegration_AvatarConfirm_MissingObjectKeyReturns422 covers
// AvatarHandler.ServeConfirm's validation branch (empty/decode-failed
// body), previously 0% coverage.
func TestIntegration_AvatarConfirm_MissingObjectKeyReturns422(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	req := httptest.NewRequest(http.MethodPost, "/v1/students/me/avatar/confirm", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("got %d, want 422, body=%s", w.Code, w.Body.String())
	}
}

// TestIntegration_AvatarConfirm_UpdatesAvatarURL covers ServeConfirm's
// success path -- studentRepo.Update is called with the built avatar URL
// and the response reflects it.
func TestIntegration_AvatarConfirm_UpdatesAvatarURL(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	body, _ := json.Marshal(map[string]string{"object_key": "avatars/x/123.webp"})
	req := httptest.NewRequest(http.MethodPost, "/v1/students/me/avatar/confirm", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200, body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if resp.AvatarURL == "" {
		t.Fatalf("expected non-empty avatar_url in response, got %+v", resp)
	}

	var dbAvatarURL string
	if err := pool.QueryRow(ctx, `SELECT avatar_url FROM users.students WHERE id = $1`, studentID).Scan(&dbAvatarURL); err != nil {
		t.Fatal(err)
	}
	if dbAvatarURL != resp.AvatarURL {
		t.Fatalf("expected DB avatar_url %q, got %q", resp.AvatarURL, dbAvatarURL)
	}
}

// TestIntegration_AvatarConfirm_CanceledContextReturnsError covers
// ServeConfirm's error-response branch when studentRepo.Update fails --
// a genuinely canceled context (real client disconnect/timeout
// behavior), previously unreachable since this handler had 0% coverage.
func TestIntegration_AvatarConfirm_CanceledContextReturnsError(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	body, _ := json.Marshal(map[string]string{"object_key": "avatars/x/456.webp"})
	req := httptest.NewRequest(http.MethodPost, "/v1/students/me/avatar/confirm", bytes.NewReader(body)).WithContext(canceled)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Fatalf("expected an error status for a canceled context, got 200 body=%s", w.Body.String())
	}
}
