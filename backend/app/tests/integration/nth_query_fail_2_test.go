package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// More "first DB call succeeds, second (or later) DB call fails" branches
// via setupWithNthQueryFailure -- see nth_query_fail_helper_test.go and
// nth_query_fail_test.go for the pattern this file continues.

// TestIntegration_ChangePassword_LaterCallFailures covers
// ChangePasswordHandler's UpdatePasswordHash/RevokeAllForCredential-fails
// branches (queries #2/#3 -- FindByID, the first call, must succeed).
func TestIntegration_ChangePassword_LaterCallFailures(t *testing.T) {
	fixtureR, fixturePool := setup(t)
	ctx := context.Background()

	// Query order inside ChangePasswordHandler.ServeHTTP:
	//   1. credRepo.FindByID
	//   2. credRepo.UpdatePasswordHash
	//   3. refreshRepo.RevokeAllForCredential
	cases := []struct {
		name string
		n    int64
	}{
		{"UpdatePasswordHash fails", 2},
		{"RevokeAllForCredential fails", 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			studentID, token := registerTestUser(t, fixtureR)
			t.Cleanup(func() { cleanupTestUser(ctx, t, fixturePool, studentID) })

			r, _ := setupWithNthQueryFailure(t, tc.n)

			body, _ := json.Marshal(map[string]string{
				"current_password": "P@ssw0rd123",
				"new_password":     "N3wP@ssw0rd456",
			})
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/password/change", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code == http.StatusNoContent {
				t.Fatalf("expected an error status when query #%d fails, got 204", tc.n)
			}
		})
	}
}

// TestIntegration_ChangeEmailConfirm_LaterCallFailures covers
// ChangeEmailConfirmHandler's MarkUsed/UpdateEmail/RevokeAllForCredential
// -fails branches (queries #2/#3/#4 -- FindActiveByCredentialAndPurpose,
// the first call, must succeed with a real active OTP).
func TestIntegration_ChangeEmailConfirm_LaterCallFailures(t *testing.T) {
	fixtureR, fixturePool := setup(t)
	ctx := context.Background()

	// Query order inside ChangeEmailConfirmHandler.ServeHTTP:
	//   1. otpRepo.FindActiveByCredentialAndPurpose
	//   2. otpRepo.MarkUsed
	//   3. credRepo.UpdateEmail
	//   4. refreshRepo.RevokeAllForCredential
	cases := []struct {
		name string
		n    int64
	}{
		{"MarkUsed fails", 2},
		{"UpdateEmail fails", 3},
		{"RevokeAllForCredential fails", 4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			studentID, token := registerTestUser(t, fixtureR)
			t.Cleanup(func() { cleanupTestUser(ctx, t, fixturePool, studentID) })

			otp := "135791"
			seedOTP(t, ctx, fixturePool, studentID, "EMAIL_CHANGE", otp)

			r, _ := setupWithNthQueryFailure(t, tc.n)

			newEmail := "changed-nthq+" + studentID + "@preuni.test"
			body, _ := json.Marshal(map[string]string{"new_email": newEmail, "otp": otp})
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/email/change/confirm", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code == http.StatusNoContent {
				t.Fatalf("expected an error status when query #%d fails, got 204", tc.n)
			}
		})
	}
}

// TestIntegration_PasswordResetConfirm_LaterCallFailures covers
// PasswordResetHandler's MarkUsed/UpdatePasswordHash/RevokeAllForCredential
// -fails branches (queries #3/#4/#5 -- FindByEmail and
// FindActiveByCredentialAndPurpose, the first two calls, must both
// succeed with a real active OTP).
func TestIntegration_PasswordResetConfirm_LaterCallFailures(t *testing.T) {
	fixtureR, fixturePool := setup(t)
	ctx := context.Background()

	// Query order inside PasswordResetHandler.ServeHTTP:
	//   1. credRepo.FindByEmail
	//   2. otpRepo.FindActiveByCredentialAndPurpose
	//   3. otpRepo.MarkUsed
	//   4. credRepo.UpdatePasswordHash
	//   5. refreshRepo.RevokeAllForCredential
	cases := []struct {
		name string
		n    int64
	}{
		{"MarkUsed fails", 3},
		{"UpdatePasswordHash fails", 4},
		{"RevokeAllForCredential fails", 5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			studentID, _ := registerTestUser(t, fixtureR)
			t.Cleanup(func() { cleanupTestUser(ctx, t, fixturePool, studentID) })
			email := studentEmail(t, ctx, fixturePool, studentID)

			otp := "975310"
			seedOTP(t, ctx, fixturePool, studentID, "PASSWORD_RESET", otp)

			r, _ := setupWithNthQueryFailure(t, tc.n)

			body, _ := json.Marshal(map[string]string{
				"email": email, "otp": otp, "new_password": "N3wP@ssw0rd789",
			})
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/password/reset/confirm", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code == http.StatusNoContent {
				t.Fatalf("expected an error status when query #%d fails, got 204", tc.n)
			}
		})
	}
}

// TestIntegration_OTPLoginVerify_LaterCallFailures covers
// OTPLoginVerifyHandler's MarkUsed/Store-fails branches (queries #3/#4 --
// FindByEmail and FindActiveByCredentialAndPurpose, the first two calls,
// must both succeed with a real active OTP).
func TestIntegration_OTPLoginVerify_LaterCallFailures(t *testing.T) {
	fixtureR, fixturePool := setup(t)
	ctx := context.Background()

	// Query order inside OTPLoginVerifyHandler.ServeHTTP:
	//   1. credRepo.FindByEmail
	//   2. otpRepo.FindActiveByCredentialAndPurpose
	//   3. otpRepo.MarkUsed
	//   4. refreshRepo.Store (IssueAccessToken/IssueRefreshToken between
	//      #3 and #4 are pure crypto, not DB calls)
	cases := []struct {
		name string
		n    int64
	}{
		{"MarkUsed fails", 3},
		{"Store fails", 4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			studentID, _ := registerTestUser(t, fixtureR)
			t.Cleanup(func() { cleanupTestUser(ctx, t, fixturePool, studentID) })
			email := studentEmail(t, ctx, fixturePool, studentID)

			otp := "864209"
			seedOTP(t, ctx, fixturePool, studentID, "LOGIN_OTP", otp)

			r, _ := setupWithNthQueryFailure(t, tc.n)

			body, _ := json.Marshal(map[string]string{"email": email, "otp": otp})
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code == http.StatusOK {
				t.Fatalf("expected an error status when query #%d fails, got 200 body=%s", tc.n, w.Body.String())
			}
		})
	}
}

// TestIntegration_DeleteAccount_AnonymizeFails covers DeleteAccountHandler's
// Anonymize-fails branch (query #2 -- RevokeAllForCredential, the first
// call, must succeed).
func TestIntegration_DeleteAccount_AnonymizeFails(t *testing.T) {
	fixtureR, fixturePool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, fixtureR)
	defer cleanupTestUser(ctx, t, fixturePool, studentID)

	// Query order inside DeleteAccountHandler.ServeHTTP:
	//   1. refreshRepo.RevokeAllForCredential
	//   2. credRepo.Anonymize
	r, _ := setupWithNthQueryFailure(t, 2)

	body, _ := json.Marshal(map[string]string{"confirmation": "DELETE"})
	req := httptest.NewRequest(http.MethodDelete, "/v1/auth/account", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusNoContent {
		t.Fatalf("expected an error status when Anonymize fails, got 204")
	}
}
