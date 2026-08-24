package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestIntegration_Register_OTPStoreFailsStillCreatesAccount covers
// RegisterHandler's otpRepo.Create-fails branch (storeErr, previously
// untested): register still succeeds (201, real tokens) since OTP storage
// failure is logged and non-fatal, but no verification OTP row is left
// behind. Query order: 1. credRepo.Create, 2. the student provisioner's
// Repo.Create, 3. otpRepo.Create.
func TestIntegration_Register_OTPStoreFailsStillCreatesAccount(t *testing.T) {
	failR, failPool := setupWithNthQueryFailure(t, 3)
	ctx := context.Background()

	email := fmt.Sprintf("register-otpfail+%d@preuni.test", time.Now().UnixNano())
	body, _ := json.Marshal(map[string]string{
		"email": email, "password": "P@ssw0rd123", "display_name": "NthQuery Register",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	failR.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		StudentID   string `json:"student_id"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.StudentID == "" || resp.AccessToken == "" {
		t.Fatal("expected registration to still succeed despite the OTP store failure")
	}
	defer cleanupTestUser(ctx, t, failPool, resp.StudentID)

	var count int
	if err := failPool.QueryRow(ctx,
		`SELECT count(*) FROM auth.otp_codes WHERE credential_id = $1 AND purpose = 'EMAIL_VERIFY'`,
		resp.StudentID,
	).Scan(&count); err != nil {
		t.Fatalf("check otp rows: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no verification OTP row after a failed store, got %d", count)
	}
}
