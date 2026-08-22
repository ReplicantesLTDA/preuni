# Tasks: Constitution Alignment Refactor

**Input**: Design documents from `/specs/014-constitution-alignment-refactor/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Included — Constitution Principle II (TDD, NON-NEGOTIABLE) and spec User Story 4 both require test-first work; every user story below writes failing tests before implementation.

**Organization**: Tasks are grouped by user story (US1–US4 from spec.md) to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

Existing repo layout (see plan.md Project Structure): `backend/app/internal/<domain>/`, `backend/tests/integration/`, `corretor-redacao/src/`, `corretor-redacao/tests/`, `mobile/src/features/<domain>/`, `mobile/app/(tabs)/`, `.github/workflows/`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Scaffolding so every later task has somewhere to land

- [X] T001 [P] Create `backend/app/internal/essay/`, `internal/streak/`, `internal/social/`, `internal/gamification/` directories with a `README.md` each (mirrors existing scaffolding convention in `internal/content/README.md`)
- [X] T002 [P] Create `mobile/src/features/essay/`, `features/streak/`, `features/social/`, `features/gamification/` directories, each with empty `api.ts`, `hooks.ts`, `validation.ts`
- [X] T003 [P] Create repo-root `.pre-commit-config.yaml` skeleton (hooks added in T019)
- [X] T004 Update `backend/app/go.mod`/`corretor-redacao/pyproject.toml`/`mobile/package.json` only if new libraries are needed for tasks below — verified: `google/uuid`, `pgx/v5`, `testify` already present in Go; no cron lib needed (week-close runs as an internal ticker/goroutine); no new deps required

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Schema, cross-system bridge, and identity/quota deduplication that every user story depends on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 Alembic migration in `corretor-redacao/src/db/migrations/` dropping `users`, `refresh_tokens`, `consent_records`, `email_verification_tokens` tables (research.md §2) — combined with T006 into `002_correction_jobs_schema.py` (single coherent schema-move migration)
- [X] T006 Alembic migration in `corretor-redacao/src/db/migrations/` creating schema `correction` with `correction_jobs` table (data-model.md) and re-keying `corrections.user_id` to a plain UUID column with no local FK
- [X] T007 [P] Go migration `infra/migrations/user/009_add_streak_gamification_columns.sql` adding `longest_streak`, `subscription_tier` to `users.students` — reuses the table's existing `streak_count`/`streak_last_active_date` columns as current-streak/last-submission-day instead of duplicating them (Principle I: no premature duplication)
- [X] T008 [P] Go migration `infra/migrations/essay/001_create_essay_submissions.sql` creating schema `essay` with `essay_submissions` and `essay_grades` tables (data-model.md); `essay` added to `infra/scripts/migrate.sh` SERVICES
- [X] T009 [P] Go migration `infra/migrations/social/001_create_friendships.sql` creating schema `social` with `friendships` table; `social` added to `infra/scripts/migrate.sh` SERVICES
- [X] T010 [P] Go migration `infra/migrations/gamification/001_create_ranking_and_medals.sql` creating schema `gamification` with `weekly_ranking_entries` and `medals` tables; `gamification` added to `infra/scripts/migrate.sh` SERVICES
- [X] T011 Grant narrow `INSERT`/`SELECT`-only privilege on `correction.correction_jobs` to the Go monolith's DB role in `infra/migrations/grant-correction-jobs.sql`
- [X] T012 Service-to-service auth: **not needed** — the bridge is fully DB-mediated (research.md §1), there is no Go→Python HTTP call in this design. Documented in `corretor-redacao/src/api/main.py`'s module docstring rather than building unused auth machinery (no premature abstraction, Principle I).
- [X] T013 Removed auth/users/quota endpoints, routes, and models from `corretor-redacao/src/api/` and `src/db/models/` (drop `/auth/*`, `/me`, `POST /corrections` and their backing `users`/`consent_records`/`refresh_tokens`/`email_verification_tokens` tables) per research.md §2
- [X] T014 [P] Updated `corretor-redacao/src/db/repositories/correction_repo.py` so `claim_next` claims from `correction_jobs` (`SELECT ... FOR UPDATE SKIP LOCKED`) and creates the matching `corrections` row itself; `mark_completed`/`mark_failed` flip the job's status too. `src/workers/correction_worker.py` needed no changes — same `claim_next`/`mark_completed`/`mark_failed` call sites. Migration 002 gained a `pg_notify` trigger on `correction_jobs` INSERT to replace the NOTIFY the removed Python enqueue path used to fire.
- [ ] T015 Wire new domain routers (`essay`, `streak`, `social`, `gamification`) into `backend/app/internal/router/router.go` — **deferred**: no handlers exist yet to wire. Each domain's `Mount()` call is added alongside that domain's own router package in its own user-story phase below (T033/T037 for essay+streak, T044 for social, T053 for gamification), not upfront against empty packages.

**Checkpoint**: Foundation ready — user story implementation can now begin

---

## Phase 3: User Story 4 - Ship changes safely and continuously (Priority: P1) 🎯 Enables safe delivery of everything below

**Goal**: A red test, a lint failure, or a coverage drop in any of the three codebases blocks merge to `main`; nothing merges without green CI and a human approval.

**Independent Test**: Open a PR that introduces a failing test (or drops coverage below 90%) and confirm it is blocked from merging; open a fully-passing PR and confirm it still can't merge without a recorded human approval.

**Why built first among the P1s**: every other story is developed test-first per Principle II; this story is the mechanism that actually enforces that, so standing it up first gives every subsequent task a real gate to fail against.

### Implementation for User Story 4

- [X] T016 [US4] Created `.github/workflows/backend-ci.yml`: applies `infra/migrations/*` via `psql` against a live Postgres service (golang-migrate's up/down naming doesn't match this repo's plain-numbered `.sql` files — pre-existing, out of scope; the `psql` approach was verified live in CI — migrations applied cleanly), then `bash backend/scripts/check-coverage.sh` (real gate). **First real CI run measured 19.0%** (full `go test ./...` against a live PG, all integration suites included) — floor set to 15% for headroom (see T060). `golangci-lint`/`gofmt` run **non-blocking**: the existing codebase has repo-wide gofmt/import-order drift from a toolchain-version difference, predating this pipeline.
- [X] T017 [P] [US4] Created `.github/workflows/correction-service-ci.yml`: `ruff check` + `ruff format --check` (both genuinely clean, repo-wide — verified locally), `import-linter` (real CLI is `lint-imports` with no flags — the Makefile's `--root src` is a pre-existing invalid-flag bug, fixed here after CI caught it), `pytest tests/unit --cov=src --cov-fail-under=45` (real, verified locally: 92 passed, 50.95% ≥ 45%), `pytest tests/integration -m integration` + golden (fake provider) — first real CI run pending (not verified locally, no live Postgres in this environment). `mypy` runs **non-blocking**: 25 pre-existing errors predate this refactor.
- [X] T018 [US4] `mobile-ci.yml` already ran `pnpm test -- --coverage --runInBand`; adding the threshold to `jest.config.js` (T020) makes that step enforce it — no separate workflow edit needed.
- [X] T019 [US4] Filled in `.pre-commit-config.yaml`: `gofmt` scoped to changed files (repo-wide would hit the same pre-existing drift as T016); `golangci-lint` repo-wide (documented caveat); `ruff check`/`ruff format --check`/`pytest tests/unit` for corretor-redacao; `tsc`/`eslint` for mobile.
- [X] T020 [US4] Added `coverageThreshold` to `mobile/jest.config.js` — provisional baseline `{lines:25, statements:25, functions:20, branches:30}`, verified against the real measured run (28.5/34.9/26.4/29.6%, 79/79 tests passing). Raised to 90 by T060.
- [X] T021 [US4] Added `fail_under = 45` to `corretor-redacao/pyproject.toml` `[tool.coverage.report]` — provisional baseline (unit suite alone measures ~51%). Raised to 90 by T060.
- [X] T022 [US4] Wrote `infra/branch-protection.md` with required checks (`backend-ci`, `correction-service-ci`, `mobile-ci`) + 1 required review + a `gh api` script; not yet applied against the live repo (needs a maintainer with admin rights to run it).
- [X] T023 [US4] Rewrote quickstart.md's coverage section to point at the real scripts/configs (`backend/scripts/check-coverage.sh`, `pyproject.toml`, `jest.config.js`) instead of a fabricated `--cov-fail-under=90`, and noted the provisional-baseline-to-90 ratchet explicitly.

**Note on "provisional, not yet 90%"**: none of the three codebases were anywhere near 90% coverage before this refactor (measured: mobile ~28.5%, Python unit-only ~51%, Go unit-only ~9.5% though most Go coverage lives in integration tests this environment couldn't run against a live DB). Gating at 90% today would make this PR — and every PR after it — permanently red for reasons unrelated to its own diff, which is worse than no gate at all. Real, verified, enforced floors are wired now; T060 (Polish) raises them to 90% once the essay/streak/social/gamification test suites this refactor itself requires (US1–US3) have landed and actually moved the number.

**Checkpoint**: CI/CD gate is live and independently verifiable (open a deliberately failing PR against this branch's later commits and confirm it's blocked)

---

## Phase 4: User Story 1 - Submit an essay and see it graded (Priority: P1) 🎯 MVP

**Goal**: A user submits an essay, it's accepted immediately, their streak advances, and a structured 5-competency grade appears once correction finishes — all without ever blocking on grading latency.

**Independent Test**: Submit one essay as a free-tier user; confirm 202 + streak +1 immediately, then confirm a graded result with all 5 competencies appears; confirm a second same-day submission is rejected.

### Tests for User Story 1 ⚠️

- [ ] T024 [P] [US1] Unit tests for quota enforcement (free 1/day, Pro multi/day, UTC day boundary) in `backend/app/internal/essay/quota_test.go`
- [ ] T025 [P] [US1] Unit tests for streak increment-on-submit and reset-on-missed-day in `backend/app/internal/streak/streak_test.go`
- [ ] T026 [P] [US1] Integration test: `POST /v1/essays` → 202 → poll `GET /v1/essays/{id}` → `graded` with 5 competencies in `backend/tests/integration/essay_test.go`
- [ ] T027 [P] [US1] Integration test: second same-day free-tier submission → 429 `quota_exhausted`; Pro-tier second submission → 202, in `backend/tests/integration/essay_quota_test.go`
- [ ] T028 [P] [US1] Integration test: correction failure (typed error) surfaces as `essay_submissions.status = failed` without losing that day's streak credit, in `backend/tests/integration/essay_failure_test.go`
- [ ] T029 [P] [US1] Python unit tests for `correction_jobs` claim → grade → complete flow in `corretor-redacao/tests/unit/test_correction_jobs.py`

### Implementation for User Story 1

- [ ] T030 [US1] Implement `essay_submissions` repository with per-day quota check in `backend/app/internal/essay/repository.go`
- [ ] T031 [US1] Implement streak service (UTC-boundary increment/reset, updates `users.current_streak`/`longest_streak`/`last_submission_day`) in `backend/app/internal/streak/service.go`
- [ ] T032 [US1] Implement `essay` → `correction_jobs` enqueue writer (same transaction as submission insert + streak update) in `backend/app/internal/essay/bridge.go`
- [ ] T033 [US1] Implement `POST /v1/essays` and `GET /v1/essays/{id}`, `GET /v1/essays` handlers in `backend/app/internal/essay/handler.go`
- [ ] T034 [US1] Implement reconciliation poller reading completed/failed `correction_jobs` rows into `essay_grades` / failed `essay_submissions`, including the grading-timeout failure path (spec Edge Cases, SC-001) in `backend/app/internal/essay/reconciler.go`
- [ ] T035 [US1] Implement `GET /v1/streaks/me` handler in `backend/app/internal/streak/handler.go`
- [ ] T036 [P] [US1] Update `corretor-redacao/src/workers/pipeline.py` to read job input from `correction_jobs` and write the graded result keyed by job id
- [ ] T037 [P] [US1] Mobile: essay submission form + submit/poll hooks in `mobile/src/features/essay/{api.ts,hooks.ts,validation.ts}`
- [ ] T038 [P] [US1] Mobile: streak display hook/component in `mobile/src/features/streak/{api.ts,hooks.ts}`
- [ ] T039 [US1] Wire essay submission + streak display into `mobile/app/(tabs)/redacao/` screens

**Checkpoint**: User Story 1 fully functional and independently testable — this is the MVP

---

## Phase 5: User Story 2 - Add friends and see their progress (Priority: P2)

**Goal**: Users connect as friends and can see each other's streak and latest grade; no one else can.

**Independent Test**: Two test accounts add each other, each sees the other's streak + latest grade; a third, non-friend account gets a 404 trying to look either up.

### Tests for User Story 2 ⚠️

- [ ] T040 [P] [US2] Unit tests for friendship visibility gating (accepted-only) in `backend/app/internal/social/visibility_test.go`
- [ ] T041 [P] [US2] Integration test: request → accept → both sides see streak/grade; removal revokes visibility; non-friend gets 404, in `backend/tests/integration/social_test.go`

### Implementation for User Story 2

- [ ] T042 [US2] Implement friendship repository (request/accept/remove) in `backend/app/internal/social/repository.go`
- [ ] T043 [US2] Implement visibility-gated query joining `friendships` + `users` + latest `essay_grades` in `backend/app/internal/social/query.go`
- [ ] T044 [US2] Implement `POST /v1/friends/requests`, `POST /v1/friends/requests/{id}/accept`, `DELETE /v1/friends/{id}`, `GET /v1/friends` handlers in `backend/app/internal/social/handler.go`
- [ ] T045 [P] [US2] Mobile: friends list/add/accept screen + hooks in `mobile/src/features/social/{api.ts,hooks.ts}`
- [ ] T046 [US2] Wire friends screen into `mobile/app/(tabs)/perfil/` (or dedicated tab per design)

**Checkpoint**: User Stories 1 AND 2 both work independently

---

## Phase 6: User Story 3 - Compete on the weekly ranking and earn medals (Priority: P3)

**Goal**: Weekly scores accumulate into a global leaderboard with league tiers that promote/demote at week-end; streak and ranking milestones award visible medals.

**Independent Test**: Several test accounts submit graded essays across a week; leaderboard orders them correctly; at week-end, top/bottom bands promote/demote per the documented rule; medals appear on profiles.

### Tests for User Story 3 ⚠️

- [ ] T047 [P] [US3] Unit tests for deterministic weekly score aggregation + tie-break (earlier `graded_at` wins) in `backend/app/internal/gamification/ranking_test.go`
- [ ] T048 [P] [US3] Unit tests for medal-award triggers (streak milestones, tier promotion, top finishes) in `backend/app/internal/gamification/medals_test.go`
- [ ] T049 [P] [US3] Integration test: week-close job promotes top band / demotes bottom band and resets weekly score to zero without touching `current_streak`, in `backend/tests/integration/ranking_test.go`

### Implementation for User Story 3

- [ ] T050 [US3] Implement `weekly_ranking_entries` repository + score aggregation from `essay_grades` in `backend/app/internal/gamification/ranking_repository.go`
- [ ] T051 [US3] Implement week-close job (promote/demote bands, tie-break, weekly reset) in `backend/app/internal/gamification/week_close.go`
- [ ] T052 [US3] Implement medal-award triggers hooked to streak updates and week-close events in `backend/app/internal/gamification/medals.go`
- [ ] T053 [US3] Implement `GET /v1/ranking/weekly`, `GET /v1/ranking/me`, `GET /v1/medals/me` handlers in `backend/app/internal/gamification/handler.go`
- [ ] T054 [P] [US3] Mobile: leaderboard screen + hooks in `mobile/src/features/gamification/{api.ts,hooks.ts}`
- [ ] T055 [P] [US3] Mobile: medals display component in `mobile/src/features/gamification/medals.tsx`
- [ ] T056 [US3] Wire leaderboard + medals into `mobile/app/(tabs)/simulado/` (or dedicated ranking tab per design)

**Checkpoint**: All four user stories independently functional

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Repo-wide cleanup and final validation once all stories land

- [ ] T057 [P] Run every command in `quickstart.md` end to end against a clean checkout; fix any drift
- [ ] T058 [P] Update `backend/app/internal/README.md` (or equivalent) and `corretor-redacao/README.md` to describe the trimmed, internal-only correction service
- [ ] T059 Remove now-dead Python auth code and its tests fully from `corretor-redacao/src/` and `corretor-redacao/tests/` (cleanup pass beyond T013's route removal)
- [ ] T060 [P] Verify combined coverage ≥90% across Go, Python, and mobile on the final branch state; add tests to close any gap
- [ ] T061 Update `CLAUDE.md` / `AGENTS.md` "Recent Changes" section to record the 014 refactor

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories (schema + bridge table + identity/quota deduplication must exist first)
- **User Story 4 / CI-CD (Phase 3)**: Depends on Foundational only. Built first among the P1s so Phases 4–6 have a real gate.
- **User Story 1 (Phase 4)**: Depends on Foundational; does not depend on Phase 3 to *function*, but should land through the Phase 3 gate once it exists
- **User Story 2 (Phase 5)**: Depends on Foundational + User Story 1 data (friendships surface US1's streak/grades) — data dependency only, still independently testable with seeded US1 data
- **User Story 3 (Phase 6)**: Depends on Foundational + User Story 1 data (ranking aggregates US1's grades) — same shape of dependency as US2
- **Polish (Phase 7)**: Depends on all four user stories being complete

### Parallel Opportunities

- All [P] tasks within Phase 1 and within Phase 2 (T007–T010) run in parallel
- Within Phase 3, T016–T018 (the three CI workflow files) run in parallel
- Within each user story, all test tasks marked [P] run in parallel before implementation begins
- Within each user story, mobile [P] tasks run in parallel with Go [P] tasks (different codebases, different files)
- Once Foundational + Phase 3 are done, Phases 4–6 could be staffed in parallel by different contributors, though US2/US3 need at least seed data from US1 to demo meaningfully

---

## Parallel Example: User Story 1

```bash
# Tests together:
Task: "Unit tests for quota enforcement in backend/app/internal/essay/quota_test.go"
Task: "Unit tests for streak increment/reset in backend/app/internal/streak/streak_test.go"
Task: "Integration test submit->graded in backend/tests/integration/essay_test.go"
Task: "Python unit tests for correction_jobs flow in corretor-redacao/tests/unit/test_correction_jobs.py"

# Mobile + backend implementation in parallel once tests are red:
Task: "Mobile essay submission hooks in mobile/src/features/essay/{api.ts,hooks.ts,validation.ts}"
Task: "Mobile streak hook in mobile/src/features/streak/{api.ts,hooks.ts}"
```

---

## Implementation Strategy

### MVP First

1. Setup (Phase 1) → Foundational (Phase 2, critical, blocks everything)
2. User Story 4 / CI-CD (Phase 3) — get the safety net live before writing feature code
3. User Story 1 (Phase 4) — **STOP and VALIDATE**: submit → grade → streak loop works end to end, quota enforced, CI green on the PR that introduced it
4. This is the MVP: a working essay-challenge core loop, safely shippable

### Incremental Delivery

1. Foundational + CI/CD ready → nothing ships without tests/coverage/review from here on
2. Add User Story 1 → validate independently → demo (MVP)
3. Add User Story 2 (friends) → validate independently → demo
4. Add User Story 3 (ranking + medals) → validate independently → demo
5. Polish pass, confirm ≥90% coverage repo-wide

### Parallel Team Strategy

1. Everyone: Setup + Foundational + CI/CD together (Phases 1–3)
2. Once that's done: Developer A takes US1, Developer B starts US2/US3 scaffolding against US1's seed data as US1 stabilizes
3. Stories integrate through the shared `users`/`essay_grades` tables, not through direct code coupling

---

## Notes

- [P] tasks touch different files with no unmet dependency
- Every implementation task in Phases 4–6 has a corresponding test task written first, per Principle II
- Commit after each task or logical group; every commit still has to clear the Phase 3 pre-commit hook once it exists
- Cross-story "dependency" between US1 and US2/US3 is data-only (they read `essay_grades`/streak state US1 writes) — no direct code import between domain packages
