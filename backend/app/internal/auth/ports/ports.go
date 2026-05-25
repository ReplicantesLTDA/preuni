// Package ports declares the cross-service collaborators auth-svc depends on.
// Concrete implementations live in adapters/ (split-service mode, HTTP) and
// in the monolith's internal/adapters/ (monolith mode, in-process).
package ports

import "context"

// StudentProvisioner creates the student profile row associated with a
// freshly registered credential. In split-service mode this is an HTTP call
// to user-svc's POST /internal/students; in monolith mode it invokes the
// user repository directly.
type StudentProvisioner interface {
	CreateStudent(ctx context.Context, req CreateStudentRequest) error
}

// CreateStudentRequest mirrors the POST /internal/students body.
type CreateStudentRequest struct {
	StudentID   string
	DisplayName string
	Username    string
	Email       string
}

// EmailSender enqueues a transactional email. In split-service mode it
// POSTs to mail-svc's /internal/email/send; in monolith mode it invokes
// the in-process mail handler.
//
// Implementations MUST be safe for fire-and-forget use (no panic).
type EmailSender interface {
	Send(ctx context.Context, emailType string, to string, params map[string]string) error
}
