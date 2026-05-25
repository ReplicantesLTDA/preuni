package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// HTTPEmailSender calls POST /internal/email/send on a remote mail-svc (or
// the monolith's self-loopback URL).
type HTTPEmailSender struct {
	BaseURL       string
	InternalToken string
	HTTPClient    *http.Client
}

// NewHTTPEmailSender returns a sender configured for split-service mode.
func NewHTTPEmailSender(baseURL, internalToken string) *HTTPEmailSender {
	return &HTTPEmailSender{
		BaseURL:       baseURL,
		InternalToken: internalToken,
		HTTPClient:    http.DefaultClient,
	}
}

// Send implements ports.EmailSender.
func (a *HTTPEmailSender) Send(ctx context.Context, emailType, to string, params map[string]string) error {
	body, _ := json.Marshal(map[string]any{
		"type":   emailType,
		"to":     to,
		"params": params,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.BaseURL+"/internal/email/send", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.InternalToken)
	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("mail-svc returned %d", resp.StatusCode)
	}
	return nil
}
