package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestIntegration_Auth_ValidationErrorPaths covers the missing-field/bad-body
// branches across the auth endpoints that don't otherwise have dedicated
// integration test files (change_password, change_email, delete_account,
// otp_login, password_reset, refresh_token) -- these are always-202 or
// always-validate-first handlers by design (anti-enumeration), so a plain
// missing-field request is enough to exercise the branch without needing
// email delivery.
func TestIntegration_Auth_ValidationErrorPaths(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	cases := []struct {
		name       string
		method     string
		path       string
		body       map[string]string
		auth       bool
		wantStatus int
	}{
		{"otp request with missing email still 202s (anti-enumeration)", http.MethodPost, "/v1/auth/otp/request", map[string]string{}, false, http.StatusAccepted},
		{"otp verify with missing fields is a validation error", http.MethodPost, "/v1/auth/otp/verify", map[string]string{}, false, http.StatusUnprocessableEntity},
		{"password reset request with missing email still 202s", http.MethodPost, "/v1/auth/password/reset/request", map[string]string{}, false, http.StatusAccepted},
		{"password reset confirm with missing fields is a validation error", http.MethodPost, "/v1/auth/password/reset/confirm", map[string]string{}, false, http.StatusUnprocessableEntity},
		{"refresh with missing token is a validation error", http.MethodPost, "/v1/auth/refresh", map[string]string{}, false, http.StatusUnprocessableEntity},
		{"change password with missing fields is a validation error", http.MethodPost, "/v1/auth/password/change", map[string]string{}, true, http.StatusUnprocessableEntity},
		{"change email request with missing new_email is a validation error", http.MethodPost, "/v1/auth/email/change/request", map[string]string{}, true, http.StatusUnprocessableEntity},
		{"change email confirm with missing fields is a validation error", http.MethodPost, "/v1/auth/email/change/confirm", map[string]string{}, true, http.StatusUnprocessableEntity},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tc.auth {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.wantStatus {
				t.Fatalf("got %d, want %d, body=%s", w.Code, tc.wantStatus, w.Body.String())
			}
		})
	}

	t.Run("delete account with wrong confirmation is a validation error", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"confirmation": "nope"})
		req := httptest.NewRequest(http.MethodDelete, "/v1/auth/account", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("got %d, want 422, body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("submit essay with missing fields is a validation error", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"prompt_theme_title": "Tema"})
		req := httptest.NewRequest(http.MethodPost, "/v1/essays", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("got %d, want 422, body=%s", w.Code, w.Body.String())
		}
	})
}
