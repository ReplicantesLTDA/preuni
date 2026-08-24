# Implementation Plan: Constitution Alignment Refactor

**Branch**: `014-constitution-alignment-refactor` | **Date**: 2026-08-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/014-constitution-alignment-refactor/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Pivot preuni's backend/mobile from an ENEM-prep platform to an essay-challenge
gamification app, and wire the existing standalone `corretor-redacao` Python
service in as preuni's AI correction engine — without duplicating identity,
quota, or gamification data across the two systems. Concretely: (1) add four
new Go domains (`essay`, `streak`, `social`, `ranking`/`gamification`) to the
monolith as the system of record for submissions, streaks, friends, weekly
ranking, and medals; (2) strip `corretor-redacao`'s own auth/users/quota
layers and repurpose its correction pipeline as an internal-only grading
engine that the monolith drives via a durable outbox table + LISTEN/NOTIFY,
reusing its existing 5-competency schema and typed-error taxonomy; (3) stand
up a monorepo CI/CD pipeline (pre-commit + GitHub Actions) that runs Go,
Python, and TypeScript test suites, enforces a 90% coverage floor per
codebase, and blocks merge to `main` without green checks + human approval.

## Technical Context

**Language/Version**: Go 1.24 (monolith), Python 3.12 (correction service), TypeScript 5.9 / React Native 0.81 + Expo SDK 54 (mobile)
**Primary Dependencies**: go-chi/chi v5, pgx/v5, golang-jwt/jwt v5, zap (Go); FastAPI, SQLAlchemy 2.x async, asyncpg, Alembic, structlog (Python); Expo Router 6, TanStack Query v5, Zustand v5 (mobile)
**Testing**: `go test` (Go, ≥90% line coverage via `go test -cover`); pytest + pytest-cov (Python, ≥90%); Jest 29 + jest-expo + `@testing-library/react-native` (mobile, ≥90% via `--coverage`)
**Target Platform**: Linux server (Go monolith + Python correction service, containerized), iOS/Android/Web via Expo (mobile)
**Project Type**: Web application (mobile app + backend monolith + internal Python service) — existing repo structure, no new top-level project
**Performance Goals**: Existing constitution budgets apply (p95 < 500ms for synchronous endpoints; LCP ≤2.5s, INP ≤200ms on mobile); essay grading is async and excluded from the synchronous budget
**Constraints**: Correction service must not own identity/quota data (Constitution: Architecture); streak/ranking math computed server-side only and deterministically (Principle III); no submission may block on LLM latency (Principle V); CI must gate on 90% coverage per codebase (Principle II) and human approval + green CI before any merge to `main` (Principle VI)
**Scale/Scope**: 4 new Go domains, 1 refactored Python service (strip 3 subsystems: auth, users, quota), 1 new outbox/bridge table, mobile screens for essay submission/streak/friends/leaderboard/medals, 1 new CI pipeline covering 3 codebases

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Status |
|---|---|---|
| I. Code Quality | New Go domains follow existing `internal/<domain>/` package convention; no god objects | PASS (structural, enforced in review) |
| II. Testing Standards | Plan mandates TDD for all 4 new domains + correction-service refactor; coverage floor raised to 90% and wired into CI (this feature's own FR-009/FR-010) | PASS — this feature *is* the mechanism that enforces this |
| III. Gamification & UX Consistency | Streak/medal/ranking calculations specified as server-side-only, deterministic (data-model.md); mobile UI work for these screens will reuse existing design tokens — flagged for the tasks phase, not a plan-time violation | PASS |
| IV. Performance Requirements | Correction remains async (submit-then-poll); no new synchronous endpoint introduces LLM latency into the p95 budget | PASS |
| V. AI Correction Integrity | Correction service keeps its structured 5-competency output, async pipeline, prompt-version discipline, provider abstraction; quota enforcement moves *fully* to the Go monolith (was partially in Python) — this is a **fix**, not a violation | PASS |
| VI. Engineering Workflow | This feature adds the pre-commit + CI/CD gate that principle VI requires; no engineering-workflow violation introduced | PASS — mechanism-providing feature |

**Initial gate result**: PASS. No complexity exceptions required — see Complexity Tracking (empty).

## Project Structure

### Documentation (this feature)

```text
specs/014-constitution-alignment-refactor/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md         # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
backend/
├── app/
│   ├── cmd/server/
│   ├── internal/
│   │   ├── auth/                 # existing — untouched
│   │   ├── user/                 # existing — gains streak/tier fields
│   │   ├── mail/                 # existing — untouched
│   │   ├── essay/                # NEW — submissions, quota gate, correction-jobs outbox writer/reader
│   │   ├── streak/               # NEW — daily streak state machine (UTC boundary)
│   │   ├── social/               # NEW — friend requests/connections, visibility gating
│   │   ├── gamification/         # NEW — weekly ranking, league tiers, medals
│   │   ├── content|learning|simulation|dissertation|notification/  # existing scaffolding — untouched this feature
│   │   └── adapters/router/config/  # existing — router gains new domain routes
│   └── pkg/                      # existing shared infra — untouched
└── tests/integration/            # existing — gains essay/streak/social/gamification suites

corretor-redacao/                 # renamed internal service, still Python/FastAPI
├── src/
│   ├── api/                      # trimmed: correction endpoints only, internal-network auth (shared secret / mTLS), auth+users+quota endpoints REMOVED
│   ├── workers/                  # unchanged shape — LISTEN/NOTIFY + poll consumer, now reads from the shared `correction_jobs` outbox table
│   ├── corrector/                # unchanged — pure grading logic, prompts, LLM abstraction (kept isolated per import-linter rule)
│   ├── db/                       # migrations trimmed: users/refresh_tokens/consent_records/email_verification_tokens DROPPED; corrections table re-keyed to opaque user_id (no local FK)
│   └── observability/            # unchanged
└── tests/                        # unit/integration/golden — unchanged shape, updated fixtures for the trimmed schema

mobile/
├── app/(tabs)/                   # gains: essay submission flow, streak view, friends, leaderboard, medals screens
└── src/features/
    ├── essay/                    # NEW — api.ts, hooks.ts, validation.ts
    ├── streak/                   # NEW
    ├── social/                   # NEW
    └── gamification/             # NEW

.github/workflows/
├── mobile-ci.yml                 # existing — gains coverage-floor enforcement (90%)
├── backend-ci.yml                # NEW — go test + coverage gate
└── correction-service-ci.yml     # NEW — pytest + coverage gate

.pre-commit-config.yaml           # NEW at repo root — fast unit tests + lint + type-check for all 3 codebases
```

**Structure Decision**: Existing repo layout (Go monolith at `backend/app/`, mobile at
`mobile/`) is preserved. `corretor-redacao/` moves in as a sibling top-level
directory (already present at repo root from the WIP import) and is treated
as an internal service, not a public one — it gains no new public surface
area, it loses its end-user-facing auth/quota surface. A repo-root
`.pre-commit-config.yaml` and two new GitHub Actions workflows are added
because this feature's own FR-009…FR-013 require CI/CD spanning all three
codebases, which did not exist as a unified pipeline before.

## Complexity Tracking

> No Constitution Check violations — table intentionally empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
