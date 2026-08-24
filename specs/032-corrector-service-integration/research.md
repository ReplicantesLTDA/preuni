# Research: Corrector Service Integration

## R1: Single Postgres instance vs. keeping separate containers on a shared network

**Decision**: Point the correction service at the *same* `postgres` service
already defined in `infra/docker-compose.yml` (database `preuni`, the
`POSTGRES_DB`/`POSTGRES_USER`/`POSTGRES_PASSWORD` values already there),
using the `correction` schema Alembic's own migration chain already
creates. Delete ai-corrector's own `postgres` container, volume, and
network entirely.

**Rationale**: `backend/app/internal/essay/repository/reconciler.go`
already executes a single SQL query that does
`FROM essay.essay_submissions es JOIN correction.correction_jobs cj ON ...`
— a cross-schema join in one round trip. That is only valid against one
database. Keeping two separate Postgres containers (even on a shared
Docker network) would require rewriting the reconciler to do two
round-trips against two connections, which is a larger, riskier change
this feature has no reason to make — the join already works today
whenever both schemas happen to live in the same database (e.g. in a
test harness), it's only the *deployed* topology that's wrong.

**Alternatives considered**:
- *Two containers, shared network, cross-database query via
  `postgres_fdw` or dblink*: adds an extra moving part and a new failure
  mode for no benefit — the whole point of the DB-mediated bridge
  (chosen over HTTP in `014`'s research.md #1) was to avoid a network
  call between the two systems in the first place.
- *Two containers, application-level two-step reconcile (query jobs,
  then query submissions, join in Go)*: works, but changes
  already-tested application code for a purely infra-driven feature;
  rejected as out of scope (spec Assumptions: "this feature changes
  deployment/infrastructure wiring, not the correction pipeline's
  grading logic").

## R2: Migration ordering between the monolith's hand-written SQL and Alembic

**Decision**: The correction service's Alembic chain is self-contained
(`CREATE SCHEMA IF NOT EXISTS correction` is the first statement of its
own migration, `002_correction_jobs_schema.py`) and does not depend on
anything the monolith's `essay` schema migration creates. The reverse is
also true — `essay.essay_submissions` references
`correction_job_id UUID` as an opaque column with no foreign key into
`correction.correction_jobs` (confirmed in
`infra/migrations/essay/001_create_essay_submissions.sql`'s comment:
"FK lives in the correction service's own schema; opaque reference
here"). Neither migration chain has a hard dependency on the other
having already run. `make dev`'s `dev` target runs `make migrate` (the
Go-side raw-SQL runner) once postgres is healthy; this plan adds running
Alembic's `upgrade head` at the equivalent point, and the two can run in
either order or in parallel — no new sequencing logic is needed beyond
"both complete before either service's traffic depends on the other's
schema existing," which is naturally satisfied by both finishing before
`monolith`/`corrector-worker` containers start serving/claiming jobs.

**Rationale**: Avoids inventing a migration-ordering coordinator for two
schemas that were already designed (in `014`) to have no FK dependency
between them specifically so they *could* migrate independently.

**Alternatives considered**:
- *A single unified migration runner*: over-engineering — two
  ecosystems (raw SQL + Alembic) already work independently in CI today;
  unifying them is a much larger change than this feature's scope.

## R3: The unused `preuni_monolith` role and `grant-correction-jobs.sql`

**Decision**: Do not create a `preuni_monolith` role for this feature.
Both services connect using the existing shared `POSTGRES_USER`
(`preuni`, already a superuser-equivalent owner role in the dev/compose
Postgres image) — the same role the monolith already uses for the
`essay`/`auth`/etc. schemas. `infra/migrations/grant-correction-jobs.sql`
is left in place, unexecuted, with a comment clarifying it's an optional
least-privilege hardening step for a deployment that has actually
provisioned a separate `preuni_monolith` role — which is not this
project's current dev or CI setup.

**Rationale**: The constitution's Security & Secrets principle explicitly
states "No enterprise-specific security theater: SSO, RBAC, and
centralized secret-vault tooling ... are deliberately NOT required
here — preuni has no multi-user staff access to gate." A role-per-service
privilege-separation script that was never actually wired up (grep
confirms `preuni_monolith` appears nowhere outside that one file) is
exactly this kind of unused theater. Introducing it now, for a feature
about *removing* accidental complexity from ai-corrector's deployment,
would cut against the feature's own point. If a future deployment
target genuinely needs role separation, that's a separate, deliberate
decision with its own ADR — not a side effect of this integration.

**Alternatives considered**:
- *Actually create the `preuni_monolith` role and wire the grant script
  into `make migrate`*: technically closer to the script's original
  intent, but adds a new secret (a second DB password) and a new startup
  dependency for no behavioral benefit today; deferred.
- *Delete the grant script*: rejected — it's harmless, already correct
  for the day someone does want role separation, and deleting working
  prior art isn't necessary to fix the actual problem (the missing
  compose wiring).

## R4: What happens to ai-corrector's own Caddy/TLS/rate-limit deploy config

**Decision**: Delete `ai-corrector/deploy/Caddyfile` and remove the
`api` service's public port mapping in favor of the internal compose
network only. `ai-corrector/deploy/docker-compose.yml` and
`ai-corrector/deploy/postgres-init.sql` are deleted; their content is
superseded by the two new services added to `infra/docker-compose.yml`.

**Rationale**: Directly required by spec FR-003/FR-004/User Story 3 — a
service with no public API surface (README already states this) should
not carry TLS termination or auth-endpoint rate-limiting for endpoints
that were removed. NGINX (`infra/nginx/nginx.conf`) is already preuni's
one public gateway and already does not route to the correction service
— nothing needs to change there.

**Alternatives considered**:
- *Keep the Caddy config for a hypothetical future public correction
  API*: rejected — YAGNI; if a public correction API is ever needed
  again, that's a product decision requiring its own spec, not a reason
  to keep dead config around now.

## R5: Health check wiring for the correction service in shared compose

**Decision**: `corrector-api` keeps its existing `HEALTHCHECK`
(`/healthz`) from `ai-corrector/Dockerfile`, reused as compose's
`depends_on: condition: service_healthy` the same way `monolith` and
`gateway` already depend on `postgres`/`redis` being healthy.
`corrector-worker` has no HTTP surface — it depends on `postgres`
(`service_healthy`) only, matching how the monolith itself is wired.

**Rationale**: Reuses infrastructure the Dockerfile already ships (no new
health-check logic to write) and matches the existing pattern in
`infra/docker-compose.yml` exactly (FR-007).

**Alternatives considered**: None — this is the already-established
pattern in the file being edited; no reason to deviate.
