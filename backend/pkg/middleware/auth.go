// Package middleware provides shared HTTP middleware for all preuni services.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	apperrors "github.com/preuni/pkg/errors"
)

type contextKey string

const (
	contextKeyUserID    contextKey = "user_id"
	contextKeyUserEmail contextKey = "user_email"
)

// Claims is the JWT payload structure shared across all services.
type Claims struct {
	UserID    string `json:"uid"`
	UserEmail string `json:"email"`
	jwt.RegisteredClaims
}

// RequireAuth validates the JWT in the Authorization header and injects
// X-User-ID + X-User-Email into the request context and headers.
// The signingKey must be the same secret used by auth-svc to sign tokens.
func RequireAuth(signingKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if raw == "" {
				writeError(w, apperrors.Unauthorized("missing authorization token"))
				return
			}

			claims := &Claims{}
			_, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, apperrors.Unauthorized("unexpected signing method")
				}
				return signingKey, nil
			})
			if err != nil {
				writeError(w, apperrors.Unauthorized("invalid or expired token"))
				return
			}

			// Inject into context and forwarding headers so downstream services can read either.
			ctx := context.WithValue(r.Context(), contextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, contextKeyUserEmail, claims.UserEmail)
			r = r.WithContext(ctx)
			r.Header.Set("X-User-ID", claims.UserID)
			r.Header.Set("X-User-Email", claims.UserEmail)

			next.ServeHTTP(w, r)
		})
	}
}

// UserIDFromContext extracts the authenticated user's ID from the request context.
// Returns an empty string if not set (i.e., unauthenticated route).
func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(contextKeyUserID).(string)
	return v
}

// UserEmailFromContext extracts the authenticated user's email from the request context.
func UserEmailFromContext(ctx context.Context) string {
	v, _ := ctx.Value(contextKeyUserEmail).(string)
	return v
}
