// Package repository persists streak state onto users.students, reusing
// its existing streak_count/streak_last_active_date columns as
// current-streak/last-submission-day plus the longest_streak column added
// alongside this feature (infra/migrations/user/009_*.sql).
package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	apperrors "github.com/preuni/pkg/errors"

	"github.com/preuni/app/internal/streak/domain"
)

// Snapshot is a user's current streak state.
type Snapshot struct {
	CurrentStreak    int
	LongestStreak    int
	LastActiveDay    *time.Time
	SubscriptionTier string
}

// Repository reads/writes streak state on users.students.
type Repository struct{}

// NewRepository constructs a Repository. It holds no state of its own —
// every method takes an explicit pgx.Tx so callers (the essay domain) can
// run the streak update in the same transaction as the submission insert,
// per the constitution: a streak increments at the moment of an accepted
// submission, not after grading completes.
func NewRepository() *Repository {
	return &Repository{}
}

// GetForUpdate locks and returns a user's current streak state.
func (r *Repository) GetForUpdate(ctx context.Context, tx pgx.Tx, userID string) (*Snapshot, error) {
	row := tx.QueryRow(ctx, `
		SELECT streak_count, longest_streak, streak_last_active_date, subscription_tier
		FROM users.students
		WHERE id = $1
		FOR UPDATE
	`, userID)

	var s Snapshot
	if err := row.Scan(&s.CurrentStreak, &s.LongestStreak, &s.LastActiveDay, &s.SubscriptionTier); err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.NotFound("student")
		}
		return nil, apperrors.Internal(err)
	}
	return &s, nil
}

// RecordSubmission advances the caller's streak for an essay submission
// accepted "now" and persists the result within tx. Returns the resulting
// current streak. Safe to call more than once for the same UTC day (Pro
// users submitting multiple essays) — the streak only advances once.
func (r *Repository) RecordSubmission(ctx context.Context, tx pgx.Tx, userID string, now time.Time) (currentStreak int, err error) {
	snap, err := r.GetForUpdate(ctx, tx, userID)
	if err != nil {
		return 0, err
	}

	newCurrent, newLongest, changed := domain.NextStreak(snap.LastActiveDay, now, snap.CurrentStreak, snap.LongestStreak)
	if !changed {
		return newCurrent, nil
	}

	today := time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day(), 0, 0, 0, 0, time.UTC)
	_, err = tx.Exec(ctx, `
		UPDATE users.students
		SET streak_count = $2, longest_streak = $3, streak_last_active_date = $4
		WHERE id = $1
	`, userID, newCurrent, newLongest, today)
	if err != nil {
		return 0, apperrors.Internal(err)
	}
	return newCurrent, nil
}
