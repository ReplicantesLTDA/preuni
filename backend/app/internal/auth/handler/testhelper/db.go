// Package testhelper provides shared utilities for handler integration tests.
// Tests that call SetupTestDB require a running PostgreSQL instance reachable
// via the TEST_DB_URL environment variable. When that variable is absent the
// calling test is skipped gracefully.
package testhelper

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupTestDB opens a pgxpool connection to the database at TEST_DB_URL,
// runs all SQL migrations from infra/migrations/auth/ (in filename order),
// and registers a cleanup function that truncates every auth table after
// the test completes. The test is skipped when TEST_DB_URL is not set.
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set — skipping integration test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("testhelper: connect to test DB: %v", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("testhelper: ping test DB: %v", err)
	}

	runMigrations(t, ctx, pool)

	t.Cleanup(func() {
		cleanupTables(t, pool)
		pool.Close()
	})

	return pool
}

// runMigrations executes all *.sql files in infra/migrations/auth/ sorted by
// filename. Tests run from backend/app/internal/auth/handler/ so the relative
// path walks five directories up to the repository root.
func runMigrations(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	migrationsDir := filepath.Join("..", "..", "..", "..", "..", "infra", "migrations", "auth")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("testhelper: read migrations dir %q: %v", migrationsDir, err)
	}

	var sqlFiles []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			sqlFiles = append(sqlFiles, filepath.Join(migrationsDir, e.Name()))
		}
	}
	sort.Strings(sqlFiles)

	for _, path := range sqlFiles {
		sql, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("testhelper: read migration %q: %v", path, err)
		}
		if _, err = pool.Exec(ctx, string(sql)); err != nil {
			// Migrations may already exist — ignore "already exists" errors so
			// the helper is safe to call in each test function.
			t.Logf("testhelper: migration %q (may be already applied): %v", filepath.Base(path), err)
		}
	}
}

// cleanupTables truncates all auth tables so each test starts with a clean slate.
func cleanupTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	tables := []string{
		"auth.otp_codes",
		"auth.refresh_tokens",
		"auth.credentials",
	}
	for _, tbl := range tables {
		if _, err := pool.Exec(ctx, "TRUNCATE "+tbl+" CASCADE"); err != nil {
			t.Logf("testhelper: truncate %s: %v", tbl, err)
		}
	}
}
