# Feature Specification: Corrector Service Integration

**Feature Branch**: `032-corrector-service-integration`
**Created**: 2026-08-24
**Status**: Draft
**Input**: User description: "so last round of coding was almost focusing on tests, quality gates and stuff to make us to be able to write good code across sessions now we gonna focus in the communication between our new ai-corrector and our backend. ive created ai-corrector as a standalone essay corrector, not related to this application. we need to transform ai-corrector to be a service to serve our Go application, it means that some features in ai-corrector needs to be removed/rethinked to make it a service inside our Go backend. i believe you can investigate it better then me, since me myself hasnt touch the ai-corrector in a while any clarification needed asks me, ai-corrector service will be the core of our MVP"

## Investigation Summary

A prior refactor (`014-constitution-alignment-refactor`) already trimmed
ai-corrector's *application* code for internal-only use: its `/auth`,
`/me`, and `POST /corrections` endpoints, and its own identity/quota
tables, were removed. Only `/healthz`, `/readyz`, and `/metrics` remain
public, and the intended integration contract (a shared-Postgres
`correction.correction_jobs` bridge table, no HTTP call between the two
systems) is documented and the Go side of it is fully implemented
(`backend/app/internal/essay/repository/reconciler.go` enqueues and
reconciles against it already).

What never followed that refactor is ai-corrector's **deployment
footprint** — it is still packaged and run as if it were still the
standalone product:

- It boots its own Postgres container (`corretor-postgres`, database
  `corretor_db`) on its own Docker network (`corretor-network`),
  disconnected from the monolith's Postgres. This directly contradicts
  the bridge design: the reconciler's SQL does a cross-schema `JOIN`
  between `essay.essay_submissions` and `correction.correction_jobs` in
  a single query, which is only possible if both schemas live in the
  *same* database — not two separate Postgres instances.
- It ships its own Caddy reverse proxy with automatic TLS and a
  per-IP rate limit scoped to `/auth/*` — a public-internet ingress
  for endpoints that no longer exist.
- Neither `infra/docker-compose.yml` nor `make dev`/`make up` starts
  ai-corrector's API or worker process. Locally, an essay submission is
  enqueued into `correction.correction_jobs` but nothing ever claims and
  grades it, so the reconciler's 10-minute timeout eventually fails
  every submission.

This feature closes that gap: fold ai-corrector into the monolith's
shared infrastructure as an internal service, so the existing bridge
contract that both sides already implement actually works end to end,
locally and in deployment.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A submitted essay gets graded, every time (Priority: P1)

A student submits an essay in the app. The grading pipeline (Go monolith
→ `correction.correction_jobs` → ai-corrector worker → back through the
reconciler) runs to completion without manual intervention, in every
environment the team runs (a developer's laptop, staging, production).

**Why this priority**: This is the MVP's core value loop per the user's
own framing ("ai-corrector service will be the core of our MVP"). Without
it, essay submission is a dead end — students never receive a grade.

**Independent Test**: Bring up the full local stack with a single
command, submit an essay through the monolith's API, and observe the
submission transition from `pending` to `graded` without any manual step
(starting a container by hand, running a migration by hand, etc.).

**Acceptance Scenarios**:

1. **Given** the local dev stack is started via the project's standard
   bring-up command, **When** an essay is submitted through the
   monolith, **Then** the correction service claims the job and the
   submission reaches a terminal graded (or typed-failure) state within
   the existing grading timeout, with no manual step by the developer.
2. **Given** the correction service and the monolith, **When** both are
   running, **Then** they read and write the same `correction` schema in
   the same Postgres database the monolith itself uses — not a separate
   database or instance.
3. **Given** the correction service is temporarily down when a job is
   enqueued, **When** it comes back up, **Then** it picks up the
   pending job on its own (existing poll/backstop behavior) without
   requiring the essay to be resubmitted.

---

### User Story 2 - Developers bring up one stack, not two (Priority: P2)

A developer working on essay grading (backend, mobile, or the correction
pipeline itself) brings up the whole local environment with the project's
existing single entry point, without needing to separately discover and
run ai-corrector's own compose file.

**Why this priority**: Directly enables User Story 1 day to day, and
removes a standing source of "works on my machine" drift — two disjoint
compose stacks and two disjoint Postgres containers is the concrete
reason grading silently doesn't work locally today.

**Independent Test**: A developer who has never touched ai-corrector
before runs the project's documented local setup and gets a working
essay-grading pipeline without reading ai-corrector's own README or
compose file.

**Acceptance Scenarios**:

1. **Given** a clean checkout, **When** a developer runs the project's
   standard local bring-up command, **Then** the correction service's
   API and worker processes are running alongside the monolith, gateway,
   and shared Postgres/Redis, using the same Postgres instance.
2. **Given** the correction service now runs from the shared compose
   stack, **When** a developer inspects running containers, **Then**
   there is no second, disconnected Postgres container or Docker network
   left over from ai-corrector's former standalone setup.

---

### User Story 3 - No public entry point that shouldn't exist (Priority: P3)

An operator or reviewer auditing the system's attack surface can confirm
that ai-corrector, having no end users or public API of its own, has no
internet-facing ingress, TLS termination, or auth-scoped rate limiting of
its own — the monolith's existing gateway is the one and only public
entry point into preuni.

**Why this priority**: Lower urgency than making grading work at all, but
matters for the "internal-only, no public API surface" guarantee the
service's own documentation already claims but its deployment artifacts
still contradict.

**Independent Test**: Review the service's deployment configuration and
confirm no reverse proxy, TLS certificate management, or public port
exposure remains for it outside of what the shared internal stack
requires (container-to-container, not internet-facing).

**Acceptance Scenarios**:

1. **Given** the correction service's deployment configuration, **When**
   reviewed, **Then** it contains no reverse proxy or TLS setup of its
   own, and no rate-limiting rule referencing endpoints that no longer
   exist (e.g. `/auth/*`).
2. **Given** the correction service is deployed, **When** its network
   exposure is inspected, **Then** its ports are reachable only from
   the internal Docker network shared with the monolith, not from the
   public internet.

---

### Edge Cases

- What happens when the correction service's Postgres migration (which
  creates the `correction` schema) has not yet run against the shared
  database when the monolith starts up? The monolith's own migrations
  and the correction service's migrations must both complete, in the
  order the schema dependency requires, before either service serves
  traffic that touches the bridge table.
- What happens if both services attempt to migrate the same shared
  database concurrently during startup? Each service's migrations must
  only ever touch its own schema, and startup ordering must not depend
  on manual sequencing by a developer.
- What happens to essay grading during the migration window while both
  the old standalone stack and the new shared-infra stack could
  theoretically exist? Out of scope for this feature to run both side by
  side; this is a one-way cutover.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The correction service MUST read and write the
  `correction` schema in the same Postgres database instance that the
  Go monolith uses, not a separate database or container.
- **FR-002**: The project's standard local development bring-up command
  MUST start the correction service's API and worker processes alongside
  the existing monolith, gateway, and shared infrastructure, with no
  additional manual steps.
- **FR-003**: The correction service's standalone deployment artifacts
  (its own Postgres container/volume/network, its own reverse proxy with
  TLS and auth-endpoint rate limiting) MUST be removed or replaced by the
  shared-infrastructure equivalent.
- **FR-004**: The correction service MUST remain reachable only from the
  monolith's internal network — it MUST NOT expose a public,
  internet-facing port or ingress of its own.
- **FR-005**: An essay submitted while the correction service is running
  MUST reach a terminal state (graded or typed failure) without any
  manual intervention, within the grading timeout the monolith already
  enforces.
- **FR-006**: Database schema setup for both the monolith (`essay`,
  `auth`, etc. schemas) and the correction service (`correction` schema)
  MUST be able to run against the shared database without one blocking
  or corrupting the other, and without requiring a developer to
  hand-sequence which migration runs first.
- **FR-007**: The correction service's health/readiness signals
  (`/healthz`, `/readyz`) MUST remain available to the shared
  infrastructure's own health checks (e.g. compose `depends_on`
  conditions), consistent with how the monolith and gateway are already
  checked.
- **FR-008**: Existing documentation describing the correction service as
  internal-only with a DB-mediated bridge (its own README, the
  internal-bridge contract) MUST stay accurate after this change — no
  regression back to describing an HTTP integration or a public API.

### Key Entities

- **Correction job** (`correction.correction_jobs`): the existing bridge
  row a student's essay submission is enqueued as; unchanged by this
  feature except for which physical database/instance it lives in.
- **Correction service deployment**: the set of processes (API, worker)
  and their infrastructure dependencies (database, network exposure)
  that make up ai-corrector as it runs in any environment; this is what
  the feature changes the shape of.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer on a clean checkout can go from the project's
  documented single bring-up command to a graded test essay submission
  with zero manual steps outside that command.
- **SC-002**: 100% of essay submissions made while the correction service
  is running reach a terminal state (graded or typed failure) without
  timing out, across repeated local runs.
- **SC-003**: Zero standalone Postgres containers, Docker networks, or
  reverse-proxy/TLS configuration remain attributable to the correction
  service outside the shared infrastructure stack.
- **SC-004**: The correction service exposes zero ports reachable from
  outside the internal Docker network in any environment (local,
  staging, production).

## Assumptions

- The correction service's existing application-level scope trim (no
  `/auth`, `/me`, `/corrections` endpoints; DB-mediated bridge only) from
  `014-constitution-alignment-refactor` is correct and final — this
  feature changes deployment/infrastructure wiring, not the correction
  pipeline's grading logic, prompts, or LLM provider choice.
- "Shared Postgres instance" means the correction service's `correction`
  schema is created inside the same database the monolith already runs
  its own schemas in (the database `infra/docker-compose.yml`'s
  `postgres` service manages), consistent with the reconciler's existing
  cross-schema `JOIN` and with the DB-privilege grants already documented
  in `infra/migrations/grant-correction-jobs.sql`.
- The correction service's choice of LLM provider (local Ollama vs.
  Ollama Cloud) and its own correction-pipeline test suite are unaffected
  by this feature; only how the service is deployed and how it reaches
  Postgres changes.
- Production/staging deployment topology (where these containers actually
  run, e.g. which host or orchestrator) is out of scope beyond the
  requirement that no public ingress remains for the correction service
  itself — this feature does not prescribe a hosting provider.
