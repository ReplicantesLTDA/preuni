package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	apperrors "github.com/preuni/pkg/errors"
	"github.com/preuni/pkg/logger"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/svc/auth/domain"
	"github.com/preuni/svc/auth/ports"
	"github.com/preuni/svc/auth/repository"
)

// PasswordResetRequestRequest is the body for POST /auth/password/reset-request.
type PasswordResetRequestRequest struct {
	Email string `json:"email"`
}

// PasswordResetRequestHandler handles POST /auth/password/reset-request.
// Always returns 202 to prevent email enumeration.
type PasswordResetRequestHandler struct {
	credRepo    *repository.CredentialsRepository
	otpRepo     *repository.OTPRepository
	emailSender ports.EmailSender
	log         *logger.Logger
}

// NewPasswordResetRequestHandler constructs the handler.
func NewPasswordResetRequestHandler(credRepo *repository.CredentialsRepository, otpRepo *repository.OTPRepository, emailSender ports.EmailSender, log *logger.Logger) *PasswordResetRequestHandler {
	return &PasswordResetRequestHandler{credRepo: credRepo, otpRepo: otpRepo, emailSender: emailSender, log: log}
}

func (h *PasswordResetRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetRequestRequest
	json.NewDecoder(r.Body).Decode(&req) //nolint:errcheck

	// Always respond 202 — prevents email enumeration
	w.WriteHeader(http.StatusAccepted)

	if req.Email == "" {
		return
	}

	go func() {
		creds, err := h.credRepo.FindByEmail(context.Background(), req.Email)
		if err != nil || !creds.EmailVerified {
			return
		}
		otpCode, otpHash, otpExpiry, err := domain.GenerateOTP()
		if err != nil {
			h.log.Error("password reset otp generation failed", logger.Err(err))
			return
		}
		if err = h.otpRepo.Create(context.Background(), creds.ID, domain.OTPPurposePasswordReset, otpHash, otpExpiry); err != nil {
			h.log.Error("password reset otp store failed", logger.Err(err))
			return
		}
		if err := h.emailSender.Send(context.Background(), "PASSWORD_RESET", req.Email, map[string]string{"otp": otpCode}); err != nil {
			h.log.Error("password reset email send failed", logger.Err(err))
		}
	}()
}

// PasswordResetRequest is the body for POST /auth/password/reset.
type PasswordResetRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

// PasswordResetHandler handles POST /auth/password/reset.
type PasswordResetHandler struct {
	credRepo    *repository.CredentialsRepository
	otpRepo     *repository.OTPRepository
	refreshRepo *repository.RefreshTokenRepository
}

// NewPasswordResetHandler constructs the handler.
func NewPasswordResetHandler(credRepo *repository.CredentialsRepository, otpRepo *repository.OTPRepository, refreshRepo *repository.RefreshTokenRepository) *PasswordResetHandler {
	return &PasswordResetHandler{credRepo: credRepo, otpRepo: otpRepo, refreshRepo: refreshRepo}
}

func (h *PasswordResetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.OTP == "" || req.NewPassword == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("body", "email, otp, and new_password are required"))
		return
	}

	creds, err := h.credRepo.FindByEmail(r.Context(), req.Email)
	if err != nil {
		pkgmw.ErrorResponse(w, apperrors.Validation("email", "invalid or expired code"))
		return
	}

	if err = domain.ValidatePassword(req.NewPassword); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	codeHash := domain.HashOTP(req.OTP)
	otpRecord, err := h.otpRepo.FindActiveByCredentialAndPurpose(r.Context(), creds.ID, domain.OTPPurposePasswordReset)
	if err != nil || otpRecord.CodeHash != codeHash || otpRecord.ExpiresAt.Before(time.Now()) {
		pkgmw.ErrorResponse(w, apperrors.Validation("otp", "invalid or expired code"))
		return
	}

	if err = h.otpRepo.MarkUsed(r.Context(), otpRecord.ID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	newHash, err := domain.HashPassword(req.NewPassword)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	if err = h.credRepo.UpdatePasswordHash(r.Context(), creds.ID, newHash); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	if err = h.refreshRepo.RevokeAllForCredential(r.Context(), creds.ID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
