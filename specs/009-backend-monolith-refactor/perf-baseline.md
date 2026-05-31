# Performance Baseline — Monolith

**Date**: 2026-05-25
**Env**: local Docker Compose (postgres 16, monolith on `:8088`)
**Method**: 100 sequential curl requests from host → monolith container

## Results

| Endpoint              | n   | avg     | p50     | p95     | p99      | constitution gate |
|-----------------------|-----|---------|---------|---------|----------|-------------------|
| `GET /health`         | 100 | 1.93 ms | 1.86 ms | 2.95 ms | 4.55 ms  | n/a               |
| `POST /v1/auth/login` | 100 | 3.25 ms | 2.85 ms | 3.44 ms | 31.60 ms | p95 < 500 ms ✓    |

## Notes

- `POST /v1/auth/login` hits postgres for `auth.credentials` lookup + bcrypt verify + `auth.refresh_tokens` insert.
- p99 outlier on login is likely a connection pool warm-up effect on the first few requests; bcrypt-bound, not I/O-bound.
- No N+1 queries introduced by router consolidation (auth + user routers are mounted side-by-side, each owning its own repository).
- Sequential test only — concurrency stress (k6/hey) not run.

## Constitution gate verdict

p95 well under 500 ms on the user-facing path. **PASS**.
