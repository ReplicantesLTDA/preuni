// Package router exposes the user router builder for the monolith.
package router

import (
	chi "github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preuni/app/internal/storage"
	"github.com/preuni/app/internal/user/handler"
	"github.com/preuni/app/internal/user/repository"
	pkgmw "github.com/preuni/pkg/middleware"
)

// Deps captures everything the user router needs.
type Deps struct {
	Pool          *pgxpool.Pool
	JWTSigningKey string
	Storage       *storage.Client
}

// Mount registers /v1/students/* onto r.
func Mount(r chi.Router, d Deps) {
	studentRepo := repository.NewStudentRepository(d.Pool)

	getStudentH := handler.NewGetStudentHandler(studentRepo)
	updateStudentH := handler.NewUpdateStudentHandler(studentRepo)
	avatarH := handler.NewAvatarHandler(studentRepo, d.Storage)
	onboardingH := handler.NewOnboardingHandler(studentRepo)
	dataExportH := handler.NewDataExportHandler(studentRepo)
	deleteStudentH := handler.NewDeleteStudentHandler(studentRepo)

	authMiddleware := pkgmw.RequireAuth([]byte(d.JWTSigningKey))

	r.Route("/v1/students", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/me", getStudentH.ServeHTTP)
		r.Patch("/me", updateStudentH.ServeHTTP)
		r.Delete("/me", deleteStudentH.ServeHTTP)
		r.Put("/me/avatar", avatarH.ServeUpload)
		r.Post("/me/avatar/confirm", avatarH.ServeConfirm)
		r.Patch("/me/onboarding", onboardingH.ServeHTTP)
		r.Get("/me/data-export", dataExportH.ServeHTTP)
	})
}
