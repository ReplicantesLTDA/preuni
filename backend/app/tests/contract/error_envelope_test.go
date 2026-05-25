// Package contract_test verifies the wire-level contract preserved by the
// monolith. Tests here exercise routes that do not require a database
// connection (auth/internal-token failures, request-id propagation).
package contract_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preuni/pkg/logger"
	"github.com/preuni/app/internal/config"
	"github.com/preuni/app/internal/router"
)

// newTestRouter constructs the monolith router. It requires TEST_DB_URL for
// DB-touching routes; non-DB routes (401/422 paths, /health, X-Request-ID)
// still work even when the pool can't reach a DB, because handlers short-
// circuit before touching the pool.
func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://preuni:preuni@localhost:5432/preuni?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skipf("TEST_DB_URL pool init failed: %v", err)
	}
	t.Cleanup(pool.Close)

	cfg := config.Config{
		Port:                 "0",
		JWTSigningKey:        "test-signing-key-32-bytes-minimum!",
		JWTAccessExpirySec:   3600,
		JWTRefreshExpiryDays: 30,
		InternalServiceToken: "test-internal-token",
		MailFromAddr:         "noreply@test",
		MailFromName:         "Test",
		SelfBaseURL:          "http://localhost:0",
		S3Bucket:             "test-bucket",
		S3Region:             "us-east-1",
	}
	return router.New(cfg, pool, logger.New("error"))
}

func TestContract_MissingJWT_Returns401Envelope(t *testing.T) {
	r := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/students/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d want 401", w.Code)
	}
	var env struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Error.Code != "UNAUTHORIZED" {
		t.Fatalf("error.code: got %q want UNAUTHORIZED", env.Error.Code)
	}
	if env.Error.Message == "" {
		t.Fatal("error.message must be non-empty")
	}
}

func TestContract_MissingInternalToken_Returns401(t *testing.T) {
	r := newTestRouter(t)
	body, _ := json.Marshal(map[string]any{
		"type": "EMAIL_VERIFY", "to": "a@b.c", "params": map[string]string{"otp": "1"},
	})
	req := httptest.NewRequest(http.MethodPost, "/internal/email/send", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d want 401", w.Code)
	}
}

func TestContract_InvalidInternalToken_Returns401(t *testing.T) {
	r := newTestRouter(t)
	body, _ := json.Marshal(map[string]any{
		"type": "EMAIL_VERIFY", "to": "a@b.c", "params": map[string]string{"otp": "1"},
	})
	req := httptest.NewRequest(http.MethodPost, "/internal/email/send", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d want 401", w.Code)
	}
}

func TestContract_EmailValidation_Returns422(t *testing.T) {
	r := newTestRouter(t)
	body, _ := json.Marshal(map[string]any{
		"type": "WELCOME", "to": "a@b.c",
		"params": map[string]string{"display_name": "Only Name"},
	})
	req := httptest.NewRequest(http.MethodPost, "/internal/email/send", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-internal-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status: got %d want 422", w.Code)
	}
	var resp map[string]string
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp["error"] == "" {
		t.Fatal("error message must be non-empty")
	}
}

func TestContract_EmailSend_Returns202(t *testing.T) {
	r := newTestRouter(t)
	body, _ := json.Marshal(map[string]any{
		"type": "EMAIL_VERIFY", "to": "a@b.c",
		"params": map[string]string{"otp": "123456"},
	})
	req := httptest.NewRequest(http.MethodPost, "/internal/email/send", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-internal-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("status: got %d want 202", w.Code)
	}
}

func TestContract_RequestID_Propagation(t *testing.T) {
	r := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "test-req-id-42")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", w.Code)
	}
	if got := w.Header().Get("X-Request-Id"); got != "test-req-id-42" {
		t.Fatalf("X-Request-Id: got %q want test-req-id-42", got)
	}
}

func TestContract_Health_Returns200(t *testing.T) {
	r := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", w.Code)
	}
}
