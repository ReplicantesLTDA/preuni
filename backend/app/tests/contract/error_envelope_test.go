// Package contract_test verifies the wire-level contract preserved by the
// monolith. Tests here cover error envelope shape, JWT auth requirements,
// X-Request-ID propagation, and health.
package contract_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preuni/app/internal/config"
	"github.com/preuni/app/internal/router"
	"github.com/preuni/pkg/logger"
)

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
		MailFromAddr:         "noreply@test",
		MailFromName:         "Test",
		StorageEndpoint:      "localhost:9000",
		StorageAccessKey:     "test",
		StorageSecretKey:     "test",
		StorageBucket:        "test-bucket",
	}
	r, _, _ := router.New(cfg, pool, logger.New("error"))
	return r
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
