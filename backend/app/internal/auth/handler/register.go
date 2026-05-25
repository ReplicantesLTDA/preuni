package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/google/uuid"
	"github.com/preuni/pkg/logger"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/app/internal/auth/domain"
	"github.com/preuni/app/internal/auth/ports"
	"github.com/preuni/app/internal/auth/repository"
)

// RegisterRequest is the JSON body for POST /auth/register.
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// AuthResponse is the JSON body returned after a successful register or login.
type AuthResponse struct {
	StudentID    string `json:"student_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// RegisterHandler handles POST /auth/register.
type RegisterHandler struct {
	credRepo           *repository.CredentialsRepository
	otpRepo            *repository.OTPRepository
	refreshRepo        *repository.RefreshTokenRepository
	jwtSvc             *domain.JWTService
	studentProvisioner ports.StudentProvisioner
	emailSender        ports.EmailSender
	log                *logger.Logger
}

// NewRegisterHandler constructs a RegisterHandler with all its dependencies.
func NewRegisterHandler(
	credRepo *repository.CredentialsRepository,
	otpRepo *repository.OTPRepository,
	refreshRepo *repository.RefreshTokenRepository,
	jwtSvc *domain.JWTService,
	studentProvisioner ports.StudentProvisioner,
	emailSender ports.EmailSender,
	log *logger.Logger,
) *RegisterHandler {
	return &RegisterHandler{
		credRepo:           credRepo,
		otpRepo:            otpRepo,
		refreshRepo:        refreshRepo,
		jwtSvc:             jwtSvc,
		studentProvisioner: studentProvisioner,
		emailSender:        emailSender,
		log:                log,
	}
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	creds, err := domain.NewCredentials(req.Email, req.Password)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	if req.DisplayName == "" {
		pkgmw.ErrorResponse(w, fmt.Errorf("display_name is required"))
		return
	}

	studentID := uuid.New().String()
	creds.ID = studentID

	if err = h.credRepo.Create(r.Context(), creds); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	if err = h.studentProvisioner.CreateStudent(r.Context(), ports.CreateStudentRequest{
		StudentID:   studentID,
		DisplayName: req.DisplayName,
		Email:       req.Email,
	}); err != nil {
		h.log.Error("failed to create student profile after registration",
			logger.String("student_id", studentID), logger.Err(err))
		// Non-fatal: the student can still complete their profile later.
	}

	otpCode, otpHash, otpExpiry, err := domain.GenerateOTP()
	if err == nil {
		if storeErr := h.otpRepo.Create(r.Context(), studentID, domain.OTPPurposeEmailVerify, otpHash, otpExpiry); storeErr != nil {
			h.log.Warn("failed to store verification OTP", logger.String("student_id", studentID), logger.Err(storeErr))
		} else {
			go h.fireAndForgetEmail("EMAIL_VERIFY", req.Email, map[string]string{
				"otp": otpCode, "display_name": req.DisplayName,
			})
		}
	}

	verifyBase := os.Getenv("APP_VERIFICATION_BASE_URL")
	if verifyBase == "" {
		verifyBase = "https://app.preuni.com.br/verify-email"
	}
	go h.fireAndForgetEmail("WELCOME", req.Email, map[string]string{
		"display_name":      req.DisplayName,
		"verification_link": verifyBase + "?email=" + url.QueryEscape(req.Email),
	})

	accessToken, err := h.jwtSvc.IssueAccessToken(studentID, creds.Email)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	rawRefresh, refreshHash, refreshExpiry, err := h.jwtSvc.IssueRefreshToken()
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	if err = h.refreshRepo.Store(r.Context(), studentID, refreshHash, refreshExpiry); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	pkgmw.JSON(w, http.StatusCreated, AuthResponse{
		StudentID:    studentID,
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    h.jwtSvc.ExpiresIn(),
	})
}

// fireAndForgetEmail dispatches via the configured EmailSender and logs failures.
// Used by RegisterHandler; sibling handlers call h.emailSender.Send directly.
func (h *RegisterHandler) fireAndForgetEmail(emailType, to string, params map[string]string) {
	if err := h.emailSender.Send(context.Background(), emailType, to, params); err != nil {
		h.log.Error("email send failed",
			logger.String("type", emailType), logger.Err(err))
	}
}
