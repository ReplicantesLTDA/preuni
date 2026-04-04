package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	apperrors "github.com/preuni/pkg/errors"
	"github.com/preuni/svc/auth/domain"
)

// CredentialsRepository handles persistence for auth.credentials.
type CredentialsRepository struct {
	db *pgxpool.Pool
}

// NewCredentialsRepository creates a CredentialsRepository backed by the given pool.
func NewCredentialsRepository(db *pgxpool.Pool) *CredentialsRepository {
	return &CredentialsRepository{db: db}
}

// Create inserts a new credentials row. The caller must set creds.ID before calling.
// Returns ErrConflict if the email is already taken.
func (r *CredentialsRepository) Create(ctx context.Context, creds *domain.Credentials) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO auth.credentials (id, email, password_hash, email_verified, created_at, updated_at)
		VALUES ($1, lower($2), $3, false, now(), now())
	`, creds.ID, creds.Email, creds.PasswordHash)
	if err != nil {
		if isDuplicateKeyError(err) {
			return apperrors.Conflict("email address is already registered")
		}
		return apperrors.Internal(err)
	}
	return nil
}

// FindByEmail retrieves credentials by email (case-insensitive).
// Returns ErrNotFound if no record exists.
func (r *CredentialsRepository) FindByEmail(ctx context.Context, email string) (*domain.Credentials, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, email_verified, email_verified_at, created_at, updated_at
		FROM auth.credentials
		WHERE lower(email) = lower($1)
	`, email)
	return scanCredentials(row)
}

// FindByID retrieves credentials by ID.
func (r *CredentialsRepository) FindByID(ctx context.Context, id string) (*domain.Credentials, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, email_verified, email_verified_at, created_at, updated_at
		FROM auth.credentials
		WHERE id = $1
	`, id)
	return scanCredentials(row)
}

// MarkEmailVerified sets email_verified = true for the given credential ID.
func (r *CredentialsRepository) MarkEmailVerified(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.db.Exec(ctx, `
		UPDATE auth.credentials
		SET email_verified = true, email_verified_at = $2, updated_at = now()
		WHERE id = $1
	`, id, now)
	if err != nil {
		return apperrors.Internal(err)
	}
	return nil
}

// UpdatePasswordHash replaces the stored password hash for a credential.
func (r *CredentialsRepository) UpdatePasswordHash(ctx context.Context, id, newHash string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE auth.credentials SET password_hash = $2, updated_at = now() WHERE id = $1
	`, id, newHash)
	if err != nil {
		return apperrors.Internal(err)
	}
	return nil
}

// UpdateEmail replaces the email address for a credential.
func (r *CredentialsRepository) UpdateEmail(ctx context.Context, id, newEmail string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE auth.credentials
		SET email = lower($2), email_verified = true, email_verified_at = now(), updated_at = now()
		WHERE id = $1
	`, id, newEmail)
	if err != nil {
		if isDuplicateKeyError(err) {
			return apperrors.Conflict("email address is already in use")
		}
		return apperrors.Internal(err)
	}
	return nil
}

// Anonymize replaces PII with non-identifying values for GDPR right-to-erasure.
func (r *CredentialsRepository) Anonymize(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE auth.credentials
		SET email = 'deleted+' || id || '@preuni.com.br',
		    password_hash = '',
		    email_verified = false,
		    email_verified_at = NULL,
		    updated_at = now()
		WHERE id = $1
	`, id)
	if err != nil {
		return apperrors.Internal(err)
	}
	return nil
}

func scanCredentials(row pgx.Row) (*domain.Credentials, error) {
	var c domain.Credentials
	var emailVerifiedAt *time.Time
	err := row.Scan(
		&c.ID, &c.Email, &c.PasswordHash,
		&c.EmailVerified, &emailVerifiedAt,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("credentials")
		}
		return nil, apperrors.Internal(err)
	}
	c.EmailVerifiedAt = emailVerifiedAt
	return &c, nil
}

func isDuplicateKeyError(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate key") || contains(err.Error(), "unique constraint"))
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && indexString(s, sub) >= 0)
}

func indexString(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
