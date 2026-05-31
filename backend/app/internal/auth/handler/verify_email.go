package handler

import (
	"context"
	"encoding/json"
	"net/http"

	apperrors "github.com/preuni/pkg/errors"
	"github.com/preuni/pkg/logger"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/app/internal/auth/domain"
	"github.com/preuni/app/internal/auth/ports"
	"github.com/preuni/app/internal/auth/repository"
)

// VerifyEmailRequest is the JSON body for POST /auth/email/verify.
type VerifyEmailRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

// VerifyEmailHandler handles POST /auth/email/verify.
type VerifyEmailHandler struct {
	credRepo *repository.CredentialsRepository
	otpRepo  *repository.OTPRepository
}

// NewVerifyEmailHandler constructs a VerifyEmailHandler.
func NewVerifyEmailHandler(credRepo *repository.CredentialsRepository, otpRepo *repository.OTPRepository) *VerifyEmailHandler {
	return &VerifyEmailHandler{credRepo: credRepo, otpRepo: otpRepo}
}

func (h *VerifyEmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgmw.ErrorResponse(w, apperrors.Validation("body", "invalid JSON"))
		return
	}
	if req.Email == "" || req.OTP == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("body", "email and otp are required"))
		return
	}

	creds, err := h.credRepo.FindByEmail(r.Context(), req.Email)
	if err != nil {
		// Don't leak whether the email exists — return the same error shape.
		pkgmw.ErrorResponse(w, apperrors.Validation("otp", "invalid or expired code"))
		return
	}

	if creds.EmailVerified {
		pkgmw.ErrorResponse(w, apperrors.Conflict("email is already verified"))
		return
	}

	codeHash := domain.HashOTP(req.OTP)
	otpRecord, err := h.otpRepo.FindActiveByCredentialAndPurpose(r.Context(), creds.ID, domain.OTPPurposeEmailVerify)
	if err != nil || otpRecord.CodeHash != codeHash {
		pkgmw.ErrorResponse(w, apperrors.Validation("otp", "invalid or expired code"))
		return
	}

	if err = h.otpRepo.MarkUsed(r.Context(), otpRecord.ID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	if err = h.credRepo.MarkEmailVerified(r.Context(), creds.ID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ResendVerificationRequest is the JSON body for POST /auth/email/verify/resend.
type ResendVerificationRequest struct {
	Email string `json:"email"`
}

// ResendVerificationHandler handles POST /auth/email/verify/resend.
type ResendVerificationHandler struct {
	credRepo    *repository.CredentialsRepository
	otpRepo     *repository.OTPRepository
	emailSender ports.EmailSender
	log         *logger.Logger
}

func NewResendVerificationHandler(
	credRepo *repository.CredentialsRepository,
	otpRepo *repository.OTPRepository,
	emailSender ports.EmailSender,
	log *logger.Logger,
) *ResendVerificationHandler {
	return &ResendVerificationHandler{
		credRepo:    credRepo,
		otpRepo:     otpRepo,
		emailSender: emailSender,
		log:         log,
	}
}

func (h *ResendVerificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("body", "email is required"))
		return
	}

	creds, err := h.credRepo.FindByEmail(r.Context(), req.Email)
	if err != nil {
		// Return 204 to prevent email enumeration
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if creds.EmailVerified {
		pkgmw.ErrorResponse(w, apperrors.Conflict("email is already verified"))
		return
	}

	otpCode, otpHash, otpExpiry, err := domain.GenerateOTP()
	if err != nil {
		pkgmw.ErrorResponse(w, apperrors.Internal(err))
		return
	}

	if err = h.otpRepo.Create(r.Context(), creds.ID, domain.OTPPurposeEmailVerify, otpHash, otpExpiry); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	go func() {
		if err := h.emailSender.Send(context.Background(), "EMAIL_VERIFY", creds.Email, map[string]string{
			"otp": otpCode,
		}); err != nil {
			h.log.Error("failed to resend verification email", logger.String("email", creds.Email), logger.Err(err))
		}
	}()

	w.WriteHeader(http.StatusNoContent)
}
