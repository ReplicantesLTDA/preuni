package repository

import "testing"

// contains/join/itoa are small internal string helpers used to build
// dynamic UPDATE statements. Their not-found/empty branches were
// previously untested -- Update() always calls them with content that
// hits the found/non-empty path.
func TestContains_NotFoundReturnsFalse(t *testing.T) {
	if contains("duplicate value violates something else", "unique constraint") {
		t.Fatal("expected contains to return false when the substring is absent")
	}
	if contains("short", "much longer substring") {
		t.Fatal("expected contains to return false when sub is longer than s")
	}
}

func TestContains_FoundReturnsTrue(t *testing.T) {
	if !contains("duplicate key value violates unique constraint", "duplicate key") {
		t.Fatal("expected contains to find a present substring")
	}
}

func TestJoin_EmptyPartsReturnsEmptyString(t *testing.T) {
	if got := join(nil, ", "); got != "" {
		t.Fatalf("expected an empty string for no parts, got %q", got)
	}
	if got := join([]string{}, ", "); got != "" {
		t.Fatalf("expected an empty string for zero parts, got %q", got)
	}
}

func TestJoin_JoinsMultipleParts(t *testing.T) {
	if got := join([]string{"a", "b", "c"}, ", "); got != "a, b, c" {
		t.Fatalf("got %q", got)
	}
}

func TestItoa_Zero(t *testing.T) {
	if got := itoa(0); got != "0" {
		t.Fatalf("got %q", got)
	}
}
