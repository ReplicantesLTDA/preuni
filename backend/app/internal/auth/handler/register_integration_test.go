package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/preuni/app/internal/auth/adapters"
	"github.com/preuni/app/internal/auth/domain"
	"github.com/preuni/app/internal/auth/handler"
	"github.com/preuni/app/internal/auth/handler/testhelper"
	"github.com/preuni/app/internal/auth/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_Register_HappyPath asserts that a valid registration returns
// HTTP 201 with all AuthResponse fields populated.
func TestIntegration_Register_HappyPath(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	ctx := context.Background()
	_ = ctx

	credRepo := repository.NewCredentialsRepository(pool)
	otpRepo := repository.NewOTPRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)

	mailSvc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mailSvc.Close()

	userSvc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer userSvc.Close()

	h := handler.NewRegisterHandler(
		credRepo, otpRepo, refreshRepo, jwtSvc,
		adapters.NewHTTPStudentProvisioner(userSvc.URL, "tok"),
		adapters.NewHTTPEmailSender(mailSvc.URL, "tok"),
		newTestLogger(t),
	)

	body, _ := json.Marshal(map[string]string{
		"email": "newuser@example.com", "password": "Passw0rd!", "display_name": "New User",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var resp handler.AuthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.StudentID)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Greater(t, resp.ExpiresIn, 0)
}

// TestIntegration_Register_DuplicateEmail asserts that a second registration
// with the same email returns HTTP 409 Conflict.
func TestIntegration_Register_DuplicateEmail(t *testing.T) {
	pool := testhelper.SetupTestDB(t)

	credRepo := repository.NewCredentialsRepository(pool)
	otpRepo := repository.NewOTPRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)

	mailSvc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mailSvc.Close()

	userSvc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer userSvc.Close()

	h := handler.NewRegisterHandler(
		credRepo, otpRepo, refreshRepo, jwtSvc,
		adapters.NewHTTPStudentProvisioner(userSvc.URL, "tok"),
		adapters.NewHTTPEmailSender(mailSvc.URL, "tok"),
		newTestLogger(t),
	)

	// First registration — must succeed.
	body, _ := json.Marshal(map[string]string{
		"email": "dup@example.com", "password": "Passw0rd!", "display_name": "Dup User",
	})
	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(body)))
	require.Equal(t, http.StatusCreated, rec1.Code, rec1.Body.String())

	// Second registration with the same email — must conflict.
	body, _ = json.Marshal(map[string]string{
		"email": "dup@example.com", "password": "Different1!", "display_name": "Dup Again",
	})
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(body)))
	assert.Equal(t, http.StatusConflict, rec2.Code, rec2.Body.String())
}
