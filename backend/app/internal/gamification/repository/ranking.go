// Package repository is the gamification domain's persistence layer:
// weekly ranking aggregation, week-close promotion/demotion, and medals.
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	apperrors "github.com/preuni/pkg/errors"

	"github.com/preuni/app/internal/gamification/domain"
)

// RankingEntry is one user's position on the current week's leaderboard.
type RankingEntry struct {
	UserID      string
	DisplayName string
	WeeklyScore int
	LeagueTier  string
	RankInTier  *int
}

// Repository is the gamification domain's persistence layer.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs a Repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// weekStart returns the UTC Monday of the week containing t.
func weekStart(t time.Time) time.Time {
	u := t.UTC()
	offset := (int(u.Weekday()) + 6) % 7 // days since Monday (Sunday=0 -> 6)
	monday := u.AddDate(0, 0, -offset)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
}

// EnsureCurrentWeekEntry creates this week's ranking row for userID if it
// doesn't exist yet, inheriting their league tier from the prior week (or
// bronze for a brand-new user), and recomputes their weekly_score from
// this week's graded essays. Called after each essay is graded so the
// leaderboard stays live.
func (r *Repository) EnsureCurrentWeekEntry(ctx context.Context, userID string, now time.Time) error {
	ws := weekStart(now)

	var tier string
	err := r.db.QueryRow(ctx, `
		SELECT league_tier FROM gamification.weekly_ranking_entries
		WHERE user_id = $1
		ORDER BY week_start DESC
		LIMIT 1
	`, userID).Scan(&tier)
	if err != nil {
		if err != pgx.ErrNoRows {
			return apperrors.Internal(err)
		}
		tier = string(domain.TierBronze)
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO gamification.weekly_ranking_entries (id, user_id, week_start, weekly_score, league_tier)
		VALUES ($1, $2, $3, 0, $4)
		ON CONFLICT (user_id, week_start) DO NOTHING
	`, uuid.New().String(), userID, ws, tier)
	if err != nil {
		return apperrors.Internal(err)
	}

	_, err = r.db.Exec(ctx, `
		UPDATE gamification.weekly_ranking_entries e
		SET weekly_score = COALESCE((
			SELECT SUM(eg.overall_score)
			FROM essay.essay_grades eg
			JOIN essay.essay_submissions es ON es.id = eg.submission_id
			WHERE es.user_id = e.user_id AND eg.graded_at >= $2 AND eg.graded_at < $2 + interval '7 days'
		), 0)
		WHERE e.user_id = $1 AND e.week_start = $2
	`, userID, ws)
	if err != nil {
		return apperrors.Internal(err)
	}
	return nil
}

// WeeklyLeaderboard returns the current week's entries for the given tier,
// ranked (domain.RankEntries), highest score first.
func (r *Repository) WeeklyLeaderboard(ctx context.Context, tier string, now time.Time) ([]RankingEntry, error) {
	ws := weekStart(now)

	rows, err := r.db.Query(ctx, `
		SELECT e.user_id, s.display_name, e.weekly_score, e.league_tier, e.rank_in_tier
		FROM gamification.weekly_ranking_entries e
		JOIN users.students s ON s.id = e.user_id
		WHERE e.week_start = $1 AND e.league_tier = $2
		ORDER BY e.weekly_score DESC
	`, ws, tier)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	defer rows.Close()

	var entries []RankingEntry
	for rows.Next() {
		var e RankingEntry
		if err := rows.Scan(&e.UserID, &e.DisplayName, &e.WeeklyScore, &e.LeagueTier, &e.RankInTier); err != nil {
			return nil, apperrors.Internal(err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// MyRanking returns userID's current-week entry.
func (r *Repository) MyRanking(ctx context.Context, userID string, now time.Time) (*RankingEntry, error) {
	ws := weekStart(now)
	row := r.db.QueryRow(ctx, `
		SELECT e.user_id, s.display_name, e.weekly_score, e.league_tier, e.rank_in_tier
		FROM gamification.weekly_ranking_entries e
		JOIN users.students s ON s.id = e.user_id
		WHERE e.user_id = $1 AND e.week_start = $2
	`, userID, ws)

	var e RankingEntry
	if err := row.Scan(&e.UserID, &e.DisplayName, &e.WeeklyScore, &e.LeagueTier, &e.RankInTier); err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.NotFound("ranking entry")
		}
		return nil, apperrors.Internal(err)
	}
	return &e, nil
}
