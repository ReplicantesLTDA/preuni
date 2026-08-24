package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/preuni/app/internal/auth/domain"
	"github.com/preuni/app/internal/auth/handler/testhelper"
	"github.com/preuni/app/internal/auth/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_OTPRepository_CanceledContextReturnsInternalError covers
// Create and MarkUsed's apperrors.Internal(err) branches (75%/80% before) --
// a canceled context is real pgx behavior (client disconnect / request
// timeout in production), not a mock.
func TestIntegration_OTPRepository_CanceledContextReturnsInternalError(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	credRepo := repository.NewCredentialsRepository(pool)
	otpRepo := repository.NewOTPRepository(pool)
	ctx := context.Background()

	creds := newCreds(t, fmt.Sprintf("otp-cancel-%s@preuni.test", uuid.NewString()))
	require.NoError(t, credRepo.Create(ctx, creds))

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("Create", func(t *testing.T) {
		err := otpRepo.Create(canceled, creds.ID, domain.OTPPurposeEmailVerify, "somehash", time.Now().Add(time.Hour))
		assert.Error(t, err)
	})

	t.Run("MarkUsed", func(t *testing.T) {
		// Seed a real OTP first (with a valid, non-canceled context) so
		// MarkUsed's own DB call is what fails, not a missing row.
		require.NoError(t, otpRepo.Create(ctx, creds.ID, domain.OTPPurposeEmailVerify, "anotherhash", time.Now().Add(time.Hour)))
		rec, err := otpRepo.FindActiveByCredentialAndPurpose(ctx, creds.ID, domain.OTPPurposeEmailVerify)
		require.NoError(t, err)

		err = otpRepo.MarkUsed(canceled, rec.ID)
		assert.Error(t, err)
	})
}
