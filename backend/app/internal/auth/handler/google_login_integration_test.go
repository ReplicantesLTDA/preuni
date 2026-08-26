package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preuni/app/internal/auth/domain"
	"github.com/preuni/app/internal/auth/handler"
	"github.com/preuni/app/internal/auth/handler/testhelper"
	"github.com/preuni/app/internal/auth/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nthQueryFailTracer deterministically fails the Nth query (1-indexed,
// counted only on the pool it's attached to) -- same technique as
// tests/integration/nth_query_fail_helper_test.go, duplicated locally so
// this package doesn't need to import a _test.go file from another
// package. See that file's docstring for why a canceled context reliably
// fails exactly one query with real pgx behavior, not a mock.
type nthQueryFailTracer struct {
	n     int64
	count int64
}

func (t *nthQueryFailTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	if atomic.AddInt64(&t.count, 1) == t.n {
		cctx, cancel := context.WithCancel(ctx)
		cancel()
		return cctx
	}
	return ctx
}

func (t *nthQueryFailTracer) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

// nthQueryFailPool opens its OWN pool (separate from testhelper.SetupTestDB's)
// whose Nth query fails -- fixtures should go through a plain SetupTestDB
// pool first so they don't count against this pool's query counter.
func nthQueryFailPool(t *testing.T, n int64) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://preuni:preuni@localhost:5432/preuni?sslmode=disable"
	}
	pgCfg, err := pgxpool.ParseConfig(dbURL)
	require.NoError(t, err)
	pgCfg.ConnConfig.Tracer = &nthQueryFailTracer{n: n}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgCfg)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(context.Background()))
	t.Cleanup(pool.Close)
	return pool
}

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

// TestIntegration_GoogleLogin_OAuthLookupFailureReturns500 covers
// ServeHTTP's `case err != nil` branch: oauthRepo.FindCredentialID failing
// for a reason other than not-found (e.g. a dropped connection) must 500,
// not silently fall through to sign-up.
func TestIntegration_GoogleLogin_OAuthLookupFailureReturns500(t *testing.T) {
	testhelper.SetupTestDB(t)
	pool := nthQueryFailPool(t, 1)

	credRepo := repository.NewCredentialsRepository(pool)
	oauthRepo := repository.NewOAuthIdentityRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)

	h := newGoogleLoginHandler(credRepo, oauthRepo, refreshRepo, jwtSvc, fakeGoogleVerifier{
		claims: domain.GoogleClaims{Subject: "google-sub-fail-1", Email: "failq1@example.com", EmailVerified: true},
	})

	rec := postGoogleLogin(t, h)
	assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
}

// TestIntegration_GoogleLogin_SignUp_FindByEmailFailureReturns500 covers
// signUp's `!apperrors.IsNotFound(err)` branch: FindByEmail failing for a
// reason other than not-found must propagate as an error, not be treated
// as "email never seen before".
func TestIntegration_GoogleLogin_SignUp_FindByEmailFailureReturns500(t *testing.T) {
	testhelper.SetupTestDB(t)
	pool := nthQueryFailPool(t, 2)

	credRepo := repository.NewCredentialsRepository(pool)
	oauthRepo := repository.NewOAuthIdentityRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)

	h := newGoogleLoginHandler(credRepo, oauthRepo, refreshRepo, jwtSvc, fakeGoogleVerifier{
		claims: domain.GoogleClaims{Subject: "google-sub-fail-2", Email: "failq2@example.com", EmailVerified: true},
	})

	rec := postGoogleLogin(t, h)
	assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
}

// TestIntegration_GoogleLogin_SignUp_LinkExistingFailureReturns500 covers
// signUp's first Link call (linking Google to an already-existing
// password credential found by email) failing.
func TestIntegration_GoogleLogin_SignUp_LinkExistingFailureReturns500(t *testing.T) {
	fixturePool := testhelper.SetupTestDB(t)
	fixtureCredRepo := repository.NewCredentialsRepository(fixturePool)
	existing, err := domain.NewCredentials("linkfail@example.com", "Passw0rd!")
	require.NoError(t, err)
	existing.ID = "00000000-0000-0000-0000-0000000000bb"
	require.NoError(t, fixtureCredRepo.Create(context.Background(), existing))

	pool := nthQueryFailPool(t, 3)
	credRepo := repository.NewCredentialsRepository(pool)
	oauthRepo := repository.NewOAuthIdentityRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)

	h := newGoogleLoginHandler(credRepo, oauthRepo, refreshRepo, jwtSvc, fakeGoogleVerifier{
		claims: domain.GoogleClaims{Subject: "google-sub-fail-3", Email: "linkfail@example.com", EmailVerified: true},
	})

	rec := postGoogleLogin(t, h)
	assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
}

// TestIntegration_GoogleLogin_SignUp_CreateFailureReturns500 covers
// signUp's credRepo.Create call failing for a brand-new email.
func TestIntegration_GoogleLogin_SignUp_CreateFailureReturns500(t *testing.T) {
	testhelper.SetupTestDB(t)
	pool := nthQueryFailPool(t, 3)

	credRepo := repository.NewCredentialsRepository(pool)
	oauthRepo := repository.NewOAuthIdentityRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)

	h := newGoogleLoginHandler(credRepo, oauthRepo, refreshRepo, jwtSvc, fakeGoogleVerifier{
		claims: domain.GoogleClaims{Subject: "google-sub-fail-4", Email: "failq4@example.com", EmailVerified: true},
	})

	rec := postGoogleLogin(t, h)
	assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
}

// TestIntegration_GoogleLogin_SignUp_LinkNewFailureReturns500 covers
// signUp's second Link call (linking Google to a brand-new credential it
// just created) failing.
func TestIntegration_GoogleLogin_SignUp_LinkNewFailureReturns500(t *testing.T) {
	testhelper.SetupTestDB(t)
	pool := nthQueryFailPool(t, 4)

	credRepo := repository.NewCredentialsRepository(pool)
	oauthRepo := repository.NewOAuthIdentityRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)

	h := newGoogleLoginHandler(credRepo, oauthRepo, refreshRepo, jwtSvc, fakeGoogleVerifier{
		claims: domain.GoogleClaims{Subject: "google-sub-fail-5", Email: "failq5@example.com", EmailVerified: true},
	})

	rec := postGoogleLogin(t, h)
	assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
}

// TestIntegration_GoogleLogin_FindByIDAfterSignUpFailureReturns500 covers
// ServeHTTP's credRepo.FindByID call (after a successful sign-up) failing.
func TestIntegration_GoogleLogin_FindByIDAfterSignUpFailureReturns500(t *testing.T) {
	testhelper.SetupTestDB(t)
	pool := nthQueryFailPool(t, 5)

	credRepo := repository.NewCredentialsRepository(pool)
	oauthRepo := repository.NewOAuthIdentityRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)

	h := newGoogleLoginHandler(credRepo, oauthRepo, refreshRepo, jwtSvc, fakeGoogleVerifier{
		claims: domain.GoogleClaims{Subject: "google-sub-fail-6", Email: "failq6@example.com", EmailVerified: true},
	})

	rec := postGoogleLogin(t, h)
	assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
}

// TestIntegration_GoogleLogin_RefreshStoreFailureReturns500 covers
// ServeHTTP's final refreshRepo.Store call failing.
func TestIntegration_GoogleLogin_RefreshStoreFailureReturns500(t *testing.T) {
	testhelper.SetupTestDB(t)
	pool := nthQueryFailPool(t, 6)

	credRepo := repository.NewCredentialsRepository(pool)
	oauthRepo := repository.NewOAuthIdentityRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService("test-signing-key-32-bytes-minimum!", 3600, 30)

	h := newGoogleLoginHandler(credRepo, oauthRepo, refreshRepo, jwtSvc, fakeGoogleVerifier{
		claims: domain.GoogleClaims{Subject: "google-sub-fail-7", Email: "failq7@example.com", EmailVerified: true},
	})

	rec := postGoogleLogin(t, h)
	assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
}
