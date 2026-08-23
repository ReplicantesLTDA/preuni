package integration_test

import (
	"context"
	"testing"
	"time"

	essayrepo "github.com/preuni/app/internal/essay/repository"
	gamificationrepo "github.com/preuni/app/internal/gamification/repository"
	socialrepo "github.com/preuni/app/internal/social/repository"
	streakrepo "github.com/preuni/app/internal/streak/repository"
	userrepo "github.com/preuni/app/internal/user/repository"
)

// TestIntegration_RepositoryFaultInjection_CanceledContextReturnsInternalError
// reaches the apperrors.Internal(err) branches across several repositories'
// DB calls by passing an already-canceled context.Context -- genuine pgx
// behavior (the same error a real client disconnect or request timeout
// produces in production), not a mock. These branches were previously
// unreachable via any success-path test.
func TestIntegration_RepositoryFaultInjection_CanceledContextReturnsInternalError(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("essay.ListByUser", func(t *testing.T) {
		repo := essayrepo.NewRepository(pool, streakrepo.NewRepository(), nil)
		_, err := repo.ListByUser(canceled, studentID)
		if err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})

	t.Run("gamification.EnsureCurrentWeekEntry", func(t *testing.T) {
		repo := gamificationrepo.NewRepository(pool)
		err := repo.EnsureCurrentWeekEntry(canceled, studentID, time.Now().UTC())
		if err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})

	t.Run("gamification.WeeklyLeaderboard", func(t *testing.T) {
		repo := gamificationrepo.NewRepository(pool)
		_, err := repo.WeeklyLeaderboard(canceled, "bronze", time.Now().UTC())
		if err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})

	t.Run("gamification.WeekClose", func(t *testing.T) {
		repo := gamificationrepo.NewRepository(pool)
		_, err := repo.WeekClose(canceled, time.Now().UTC())
		if err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})

	t.Run("social.ListFriends", func(t *testing.T) {
		repo := socialrepo.NewRepository(pool)
		_, err := repo.ListFriends(canceled, studentID)
		if err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})

	t.Run("user.Update", func(t *testing.T) {
		repo := userrepo.NewStudentRepository(pool)
		name := "Won't Persist"
		_, err := repo.Update(canceled, studentID, &userrepo.StudentPatch{DisplayName: &name})
		if err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})
}
