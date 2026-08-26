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

func newGoogleLoginHandler(
	credRepo *repository.CredentialsRepository, oauthRepo *repository.OAuthIdentityRepository,
	refreshRepo *repository.RefreshTokenRepository, jwtSvc *domain.JWTService, verifier fakeGoogleVerifier,
) *handler.GoogleLoginHandler {
	return handler.NewGoogleLoginHandler(
		verifier, "configured-client-id",
		credRepo, oauthRepo, refreshRepo, jwtSvc, fakeStudentProvisioner{}, nil,
	)
}

func postGoogleLogin(t *testing.T, h *handler.GoogleLoginHandler) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"id_token": "irrelevant-fake-verifier-ignores-it"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/google", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	return rec
}

// TestIntegration_GoogleLogin_NewEmail_CreatesAccount covers the first-ever
// sign-in for an email Google has never told us about: a new credential +
// oauth_identities link must be created, and a normal AuthResponse returned.
func TestIntegration_GoogleLogin_NewEmail_CreatesAccount(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	ctx := context.Background()

	credRepo := repository.NewCredentialsRepository(pool)
	oauthRepo := repository.NewOAuthIdentityRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)

	h := newGoogleLoginHandler(credRepo, oauthRepo, refreshRepo, jwtSvc, fakeGoogleVerifier{
		claims: domain.GoogleClaims{Subject: "google-sub-1", Email: "newviagoogle@example.com", EmailVerified: true},
	})

	rec := postGoogleLogin(t, h)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var resp handler.AuthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)

	creds, err := credRepo.FindByEmail(ctx, "newviagoogle@example.com")
	require.NoError(t, err)
	assert.True(t, creds.EmailVerified)
	assert.Equal(t, resp.StudentID, creds.ID)

	linkedID, err := oauthRepo.FindCredentialID(ctx, "google", "google-sub-1")
	require.NoError(t, err)
	assert.Equal(t, creds.ID, linkedID)
}

// TestIntegration_GoogleLogin_ExistingLinkedIdentity_ReturnsSameAccount
// covers the returning-user path: a second sign-in with the same Google sub
// must reuse the already-linked credential, not create a duplicate.
func TestIntegration_GoogleLogin_ExistingLinkedIdentity_ReturnsSameAccount(t *testing.T) {
	pool := testhelper.SetupTestDB(t)

	credRepo := repository.NewCredentialsRepository(pool)
	oauthRepo := repository.NewOAuthIdentityRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)

	verifier := fakeGoogleVerifier{
		claims: domain.GoogleClaims{Subject: "google-sub-2", Email: "returning@example.com", EmailVerified: true},
	}
	h := newGoogleLoginHandler(credRepo, oauthRepo, refreshRepo, jwtSvc, verifier)

	first := postGoogleLogin(t, h)
	require.Equal(t, http.StatusOK, first.Code, first.Body.String())
	var firstResp handler.AuthResponse
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &firstResp))

	second := postGoogleLogin(t, h)
	require.Equal(t, http.StatusOK, second.Code, second.Body.String())
	var secondResp handler.AuthResponse
	require.NoError(t, json.Unmarshal(second.Body.Bytes(), &secondResp))

	assert.Equal(t, firstResp.StudentID, secondResp.StudentID)
}

// TestIntegration_GoogleLogin_ExistingEmailPasswordAccount_LinksInstead
// covers a user who registered with email/password first, then signs in
// with Google using the same address: it must link to the existing
// credential rather than erroring or creating a duplicate.
func TestIntegration_GoogleLogin_ExistingEmailPasswordAccount_LinksInstead(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	ctx := context.Background()

	credRepo := repository.NewCredentialsRepository(pool)
	oauthRepo := repository.NewOAuthIdentityRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)

	existing, err := domain.NewCredentials("linkme@example.com", "Passw0rd!")
	require.NoError(t, err)
	existing.ID = "00000000-0000-0000-0000-0000000000aa"
	require.NoError(t, credRepo.Create(ctx, existing))

	h := newGoogleLoginHandler(credRepo, oauthRepo, refreshRepo, jwtSvc, fakeGoogleVerifier{
		claims: domain.GoogleClaims{Subject: "google-sub-3", Email: "linkme@example.com", EmailVerified: true},
	})

	rec := postGoogleLogin(t, h)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var resp handler.AuthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, existing.ID, resp.StudentID)

	linkedID, err := oauthRepo.FindCredentialID(ctx, "google", "google-sub-3")
	require.NoError(t, err)
	assert.Equal(t, existing.ID, linkedID)
}
