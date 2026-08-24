package router

import (
	"testing"

	"github.com/preuni/app/internal/config"
	"github.com/preuni/app/internal/mail"
	"github.com/preuni/pkg/logger"
)

// TestBuildSender_NoSMTPHostReturnsNoop covers buildSender's "SMTP not
// configured" branch (75% before), previously untested.
func TestBuildSender_NoSMTPHostReturnsNoop(t *testing.T) {
	cfg := config.Config{SMTPHost: ""}
	log := logger.New("error")

	sender := buildSender(cfg, log)

	if _, ok := sender.(mail.NoopSender); !ok {
		t.Fatalf("expected a mail.NoopSender when SMTPHost is empty, got %T", sender)
	}
}

// TestBuildSender_WithSMTPHostReturnsSMTPSender covers buildSender's
// "SMTP configured" branch, previously untested.
func TestBuildSender_WithSMTPHostReturnsSMTPSender(t *testing.T) {
	cfg := config.Config{
		SMTPHost:   "smtp.example.com",
		SMTPPort:   587,
		SMTPUser:   "user",
		SMTPPass:   "pass",
		SMTPUseTLS: false,
	}
	log := logger.New("error")

	sender := buildSender(cfg, log)

	if _, ok := sender.(*mail.SMTPSender); !ok {
		t.Fatalf("expected a *mail.SMTPSender when SMTPHost is set, got %T", sender)
	}
}
