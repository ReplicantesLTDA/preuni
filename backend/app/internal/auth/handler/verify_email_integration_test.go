package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/preuni/app/internal/auth/domain"
	"github.com/preuni/app/internal/auth/handler"
	"github.com/preuni/app/internal/auth/handler/testhelper"
	"github.com/preuni/app/internal/auth/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


// TestIntegration_VerifyEmail_ValidOTP asserts that submitting the correct
// 6-digit OTP returns HTTP 204 No Content.
func TestIntegration_VerifyEmail_ValidOTP(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	ctx := context.Background()

	credRepo := repository.NewCredentialsRepository(pool)
	otpRepo := repository.NewOTPRepository(pool)
	h := handler.NewVerifyEmailHandler(credRepo, otpRepo)

	const email = "verifyok@example.com"

	// Create an unverified credential.
	creds, err := domain.NewCredentials(email, "Passw0rd!")
	require.NoError(t, err)
	creds.ID = "00000000-0000-0000-0000-000000000010"
	require.NoError(t, credRepo.Create(ctx, creds))

	// Generate and persist a real OTP for that credential.
	plaintext, hash, expiry, err := domain.GenerateOTP()
	require.NoError(t, err)
	require.NoError(t, otpRepo.Create(ctx, creds.ID, domain.OTPPurposeEmailVerify, hash, expiry))

	body, _ := json.Marshal(map[string]string{"email": email, "otp": plaintext})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/email/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

	// Confirm the credential is now marked verified in the DB.
	verified, err := credRepo.FindByEmail(ctx, email)
	require.NoError(t, err)
	assert.True(t, verified.EmailVerified)
}

// TestIntegration_VerifyEmail_InvalidOTP asserts that submitting a wrong code
// returns HTTP 422 (validation error).
func TestIntegration_VerifyEmail_InvalidOTP(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	ctx := context.Background()

	credRepo := repository.NewCredentialsRepository(pool)
	otpRepo := repository.NewOTPRepository(pool)
	h := handler.NewVerifyEmailHandler(credRepo, otpRepo)

	const email = "verifybad@example.com"

	creds, err := domain.NewCredentials(email, "Passw0rd!")
	require.NoError(t, err)
	creds.ID = "00000000-0000-0000-0000-000000000011"
	require.NoError(t, credRepo.Create(ctx, creds))

	_, hash, expiry, err := domain.GenerateOTP()
	require.NoError(t, err)
	require.NoError(t, otpRepo.Create(ctx, creds.ID, domain.OTPPurposeEmailVerify, hash, expiry))

	body, _ := json.Marshal(map[string]string{"email": email, "otp": "000000"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/email/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
}
