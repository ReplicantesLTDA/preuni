package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOAuthCredentials_Succeeds(t *testing.T) {
	creds, err := NewOAuthCredentials("Student@Example.com")
	require.NoError(t, err)
	assert.Equal(t, "student@example.com", creds.Email)
	assert.True(t, creds.EmailVerified)
	assert.NotNil(t, creds.EmailVerifiedAt)
	assert.NotEmpty(t, creds.PasswordHash)

	// The generated password hash must never validate against an empty or
	// guessable password — it exists only so the column's NOT NULL
	// constraint is satisfied.
	assert.Error(t, creds.CheckPassword(""))
	assert.Error(t, creds.CheckPassword("password"))
}

func TestNewOAuthCredentials_ErrorsOnInvalidEmail(t *testing.T) {
	_, err := NewOAuthCredentials("not-an-email")
	assert.Error(t, err)
}

func TestNewOAuthCredentials_ErrorsOnEmptyEmail(t *testing.T) {
	_, err := NewOAuthCredentials("   ")
	assert.Error(t, err)
}
