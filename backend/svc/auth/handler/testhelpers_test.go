package handler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/preuni/pkg/logger"
)

// newTestLogger returns a logger that suppresses output during tests.
func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	return logger.New("error")
}

// captureMailOTP returns an http.Handler and a channel that receives the
// plaintext OTP code from the first EMAIL_VERIFY request the register handler
// fires to the mail service (asynchronously).
func captureMailOTP() (http.Handler, <-chan string) {
	ch := make(chan string, 1)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return
		}

		var payload struct {
			Type   string            `json:"type"`
			Params map[string]string `json:"params"`
		}
		if err = json.Unmarshal(body, &payload); err != nil {
			return
		}
		if payload.Type == "EMAIL_VERIFY" {
			if otp, ok := payload.Params["otp"]; ok {
				select {
				case ch <- otp:
				default:
				}
			}
		}
	})
	return h, ch
}
