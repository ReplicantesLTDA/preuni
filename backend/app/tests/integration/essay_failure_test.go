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
	// as corretor-redacao's correction_repo.mark_failed would.
	if _, err := pool.Exec(ctx, `UPDATE correction.correction_jobs SET status = 'failed', completed_at = now() WHERE id = $1`, jobID); err != nil {
		t.Fatalf("mark job failed: %v", err)
	}

	essayRepo := essayrepo.NewRepository(pool, streakrepo.NewRepository())
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
