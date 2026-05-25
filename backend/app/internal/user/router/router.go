// Package router exposes the user router builder for the monolith.
package router

import (
	chi "github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/app/internal/user/handler"
	"github.com/preuni/app/internal/user/repository"
)

// Deps captures everything the user router needs.
type Deps struct {
	Pool          *pgxpool.Pool
	JWTSigningKey string
	S3Bucket      string
	S3Region      string
}

// Mount registers /v1/students/* onto r.
func Mount(r chi.Router, d Deps) {
	studentRepo := repository.NewStudentRepository(d.Pool)

	getStudentH := handler.NewGetStudentHandler(studentRepo)
	updateStudentH := handler.NewUpdateStudentHandler(studentRepo)
	avatarH := handler.NewAvatarHandler(studentRepo, d.S3Bucket, d.S3Region)
	onboardingH := handler.NewOnboardingHandler(studentRepo)
	dataExportH := handler.NewDataExportHandler(studentRepo)
	deleteStudentH := handler.NewDeleteStudentHandler(studentRepo)

	authMiddleware := pkgmw.RequireAuth([]byte(d.JWTSigningKey))

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
}
