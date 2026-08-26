package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	apperrors "github.com/preuni/pkg/errors"
)

// OAuthIdentityRepository handles persistence for auth.oauth_identities.
type OAuthIdentityRepository struct {
	db *pgxpool.Pool
}

// NewOAuthIdentityRepository creates an OAuthIdentityRepository backed by the given pool.
func NewOAuthIdentityRepository(db *pgxpool.Pool) *OAuthIdentityRepository {
	return &OAuthIdentityRepository{db: db}
}

// FindCredentialID returns the linked credential ID for a given provider +
// provider-user-id pair. Returns ErrNotFound if no link exists yet.
func (r *OAuthIdentityRepository) FindCredentialID(ctx context.Context, provider, providerUserID string) (string, error) {
	row := r.db.QueryRow(ctx, `
		SELECT credential_id FROM auth.oauth_identities
		WHERE provider = $1 AND provider_user_id = $2
	`, provider, providerUserID)

	var credentialID string
	if err := row.Scan(&credentialID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.NotFound("oauth identity")
		}
		return "", apperrors.Internal(err)
	}
	return credentialID, nil
}

// Link creates a new provider identity pointing at an existing credential.
func (r *OAuthIdentityRepository) Link(ctx context.Context, credentialID, provider, providerUserID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO auth.oauth_identities (credential_id, provider, provider_user_id)
		VALUES ($1, $2, $3)
	`, credentialID, provider, providerUserID)
	if err != nil {
		if isDuplicateKeyError(err) {
			return apperrors.Conflict("this provider account is already linked")
		}
		return apperrors.Internal(err)
	}
	return nil
}
