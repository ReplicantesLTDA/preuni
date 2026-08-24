package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	essayrepo "github.com/preuni/app/internal/essay/repository"
	streakrepo "github.com/preuni/app/internal/streak/repository"
)

// TestIntegration_ReconcileOnce_MarkTimedOutExecFails covers
// reconcileFailed/markTimedOut's Exec-fails branch: the initial pending-rows
// SELECT (query #1) must succeed to reach it, then the UPDATE inside
// markTimedOut (query #2) fails via the Nth-query tracer. Genuinely
// distinct from the already-covered "whole context canceled before the
// first query" case, which never reaches markTimedOut at all.
func TestIntegration_ReconcileOnce_MarkTimedOutExecFails(t *testing.T) {
	fixtureR, fixturePool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, fixtureR)
	defer cleanupTestUser(ctx, t, fixturePool, studentID)

	w := submitEssay(t, fixtureR, token)
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
	if err := fixturePool.QueryRow(ctx, `SELECT correction_job_id FROM essay.essay_submissions WHERE id = $1`, resp.ID).Scan(&jobID); err != nil {
		t.Fatalf("read correction_job_id: %v", err)
	}
	if _, err := fixturePool.Exec(ctx, `UPDATE correction.correction_jobs SET status = 'failed', completed_at = now() WHERE id = $1`, jobID); err != nil {
		t.Fatalf("mark job failed: %v", err)
	}

	// Query order inside ReconcileOnce for this single pending 'failed'
	// row: 1. the pending-rows SELECT, 2. markTimedOut's UPDATE Exec.
	failPool := nthQueryFailPool(t, 2)
	essayRepo := essayrepo.NewRepository(failPool, streakrepo.NewRepository(), nil)
	if _, err := essayRepo.ReconcileOnce(ctx); err == nil {
		t.Fatal("expected ReconcileOnce to surface markTimedOut's Exec failure")
	}
}

// TestIntegration_ReconcileOnce_ReconcileCompleted_LaterCallFailures covers
// reconcileCompleted's tx.Begin-fails, tx.Exec(insert essay_grades)-fails,
// and tx.Exec(update submission)-fails branches -- three genuinely distinct
// gaps found by probing which Nth query each one falls on (query #1 is the
// pending-rows SELECT, #2 is the final_score/competencies QueryRow, #3 is
// tx.Begin's implicit BEGIN, #4 is the essay_grades INSERT, #5 is the
// submissions UPDATE). Confirmed via a coverage-diff probe before writing
// this: n=3 hits the Begin-fails branch, n=4 the insert-fails branch, n=5
// the update-fails branch, and n=6 (tx.Commit) hits nothing new the
// success-path test doesn't already cover.
func TestIntegration_ReconcileOnce_ReconcileCompleted_LaterCallFailures(t *testing.T) {
	cases := []struct {
		name string
		n    int64
	}{
		{"tx-begin_fails", 3},
		{"insert-essay_grades_fails", 4},
		{"update-submission_status_fails", 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixtureR, fixturePool := setup(t)
			ctx := context.Background()
			studentID, token := registerTestUser(t, fixtureR)
			defer cleanupTestUser(ctx, t, fixturePool, studentID)

			w := submitEssay(t, fixtureR, token)
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
			if err := fixturePool.QueryRow(ctx, `SELECT correction_job_id FROM essay.essay_submissions WHERE id = $1`, resp.ID).Scan(&jobID); err != nil {
				t.Fatalf("read correction_job_id: %v", err)
			}
			_, err := fixturePool.Exec(ctx, `
				INSERT INTO correction.corrections
					(id, user_id, job_id, essay_text, prompt_theme_title, prompt_theme_context, input_hash,
					 status, queued_at, completed_at, final_score, c1_score, c2_score, c3_score, c4_score, c5_score,
					 competencies, eliminatory_flags, prompt_version, model_identifier, output_schema_version, quota_consumed)
				VALUES
					(gen_random_uuid(), $1, $2, 'texto', 'tema', 'contexto', decode('00','hex'),
					 'completed', now(), now(), 800, 160, 160, 160, 160, 160,
					 '[]'::jsonb, '[]'::jsonb, '1.0.0', 'test-model', 'v1', true)
			`, studentID, jobID)
			if err != nil {
				t.Fatalf("simulate worker completion: %v", err)
			}
			if _, err := fixturePool.Exec(ctx, `UPDATE correction.correction_jobs SET status = 'completed', completed_at = now() WHERE id = $1`, jobID); err != nil {
				t.Fatalf("mark job completed: %v", err)
			}

			failPool := nthQueryFailPool(t, tc.n)
			essayRepo := essayrepo.NewRepository(failPool, streakrepo.NewRepository(), nil)
			if _, err := essayRepo.ReconcileOnce(ctx); err == nil {
				t.Fatalf("expected ReconcileOnce to surface a failure at query #%d", tc.n)
			}
		})
	}
}
