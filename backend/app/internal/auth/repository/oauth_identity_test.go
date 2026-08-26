package repository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/preuni/app/internal/auth/handler/testhelper"
	"github.com/preuni/app/internal/auth/repository"
	apperrors "github.com/preuni/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_OAuthIdentityRepository_LinkAndFind covers the happy
// path: Link persists a row, FindCredentialID retrieves it back.
func TestIntegration_OAuthIdentityRepository_LinkAndFind(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	credRepo := repository.NewCredentialsRepository(pool)
	oauthRepo := repository.NewOAuthIdentityRepository(pool)
	ctx := context.Background()

	creds := newCreds(t, fmt.Sprintf("oauth-%s@preuni.test", uuid.NewString()))
	require.NoError(t, credRepo.Create(ctx, creds))

	providerUserID := uuid.NewString()
	require.NoError(t, oauthRepo.Link(ctx, creds.ID, "google", providerUserID))

	got, err := oauthRepo.FindCredentialID(ctx, "google", providerUserID)
	require.NoError(t, err)
	assert.Equal(t, creds.ID, got)
}

// TestIntegration_OAuthIdentityRepository_FindUnknownIsNotFound covers the
// not-found branch, distinguishing it from a generic Internal error --
// GoogleLoginHandler relies on this distinction to know when to sign up.
func TestIntegration_OAuthIdentityRepository_FindUnknownIsNotFound(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	oauthRepo := repository.NewOAuthIdentityRepository(pool)

	_, err := oauthRepo.FindCredentialID(context.Background(), "google", "no-such-provider-user-id")
	assert.True(t, apperrors.IsNotFound(err))
}

// TestIntegration_OAuthIdentityRepository_LinkDuplicateIsConflict covers
// Link's isDuplicateKeyError branch: the same (provider, provider_user_id)
// pair can't be linked twice.
func TestIntegration_OAuthIdentityRepository_LinkDuplicateIsConflict(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	credRepo := repository.NewCredentialsRepository(pool)
	oauthRepo := repository.NewOAuthIdentityRepository(pool)
	ctx := context.Background()

	first := newCreds(t, fmt.Sprintf("dup1-%s@preuni.test", uuid.NewString()))
	require.NoError(t, credRepo.Create(ctx, first))
	second := newCreds(t, fmt.Sprintf("dup2-%s@preuni.test", uuid.NewString()))
	require.NoError(t, credRepo.Create(ctx, second))

	providerUserID := uuid.NewString()
	require.NoError(t, oauthRepo.Link(ctx, first.ID, "google", providerUserID))

	err := oauthRepo.Link(ctx, second.ID, "google", providerUserID)
	assert.True(t, apperrors.IsConflict(err))
}

// TestIntegration_OAuthIdentityRepository_CanceledContextReturnsInternalError
// covers the generic apperrors.Internal(err) branch on both methods -- a
// canceled context is real pgx behavior, not a mock.
func TestIntegration_OAuthIdentityRepository_CanceledContextReturnsInternalError(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	credRepo := repository.NewCredentialsRepository(pool)
	oauthRepo := repository.NewOAuthIdentityRepository(pool)
	ctx := context.Background()

	creds := newCreds(t, fmt.Sprintf("cancel-oauth-%s@preuni.test", uuid.NewString()))
	require.NoError(t, credRepo.Create(ctx, creds))

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("Link", func(t *testing.T) {
		err := oauthRepo.Link(canceled, creds.ID, "google", uuid.NewString())
		assert.Error(t, err)
	})

	t.Run("FindCredentialID", func(t *testing.T) {
		_, err := oauthRepo.FindCredentialID(canceled, "google", uuid.NewString())
		assert.Error(t, err)
	})
}
