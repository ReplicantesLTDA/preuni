package repository

import (
	"context"
	"time"

	apperrors "github.com/preuni/pkg/errors"
)

// GradingTimeout bounds how long a submission may sit "pending" before the
// reconciler gives up and marks it failed (spec SC-001: no submission is
// left in an indefinite pending state).
const GradingTimeout = 10 * time.Minute

// ReconcileOnce pulls completed/failed correction_jobs results back onto
// their originating essay_submissions: writes essay_grades on success,
// marks the submission failed (surfacing a typed error) on failure or
// timeout. Intended to be called on a short interval by a background loop
// (cmd/server wires this up); safe to call concurrently/repeatedly — each
// row is only ever transitioned out of "pending" once.
func (r *Repository) ReconcileOnce(ctx context.Context) (reconciled int, err error) {
	rows, err := r.db.Query(ctx, `
		SELECT es.id, es.user_id, es.correction_job_id, es.submitted_at,
		       cj.status, cj.completed_at
		FROM essay.essay_submissions es
		JOIN correction.correction_jobs cj ON cj.id = es.correction_job_id
		WHERE es.status = 'pending'
	`)
	if err != nil {
		return 0, apperrors.Internal(err)
	}

	type pendingRow struct {
		submissionID string
		userID       string
		jobID        string
		submittedAt  time.Time
		jobStatus    string
	}
	var pending []pendingRow
	for rows.Next() {
		var pr pendingRow
		var completedAt *time.Time
		if err := rows.Scan(&pr.submissionID, &pr.userID, &pr.jobID, &pr.submittedAt, &pr.jobStatus, &completedAt); err != nil {
			rows.Close()
			return 0, apperrors.Internal(err)
		}
		pending = append(pending, pr)
	}
	rows.Close()

	now := time.Now().UTC()
	for _, pr := range pending {
		switch pr.jobStatus {
		case "completed":
			if err := r.reconcileCompleted(ctx, pr.submissionID, pr.jobID); err != nil {
				return reconciled, err
			}
			if r.gamification != nil {
				_ = r.gamification.EnsureCurrentWeekEntry(ctx, pr.userID, now)
			}
			reconciled++
		case "failed":
			if err := r.reconcileFailed(ctx, pr.submissionID, pr.jobID); err != nil {
				return reconciled, err
			}
			reconciled++
		default:
			if now.Sub(pr.submittedAt) > GradingTimeout {
				if err := r.markTimedOut(ctx, pr.submissionID); err != nil {
					return reconciled, err
				}
				reconciled++
			}
		}
	}
	return reconciled, nil
}

func (r *Repository) reconcileCompleted(ctx context.Context, submissionID, jobID string) error {
	row := r.db.QueryRow(ctx, `
		SELECT final_score, competencies
		FROM correction.corrections
		WHERE job_id = $1
	`, jobID)

	var overallScore int
	var competencies []byte
	if err := row.Scan(&overallScore, &competencies); err != nil {
		return apperrors.Internal(err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return apperrors.Internal(err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `
		INSERT INTO essay.essay_grades (submission_id, overall_score, competencies, graded_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (submission_id) DO NOTHING
	`, submissionID, overallScore, competencies); err != nil {
		return apperrors.Internal(err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE essay.essay_submissions SET status = 'graded' WHERE id = $1
	`, submissionID); err != nil {
		return apperrors.Internal(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return apperrors.Internal(err)
	}
	return nil
}

func (r *Repository) reconcileFailed(ctx context.Context, submissionID, _ string) error {
	return r.markTimedOut(ctx, submissionID)
}

// markTimedOut marks a submission failed. Streak credit is intentionally
// left untouched — the streak already advanced at submission time (spec
// Edge Cases: a grading failure doesn't cost the user their day's streak).
func (r *Repository) markTimedOut(ctx context.Context, submissionID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE essay.essay_submissions SET status = 'failed' WHERE id = $1
	`, submissionID)
	if err != nil {
		return apperrors.Internal(err)
	}
	return nil
}
