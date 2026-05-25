// Deprecated: standalone auth-svc binary kept only for rollback.
// Production traffic should be served by backend/svc/monolith.
// Remove once monolith cutover is validated in production.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preuni/pkg/config"
	"github.com/preuni/pkg/logger"
	"github.com/preuni/svc/auth/adapters"
	"github.com/preuni/svc/auth/router"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	jwtSigningKey := requireEnv("JWT_SIGNING_KEY")
	jwtAccessExpiry := envInt("JWT_ACCESS_EXPIRY_SECONDS", 3600)
	jwtRefreshExpiry := envInt("JWT_REFRESH_EXPIRY_DAYS", 30)
	mailSvcURL := envOrDefault("MAIL_SERVICE_URL", "http://localhost:4000")
	userSvcURL := envOrDefault("USER_SERVICE_URL", "http://localhost:8082")

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", logger.Err(err))
		os.Exit(1)
	}
	defer pool.Close()

	r := router.NewStandalone(router.Deps{
		Pool:                 pool,
		JWTSigningKey:        jwtSigningKey,
		JWTAccessExpirySec:   jwtAccessExpiry,
		JWTRefreshExpiryDays: jwtRefreshExpiry,
		StudentProvisioner:   adapters.NewHTTPStudentProvisioner(userSvcURL, cfg.InternalServiceToken),
		EmailSender:          adapters.NewHTTPEmailSender(mailSvcURL, cfg.InternalServiceToken),
		UserSvcURL:           userSvcURL,
		Log:                  log,
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
