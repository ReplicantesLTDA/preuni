# Implementation Plan: Backend Monolith Refactor

**Branch**: `009-backend-monolith-refactor` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/009-backend-monolith-refactor/spec.md`

## Summary

Refactor the backend from multiple deployable services into a single monolith binary while preserving the existing HTTP API surface exposed by the gateway. Email sending moves from the Elixir mail service into the unified backend while keeping the same internal send contract and supported email types.

Key outcomes:
- One backend release artifact can serve all current `/v1/*` routes.
- Existing clients remain compatible (no client updates required).
- Auth/User flows and transactional emails work end-to-end.

Primary decisions and tradeoffs are documented in [research.md](./research.md).

## Technical Context

**Language/Version**: Go 1.24 (per `go.work`/`go.mod`); Elixir mail service currently exists and will be replaced for this feature
**Primary Dependencies**: go-chi/chi (HTTP routing), pgx/v5 (PostgreSQL), golang-jwt/jwt (JWT), zap (logging), testify (tests), NGINX (gateway routing + rate limiting)
**Storage**: PostgreSQL 16 (auth + users schemas), Redis 7 (existing infra), S3-compatible object storage (avatar flows)
**Testing**: `go test` + `testify`, `httptest` for handler-level tests, integration tests against local Postgres via `TEST_DB_URL`
**Target Platform**: Linux server (containerized) + local development via `docker compose`
**Project Type**: HTTP API backend consolidation (multi-service → monolith)
**Performance Goals**: Preserve constitution target p95 < 500 ms for user-facing endpoints; no new N+1 query patterns introduced by consolidation
**Constraints**: Backward-compatible routes + error envelopes; preserve JWT semantics; internal endpoints remain protected by the internal token; migration must be incremental with a rollback path
**Scale/Scope**: New monolith entrypoint + router extraction from auth/user + mail rewrite to Go + infra updates (compose + gateway routing) + contract/integration tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Readability / single responsibility | PASS | Monolith is assembled from domain packages and router builders; no single “god” package for all concerns |
| No dead code | PASS | Remove or quarantine legacy entrypoints only after cutover is verified |
| Unit tests for business logic | PASS | Mail validation + auth/user domain logic covered by unit tests |
| Integration tests for API boundaries | PASS | Add monolith HTTP integration coverage and keep existing auth/user integration tests |
| Performance requirements | PASS | Keep handler + DB access patterns intact during cutover; treat regressions as blockers |
| Coverage floor ≥ 80% | PASS | New tests are additive; consolidation work must not reduce overall coverage |

Post-design re-check: still PASS, no constitution violations introduced.

## Project Structure

### Documentation (this feature)

```text
specs/009-backend-monolith-refactor/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── backend-monolith-api-contract.md
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── go.work
├── pkg/                           # shared config/errors/logger/middleware
└── svc/
    ├── auth/                      # existing module; extract router builder for monolith reuse
    ├── user/                      # existing module; extract router builder for monolith reuse
    ├── content/                   # currently health-only
    ├── learning/                  # currently health-only
    ├── simulation/                # currently health-only
    ├── dissertation/              # currently health-only
    ├── notification/              # currently health-only
    └── monolith/                  # NEW: unified backend binary

infra/
├── docker-compose.yml             # add monolith service, remove mail service after cutover
└── nginx/nginx.conf               # route /v1/* prefixes to monolith after cutover
```

**Structure Decision**: Create a new Go module `backend/svc/monolith` that builds a single HTTP server by mounting the existing auth/user routers (and later other routers) under the same route prefixes as today. Keep the existing service modules runnable during transition for rollback.

## Implementation Reference

### Phase A — Monolith skeleton and route compatibility

1. Create `backend/svc/monolith/cmd/server/main.go` with a top-level chi router.
2. Refactor `backend/svc/auth` and `backend/svc/user` to expose router construction functions (so both their existing mains and the monolith can reuse the same routing definitions).
3. Mount routers into the monolith under:
   - `/v1/auth` (auth routes)
   - `/v1/students` (user routes)
   - `/health` (health check)
4. Ensure `POST /v1/auth/refresh` is mounted (handler exists today but is not wired in the auth router), since the gateway routes refresh traffic.
5. Ensure shared middleware behavior remains consistent:
   - request ID propagation (`X-Request-ID`)
   - standard JSON error envelope
   - JWT auth requirements on protected routes

### Phase B — Mail rewrite (Elixir → Go) while preserving contract

5. Implement a Go mail component that supports the current internal mail contract:
   - `POST /internal/email/send` (protected by internal token)
   - supported `type` values and required `params` validation
6. Replace auth’s dependency on the external mail service with an in-process interface in monolith mode (keep an HTTP-based adapter for split-service mode during transition).
7. Remove the Elixir mail service from the local stack only after monolith email delivery verification is green.

### Phase C — Cutover + rollback safety

8. Update `infra/docker-compose.yml` and `infra/nginx/nginx.conf` to route API traffic to the monolith.
   - Ensure gateway routing covers all auth endpoints that exist in the service contract (for example `/v1/auth/account`).
9. Keep the previous multi-service deployment option available until production cutover is validated.

## Test Plan

### Contract tests (HTTP-level)

- Validate that the monolith preserves:
  - route prefixes and HTTP methods used by current clients
  - status codes on success/error paths
  - standard error envelope shape
  - authentication requirements (JWT-protected vs public routes)

### Integration tests (Go)

- Auth flow coverage (monolith): register → email verify (OTP) → login → access protected user endpoints.
- User coverage (monolith): `GET /v1/students/me`, `PATCH /v1/students/me`, onboarding update, avatar URL endpoints.
- Internal endpoints: `/internal/students` and `/internal/email/send` require the internal token.

### Email verification

- For each supported email type (`EMAIL_VERIFY`, `OTP_LOGIN`, `EMAIL_CHANGE`, `PASSWORD_RESET`, `WELCOME`):
  - request is accepted (202) when params are valid
  - invalid params return 422 with a meaningful error
  - delivery failures are observable to operators (logs/metrics), and user-facing flows remain non-blocking where designed (fire-and-forget)

## Complexity Tracking

No constitution violations requiring justification.
