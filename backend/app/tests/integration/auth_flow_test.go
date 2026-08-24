// Package integration_test exercises the full monolith router against a real
// Postgres. Set TEST_DB_URL (or rely on the default local compose URL) to run.
package integration_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preuni/pkg/logger"
	"github.com/preuni/app/internal/config"
	"github.com/preuni/app/internal/router"
)

func setup(t *testing.T) (http.Handler, *pgxpool.Pool) {
	t.Helper()
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://preuni:preuni@localhost:5432/preuni?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skipf("postgres unreachable: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Skipf("postgres ping failed: %v", err)
	}
	t.Cleanup(pool.Close)

	cfg := config.Config{
		JWTSigningKey:        "test-signing-key-32-bytes-minimum!",
		JWTAccessExpirySec:   3600,
		JWTRefreshExpiryDays: 30,
		MailFromAddr:         "noreply@test",
		MailFromName:         "Test",
		S3Bucket:             "test-bucket",
		S3Region:             "us-east-1",
	}
	r, _ := router.New(cfg, pool, logger.New("error"))
	return r, pool
}

func TestIntegration_FullAuthFlow(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	email := fmt.Sprintf("integ+%d@preuni.test", time.Now().UnixNano())
	password := "P@ssw0rd123"

	// 1. Register
	body, _ := json.Marshal(map[string]string{
		"email": email, "password": password, "display_name": "Integ",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: got %d body=%s", w.Code, w.Body.String())
	}
	var reg struct {
		StudentID    string `json:"student_id"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &reg); err != nil {
		t.Fatal(err)
	}
	if reg.StudentID == "" || reg.AccessToken == "" {
		t.Fatal("missing student_id or access_token")
	}

	// 2. /v1/students/me with the access token must work (proves in-process student provisioning).
	req2 := httptest.NewRequest(http.MethodGet, "/v1/students/me", nil)
	req2.Header.Set("Authorization", "Bearer "+reg.AccessToken)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("me: got %d body=%s", w2.Code, w2.Body.String())
	}

	// 3. Inject a known OTP and verify email.
	otp := "424242"
	sum := sha256.Sum256([]byte(otp))
	hash := hex.EncodeToString(sum[:])
	_, err := pool.Exec(ctx,
		`UPDATE auth.otp_codes SET code_hash=$1 WHERE credential_id=$2 AND purpose='EMAIL_VERIFY' AND used_at IS NULL`,
		hash, reg.StudentID)
	if err != nil {
		t.Fatal(err)
	}
	verifyBody, _ := json.Marshal(map[string]string{"email": email, "otp": otp})
	req3 := httptest.NewRequest(http.MethodPost, "/v1/auth/email/verify", bytes.NewReader(verifyBody))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusNoContent {
		t.Fatalf("verify: got %d body=%s", w3.Code, w3.Body.String())
	}

	// 4. Login after verify.
	loginBody, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req4 := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(loginBody))
	req4.Header.Set("Content-Type", "application/json")
	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, req4)
	if w4.Code != http.StatusOK {
		t.Fatalf("login: got %d body=%s", w4.Code, w4.Body.String())
	}

	// 5. Refresh token rotates.
	refreshBody, _ := json.Marshal(map[string]string{"refresh_token": reg.RefreshToken})
	req5 := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewReader(refreshBody))
	req5.Header.Set("Content-Type", "application/json")
	w5 := httptest.NewRecorder()
	r.ServeHTTP(w5, req5)
	if w5.Code != http.StatusOK {
		t.Fatalf("refresh: got %d body=%s", w5.Code, w5.Body.String())
	}

	// Cleanup: anonymize this account's data.
	_, _ = pool.Exec(ctx, `DELETE FROM auth.refresh_tokens WHERE credential_id=$1`, reg.StudentID)
	_, _ = pool.Exec(ctx, `DELETE FROM auth.otp_codes WHERE credential_id=$1`, reg.StudentID)
	_, _ = pool.Exec(ctx, `DELETE FROM auth.credentials WHERE id=$1`, reg.StudentID)
	_, _ = pool.Exec(ctx, `DELETE FROM users.students WHERE id=$1`, reg.StudentID)
}
