package mail

import (
	"strings"
	"testing"
)

func TestSanitizeHeader_StripsCRLFAndNUL(t *testing.T) {
	cases := map[string]string{
		"plain":             "plain",
		"with\r\nbcc":       "withbcc",
		"with\rnewline":     "withnewline",
		"with\nnewline":     "withnewline",
		"with\x00nul":       "withnul",
		"Daniel\r\nBcc: a": "DanielBcc: a",
	}
	for in, want := range cases {
		if got := sanitizeHeader(in); got != want {
			t.Errorf("sanitizeHeader(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEncodeMIME_HeaderInjectionStripped(t *testing.T) {
	msg := Message{
		From: "noreply@preuni.com.br", FromName: "PreUni",
		To:       "victim@example.com\r\nBcc: attacker@evil.example",
		Subject:  "Hello\r\nBcc: pwn@evil.example",
		HTMLBody: "<p>hi</p>",
		TextBody: "hi",
	}
	raw := encodeMIME(msg)
	headerSection := strings.SplitN(raw, "\r\n\r\n", 2)[0]
	// CRLF stripped → no new header line starts with "Bcc:".
	for _, line := range strings.Split(headerSection, "\r\n") {
		if strings.HasPrefix(line, "Bcc:") {
			t.Fatalf("Bcc header was injected as its own line: %q\nFull headers:\n%s", line, headerSection)
		}
	}
	// Attacker-controlled bytes survive as part of the legitimate header
	// value (concatenated, not split). That is the intended sanitization
	// behavior — the SMTP server sees one To value, one Subject value.
	if !strings.Contains(headerSection, "To: victim@example.comBcc: attacker@evil.example") {
		t.Errorf("To header not sanitized as expected; got:\n%s", headerSection)
	}
}
