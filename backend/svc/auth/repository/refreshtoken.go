package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	apperrors "github.com/preuni/pkg/errors"
)

// RefreshToken is the persistence model for auth.refresh_tokens.
type RefreshToken struct {
	ID           string
	CredentialID string
	TokenHash    string
	ExpiresAt    time.Time
	RevokedAt    *time.Time
	IssuedAt     time.Time
}

// RefreshTokenRepository manages refresh token persistence.
type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

// NewRefreshTokenRepository creates a RefreshTokenRepository.
func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Store saves a new refresh token record.
func (r *RefreshTokenRepository) Store(ctx context.Context, credentialID, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO auth.refresh_tokens (credential_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, credentialID, tokenHash, expiresAt)
	if err != nil {
		return apperrors.Internal(err)
	}
	return nil
}

// FindByHash looks up an active (non-revoked, non-expired) token by its hash.
func (r *RefreshTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, credential_id, token_hash, expires_at, revoked_at, issued_at
		FROM auth.refresh_tokens
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > now()
	`, tokenHash)

	var t RefreshToken
	err := row.Scan(&t.ID, &t.CredentialID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt, &t.IssuedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.Unauthorized("refresh token is invalid or expired")
		}
		return nil, apperrors.Internal(err)
	}
	return &t, nil
}

// Revoke marks a specific token as revoked.
func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE auth.refresh_tokens SET revoked_at = now() WHERE id = $1
	`, id)
	if err != nil {
		return apperrors.Internal(err)
	}
	return nil
}

// RevokeAllForCredential revokes every refresh token for a credential (used on password change or account deletion).
func (r *RefreshTokenRepository) RevokeAllForCredential(ctx context.Context, credentialID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE auth.refresh_tokens
		SET revoked_at = now()
		WHERE credential_id = $1 AND revoked_at IS NULL
	`, credentialID)
	if err != nil {
		return apperrors.Internal(err)
	}
	return nil
}
