// Package router exposes the streak router builder for the monolith.
package router

import (
	chi "github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pkgmw "github.com/preuni/pkg/middleware"

	"github.com/preuni/app/internal/streak/handler"
)

// Deps captures everything the streak router needs.
type Deps struct {
	Pool          *pgxpool.Pool
	JWTSigningKey string
}

// Mount registers /v1/streaks/* onto r.
func Mount(r chi.Router, d Deps) {
	getStreakH := handler.NewGetStreakHandler(d.Pool)

	authMiddleware := pkgmw.RequireAuth([]byte(d.JWTSigningKey))

	r.Route("/v1/streaks", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/me", getStreakH.ServeHTTP)
	})
}
