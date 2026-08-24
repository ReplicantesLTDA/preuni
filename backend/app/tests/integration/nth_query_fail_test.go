package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	essayrepo "github.com/preuni/app/internal/essay/repository"
	streakrepo "github.com/preuni/app/internal/streak/repository"
)

// These tests reach "first DB call succeeds, second (or later) DB call
// fails" branches using setupWithNthQueryFailure (see
// nth_query_fail_helper_test.go for how and why this is deterministic,
// not timing-based). Fixtures are created via a plain setup(t) router so
// they don't count against the traced pool's query counter -- only the
// single request under test, against a fresh router bound to the traced
// pool, does.

// TestIntegration_RefreshToken_SecondCallFailures covers
// RefreshTokenHandler's FindByID/Store/Revoke-fails branches (creds
// lookup, new-token persistence, old-token revocation) -- previously
// unreachable since only FindByHash's failure (the *first* call) could
// be fault-injected.
func TestIntegration_RefreshToken_SecondCallFailures(t *testing.T) {
	fixtureR, fixturePool := setup(t)
	ctx := context.Background()

	newRefreshToken := func(t *testing.T) string {
		t.Helper()
		email := studentEmailPlaceholder()
		regBody, _ := json.Marshal(map[string]string{
			"email": email, "password": "P@ssw0rd123", "display_name": "NthQuery",
		})
		regReq := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(regBody))
		regReq.Header.Set("Content-Type", "application/json")
		regW := httptest.NewRecorder()
		fixtureR.ServeHTTP(regW, regReq)
		if regW.Code != http.StatusCreated {
			t.Fatalf("register: got %d body=%s", regW.Code, regW.Body.String())
		}
		var reg struct {
			StudentID    string `json:"student_id"`
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.Unmarshal(regW.Body.Bytes(), &reg); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { cleanupTestUser(ctx, t, fixturePool, reg.StudentID) })
		return reg.RefreshToken
	}

	// Query order inside RefreshTokenHandler.ServeHTTP:
	//   1. refreshRepo.FindByHash
	//   2. credRepo.FindByID
	//   3. refreshRepo.Store
	//   4. refreshRepo.Revoke
	cases := []struct {
		name string
		n    int64
	}{
		{"FindByID fails (account not found)", 2},
		{"Store fails (new token persistence)", 3},
		{"Revoke fails (old token rotation)", 4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			token := newRefreshToken(t)
			r, _ := setupWithNthQueryFailure(t, tc.n)

			body, _ := json.Marshal(map[string]string{"refresh_token": token})
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code == http.StatusOK {
				t.Fatalf("expected an error status when query #%d fails, got 200 body=%s", tc.n, w.Body.String())
			}
		})
	}
}

// TestIntegration_Logout_RevokeFails covers LogoutHandler's Revoke-fails
// branch (query #2 -- FindByHash, the first call, must succeed to even
// reach it).
func TestIntegration_Logout_RevokeFails(t *testing.T) {
	fixtureR, fixturePool := setup(t)
	ctx := context.Background()

	email := "logout-nthq+" + time.Now().Format("20060102150405.000000000") + "@preuni.test"
	regBody, _ := json.Marshal(map[string]string{
		"email": email, "password": "P@ssw0rd123", "display_name": "NthQuery",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	fixtureR.ServeHTTP(regW, regReq)
	if regW.Code != http.StatusCreated {
		t.Fatalf("register: got %d body=%s", regW.Code, regW.Body.String())
	}
	var reg struct {
		StudentID    string `json:"student_id"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(regW.Body.Bytes(), &reg); err != nil {
		t.Fatal(err)
	}
	defer cleanupTestUser(ctx, t, fixturePool, reg.StudentID)

	// Query order inside LogoutHandler.ServeHTTP: 1. FindByHash, 2. Revoke.
	r, _ := setupWithNthQueryFailure(t, 2)

	body, _ := json.Marshal(map[string]string{"refresh_token": reg.RefreshToken})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+reg.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusNoContent {
		t.Fatalf("expected an error status when Revoke fails, got 204")
	}
}

// TestIntegration_Onboarding_FindByIDFails covers OnboardingHandler's
// FindByID-fails branch (query #2 -- SetOnboardingCompleted, the first
// call, must succeed to even reach it).
func TestIntegration_Onboarding_FindByIDFails(t *testing.T) {
	fixtureR, fixturePool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, fixtureR)
	defer cleanupTestUser(ctx, t, fixturePool, studentID)

	// Query order inside OnboardingHandler.ServeHTTP:
	//   1. studentRepo.SetOnboardingCompleted
	//   2. studentRepo.FindByID
	r, _ := setupWithNthQueryFailure(t, 2)

	req := httptest.NewRequest(http.MethodPatch, "/v1/students/me/onboarding", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Fatalf("expected an error status when the post-update FindByID fails, got 200 body=%s", w.Body.String())
	}

	// SetOnboardingCompleted (query #1) genuinely ran and committed before
	// query #2 failed -- confirm the side effect actually happened, this
	// isn't testing a no-op.
	var onboardingCompleted bool
	if err := fixturePool.QueryRow(ctx, `SELECT onboarding_completed FROM users.students WHERE id = $1`, studentID).Scan(&onboardingCompleted); err != nil {
		t.Fatal(err)
	}
	if !onboardingCompleted {
		t.Fatal("expected onboarding_completed = true even though the handler's response was an error")
	}
}

// TestIntegration_GetEssay_GetGradeFails covers essay.Repository.getGrade's
// apperrors.Internal(err) branch (query #2 inside GetByID -- scanSubmission,
// the first call, must succeed and find a 'graded' submission to even
// reach it). Previously only getGrade's ErrNoRows branch was reachable.
func TestIntegration_GetEssay_GetGradeFails(t *testing.T) {
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
	competencies := `[
		{"competency":1,"score":160,"justification_pt_br":"ok","excerpt":"trecho 1"},
		{"competency":2,"score":160,"justification_pt_br":"ok","excerpt":"trecho 2"},
		{"competency":3,"score":160,"justification_pt_br":"ok","excerpt":"trecho 3"},
		{"competency":4,"score":160,"justification_pt_br":"ok","excerpt":"trecho 4"},
		{"competency":5,"score":160,"justification_pt_br":"ok","excerpt":"trecho 5"}
	]`
	if _, err := fixturePool.Exec(ctx, `
		INSERT INTO correction.corrections
			(id, user_id, job_id, essay_text, prompt_theme_title, prompt_theme_context, input_hash,
			 status, queued_at, completed_at, final_score, c1_score, c2_score, c3_score, c4_score, c5_score,
			 competencies, eliminatory_flags, prompt_version, model_identifier, output_schema_version, quota_consumed)
		VALUES
			(gen_random_uuid(), $1, $2, 'texto', 'tema', 'contexto', decode('00','hex'),
			 'completed', now(), now(), 800, 160, 160, 160, 160, 160,
			 $3::jsonb, '[]'::jsonb, '1.0.0', 'test-model', 'v1', true)
	`, studentID, jobID, competencies); err != nil {
		t.Fatalf("simulate worker completion: %v", err)
	}
	if _, err := fixturePool.Exec(ctx, `UPDATE correction.correction_jobs SET status = 'completed', completed_at = now() WHERE id = $1`, jobID); err != nil {
		t.Fatalf("mark job completed: %v", err)
	}

	essayRepo := essayrepo.NewRepository(fixturePool, streakrepo.NewRepository(), nil)
	if _, err := essayRepo.ReconcileOnce(ctx); err != nil {
		t.Fatalf("ReconcileOnce: %v", err)
	}

	// Query order inside GetByID (once status='graded'):
	//   1. scanSubmission's own SELECT
	//   2. getGrade's SELECT
	r, _ := setupWithNthQueryFailure(t, 2)

	getReq := httptest.NewRequest(http.MethodGet, "/v1/essays/"+resp.ID, nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	if getW.Code == http.StatusOK {
		t.Fatalf("expected an error status when getGrade's query fails, got 200 body=%s", getW.Body.String())
	}
}

// TestIntegration_VerifyEmail_LaterCallFailures covers VerifyEmailHandler's
// MarkUsed/MarkEmailVerified-fails branches (queries #3 and #4 -- FindByEmail
// and FindActiveByCredentialAndPurpose, the first two calls, must both
// succeed with a real active OTP to even reach them).
func TestIntegration_VerifyEmail_LaterCallFailures(t *testing.T) {
	fixtureR, fixturePool := setup(t)
	ctx := context.Background()

	// Query order inside VerifyEmailHandler.ServeHTTP:
	//   1. credRepo.FindByEmail
	//   2. otpRepo.FindActiveByCredentialAndPurpose
	//   3. otpRepo.MarkUsed
	//   4. credRepo.MarkEmailVerified
	cases := []struct {
		name string
		n    int64
	}{
		{"MarkUsed fails", 3},
		{"MarkEmailVerified fails", 4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			studentID, _ := registerTestUser(t, fixtureR)
			t.Cleanup(func() { cleanupTestUser(ctx, t, fixturePool, studentID) })
			email := studentEmail(t, ctx, fixturePool, studentID)

			otp := "246813"
			seedOTP(t, ctx, fixturePool, studentID, "EMAIL_VERIFY", otp)

			r, _ := setupWithNthQueryFailure(t, tc.n)

			body, _ := json.Marshal(map[string]string{"email": email, "otp": otp})
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/email/verify", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code == http.StatusNoContent {
				t.Fatalf("expected an error status when query #%d fails, got 204", tc.n)
			}
		})
	}
}

// studentEmailPlaceholder generates a unique test email -- kept separate
// from the package's other email-generation helpers to avoid colliding
// with fmt.Sprintf-based counters used elsewhere in parallel subtests.
func studentEmailPlaceholder() string {
	return "nthq+" + time.Now().Format("20060102150405.000000000") + "@preuni.test"
}
