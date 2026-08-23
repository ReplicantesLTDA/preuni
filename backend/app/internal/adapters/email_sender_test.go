package adapters

import (
	"context"
	"errors"
	"testing"

	"github.com/preuni/app/internal/mail"
)

type fakeSender struct {
	sent    []mail.Message
	sendErr error
}

func (f *fakeSender) Send(msg mail.Message) error {
	if f.sendErr != nil {
		return f.sendErr
	}
	f.sent = append(f.sent, msg)
	return nil
}

// TestInProcessEmailSender_Send_Succeeds covers the happy path: validate,
// build, and forward to the underlying Sender.
func TestInProcessEmailSender_Send_Succeeds(t *testing.T) {
	fake := &fakeSender{}
	sender := NewInProcessEmailSender("noreply@preuni.com.br", "PreUni", fake, nil)

	err := sender.Send(context.Background(), string(mail.TypeEmailVerify), "user@example.com", map[string]string{"otp": "123456"})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(fake.sent) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(fake.sent))
	}
	if fake.sent[0].To != "user@example.com" {
		t.Errorf("To = %q", fake.sent[0].To)
	}
}

// TestInProcessEmailSender_Send_ValidationFailure covers mail.Validate's
// error branch (missing a required param for the email type).
func TestInProcessEmailSender_Send_ValidationFailure(t *testing.T) {
	fake := &fakeSender{}
	sender := NewInProcessEmailSender("noreply@preuni.com.br", "PreUni", fake, nil)

	err := sender.Send(context.Background(), string(mail.TypeEmailVerify), "user@example.com", nil)
	if err == nil {
		t.Fatal("expected an error for a request missing the required otp param")
	}
	if len(fake.sent) != 0 {
		t.Error("expected no message to reach the underlying Sender on validation failure")
	}
}

// TestInProcessEmailSender_Send_UnderlyingSenderFailure covers the
// Sender.Send-fails branch.
func TestInProcessEmailSender_Send_UnderlyingSenderFailure(t *testing.T) {
	fake := &fakeSender{sendErr: errors.New("smtp unavailable")}
	sender := NewInProcessEmailSender("noreply@preuni.com.br", "PreUni", fake, nil)

	err := sender.Send(context.Background(), string(mail.TypeEmailVerify), "user@example.com", map[string]string{"otp": "123456"})
	if err == nil {
		t.Fatal("expected the underlying Sender's error to propagate")
	}
}
