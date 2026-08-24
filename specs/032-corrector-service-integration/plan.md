# Implementation Plan: Corrector Service Integration

**Branch**: `032-corrector-service-integration` | **Date**: 2026-08-24 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/032-corrector-service-integration/spec.md`

## Summary

ai-corrector's *application* code was already trimmed to internal-only in
`014-constitution-alignment-refactor` (no public endpoints beyond
health/metrics, DB-mediated bridge via `correction.correction_jobs`), and
the Go side of that bridge is fully implemented. What was never carried
through is the *deployment* footprint: ai-corrector still runs its own
standalone Postgres container/network and its own TLS reverse proxy, and
is entirely absent from `infra/docker-compose.yml` / `make dev`. This
plan folds ai-corrector into the shared `preuni` compose stack — one
Postgres instance, one bring-up command, no public ingress of its own —
so the already-built bridge actually runs end to end.

## Technical Context

**Language/Version**: Python 3.12 (correction service, unchanged) / Go 1.25 (monolith, unchanged) — this feature touches infra only, not application code
**Primary Dependencies**: Docker Compose (orchestration), Alembic (correction service's existing migration tool), the Go monolith's own `infra/migrations/*.sql` + `make migrate` runner
**Storage**: PostgreSQL 16 — single shared instance (`infra/docker-compose.yml`'s `postgres` service, database `preuni`), correction service moves from its own `corretor_db` on its own container into the `correction` schema of that same database
**Testing**: No new application logic; validated via the existing `make dev` bring-up + a manual/scripted end-to-end essay-submission smoke test (quickstart.md), plus whatever CI already covers each side independently (`backend-ci.yml`, `correction-service-ci.yml`)
**Target Platform**: Docker Compose (local dev today; the same compose shape is the basis for whatever staging/prod deploy exists — out of scope to prescribe further per spec Assumptions)
**Project Type**: Infrastructure/deployment reconfiguration of an existing web-service pair (no new service, no new UI)
**Performance Goals**: N/A — no change to request-path or grading performance; this is a topology change
**Constraints**: Migration ordering must not require manual sequencing (FR-006); correction service must remain reachable only from the internal network (FR-004)
**Scale/Scope**: Two existing services (Go monolith, correction service) + their two independent migration mechanisms, unified onto one Postgres instance

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Architecture section**: describes the bridge as "the Go monolith calls
  this service over internal HTTP" — this is **stale**. The actual,
  already-implemented, already-documented contract
  (`specs/014-constitution-alignment-refactor/contracts/internal-bridge.md`,
  and `backend/app/internal/essay/repository/reconciler.go`) is a
  DB-mediated bridge table, explicitly *not* HTTP. This plan does not
  introduce that drift — it predates this feature — but this plan also
  does not resolve it, since amending the constitution requires the
  "written proposal, team discussion" process (Governance section), not
  a unilateral edit inside an unrelated feature's plan. **Flagged for the
  user**; recommend a follow-up constitution amendment PR pointing at the
  existing internal-bridge contract as the source of truth. Not a gate
  failure for this feature since this feature only changes *deployment*
  of the already-correct DB-bridge design.
- **Principle II (Testing Standards)**: no new business logic is added by
  this feature, so no new unit/integration tests are required by that
  principle's letter; however the "integration boundary" language
  ("Go monolith ↔ correction service boundary") already mandates
  integration coverage of that boundary — which today exists
  (`backend/app/tests/integration/*fault_injection*`, `reconciler_*`).
  This plan does not need to add new tests to satisfy Principle II, but
  Phase 1's quickstart doubles as the manual verification the spec's
  User Story 1 acceptance scenarios call for.
- **Principle VI (migration reversibility)**: any new SQL this plan
  introduces (a role/grant script, if the grant script is actually wired
  up — see research.md) must ship a working downgrade. Alembic's
  existing `alembic downgrade base && upgrade head` CI check already
  covers the correction service's own migration chain and does not
  change. No new Go-side (`infra/migrations/*.sql`) migration is expected
  — this plan reuses the existing `essay` schema and `grant-correction-jobs.sql`
  as-is; if that script needs edits, it inherits the sibling
  `.down.sql` requirement.
- **Principle VII (Security & Secrets)**: consolidating onto one Postgres
  instance means the correction service's DB credentials move out of its
  own `.env`/compose file into the shared one — still gitignored, still
  never committed. No new secret category is introduced.
- **No enterprise-specific security theater** (Principle VII closing
  bullet): the existing `grant-correction-jobs.sql` assumes a
  `preuni_monolith` role that is never actually created anywhere in the
  repo — the monolith currently connects as the same `POSTGRES_USER`
  superuser-equivalent role everything else uses. Per the constitution's
  explicit rejection of RBAC-style access control for this project, this
  plan does **not** introduce a new role just to satisfy that script;
  see research.md for the resulting decision on whether to run it, adapt
  it, or retire it.

**Gate result**: PASS, with one flagged pre-existing doc drift (Architecture
section's stale HTTP description) surfaced to the user rather than
silently resolved or silently ignored.

## Project Structure

### Documentation (this feature)

```text
specs/032-corrector-service-integration/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md         # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks — not created here)
```

No `contracts/` directory: this feature does not add or change an
external interface. The bridge table's contract already exists at
`specs/014-constitution-alignment-refactor/contracts/internal-bridge.md`
and is unchanged by this plan.

### Source Code (repository root)

```text
infra/
├── docker-compose.yml          # gains: corrector-api, corrector-worker services;
│                                #        postgres service gains no schema change
│                                #        (correction schema created by Alembic)
├── .env.example                 # gains: correction-service env vars
│                                #        (LLM provider, prompt version, etc.)
├── migrations/
│   └── grant-correction-jobs.sql  # kept, adapted, or retired — see research.md
└── nginx/nginx.conf             # unchanged — gateway already doesn't route to
                                  # the correction service and shouldn't

ai-corrector/
├── deploy/
│   ├── docker-compose.yml       # retired (superseded by infra/docker-compose.yml)
│   ├── Caddyfile                # retired (no public ingress — FR-004)
│   └── postgres-init.sql        # retired (no more standalone Postgres container)
├── Dockerfile                   # kept as-is — still how both new compose
│                                 # services build their image
├── .env.example                  # updated: DATABASE_URL example now points at
│                                 # the shared preuni Postgres/`correction` schema
└── README.md                     # updated: quick-start section references the
                                  # shared `make dev`, not `make up` inside ai-corrector

Makefile                          # `dev` target gains the correction service;
                                   # `migrate` target's ordering documented/enforced
```

**Structure Decision**: No new top-level directories. This is a
reconfiguration of existing `infra/` and `ai-corrector/deploy/` files —
add two services to the one compose file the rest of the stack already
uses, delete the standalone compose/proxy files that duplicated it.

## Complexity Tracking

*No constitution violations requiring justification — table omitted.*
