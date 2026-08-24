package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestIntegration_OTPLoginRequest_CreateFailsLeavesNoOTPRow covers
// OTPLoginRequestHandler's background goroutine otpRepo.Create-fails
// branch (query #2 -- credRepo.FindByEmail, the first call, must succeed
// and find a verified user to even reach it). The handler always responds
// 202 regardless, so the only observable difference is DB state: no OTP
// row gets persisted when Create fails, vs. one does on success (covered
// by TestIntegration_OTPLoginRequest_GeneratesOTPForVerifiedUser).
func TestIntegration_OTPLoginRequest_CreateFailsLeavesNoOTPRow(t *testing.T) {
	fixtureR, fixturePool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, fixtureR)
	defer cleanupTestUser(ctx, t, fixturePool, studentID)
	markVerified(t, ctx, fixturePool, studentID)
	email := studentEmail(t, ctx, fixturePool, studentID)

	// Query order in the goroutine: 1. credRepo.FindByEmail, 2. otpRepo.Create.
	failR, failPool := setupWithNthQueryFailure(t, 2)
	_ = failPool

	body, _ := json.Marshal(map[string]string{"email": email})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/request", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	failR.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("otp request: got %d body=%s", w.Code, w.Body.String())
	}

	// Give the goroutine time to run and (fail to) persist -- then confirm
	// no OTP row was created, using the untraced fixture pool so this
	// check doesn't itself count against failR's query counter.
	time.Sleep(300 * time.Millisecond)
	var count int
	if err := fixturePool.QueryRow(ctx,
		`SELECT count(*) FROM auth.otp_codes WHERE credential_id = $1 AND purpose = 'LOGIN_OTP' AND used_at IS NULL`,
		studentID,
	).Scan(&count); err != nil {
		t.Fatalf("check otp rows: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no OTP row after a failed Create, got %d", count)
	}
}

// TestIntegration_PasswordResetRequest_CreateFailsLeavesNoOTPRow is the
// same gap for PasswordResetRequestHandler's background goroutine.
func TestIntegration_PasswordResetRequest_CreateFailsLeavesNoOTPRow(t *testing.T) {
	fixtureR, fixturePool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, fixtureR)
	defer cleanupTestUser(ctx, t, fixturePool, studentID)
	markVerified(t, ctx, fixturePool, studentID)
	email := studentEmail(t, ctx, fixturePool, studentID)

	failR, failPool := setupWithNthQueryFailure(t, 2)
	_ = failPool

	body, _ := json.Marshal(map[string]string{"email": email})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/password/reset/request", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	failR.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("password reset request: got %d body=%s", w.Code, w.Body.String())
	}

	time.Sleep(300 * time.Millisecond)
	var count int
	if err := fixturePool.QueryRow(ctx,
		`SELECT count(*) FROM auth.otp_codes WHERE credential_id = $1 AND purpose = 'PASSWORD_RESET' AND used_at IS NULL`,
		studentID,
	).Scan(&count); err != nil {
		t.Fatalf("check otp rows: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no OTP row after a failed Create, got %d", count)
	}
}
