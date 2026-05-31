package domain

import (
	"regexp"

	apperrors "github.com/preuni/pkg/errors"
)

// USERNAME_REGEX: lowercase letters, digits, hyphen and underscore only;
// must start with a letter; must end with a letter or digit; length 3-30.
var usernameRegex = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,28}[a-z0-9]$`)

// ValidateUsername validates a username according to the PreUni rules.
// Returns a typed ValidationError if invalid, nil if valid.
func ValidateUsername(username string) error {
	if username == "" {
		return apperrors.Validation("username", "Username is required")
	}
	if len(username) < 3 {
		return apperrors.Validation("username", "Username must be at least 3 characters")
	}
	if len(username) > 30 {
		return apperrors.Validation("username", "Username must be at most 30 characters")
	}
	if !usernameRegex.MatchString(username) {
		return apperrors.Validation(
			"username",
			"Username may only contain lowercase letters, digits, - and _; must start with a letter",
		)
	}
	return nil
}
