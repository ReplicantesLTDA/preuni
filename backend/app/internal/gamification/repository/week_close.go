package repository

import (
	"context"
	"time"

	apperrors "github.com/preuni/pkg/errors"

	"github.com/preuni/app/internal/gamification/domain"
)

var allTiers = []domain.Tier{
	domain.TierBronze, domain.TierSilver, domain.TierGold, domain.TierPlatinum, domain.TierDiamond,
}

type weekCloseRow struct {
	userID           string
	score            int
	earliestGradedAt time.Time
}

// WeekClose finalizes the most recently completed week (the UTC week
// before the one containing `now`): ranks each tier (domain.RankEntries),
// records rank_in_tier, promotes/demotes users into next week's entry
// (domain.TierMovement/NextTier), resets their weekly score to zero for
// the new week (without touching their all-time streak), and awards
// tier_promotion / weekly_top_finish medals.
//
// Idempotent: a tier-week already closed (rank_in_tier populated) is
// skipped, so calling this repeatedly (e.g. from a daily ticker) is safe.
func (r *Repository) WeekClose(ctx context.Context, now time.Time) (processedTiers int, err error) {
	closedWeek := weekStart(now.AddDate(0, 0, -7))
	nextWeek := weekStart(now)

	for _, tier := range allTiers {
		rows, err := r.db.Query(ctx, `
			SELECT e.user_id, e.weekly_score,
			       COALESCE(MIN(eg.graded_at), now()) AS earliest_graded_at
			FROM gamification.weekly_ranking_entries e
			LEFT JOIN essay.essay_submissions es ON es.user_id = e.user_id
			LEFT JOIN essay.essay_grades eg ON eg.submission_id = es.id
			                                 AND eg.graded_at >= e.week_start AND eg.graded_at < e.week_start + interval '7 days'
			WHERE e.week_start = $1 AND e.league_tier = $2 AND e.rank_in_tier IS NULL
			GROUP BY e.user_id, e.weekly_score
		`, closedWeek, string(tier))
		if err != nil {
			return processedTiers, apperrors.Internal(err)
		}

		var entries []weekCloseRow
		for rows.Next() {
			var wr weekCloseRow
			if err := rows.Scan(&wr.userID, &wr.score, &wr.earliestGradedAt); err != nil {
				rows.Close()
				return processedTiers, apperrors.Internal(err)
			}
			entries = append(entries, wr)
		}
		rows.Close()

		if len(entries) == 0 {
			continue
		}

		scoreEntries := make([]domain.ScoreEntry, len(entries))
		for i, e := range entries {
			scoreEntries[i] = domain.ScoreEntry{UserID: e.userID, Score: e.score, EarliestGradedAt: e.earliestGradedAt}
		}
		ranked := domain.RankEntries(scoreEntries)

		for i, e := range ranked {
			rankInTier := i + 1
			if err := r.closeOneEntry(ctx, e.UserID, closedWeek, nextWeek, tier, rankInTier, len(ranked), e.Score); err != nil {
				return processedTiers, err
			}
		}
		processedTiers++
	}
	return processedTiers, nil
}

func (r *Repository) closeOneEntry(ctx context.Context, userID string, closedWeek, nextWeek time.Time, tier domain.Tier, rankInTier, tierSize, score int) error {
	if _, err := r.db.Exec(ctx, `
		UPDATE gamification.weekly_ranking_entries
		SET rank_in_tier = $3
		WHERE user_id = $1 AND week_start = $2
	`, userID, closedWeek, rankInTier); err != nil {
		return apperrors.Internal(err)
	}

	movement := domain.TierMovement(rankInTier, tierSize)
	nextTier := domain.NextTier(tier, movement)

	if _, err := r.db.Exec(ctx, `
		INSERT INTO gamification.weekly_ranking_entries (id, user_id, week_start, weekly_score, league_tier)
		VALUES (gen_random_uuid(), $1, $2, 0, $3)
		ON CONFLICT (user_id, week_start) DO UPDATE SET league_tier = EXCLUDED.league_tier
	`, userID, nextWeek, string(nextTier)); err != nil {
		return apperrors.Internal(err)
	}

	if domain.TierPromotionMedalEarned(movement) {
		if err := r.awardMedal(ctx, userID, domain.MedalTierPromotion); err != nil {
			return err
		}
	}
	if domain.WeeklyTopFinishMedalEarned(rankInTier) {
		if err := r.awardMedal(ctx, userID, domain.MedalWeeklyTopFinish); err != nil {
			return err
		}
	}
	_ = score
	return nil
}
