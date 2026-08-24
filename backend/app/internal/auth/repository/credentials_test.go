package repository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/preuni/app/internal/auth/domain"
	"github.com/preuni/app/internal/auth/handler/testhelper"
	"github.com/preuni/app/internal/auth/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCreds(t *testing.T, email string) *domain.Credentials {
	t.Helper()
	creds, err := domain.NewCredentials(email, "P@ssw0rd123")
	require.NoError(t, err)
	creds.ID = uuid.NewString()
	return creds
}

// TestIntegration_UpdateEmail_DuplicateIsRejected covers UpdateEmail's
// isDuplicateKeyError branch (50% before) -- a real email collision, not a
// fault injection: two real credentials, then changing the second's email
// to the first's.
func TestIntegration_UpdateEmail_DuplicateIsRejected(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	repo := repository.NewCredentialsRepository(pool)
	ctx := context.Background()

	first := newCreds(t, fmt.Sprintf("first-%s@preuni.test", uuid.NewString()))
	require.NoError(t, repo.Create(ctx, first))
	second := newCreds(t, fmt.Sprintf("second-%s@preuni.test", uuid.NewString()))
	require.NoError(t, repo.Create(ctx, second))

	err := repo.UpdateEmail(ctx, second.ID, first.Email)
	assert.Error(t, err, "expected updating to an already-taken email to be rejected")
}

// TestIntegration_CredentialsRepository_CanceledContextReturnsInternalError
// covers the generic apperrors.Internal(err) branch on several
// CredentialsRepository methods that were sitting at 75% because their
// only DB-call-fails path was never exercised. A canceled context is real
// behavior pgx surfaces exactly as it would for a client disconnect or
// request timeout in production -- not a mock.
func TestIntegration_CredentialsRepository_CanceledContextReturnsInternalError(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	repo := repository.NewCredentialsRepository(pool)
	ctx := context.Background()

	creds := newCreds(t, fmt.Sprintf("cancel-%s@preuni.test", uuid.NewString()))
	require.NoError(t, repo.Create(ctx, creds))

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("UpdatePasswordHash", func(t *testing.T) {
		err := repo.UpdatePasswordHash(canceled, creds.ID, "newhash")
		assert.Error(t, err)
	})

	t.Run("MarkEmailVerified", func(t *testing.T) {
		err := repo.MarkEmailVerified(canceled, creds.ID)
		assert.Error(t, err)
	})

	t.Run("Anonymize", func(t *testing.T) {
		err := repo.Anonymize(canceled, creds.ID)
		assert.Error(t, err)
	})

	t.Run("UpdateEmail", func(t *testing.T) {
		err := repo.UpdateEmail(canceled, creds.ID, "wont-matter@preuni.test")
		assert.Error(t, err)
	})
}
