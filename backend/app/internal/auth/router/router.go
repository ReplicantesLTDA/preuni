// Package router exposes a router builder so both the standalone auth-svc
// binary and the unified monolith binary can mount the same routes.
package router

import (
	"fmt"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preuni/pkg/logger"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/app/internal/auth/domain"
	"github.com/preuni/app/internal/auth/handler"
	"github.com/preuni/app/internal/auth/ports"
	"github.com/preuni/app/internal/auth/repository"
)

// Deps captures everything the auth router needs.
//
// StudentProvisioner + EmailSender are required: the caller (standalone
// binary or monolith) constructs the concrete adapter (HTTP or in-process)
// and injects it here.
type Deps struct {
	Pool                 *pgxpool.Pool
	JWTSigningKey        string
	JWTAccessExpirySec   int
	JWTRefreshExpiryDays int
	StudentProvisioner   ports.StudentProvisioner
	EmailSender          ports.EmailSender
	UserSvcURL           string // only used by delete-account legacy path
	Log                  *logger.Logger
}

// Mount registers /v1/auth/* routes onto r.
func Mount(r chi.Router, d Deps) {
	credRepo := repository.NewCredentialsRepository(d.Pool)
	otpRepo := repository.NewOTPRepository(d.Pool)
	refreshRepo := repository.NewRefreshTokenRepository(d.Pool)
	jwtSvc := domain.NewJWTService(d.JWTSigningKey, d.JWTAccessExpirySec, d.JWTRefreshExpiryDays)

	registerH := handler.NewRegisterHandler(credRepo, otpRepo, refreshRepo, jwtSvc, d.StudentProvisioner, d.EmailSender, d.Log)
	loginH := handler.NewLoginHandler(credRepo, refreshRepo, jwtSvc)
	refreshH := handler.NewRefreshTokenHandler(credRepo, refreshRepo, jwtSvc)
	verifyEmailH := handler.NewVerifyEmailHandler(credRepo, otpRepo)
	otpRequestH := handler.NewOTPLoginRequestHandler(credRepo, otpRepo, d.EmailSender, d.Log)
	otpVerifyH := handler.NewOTPLoginVerifyHandler(credRepo, otpRepo, refreshRepo, jwtSvc)
	logoutH := handler.NewLogoutHandler(refreshRepo)
	changePasswordH := handler.NewChangePasswordHandler(credRepo, refreshRepo)
	pwResetRequestH := handler.NewPasswordResetRequestHandler(credRepo, otpRepo, d.EmailSender, d.Log)
	pwResetH := handler.NewPasswordResetHandler(credRepo, otpRepo, refreshRepo)
	changeEmailReqH := handler.NewChangeEmailRequestHandler(credRepo, otpRepo, d.EmailSender, d.Log)
	changeEmailConfH := handler.NewChangeEmailConfirmHandler(credRepo, otpRepo, refreshRepo)
	deleteAccountH := handler.NewDeleteAccountHandler(credRepo, refreshRepo, d.UserSvcURL)

	authMiddleware := pkgmw.RequireAuth([]byte(d.JWTSigningKey))

	r.Route("/v1/auth", func(r chi.Router) {
		r.Post("/register", registerH.ServeHTTP)
		r.Post("/login", loginH.ServeHTTP)
		r.Post("/refresh", refreshH.ServeHTTP)
		r.Post("/email/verify", verifyEmailH.ServeHTTP)
		r.Post("/otp/request", otpRequestH.ServeHTTP)
		r.Post("/otp/verify", otpVerifyH.ServeHTTP)
		r.Post("/password/reset/request", pwResetRequestH.ServeHTTP)
		r.Post("/password/reset/confirm", pwResetH.ServeHTTP)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Post("/logout", logoutH.ServeHTTP)
			r.Post("/password/change", changePasswordH.ServeHTTP)
			r.Post("/email/change/request", changeEmailReqH.ServeHTTP)
			r.Post("/email/change/confirm", changeEmailConfH.ServeHTTP)
			r.Delete("/account", deleteAccountH.ServeHTTP)
		})
	})
}

// NewStandalone returns a fully configured chi.Router for the standalone
// auth-svc binary (includes /health + standard middleware).
func NewStandalone(d Deps) chi.Router {
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	Mount(r, d)
	return r
}
