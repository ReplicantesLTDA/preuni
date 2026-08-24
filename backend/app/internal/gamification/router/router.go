// Package router exposes the gamification (ranking + medals) router
// builder for the monolith.
package router

import (
	chi "github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pkgmw "github.com/preuni/pkg/middleware"

	"github.com/preuni/app/internal/gamification/handler"
	"github.com/preuni/app/internal/gamification/repository"
)

// Deps captures everything the gamification router needs.
type Deps struct {
	Pool          *pgxpool.Pool
	JWTSigningKey string
}

// Mount registers /v1/ranking/* and /v1/medals/* onto r and returns the
// Repository so cmd/server can wire it into the essay domain's
// GamificationHooks and drive the weekly week-close job.
func Mount(r chi.Router, d Deps) *repository.Repository {
	repo := repository.NewRepository(d.Pool)
	h := handler.NewHandler(repo)

	authMiddleware := pkgmw.RequireAuth([]byte(d.JWTSigningKey))

	r.Route("/v1/ranking", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/weekly", h.WeeklyLeaderboard)
		r.Get("/me", h.MyRanking)
	})

	r.Route("/v1/medals", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/me", h.MyMedals)
	})

	return repo
}
