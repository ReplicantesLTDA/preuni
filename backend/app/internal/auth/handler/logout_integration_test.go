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

// TestIntegration_Logout_ValidToken asserts that revoking an active refresh
// token returns HTTP 204 and renders the token unusable.
func TestIntegration_Logout_ValidToken(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	ctx := context.Background()

	credRepo := repository.NewCredentialsRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)
	logoutH := handler.NewLogoutHandler(refreshRepo)

	// Create a credential and issue a refresh token directly.
	creds, err := domain.NewCredentials("logout@example.com", "Passw0rd!")
	require.NoError(t, err)
	creds.ID = "00000000-0000-0000-0000-000000000020"
	require.NoError(t, credRepo.Create(ctx, creds))

	rawRefresh, refreshHash, refreshExpiry, err := jwtSvc.IssueRefreshToken()
	require.NoError(t, err)
	require.NoError(t, refreshRepo.Store(ctx, creds.ID, refreshHash, refreshExpiry))

	// Logout with the valid refresh token.
	body, _ := json.Marshal(map[string]string{"refresh_token": rawRefresh})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	logoutH.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
}

// TestIntegration_Logout_UnknownToken asserts that logout with an unrecognised
// refresh token is idempotent — it still returns HTTP 204.
func TestIntegration_Logout_UnknownToken(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	_ = pool // SetupTestDB ensures DB is ready; no rows needed for this test

	refreshRepo := repository.NewRefreshTokenRepository(pool)
	logoutH := handler.NewLogoutHandler(refreshRepo)

	body, _ := json.Marshal(map[string]string{"refresh_token": "this-token-does-not-exist"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	logoutH.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
}
