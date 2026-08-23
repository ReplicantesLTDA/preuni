package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

const (
	otpLength = 6
	otpExpiry = 15 * time.Minute
)

// OTPPurpose enumerates the valid uses for an OTP code.
type OTPPurpose string

const (
	OTPPurposeEmailVerify   OTPPurpose = "EMAIL_VERIFY"
	OTPPurposePasswordReset OTPPurpose = "PASSWORD_RESET"
	OTPPurposeLoginOTP      OTPPurpose = "LOGIN_OTP"
	OTPPurposeEmailChange   OTPPurpose = "EMAIL_CHANGE"
)

// OTPCode is the domain model for a one-time password event.
type OTPCode struct {
	ID           string
	CredentialID string
	Purpose      OTPPurpose
	CodeHash     string
	ExpiresAt    time.Time
	UsedAt       *time.Time
	CreatedAt    time.Time
}

// GenerateOTP creates a random 6-digit code and its SHA-256 hash.
// Return the plaintext code to the caller (to send via email) and store only the hash.
func GenerateOTP() (plaintext, hash string, expiresAt time.Time, err error) {
	code, err := randomDigits(otpLength)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("generate otp: %w", err)
	}
	sum := sha256.Sum256([]byte(code))
	return code, hex.EncodeToString(sum[:]), time.Now().Add(otpExpiry), nil
}

// HashOTP returns the SHA-256 hex digest of a plaintext code.
// Use this when verifying a code submitted by a user.
func HashOTP(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

// randomDigits generates a cryptographically-random string of n decimal digits.
func randomDigits(n int) (string, error) {
	digits := make([]byte, n)
	for i := range digits {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		digits[i] = byte('0' + num.Int64())
	}
	return string(digits), nil
}
