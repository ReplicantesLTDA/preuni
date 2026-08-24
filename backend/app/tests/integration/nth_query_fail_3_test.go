package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	gamificationrepo "github.com/preuni/app/internal/gamification/repository"
)

// TestIntegration_Login_InvalidJSONIsRejected covers LoginHandler's
// decode-error branch, previously untested (only the "missing
// email/password" and various credential-lookup branches were).
func TestIntegration_Login_InvalidJSONIsRejected(t *testing.T) {
	r, _ := setup(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader([]byte("{not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("got %d, want 422, body=%s", w.Code, w.Body.String())
	}
}

// TestIntegration_Login_StoreFails covers LoginHandler's
// refreshRepo.Store-fails branch (query #2 -- credRepo.FindByEmail, the
// first call, must succeed to even reach it).
func TestIntegration_Login_StoreFails(t *testing.T) {
	fixtureR, fixturePool := setup(t)
	ctx := context.Background()

	email := "login-nthq+" + time.Now().Format("20060102150405.000000000") + "@preuni.test"
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
		StudentID string `json:"student_id"`
	}
	if err := json.Unmarshal(regW.Body.Bytes(), &reg); err != nil {
		t.Fatal(err)
	}
	defer cleanupTestUser(ctx, t, fixturePool, reg.StudentID)
	markVerified(t, ctx, fixturePool, reg.StudentID)

	// Query order inside LoginHandler.ServeHTTP: 1. credRepo.FindByEmail,
	// 2. refreshRepo.Store.
	r, _ := setupWithNthQueryFailure(t, 2)

	body, _ := json.Marshal(map[string]string{"email": email, "password": "P@ssw0rd123"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Fatalf("expected an error status when Store fails, got 200 body=%s", w.Body.String())
	}
}

// TestIntegration_ChangeEmailRequest_OTPCreateFails covers
// ChangeEmailRequestHandler's otpRepo.Create-fails branch (the first, and
// only, DB call in this handler -- plain canceled-context injection
// reaches it directly).
func TestIntegration_ChangeEmailRequest_OTPCreateFails(t *testing.T) {
	r, pool := setup(t)
	ctx := context.Background()
	studentID, token := registerTestUser(t, r)
	defer cleanupTestUser(ctx, t, pool, studentID)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	body, _ := json.Marshal(map[string]string{"new_email": "wontwork@preuni.test"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/email/change/request", bytes.NewReader(body)).WithContext(canceled)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusAccepted {
		t.Fatalf("expected an error status for a canceled context, got 202")
	}
}

// TestIntegration_ResendVerification_OTPCreateFails covers
// ResendVerificationHandler's otpRepo.Create-fails branch (query #2 --
// credRepo.FindByEmail, the first call, must succeed to even reach it).
func TestIntegration_ResendVerification_OTPCreateFails(t *testing.T) {
	fixtureR, fixturePool := setup(t)
	ctx := context.Background()
	studentID, _ := registerTestUser(t, fixtureR)
	defer cleanupTestUser(ctx, t, fixturePool, studentID)
	email := studentEmail(t, ctx, fixturePool, studentID)

	// Query order inside ResendVerificationHandler.ServeHTTP:
	//   1. credRepo.FindByEmail, 2. otpRepo.Create.
	r, _ := setupWithNthQueryFailure(t, 2)

	body, _ := json.Marshal(map[string]string{"email": email})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/email/verify-resend", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusNoContent {
		t.Fatalf("expected an error status when otpRepo.Create fails, got 204")
	}
}

// tracedPool builds a raw *pgxpool.Pool (no router) whose Nth query fails,
// for reaching repository methods not exposed over HTTP (e.g. WeekClose,
// which only runs from a scheduled job in cmd/server).
func tracedPool(t *testing.T, n int64) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://preuni:preuni@localhost:5432/preuni?sslmode=disable"
	}
	pgCfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		t.Fatalf("parse db url: %v", err)
	}
	pgCfg.ConnConfig.Tracer = &nthQueryFailTracer{n: n}
	pool, err := pgxpool.NewWithConfig(context.Background(), pgCfg)
	if err != nil {
		t.Skipf("postgres unreachable: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Skipf("postgres ping failed: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// TestIntegration_WeekClose_CloseOneEntryLaterCallFailures covers
// closeOneEntry's rank-update and next-week-insert Exec-fail branches
// (queries #2 and #3 -- query #1 is the tier's own SELECT, already
// covered by a canceled-context test that fails on the very first call).
func TestIntegration_WeekClose_CloseOneEntryLaterCallFailures(t *testing.T) {
	fixtureR, fixturePool := setup(t)
	ctx := context.Background()

	var userIDs []string
	for i := 0; i < 5; i++ {
		id, _ := registerTestUser(t, fixtureR)
		userIDs = append(userIDs, id)
	}
	defer func() {
		for _, id := range userIDs {
			cleanupTestUser(ctx, t, fixturePool, id)
		}
	}()

	now := time.Now().UTC()
	lastWeekStart := now.AddDate(0, 0, -7)
	offset := (int(lastWeekStart.Weekday()) + 6) % 7
	monday := lastWeekStart.AddDate(0, 0, -offset)
	closedWeek := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)

	seed := func(t *testing.T) {
		t.Helper()
		scores := []int{500, 400, 300, 200, 100}
		for i, uid := range userIDs {
			if _, err := fixturePool.Exec(ctx, `
				INSERT INTO gamification.weekly_ranking_entries (id, user_id, week_start, weekly_score, league_tier)
				VALUES (gen_random_uuid(), $1, $2, $3, 'bronze')
				ON CONFLICT (user_id, week_start) DO UPDATE SET weekly_score = EXCLUDED.weekly_score, rank_in_tier = NULL
			`, uid, closedWeek, scores[i]); err != nil {
				t.Fatalf("seed ranking entry %d: %v", i, err)
			}
		}
	}

	cases := []struct {
		name string
		n    int64
	}{
		{"rank update fails", 2},
		{"next-week insert fails", 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			seed(t)
			pool := tracedPool(t, tc.n)
			repo := gamificationrepo.NewRepository(pool)
			if _, err := repo.WeekClose(ctx, now); err == nil {
				t.Fatalf("expected an error when query #%d fails", tc.n)
			}
		})
	}
}
