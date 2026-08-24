---

description: "Task list for 032-corrector-service-integration"
---

# Tasks: Corrector Service Integration

**Input**: Design documents from `/specs/032-corrector-service-integration/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Not explicitly requested for this infra-only feature (no new
business logic — see plan.md Constitution Check). Verification is the
quickstart.md smoke test plus explicit checks that the retired standalone
artifacts are actually gone, called out as their own tasks below.

**Organization**: Tasks are grouped by user story per spec.md (US1 = P1
grading works end-to-end, US2 = P2 single bring-up command, US3 = P3 no
public ingress).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to

## Path Conventions

Existing repo layout — see plan.md's Project Structure. No new
directories.

---

## Phase 1: Setup

**Purpose**: Confirm the environment this feature edits before touching it.

- [X] T001 Confirm the local `infra/.env` (gitignored, not `.env.example`)
      has `POSTGRES_USER`/`POSTGRES_PASSWORD`/`POSTGRES_DB` set, since
      the correction service's `DATABASE_URL` will be built from the
      same values (`infra/.env`, no code change — verification only)
- [X] T002 [P] Confirm `ai-corrector/src/db/migrations/env.py`'s
      `_resolve_database_url()` reads `DATABASE_URL` from the process
      environment with no hardcoded fallback host (read-only check,
      `ai-corrector/src/db/engine.py`)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Wire the correction service into the shared compose stack
against the shared Postgres instance. Every user story depends on this.

**⚠️ CRITICAL**: No user story task can be verified until this phase is
complete — it's what makes the correction service reachable at all.

- [X] T003 Add `corrector-api` and `corrector-worker` services to
      `infra/docker-compose.yml`: build from `../ai-corrector`
      (`dockerfile: Dockerfile`), `DATABASE_URL` built from the existing
      `${POSTGRES_USER}`/`${POSTGRES_PASSWORD}`/`${POSTGRES_DB}` vars
      pointed at the `postgres` service (asyncpg URL form:
      `postgresql+asyncpg://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}`),
      `depends_on: postgres: condition: service_healthy`, no host port
      published for either (internal network only — FR-004); `corrector-worker`
      overrides `command: python -m src.workers.correction_worker`
      matching `ai-corrector/deploy/docker-compose.yml`'s prior worker
      command
- [X] T004 Reuse `corrector-api`'s existing `HEALTHCHECK`
      (`ai-corrector/Dockerfile`) for compose's own
      `depends_on: condition: service_healthy` wherever another service
      needs to wait on it (research.md R5); no new healthcheck logic
      — the Dockerfile's `HEALTHCHECK` is inherited automatically by
      compose; nothing currently needs `depends_on` on `corrector-api`
      itself, so no explicit compose `healthcheck:` block was added
- [X] T005 Add correction-service env vars (`LLM_PROVIDER`,
      `OLLAMA_BASE_URL`, `PROMPT_VERSION`, `PROMPT_TEMPERATURE`, and any
      other vars `corrector-api`/`corrector-worker` require per
      `ai-corrector/.env.example`, minus everything already removed in
      `014` — no JWT/SMTP/CORS/rate-limit vars) to `infra/.env.example`,
      grouped under a new `# --- Correction service ---` section
- [X] T006 Add an Alembic migration step for the correction service to
      the `migrate` target in `Makefile` (currently only runs the Go
      monolith's raw `infra/migrations/*.sql` files), running
      `alembic upgrade head` against the shared `DATABASE_URL` — e.g. via
      `docker compose -f infra/docker-compose.yml run --rm corrector-api alembic upgrade head`,
      consistent with how `migrate` already targets the `postgres`
      service via `docker compose exec` (research.md R2: no ordering
      dependency on the Go-side migrations, either order is safe)
      — also fixed a pre-existing bug found while editing this target:
      `grant-correction-jobs.sql` was being picked up unconditionally by
      the glob and would fail (`preuni_monolith` role doesn't exist);
      excluded it from the automatic loop per research.md R3
- [X] T007 Update the `dev` target in `Makefile` to bring up
      `corrector-api` and `corrector-worker` alongside `monolith` and
      `gateway`, and to run the new Alembic migration step from T006 in
      the same place `$(MAKE) migrate` already runs for the Go monolith

**Checkpoint**: `make dev` starts every container the stack needs,
against one Postgres instance, with both migration chains applied.
Individual user stories can now be verified.

---

## Phase 3: User Story 1 - A submitted essay gets graded, every time (Priority: P1) 🎯 MVP

**Goal**: The existing enqueue → grade → reconcile pipeline actually
completes locally, because the correction service is now running and
pointed at the schema the monolith already writes to.

**Independent Test**: Run `make dev`, submit a test essay through the
monolith, observe it reach `graded` (or a typed `failed`) without any
manual step — per quickstart.md's end-to-end smoke test.

### Implementation for User Story 1

- [X] T008 [US1] Run `make dev` from a clean state and confirm via
      `docker compose -f infra/docker-compose.yml ps` that `postgres`,
      `redis`, `gateway`, `monolith`, `corrector-api`, and
      `corrector-worker` are all `Up`/`healthy`
      — deviation: `backend/app/Dockerfile` (`FROM golang:1.24-alpine`)
      currently fails to build (`go: go.mod requires go >= 1.25.0
      (running go 1.24.13; GOTOOLCHAIN=local)` — the official
      `golang:1.24-alpine` image ships `GOTOOLCHAIN=local`, so it refuses
      to auto-fetch the `go1.26.6` toolchain `go.mod`/`go.work` pin. This
      is a pre-existing bug unrelated to this feature (out of scope —
      not fixed here; flagged for the user). Verified `postgres`,
      `redis`, `corrector-api` (healthy), `corrector-worker` all come up
      via `docker compose -f infra/docker-compose.yml up -d ...`
      instead, and ran the monolith natively
      (`cd backend/app && go run ./cmd/server`, go1.26.2 installed
      locally) against the same dockerized postgres/redis on
      `localhost` — this exercises the exact same code path `make dev`
      would once the Dockerfile issue is fixed separately. Also fixed:
      `corrector-worker` inherited `corrector-api`'s Dockerfile
      `HEALTHCHECK` (an HTTP curl to `/healthz`) despite having no HTTP
      surface, so it always showed `unhealthy`; added
      `healthcheck: disable: true` on the `corrector-worker` service in
      `infra/docker-compose.yml` per research.md R5's original intent
- [X] T009 [US1] Confirm both schemas exist in the one shared database:
      `docker compose -f infra/docker-compose.yml exec -T postgres psql -U preuni -d preuni -c "\dn"`
      lists both `essay` and `correction` — confirmed (also `auth`,
      `gamification`, `social`, `users`, all in the one `preuni`
      database)
- [X] T010 [US1] Execute the quickstart.md end-to-end grading smoke
      test (register/login, submit an essay, poll status) and confirm
      the submission reaches a terminal state within
      `backend/app/internal/essay/repository/reconciler.go`'s
      `GradingTimeout` without any manual container restart or migration
      — confirmed. Registered via `POST /v1/auth/register` (returns a
      usable access token immediately, no email-verification gate on
      this endpoint), submitted via `POST /v1/essays` (route is
      `/v1/essays`, not `/v1/students/essays` as quickstart.md's
      illustrative curl guessed — quickstart.md corrected), polled
      `GET /v1/essays/{id}`. Result: `pending` → **`failed`** in well
      under a second (LISTEN/NOTIFY fired immediately, no 5s poll wait),
      with `correction.corrections.error_code = 'provider_unavailable'`
      — expected and correct: no Ollama container exists in this
      compose stack (`CORRECTOR_OLLAMA_BASE_URL` points at an `ollama`
      hostname nothing serves), so the worker correctly typed-failed
      the LLM call per its existing failure-classifier rather than
      hanging or crashing. This validates the bridge mechanism
      end-to-end exactly as intended — the DB round trip
      (enqueue → NOTIFY → claim → grade-attempt → typed failure →
      reconcile) all worked; only the actual LLM provider is
      unconfigured, which is out of this feature's scope (deployment
      wiring, not LLM provisioning — see plan.md Technical Context)
- [X] T011 [US1] Stop `corrector-worker` mid-flight
      (`docker compose -f infra/docker-compose.yml stop corrector-worker`),
      submit an essay, confirm it sits `pending`, then restart the worker
      and confirm the job is claimed and completes on its own — validates
      spec Acceptance Scenario 3 (job survives a service restart, no
      resubmission needed) — confirmed. Used a second registered user
      (first had already exhausted the free-tier 1/day quota from T010).
      Job sat `pending` for 8s with the worker stopped, then `start`ed
      the worker (no new job inserted, so no LISTEN/NOTIFY fired) and it
      was still picked up via the worker's 5s poll backstop within ~10s,
      reaching `failed` (same `provider_unavailable`, no resubmission)

**Checkpoint**: User Story 1 fully verified — grading works end-to-end,
locally, with zero manual steps.

---

## Phase 4: User Story 2 - Developers bring up one stack, not two (Priority: P2)

**Goal**: A developer who has never touched ai-corrector gets a working
grading pipeline from the project's existing single entry point, with no
need to discover or run ai-corrector's own compose file.

**Independent Test**: A developer runs only the repo-root `make dev` and
never reads `ai-corrector/README.md`'s quick-start section or
`ai-corrector/deploy/`.

### Implementation for User Story 2

- [X] T012 [US2] Update `ai-corrector/README.md`'s "Quick start" section
      to point at the repo-root `make dev` instead of
      `make up` / `make db-migrate` run from inside `ai-corrector/`
      (those `ai-corrector/Makefile` targets are superseded, not deleted,
      for the case someone runs the Python side in true isolation for
      correction-pipeline-only development — see T017) — done. Note:
      `db-migrate` is kept working as-is (uses `DATABASE_URL` env, no
      dependency on the deleted `deploy/`); `up` could not be kept
      working as originally scoped here — it hard-depends on
      `deploy/docker-compose.yml`, which research.md R4 explicitly
      deletes. Repointed in T015/T017 instead of left broken.
- [X] T013 [US2] Update the root `README.md` (if it documents local setup)
      and/or `CLAUDE.md`'s Commands section to mention that `make dev`
      now brings up the correction service too, so it isn't undocumented
      — updated both `README.md` and `CLAUDE.md`'s `make dev` comment
- [X] T014 [US2] Confirm `make doctor` still passes on a machine that has
      never run anything under `ai-corrector/` directly (no new
      prerequisite this feature silently requires beyond what `make dev`
      itself now installs via Docker build) — confirmed:
      `✓ docker running, go 1.24+ available`

**Checkpoint**: User Stories 1 AND 2 both verified — one command, one
stack, grading works.

---

## Phase 5: User Story 3 - No public entry point that shouldn't exist (Priority: P3)

**Goal**: The correction service carries no reverse proxy, TLS, or
auth-scoped rate limiting of its own, and is reachable only from the
internal Docker network — matching what its own README already claims.

**Independent Test**: Inspect the correction service's deployment
configuration and running containers; confirm no public port, no TLS
setup, no rate-limit rule referencing removed endpoints.

### Implementation for User Story 3

- [X] T015 [US3] Delete `ai-corrector/deploy/docker-compose.yml` and
      `ai-corrector/deploy/postgres-init.sql` — fully superseded by
      T003's shared-compose services and the shared `postgres` service's
      own init (research.md R4) — done. Note: also fixed
      `ai-corrector/Makefile`'s `up`/`down`/`logs` targets, which
      hard-depended on the now-deleted `deploy/docker-compose.yml`;
      `up`/`down` now point the user at the repo-root `make dev` /
      `docker compose -f ../infra/docker-compose.yml down` instead of
      failing with a confusing "file not found"; `logs` now tails the
      shared compose's `corrector-api`/`corrector-worker` services.
      Left `ai-corrector/deploy/docker-compose.prod.yml` (and the nested
      `deploy/deploy/postgres-init.sql`, `deploy/backup/` pg_dump
      sidecar) untouched — same standalone-product pattern, but scoped
      to production topology, which spec.md's Assumptions explicitly
      puts out of scope for this feature. Flagged for the user; it also
      still references dead env vars (`JWT_SECRET_KEY`, `SMTP_*`) from
      before auth/mail were removed from this service in `014`.
- [X] T016 [US3] Delete `ai-corrector/deploy/Caddyfile` — no public
      ingress remains for this service, so no reverse proxy, TLS, or
      `/auth/*` rate limiting is needed (FR-003, FR-004; the routes it
      rate-limited no longer exist per `014`) — done
- [X] T017 [US3] Update `ai-corrector/README.md`'s Architecture/deploy
      references (and any mention of `deploy/Caddyfile` or
      `deploy/docker-compose.yml`) to describe the shared
      `infra/docker-compose.yml` as the only way this service is deployed
      — added a "Deployment" section; rewrote "Quick start" (T012)
- [X] T018 [US3] Confirm no host port is published for `corrector-api` or
      `corrector-worker` in `infra/docker-compose.yml` (per T003) and
      confirm `docker compose -f infra/docker-compose.yml ps` shows no
      `0.0.0.0:8000->8000` (or similar) mapping — only reachable
      container-to-container — confirmed, `ps` shows no port mapping for
      either service (only `postgres`/`redis` publish host ports)
- [X] T019 [US3] Confirm `infra/nginx/nginx.conf` has no `location` block
      routing to the correction service (it shouldn't, and this task is
      a check, not a change — the gateway is preuni's one public entry
      point) — confirmed, no reference to the correction service anywhere
      in `nginx.conf`

**Checkpoint**: All three user stories verified independently. Feature
complete.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Close out the pre-existing doc drift and unused-artifact
questions surfaced during planning, without expanding this feature's
scope.

- [X] T020 [P] Flag the constitution's Architecture section ("The Go
      monolith calls this service over internal HTTP") as stale to the
      user, per plan.md's Constitution Check — recommend (do not
      unilaterally make) a follow-up constitution-amendment PR pointing
      at `specs/014-constitution-alignment-refactor/contracts/internal-bridge.md`
      as the source of truth — NOT edited (by design, needs the
      documented team-discussion amendment process); surfaced in the
      final report to the user instead
- [X] T021 [P] Add a one-line comment to
      `infra/migrations/grant-correction-jobs.sql` clarifying it is an
      optional least-privilege hardening step for a deployment that has
      actually provisioned a separate `preuni_monolith` role, which this
      project's current dev/CI setup does not do (research.md R3) —
      leave the script itself unexecuted and unchanged otherwise — done
- [X] T022 Run the full quickstart.md checklist end-to-end one more time
      after all deletions (T015-T016) to confirm nothing regressed —
      confirmed: all four services still `Up`/`healthy`, no `corretor-*`
      network survives, both `essay` and `correction` schemas still
      present in the one shared `preuni` database

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
  (T003 in particular is the single task everything else verifies)
- **User Stories (Phase 3-5)**: All depend on Foundational completion;
  independently verifiable from there, though US3's deletions (T015-T016)
  are safest to do after US1 (T008-T011) has already proven the new
  compose wiring works, so there's no standalone stack to fall back to
  mid-verification
- **Polish (Phase 6)**: Depends on all three user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends only on Foundational
- **User Story 2 (P2)**: Depends only on Foundational; documentation-only,
  no code dependency on US1, but verifying it is more meaningful once
  US1 has confirmed grading actually works
- **User Story 3 (P3)**: Depends only on Foundational; recommended after
  US1 for the reason above, not a hard technical dependency

### Parallel Opportunities

- T001/T002 (Setup) in parallel
- T012/T013/T014 (US2, all documentation) largely parallel with each
  other and with US3's tasks once Foundational is done
- T020/T021 (Polish, different files) in parallel

---

## Parallel Example: Foundational Phase

```bash
Task: "Confirm infra/.env has POSTGRES_* set (T001)"
Task: "Confirm ai-corrector's env.py reads DATABASE_URL with no hardcoded fallback (T002)"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (T003-T007 — this is the actual
   integration work; everything after is verification/cleanup)
3. Complete Phase 3: User Story 1 (T008-T011)
4. **STOP and VALIDATE**: essay submitted locally reaches `graded`
5. This alone delivers the spec's stated MVP core

### Incremental Delivery

1. Setup + Foundational → correction service is part of the shared stack
2. User Story 1 → grading verified end-to-end (MVP)
3. User Story 2 → docs updated so the single-command bring-up is
   actually discoverable, not just technically true
4. User Story 3 → standalone deploy artifacts removed, public-surface
   claim verified
5. Polish → doc-drift flagged, unused-grant-script clarified

---

## Notes

- No new application code, no new tests requested — this is
  infrastructure reconfiguration; "implementation" tasks in each user
  story phase are largely verification against the Foundational phase's
  compose/Makefile changes (T003-T007), plus the deletions in US3
- T003 is the load-bearing task of this entire feature — get it right
  and US1/US2 verification should pass with no further code changes
- Commit after each phase, not each task — T003-T007 form one coherent
  compose-wiring change; T015-T017 form one coherent cleanup change
- Avoid re-introducing a second Postgres container/network under any
  name — that's the exact problem this feature exists to remove
