package domain

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	apperrors "github.com/preuni/pkg/errors"
)

// GoogleClaims is the subset of a verified Google ID token's claims the
// auth domain needs.
type GoogleClaims struct {
	Subject       string // Google's stable per-user "sub" claim
	Email         string
	EmailVerified bool
}

// GoogleIDTokenVerifier validates a Google-issued OpenID Connect ID token
// and returns its verified claims. The production implementation checks
// the signature against Google's published JWKS plus audience/issuer;
// tests inject a fake.
type GoogleIDTokenVerifier interface {
	Verify(ctx context.Context, idToken, audience string) (GoogleClaims, error)
}

// NewOAuthCredentials builds a Credentials row for a user signing up via an
// OAuth provider (no password of their own). The password hash is a random
// value the user never sees and can never authenticate with directly --
// only Google-issued tokens can unlock this account. Email is treated as
// pre-verified since the provider already confirmed it.
func NewOAuthCredentials(email string) (*Credentials, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, apperrors.Validation("email", "email is required")
	}
	if !isValidEmail(email) {
		return nil, apperrors.Validation("email", "email format is invalid")
	}

	randPassword, err := randomOpaquePassword()
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	hash, err := HashPassword(randPassword)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Credentials{
		Email:           email,
		PasswordHash:    hash,
		EmailVerified:   true,
		EmailVerifiedAt: &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// randomOpaquePassword returns a 32-byte cryptographically random value,
// base64-encoded, used only as bcrypt input -- never returned to a caller.
func randomOpaquePassword() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
