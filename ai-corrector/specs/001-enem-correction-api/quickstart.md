# Quickstart — ENEM Essay Correction API (MVP)

**Date**: 2026-05-28
**Audience**: developers booting the project from a fresh clone, plus reviewers running the
end-to-end smoke.

## Prerequisites

- Docker + docker-compose v2
- `uv` ≥ 0.4 installed locally (for editor/IDE workflows; CI uses the same)
- An Ollama Cloud API key OR enough local horsepower for `qwen3:14b`
- 8 GB RAM minimum; 16 GB recommended if running local Ollama

## 1. Bring up the dev environment

```bash
git clone <repo> ai-corrector
cd ai-corrector
cp .env.example .env       # fill in OLLAMA_BASE_URL, OLLAMA_API_KEY (if Cloud), SMTP_*
docker compose up -d        # starts: api, worker, postgres, optionally local ollama
docker compose logs -f api
```

`.env` defaults:

```
DB_URL=postgresql+asyncpg://corretor:corretor@postgres:5432/corretor
OLLAMA_BASE_URL=https://ollama.cloud   # or http://ollama:11434 for local
OLLAMA_API_KEY=                         # set for cloud; empty for local
LLM_MODEL_ID=kimi-k2:1t       # use qwen3:14b for local dev
LLM_TEMPERATURE=0.1
LLM_MAX_TOKENS=2048
LLM_TIMEOUT_S=75
JWT_SECRET=change-me-in-prod
JWT_ACCESS_TTL_S=900                    # 15 min
JWT_REFRESH_TTL_S=2592000                # 30 days
PROMPT_VERSION=                          # empty = use newest on disk
SMTP_HOST=
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=noreply@corretor-enem.local
LOG_LEVEL=INFO
OTEL_EXPORTER=console                    # console | otlp
```

## 2. Run migrations

```bash
docker compose exec api uv run alembic upgrade head
```

## 3. End-to-end smoke

```bash
# Register
curl -X POST http://localhost:8000/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"aluno@example.com","password":"correctHorseBatteryStaple12!"}'

# Verify email — pull token from docker compose logs api (SMTP_HOST empty = log delivery)
TOKEN=<paste verification token from logs>
curl -X POST http://localhost:8000/auth/verify-email \
  -H 'Content-Type: application/json' \
  -d "{\"token\":\"$TOKEN\"}"

# Login
curl -s -X POST http://localhost:8000/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"aluno@example.com","password":"correctHorseBatteryStaple12!"}' | tee /tmp/login.json
ACCESS=$(jq -r .access_token /tmp/login.json)

# Profile / quota
curl -s -H "Authorization: Bearer $ACCESS" http://localhost:8000/me | jq

# Submit
curl -s -X POST http://localhost:8000/corrections \
  -H "Authorization: Bearer $ACCESS" \
  -H 'Content-Type: application/json' \
  -d @tests/golden/examples/exemplo-base/submit.json | tee /tmp/sub.json
CID=$(jq -r .correction_id /tmp/sub.json)

# Poll until done
while true; do
  STATUS=$(curl -s -H "Authorization: Bearer $ACCESS" \
    http://localhost:8000/corrections/$CID | jq -r .status)
  echo "status=$STATUS"
  case "$STATUS" in completed|failed) break;; esac
  sleep 2
done

# Inspect completed payload
curl -s -H "Authorization: Bearer $ACCESS" http://localhost:8000/corrections/$CID | jq
```

Expected: `status` transitions `pending → processing → completed` within the 90-second p95
budget for a small model on a workstation; the `completed` payload conforms to
`contracts/correction_output.schema.json` once unwrapped.

## 4. Run the test suites

```bash
# unit + integration (no real LLM)
docker compose exec api uv run pytest tests/unit tests/integration -v

# golden — pipeline against FakeProvider (fast, every PR)
docker compose exec api uv run pytest tests/golden -k fake -v

# golden — real provider metric check (slow, paid)
OLLAMA_API_KEY=$REAL_KEY docker compose exec api \
  uv run pytest tests/golden -k real --golden-tier=mvp
```

### Ingesting the Essay-BR corpus (one-shot)

The Essay-BR corpus (Marinho, Anchiêta & Moura, 2022; MIT) is already committed under
`tests/golden/essay_br/{test,valid}/`. To regenerate (e.g., after bumping the upstream pin):

```bash
docker compose exec api uv run python tests/golden/essay_br/ingest.py
# add --force to overwrite an existing tree
```

The script verifies the upstream tarball against the pinned commit hash + SHA-256, emits
byte-identical files, and writes `MANIFEST.json`. CI re-runs the manifest check to guard
against drift.

### Golden harness report shape

```
=== Golden Harness Report ===
Essay-BR test split (merge-gating)
  overall:
    total-score MAE     : <value>  (gate ≤ 120)
    per-comp MAE        : <value>  (gate ≤ 60)
    within ±120 hit rate: <value>  (gate ≥ 70%)
  band 0–400  (n=<>): total MAE <>, per-comp MAE <>, hit rate <>   (each gate as above)
  band 401–700(n=<>): ...
  band 701–1000(n=<>): ...

INEP exemplary (secondary sanity, per-essay gate ±120 total / ±60 per comp)
  pass <m>/<n>; failures: [<slug>, ...]

Long-term progress (non-blocking)
  total MAE   : <value>  (target ≤ 80)
  per-comp MAE: <value>  (target ≤ 40)
  within ±80  : <value>  (target ≥ 80%)
```

CI blocks merge on any of: overall MVP-tier regression, per-band MVP-tier regression in any
band, or any INEP-secondary per-essay failure. Long-term targets are reported only.

### Note on `valid/` and `train/`

- `tests/golden/essay_br/valid/` is intended for **prompt iteration** during development. It
  is NOT a CI gate. Use it freely while tuning prompts.
- The Essay-BR `train/` split is **never** used here. Prompt-engineering against the test
  split contaminates the gate; against the train split would be the easiest way to do that
  invisibly. We avoid both by not ingesting `train/`.

## 5. Production deployment (single VPS)

```bash
# on the VPS
git clone <repo> /opt/ai-corrector
cd /opt/ai-corrector
cp .env.example .env.prod   # set OLLAMA_API_KEY, JWT_SECRET, SMTP_*, BACKUP_*
docker compose -f deploy/docker-compose.prod.yml --env-file .env.prod up -d
docker compose -f deploy/docker-compose.prod.yml exec api uv run alembic upgrade head

# Caddy auto-provisions TLS via Let's Encrypt; ensure :80 and :443 are reachable.
```

### Backups

`deploy/backup/pg_dump_cron.sh` runs daily at 03:00 UTC inside a sidecar container and uploads
`pg_dump -Fc` archives to the S3-compatible bucket configured by `BACKUP_*` env vars (Backblaze
B2 by default). 30-day rolling retention.

Restore drill:

```bash
docker compose -f deploy/docker-compose.prod.yml exec backup \
  sh -c 'aws --endpoint-url $BACKUP_S3_ENDPOINT s3 cp s3://$BACKUP_BUCKET/$(date -d yesterday +%Y-%m-%d)/db.dump /tmp/'
docker compose -f deploy/docker-compose.prod.yml exec postgres \
  pg_restore -U corretor -d corretor --clean --if-exists /tmp/db.dump
```

## 6. Deletion (LGPD 15-day window)

A deletion is initiated via support today (MVP). Operator command:

```bash
docker compose exec api uv run python -m scripts.delete_user --user-id <uuid> --reason 'LGPD request #1234'
```

This:
1. Cascades `DELETE FROM users` — removes the user, all their corrections, grader passes,
   audit-log payloads, refresh + verification tokens, consent record.
2. Emits a final `correction_audit_logs` entry with `event_type='deleted'` for each correction
   before cascade fires (used by the post-MVP `deletion_log` mechanism).
3. Logs the operator action to a separate `operator_actions` table (no PII).

The 15-day SLA is operational; the script itself is synchronous in seconds.

## 7. Troubleshooting

- **Worker doesn't pick up a job**: check `docker compose logs worker`. If `LISTEN` dropped, the
  5-second poll fallback should claim within one poll cycle. If not, restart `worker`.
- **Schema-violation loop**: the audit log will show `retry_exhausted` with the last
  `validator_message` (PII-scrubbed). Increase logging to `DEBUG` and re-run the same essay via
  `dry_run=true` first to rule out a pre-validation regression.
- **Caddy not issuing cert**: confirm DNS A-record points to the VPS, and the host's 80/443 are
  not firewalled.
- **Golden tier failing locally on Qwen3 14B**: expected — the 14B model is below MVP-tier; use
  it for plumbing, not for the metric gate. Run the metric gate against 72B (Cloud).
