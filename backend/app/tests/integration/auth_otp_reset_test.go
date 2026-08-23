package integration_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestIntegration_OTPLogin_VerifySucceeds covers OTPLoginVerifyHandler's
// success path. The request side (OTPLoginRequestHandler) generates its OTP
// in a background goroutine, so this test bypasses it and seeds the OTP row
// directly -- the same trick auth_flow_test.go uses for email verification.
func TestIntegration_OTPLogin_VerifySucceeds(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	otp := "135790"
	seedOTP(t, ctx, pool, studentID, "LOGIN_OTP", otp)

	email := studentEmail(t, ctx, pool, studentID)
	body, _ := json.Marshal(map[string]string{"email": email, "otp": otp})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("otp verify: got %d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.AccessToken == "" {
		t.Fatal("expected a non-empty access_token")
	}
}

// TestIntegration_PasswordReset_ConfirmSucceeds covers PasswordResetHandler's
// success path (the confirm side), bypassing the async request side the
// same way TestIntegration_OTPLogin_VerifySucceeds does.
func TestIntegration_PasswordReset_ConfirmSucceeds(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	otp := "246810"
	seedOTP(t, ctx, pool, studentID, "PASSWORD_RESET", otp)

	email := studentEmail(t, ctx, pool, studentID)
	body, _ := json.Marshal(map[string]string{"email": email, "otp": otp, "new_password": "Reset1234!"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/password/reset/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("password reset: got %d body=%s", w.Code, w.Body.String())
	}

	// The hash was actually updated: the new password now checks out as
	// "current" against it via the change-password endpoint.
	confirmBody, _ := json.Marshal(map[string]string{
		"current_password": "Reset1234!",
		"new_password":     "AnotherOne789",
	})
	confirmReq := httptest.NewRequest(http.MethodPost, "/v1/auth/password/change", bytes.NewReader(confirmBody))
	confirmReq.Header.Set("Content-Type", "application/json")
	confirmReq.Header.Set("Authorization", "Bearer "+token)
	confirmW := httptest.NewRecorder()
	r.ServeHTTP(confirmW, confirmReq)
	if confirmW.Code != http.StatusNoContent {
		t.Fatalf("expected the reset password to be accepted as current_password: got %d body=%s", confirmW.Code, confirmW.Body.String())
	}
}

// seedOTP inserts an OTP row directly, bypassing the async request handlers
// that would normally create one, with a known code so the test can present
// it back to the confirm/verify endpoint.
func seedOTP(t *testing.T, ctx context.Context, pool *pgxpool.Pool, credentialID, purpose, code string) {
	t.Helper()
	sum := sha256.Sum256([]byte(code))
	hash := hex.EncodeToString(sum[:])
	_, err := pool.Exec(ctx,
		`INSERT INTO auth.otp_codes (credential_id, purpose, code_hash, expires_at) VALUES ($1, $2, $3, $4)`,
		credentialID, purpose, hash, time.Now().Add(10*time.Minute),
	)
	if err != nil {
		t.Fatalf("seed otp: %v", err)
	}
}
