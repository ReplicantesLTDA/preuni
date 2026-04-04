package domain

import (
	"strings"
	"time"

	apperrors "github.com/preuni/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// Credentials represents an auth record for a student.
type Credentials struct {
	ID               string
	Email            string
	PasswordHash     string
	EmailVerified    bool
	EmailVerifiedAt  *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewCredentials validates the email and password, hashes the password,
// and returns a ready-to-persist Credentials value.
//
// The caller is responsible for assigning ID (which should match the student ID
// created in user-svc immediately after).
func NewCredentials(email, password string) (*Credentials, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, apperrors.Validation("email", "email is required")
	}
	if !isValidEmail(email) {
		return nil, apperrors.Validation("email", "email format is invalid")
	}

	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, apperrors.Internal(err)
	}

	now := time.Now()
	return &Credentials{
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// CheckPassword verifies a plaintext password against the stored hash.
func (c *Credentials) CheckPassword(password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(c.PasswordHash), []byte(password))
	if err != nil {
		return apperrors.Unauthorized("invalid credentials")
	}
	return nil
}

// UpdatePassword re-hashes and replaces the password hash.
func (c *Credentials) UpdatePassword(newPassword string) error {
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return apperrors.Internal(err)
	}
	c.PasswordHash = string(hash)
	return nil
}

// isValidEmail performs a minimal RFC 5322-style check.
func isValidEmail(email string) bool {
	at := strings.LastIndex(email, "@")
	if at < 1 {
		return false
	}
	domain := email[at+1:]
	dot := strings.LastIndex(domain, ".")
	return dot >= 1 && dot < len(domain)-1
}
