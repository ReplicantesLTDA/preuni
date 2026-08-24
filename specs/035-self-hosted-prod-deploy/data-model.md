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
| `minio` | Self-hosted S3-compatible object store, avatars (research.md R8) | `monolith` (presign only, no live connection needed), the uploading client directly (public `storage.preuni.com.br`, via `cloudflared`) |
| `minio-init` | One-shot: creates the avatar bucket + sets public-read, then exits | `minio` |
| GitHub Actions runner | Long-lived, separate from the app stack; executes deploy jobs | GitHub (outbound only) |

## Secrets (GitHub Actions repository secrets — research.md R7)

Source of truth for every production credential. `deploy-prod.yml`
writes `infra/.env` on the NAS fresh from these on every deploy; nothing
here is hand-edited on the NAS. Each secret name below is prefixed
`PROD_` in GitHub Settings → Secrets and variables → Actions, to keep
them visibly distinct from any dev/CI secret that might exist.

| Secret name | Used by | Notes |
|---|---|---|
| `PROD_POSTGRES_USER` | `postgres`, `monolith`, `corrector-api`/`worker` | |
| `PROD_POSTGRES_PASSWORD` | same | Distinct value from dev's |
| `PROD_POSTGRES_DB` | same | |
| `PROD_POSTGRES_DATA_PATH` | `postgres` (bind mount) | Real NAS dataset path (tasks.md T007) — not secret data, but NAS-specific, so kept alongside the rest rather than baked into the compose file |
| `PROD_JWT_SIGNING_KEY` | `monolith` | Distinct value from dev's, ≥32 chars per existing config validation |
| `PROD_JWT_ACCESS_EXPIRY_SECONDS` / `PROD_JWT_REFRESH_EXPIRY_DAYS` | `monolith` | |
| `PROD_STORAGE_ENDPOINT` | `monolith` | `storage.preuni.com.br` — the second public tunnel hostname (tasks.md T008a), not `minio:9000` |
| `PROD_STORAGE_ACCESS_KEY` / `PROD_STORAGE_SECRET_KEY` | `monolith`, `minio` (as `MINIO_ROOT_USER`/`MINIO_ROOT_PASSWORD`), `minio-init` | Same value on both sides — MinIO's root credentials double as the S3-API access/secret key |
| `PROD_STORAGE_BUCKET` | `monolith`, `minio-init` | |
| `PROD_MINIO_DATA_PATH` | `minio` (bind mount) | Real NAS dataset path, same pattern as `PROD_POSTGRES_DATA_PATH` |
| `PROD_FROM_EMAIL` / `PROD_MAIL_FROM_NAME` / `PROD_SMTP_*` | `monolith` | Optional — empty `SMTP_HOST` still works (NoopSender), same as dev |
| `PROD_APP_VERIFICATION_BASE_URL` | `monolith` | `https://preuni.com.br/verify-email` |
| `PROD_CORRECTOR_OLLAMA_CLOUD_API_KEY` | `corrector-worker` | Same key verified working in `034`, or a production-tier key if usage warrants a separate one later |
| `PROD_CORRECTOR_OLLAMA_CLOUD_BASE_URL` / `PROD_CORRECTOR_OLLAMA_BASE_URL` / `PROD_CORRECTOR_OLLAMA_MODEL` / `PROD_CORRECTOR_PROMPT_VERSION` | `corrector-worker` | |
| `PROD_CLOUDFLARE_TUNNEL_TOKEN` | `cloudflared` | From the Cloudflare Tunnel created for `preuni-prod` (tasks.md T008) |
| `PROD_BACKUP_S3_*` / `PROD_RETENTION_DAYS` | `backup` sidecar | Wherever dumps land — see research.md R4 |

## GitHub Actions Secrets not covered above

| Secret | Purpose |
|---|---|
| Runner registration token | One-time, used only to register the self-hosted runner (tasks.md T014) — short-lived, not stored long-term as a repository secret the way the `PROD_*` values above are |

## Out of scope

- No new Postgres schema, table, or migration.
- No new API endpoint or contract change to the monolith or correction
  service — this feature is entirely about *where* the existing,
  already-correct system runs.
