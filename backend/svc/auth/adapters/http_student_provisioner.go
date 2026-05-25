package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/preuni/svc/auth/ports"
)

// HTTPStudentProvisioner calls POST /internal/students on a remote user-svc
// (or the monolith's self-loopback URL).
type HTTPStudentProvisioner struct {
	BaseURL       string
	InternalToken string
	HTTPClient    *http.Client
}

// NewHTTPStudentProvisioner returns a provisioner configured for split-service mode.
func NewHTTPStudentProvisioner(baseURL, internalToken string) *HTTPStudentProvisioner {
	return &HTTPStudentProvisioner{
		BaseURL:       baseURL,
		InternalToken: internalToken,
		HTTPClient:    http.DefaultClient,
	}
}

// CreateStudent implements ports.StudentProvisioner.
func (a *HTTPStudentProvisioner) CreateStudent(ctx context.Context, req ports.CreateStudentRequest) error {
	body, _ := json.Marshal(map[string]string{
		"student_id":   req.StudentID,
		"display_name": req.DisplayName,
		"username":     req.Username,
		"email":        req.Email,
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.BaseURL+"/internal/students", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.InternalToken)
	resp, err := a.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("user-svc returned %d", resp.StatusCode)
	}
	return nil
}
