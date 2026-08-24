// Package errors defines the typed error catalogue for all preuni services.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Code represents a machine-readable error category.
type Code string

const (
	CodeNotFound      Code = "NOT_FOUND"
	CodeConflict      Code = "CONFLICT"
	CodeUnauthorized  Code = "UNAUTHORIZED"
	CodeForbidden     Code = "FORBIDDEN"
	CodeValidation    Code = "VALIDATION_ERROR"
	CodeInternal      Code = "INTERNAL_ERROR"
	CodeQuotaExceeded Code = "QUOTA_EXCEEDED"
)

// AppError is the standard error type for all service boundaries.
type AppError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	// Field is set for validation errors to identify the offending field.
	Field string `json:"field,omitempty"`
	// Cause is an optional wrapped error for internal diagnostics (never serialised).
	Cause error `json:"-"`
}

func (e *AppError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Cause }

// HTTPStatus maps an AppError to the appropriate HTTP status code.
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeValidation:
		return http.StatusUnprocessableEntity
	case CodeQuotaExceeded:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// ── Constructors ──────────────────────────────────────────────────────────────

func NotFound(resource string) *AppError {
	return &AppError{Code: CodeNotFound, Message: resource + " not found"}
}

func Conflict(message string) *AppError {
	return &AppError{Code: CodeConflict, Message: message}
}

func Unauthorized(message string) *AppError {
	return &AppError{Code: CodeUnauthorized, Message: message}
}

func Forbidden(message string) *AppError {
	return &AppError{Code: CodeForbidden, Message: message}
}

func Validation(field, message string) *AppError {
	return &AppError{Code: CodeValidation, Message: message, Field: field}
}

func Internal(cause error) *AppError {
	return &AppError{Code: CodeInternal, Message: "an internal error occurred", Cause: cause}
}

func QuotaExceeded(message string) *AppError {
	return &AppError{Code: CodeQuotaExceeded, Message: message}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// As unwraps an error into *AppError. Returns nil if the chain contains no AppError.
func As(err error) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return nil
}

// IsNotFound reports whether err is a CodeNotFound AppError.
func IsNotFound(err error) bool {
	ae := As(err)
	return ae != nil && ae.Code == CodeNotFound
}

// IsConflict reports whether err is a CodeConflict AppError.
func IsConflict(err error) bool {
	ae := As(err)
	return ae != nil && ae.Code == CodeConflict
}
