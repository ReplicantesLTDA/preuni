package handler_test

import (
	"context"
	"sync"
	"testing"

	"github.com/preuni/pkg/logger"
	"github.com/preuni/app/internal/auth/ports"
)

// newTestLogger returns a logger that suppresses output during tests.
func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	return logger.New("error")
}

// fakeStudentProvisioner accepts every CreateStudent call without side effects.
type fakeStudentProvisioner struct{}

func (fakeStudentProvisioner) CreateStudent(_ context.Context, _ ports.CreateStudentRequest) error {
	return nil
}

// fakeEmailSender accepts every Send call without side effects.
type fakeEmailSender struct{}

func (fakeEmailSender) Send(_ context.Context, _, _ string, _ map[string]string) error {
	return nil
}

// otpCapturingEmailSender records the OTP from the first EMAIL_VERIFY call.
type otpCapturingEmailSender struct {
	mu  sync.Mutex
	ch  chan string
}

func newOTPCapturingEmailSender() *otpCapturingEmailSender {
	return &otpCapturingEmailSender{ch: make(chan string, 1)}
}

func (s *otpCapturingEmailSender) Send(_ context.Context, emailType, _ string, params map[string]string) error {
	if emailType == "EMAIL_VERIFY" {
		if otp, ok := params["otp"]; ok {
			select {
			case s.ch <- otp:
			default:
			}
		}
	}
	return nil
}

func (s *otpCapturingEmailSender) OTPChan() <-chan string { return s.ch }
