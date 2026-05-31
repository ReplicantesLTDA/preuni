# Implementation Plan: Backend Folder Reorganization for Monolith Architecture

**Branch**: `010-backend-monolith-cleanup` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/010-backend-monolith-cleanup/spec.md`

## Summary

Collapse `backend/svc/{auth,user,mail,content,learning,simulation,dissertation,notification,monolith}/` into a single `backend/app/` Go module organized by `internal/<domain>/`. Delete the standalone auth + user mains and the Elixir mail tree. Five stub services are folded as empty domain packages with README markers. Constitution + `CLAUDE.md` updated to describe the backend as a single Go binary. No HTTP contracts change.

Primary technical approach: mechanical `git mv` + import-path rewrite across all internal consumers in one coordinated PR, validated by the existing contract + integration test suite from feature 009.

## Technical Context

**Language/Version**: Go 1.24 (per `backend/go.work`)
**Primary Dependencies**: go-chi/chi v5 (router), pgx/v5 (Postgres), golang-jwt/jwt v5 (JWT), zap (logging), testify (tests), `net/smtp` (mail). No new dependencies introduced.
**Storage**: PostgreSQL 16 (`auth.*`, `users.*` schemas — unchanged). Redis 7 (existing). S3 (avatars).
**Testing**: `go test` + `testify`; contract tests at `<app>/tests/contract/`, integration at `<app>/tests/integration/`. Test setup unchanged apart from import paths.
**Target Platform**: Linux server (containerized via existing distroless Dockerfile) + local dev via `docker compose`.
**Project Type**: Backend source-tree reorganization (no new runtime behavior).
**Performance Goals**: No regression vs feature-009 baseline (`/v1/auth/login` p95 = 3.44 ms recorded). Constitution gate: p95 < 500 ms.
**Constraints**:
- HTTP surface MUST not change (routes, status codes, envelope shape).
- Git blame MUST survive (`git mv`, no delete+add).
- Single PR — atomic move + rewrite.
**Scale/Scope**: ~30 Go files moved; ~50 import paths rewritten; 1 module disappears (`svc/monolith` → `app`); 7 modules deleted (auth, user, mail, content, learning, simulation, dissertation, notification); 1 Elixir tree deleted.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Readability / single responsibility | PASS | Reorg increases readability by collapsing duplicated trees into one canonical layout. |
| No dead code | PASS — driver | Deleting Elixir mail + stub binaries DIRECTLY satisfies "no dead code" gate. |
| Unit tests for business logic | PASS | No business-logic code changes; existing tests move with their packages. |
| Integration tests for API boundaries | PASS | feature-009 contract + integration tests act as the regression suite for this reorg. |
| Performance requirements | PASS | No code-path changes; binary is bit-for-bit equivalent apart from import paths and package names. |
| Coverage floor ≥ 80% | PASS | No test removal; only path renames. |

Post-design re-check: still PASS — no constitution violations.

## Project Structure

### Documentation (this feature)

```text
specs/010-backend-monolith-cleanup/
├── spec.md
├── plan.md
├── research.md            # Phase 0 — git-mv strategy, import rewrite approach
├── quickstart.md          # Phase 1 — how to verify the reorg
└── tasks.md               # Phase 2 (/speckit.tasks)
```

No `data-model.md` (no entities). No `contracts/` (HTTP contracts unchanged — see feature 009).

### Source Code (repository root)

**Before (current):**

```text
backend/
├── go.work                          # lists pkg + 9 svc modules
├── pkg/                             # shared infra
└── svc/
    ├── auth/                        # legacy, retained for rollback
    ├── user/                        # legacy, retained for rollback
    ├── mail/                        # deprecated Elixir tree
    ├── monolith/                    # the actually-deployed binary
    ├── content/                     # stub main
    ├── learning/                    # stub main
    ├── simulation/                  # stub main
    ├── dissertation/                # stub main
    └── notification/                # stub main
```

**After (target):**

```text
backend/
├── go.work                          # lists pkg + app only
├── pkg/                             # shared infra (unchanged)
└── app/
    ├── go.mod                       # renamed from svc/monolith/go.mod
    ├── Dockerfile                   # updated COPY paths
    ├── README.md
    ├── .env.example
    ├── cmd/server/
    ├── internal/
    │   ├── auth/                    # from svc/auth/{adapters,domain,handler,ports,repository,router}
    │   ├── user/                    # from svc/user/{domain,handler,repository,router}
    │   ├── mail/                    # from svc/monolith/internal/mail
    │   ├── content/                 # NEW: empty + README ("scaffolding only — no endpoints yet")
    │   ├── learning/                # NEW: empty + README
    │   ├── simulation/              # NEW: empty + README
    │   ├── dissertation/            # NEW: empty + README
    │   ├── notification/            # NEW: empty + README
    │   ├── adapters/                # from svc/monolith/internal/adapters (in-process student/email)
    │   ├── config/                  # from svc/monolith/internal/config
    │   └── router/                  # from svc/monolith/internal/router
    └── tests/
        ├── contract/
        └── integration/
```

**Module path**: `github.com/preuni/svc/monolith` → `github.com/preuni/app`. All `github.com/preuni/svc/auth/...` and `github.com/preuni/svc/user/...` imports become `github.com/preuni/app/internal/{auth,user}/...`.

**Structure Decision**: One Go module under `backend/app/`. Domains live under `app/internal/<domain>/`. Sub-trees of each domain (handler, repository, domain, router, ports, adapters) preserve their original names, just relocated.

## Implementation Reference

### Phase A — Pre-reorg safety

1. Tag pre-reorg HEAD as `pre-monolith-reorg` (rollback path per FR-003).
2. Stop docker stack; commit unrelated WIP if any.

### Phase B — Tree moves

3. `git mv backend/svc/monolith backend/app`
4. `git mv backend/svc/auth/{adapters,domain,handler,ports,repository,router} backend/app/internal/auth/`
5. `git mv backend/svc/user/{domain,handler,repository,router} backend/app/internal/user/`
6. `git rm -r backend/svc/auth backend/svc/user backend/svc/mail backend/svc/content backend/svc/learning backend/svc/simulation backend/svc/dissertation backend/svc/notification`
7. `rmdir backend/svc` (should be empty).
8. Create empty `backend/app/internal/{content,learning,simulation,dissertation,notification}/README.md` with one-line marker.

### Phase C — Module + import rewrite

9. Edit `backend/app/go.mod`: rename module `github.com/preuni/svc/monolith` → `github.com/preuni/app`. Drop `replace github.com/preuni/svc/auth => ../auth` + `replace github.com/preuni/svc/user => ../user`.
10. Edit `backend/go.work`: drop all `./svc/*` entries except keep `./pkg` + new `./app`.
11. Bulk replace import paths:
    - `github.com/preuni/svc/monolith/` → `github.com/preuni/app/`
    - `github.com/preuni/svc/auth` → `github.com/preuni/app/internal/auth`
    - `github.com/preuni/svc/user` → `github.com/preuni/app/internal/user`
12. Move internal-test helper packages (auth `testhelper`, user `testhelper`) to their new homes; update their package declarations only if path changes.
13. `go mod tidy` inside `backend/app/`.

### Phase D — Infra + tooling

14. Update `backend/app/Dockerfile`: COPY paths drop `svc/` segment; only `pkg/` + `app/` needed.
15. Update `infra/docker-compose.yml`: `monolith` service `dockerfile: app/Dockerfile`, context `../backend`. Remove any remaining legacy service blocks if present.
16. Update `infra/nginx/nginx.conf`: drop dead `upstream auth_svc { ... }`, `upstream user_svc { ... }`, and all other upstreams whose services no longer exist in compose.
17. Update root `Makefile`: `cd backend/app && go run ./cmd/server`, `cd backend/app && go test ./...`. Remove `mail`, `auth`, `user` references from `run-backend`/`stop-backend`.
18. Update `backend/.golangci.yml` if its goimports `local-prefixes` needs adjustment (`github.com/preuni` covers both old and new — no change needed).

### Phase E — Documentation

19. Rewrite "Architecture / Backend" section of `.specify/memory/constitution.md` to describe the backend as a single Go monolith composed of domain packages.
20. Update `CLAUDE.md`: Project Structure section replaced with the new tree; Active Technologies trimmed to current truth.
21. Add a `backend/README.md` (or update existing) with the two-line layout summary.
22. Update `specs/009-backend-monolith-refactor/quickstart.md` if any paths it references no longer exist (mark with a "post-feature-010" note rather than rewriting history).

### Phase F — Validation

23. `cd backend/app && go build ./...` → zero errors.
24. `cd backend/app && go test ./...` → all green (unit + contract + integration, given Postgres reachable).
25. `cd /Users/dwbessa/projects/preuni/infra && docker compose build monolith && docker compose up -d monolith gateway`.
26. Live curl smoke: `/health`, `/v1/auth/register` → 201, `/v1/students/me` (Bearer) → 200, `/internal/email/send` with internal token → 202. Mirror feature-009 live validation.
27. Commit as a single PR titled `refactor: collapse backend into app/ module`.

### Phase G — Cleanup

28. After production cutover validated, delete the `pre-monolith-reorg` tag rollback announcement (optional).

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Import-path rewrite misses a reference → broken build | `go build ./...` in CI catches it deterministically before merge. |
| `git mv` loses blame on rename detection | Use `git mv` (not delete+add). Verify with `git log --follow` on a moved file pre-PR. |
| In-flight feature branches conflict heavily | Announce merge window in advance; rebase queue (mail thread + team chat). |
| External scripts hard-coding `backend/svc/...` | Greppable change. Phase D enumerates known surfaces; `grep -r "backend/svc/" .` after Phase D MUST return zero hits. |
| Docker build path drift | Verified by Phase F step 25 (real container build). |

## Test Plan

- Existing contract tests (`backend/svc/monolith/tests/contract/error_envelope_test.go` → moves to `backend/app/tests/contract/`): all 7 must pass post-reorg without assertion changes.
- Existing integration test (`auth_flow_test.go`): must pass against the same local Postgres.
- Live curl smoke per Phase F step 26.
- No new tests required by this feature (reorganization, not new behavior).

## Complexity Tracking

No constitution violations. Single mechanical PR; no new abstractions; no new dependencies.
