package integration_test

import (
	"context"
	"net/http"
	"testing"
)

// TestIntegration_Submit_LaterCallFailures covers essay.Repository.Submit's
// remaining failure branches beyond the already-tested "whole context
// canceled before the first query" case (which only reaches tx.Begin
// failing). Query order inside Submit for a first-ever submission on a
// brand-new user: 1. tx.Begin, 2. streakRepo.GetForUpdate's SELECT ... FOR
// UPDATE, 3. the alreadySubmitted EXISTS SELECT, 4. the essay_submissions
// INSERT, 5. streakRepo.RecordSubmission's UPDATE, 6. the correction_jobs
// INSERT, 7. tx.Commit. n=1/2 are already covered elsewhere (canceled-
// context and streak's closed-tx tests); n=3..7 were not.
func TestIntegration_Submit_LaterCallFailures(t *testing.T) {
	cases := []struct {
		name string
		n    int64
	}{
		{"already-submitted_check_fails", 3},
		{"insert-essay_submissions_fails", 4},
		{"record-submission_update_fails", 5},
		{"insert-correction_jobs_fails", 6},
		{"tx-commit_fails", 7},
	}

	fixtureR, fixturePool := setup(t)
	ctx := context.Background()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			studentID, token := registerTestUser(t, fixtureR)
			defer cleanupTestUser(ctx, t, fixturePool, studentID)

			failR, _ := setupWithNthQueryFailure(t, tc.n)
			w := submitEssay(t, failR, token)
			if w.Code != http.StatusInternalServerError {
				t.Fatalf("expected a 500 when query #%d fails, got %d body=%s", tc.n, w.Code, w.Body.String())
			}
		})
	}
}
