package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestIntegration_Login_MissingFieldsIsRejected covers LoginHandler's
// missing-email-or-password validation branch, previously untested (only
// invalid-JSON, unknown-email, unverified-email, and wrong-password
// branches had coverage).
func TestIntegration_Login_MissingFieldsIsRejected(t *testing.T) {
	r, _ := setup(t)

	body, _ := json.Marshal(map[string]string{"email": "someone@preuni.test"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("login with missing password: got %d body=%s", w.Code, w.Body.String())
	}
}

// TestIntegration_OTPLoginVerify_WrongCodeIsRejected covers
// OTPLoginVerifyHandler's otpRecord.CodeHash-mismatch branch specifically
// (a real OTP row exists, but the submitted code doesn't match) --
// distinct from the already-covered "no OTP requested at all"
// (FindActiveByCredentialAndPurpose error) and unknown-email branches.
func TestIntegration_OTPLoginVerify_WrongCodeIsRejected(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)
	seedOTP(t, ctx, pool, studentID, "LOGIN_OTP", "111111")
	email := studentEmail(t, ctx, pool, studentID)

	body, _ := json.Marshal(map[string]string{"email": email, "otp": "999999"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("otp verify with wrong code: got %d body=%s", w.Code, w.Body.String())
	}
}

// TestIntegration_VerifyEmail_InvalidJSONIsRejected covers
// VerifyEmailHandler's json.Decode-fails branch, previously untested.
func TestIntegration_VerifyEmail_InvalidJSONIsRejected(t *testing.T) {
	r, _ := setup(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/email/verify", bytes.NewReader([]byte("{not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("verify email with invalid JSON: got %d body=%s", w.Code, w.Body.String())
	}
}

// TestIntegration_VerifyEmail_MissingFieldsIsRejected covers
// VerifyEmailHandler's missing-email-or-otp validation branch, previously
// untested.
func TestIntegration_VerifyEmail_MissingFieldsIsRejected(t *testing.T) {
	r, _ := setup(t)

	body, _ := json.Marshal(map[string]string{"email": "someone@preuni.test"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/email/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("verify email with missing otp: got %d body=%s", w.Code, w.Body.String())
	}
}

// TestIntegration_VerifyEmail_WrongCodeIsRejected covers VerifyEmailHandler's
// otpRecord.CodeHash-mismatch branch specifically (a real OTP row exists,
// submitted code doesn't match) -- distinct from the already-covered
// unknown-email branch.
func TestIntegration_VerifyEmail_WrongCodeIsRejected(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)
	seedOTP(t, ctx, pool, studentID, "EMAIL_VERIFY", "222222")
	email := studentEmail(t, ctx, pool, studentID)

	body, _ := json.Marshal(map[string]string{"email": email, "otp": "000000"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/email/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("verify email with wrong code: got %d body=%s", w.Code, w.Body.String())
	}
}

// TestIntegration_ResendVerification_InvalidJSONIsRejected covers
// ResendVerificationHandler's json.Decode-fails-or-missing-email branch,
// previously untested (unknown-email, already-verified, and success
// branches were covered).
func TestIntegration_ResendVerification_InvalidJSONIsRejected(t *testing.T) {
	r, _ := setup(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/email/verify-resend", bytes.NewReader([]byte("{not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("resend with invalid JSON: got %d body=%s", w.Code, w.Body.String())
	}
}
