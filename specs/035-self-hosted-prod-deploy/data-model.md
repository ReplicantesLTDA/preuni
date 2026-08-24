# Data Model: Self-Hosted Production Deployment

This feature adds no application-level table, column, or business
entity — the production database is the same schema set
(`auth`/`user`/`essay`/`correction`/`gamification`/`social`) already
defined by existing migrations, just running on real NAS-backed storage
instead of a dev container volume. What this feature actually models is
deployment/ops configuration, not data.

## Production stack topology (conceptual, not a DB entity)

| Component | Role | Reachable from |
|---|---|---|
| `cloudflared` | Outbound tunnel to Cloudflare edge; the only thing initiating a connection off the NAS for ingress purposes | Cloudflare (outbound only) |
| `gateway` (nginx) | Same config as dev — single public entry point | `cloudflared` only |
| `monolith` | Same image/build as dev | `gateway`, `postgres`, `redis` |
| `postgres` | Same schemas as dev, NAS-backed volume | `monolith`, `corrector-api` (migrations), `corrector-worker`, backup sidecar |
| `redis` | Same as dev | `monolith` |
| `corrector-api` | Health/metrics + one-off Alembic migrations only, per `032` | `postgres` (no public port) |
| `corrector-worker` | Claims correction jobs, calls Ollama Cloud | `postgres`, Ollama Cloud (outbound) |
| `backup` sidecar | Scheduled `pg_dump` of the whole `preuni` database | `postgres` |
| GitHub Actions runner | Long-lived, separate from the app stack; executes deploy jobs | GitHub (outbound only) |

## Secrets (NAS-local `.env`, never committed)

| Variable | Used by | Notes |
|---|---|---|
| `POSTGRES_PASSWORD` | `postgres`, `monolith`, `corrector-api`/`worker` | Distinct value from dev's |
| `JWT_SIGNING_KEY` | `monolith` | Distinct value from dev's, ≥32 chars per existing config validation |
| `CORRECTOR_OLLAMA_CLOUD_API_KEY` | `corrector-worker` | Same key as verified working in `034`, or a production-tier key if usage warrants a separate one later |
| `CLOUDFLARE_TUNNEL_TOKEN` | `cloudflared` | From the Cloudflare dashboard's tunnel setup |
| `BACKUP_S3_*` (or NAS-local path) | `backup` sidecar | Wherever dumps land — see research.md R4 |

## GitHub Actions Secrets (CI/CD-side, distinct store from the NAS `.env`)

| Secret | Purpose |
|---|---|
| Runner registration token | One-time, used only to register the self-hosted runner — not a long-lived secret the runner reads per-job |

No application runtime secret needs to live in GitHub Actions Secrets —
the runner executes on the NAS where the real `.env` already exists;
GitHub only needs what it takes to dispatch the job to that runner.

## Out of scope

- No new Postgres schema, table, or migration.
- No new API endpoint or contract change to the monolith or correction
  service — this feature is entirely about *where* the existing,
  already-correct system runs.
