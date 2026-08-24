# Data Model: Corrector Service Integration

This feature does not add, remove, or alter any application-level table,
column, or business entity — `correction.correction_jobs` and
`correction.corrections` keep the shape defined in
`014-constitution-alignment-refactor`'s data-model.md and the Alembic
migrations that create them. What changes is *where* that schema lives
and *how* the two services that own/consume it are deployed.

## Correction job (`correction.correction_jobs`)

Unchanged. See
`specs/014-constitution-alignment-refactor/contracts/internal-bridge.md`
for the authoritative column-level contract. Relevant to this feature
only insofar as its physical location moves from a standalone
`corretor_db` database (on ai-corrector's own Postgres container) to the
`correction` schema of the shared `preuni` database.

## Correction service deployment (new conceptual entity for this feature)

Not a database table — the set of runtime processes and infrastructure
dependencies that constitute "the correction service" as a deployable
unit. This is what the feature actually changes the shape of.

| Attribute | Before this feature | After this feature |
|---|---|---|
| API process | `corretor-api` container, own compose file | `corrector-api` service in `infra/docker-compose.yml` |
| Worker process | `corretor-worker` container, own compose file | `corrector-worker` service in `infra/docker-compose.yml` |
| Database | Own Postgres container (`corretor-postgres`), database `corretor_db` | Shared `postgres` service already in `infra/docker-compose.yml`, database `preuni`, schema `correction` |
| Network | `corretor-network` (isolated) | `preuni`'s default compose network (shared with `monolith`, `gateway`, `postgres`, `redis`) |
| Public ingress | Own Caddy reverse proxy, TLS, `/auth/*` rate limit | None — internal network only, matching FR-004 |
| Bring-up | `cd ai-corrector && docker compose -f deploy/docker-compose.yml up` | `make dev` (repo root) |
| Migrations | `alembic upgrade head` run manually inside `ai-corrector/` | Same Alembic command, invoked against the shared `DATABASE_URL`, wired into the shared bring-up flow (see quickstart.md) |
| Health checks | Its own compose's `healthcheck:` blocks | Reused as-is, referenced from `infra/docker-compose.yml`'s `depends_on: condition: service_healthy` |

No state transitions apply — this is static deployment configuration,
not a runtime state machine.

## Out of scope

- `correction.corrections`, `correction.grader_passes`,
  `correction.correction_audit_logs` — unchanged shape, unchanged
  ownership (correction service only), just relocated alongside
  `correction_jobs`.
- `essay.essay_submissions` — unchanged; still owned by the monolith,
  still references `correction_job_id` as an opaque UUID.
