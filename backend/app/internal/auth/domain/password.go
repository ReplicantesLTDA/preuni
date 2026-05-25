package domain

import (
	"unicode"

	apperrors "github.com/preuni/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const bcryptHashCost = 12

// HashPassword hashes the given plaintext password using bcrypt (cost 12).
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptHashCost)
	if err != nil {
		return "", apperrors.Internal(err)
	}
	return string(hash), nil
}

const (
	passwordMinLength = 6
	passwordMaxLength = 255
)

// ValidatePassword checks that the password meets the minimum security requirements:
//   - Length: 6–255 characters
//   - At least one uppercase letter
//   - At least one lowercase letter
//   - At least one digit
//
// Returns a *AppError with CodeValidation if any rule is violated.
func ValidatePassword(password string) error {
	runes := []rune(password)
	length := len(runes)

	if length < passwordMinLength {
		return apperrors.Validation("password",
			"password must be at least 6 characters long")
	}
	if length > passwordMaxLength {
		return apperrors.Validation("password",
			"password must be at most 255 characters long")
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range runes {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasUpper {
		return apperrors.Validation("password",
			"password must contain at least one uppercase letter")
	}
	if !hasLower {
		return apperrors.Validation("password",
			"password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return apperrors.Validation("password",
			"password must contain at least one digit")
	}

	return nil
}
