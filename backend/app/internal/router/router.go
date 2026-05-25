// Package router builds the unified monolith chi.Router by mounting
// per-domain routers (auth, user) and the in-process mail handler.
package router

import (
	"fmt"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preuni/pkg/logger"
	pkgmw "github.com/preuni/pkg/middleware"
	authrouter "github.com/preuni/app/internal/auth/router"
	"github.com/preuni/app/internal/adapters"
	"github.com/preuni/app/internal/config"
	"github.com/preuni/app/internal/mail"
	userrepo "github.com/preuni/app/internal/user/repository"
	userrouter "github.com/preuni/app/internal/user/router"
)

// New constructs the monolith's top-level chi.Router with all routes mounted.
func New(cfg config.Config, pool *pgxpool.Pool, log *logger.Logger) chi.Router {
	r := chi.NewRouter()
	r.Use(pkgmw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	// Build shared collaborators once.
	studentRepo := userrepo.NewStudentRepository(pool)
	sender := buildSender(cfg, log)
	provisioner := adapters.NewInProcessStudentProvisioner(studentRepo)
	emailSender := adapters.NewInProcessEmailSender(cfg.MailFromAddr, cfg.MailFromName, sender, log)

	// Auth — /v1/auth/* (uses in-process adapters; no self-HTTP loopback)
	authrouter.Mount(r, authrouter.Deps{
		Pool:                 pool,
		JWTSigningKey:        cfg.JWTSigningKey,
		JWTAccessExpirySec:   cfg.JWTAccessExpirySec,
		JWTRefreshExpiryDays: cfg.JWTRefreshExpiryDays,
		StudentProvisioner:   provisioner,
		EmailSender:          emailSender,
		UserSvcURL:           cfg.SelfBaseURL, // delete-account legacy path
		Log:                  log,
	})

	// User — /v1/students/* + /internal/students
	userrouter.Mount(r, userrouter.Deps{
		Pool:                 pool,
		JWTSigningKey:        cfg.JWTSigningKey,
		S3Bucket:             cfg.S3Bucket,
		S3Region:             cfg.S3Region,
		InternalServiceToken: cfg.InternalServiceToken,
	})

	// Mail — /internal/email/send (internal-token protected; serves split-service callers)
	mailHandler := mail.NewHandler(cfg.MailFromAddr, cfg.MailFromName, sender, log)
	internalAuth := pkgmw.InternalAuth(cfg.InternalServiceToken)
	r.With(internalAuth).Post("/internal/email/send", mailHandler.ServeHTTP)

	return r
}

func buildSender(cfg config.Config, log *logger.Logger) mail.Sender {
	if cfg.SMTPHost == "" {
		log.Warn("SMTP_HOST not set — using NoopSender (emails will be discarded)")
		return mail.NoopSender{}
	}
	return mail.NewSMTPSender(mail.SMTPConfig{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		Username: cfg.SMTPUser,
		Password: cfg.SMTPPass,
		UseTLS:   cfg.SMTPUseTLS,
	})
}
