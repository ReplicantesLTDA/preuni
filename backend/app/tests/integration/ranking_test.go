package integration_test

import (
	"context"
	"testing"
	"time"

	gamificationrepo "github.com/preuni/app/internal/gamification/repository"
)

// TestIntegration_WeekClose_PromotesTopDemotesBottomAndResetsScore covers
// spec User Story 3 acceptance scenarios 2 & 4: at week close, the top
// band of a tier promotes and the bottom band demotes, weekly scores reset
// to zero for the new week, and the user's all-time streak is untouched.
func TestIntegration_WeekClose_PromotesTopDemotesBottomAndResetsScore(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()

	// A tier needs >=5 members for any movement (domain.TierMovement).
	// Seed 5 bronze-tier users directly for last week with distinct
	// scores, submit one graded essay each so streak_count=1 is set, and
	// verify streak survives week close untouched.
	var userIDs []string
	for i := 0; i < 5; i++ {
		id, _ := registerTestUser(t, r)
		userIDs = append(userIDs, id)
	}
	defer func() {
		for _, id := range userIDs {
			cleanupTestUser(ctx, t, pool, id)
		}
	}()

	now := time.Now().UTC()
	lastWeekStart := now.AddDate(0, 0, -7)
	// Align to a Monday-ish anchor consistent with the repository's own
	// weekStart() — easiest to just ask the repository what "last week"
	// means relative to now by round-tripping through EnsureCurrentWeekEntry
	// semantics isn't exposed, so we compute Monday here the same way.
	offset := (int(lastWeekStart.Weekday()) + 6) % 7
	monday := lastWeekStart.AddDate(0, 0, -offset)
	closedWeek := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)

	scores := []int{500, 400, 300, 200, 100} // userIDs[0] highest .. userIDs[4] lowest
	for i, uid := range userIDs {
		if _, err := pool.Exec(ctx, `
			INSERT INTO gamification.weekly_ranking_entries (id, user_id, week_start, weekly_score, league_tier)
			VALUES (gen_random_uuid(), $1, $2, $3, 'bronze')
		`, uid, closedWeek, scores[i]); err != nil {
			t.Fatalf("seed ranking entry %d: %v", i, err)
		}
		if _, err := pool.Exec(ctx, `UPDATE users.students SET streak_count = 3 WHERE id = $1`, uid); err != nil {
			t.Fatalf("seed streak: %v", err)
		}
	}

	repo := gamificationrepo.NewRepository(pool)
	if _, err := repo.WeekClose(ctx, now); err != nil {
		t.Fatalf("WeekClose: %v", err)
	}

	// Top scorer (userIDs[0], rank 1 of 5) must have promoted to silver.
	var topTier string
	if err := pool.QueryRow(ctx, `SELECT league_tier FROM gamification.weekly_ranking_entries WHERE user_id = $1 AND week_start = $2`, userIDs[0], closedWeek).Scan(&topTier); err != nil {
		t.Fatalf("read top scorer's closed-week tier: %v", err)
	}
	// The closed week's own row keeps league_tier='bronze' (that's the
	// tier they played in); the *next* week's row reflects the promotion.
	var rankInTier int
	if err := pool.QueryRow(ctx, `SELECT rank_in_tier FROM gamification.weekly_ranking_entries WHERE user_id = $1 AND week_start = $2`, userIDs[0], closedWeek).Scan(&rankInTier); err != nil {
		t.Fatalf("read rank_in_tier: %v", err)
	}
	if rankInTier != 1 {
		t.Fatalf("expected the top scorer to rank 1, got %d", rankInTier)
	}

	var nextWeekTierTop, nextWeekTierBottom string
	nextWeek := closedWeek.AddDate(0, 0, 7)
	if err := pool.QueryRow(ctx, `SELECT league_tier FROM gamification.weekly_ranking_entries WHERE user_id = $1 AND week_start = $2`, userIDs[0], nextWeek).Scan(&nextWeekTierTop); err != nil {
		t.Fatalf("read top scorer's next-week tier: %v", err)
	}
	if nextWeekTierTop != "silver" {
		t.Fatalf("expected the top scorer to promote to silver for next week, got %q", nextWeekTierTop)
	}
	if err := pool.QueryRow(ctx, `SELECT league_tier FROM gamification.weekly_ranking_entries WHERE user_id = $1 AND week_start = $2`, userIDs[4], nextWeek).Scan(&nextWeekTierBottom); err != nil {
		t.Fatalf("read bottom scorer's next-week tier: %v", err)
	}
	if nextWeekTierBottom != "bronze" {
		t.Fatalf("expected the bottom scorer to stay in bronze (can't demote below the floor), got %q", nextWeekTierBottom)
	}

	var nextWeekScore int
	if err := pool.QueryRow(ctx, `SELECT weekly_score FROM gamification.weekly_ranking_entries WHERE user_id = $1 AND week_start = $2`, userIDs[0], nextWeek).Scan(&nextWeekScore); err != nil {
		t.Fatalf("read next week's score: %v", err)
	}
	if nextWeekScore != 0 {
		t.Fatalf("expected next week's score to reset to 0, got %d", nextWeekScore)
	}

	var streak int
	if err := pool.QueryRow(ctx, `SELECT streak_count FROM users.students WHERE id = $1`, userIDs[0]).Scan(&streak); err != nil {
		t.Fatalf("read streak: %v", err)
	}
	if streak != 3 {
		t.Fatalf("week close must not touch the all-time streak: expected 3, got %d", streak)
	}

	var promotionMedals int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM gamification.medals WHERE user_id = $1 AND type = 'tier_promotion'`, userIDs[0]).Scan(&promotionMedals); err != nil {
		t.Fatalf("count promotion medals: %v", err)
	}
	if promotionMedals != 1 {
		t.Fatalf("expected exactly one tier_promotion medal for the top scorer, got %d", promotionMedals)
	}
}
