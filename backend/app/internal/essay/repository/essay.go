// Package repository is the essay domain's persistence layer: submission
// intake (quota-gated, streak-triggering), and reconciliation of graded
// results from the correction service via the correction_jobs bridge table
// (contracts/internal-bridge.md).
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	apperrors "github.com/preuni/pkg/errors"

	essaydomain "github.com/preuni/app/internal/essay/domain"
	streakrepo "github.com/preuni/app/internal/streak/repository"
)

// Submission mirrors an essay.essay_submissions row.
type Submission struct {
	ID                 string
	UserID             string
	PromptThemeTitle   string
	PromptThemeContext string
	EssayText          string
	SubmissionDay      time.Time
	Status             string
	CorrectionJobID    *string
	SubmittedAt        time.Time
}

// Grade mirrors an essay.essay_grades row.
type Grade struct {
	SubmissionID string
	OverallScore int
	Competencies []byte // raw JSON, decoded by the handler layer
	GradedAt     time.Time
}

// SubmissionView is a submission plus its grade, if graded.
type SubmissionView struct {
	Submission
	Grade *Grade
}

// GamificationHooks lets the essay domain trigger gamification side
// effects (streak medals, weekly leaderboard updates) without importing
// the gamification package's full API — gamification.Repository satisfies
// this structurally. Best-effort: these run after the essay transaction
// commits, so a hook failure never rolls back a submission or a grade.
type GamificationHooks interface {
	AwardStreakMedals(ctx context.Context, userID string, oldStreak, newStreak int) error
	EnsureCurrentWeekEntry(ctx context.Context, userID string, now time.Time) error
}

// Repository is the essay domain's persistence layer.
type Repository struct {
	db           *pgxpool.Pool
	streakRepo   *streakrepo.Repository
	gamification GamificationHooks
}

// NewRepository constructs a Repository. gamification may be nil (no
// medal/leaderboard side effects — e.g. in tests that don't need them).
func NewRepository(db *pgxpool.Pool, streakRepo *streakrepo.Repository, gamification GamificationHooks) *Repository {
	return &Repository{db: db, streakRepo: streakRepo, gamification: gamification}
}

// Submit accepts an essay submission: checks the caller's daily quota
// (free tier: 1/day; Pro: unlimited), records it, advances the caller's
// streak, and enqueues a correction_jobs row for the correction service —
// all in one transaction (constitution: streak advances on acceptance, not
// on completed grading; quota is enforced by the monolith, before the
// request reaches the correction service).
func (r *Repository) Submit(ctx context.Context, userID, themeTitle, themeContext, essayText string) (*Submission, error) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	snap, err := r.streakRepo.GetForUpdate(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	var alreadySubmitted bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM essay.essay_submissions
			WHERE user_id = $1 AND submission_day = $2
		)
	`, userID, today).Scan(&alreadySubmitted)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if essaydomain.QuotaExceeded(snap.SubscriptionTier, alreadySubmitted) {
		return nil, apperrors.QuotaExceeded("free tier allows one essay submission per day")
	}

	submissionID := uuid.New().String()
	jobID := uuid.New().String()

	_, err = tx.Exec(ctx, `
		INSERT INTO essay.essay_submissions
			(id, user_id, prompt_theme_title, prompt_theme_context, essay_text, submission_day, status, correction_job_id, submitted_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7, now())
	`, submissionID, userID, themeTitle, themeContext, essayText, today, jobID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}

	newStreak, err := r.streakRepo.RecordSubmission(ctx, tx, userID, now)
	if err != nil {
		return nil, err
	}
	oldStreak := snap.CurrentStreak

	// The only write this service is granted on the correction schema
	// (contracts/internal-bridge.md) — INSERT-only into correction_jobs.
	_, err = tx.Exec(ctx, `
		INSERT INTO correction.correction_jobs
			(id, user_id, essay_text, prompt_theme_title, prompt_theme_context, status, queued_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', now())
	`, jobID, userID, essayText, themeTitle, themeContext)
	if err != nil {
		return nil, apperrors.Internal(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperrors.Internal(err)
	}

	if r.gamification != nil && newStreak != oldStreak {
		// Best-effort: a medal-award failure must never fail the
		// submission that already committed successfully.
		_ = r.gamification.AwardStreakMedals(ctx, userID, oldStreak, newStreak)
	}

	return &Submission{
		ID:                 submissionID,
		UserID:             userID,
		PromptThemeTitle:   themeTitle,
		PromptThemeContext: themeContext,
		EssayText:          essayText,
		SubmissionDay:      today,
		Status:             "pending",
		CorrectionJobID:    &jobID,
		SubmittedAt:        now,
	}, nil
}

// GetByID returns a submission (with its grade, if graded) owned by userID.
func (r *Repository) GetByID(ctx context.Context, id, userID string) (*SubmissionView, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, prompt_theme_title, prompt_theme_context, essay_text,
		       submission_day, status, correction_job_id, submitted_at
		FROM essay.essay_submissions
		WHERE id = $1 AND user_id = $2
	`, id, userID)

	view, err := scanSubmission(row)
	if err != nil {
		return nil, err
	}

	if view.Status == "graded" {
		grade, err := r.getGrade(ctx, id)
		if err != nil {
			return nil, err
		}
		view.Grade = grade
	}

	return view, nil
}

// ListByUser returns userID's submissions, most recent first.
func (r *Repository) ListByUser(ctx context.Context, userID string) ([]*SubmissionView, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, prompt_theme_title, prompt_theme_context, essay_text,
		       submission_day, status, correction_job_id, submitted_at
		FROM essay.essay_submissions
		WHERE user_id = $1
		ORDER BY submitted_at DESC
	`, userID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	defer rows.Close()

	var views []*SubmissionView
	for rows.Next() {
		view, err := scanSubmission(rows)
		if err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, nil
}

func (r *Repository) getGrade(ctx context.Context, submissionID string) (*Grade, error) {
	row := r.db.QueryRow(ctx, `
		SELECT submission_id, overall_score, competencies, graded_at
		FROM essay.essay_grades
		WHERE submission_id = $1
	`, submissionID)

	var g Grade
	if err := row.Scan(&g.SubmissionID, &g.OverallScore, &g.Competencies, &g.GradedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, apperrors.Internal(err)
	}
	return &g, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSubmission(row rowScanner) (*SubmissionView, error) {
	var s Submission
	err := row.Scan(
		&s.ID, &s.UserID, &s.PromptThemeTitle, &s.PromptThemeContext, &s.EssayText,
		&s.SubmissionDay, &s.Status, &s.CorrectionJobID, &s.SubmittedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.NotFound("essay submission")
		}
		return nil, apperrors.Internal(err)
	}
	return &SubmissionView{Submission: s}, nil
}
