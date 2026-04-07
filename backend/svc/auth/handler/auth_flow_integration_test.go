package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/preuni/svc/auth/domain"
	"github.com/preuni/svc/auth/handler"
	"github.com/preuni/svc/auth/handler/testhelper"
	"github.com/preuni/svc/auth/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_FullAuthFlow exercises the complete register → verify-email → login
// sequence against a real PostgreSQL instance.
func TestIntegration_FullAuthFlow(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	_ = context.Background() // used by helpers

	credRepo := repository.NewCredentialsRepository(pool)
	otpRepo := repository.NewOTPRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)

	// Capture the OTP from the mail service call.
	mailHandler, otpCh := captureMailOTP()
	mailSvc := httptest.NewServer(mailHandler)
	defer mailSvc.Close()

	userSvc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer userSvc.Close()

	registerH := handler.NewRegisterHandler(
		credRepo, otpRepo, refreshRepo, jwtSvc,
		userSvc.URL, mailSvc.URL, "test-internal-token",
		newTestLogger(t),
	)
	verifyEmailH := handler.NewVerifyEmailHandler(credRepo, otpRepo)
	loginH := handler.NewLoginHandler(credRepo, refreshRepo, jwtSvc)

	const email = "flow@example.com"
	const password = "Passw0rd!"
	const displayName = "Flow Tester"

	// ── Step 1: Register ──────────────────────────────────────────────────────
	regBody, _ := json.Marshal(map[string]string{
		"email": email, "password": password, "display_name": displayName,
	})
	regRec := httptest.NewRecorder()
	regReq := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	registerH.ServeHTTP(regRec, regReq)

	require.Equal(t, http.StatusCreated, regRec.Code, "register: %s", regRec.Body.String())

	var regResp handler.AuthResponse
	require.NoError(t, json.Unmarshal(regRec.Body.Bytes(), &regResp))
	assert.NotEmpty(t, regResp.StudentID)
	assert.NotEmpty(t, regResp.AccessToken)
	assert.NotEmpty(t, regResp.RefreshToken)
	assert.Equal(t, 3600, regResp.ExpiresIn)

	// ── Step 2: Capture OTP from mail service (sent asynchronously) ───────────
	var otpPlaintext string
	select {
	case otpPlaintext = <-otpCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for verification email OTP")
	}
	require.Len(t, otpPlaintext, 6, "OTP must be 6 digits")

	// ── Step 3: Verify email ──────────────────────────────────────────────────
	verifyBody, _ := json.Marshal(map[string]string{"email": email, "otp": otpPlaintext})
	verifyRec := httptest.NewRecorder()
	verifyReq := httptest.NewRequest(http.MethodPost, "/v1/auth/email/verify", bytes.NewReader(verifyBody))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyEmailH.ServeHTTP(verifyRec, verifyReq)

	require.Equal(t, http.StatusNoContent, verifyRec.Code, "verify-email: %s", verifyRec.Body.String())

	// ── Step 4: Login ─────────────────────────────────────────────────────────
	loginBody, _ := json.Marshal(map[string]string{"email": email, "password": password})
	loginRec := httptest.NewRecorder()
	loginReq := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginH.ServeHTTP(loginRec, loginReq)

	require.Equal(t, http.StatusOK, loginRec.Code, "login: %s", loginRec.Body.String())

	var loginResp handler.AuthResponse
	require.NoError(t, json.Unmarshal(loginRec.Body.Bytes(), &loginResp))
	assert.Equal(t, regResp.StudentID, loginResp.StudentID)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, loginResp.RefreshToken)
}
