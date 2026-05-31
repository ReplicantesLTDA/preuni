package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	apperrors "github.com/preuni/pkg/errors"
	"github.com/preuni/pkg/logger"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/app/internal/auth/domain"
	"github.com/preuni/app/internal/auth/ports"
	"github.com/preuni/app/internal/auth/repository"
)

// ChangeEmailRequestRequest is the body for POST /auth/email/change-request.
type ChangeEmailRequestRequest struct {
	NewEmail string `json:"new_email"`
}

// ChangeEmailConfirmRequest is the body for POST /auth/email/change-confirm.
type ChangeEmailConfirmRequest struct {
	NewEmail string `json:"new_email"`
	OTP      string `json:"otp"`
}

// ChangeEmailRequestHandler handles POST /auth/email/change-request.
// Generates an OTP and sends it to the NEW email address for verification.
type ChangeEmailRequestHandler struct {
	credRepo    *repository.CredentialsRepository
	otpRepo     *repository.OTPRepository
	emailSender ports.EmailSender
	log         *logger.Logger
}

// NewChangeEmailRequestHandler constructs the handler.
func NewChangeEmailRequestHandler(credRepo *repository.CredentialsRepository, otpRepo *repository.OTPRepository, emailSender ports.EmailSender, log *logger.Logger) *ChangeEmailRequestHandler {
	return &ChangeEmailRequestHandler{credRepo: credRepo, otpRepo: otpRepo, emailSender: emailSender, log: log}
}

func (h *ChangeEmailRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req ChangeEmailRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NewEmail == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("new_email", "new_email is required"))
		return
	}

	userID := pkgmw.UserIDFromContext(r.Context())

	otpCode, otpHash, otpExpiry, err := domain.GenerateOTP()
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	// Store OTP with purpose EMAIL_CHANGE referencing the credential ID
	if err = h.otpRepo.Create(r.Context(), userID, domain.OTPPurposeEmailChange, otpHash, otpExpiry); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	go func() {
		if err := h.emailSender.Send(context.Background(), "EMAIL_CHANGE", req.NewEmail, map[string]string{"otp": otpCode}); err != nil {
			h.log.Error("email change send failed", logger.Err(err))
		}
	}()

	w.WriteHeader(http.StatusAccepted)
}

// ChangeEmailConfirmHandler handles POST /auth/email/change-confirm.
type ChangeEmailConfirmHandler struct {
	credRepo    *repository.CredentialsRepository
	otpRepo     *repository.OTPRepository
	refreshRepo *repository.RefreshTokenRepository
}

// NewChangeEmailConfirmHandler constructs the handler.
func NewChangeEmailConfirmHandler(credRepo *repository.CredentialsRepository, otpRepo *repository.OTPRepository, refreshRepo *repository.RefreshTokenRepository) *ChangeEmailConfirmHandler {
	return &ChangeEmailConfirmHandler{credRepo: credRepo, otpRepo: otpRepo, refreshRepo: refreshRepo}
}

func (h *ChangeEmailConfirmHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req ChangeEmailConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NewEmail == "" || req.OTP == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("body", "new_email and otp are required"))
		return
	}

	userID := pkgmw.UserIDFromContext(r.Context())

	codeHash := domain.HashOTP(req.OTP)
	otpRecord, err := h.otpRepo.FindActiveByCredentialAndPurpose(r.Context(), userID, domain.OTPPurposeEmailChange)
	if err != nil || otpRecord.CodeHash != codeHash {
		pkgmw.ErrorResponse(w, apperrors.Validation("otp", "invalid or expired code"))
		return
	}
	if otpRecord.ExpiresAt.Before(time.Now()) {
		pkgmw.ErrorResponse(w, apperrors.Validation("otp", "invalid or expired code"))
		return
	}

	if err = h.otpRepo.MarkUsed(r.Context(), otpRecord.ID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	if err = h.credRepo.UpdateEmail(r.Context(), userID, req.NewEmail); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	// Revoke all refresh tokens — user must re-auth with new email
	if err = h.refreshRepo.RevokeAllForCredential(r.Context(), userID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
