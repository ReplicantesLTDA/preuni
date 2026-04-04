package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	chi "github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preuni/pkg/config"
	"github.com/preuni/pkg/logger"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/svc/user/handler"
	"github.com/preuni/svc/user/repository"
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

	studentRepo := repository.NewStudentRepository(pool)

	createStudentH := handler.NewCreateStudentHandler(studentRepo)
	getStudentH := handler.NewGetStudentHandler(studentRepo)
	updateStudentH := handler.NewUpdateStudentHandler(studentRepo)
	avatarH := handler.NewAvatarHandler(studentRepo, s3Bucket, s3Region)
	onboardingH := handler.NewOnboardingHandler(studentRepo)
	dataExportH := handler.NewDataExportHandler(studentRepo)
	deleteStudentH := handler.NewDeleteStudentHandler(studentRepo)

	authMiddleware := pkgmw.RequireAuth([]byte(jwtSigningKey))
	internalMiddleware := pkgmw.InternalAuth(cfg.InternalServiceToken)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	// Internal endpoint — service-to-service only
	r.With(internalMiddleware).Post("/internal/students", createStudentH.ServeHTTP)

	// Authenticated student endpoints
	r.Route("/v1/students", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/me", getStudentH.ServeHTTP)
		r.Patch("/me", updateStudentH.ServeHTTP)
		r.Delete("/me", deleteStudentH.ServeHTTP)
		r.Get("/me/avatar/upload-url", avatarH.ServeUpload)
		r.Post("/me/avatar/confirm", avatarH.ServeConfirm)
		r.Patch("/me/onboarding", onboardingH.ServeHTTP)
		r.Get("/me/data-export", dataExportH.ServeHTTP)
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
