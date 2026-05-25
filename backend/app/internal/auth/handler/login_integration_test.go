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

// TestIntegration_Login_UnverifiedEmail asserts that login with a registered
// but unverified account returns HTTP 403 Forbidden.
func TestIntegration_Login_UnverifiedEmail(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	ctx := context.Background()

	credRepo := repository.NewCredentialsRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)
	h := handler.NewLoginHandler(credRepo, refreshRepo, jwtSvc)

	// email_verified defaults to false — no extra step needed.
	creds, err := domain.NewCredentials("unverified@example.com", "Passw0rd!")
	require.NoError(t, err)
	creds.ID = "00000000-0000-0000-0000-000000000001"
	require.NoError(t, credRepo.Create(ctx, creds))

	body, _ := json.Marshal(map[string]string{
		"email": "unverified@example.com", "password": "Passw0rd!",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

// TestIntegration_Login_InvalidPassword asserts that login with a verified
// account but wrong password returns HTTP 401 Unauthorized.
func TestIntegration_Login_InvalidPassword(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	ctx := context.Background()

	credRepo := repository.NewCredentialsRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)
	h := handler.NewLoginHandler(credRepo, refreshRepo, jwtSvc)

	const email = "wrongpw@example.com"
	creds, err := domain.NewCredentials(email, "CorrectPass1!")
	require.NoError(t, err)
	creds.ID = "00000000-0000-0000-0000-000000000002"
	require.NoError(t, credRepo.Create(ctx, creds))
	_, err = pool.Exec(ctx,
		"UPDATE auth.credentials SET email_verified = true, email_verified_at = now() WHERE lower(email) = lower($1)",
		email,
	)
	require.NoError(t, err)

	body, _ := json.Marshal(map[string]string{
		"email": email, "password": "WrongPass9!",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
}

// TestIntegration_Login_Success asserts that login with a verified account and
// correct credentials returns HTTP 200 with valid tokens.
func TestIntegration_Login_Success(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	ctx := context.Background()

	credRepo := repository.NewCredentialsRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)
	h := handler.NewLoginHandler(credRepo, refreshRepo, jwtSvc)

	const email = "loginsuccess@example.com"
	const password = "Passw0rd!"

	creds, err := domain.NewCredentials(email, password)
	require.NoError(t, err)
	creds.ID = "00000000-0000-0000-0000-000000000003"
	require.NoError(t, credRepo.Create(ctx, creds))
	_, err = pool.Exec(ctx,
		"UPDATE auth.credentials SET email_verified = true, email_verified_at = now() WHERE lower(email) = lower($1)",
		email,
	)
	require.NoError(t, err)

	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var resp handler.AuthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
}
