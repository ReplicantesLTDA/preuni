package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/preuni/pkg/errors"

	"github.com/preuni/app/internal/gamification/domain"
)

// Medal is one earned achievement.
type Medal struct {
	Type     string
	EarnedAt time.Time
}

// awardMedal inserts a medal row. Not deduplicated at the DB level — a
// user can earn the same milestone type more than once (e.g.
// tier_promotion on every promotion); callers that mean "at most once"
// (a specific streak milestone) rely on AwardStreakMedals only calling
// this when domain.StreakMedalsEarned reports a *newly* crossed milestone.
func (r *Repository) awardMedal(ctx context.Context, userID string, medalType domain.MedalType) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO gamification.medals (id, user_id, type, earned_at)
		VALUES ($1, $2, $3, now())
	`, uuid.New().String(), userID, string(medalType))
	if err != nil {
		return apperrors.Internal(err)
	}
	return nil
}

// AwardStreakMedals awards any milestone medals newly crossed by a streak
// advancing from oldStreak to newStreak. Called by the streak/essay flow
// whenever a submission advances a user's streak.
func (r *Repository) AwardStreakMedals(ctx context.Context, userID string, oldStreak, newStreak int) error {
	for _, medalType := range domain.StreakMedalsEarned(oldStreak, newStreak) {
		if err := r.awardMedal(ctx, userID, medalType); err != nil {
			return err
		}
	}
	return nil
}

// ListMedals returns userID's earned medals, most recent first.
func (r *Repository) ListMedals(ctx context.Context, userID string) ([]Medal, error) {
	rows, err := r.db.Query(ctx, `
		SELECT type, earned_at FROM gamification.medals
		WHERE user_id = $1
		ORDER BY earned_at DESC
	`, userID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	defer rows.Close()

	var medals []Medal
	for rows.Next() {
		var m Medal
		if err := rows.Scan(&m.Type, &m.EarnedAt); err != nil {
			return nil, apperrors.Internal(err)
		}
		medals = append(medals, m)
	}
	return medals, nil
}
