package handler

import (
	"context"
	"encoding/json"
	"net/http"

	apperrors "github.com/preuni/pkg/errors"
	"github.com/preuni/pkg/logger"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/svc/auth/domain"
	"github.com/preuni/svc/auth/repository"
)

// OTPLoginRequestRequest is the JSON body for POST /auth/otp/request.
type OTPLoginRequestRequest struct {
	Email string `json:"email"`
}

// OTPLoginRequestHandler handles POST /auth/otp/request.
// Always returns 202 to prevent email enumeration.
type OTPLoginRequestHandler struct {
	credRepo   *repository.CredentialsRepository
	otpRepo    *repository.OTPRepository
	mailSvcURL string
	log        *logger.Logger
}

func NewOTPLoginRequestHandler(credRepo *repository.CredentialsRepository, otpRepo *repository.OTPRepository, mailSvcURL string, log *logger.Logger) *OTPLoginRequestHandler {
	return &OTPLoginRequestHandler{credRepo: credRepo, otpRepo: otpRepo, mailSvcURL: mailSvcURL, log: log}
}

func (h *OTPLoginRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req OTPLoginRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	// Always respond 202 — don't reveal whether the email is registered.
	w.WriteHeader(http.StatusAccepted)

	go func() {
		creds, err := h.credRepo.FindByEmail(context.Background(), req.Email)
		if err != nil || !creds.EmailVerified {
			return // silently ignore unknown or unverified emails
		}
		otpCode, otpHash, otpExpiry, err := domain.GenerateOTP()
		if err != nil {
			h.log.Error("otp generation failed", logger.Err(err))
			return
		}
		if err = h.otpRepo.Create(context.Background(), creds.ID, domain.OTPPurposeLoginOTP, otpHash, otpExpiry); err != nil {
			h.log.Error("otp store failed", logger.Err(err))
			return
		}
		// Deliberately not storing the OTP handler reference in struct to keep it clean
		rh := &RegisterHandler{mailSvcURL: h.mailSvcURL, log: h.log}
		rh.sendEmail(context.Background(), "OTP_LOGIN", req.Email, map[string]string{"otp": otpCode})
	}()
}

// ── OTP Login Verify ──────────────────────────────────────────────────────────

// OTPLoginVerifyRequest is the JSON body for POST /auth/otp/verify.
type OTPLoginVerifyRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

// OTPLoginVerifyHandler handles POST /auth/otp/verify.
type OTPLoginVerifyHandler struct {
	credRepo    *repository.CredentialsRepository
	otpRepo     *repository.OTPRepository
	refreshRepo *repository.RefreshTokenRepository
	jwtSvc      *domain.JWTService
}

func NewOTPLoginVerifyHandler(credRepo *repository.CredentialsRepository, otpRepo *repository.OTPRepository, refreshRepo *repository.RefreshTokenRepository, jwtSvc *domain.JWTService) *OTPLoginVerifyHandler {
	return &OTPLoginVerifyHandler{credRepo: credRepo, otpRepo: otpRepo, refreshRepo: refreshRepo, jwtSvc: jwtSvc}
}

func (h *OTPLoginVerifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req OTPLoginVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.OTP == "" {
		pkgmw.ErrorResponse(w, apperrors.Validation("body", "email and otp are required"))
		return
	}

	creds, err := h.credRepo.FindByEmail(r.Context(), req.Email)
	if err != nil {
		pkgmw.ErrorResponse(w, apperrors.Validation("otp", "invalid or expired code"))
		return
	}

	codeHash := domain.HashOTP(req.OTP)
	otpRecord, err := h.otpRepo.FindActiveByCredentialAndPurpose(r.Context(), creds.ID, domain.OTPPurposeLoginOTP)
	if err != nil || otpRecord.CodeHash != codeHash {
		pkgmw.ErrorResponse(w, apperrors.Validation("otp", "invalid or expired code"))
		return
	}
	if err = h.otpRepo.MarkUsed(r.Context(), otpRecord.ID); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	accessToken, err := h.jwtSvc.IssueAccessToken(creds.ID, creds.Email)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	rawRefresh, refreshHash, refreshExpiry, err := h.jwtSvc.IssueRefreshToken()
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	if err = h.refreshRepo.Store(r.Context(), creds.ID, refreshHash, refreshExpiry); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	pkgmw.JSON(w, http.StatusOK, AuthResponse{
		StudentID:    creds.ID,
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    h.jwtSvc.ExpiresIn(),
	})
}
