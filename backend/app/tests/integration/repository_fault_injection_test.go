package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	essayrepo "github.com/preuni/app/internal/essay/repository"
	gamificationrepo "github.com/preuni/app/internal/gamification/repository"
	socialrepo "github.com/preuni/app/internal/social/repository"
	streakrepo "github.com/preuni/app/internal/streak/repository"
	userrepo "github.com/preuni/app/internal/user/repository"
)

// TestIntegration_RepositoryFaultInjection_CanceledContextReturnsInternalError
// reaches the apperrors.Internal(err) branches across several repositories'
// DB calls by passing an already-canceled context.Context -- genuine pgx
// behavior (the same error a real client disconnect or request timeout
// produces in production), not a mock. These branches were previously
// unreachable via any success-path test.
func TestIntegration_RepositoryFaultInjection_CanceledContextReturnsInternalError(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("essay.ListByUser", func(t *testing.T) {
		repo := essayrepo.NewRepository(pool, streakrepo.NewRepository(), nil)
		_, err := repo.ListByUser(canceled, studentID)
		if err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})

	t.Run("gamification.EnsureCurrentWeekEntry", func(t *testing.T) {
		repo := gamificationrepo.NewRepository(pool)
		err := repo.EnsureCurrentWeekEntry(canceled, studentID, time.Now().UTC())
		if err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})

	t.Run("gamification.WeeklyLeaderboard", func(t *testing.T) {
		repo := gamificationrepo.NewRepository(pool)
		_, err := repo.WeeklyLeaderboard(canceled, "bronze", time.Now().UTC())
		if err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})

	t.Run("gamification.WeekClose", func(t *testing.T) {
		repo := gamificationrepo.NewRepository(pool)
		_, err := repo.WeekClose(canceled, time.Now().UTC())
		if err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})

	t.Run("social.ListFriends", func(t *testing.T) {
		repo := socialrepo.NewRepository(pool)
		_, err := repo.ListFriends(canceled, studentID)
		if err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})

	t.Run("user.Update", func(t *testing.T) {
		repo := userrepo.NewStudentRepository(pool)
		name := "Won't Persist"
		_, err := repo.Update(canceled, studentID, &userrepo.StudentPatch{DisplayName: &name})
		if err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})

	t.Run("essay.ReconcileOnce", func(t *testing.T) {
		repo := essayrepo.NewRepository(pool, streakrepo.NewRepository(), nil)
		if _, err := repo.ReconcileOnce(canceled); err == nil {
			t.Fatal("expected an error from a canceled context")
		}
	})
}

// TestIntegration_StreakRepository_ClosedTxReturnsInternalError reaches
// streak.Repository.GetForUpdate/RecordSubmission's apperrors.Internal(err)
// branches with a genuinely closed pgx.Tx (begun then immediately rolled
// back by the test) -- every real call site opens a fresh tx and passes it
// straight in, so these branches were previously unreachable.
func TestIntegration_StreakRepository_ClosedTxReturnsInternalError(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	repo := streakrepo.NewRepository()

	t.Run("GetForUpdate", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("rollback tx: %v", err)
		}
		if _, err := repo.GetForUpdate(ctx, tx, studentID); err == nil {
			t.Fatal("expected an error from a closed tx")
		}
	})

	t.Run("RecordSubmission", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("rollback tx: %v", err)
		}
		if _, err := repo.RecordSubmission(ctx, tx, studentID, time.Now().UTC()); err == nil {
			t.Fatal("expected an error from a closed tx")
		}
	})
}

// TestIntegration_Reconciler_MissingCorrectionsRowIsInternalError covers
// reconcileCompleted's row.Scan(err) branch: a correction_jobs row marked
// 'completed' with no matching correction.corrections row (an inconsistent
// state that shouldn't happen in practice, but the code has no special
// handling for it -- it falls straight into apperrors.Internal(err) same
// as any other DB error). Reached indirectly via ReconcileOnce since
// reconcileCompleted itself is unexported.
func TestIntegration_Reconciler_MissingCorrectionsRowIsInternalError(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	w := submitEssay(t, r, token)
	if w.Code != 202 {
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

	// Mark the job completed WITHOUT ever inserting a matching
	// correction.corrections row -- reconcileCompleted's QueryRow will
	// genuinely find nothing.
	if _, err := pool.Exec(ctx, `UPDATE correction.correction_jobs SET status = 'completed', completed_at = now() WHERE id = $1`, jobID); err != nil {
		t.Fatalf("mark job completed: %v", err)
	}

	essayRepo := essayrepo.NewRepository(pool, streakrepo.NewRepository(), nil)
	if _, err := essayRepo.ReconcileOnce(ctx); err == nil {
		t.Fatal("expected ReconcileOnce to surface reconcileCompleted's missing-corrections-row error")
	}
}

// TestIntegration_HandlerFaultInjection_CanceledContextReturnsErrorStatus
// covers several GET-list handlers' error-response branches by giving the
// *http.Request an already-canceled context before it reaches the router --
// the same real pgx failure OnboardingHandler's canceled-context test
// proved works at the HTTP layer, not just the repository layer directly.
func TestIntegration_HandlerFaultInjection_CanceledContextReturnsErrorStatus(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{"ListEssays", "GET", "/v1/essays"},
		{"GetStreak", "GET", "/v1/streaks/me"},
		{"ListFriends", "GET", "/v1/friends"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil).WithContext(canceled)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code == 200 {
				t.Fatalf("%s: expected an error status for a canceled context, got 200 body=%s", tc.name, w.Body.String())
			}
		})
	}
}

// TestIntegration_EssaySubmit_CanceledContextReturnsInternalError covers
// essay.Repository.Submit's apperrors.Internal(err) branch: Submit opens
// its own tx via r.db.Begin(ctx), exactly like reconciler.go, so a
// canceled context fails Begin itself -- genuine pgx behavior.
func TestIntegration_EssaySubmit_CanceledContextReturnsInternalError(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	repo := essayrepo.NewRepository(pool, streakrepo.NewRepository(), nil)
	_, err := repo.Submit(canceled, studentID, "Tema", "Contexto qualquer com mais de vinte caracteres.", "Texto de redação de teste para o fluxo de integração.")
	if err == nil {
		t.Fatal("expected an error from a canceled context")
	}
}

// TestIntegration_MoreHandlerFaultInjection_CanceledContextReturnsErrorStatus
// extends the canceled-context HTTP-layer technique to more handlers whose
// error-response branch was previously untested: DeleteStudent, DataExport,
// UpdateStudent (a PATCH with a real body, so the DB call -- not JSON
// decoding -- is what fails), gamification's WeeklyLeaderboard, and MyMedals.
func TestIntegration_MoreHandlerFaultInjection_CanceledContextReturnsErrorStatus(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("DeleteStudent", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/v1/students/me", nil).WithContext(canceled)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == 200 || w.Code == 204 {
			t.Fatalf("expected an error status for a canceled context, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("DataExport", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/students/me/data-export", nil).WithContext(canceled)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == 200 {
			t.Fatalf("expected an error status for a canceled context, got 200 body=%s", w.Body.String())
		}
	})

	t.Run("UpdateStudent", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"display_name": "Won't Persist"})
		req := httptest.NewRequest("PATCH", "/v1/students/me", bytes.NewReader(body)).WithContext(canceled)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == 200 {
			t.Fatalf("expected an error status for a canceled context, got 200 body=%s", w.Body.String())
		}
	})

	t.Run("WeeklyLeaderboard", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/ranking/weekly?tier=bronze", nil).WithContext(canceled)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == 200 {
			t.Fatalf("expected an error status for a canceled context, got 200 body=%s", w.Body.String())
		}
	})

	t.Run("MyMedals", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/medals/me", nil).WithContext(canceled)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == 200 {
			t.Fatalf("expected an error status for a canceled context, got 200 body=%s", w.Body.String())
		}
	})
}
