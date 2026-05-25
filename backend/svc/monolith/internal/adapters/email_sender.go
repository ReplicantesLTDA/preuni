package adapters

import (
	"context"
	"fmt"

	"github.com/preuni/pkg/logger"
	"github.com/preuni/svc/monolith/internal/mail"
)

// InProcessEmailSender hands an email to the in-process mail pipeline
// (validate → build → SMTP send) without going through HTTP.
type InProcessEmailSender struct {
	FromEmail string
	FromName  string
	Sender    mail.Sender
	Log       *logger.Logger
}

// NewInProcessEmailSender constructs an InProcessEmailSender.
func NewInProcessEmailSender(fromEmail, fromName string, sender mail.Sender, log *logger.Logger) *InProcessEmailSender {
	return &InProcessEmailSender{FromEmail: fromEmail, FromName: fromName, Sender: sender, Log: log}
}

// Send implements ports.EmailSender. Runs synchronously; callers wrap in a
// goroutine for fire-and-forget semantics.
func (a *InProcessEmailSender) Send(_ context.Context, emailType, to string, params map[string]string) error {
	req := mail.SendRequest{
		Type:   mail.EmailType(emailType),
		To:     to,
		Params: params,
	}
	if err := mail.Validate(req); err != nil {
		return fmt.Errorf("mail validate: %w", err)
	}
	msg, err := mail.Build(req, a.FromEmail, a.FromName)
	if err != nil {
		return fmt.Errorf("mail build: %w", err)
	}
	if err := a.Sender.Send(msg); err != nil {
		return fmt.Errorf("mail send: %w", err)
	}
	return nil
}
