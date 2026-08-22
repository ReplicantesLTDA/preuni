// Package router exposes the social (friends) router builder for the monolith.
package router

import (
	chi "github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pkgmw "github.com/preuni/pkg/middleware"

	"github.com/preuni/app/internal/social/handler"
	"github.com/preuni/app/internal/social/repository"
)

// Deps captures everything the social router needs.
type Deps struct {
	Pool          *pgxpool.Pool
	JWTSigningKey string
}

// Mount registers /v1/friends/* onto r.
func Mount(r chi.Router, d Deps) {
	repo := repository.NewRepository(d.Pool)
	h := handler.NewHandler(repo)

	authMiddleware := pkgmw.RequireAuth([]byte(d.JWTSigningKey))

	r.Route("/v1/friends", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/", h.ListFriends)
		r.Post("/requests", h.SendRequest)
		r.Post("/requests/{id}/accept", h.AcceptRequest)
		r.Delete("/{id}", h.RemoveFriend)
	})
}
