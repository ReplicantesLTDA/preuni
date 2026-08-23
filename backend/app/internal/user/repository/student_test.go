package repository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/preuni/app/internal/user/handler/testhelper"
	"github.com/preuni/app/internal/user/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_Create_DuplicateUsernameIsRejected covers Create's
// isDuplicateKeyError branch, previously untested -- every real caller
// derives a username from a fresh UUID, so a collision never happens in
// the normal registration flow.
func TestIntegration_Create_DuplicateUsernameIsRejected(t *testing.T) {
	pool := testhelper.SetupTestDB(t)
	repo := repository.NewStudentRepository(pool)
	ctx := context.Background()

	username := fmt.Sprintf("dup-%d", uuid.New().ID())
	first := &repository.Student{
		ID:          uuid.NewString(),
		DisplayName: "First",
		Username:    username,
		Email:       fmt.Sprintf("first-%s@preuni.test", uuid.NewString()),
	}
	require.NoError(t, repo.Create(ctx, first))

	second := &repository.Student{
		ID:          uuid.NewString(),
		DisplayName: "Second",
		Username:    username,
		Email:       fmt.Sprintf("second-%s@preuni.test", uuid.NewString()),
	}
	err := repo.Create(ctx, second)
	assert.Error(t, err, "expected a duplicate username to be rejected")
}
