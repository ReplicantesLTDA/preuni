// Package router exposes the essay router builder for the monolith.
package router

import (
	chi "github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pkgmw "github.com/preuni/pkg/middleware"

	"github.com/preuni/app/internal/essay/handler"
	"github.com/preuni/app/internal/essay/repository"
	streakrepo "github.com/preuni/app/internal/streak/repository"
)

// Deps captures everything the essay router needs.
type Deps struct {
	Pool          *pgxpool.Pool
	JWTSigningKey string
	Gamification  repository.GamificationHooks
}

// Mount registers /v1/essays/* onto r and returns the Repository so
// cmd/server can drive the background reconciler loop from it.
func Mount(r chi.Router, d Deps) *repository.Repository {
	repo := repository.NewRepository(d.Pool, streakrepo.NewRepository(), d.Gamification)

	submitH := handler.NewSubmitEssayHandler(repo)
	getH := handler.NewGetEssayHandler(repo)
	listH := handler.NewListEssaysHandler(repo)

	authMiddleware := pkgmw.RequireAuth([]byte(d.JWTSigningKey))

	r.Route("/v1/essays", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/", submitH.ServeHTTP)
		r.Get("/", listH.ServeHTTP)
		r.Get("/{id}", getH.ServeHTTP)
	})

	return repo
}
