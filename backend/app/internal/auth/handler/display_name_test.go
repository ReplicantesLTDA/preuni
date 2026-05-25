package handler

import (
	"strings"
	"testing"
)

func TestValidateDisplayName(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", true},
		{"whitespace only", "   ", true},
		{"plain", "Daniel", false},
		{"accented", "Ana Lima", false},
		{"emoji", "Ana 🚀", false},
		{"64 printable runes ok", strings.Repeat("a", 64), false},
		{"65 runes too long", strings.Repeat("a", 65), true},
		{"crlf injection", "Daniel\r\nBcc: x@y.z", true},
		{"lf only", "Daniel\nBcc", true},
		{"cr only", "Daniel\rBcc", true},
		{"embedded tab", "Dan\tiel", true},
		{"nul", "Daniel\x00", true},
		{"backspace", "Daniel\x08", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateDisplayName(c.input)
			if c.wantErr && err == nil {
				t.Fatalf("expected error for %q, got nil", c.input)
			}
			if !c.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", c.input, err)
			}
		})
	}
}
