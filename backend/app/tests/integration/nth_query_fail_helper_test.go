package integration_test

import (
	"context"
	"os"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preuni/pkg/logger"
	"net/http"

	"github.com/preuni/app/internal/config"
	"github.com/preuni/app/internal/router"
)

// nthQueryFailTracer deterministically fails the Nth query (1-indexed,
// counted only on the pool it's attached to) with a real, non-flaky error.
//
// Confirmed by reading pgx v5.9.2's Conn.Query: the context returned by
// QueryTracer.TraceQueryStart is the context actually used for that
// query's execution. Returning an already-canceled context on the Nth
// call makes pgx genuinely fail that one query with "context canceled" --
// the same class of error a real client disconnect/timeout produces --
// while every other query on the same pool is unaffected. This reaches
// "first DB call succeeds, second DB call fails" branches that plain
// canceled-context fault injection (which fails the *first* call
// uniformly) can never target.
type nthQueryFailTracer struct {
	n     int64
	count int64
}

func (t *nthQueryFailTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	if atomic.AddInt64(&t.count, 1) == t.n {
		cctx, cancel := context.WithCancel(ctx)
		cancel()
		return cctx
	}
	return ctx
}

func (t *nthQueryFailTracer) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

// setupWithNthQueryFailure builds a router bound to its OWN pool (separate
// from setup(t)'s) whose Nth query fails. Fixtures (registering a user,
// seeding an OTP, etc.) should go through a plain setup(t) pool/router
// first so they don't count against this pool's query counter -- only
// requests made against the router returned here count.
func setupWithNthQueryFailure(t *testing.T, n int64) (http.Handler, *pgxpool.Pool) {
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

	cfg := config.Config{
		JWTSigningKey:        "test-signing-key-32-bytes-minimum!",
		JWTAccessExpirySec:   3600,
		JWTRefreshExpiryDays: 30,
		MailFromAddr:         "noreply@test",
		MailFromName:         "Test",
		StorageEndpoint:      "localhost:9000",
		StorageAccessKey:     "test",
		StorageSecretKey:     "test",
		StorageBucket:        "test-bucket",
	}
	r, _, _ := router.New(cfg, pool, logger.New("error"))
	return r, pool
}

// nthQueryFailPool builds a standalone pool (no router) whose Nth query
// fails, for exercising repository methods directly rather than through
// HTTP handlers -- e.g. essay.Repository.ReconcileOnce, which is called
// from a background loop, not a request.
func nthQueryFailPool(t *testing.T, n int64) *pgxpool.Pool {
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
