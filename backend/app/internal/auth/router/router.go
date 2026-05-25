// Package router exposes the auth router builder for the monolith.
package router

import (
	chi "github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preuni/pkg/logger"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/app/internal/auth/domain"
	"github.com/preuni/app/internal/auth/handler"
	"github.com/preuni/app/internal/auth/ports"
	"github.com/preuni/app/internal/auth/repository"
)

// Deps captures everything the auth router needs.
// StudentProvisioner + EmailSender are injected by the monolith wiring with
// in-process implementations from app/internal/adapters/.
type Deps struct {
	Pool                 *pgxpool.Pool
	JWTSigningKey        string
	JWTAccessExpirySec   int
	JWTRefreshExpiryDays int
	StudentProvisioner   ports.StudentProvisioner
	EmailSender          ports.EmailSender
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
	deleteAccountH := handler.NewDeleteAccountHandler(credRepo, refreshRepo)

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

