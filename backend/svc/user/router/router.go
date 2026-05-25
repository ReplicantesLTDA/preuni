// Package router exposes a router builder so both the standalone user-svc
// binary and the unified monolith binary can mount the same routes.
package router

import (
	"fmt"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	pkgmw "github.com/preuni/pkg/middleware"
	"github.com/preuni/svc/user/handler"
	"github.com/preuni/svc/user/repository"
)

// Deps captures everything the user router needs.
type Deps struct {
	Pool                 *pgxpool.Pool
	JWTSigningKey        string
	S3Bucket             string
	S3Region             string
	InternalServiceToken string
}

// Mount registers /internal/students and /v1/students/* onto r.
func Mount(r chi.Router, d Deps) {
	studentRepo := repository.NewStudentRepository(d.Pool)

	createStudentH := handler.NewCreateStudentHandler(studentRepo)
	getStudentH := handler.NewGetStudentHandler(studentRepo)
	updateStudentH := handler.NewUpdateStudentHandler(studentRepo)
	avatarH := handler.NewAvatarHandler(studentRepo, d.S3Bucket, d.S3Region)
	onboardingH := handler.NewOnboardingHandler(studentRepo)
	dataExportH := handler.NewDataExportHandler(studentRepo)
	deleteStudentH := handler.NewDeleteStudentHandler(studentRepo)

	authMiddleware := pkgmw.RequireAuth([]byte(d.JWTSigningKey))
	internalMiddleware := pkgmw.InternalAuth(d.InternalServiceToken)

	r.With(internalMiddleware).Post("/internal/students", createStudentH.ServeHTTP)

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

// NewStandalone returns a fully configured chi.Router for the standalone user-svc binary.
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
