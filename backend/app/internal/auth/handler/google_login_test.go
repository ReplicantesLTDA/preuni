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
	"github.com/stretchr/testify/assert"
)

// fakeGoogleVerifier lets tests control what domain.GoogleIDTokenVerifier
// returns without touching the network or a real Google token.
type fakeGoogleVerifier struct {
	claims domain.GoogleClaims
	err    error
}

func (f fakeGoogleVerifier) Verify(_ context.Context, _, _ string) (domain.GoogleClaims, error) {
	return f.claims, f.err
}

func TestGoogleLogin_NotConfiguredReturns500(t *testing.T) {
	h := handler.NewGoogleLoginHandler(
		fakeGoogleVerifier{}, "", // empty client ID => not configured
		nil, nil, nil, nil, nil, nil,
	)

	body, _ := json.Marshal(map[string]string{"id_token": "whatever"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/google", bytes.NewReader(body))
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
}

func TestGoogleLogin_MissingIDTokenReturns422(t *testing.T) {
	h := handler.NewGoogleLoginHandler(
		fakeGoogleVerifier{}, "configured-client-id",
		nil, nil, nil, nil, nil, nil,
	)

	body, _ := json.Marshal(map[string]string{"id_token": ""})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/google", bytes.NewReader(body))
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
}

func TestGoogleLogin_InvalidTokenReturns401(t *testing.T) {
	h := handler.NewGoogleLoginHandler(
		fakeGoogleVerifier{err: assert.AnError}, "configured-client-id",
		nil, nil, nil, nil, nil, nil,
	)

	body, _ := json.Marshal(map[string]string{"id_token": "bad-token"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/google", bytes.NewReader(body))
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
}

func TestGoogleLogin_UnverifiedEmailReturns403(t *testing.T) {
	h := handler.NewGoogleLoginHandler(
		fakeGoogleVerifier{claims: domain.GoogleClaims{
			Subject: "123", Email: "student@example.com", EmailVerified: false,
		}}, "configured-client-id",
		nil, nil, nil, nil, nil, nil,
	)

	body, _ := json.Marshal(map[string]string{"id_token": "some-token"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/google", bytes.NewReader(body))
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

func TestGoogleLogin_InvalidJSONReturns422(t *testing.T) {
	h := handler.NewGoogleLoginHandler(
		fakeGoogleVerifier{}, "configured-client-id",
		nil, nil, nil, nil, nil, nil,
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/google", bytes.NewReader([]byte("{not json")))
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
}
