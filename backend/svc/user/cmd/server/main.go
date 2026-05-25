// Deprecated: standalone user-svc binary kept only for rollback.
// Production traffic should be served by backend/svc/monolith.
// Remove once monolith cutover is validated in production.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preuni/pkg/config"
	"github.com/preuni/pkg/logger"
	"github.com/preuni/svc/user/router"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	jwtSigningKey := requireEnv("JWT_SIGNING_KEY")
	s3Bucket := envOrDefault("S3_BUCKET", "preuni-avatars")
	s3Region := envOrDefault("S3_REGION", "us-east-1")

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", logger.Err(err))
		os.Exit(1)
	}
	defer pool.Close()

	r := router.NewStandalone(router.Deps{
		Pool:                 pool,
		JWTSigningKey:        jwtSigningKey,
		S3Bucket:             s3Bucket,
		S3Region:             s3Region,
		InternalServiceToken: cfg.InternalServiceToken,
	})

	log.Info("starting user service", logger.String("port", cfg.Port))
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
