package domain_test

import (
	"strings"
	"testing"

	"github.com/preuni/app/internal/auth/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCredentials_ValidInputSucceeds(t *testing.T) {
	creds, err := domain.NewCredentials("  User@Example.com  ", "StrongPass1")
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", creds.Email, "email should be lowercased and trimmed")
	assert.NotEmpty(t, creds.PasswordHash)
}

func TestNewCredentials_EmptyEmailIsRejected(t *testing.T) {
	_, err := domain.NewCredentials("   ", "StrongPass1")
	assert.Error(t, err)
}

func TestNewCredentials_InvalidEmailFormatIsRejected(t *testing.T) {
	for _, email := range []string{"not-an-email", "missing-domain@", "@nodomain.com", "user@nodot"} {
		_, err := domain.NewCredentials(email, "StrongPass1")
		assert.Error(t, err, "expected %q to be rejected", email)
	}
}

func TestNewCredentials_WeakPasswordIsRejected(t *testing.T) {
	_, err := domain.NewCredentials("user@example.com", "weak")
	assert.Error(t, err)
}

func TestHashPassword_ValidPasswordSucceeds(t *testing.T) {
	hash, err := domain.HashPassword("StrongPass1")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestHashPassword_OverBcryptLimitReturnsAnError(t *testing.T) {
	// bcrypt caps input at 72 bytes; ValidatePassword allows up to 255
	// characters, so a password in that gap passes validation but fails
	// hashing -- HashPassword's error branch, previously untested.
	over72Bytes := "Aa1" + strings.Repeat("x", 100)
	_, err := domain.HashPassword(over72Bytes)
	assert.Error(t, err)
}
