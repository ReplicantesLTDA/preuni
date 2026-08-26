package adapters

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/preuni/app/internal/auth/domain"
)

const googleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"

var googleIssuers = map[string]bool{
	"accounts.google.com":         true,
	"https://accounts.google.com": true,
}

type googleJWK struct {
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type googleJWKSet struct {
	Keys []googleJWK `json:"keys"`
}

// GoogleVerifier verifies Google-issued OpenID Connect ID tokens against
// Google's published JWKS (fetched over HTTP, no google-api-go-client
// dependency needed). Keys are cached for an hour.
type GoogleVerifier struct {
	certsURL   string
	httpClient *http.Client

	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
}

// NewGoogleVerifier constructs a GoogleVerifier that fetches keys from
// Google's real certs endpoint.
func NewGoogleVerifier() *GoogleVerifier {
	return &GoogleVerifier{
		certsURL:   googleCertsURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// Verify implements domain.GoogleIDTokenVerifier.
func (v *GoogleVerifier) Verify(ctx context.Context, idToken, audience string) (domain.GoogleClaims, error) {
	keys, err := v.getKeys(ctx)
	if err != nil {
		return domain.GoogleClaims{}, fmt.Errorf("google verifier: fetching keys: %w", err)
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(idToken, claims, func(t *jwt.Token) (interface{}, error) {
		kid, _ := t.Header["kid"].(string)
		key, ok := keys[kid]
		if !ok {
			return nil, fmt.Errorf("unknown key id %q", kid)
		}
		return key, nil
	}, jwt.WithValidMethods([]string{"RS256"}))
	if err != nil || !token.Valid {
		return domain.GoogleClaims{}, fmt.Errorf("google verifier: invalid token: %w", err)
	}

	aud, _ := claims["aud"].(string)
	if aud != audience {
		return domain.GoogleClaims{}, fmt.Errorf("google verifier: audience mismatch")
	}
	iss, _ := claims["iss"].(string)
	if !googleIssuers[iss] {
		return domain.GoogleClaims{}, fmt.Errorf("google verifier: unexpected issuer %q", iss)
	}

	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	emailVerified, _ := claims["email_verified"].(bool)
	if sub == "" || email == "" {
		return domain.GoogleClaims{}, fmt.Errorf("google verifier: missing sub or email claim")
	}

	return domain.GoogleClaims{Subject: sub, Email: email, EmailVerified: emailVerified}, nil
}

func (v *GoogleVerifier) getKeys(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	v.mu.RLock()
	if time.Since(v.fetchedAt) < time.Hour && len(v.keys) > 0 {
		keys := v.keys
		v.mu.RUnlock()
		return keys, nil
	}
	v.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.certsURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching google certs", resp.StatusCode)
	}

	var set googleJWKSet
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
		return nil, err
	}

	keys := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, k := range set.Keys {
		pub, err := googleJWKToRSAPublicKey(k)
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}

	v.mu.Lock()
	v.keys = keys
	v.fetchedAt = time.Now()
	v.mu.Unlock()

	return keys, nil
}

func googleJWKToRSAPublicKey(k googleJWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, err
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}, nil
}
