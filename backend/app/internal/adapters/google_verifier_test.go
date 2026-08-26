package adapters

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testKid = "test-key-1"
const testAudience = "test-client-id.apps.googleusercontent.com"

// newTestVerifier spins up a fake JWKS server backed by key, and returns a
// GoogleVerifier pointed at it plus a func to sign valid test ID tokens.
func newTestVerifier(t *testing.T, key *rsa.PrivateKey) (*GoogleVerifier, func(claims jwt.MapClaims) string) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		set := googleJWKSet{Keys: []googleJWK{{
			Kid: testKid,
			N:   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
			E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes()),
		}}}
		_ = json.NewEncoder(w).Encode(set)
	}))
	t.Cleanup(srv.Close)

	v := &GoogleVerifier{certsURL: srv.URL, httpClient: srv.Client()}

	sign := func(claims jwt.MapClaims) string {
		tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tok.Header["kid"] = testKid
		s, err := tok.SignedString(key)
		require.NoError(t, err)
		return s
	}

	return v, sign
}

func validClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"iss":            "https://accounts.google.com",
		"aud":            testAudience,
		"sub":            "1234567890",
		"email":          "student@example.com",
		"email_verified": true,
		"exp":            time.Now().Add(time.Hour).Unix(),
	}
}

func TestGoogleVerifier_Verify_AcceptsValidToken(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	v, sign := newTestVerifier(t, key)

	claims, err := v.Verify(context.Background(), sign(validClaims()), testAudience)
	require.NoError(t, err)
	assert.Equal(t, "1234567890", claims.Subject)
	assert.Equal(t, "student@example.com", claims.Email)
	assert.True(t, claims.EmailVerified)
}

func TestGoogleVerifier_Verify_RejectsWrongAudience(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	v, sign := newTestVerifier(t, key)

	_, err = v.Verify(context.Background(), sign(validClaims()), "some-other-client-id")
	assert.Error(t, err)
}

func TestGoogleVerifier_Verify_RejectsWrongIssuer(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	v, sign := newTestVerifier(t, key)

	claims := validClaims()
	claims["iss"] = "https://evil.example.com"
	_, err = v.Verify(context.Background(), sign(claims), testAudience)
	assert.Error(t, err)
}

func TestGoogleVerifier_Verify_RejectsExpiredToken(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	v, sign := newTestVerifier(t, key)

	claims := validClaims()
	claims["exp"] = time.Now().Add(-time.Hour).Unix()
	_, err = v.Verify(context.Background(), sign(claims), testAudience)
	assert.Error(t, err)
}

func TestGoogleVerifier_Verify_RejectsBadSignature(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	v, sign := newTestVerifier(t, key)

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, validClaims())
	tok.Header["kid"] = testKid
	badToken, err := tok.SignedString(otherKey)
	require.NoError(t, err)
	_ = sign // keep helper referenced for symmetry with other tests

	_, err = v.Verify(context.Background(), badToken, testAudience)
	assert.Error(t, err)
}

func TestGoogleVerifier_Verify_RejectsUnverifiedEmailClaimMissing(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	v, sign := newTestVerifier(t, key)

	claims := validClaims()
	delete(claims, "email")
	_, err = v.Verify(context.Background(), sign(claims), testAudience)
	assert.Error(t, err)
}

func TestGoogleVerifier_Verify_KeysCached(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	v, sign := newTestVerifier(t, key)

	tok := sign(validClaims())
	_, err = v.Verify(context.Background(), tok, testAudience)
	require.NoError(t, err)
	require.False(t, v.fetchedAt.IsZero())

	// Second call must not error even if we now point at a dead URL —
	// proves the cached keys (not a fresh fetch) served the request.
	v.certsURL = "http://127.0.0.1:1"
	_, err = v.Verify(context.Background(), tok, testAudience)
	assert.NoError(t, err)
}
