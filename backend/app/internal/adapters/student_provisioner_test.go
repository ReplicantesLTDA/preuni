package adapters

import "testing"

// TestDefaultUsername_ShortID covers defaultUsername's "id shorter than 8
// chars" branch -- every real caller passes a UUID (36 chars), so only the
// len(id)>=8 branch was ever exercised before.
func TestDefaultUsername_ShortID(t *testing.T) {
	if got := defaultUsername("abc"); got != "userabc" {
		t.Errorf("defaultUsername(%q) = %q, want %q", "abc", got, "userabc")
	}
}

// TestDefaultUsername_LongID covers the len(id)>=8 branch explicitly (the
// >8-char UUID case was already exercised indirectly via integration
// tests, but never asserted on directly).
func TestDefaultUsername_LongID(t *testing.T) {
	if got := defaultUsername("12345678-abcd"); got != "user12345678" {
		t.Errorf("defaultUsername with a long id = %q, want %q", got, "user12345678")
	}
}
