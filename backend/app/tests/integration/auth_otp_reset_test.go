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

// TestIntegration_OTPLoginRequest_GeneratesOTPForVerifiedUser covers
// OTPLoginRequestHandler's success branch (email found + verified), which
// runs in a background goroutine and was otherwise only exercised via its
// "missing email still 202s" validation branch.
func TestIntegration_OTPLoginRequest_GeneratesOTPForVerifiedUser(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)
	markVerified(t, ctx, pool, studentID)
	email := studentEmail(t, ctx, pool, studentID)

	body, _ := json.Marshal(map[string]string{"email": email})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/request", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("otp request: got %d body=%s", w.Code, w.Body.String())
	}

	waitForOTP(t, ctx, pool, studentID, "LOGIN_OTP")
}

// TestIntegration_PasswordResetRequest_GeneratesOTPForVerifiedUser is the
// same coverage gap as above, for PasswordResetRequestHandler.
func TestIntegration_PasswordResetRequest_GeneratesOTPForVerifiedUser(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)
	markVerified(t, ctx, pool, studentID)
	email := studentEmail(t, ctx, pool, studentID)

	body, _ := json.Marshal(map[string]string{"email": email})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/password/reset/request", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("password reset request: got %d body=%s", w.Code, w.Body.String())
	}

	waitForOTP(t, ctx, pool, studentID, "PASSWORD_RESET")
}

// TestIntegration_RefreshToken_RotatesAndReturnsNewTokens covers
// RefreshTokenHandler's success path (find, rotate, issue), previously
// only exercised via its missing-field validation branch.
func TestIntegration_RefreshToken_RotatesAndReturnsNewTokens(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	// registerTestUser doesn't capture refresh_token, so grab it directly
	// with a fresh registration on this same handler.
	email := fmt.Sprintf("refresh-integ+%d@preuni.test", time.Now().UnixNano())
	regBody, _ := json.Marshal(map[string]string{
		"email": email, "password": "P@ssw0rd123", "display_name": "Refresh Integ",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	r.ServeHTTP(regW, regReq)
	if regW.Code != http.StatusCreated {
		t.Fatalf("register: got %d body=%s", regW.Code, regW.Body.String())
	}
	var reg struct {
		StudentID    string `json:"student_id"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(regW.Body.Bytes(), &reg); err != nil {
		t.Fatal(err)
	}
	defer cleanupTestUser(ctx, t, pool, reg.StudentID)

	body, _ := json.Marshal(map[string]string{"refresh_token": reg.RefreshToken})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("refresh: got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.RefreshToken == "" || resp.RefreshToken == reg.RefreshToken {
		t.Fatalf("expected a newly rotated refresh token, got %q", resp.RefreshToken)
	}

	// The old (rotated-out) refresh token no longer works.
	oldBody, _ := json.Marshal(map[string]string{"refresh_token": reg.RefreshToken})
	oldReq := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewReader(oldBody))
	oldReq.Header.Set("Content-Type", "application/json")
	oldW := httptest.NewRecorder()
	r.ServeHTTP(oldW, oldReq)
	if oldW.Code != http.StatusUnauthorized {
		t.Fatalf("expected the rotated-out token to be rejected: got %d body=%s", oldW.Code, oldW.Body.String())
	}
}

// TestIntegration_VerifyEmail_AlreadyVerifiedIsConflict covers
// VerifyEmailHandler's "already verified" branch, which the register->verify
// flow in auth_flow_test.go never reaches (it only verifies once).
func TestIntegration_VerifyEmail_AlreadyVerifiedIsConflict(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)
	markVerified(t, ctx, pool, studentID)
	email := studentEmail(t, ctx, pool, studentID)

	body, _ := json.Marshal(map[string]string{"email": email, "otp": "000000"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/email/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("verify for already-verified email: got %d body=%s", w.Code, w.Body.String())
	}
}

func markVerified(t *testing.T, ctx context.Context, pool *pgxpool.Pool, credentialID string) {
	t.Helper()
	if _, err := pool.Exec(ctx, `UPDATE auth.credentials SET email_verified = true WHERE id = $1`, credentialID); err != nil {
		t.Fatalf("mark verified: %v", err)
	}
}

// waitForOTP polls for the request handlers' background goroutine to
// persist its OTP row, up to a couple seconds.
func waitForOTP(t *testing.T, ctx context.Context, pool *pgxpool.Pool, credentialID, purpose string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var count int
		if err := pool.QueryRow(ctx,
			`SELECT count(*) FROM auth.otp_codes WHERE credential_id = $1 AND purpose = $2 AND used_at IS NULL`,
			credentialID, purpose,
		).Scan(&count); err != nil {
			t.Fatalf("poll for otp: %v", err)
		}
		if count > 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for a %s otp to be generated", purpose)
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
