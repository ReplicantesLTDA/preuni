package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	essayrepo "github.com/preuni/app/internal/essay/repository"
	streakrepo "github.com/preuni/app/internal/streak/repository"
)

// TestIntegration_SubmitEssay_CorrectionFailureDoesNotLoseStreakCredit
// covers spec Edge Cases: when grading fails, the user is notified via a
// failed status, and their streak (already advanced at submission time)
// is left untouched -- they don't lose today's credit and can resubmit.
func TestIntegration_SubmitEssay_CorrectionFailureDoesNotLoseStreakCredit(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	w := submitEssay(t, r, token)
	if w.Code != http.StatusAccepted {
		t.Fatalf("submit: got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	var jobID string
	if err := pool.QueryRow(ctx, `SELECT correction_job_id FROM essay.essay_submissions WHERE id = $1`, resp.ID).Scan(&jobID); err != nil {
		t.Fatalf("read correction_job_id: %v", err)
	}

	// Simulate the correction worker failing the job (e.g. provider_timeout),
	// as ai-corrector's correction_repo.mark_failed would.
	if _, err := pool.Exec(ctx, `UPDATE correction.correction_jobs SET status = 'failed', completed_at = now() WHERE id = $1`, jobID); err != nil {
		t.Fatalf("mark job failed: %v", err)
	}

	essayRepo := essayrepo.NewRepository(pool, streakrepo.NewRepository(), nil)
	if _, err := essayRepo.ReconcileOnce(ctx); err != nil {
		t.Fatalf("ReconcileOnce: %v", err)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/essays/"+resp.ID, nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	var got struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(getW.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != "failed" {
		t.Fatalf("expected status=failed after a worker failure is reconciled, got %q", got.Status)
	}

	streakReq := httptest.NewRequest(http.MethodGet, "/v1/streaks/me", nil)
	streakReq.Header.Set("Authorization", "Bearer "+token)
	streakW := httptest.NewRecorder()
	r.ServeHTTP(streakW, streakReq)
	var streak struct {
		CurrentStreak int `json:"current_streak"`
	}
	if err := json.Unmarshal(streakW.Body.Bytes(), &streak); err != nil {
		t.Fatal(err)
	}
	if streak.CurrentStreak != 1 {
		t.Fatalf("streak credit must survive a grading failure: expected current_streak=1, got %d", streak.CurrentStreak)
	}
}

// TestIntegration_SubmitEssay_ReconcileOnce_TimesOutStalePendingSubmission
// covers ReconcileOnce's "still pending, past GradingTimeout" default
// branch -- previously only its "completed" and "failed" job-status
// branches were tested.
func TestIntegration_SubmitEssay_ReconcileOnce_TimesOutStalePendingSubmission(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	w := submitEssay(t, r, token)
	if w.Code != http.StatusAccepted {
		t.Fatalf("submit: got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	// Backdate submitted_at past GradingTimeout (10m) so this submission is
	// stale, but leave its correction_jobs row 'pending' (never completed or
	// failed) -- ReconcileOnce's timeout branch is the only thing that can
	// resolve a submission stuck like this.
	if _, err := pool.Exec(ctx,
		`UPDATE essay.essay_submissions SET submitted_at = now() - interval '11 minutes' WHERE id = $1`,
		resp.ID,
	); err != nil {
		t.Fatalf("backdate submitted_at: %v", err)
	}

	essayRepo := essayrepo.NewRepository(pool, streakrepo.NewRepository(), nil)
	reconciled, err := essayRepo.ReconcileOnce(ctx)
	if err != nil {
		t.Fatalf("ReconcileOnce: %v", err)
	}
	if reconciled < 1 {
		t.Fatalf("expected at least 1 submission reconciled (timed out), got %d", reconciled)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/essays/"+resp.ID, nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	var got struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(getW.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != "failed" {
		t.Fatalf("expected status=failed after a grading timeout, got %q", got.Status)
	}
}

// TestIntegration_SubmitEssay_NoStudentRowReturns404 covers
// streak.Repository.GetForUpdate's pgx.ErrNoRows branch: a valid JWT for a
// credential whose users.students row doesn't exist (simulated here by
// deleting it directly -- in practice this only happens if student
// provisioning failed at register time, which register.go treats as
// non-fatal). Previously untested.
func TestIntegration_SubmitEssay_NoStudentRowReturns404(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	if _, err := pool.Exec(ctx, `DELETE FROM users.students WHERE id = $1`, studentID); err != nil {
		t.Fatalf("delete student row: %v", err)
	}

	w := submitEssay(t, r, token)
	if w.Code != http.StatusNotFound {
		t.Fatalf("submit essay with no backing student row: got %d body=%s", w.Code, w.Body.String())
	}
}

// TestIntegration_GetEssay_UnknownIDReturns404 covers the not-found branch
// of essay.Repository.scanSubmission (pgx.ErrNoRows -> apperrors.NotFound),
// which the graded-flow tests never exercise since they only ever fetch a
// submission they just created.
func TestIntegration_GetEssay_UnknownIDReturns404(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	getReq := httptest.NewRequest(http.MethodGet, "/v1/essays/00000000-0000-4000-a000-000000000999", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for an unknown essay id, got %d body=%s", getW.Code, getW.Body.String())
	}
}
