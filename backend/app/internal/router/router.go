// Package router builds the unified monolith chi.Router by mounting
// per-domain routers (auth, user). Mail is invoked in-process via the
// EmailSender adapter — there is no HTTP mail endpoint.
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
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	studentRepo := userrepo.NewStudentRepository(pool)
	sender := buildSender(cfg, log)
	provisioner := adapters.NewInProcessStudentProvisioner(studentRepo)
	emailSender := adapters.NewInProcessEmailSender(cfg.MailFromAddr, cfg.MailFromName, sender, log)

	authrouter.Mount(r, authrouter.Deps{
		Pool:                 pool,
		JWTSigningKey:        cfg.JWTSigningKey,
		JWTAccessExpirySec:   cfg.JWTAccessExpirySec,
		JWTRefreshExpiryDays: cfg.JWTRefreshExpiryDays,
		StudentProvisioner:   provisioner,
		EmailSender:          emailSender,
		Log:                  log,
	})

	userrouter.Mount(r, userrouter.Deps{
		Pool:          pool,
		JWTSigningKey: cfg.JWTSigningKey,
		S3Bucket:      cfg.S3Bucket,
		S3Region:      cfg.S3Region,
	})

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
