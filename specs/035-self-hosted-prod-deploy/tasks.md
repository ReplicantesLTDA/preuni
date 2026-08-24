---

description: "Task list for 035-self-hosted-prod-deploy"
---

# Tasks: Self-Hosted Production Deployment

**Input**: Design documents from `/specs/035-self-hosted-prod-deploy/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/deploy-pipeline.md, quickstart.md

**Tests**: Not explicitly requested for this infra/ops feature (no new
application code — plan.md's Constitution Check). Verification is
quickstart.md's checklist against spec's Success Criteria.

**Organization**: Tasks are grouped by user story per spec.md (US1 = P1
external reachability, US2 = P2 CD pipeline, US3 = P1 essay grading in
production).

**⚠️ Manual/operator tasks**: Several tasks require physical/account
access this agent does not have — the user's home NAS, their Cloudflare
account, GitHub org/repo admin settings. Those are marked `[MANUAL]` with
step-by-step instructions instead of an implementation the agent can run.
Everything else (compose files, workflow YAML, ADR, docs) is a normal
file-editing task.

## Format: `[ID] [P?] [Story?] [MANUAL?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to
- **[MANUAL]**: Requires the user to act outside this repo/session

## Path Conventions

Existing repo layout — see plan.md's Project Structure. No new
application directories; this feature adds ops/deploy config only.

---

## Phase 1: Setup

**Purpose**: Governance and documentation groundwork the rest of the
feature builds on.

- [X] T001 Write `docs/decisions/0002-self-hosted-nas-production.md`
      (template: `docs/decisions/template.md`) documenting the hosting
      target (self-hosted TrueNAS SCALE), ingress mechanism (Cloudflare
      Tunnel), and CD mechanism (self-hosted GitHub Actions runner)
      decisions from research.md R1/R2 — Governance requires an ADR for
      exactly this class of decision (plan.md Constitution Check)
- [X] T002 [P] Update `infra/branch-protection.md`: fix the stale
      reference to `001-enem-prep-platform` as "this repo's main branch
      for PRs," replacing it with the current reality — `dev` is the
      integration branch, `main` (revived by this feature) is the
      release branch the documented `gh api` command should be applied
      against (research.md R6)
- [X] T003 [P] Create `infra/.env.prod.example` documenting the
      production secret shape from data-model.md's secrets table
      (`POSTGRES_PASSWORD`, `JWT_SIGNING_KEY`,
      `CORRECTOR_OLLAMA_CLOUD_API_KEY`, `CLOUDFLARE_TUNNEL_TOKEN`,
      `BACKUP_*`) — placeholder values only, mirrors how
      `infra/.env.example` documents the dev shape

**Checkpoint**: Governance/documentation groundwork in place.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The production compose stack and its backup sidecar — every
user story depends on this existing before it can be verified.

**⚠️ CRITICAL**: No user story task can be verified until T004 exists and
runs successfully at least once.

- [X] T004 Create `infra/docker-compose.prod.yml`: `monolith`, `gateway`
      (reusing `infra/nginx/nginx.conf` unchanged), `postgres` (volume
      mounted to a NAS dataset path, not a Docker-managed volume),
      `redis`, `corrector-api`, `corrector-worker`, `cloudflared`
      (`cloudflare/cloudflared:latest`, `tunnel run --token
      $CLOUDFLARE_TUNNEL_TOKEN`), and the generalized `backup` sidecar
      from T005 — **no host ports published anywhere** (research.md R5;
      ingress is exclusively via `cloudflared` → `gateway` inside the
      compose network). Service definitions otherwise mirror
      `infra/docker-compose.yml`/`032`'s `corrector-api`/`corrector-worker`
      blocks, with prod secrets read from the NAS-local `infra/.env`
      (never committed). Validated with `docker compose config` against
      dummy env values — syntactically correct. `restart: unless-stopped`
      throughout (not dev's `on-failure`) for reboot recovery (SC-005).
      Deviation found and fixed while writing T013: `infra/nginx/nginx.conf`
      itself had a separate, pre-existing bug (5 of 7 API domains never
      routed — see PR #80, split out since it also affects dev) that
      would have made this file ship the same gap to production; T004
      itself needed no change once #80 lands, since it reuses the file
      unchanged as planned
- [X] T005 [P] Generalize `ai-corrector/deploy/backup/pg_dump_cron.sh`
      and `ai-corrector/deploy/backup/Dockerfile` to dump the **whole**
      shared `preuni` database (all schemas), not just a
      correction-only database — reuse the existing cron/`pg_dump`
      mechanism, only the target DSN/database name changes (research.md
      R4). Turned out to need zero logic changes: the script was
      already parameterized entirely by `POSTGRES_DB`, dumping whichever
      database that names in full — "correction-only" was purely a
      byproduct of the old standalone `corretor_db` being a separate
      database, not anything the script itself restricted. Only changed
      the `corretor-*` backup filename prefix to `preuni-*` since it now
      backs up the whole shared database, not a correction-only one
- [ ] T006 [MANUAL] Apply `main` branch protection for real: run the
      `gh api` command in `infra/branch-protection.md` (as updated by
      T002) against the actual repo. Requires GitHub admin access on
      `preuni-br/preuni`
- [ ] T007 [MANUAL] On the TrueNAS box itself: confirm Docker/Apps
      support is enabled (Settings → Apps, or `docker compose version`
      over SSH if shell access is enabled), and identify/create the
      dataset path `postgres` and the `backup` sidecar will write to
      (needs to be a real NAS dataset a TrueNAS snapshot/backup task can
      reach, per FR-008)

**Checkpoint**: The production stack definition exists and the NAS is
confirmed ready to run it. User stories can now proceed.

---

## Phase 3: User Story 1 - preuni is reachable on the internet, for real (Priority: P1) 🎯 MVP

**Goal**: `preuni.com.br` resolves to a real, running instance of the
app, over HTTPS, with no port forwarded on the home router.

**Independent Test**: From a network outside the NAS's home network, hit
a health endpoint under `preuni.com.br` and get a valid HTTPS response
(quickstart.md SC-001).

### Implementation for User Story 1

- [ ] T008 [MANUAL] [US1] In the Cloudflare dashboard: create a tunnel
      for the `preuni.com.br` zone (token-only/remotely-managed mode,
      research.md R1), add a public-hostname route
      `preuni.com.br` → `http://gateway:8080` (the internal compose
      service name/port), and copy the tunnel token
- [ ] T009 [MANUAL] [US1] Populate the NAS-local `infra/.env` (gitignored,
      never committed) with real production values per
      `infra/.env.prod.example` (T003): `CLOUDFLARE_TUNNEL_TOKEN` from
      T008, distinct production `POSTGRES_PASSWORD`, `JWT_SIGNING_KEY`
      (≥32 chars), `CORRECTOR_OLLAMA_CLOUD_API_KEY`, and `BACKUP_*`
      values for wherever dumps land (T007's dataset path)
- [ ] T010 [US1] First manual bring-up (before CD exists — T013 replaces
      this for every subsequent deploy):
      `docker compose -f infra/docker-compose.prod.yml up -d --build`
      on the NAS
- [ ] T011 [US1] Verify quickstart.md SC-001: from a network outside the
      NAS's home network, `curl -s -o /dev/null -w "%{http_code}\n"
      https://preuni.com.br/health` returns `200` with a valid
      Cloudflare-issued TLS certificate, and confirm no port is
      forwarded on the home router
- [ ] T012 [US1] Verify quickstart.md SC-005: reboot the NAS, wait for
      it to come back online, re-run T011's `curl` and confirm the
      stack restarted on its own with no manual container start

**Checkpoint**: `preuni.com.br` is live and durable across a reboot —
the MVP of this feature (the actual launch gate, per #68).

---

## Phase 4: User Story 2 - Shipping code means it goes live, not just merges (Priority: P2)

**Goal**: Merging into `main` results in that change running in
production without anyone touching the NAS by hand.

**Independent Test**: Merge a trivial, observable change and confirm it
appears at `preuni.com.br` with zero manual NAS steps
(quickstart.md SC-002).

### Implementation for User Story 2

- [X] T013 [US2] Write `.github/workflows/deploy-prod.yml` per
      `contracts/deploy-pipeline.md`'s contract: triggers on `push` to
      `main` only, `runs-on: [self-hosted, self-hosted-nas]`, steps:
      checkout → `docker compose -f infra/docker-compose.prod.yml build`
      → `up -d` → health-check `gateway`/`monolith` from the runner →
      on health-check failure, leave prior images in place (no prune) —
      does not re-run lint/test (already gated by required status
      checks on both the `dev` PR and the `dev`→`main` promotion PR).
      Deviation: the health-check step originally planned to `curl`
      `localhost:8080` from the runner, but `gateway` publishes no host
      port at all (T004) — fixed to `docker compose exec gateway wget`
      instead, checking from inside the compose network
- [ ] T014 [MANUAL] [US2] Register a self-hosted GitHub Actions runner on
      the NAS with label `self-hosted-nas` (research.md R2) — a
      long-lived container **separate from** `infra/docker-compose.prod.yml`
      (must stay up while the app stack it deploys cycles), with the
      Docker socket mounted so it can run `docker compose` against the
      host daemon, and a full repo checkout on each job
- [ ] T015 [US2] Verify quickstart.md SC-002: make a trivial, observable
      change (e.g. a version string surfaced by the health endpoint),
      PR into `dev` and merge once green, open the `dev`→`main`
      promotion PR (research.md R3) and merge once green and approved,
      watch `deploy-prod.yml` run against the `self-hosted-nas` runner,
      re-run T011's `curl` and confirm the version string changed with
      zero manual NAS commands
- [ ] T016 [US2] Verify FR-006 (contracts/deploy-pipeline.md's rollback
      guarantee): intentionally push a change that fails the build or
      health check, confirm `deploy-prod.yml` does not remove the
      previously-working containers, and confirm `preuni.com.br` keeps
      serving the last-good version throughout

**Checkpoint**: Deploys are automatic and safe against a broken push —
User Stories 1 and 2 both verified.

---

## Phase 5: User Story 3 - Essay grading actually works in production (Priority: P1)

**Goal**: The correction service in production shares the real Postgres
instance (no standalone DB/JWT/API of its own) and a submitted essay
reaches a graded state — closes #69 for real, against a real target.

**Independent Test**: Submit an essay through `https://preuni.com.br`
and observe it reach `graded` (quickstart.md SC-003).

### Implementation for User Story 3

- [X] T017 [US3] Delete `ai-corrector/deploy/docker-compose.prod.yml`
      and `ai-corrector/deploy/Caddyfile` — fully superseded by T004's
      `infra/docker-compose.prod.yml` (closes #69: no more standalone
      Postgres, JWT, or reverse proxy for the correction service).
      `Caddyfile` did not actually exist (dead reference inside the
      deleted compose file, never committed) — nothing to delete there.
      Also removed `ai-corrector/deploy/deploy/postgres-init.sql`, an
      untracked directory (not a file) left over from a stale Docker
      bind-mount path resolution against the now-deleted compose file —
      pure local cruft, not tracked by git
- [ ] T018 [US3] Verify the `correction` schema lives in the same shared
      production database: `docker compose -f
      infra/docker-compose.prod.yml exec -T postgres psql -U
      <prod_user> -d preuni -c "\dn"` (run via the NAS/runner) lists
      both `correction` and `essay` in one database, matching `032`'s
      already-proven dev shape
- [ ] T019 [US3] Verify no public port or reverse proxy remains for the
      correction service in production: `docker compose -f
      infra/docker-compose.prod.yml ps` shows no published port for
      `corrector-api`/`corrector-worker`, and confirm `cloudflared`'s
      only configured public hostname (T008) points at `gateway`, never
      at either corrector service
- [ ] T020 [US3] Run quickstart.md SC-003 end to end against
      `https://preuni.com.br`: register/login a test account, submit an
      essay, poll until it reaches `graded` (or a typed `failed`),
      confirming the same pipeline already proven in `032`/`034`'s
      local verification now works in production

**Checkpoint**: All three user stories verified. Feature complete —
#68 and #69 both actually closeable.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Close out the GitHub issues this feature resolves and leave
documentation accurate.

- [ ] T021 [P] Comment on and close GitHub issue #68, linking this
      feature's branch/PR and quickstart.md's verified Success Criteria
      as evidence
- [ ] T022 [P] Comment on and close GitHub issue #69, linking T017-T020
      as evidence the correction service's production deployment now
      matches the DB-mediated, no-public-API architecture the
      constitution requires
- [X] T023 [P] Update `ai-corrector/README.md`'s Deployment section
      (added in `032`) to name `infra/docker-compose.prod.yml`
      specifically as the production deployment, not just "the shared
      stack" generically
- [ ] T024 Run quickstart.md's full checklist end to end one more time
      after T017's deletions, to confirm nothing regressed

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup (T002 informs T006;
  T003 informs T009) — BLOCKS all user stories; T004 in particular is
  the load-bearing task
- **User Stories (Phase 3-5)**: All depend on Foundational completion.
  US1 (Phase 3) is the true prerequisite for verifying US2 and US3 in
  practice (nothing to deploy-to or grade-against until the stack is
  reachable at all), even though US2/US3 don't strictly *require* US1's
  tasks to be code-complete first
- **Polish (Phase 6)**: Depends on all three user stories being verified
  (T021/T022 close issues based on that evidence)

### User Story Dependencies

- **User Story 1 (P1)**: Depends only on Foundational — this is the MVP
- **User Story 2 (P2)**: Depends on Foundational; practically sequenced
  after US1 since there's nothing to redeploy onto until the stack is
  first up (T010)
- **User Story 3 (P1)**: Depends on Foundational; practically sequenced
  after US1 for the same reason — grading needs a reachable stack to
  submit against

### Parallel Opportunities

- T001/T002/T003 (Setup) in parallel
- T004/T005 (Foundational) in parallel — different files
- T021/T022/T023 (Polish) in parallel

---

## Parallel Example: Setup Phase

```bash
Task: "Write ADR 0002 for hosting/ingress/CD decisions (T001)"
Task: "Fix infra/branch-protection.md's stale branch reference (T002)"
Task: "Create infra/.env.prod.example (T003)"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (T004 is the actual stack definition)
3. Complete Phase 3: User Story 1 (T008-T012, several `[MANUAL]`)
4. **STOP and VALIDATE**: `preuni.com.br` is reachable and survives a
   reboot
5. This alone resolves #68's actual launch gate

### Incremental Delivery

1. Setup + Foundational → the production stack is defined and the NAS
   is confirmed ready
2. User Story 1 → preuni is live at `preuni.com.br` (MVP, #68 closeable)
3. User Story 2 → deploys become automatic and safe (CD pipeline live)
4. User Story 3 → correction service's production shape is correct and
   proven grading essays for real (#69 closeable)
5. Polish → issues closed with evidence, docs accurate

### A note on `[MANUAL]` tasks

T006, T007, T008, T009, T014 cannot be executed by an agent — they
require the user's Cloudflare account, GitHub org admin access, and
physical/SSH access to their own NAS. Everything else in this list
produces a file this agent can write and the user can review as a
normal PR. Treat `/speckit.implement` on this feature as producing the
config/code half of the work; the `[MANUAL]` tasks are a checklist for
the user to work through alongside it, in the order they appear (T006
and T007 before Phase 3; T008-T009 before T010; T014 before T015).

---

## Notes

- No new tests — infra/ops feature; verification is quickstart.md
  against spec's Success Criteria, referenced directly from each
  story's tasks above
- T004 is the single most load-bearing task — every story's
  verification tasks assume it exists and is correct
- Commit after each phase's file-producing tasks, not each task
  individually — T001-T003 form one coherent "governance/docs" PR;
  T004-T005 form one coherent "prod stack" PR; T013 stands alone as the
  CD workflow; T017 stands alone as the ai-corrector cleanup (mirrors
  `032`'s own commit granularity)
- Avoid committing any real secret value at any point — T003/T009 are
  explicitly placeholder-shape-only vs. real-value-on-the-NAS-only
