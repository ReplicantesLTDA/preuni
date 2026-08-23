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

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestIntegration_ChangePassword_Succeeds covers ChangePasswordHandler's
// success path -- CredentialsRepository.UpdatePasswordHash and
// RefreshTokenRepository.RevokeAllForCredential, previously only exercised
// via their missing-field validation branch (auth_validation_test.go).
func TestIntegration_ChangePassword_Succeeds(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	body, _ := json.Marshal(map[string]string{
		"current_password": "P@ssw0rd123",
		"new_password":     "N3wP@ssw0rd456",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/password/change", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("change password: got %d body=%s", w.Code, w.Body.String())
	}

	// The hash was actually updated: the old password no longer checks out
	// as "current" against it, and the new one does.
	retryOld, _ := json.Marshal(map[string]string{
		"current_password": "P@ssw0rd123",
		"new_password":     "AnotherOne789",
	})
	oldReq := httptest.NewRequest(http.MethodPost, "/v1/auth/password/change", bytes.NewReader(retryOld))
	oldReq.Header.Set("Content-Type", "application/json")
	oldReq.Header.Set("Authorization", "Bearer "+token)
	oldW := httptest.NewRecorder()
	r.ServeHTTP(oldW, oldReq)
	if oldW.Code != http.StatusUnprocessableEntity {
		t.Fatalf("old password should no longer be accepted as current_password: got %d body=%s", oldW.Code, oldW.Body.String())
	}

	retryNew, _ := json.Marshal(map[string]string{
		"current_password": "N3wP@ssw0rd456",
		"new_password":     "AnotherOne789",
	})
	newReq := httptest.NewRequest(http.MethodPost, "/v1/auth/password/change", bytes.NewReader(retryNew))
	newReq.Header.Set("Content-Type", "application/json")
	newReq.Header.Set("Authorization", "Bearer "+token)
	newW := httptest.NewRecorder()
	r.ServeHTTP(newW, newReq)
	if newW.Code != http.StatusNoContent {
		t.Fatalf("new password should be accepted as current_password: got %d body=%s", newW.Code, newW.Body.String())
	}
}

// TestIntegration_ChangePassword_WrongCurrentPasswordIsRejected covers
// ChangePasswordHandler's CheckPassword-fails branch, previously untested.
func TestIntegration_ChangePassword_WrongCurrentPasswordIsRejected(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	body, _ := json.Marshal(map[string]string{
		"current_password": "TotallyWrong123",
		"new_password":     "N3wP@ssw0rd456",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/password/change", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("change password with the wrong current password: got %d body=%s", w.Code, w.Body.String())
	}
}

// TestIntegration_ChangePassword_WeakNewPasswordIsRejected covers
// ChangePasswordHandler's ValidatePassword branch, previously untested.
func TestIntegration_ChangePassword_WeakNewPasswordIsRejected(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	body, _ := json.Marshal(map[string]string{
		"current_password": "P@ssw0rd123",
		"new_password":     "short",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/password/change", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("change password with a weak new password: got %d body=%s", w.Code, w.Body.String())
	}
}

// TestIntegration_ChangeEmail_RequestThenConfirmSucceeds covers both
// ChangeEmail handlers' success paths -- OTPRepository.Create/MarkUsed and
// CredentialsRepository.UpdateEmail -- previously only exercised via their
// missing-field validation branch.
func TestIntegration_ChangeEmail_RequestThenConfirmSucceeds(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	newEmail := "changed+" + studentID + "@preuni.test"

	reqBody, _ := json.Marshal(map[string]string{"new_email": newEmail})
	reqReq := httptest.NewRequest(http.MethodPost, "/v1/auth/email/change/request", bytes.NewReader(reqBody))
	reqReq.Header.Set("Content-Type", "application/json")
	reqReq.Header.Set("Authorization", "Bearer "+token)
	reqW := httptest.NewRecorder()
	r.ServeHTTP(reqW, reqReq)
	if reqW.Code != http.StatusAccepted {
		t.Fatalf("change email request: got %d body=%s", reqW.Code, reqW.Body.String())
	}

	otp := "424242"
	sum := sha256.Sum256([]byte(otp))
	hash := hex.EncodeToString(sum[:])
	if _, err := pool.Exec(ctx,
		`UPDATE auth.otp_codes SET code_hash=$1 WHERE credential_id=$2 AND purpose='EMAIL_CHANGE' AND used_at IS NULL`,
		hash, studentID,
	); err != nil {
		t.Fatalf("inject known otp: %v", err)
	}

	confirmBody, _ := json.Marshal(map[string]string{"new_email": newEmail, "otp": otp})
	confirmReq := httptest.NewRequest(http.MethodPost, "/v1/auth/email/change/confirm", bytes.NewReader(confirmBody))
	confirmReq.Header.Set("Content-Type", "application/json")
	confirmReq.Header.Set("Authorization", "Bearer "+token)
	confirmW := httptest.NewRecorder()
	r.ServeHTTP(confirmW, confirmReq)
	if confirmW.Code != http.StatusNoContent {
		t.Fatalf("change email confirm: got %d body=%s", confirmW.Code, confirmW.Body.String())
	}

	if got := studentEmail(t, ctx, pool, studentID); got != newEmail {
		t.Fatalf("email not updated: got %q, want %q", got, newEmail)
	}
}

// TestIntegration_ChangeEmail_ConfirmWithoutAnyOTPIsRejected covers
// ChangeEmailConfirmHandler's FindActiveByCredentialAndPurpose-fails
// branch (no OTP was ever requested), previously untested.
func TestIntegration_ChangeEmail_ConfirmWithoutAnyOTPIsRejected(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	body, _ := json.Marshal(map[string]string{"new_email": "new+" + studentID + "@preuni.test", "otp": "123456"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/email/change/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("change email confirm without any OTP requested: got %d body=%s", w.Code, w.Body.String())
	}
}

// TestIntegration_DeleteAccount_Succeeds covers DeleteAccountHandler's
// success path -- CredentialsRepository.Anonymize and
// RefreshTokenRepository.RevokeAllForCredential.
func TestIntegration_DeleteAccount_Succeeds(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	body, _ := json.Marshal(map[string]string{"confirmation": "DELETE"})
	req := httptest.NewRequest(http.MethodDelete, "/v1/auth/account", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete account: got %d body=%s", w.Code, w.Body.String())
	}

	got := studentEmail(t, ctx, pool, studentID)
	want := "deleted+" + studentID + "@preuni.com.br"
	if got != want {
		t.Fatalf("expected credentials to be anonymized after account deletion: got email %q, want %q", got, want)
	}
}

// TestIntegration_ResendVerification_UnknownEmailReturns204 covers the
// anti-enumeration branch of ResendVerificationHandler.ServeHTTP, which was
// entirely untested (0% coverage).
func TestIntegration_ResendVerification_UnknownEmailReturns204(t *testing.T) {
	r, _ := setup(t)

	body, _ := json.Marshal(map[string]string{"email": "nobody-here@preuni.test"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/email/verify-resend", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("resend for unknown email: got %d body=%s", w.Code, w.Body.String())
	}
}

// TestIntegration_ResendVerification_AlreadyVerifiedIsConflict covers the
// "already verified" branch, and TestIntegration_ResendVerification_Succeeds
// covers the success branch (OTP generation + storage) -- also untested.
func TestIntegration_ResendVerification_AlreadyVerifiedIsConflict(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)
	email := studentEmail(t, ctx, pool, studentID)

	if _, err := pool.Exec(ctx, `UPDATE auth.credentials SET email_verified = true WHERE id = $1`, studentID); err != nil {
		t.Fatalf("mark verified: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"email": email})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/email/verify-resend", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("resend for already-verified email: got %d body=%s", w.Code, w.Body.String())
	}
}

func TestIntegration_ResendVerification_Succeeds(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)
	email := studentEmail(t, ctx, pool, studentID)

	body, _ := json.Marshal(map[string]string{"email": email})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/email/verify-resend", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("resend: got %d body=%s", w.Code, w.Body.String())
	}
}

func studentEmail(t *testing.T, ctx context.Context, pool *pgxpool.Pool, studentID string) string {
	t.Helper()
	var email string
	if err := pool.QueryRow(ctx, `SELECT email FROM auth.credentials WHERE id = $1`, studentID).Scan(&email); err != nil {
		t.Fatalf("read email: %v", err)
	}
	return email
}
