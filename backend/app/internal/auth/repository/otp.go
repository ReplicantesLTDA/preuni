package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	apperrors "github.com/preuni/pkg/errors"
	"github.com/preuni/app/internal/auth/domain"
)

// OTPRepository handles persistence for auth.otp_codes.
type OTPRepository struct {
	db *pgxpool.Pool
}

// NewOTPRepository creates an OTPRepository backed by the given pool.
func NewOTPRepository(db *pgxpool.Pool) *OTPRepository {
	return &OTPRepository{db: db}
}

// Create stores a new OTP record.
func (r *OTPRepository) Create(ctx context.Context, credentialID string, purpose domain.OTPPurpose, codeHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO auth.otp_codes (credential_id, purpose, code_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`, credentialID, string(purpose), codeHash, expiresAt)
	if err != nil {
		return apperrors.Internal(err)
	}
	return nil
}

// FindActiveByCredentialAndPurpose returns the most recently created, unused, non-expired OTP
// for the given credential and purpose.
func (r *OTPRepository) FindActiveByCredentialAndPurpose(ctx context.Context, credentialID string, purpose domain.OTPPurpose) (*domain.OTPCode, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, credential_id, purpose, code_hash, expires_at, used_at, created_at
		FROM auth.otp_codes
		WHERE credential_id = $1
		  AND purpose = $2
		  AND used_at IS NULL
		  AND expires_at > now()
		ORDER BY created_at DESC
		LIMIT 1
	`, credentialID, string(purpose))

	var o domain.OTPCode
	var usedAt *time.Time
	err := row.Scan(&o.ID, &o.CredentialID, (*string)(&o.Purpose), &o.CodeHash, &o.ExpiresAt, &usedAt, &o.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("otp code")
		}
		return nil, apperrors.Internal(err)
	}
	o.UsedAt = usedAt
	return &o, nil
}

// MarkUsed marks a specific OTP as consumed.
func (r *OTPRepository) MarkUsed(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.db.Exec(ctx, `
		UPDATE auth.otp_codes SET used_at = $2 WHERE id = $1
	`, id, now)
	if err != nil {
		return apperrors.Internal(err)
	}
	return nil
}
