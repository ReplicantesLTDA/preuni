// Package integration_test exercises the full essay submission -> streak ->
// correction bridge -> reconciliation loop against a real Postgres. Set
// TEST_DB_URL to run (see backend/app/tests/integration/auth_flow_test.go
// for the shared setup() helper).
package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	essayrepo "github.com/preuni/app/internal/essay/repository"
	streakrepo "github.com/preuni/app/internal/streak/repository"
)

// registerTestUser registers a brand-new user via the real /v1/auth/register
// endpoint and returns their student id + access token, so essay tests
// exercise the same users.students row essay/streak repos read.
func registerTestUser(t *testing.T, r http.Handler) (studentID, accessToken string) {
	t.Helper()
	email := fmt.Sprintf("essay-integ+%d@preuni.test", time.Now().UnixNano())
	body, _ := json.Marshal(map[string]string{
		"email": email, "password": "P@ssw0rd123", "display_name": "Essay Integ",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: got %d body=%s", w.Code, w.Body.String())
	}
	var reg struct {
		StudentID   string `json:"student_id"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &reg); err != nil {
		t.Fatal(err)
	}
	return reg.StudentID, reg.AccessToken
}

func submitEssay(t *testing.T, r http.Handler, accessToken string) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"prompt_theme_title":   "A importância da leitura",
		"prompt_theme_context": "Discuta o papel da leitura na formação do cidadão.",
		"essay_text":           "Texto de redação de teste para o fluxo de integração.",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/essays", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestIntegration_SubmitEssay_AcceptsImmediatelyAndIncrementsStreak covers
// spec User Story 1, acceptance scenario 1: submission is accepted
// immediately (202), and the streak increments in the same transaction —
// well before any grading could plausibly finish.
func TestIntegration_SubmitEssay_AcceptsImmediatelyAndIncrementsStreak(t *testing.T) {
	r, pool := setup(t)
	studentID, token := registerTestUser(t, r)

	w := submitEssay(t, r, token)
	if w.Code != http.StatusAccepted {
		t.Fatalf("submit: got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Status != "pending" {
		t.Fatalf("expected status=pending immediately after submit, got %q", resp.Status)
	}

	streakReq := httptest.NewRequest(http.MethodGet, "/v1/streaks/me", nil)
	streakReq.Header.Set("Authorization", "Bearer "+token)
	streakW := httptest.NewRecorder()
	r.ServeHTTP(streakW, streakReq)
	if streakW.Code != http.StatusOK {
		t.Fatalf("get streak: got %d body=%s", streakW.Code, streakW.Body.String())
	}
	var streak struct {
		CurrentStreak int `json:"current_streak"`
	}
	if err := json.Unmarshal(streakW.Body.Bytes(), &streak); err != nil {
		t.Fatal(err)
	}
	if streak.CurrentStreak != 1 {
		t.Fatalf("expected current_streak=1 right after the first-ever submission, got %d", streak.CurrentStreak)
	}

	cleanupTestUser(context.Background(), t, pool, studentID)
}

// TestIntegration_SubmitEssay_ReconciliationGradesTheSubmission simulates
// the correction worker completing the job (writes correction.corrections
// directly, as the Python worker would) and verifies the monolith's
// reconciler picks it up and produces a graded essay_grades row.
func TestIntegration_SubmitEssay_ReconciliationGradesTheSubmission(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)

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

	// Simulate the correction worker: claim + complete the job, exactly as
	// corretor-redacao's correction_repo.claim_next/mark_completed would.
	competencies := `[
		{"competency":1,"score":160,"justification_pt_br":"Bom domínio da norma culta.","excerpt":"trecho 1"},
		{"competency":2,"score":160,"justification_pt_br":"Compreende bem o tema.","excerpt":"trecho 2"},
		{"competency":3,"score":160,"justification_pt_br":"Argumentação consistente.","excerpt":"trecho 3"},
		{"competency":4,"score":160,"justification_pt_br":"Boa coesão textual.","excerpt":"trecho 4"},
		{"competency":5,"score":160,"justification_pt_br":"Proposta de intervenção clara.","excerpt":"trecho 5"}
	]`
	_, err := pool.Exec(ctx, `
		INSERT INTO correction.corrections
			(id, user_id, job_id, essay_text, prompt_theme_title, prompt_theme_context, input_hash,
			 status, queued_at, completed_at, final_score, c1_score, c2_score, c3_score, c4_score, c5_score,
			 competencies, eliminatory_flags, prompt_version, model_identifier, output_schema_version, quota_consumed)
		VALUES
			(gen_random_uuid(), $1, $2, 'texto', 'tema', 'contexto', decode('00','hex'),
			 'completed', now(), now(), 800, 160, 160, 160, 160, 160,
			 $3::jsonb, '[]'::jsonb, '1.0.0', 'test-model', 'v1', true)
	`, studentID, jobID, competencies)
	if err != nil {
		t.Fatalf("simulate worker completion: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE correction.correction_jobs SET status = 'completed', completed_at = now() WHERE id = $1`, jobID); err != nil {
		t.Fatalf("mark job completed: %v", err)
	}

	essayRepo := essayrepo.NewRepository(pool, streakrepo.NewRepository(), nil)
	if _, err := essayRepo.ReconcileOnce(ctx); err != nil {
		t.Fatalf("ReconcileOnce: %v", err)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/essays/"+resp.ID, nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("get essay: got %d body=%s", getW.Code, getW.Body.String())
	}
	var graded struct {
		Status string `json:"status"`
		Grade  *struct {
			OverallScore int `json:"overall_score"`
			Competencies []struct {
				Competency int `json:"competency"`
			} `json:"competencies"`
		} `json:"grade"`
	}
	if err := json.Unmarshal(getW.Body.Bytes(), &graded); err != nil {
		t.Fatal(err)
	}
	if graded.Status != "graded" {
		t.Fatalf("expected status=graded after reconciliation, got %q", graded.Status)
	}
	if graded.Grade == nil || graded.Grade.OverallScore != 800 || len(graded.Grade.Competencies) != 5 {
		t.Fatalf("expected a full 5-competency grade with overall_score=800, got %+v", graded.Grade)
	}

	cleanupTestUser(ctx, t, pool, studentID)
}

// cleanupTestUser removes everything this test file creates for one test
// user, cascading through essay/streak/auth data (mirrors auth_flow_test.go's
// inline cleanup).
func cleanupTestUser(ctx context.Context, t *testing.T, pool *pgxpool.Pool, studentID string) {
	t.Helper()
	_, _ = pool.Exec(ctx, `DELETE FROM essay.essay_grades WHERE submission_id IN (SELECT id FROM essay.essay_submissions WHERE user_id = $1)`, studentID)
	_, _ = pool.Exec(ctx, `DELETE FROM essay.essay_submissions WHERE user_id = $1`, studentID)
	_, _ = pool.Exec(ctx, `DELETE FROM correction.corrections WHERE user_id = $1`, studentID)
	_, _ = pool.Exec(ctx, `DELETE FROM correction.correction_jobs WHERE user_id = $1`, studentID)
	_, _ = pool.Exec(ctx, `DELETE FROM auth.refresh_tokens WHERE credential_id = $1`, studentID)
	_, _ = pool.Exec(ctx, `DELETE FROM auth.otp_codes WHERE credential_id = $1`, studentID)
	_, _ = pool.Exec(ctx, `DELETE FROM auth.credentials WHERE id = $1`, studentID)
	_, _ = pool.Exec(ctx, `DELETE FROM users.students WHERE id = $1`, studentID)
}
