package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestIntegration_SubmitEssay_FreeTierSecondSubmissionBlocked covers spec
// User Story 1, acceptance scenario 3: a free-tier user's second same-UTC-day
// submission is rejected with 429 quota_exhausted-equivalent, and the first
// submission is unaffected.
func TestIntegration_SubmitEssay_FreeTierSecondSubmissionBlocked(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	first := submitEssay(t, r, token)
	if first.Code != http.StatusAccepted {
		t.Fatalf("first submit: got %d body=%s", first.Code, first.Body.String())
	}

	second := submitEssay(t, r, token)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second same-day submit: got %d, want 429, body=%s", second.Code, second.Body.String())
	}
	var errBody struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(second.Body.Bytes(), &errBody); err != nil {
		t.Fatal(err)
	}
	if errBody.Error.Code != "QUOTA_EXCEEDED" {
		t.Fatalf("expected error code QUOTA_EXCEEDED, got %q", errBody.Error.Code)
	}

	// Only one submission must exist.
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM essay.essay_submissions WHERE user_id = $1`, studentID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 submission row after the blocked second attempt, got %d", count)
	}
}

// TestIntegration_SubmitEssay_ProTierSecondSubmissionAllowed covers spec
// User Story 1, acceptance scenario 4.
func TestIntegration_SubmitEssay_ProTierSecondSubmissionAllowed(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	if _, err := pool.Exec(ctx, `UPDATE users.students SET subscription_tier = 'pro' WHERE id = $1`, studentID); err != nil {
		t.Fatalf("upgrade to pro: %v", err)
	}

	first := submitEssay(t, r, token)
	if first.Code != http.StatusAccepted {
		t.Fatalf("first submit: got %d body=%s", first.Code, first.Body.String())
	}
	second := submitEssay(t, r, token)
	if second.Code != http.StatusAccepted {
		t.Fatalf("Pro tier second same-day submit: got %d, want 202, body=%s", second.Code, second.Body.String())
	}

	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM essay.essay_submissions WHERE user_id = $1`, studentID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected 2 submission rows for a Pro user, got %d", count)
	}

	// Streak must still only be 1 -- multiple same-day submissions don't
	// inflate the streak (streak/domain.NextStreak's idempotency rule).
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
		t.Fatalf("expected current_streak=1 after two same-day Pro submissions, got %d", streak.CurrentStreak)
	}
}
