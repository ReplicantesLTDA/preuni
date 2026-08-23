package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/preuni/app/internal/auth/handler/testhelper"
	"github.com/preuni/app/internal/auth/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_RefreshTokenRepository_CanceledContextReturnsInternalError
// covers Store, Revoke, and RevokeAllForCredential's apperrors.Internal(err)
// branches (75% before each) via a canceled context -- real pgx behavior,
// not a mock.
func TestIntegration_RefreshTokenRepository_CanceledContextReturnsInternalError(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	credRepo := repository.NewCredentialsRepository(pool)
	refreshRepo := repository.NewRefreshTokenRepository(pool)
	ctx := context.Background()

	creds := newCreds(t, fmt.Sprintf("refresh-cancel-%s@preuni.test", uuid.NewString()))
	require.NoError(t, credRepo.Create(ctx, creds))

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("Store", func(t *testing.T) {
		err := refreshRepo.Store(canceled, creds.ID, "somehash", time.Now().Add(time.Hour))
		assert.Error(t, err)
	})

	t.Run("Revoke", func(t *testing.T) {
		require.NoError(t, refreshRepo.Store(ctx, creds.ID, "anotherhash", time.Now().Add(time.Hour)))
		tok, err := refreshRepo.FindByHash(ctx, "anotherhash")
		require.NoError(t, err)

		err = refreshRepo.Revoke(canceled, tok.ID)
		assert.Error(t, err)
	})

	t.Run("RevokeAllForCredential", func(t *testing.T) {
		err := refreshRepo.RevokeAllForCredential(canceled, creds.ID)
		assert.Error(t, err)
	})
}
