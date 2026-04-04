package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/preuni/pkg/logger"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/svc/auth/domain"
	"github.com/preuni/svc/auth/repository"
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
	credRepo     *repository.CredentialsRepository
	otpRepo      *repository.OTPRepository
	refreshRepo  *repository.RefreshTokenRepository
	jwtSvc       *domain.JWTService
	userSvcURL   string
	mailSvcURL   string
	log          *logger.Logger
}

// NewRegisterHandler constructs a RegisterHandler with all its dependencies.
func NewRegisterHandler(
	credRepo *repository.CredentialsRepository,
	otpRepo *repository.OTPRepository,
	refreshRepo *repository.RefreshTokenRepository,
	jwtSvc *domain.JWTService,
	userSvcURL, mailSvcURL string,
	internalToken string,
	log *logger.Logger,
) *RegisterHandler {
	return &RegisterHandler{
		credRepo:    credRepo,
		otpRepo:     otpRepo,
		refreshRepo: refreshRepo,
		jwtSvc:      jwtSvc,
		userSvcURL:  userSvcURL,
		mailSvcURL:  mailSvcURL,
		log:         log,
	}
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	// Build and validate credentials domain object (validates email + password).
	creds, err := domain.NewCredentials(req.Email, req.Password)
	if err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}
	if req.DisplayName == "" {
		pkgmw.ErrorResponse(w, fmt.Errorf("display_name is required"))
		return
	}

	// Assign a shared ID that will be used across auth-svc and user-svc.
	studentID := uuid.New().String()
	creds.ID = studentID

	// Persist credentials.
	if err = h.credRepo.Create(r.Context(), creds); err != nil {
		pkgmw.ErrorResponse(w, err)
		return
	}

	// Create the student profile in user-svc (internal call).
	if err = h.createStudentProfile(r.Context(), studentID, req.Email, req.DisplayName); err != nil {
		h.log.Error("failed to create student profile after registration",
			logger.String("student_id", studentID), logger.Err(err))
		// Non-fatal for now: the student can still complete their profile later.
	}

	// Generate and store email-verification OTP.
	otpCode, otpHash, otpExpiry, err := domain.GenerateOTP()
	if err == nil {
		if storeErr := h.otpRepo.Create(r.Context(), studentID, domain.OTPPurposeEmailVerify, otpHash, otpExpiry); storeErr != nil {
			h.log.Warn("failed to store verification OTP", logger.String("student_id", studentID), logger.Err(storeErr))
		} else {
			// Send verification email (fire-and-forget: failure doesn't block registration).
			go h.sendEmail(context.Background(), "EMAIL_VERIFY", req.Email, map[string]string{
				"otp": otpCode, "display_name": req.DisplayName,
			})
		}
	}

	// Send welcome email.
	go h.sendEmail(context.Background(), "WELCOME", req.Email, map[string]string{
		"display_name": req.DisplayName,
	})

	// Issue tokens.
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

func (h *RegisterHandler) createStudentProfile(ctx context.Context, id, email, displayName string) error {
	body, _ := json.Marshal(map[string]string{"id": id, "email": email, "display_name": displayName})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.userSvcURL+"/internal/students", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+getInternalToken())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("user-svc returned %d", resp.StatusCode)
	}
	return nil
}

func (h *RegisterHandler) sendEmail(ctx context.Context, emailType, to string, params map[string]string) {
	body, _ := json.Marshal(map[string]any{"type": emailType, "to": to, "params": params})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.mailSvcURL+"/internal/email/send", bytes.NewReader(body))
	if err != nil {
		h.log.Error("failed to build mail request", logger.Err(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+getInternalToken())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		h.log.Error("failed to send email", logger.String("type", emailType), logger.Err(err))
		return
	}
	defer resp.Body.Close()
}

// getInternalToken reads from env at call-time (lazily) to avoid import cycles.
func getInternalToken() string {
	return os.Getenv("INTERNAL_SERVICE_TOKEN")
}
