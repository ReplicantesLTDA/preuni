# Implementation Plan: Automated ENEM Essay Correction API (MVP)

**Branch**: `001-enem-correction-api` | **Date**: 2026-05-28 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification at `/specs/001-enem-correction-api/spec.md` (v2.0.0)
**Constitution**: `.specify/memory/constitution.md` (v2.0.0)

## Summary

Build a B2C MVP API that lets a registered, verified Brazilian high-school student submit an
ENEM essay + prompt theme, receive an immediate 202 acknowledgement, and poll until a structured,
matrix-conformant correction is produced asynchronously by a worker that calls an LLM (Ollama
Cloud, default model **Kimi K2 1T**) through a provider-agnostic abstraction. The
correction is broken down by the 5 official competencies, each with a verbatim excerpt, a pt-BR
justification, and an improvement path; eliminatory criteria are first-class flags.

The MVP uses a **single Postgres deployment** as both system-of-record and task queue (no Redis,
no Celery, no RabbitMQ), a **single-grader pass** per correction with a data model already
designed for multi-grader on day one, and ships through a single `docker-compose.yml` on a single
VPS behind Caddy with daily `pg_dump` backups to S3-compatible storage.

## Technical Context

**Language/Version**: Python 3.12+.
**Primary Dependencies**: FastAPI, SQLAlchemy 2.x (async), Alembic, asyncpg, pydantic 2.x, httpx,
jsonschema, PyJWT, argon2-cffi, structlog, prometheus-client, opentelemetry-sdk, pytest,
pytest-asyncio, ruff, mypy.
**Storage**: PostgreSQL 16+ (single instance) — durable state **and** task queue via
`SELECT ... FOR UPDATE SKIP LOCKED`.
**LLM provider**: Ollama Cloud (HTTPS) for MVP, local Ollama for dev. Default model
`kimi-k2:1t` (configurable per env). Accessed via a single `LLMProvider` Protocol;
swapping to Anthropic / OpenAI / Google requires only a new adapter implementation.
**Testing**: pytest with three suites — `tests/unit/`, `tests/integration/`, `tests/golden/`.
The golden suite runs the full pipeline against `tests/golden/` examples and computes
Constitution IX MVP-tier metrics.
**Target Platform**: Linux server (single VPS), containerized via Docker. Local dev on macOS /
Linux via `docker-compose up`.
**Project Type**: Web service (FastAPI API + background worker, single repo, single project).
**Performance Goals**: Submission ack p95 < 500 ms; end-to-end correction p95 < 90 s; 5 concurrent
in-flight corrections; 99% API-surface uptime decoupled from worker uptime.
**Constraints**: Constitution v2.0.0 — provider-neutral abstraction (Art. II), temperature ≤ 0.2
with fixed seed when supported (Art. III), JSON Schema with 2-attempt corrective retry then typed
failure (Art. IV), strict layer separation: LLM only via abstraction, never touches DB (Art. VI),
PII-scrubbed structured logs only (Art. VII), prompts as versioned files (Art. VIII), MVP-tier
golden-dataset gates: total MAE ≤ 120, per-comp MAE ≤ 60, ≥ 70% within ±120 (Art. IX), B2C-only
+ Zero Data Retention provider config + parental consent for minors + 15-day deletion
(Art. X), library-first per competency (Art. XI).
**Scale/Scope**: MVP target ≤ 50 corrections/day across early users; data model and quota ledger
sized for an order of magnitude headroom; no horizontal scaling planned for MVP.

## Constitution Check

*GATE: must pass before Phase 0. Re-check after Phase 1 design (bottom of file).*

| Article | Principle | Gate Status (pre-Phase 0) | How Plan Satisfies |
|---------|-----------|--------------------------|--------------------|
| I | ENEM matrix fidelity (NN) | PASS | Output schema enforces scores ∈ `{0,40,80,120,160,200}`, sum check, eliminatory flags as typed enum; DB `CHECK` constraint on `final_score = c1+c2+c3+c4+c5`. |
| II | Provider neutrality (NN) | PASS | `LLMProvider` Protocol in `corrector/llm/base.py`; `OllamaProvider` is the only MVP adapter; correction pipeline depends on Protocol, not adapter; vendor name absent from public API surface. |
| III | Determinism & auditability (NN) | PASS | `temperature=0.1`, `seed = int.from_bytes(sha256(correction_id)[:8], "big")`, full inference params persisted in `grader_passes` + `prompt_version` + `model_identifier`. |
| IV | Structured output contract (NN) | PASS | `schemas/v1/correction_output.schema.json` validated post-call; 2-attempt corrective retry then typed `schema_violation` failure. Import-linter forbids free-form fallback. |
| V | Per-competency justification | PASS | Output schema requires `excerpt`, `justification_pt_br`, `improvement_path_pt_br` (when score < 200) per competency; pipeline verifies excerpt is a verbatim normalized-substring of the essay. |
| VI | Strict layer separation | PASS | `api/`, `workers/`, `corrector/` (pure), `db/`, `auth/`. Import-linter contract forbids `corrector → db | api`; CI gate. |
| VII | Observability + PII privacy (NN) | PASS | `observability/logging.py` filter strips PII; logs key on opaque UUIDs (`user_id`, `correction_id`); audit log payload allowlist excludes `essay_text`, `email`. |
| VIII | Prompts-as-code | PASS | `corrector/prompts/v{semver}/` markdown; runtime loader caches; prompt SemVer persisted per grader pass; CI requires metric delta on prompt-touching PR. |
| IX | Golden-dataset MVP-tier gate (NN) | PASS | `tests/golden/essay_br/test/` (Essay-BR test split, MIT) + `tests/golden/inep_exemplary/` (INEP nota-1000 secondary set) feed `harness.py`. MVP-tier gate (MAE-total ≤ 120, MAE-per-comp ≤ 60, hit-rate ≥ 70% within ±120) applies to the Essay-BR test split AND a per-band stratified slice (0–400 / 401–700 / 701–1000). INEP secondary set is a hard sanity gate — any failure blocks merge. Long-term targets reported, non-blocking. |
| X | LGPD, ECA, B2C, ZDR, minor consent (NN) | PASS | Provider config asserts ZDR + no-train; consent record FK on `users` for minors; 15-day deletion documented; zero PII in logs. |
| XI | Library-first per competency | PASS | `corrector/competencies/c{1..5}/` — each owns its prompt fragment + parser + tests + golden slice. Orchestrator in `corrector/graders/single_grader.py` composes only. |

**Pre-Phase 0 gate: PASS.** No complexity-tracking entries.

## Project Structure

### Documentation (this feature)

```text
specs/001-enem-correction-api/
├── plan.md                                 # This file
├── research.md                             # Phase 0 — open decisions resolved
├── data-model.md                           # Phase 1 — entities, relationships, indexes
├── quickstart.md                           # Phase 1 — bring up dev env, run e2e smoke
├── contracts/                              # Phase 1 — versioned interface contracts
│   ├── openapi.yaml
│   ├── correction_output.schema.json
│   └── error_codes.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

Single-project layout, library-first per Constitution Article XI:

```text
src/
├── api/                              # FastAPI surface; thin
│   ├── main.py                       # app factory, lifespan, middleware
│   ├── deps.py                       # auth/quota dependencies
│   ├── routes/
│   │   ├── auth.py                   # register, login, refresh, logout, verify
│   │   ├── me.py                     # GET /me (profile + quota)
│   │   ├── corrections.py            # POST, GET id, GET list, POST reevaluate
│   │   └── health.py                 # /healthz, /readyz
│   ├── schemas/                      # pydantic request/response models
│   └── errors.py                     # typed HTTP error -> error code mapping
│
├── workers/
│   └── correction_worker.py          # Postgres-queue consumer; LISTEN/NOTIFY + poll fallback
│
├── corrector/                        # PURE business logic; no I/O
│   ├── competencies/
│   │   ├── c1/                       # Domínio da norma culta
│   │   │   ├── prompt_fragment.md
│   │   │   ├── parser.py
│   │   │   └── tests/
│   │   ├── c2/                       # Compreensão da proposta
│   │   ├── c3/                       # Seleção e organização de argumentos
│   │   ├── c4/                       # Mecanismos linguísticos
│   │   └── c5/                       # Proposta de intervenção
│   ├── graders/
│   │   ├── single_grader.py          # MVP: 1 pass per correction
│   │   └── aggregator.py             # MVP: identity; multi-grader ENEM rule documented
│   ├── llm/
│   │   ├── base.py                   # LLMProvider Protocol
│   │   ├── errors.py                 # error taxonomy
│   │   ├── ollama.py                 # OllamaProvider (Cloud + local)
│   │   └── fake.py                   # in-memory adapter for tests
│   ├── prompts/
│   │   ├── loader.py                 # versioned loader + cache
│   │   └── v1.0.0/
│   │       ├── system.md
│   │       ├── eliminatory_check.md
│   │       └── per_competency_assembly.md
│   ├── prevalidation/
│   │   ├── length.py
│   │   ├── language.py               # pt-BR detection
│   │   └── theme.py
│   └── pipeline.py                   # composes prevalidation + grader + schema validation
│
├── schemas/                          # JSON Schemas, versioned, mirrored from contracts/
│   └── v1/
│       └── correction_output.schema.json
│
├── auth/
│   ├── passwords.py                  # argon2-cffi hash/verify
│   ├── jwt.py                        # PyJWT issue/verify access tokens
│   ├── refresh_tokens.py             # SHA-256 hashed, revocable
│   ├── verification.py               # email token issue/verify
│   └── quota.py                      # quota check + atomic decrement
│
├── db/
│   ├── engine.py                     # async SQLAlchemy engine
│   ├── models/                       # one module per entity
│   ├── repositories/                 # narrow data-access methods
│   └── migrations/                   # Alembic
│
├── config/
│   └── settings.py                   # pydantic-settings
│
└── observability/
    ├── logging.py                    # structlog + PII scrub filter
    ├── metrics.py                    # prometheus-client registry
    └── tracing.py                    # OpenTelemetry setup

tests/
├── unit/                             # pure-logic tests; no DB, no LLM
├── integration/                      # against real Postgres + fake LLM
└── golden/                           # full pipeline against reference essays
    ├── conftest.py
    ├── harness.py                    # MAE / hit-rate / stratified bands; emits JUnit + JSON
    ├── test_golden_dataset.py        # pytest entrypoint; iterates examples
    ├── essay_br/                     # Essay-BR extended corpus (Marinho et al., 2022, MIT)
    │   ├── README.md                 # upstream URL + pinned commit hash
    │   ├── ingest.py                 # one-shot fetch + convert; idempotent, byte-stable
    │   ├── test/<slug>/              # MERGE-GATING split
    │   │   ├── essay.txt
    │   │   ├── prompt_theme.txt
    │   │   ├── motivational_texts.txt
    │   │   └── expected.json
    │   └── valid/<slug>/             # for prompt iteration only; NOT a merge gate
    └── inep_exemplary/<slug>/        # 10–20 INEP nota-1000 essays; secondary sanity check

NOTICE                                # repo-root attribution: Essay-BR MIT license text + cite

deploy/
├── docker-compose.yml                # dev: api + worker + postgres + ollama (local)
├── docker-compose.prod.yml           # prod: api + worker + postgres + caddy
├── Caddyfile                         # TLS + reverse proxy
└── backup/
    └── pg_dump_cron.sh               # daily backup to S3-compatible
```

**Structure Decision**: Single repo, single Python project. Source under `src/` with the
library-first split mandated by Article XI. Workers and API share `corrector/` (pure logic),
`db/`, `auth/`, `observability/`, `config/`. Only `api/` and `workers/` import `db/`; `corrector/`
is pure (enforced by import-linter).

## Phase 0, 1, 2 outputs

- **Phase 0 — Research**: see [research.md](./research.md). Resolves the 5 open planner decisions:
  worker wake-up strategy, quota counter design, reverse-proxy choice, dependency manager, JWT +
  password-hash library choices. Also captures Kimi K2 1T rationale and the
  Postgres-as-queue pattern citations.
- **Phase 1 — Design & Contracts**:
  - [data-model.md](./data-model.md): 7 tables, foreign keys, indexes, state machine,
    multi-grader-ready shape.
  - [contracts/openapi.yaml](./contracts/openapi.yaml): REST contract — auth, profile,
    corrections.
  - [contracts/correction_output.schema.json](./contracts/correction_output.schema.json):
    versioned LLM output schema (v1).
  - [contracts/error_codes.md](./contracts/error_codes.md): typed error taxonomy.
  - [quickstart.md](./quickstart.md): bring up dev env, run e2e smoke, run golden harness.
- **Phase 2 — Tasks**: produced by `/speckit.tasks` (not by this command).

## Complexity Tracking

No entries. Every Constitution gate passes by construction.

The `grader_passes` table is the only piece beyond a strict single-grader MVP; it is justified by
Article-XI-aligned future-proofing for the ENEM two-grader rule without a future migration. MVP
behavior is exactly one `grader_passes` row per `corrections` row, with values copied to the
aggregate columns.

## Post-Design Constitution Re-check

*Re-evaluated after Phase 1 design artifacts (data-model.md, contracts/, quickstart.md).*

| Article | Concern surfaced in Phase 1? | Resolution |
|---------|------------------------------|------------|
| I | DB stores `final_score` + per-comp scores; risk of drift from sum. | `corrections` has a `CHECK (final_score = c1+c2+c3+c4+c5)` constraint, applied via Alembic. |
| II | OpenAPI exposes no vendor field; only `model_identifier` (opaque string). | OK — vendor name appears only in env config, never in API surface. |
| III | Seed reproducibility across worker restarts. | Seed = `int.from_bytes(sha256(correction_id.bytes)[:8], "big")`. Stable, audit-friendly. |
| IV | Two-attempt retry path traceable in audit log. | Each LLM call writes `correction_audit_logs` rows with `event_type ∈ {llm_started, llm_completed, schema_failed, retry, retry_exhausted}`. |
| V | Excerpt-is-verbatim runtime check sits in pipeline. | Pipeline rejects a parsed output where any cited `excerpt` is not a normalized substring of the essay; counts as schema failure → corrective retry. |
| VI | Layer-separation enforced. | Import-linter config in `pyproject.toml` forbids `corrector` → `db`/`api` imports; CI gate. |
| VII | PII in `audit_logs.event_payload` jsonb? | Filter at write site — `event_payload` is allowlisted keys only; `essay_text` and `email` are forbidden in jsonb payloads. |
| VIII | Prompt change vs schema change orthogonal. | Prompt version stored on `grader_passes.prompt_version`; schema version stored as `output_schema_version` on the same row. |
| IX | Golden harness emits MVP-tier + per-band-stratified + INEP-secondary reports; CI blocks on (a) overall MVP-tier regression, (b) per-band MVP-tier regression in any band, (c) any INEP-secondary failure. Long-term metrics reported, non-blocking. | `tests/golden/harness.py` emits three reports; CI workflow asserts all three gates. |
| X | Deletion within 15 days covers essay + audit + derived. | `users.id` cascades to `corrections`, `grader_passes`, and to `correction_audit_logs` (essay_text purged, hash retained). Documented in quickstart. |
| XI | Orchestrator (`graders/single_grader.py`) contains no scoring logic. | Composes prevalidation + LLM call + per-competency parsers. Lint rule: no integer literal in `{40,80,120,160,200}` outside `competencies/`. |

**Post-design gate: PASS.** No complexity-tracking entries needed.
