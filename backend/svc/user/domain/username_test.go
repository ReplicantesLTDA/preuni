package domain_test

import (
	"testing"

	"github.com/preuni/svc/user/domain"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		wantError bool
		errMsg    string
	}{
		// ── Valid cases ───────────────────────────────────────────────────────────
		{name: "simple letters and digits", username: "ana01", wantError: false},
		{name: "hyphen and underscore", username: "jo-se_01", wantError: false},
		{name: "exactly 3 chars", username: "abc", wantError: false},
		{name: "exactly 30 chars", username: "a" + repeat("b", 28) + "c", wantError: false},

		// ── Invalid cases ────────────────────────────────────────────────────────
		{name: "empty", username: "", wantError: true},
		{name: "too short (2 chars)", username: "ab", wantError: true, errMsg: "3 characters"},
		{name: "too long (31 chars)", username: "a" + repeat("b", 30), wantError: true, errMsg: "30 characters"},
		{name: "uppercase letter", username: "AnaUser", wantError: true},
		{name: "starts with digit", username: "1user", wantError: true},
		{name: "contains forward slash", username: "user/name", wantError: true},
		{name: "contains space", username: "user name", wantError: true},
		{name: "contains dot", username: "user.name", wantError: true},
		{name: "only digits", username: "12345", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateUsername(tt.username)
			if tt.wantError && err == nil {
				t.Errorf("ValidateUsername(%q) expected error, got nil", tt.username)
			}
			if !tt.wantError && err != nil {
				t.Errorf("ValidateUsername(%q) unexpected error: %v", tt.username, err)
			}
			if tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateUsername(%q) error %q does not contain %q", tt.username, err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func repeat(s string, n int) string {
	result := ""
	for range n {
		result += s
	}
	return result
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
