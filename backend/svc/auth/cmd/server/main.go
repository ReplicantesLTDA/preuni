package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	chi "github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preuni/pkg/config"
	"github.com/preuni/pkg/logger"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/svc/auth/domain"
	"github.com/preuni/svc/auth/handler"
	"github.com/preuni/svc/auth/repository"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	// Auth-specific config
	jwtSigningKey := requireEnv("JWT_SIGNING_KEY")
	jwtAccessExpiry := envInt("JWT_ACCESS_EXPIRY_SECONDS", 3600)
	jwtRefreshExpiry := envInt("JWT_REFRESH_EXPIRY_DAYS", 30)
	mailSvcURL := envOrDefault("MAIL_SERVICE_URL", "http://localhost:4000")
	userSvcURL := envOrDefault("USER_SERVICE_URL", "http://localhost:8082")

	// Connect to PostgreSQL
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", logger.Err(err))
		os.Exit(1)
	}
	defer pool.Close()

	// Build shared dependencies
	credRepo := repository.NewCredentialsRepository(pool)
	otpRepo := repository.NewOTPRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	jwtSvc := domain.NewJWTService(jwtSigningKey, jwtAccessExpiry, jwtRefreshExpiry)

	// Build handlers
	registerH := handler.NewRegisterHandler(
		credRepo, otpRepo, refreshRepo, jwtSvc,
		userSvcURL, mailSvcURL, cfg.InternalServiceToken, log,
	)
	loginH := handler.NewLoginHandler(credRepo, refreshRepo, jwtSvc)
	verifyEmailH := handler.NewVerifyEmailHandler(credRepo, otpRepo)
	otpRequestH := handler.NewOTPLoginRequestHandler(credRepo, otpRepo, mailSvcURL, log)
	otpVerifyH := handler.NewOTPLoginVerifyHandler(credRepo, otpRepo, refreshRepo, jwtSvc)
	logoutH := handler.NewLogoutHandler(refreshRepo)
	changePasswordH := handler.NewChangePasswordHandler(credRepo, refreshRepo)
	pwResetRequestH := handler.NewPasswordResetRequestHandler(credRepo, otpRepo, mailSvcURL, log)
	pwResetH := handler.NewPasswordResetHandler(credRepo, otpRepo, refreshRepo)
	changeEmailReqH := handler.NewChangeEmailRequestHandler(credRepo, otpRepo, mailSvcURL, log)
	changeEmailConfH := handler.NewChangeEmailConfirmHandler(credRepo, otpRepo, refreshRepo)
	deleteAccountH := handler.NewDeleteAccountHandler(credRepo, refreshRepo, userSvcURL)

	authMiddleware := pkgmw.RequireAuth([]byte(jwtSigningKey))

	// Build router
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	r.Route("/v1/auth", func(r chi.Router) {
		// Public endpoints
		r.Post("/register", registerH.ServeHTTP)
		r.Post("/login", loginH.ServeHTTP)
		r.Post("/email/verify", verifyEmailH.ServeHTTP)
		r.Post("/otp/request", otpRequestH.ServeHTTP)
		r.Post("/otp/verify", otpVerifyH.ServeHTTP)
		r.Post("/password/reset/request", pwResetRequestH.ServeHTTP)
		r.Post("/password/reset/confirm", pwResetH.ServeHTTP)

		// Authenticated endpoints
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Post("/logout", logoutH.ServeHTTP)
			r.Post("/password/change", changePasswordH.ServeHTTP)
			r.Post("/email/change/request", changeEmailReqH.ServeHTTP)
			r.Post("/email/change/confirm", changeEmailConfH.ServeHTTP)
			r.Delete("/account", deleteAccountH.ServeHTTP)
		})
	})

	log.Info("starting auth service", logger.String("port", cfg.Port))
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Error("server exited", logger.Err(err))
		os.Exit(1)
	}
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
