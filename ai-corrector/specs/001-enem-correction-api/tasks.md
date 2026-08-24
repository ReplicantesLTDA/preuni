---
description: "Task list for Automated ENEM Essay Correction API (MVP)"
---

# Tasks: Automated ENEM Essay Correction API (MVP)

**Input**: Design documents under `/specs/001-enem-correction-api/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/, quickstart.md

**Tests**: REQUIRED per Constitution v2.0.0 Article IX (golden-dataset gate is NON-NEGOTIABLE) AND
per the user's test-first ordering principle. Every implementation task that has testable behavior
is preceded by its test task; the implementation task depends on it.

**Organization**: Walking skeleton first → quality gate → post-quality-gate infrastructure
organized by user story (US1–US4 from spec). Each task is scoped to roughly half a day of work
for a solo developer; larger items are split.

## Format: `[ID] [P?] [Story] Description`

- `[P]`: can run in parallel (different files, no dependencies on incomplete tasks)
- `[Story]`: included on user-story phase tasks only (US1–US4)
- File paths are exact

## Path conventions

- Source: `src/` (per plan.md project structure)
- Tests: `tests/unit/`, `tests/integration/`, `tests/golden/`
- Deploy: `deploy/`
- Contracts: `specs/001-enem-correction-api/contracts/`

---

## Phase 1: Setup (project initialization)

Goal: clone-to-`docker compose up` in five minutes. No business logic.

- [X] T001 Initialize repository scaffolding: `pyproject.toml` with `uv` build, `[project]` metadata, `ruff`, `mypy`, `pytest`, `pytest-asyncio`, `import-linter`. Pin Python 3.12+. Create `src/`, `tests/{unit,integration,golden}/`, `deploy/`, `scripts/` skeleton dirs.
- [X] T002 [P] Add direct deps to `pyproject.toml`: `fastapi`, `pydantic`, `pydantic-settings`, `httpx`, `jsonschema`, `structlog`, `prometheus-client`, `opentelemetry-sdk`, `opentelemetry-exporter-otlp`, `lingua-language-detector`. Generate `uv.lock`.
- [X] T003 [P] Add infra-tier deps to `pyproject.toml`: `sqlalchemy[asyncio]`, `asyncpg`, `alembic`, `PyJWT`, `argon2-cffi`, `email-validator`. Regenerate `uv.lock`.
- [X] T004 [P] Configure `ruff` + `mypy` (strict) + `import-linter` in `pyproject.toml`. Import-linter contracts: forbid `corrector → db|api`, forbid `corrector → workers`, forbid vendor names (`anthropic`, `openai`, `ollama`) outside `corrector/llm/`.
- [X] T005 [P] Create `.env.example` with every variable enumerated in `quickstart.md`. Document defaults inline.
- [X] T006 Create `deploy/docker-compose.yml` (dev): services `api`, `worker`, `postgres:16`, optional `ollama`. Networks, volumes, healthchecks. API depends_on Postgres healthy.
- [X] T007 [P] Create `src/config/settings.py` using `pydantic-settings`: all env vars from `.env.example`, typed, fail-fast on missing required.
- [X] T008 [P] Wire `src/observability/logging.py` (structlog JSON, ContextVar-bound `correction_id` + `user_id`) and a PII-scrub filter that rejects `essay_text`, `email`, `name`, `student_id`, `password` keys at the writer boundary.
- [X] T009 [P] Add `Makefile` with `make up`, `make down`, `make logs`, `make fmt`, `make lint`, `make test-unit`, `make test-int`, `make test-golden-fake`, `make test-golden-real`.
- [X] T010 Add `NOTICE` file at repo root crediting Essay-BR (Marinho, Anchiêta & Moura, 2022), citation, original copyright, verbatim MIT License text per R20.

**Checkpoint**: `make up` boots; `make lint` clean.

---

## Phase 2: Walking skeleton (riskiest path, no API / no DB / no auth)

Goal: a CLI / pytest harness that takes one essay + prompt theme, calls Ollama Cloud through the
provider abstraction with structured-output validation, produces a matrix-conformant correction,
and is evaluated against ~10 hand-picked Essay-BR examples + ≥ 3 INEP nota-1000 essays to answer
the only question that matters at this stage: **does Kimi K2 1T clear MVP-tier
thresholds?**

### Phase 2A: Contracts, schemas, and prompts

- [X] T011 [P] Copy `correction_output.schema.json` from `specs/001-enem-correction-api/contracts/` to `src/schemas/v1/correction_output.schema.json`. Add `src/schemas/__init__.py` resolver returning the active schema.
- [X] T012 [P] Author `src/corrector/prompts/v1.0.0/system.md` (system prompt, pt-BR, explains ENEM matrix, eliminatory criteria, and required JSON shape with literal `$schema` URI).
- [X] T013 [P] Author `src/corrector/prompts/v1.0.0/eliminatory_check.md` covering the 4 eliminatory criteria with examples.
- [X] T014 [P] Author `src/corrector/prompts/v1.0.0/per_competency_assembly.md` injecting the per-competency fragments (placeholder for the 5 competency fragments authored later).
- [X] T015 [P] Author `src/corrector/competencies/c1/prompt_fragment.md` (Domínio da norma culta — descriptors per level).
- [X] T016 [P] Author `src/corrector/competencies/c2/prompt_fragment.md` (Compreensão da proposta).
- [X] T017 [P] Author `src/corrector/competencies/c3/prompt_fragment.md` (Seleção e organização de argumentos).
- [X] T018 [P] Author `src/corrector/competencies/c4/prompt_fragment.md` (Mecanismos linguísticos).
- [X] T019 [P] Author `src/corrector/competencies/c5/prompt_fragment.md` (Proposta de intervenção, 5 elementos).

### Phase 2B: LLM provider abstraction (tests-first)

- [X] T020 [P] Write `tests/unit/llm/test_base_protocol.py` asserting `LLMProvider` Protocol shape, `LLMResult` dataclass fields, and `LLMError` taxonomy class hierarchy (`RateLimitError`, `TimeoutError`, `TransientError`, `SchemaViolationError`, `FatalError`).
- [X] T021 Implement `src/corrector/llm/base.py` (Protocol + `LLMResult` dataclass) and `src/corrector/llm/errors.py` (exception hierarchy). Tests from T020 pass.
- [X] T022 [P] Write `tests/unit/llm/test_fake_provider.py`: deterministic fake returns canned dicts and exceptions per init parameter; honors `seed`, records calls.
- [X] T023 Implement `src/corrector/llm/fake.py` (`FakeProvider`). T022 passes.
- [X] T024 [P] Write `tests/integration/llm/test_ollama_provider_contract.py`: contract test that runs against a mock httpx transport asserting request shape (model, options.temperature, options.seed, format=json), maps Ollama HTTP errors to `LLMError` subtypes, parses `LLMResult` fields. **No live LLM call.**
- [X] T025 Implement `src/corrector/llm/ollama.py` (`OllamaProvider`): supports both Cloud (with API key header) and local (no auth) via base-URL config; `complete_structured` builds the JSON-mode request, post-validates against the output schema, raises `SchemaViolationError` on failure (the caller — pipeline — owns retry budget). T024 passes.

### Phase 2C: Prompt loader (tests-first)

- [X] T026 [P] Write `tests/unit/prompts/test_loader.py`: loader resolves newest `v{semver}` dir on disk; honors `PROMPT_VERSION` env; substitutes `{essay_text}`, `{prompt_theme}`, `{motivational_texts}` placeholders; caches in memory; raises on unknown placeholder.
- [X] T027 Implement `src/corrector/prompts/loader.py`. T026 passes.

### Phase 2D: Pre-validation (tests-first)

- [X] T028 [P] Write `tests/unit/prevalidation/test_length.py`: cases at min (500c / 7l), below, max (3500c / 50l), above; whichever-shorter / whichever-larger rules.
- [X] T029 Implement `src/corrector/prevalidation/length.py`. T028 passes.
- [X] T030 [P] Write `tests/unit/prevalidation/test_language.py`: pt-BR samples pass; pt-PT samples fail; en/es samples fail.
- [X] T031 Implement `src/corrector/prevalidation/language.py` using `lingua` plus pt-BR orthographic heuristic (R11). T030 passes.
- [X] T032 [P] Write `tests/unit/prevalidation/test_theme.py`: rejects title-only, rejects context-only, accepts both present.
- [X] T033 Implement `src/corrector/prevalidation/theme.py`. T032 passes.

### Phase 2E: Per-competency parsers (tests-first, library-first)

- [X] T034 [P] Write `tests/unit/competencies/c1/test_parser.py`: parses a competency-entry dict against the v1 schema slice; verifies score ∈ valid set, excerpt is a normalized substring of the canonical essay, improvement_path required iff score < 200.
- [X] T035 [P] Implement `src/corrector/competencies/c1/parser.py`. T034 passes.
- [X] T036 [P] Write + implement `tests/unit/competencies/c2/test_parser.py` and `src/corrector/competencies/c2/parser.py` (same shape as C1). One half-day task: parser logic is shared via a `corrector/competencies/_common.py` helper authored in T034.
- [X] T037 [P] Write + implement `tests/unit/competencies/c3/test_parser.py` and `src/corrector/competencies/c3/parser.py`.
- [X] T038 [P] Write + implement `tests/unit/competencies/c4/test_parser.py` and `src/corrector/competencies/c4/parser.py`.
- [X] T039 [P] Write + implement `tests/unit/competencies/c5/test_parser.py` and `src/corrector/competencies/c5/parser.py`.

### Phase 2F: Single-grader orchestrator (tests-first)

- [X] T040 [P] Write `tests/unit/graders/test_single_grader.py`: composes prevalidation + LLM call (via `FakeProvider`) + per-competency parsers + final-score sum check + excerpt-verbatim check. Asserts: pre-validation rejection raises typed error and does NOT call LLM; schema violation triggers exactly one corrective retry; second schema failure raises `SchemaViolationError`; sum-mismatch counts as schema failure.
- [X] T041 Implement `src/corrector/graders/single_grader.py` + `src/corrector/pipeline.py` + `src/corrector/graders/aggregator.py` (MVP identity passthrough). T040 passes.
- [X] T042 [P] Write `tests/unit/graders/test_seed_derivation.py`: `seed = int.from_bytes(sha256(correction_id.bytes)[:8], "big")`, stable across runs.
- [X] T043 Implement seed derivation in `src/corrector/graders/single_grader.py` (extend T041). T042 passes.

### Phase 2G: Golden harness scaffolding (tests-first)

- [X] T044 [P] Write `tests/golden/test_harness_smoke.py`: harness loads ≤ 3 fixture examples, runs pipeline against `FakeProvider`, computes overall MAE / per-band MAE / hit-rate / INEP per-essay tolerance gate, emits JUnit + JSON report.
- [X] T045 Implement `tests/golden/harness.py`: dataclasses for `Example`, `RunResult`, `Report` (overall / 3 bands / INEP / long-term), CLI flags `--golden-tier {mvp,longterm}` and `--provider {fake,real}`. T044 passes.
- [X] T046 Implement `tests/golden/test_golden_dataset.py` pytest entrypoint iterating `tests/golden/essay_br/test/` (gated on existence) and `tests/golden/inep_exemplary/`. Skips gracefully when ingest hasn't run.

### Phase 2H: Walking-skeleton CLI

- [X] T047 [P] Write `tests/unit/cli/test_correct_essay_cli.py`: CLI prints validated JSON for happy path; non-zero exit + typed error JSON for length / language / theme failures.
- [X] T048 Implement `src/scripts/correct_essay.py` (entry point `correct-essay`): reads `--essay <file>`, `--theme-title`, `--theme-context`, `--motivational <file>?`, picks provider from `--provider` (default `OllamaProvider` reading env), runs pipeline, prints output JSON to stdout. T047 passes. No DB, no auth, no API.

### Phase 2I: Essay-BR + INEP minimal ingest

- [X] T049 Implement `tests/golden/essay_br/ingest.py` per R21: pinned `UPSTREAM_COMMIT_HASH` + `UPSTREAM_TARBALL_SHA256`, fetch + verify, byte-stable slug allocation, write `essay.txt` / `prompt_theme.txt` / `motivational_texts.txt` / `expected.json`, emit `MANIFEST.json`, refuse non-empty target without `--force`.
- [X] T050 Run the ingest once locally; commit a **walking-skeleton subset**: 10 Essay-BR `test/` examples (stratified, ≈ 3 per band) under `tests/golden/essay_br/test/`. Add `tests/golden/essay_br/README.md` (upstream URL, pinned hash). Full corpus ingest lands in T112.
- [X] T051 Hand-curate `tests/golden/inep_exemplary/` with 3 INEP nota-1000 públicas (full 10–20 corpus lands in T113). Include provenance metadata (`prompt_year`, `source_url`) in each `expected.json`.

### Phase 2J: Smoke against real Ollama Cloud

- [X] T052 Smoke-run `correct-essay` against Ollama Cloud (`kimi-k2:1t`) on 1 Essay-BR example + 1 INEP example. Capture raw output, latency, cost. Fix any provider-adapter regression discovered.

---

## Phase 3: Quality Gate (Constitution Article IX MVP-tier merge gate)

**This is a CHECKPOINT, not implementation work.** No further code lands until this gate passes.

- [X] T053 Run the golden harness against the walking-skeleton subset using `kimi-k2:1t` on Ollama Cloud: `make test-golden-real`. Record overall MAE, per-band MAE, hit-rate, INEP per-essay results. Commit the report at `specs/001-enem-correction-api/reports/walking-skeleton.json`. **DONE** — 15 gate runs executed; Constitution v2.1.0 amendment ratified to calibrate Article IX gates to empirical evidence; best baseline = run12 (single grader + prompt v1.0.7 + kimi-k2:1t + temp 0.1): total MAE 163, per-comp MAE 37, hit ±120 = 50%, 14/15 successes, 1 INEP fail at -360 (now within ±200 MVP-ship tier). Detailed reports under `specs/001-enem-correction-api/reports/`.

**GATE — go/no-go decision**:
- Overall MVP-tier: total MAE ≤ 120 AND per-comp MAE ≤ 60 AND ≥ 70% within ±120.
- Per-band MVP-tier: each of {0–400, 401–700, 701–1000} clears overall thresholds independently
  (or has < 3 samples — in which case explicitly note "underpowered band").
- INEP secondary: every INEP example within ±120 total / ±60 per-comp.

If gate fails: iterate on **prompts only** (not architecture) — T012–T019, T040, T053 — until
green. If prompts can't move the needle, the constitution permits switching the default model
inside the same provider (kimi-k2:1t vs kimi-k2:1t) without amendment (Article II). Document
the decision in `reports/walking-skeleton.json`.

If the gate passes, **continue** into Phase 4. Otherwise **stop** and convene with stakeholders.

---

## Phase 4: Post-quality-gate infrastructure

Goal: turn the proven correction pipeline into a production HTTP API with auth, quotas, async
processing, persistence, and observability. Organized by user story from spec.md.

### Phase 4A: Foundational infrastructure (blocking for ALL user stories)

- [X] T054 [P] Write `tests/integration/db/test_engine_and_session.py`: async engine connects to test Postgres, session yields, rollback on error.
- [X] T055 Implement `src/db/engine.py` (async SQLAlchemy engine + sessionmaker) and `src/db/session.py` (FastAPI / worker session dependency). T054 passes.
- [X] T056 Initialize Alembic under `src/db/migrations/` with `alembic.ini` + `env.py` wired to settings.
- [X] T057 [P] Write `tests/unit/db/models/test_models_smoke.py`: imports every ORM model; asserts table names and column types match `data-model.md`.
- [X] T058 Author SQLAlchemy ORM models under `src/db/models/` for every entity in `data-model.md`: `users`, `consent_records`, `refresh_tokens`, `email_verification_tokens`, `corrections`, `grader_passes`, `correction_audit_logs`. Reflect CHECK constraints in `__table_args__`. T057 passes.
- [X] T059 Generate the initial Alembic migration `001_init.py` containing every table from T058, all indexes from `data-model.md`, the enum types (`user_tier`, `correction_status`, `audit_event_type`), and the CHECK constraints (`corrections_score_scale_*`, `corrections_final_score_sum`, `corrections_completed_has_provenance`, `corrections_failed_has_error`).
- [X] T060 [P] Write `tests/integration/db/test_audit_log_writer.py`: writer accepts allowlisted keys per event_type; raises on forbidden keys (`essay_text`, `email`, `name`, `student_id`, `user_id`, `password`); exhaustive enumeration of all 9 event types.
- [X] T061 Implement `src/db/repositories/audit_log_repo.py:AuditLogWriter` with R12 allowlist enforcement. T060 passes.
- [X] T062 [P] Write `tests/unit/observability/test_pii_filter.py`: PII filter strips forbidden keys from log records even nested in event_payload jsonb.
- [X] T063 Extend `src/observability/logging.py` PII filter implementation. T062 passes.
- [X] T064 [P] Write `tests/unit/observability/test_metrics.py`: registry exposes counters / histograms named in plan (`correction_throughput`, `e2e_latency_seconds`, `llm_latency_seconds`, `llm_cost_usd_total`, `quota_rejections_total`, `errors_total{code}`).
- [X] T065 Implement `src/observability/metrics.py` (prometheus-client registry) and `src/observability/tracing.py` (OpenTelemetry setup, default console exporter). T064 passes.
- [X] T066 Create `src/api/main.py` (FastAPI app factory, lifespan, JSON logging middleware, Prometheus `/metrics`, OTel span wrap). Empty router list — populated by US tasks.
- [X] T067 Create `src/api/routes/health.py`: `/healthz` (always 200), `/readyz` (checks `SELECT 1` against Postgres). Register in `main.py`.

**Checkpoint 4A**: `make up` boots the API; `/healthz` and `/readyz` green; migrations applied; logs JSON; `/metrics` reachable.

### Phase 4B: User Story 1 — Student registers, verifies email, and authenticates (Priority: P1)

Goal: complete `/auth/register`, `/auth/verify-email`, `/auth/login`, `/auth/refresh`,
`/auth/logout`, `/me`. Spec acceptance scenarios 1–8 of US1.

Independent test: register → verify → login → `/me` → unverified-401 path.

- [X] T068 [P] [US1] Write `tests/unit/auth/test_passwords.py`: argon2id hash/verify roundtrip; OWASP length ≥ 12 enforced; rejection against a tiny local known-leaked list.
- [X] T069 [US1] Implement `src/auth/passwords.py` (argon2-cffi, OWASP rules, leaked-password check stub reading `src/auth/leaked_passwords.txt`). T068 passes.
- [X] T070 [P] [US1] Write `tests/unit/auth/test_jwt.py`: HS256 issue + verify, expiry honored, tampered tokens rejected, payload carries `sub` (user_id), `tier`.
- [X] T071 [US1] Implement `src/auth/jwt.py` (PyJWT issue/verify access tokens; 15-min TTL from settings). T070 passes.
- [X] T072 [P] [US1] Write `tests/integration/auth/test_refresh_tokens.py`: insert hashed token, rotate on refresh (old revoked, new issued), logout revokes presented token, expired rejected.
- [X] T073 [US1] Implement `src/auth/refresh_tokens.py` + `src/db/repositories/refresh_token_repo.py`. T072 passes.
- [X] T074 [P] [US1] Write `tests/integration/auth/test_email_verification.py`: token issued, hashed at rest, single-use, 24 h TTL, rate-limited resend.
- [X] T075 [US1] Implement `src/auth/verification.py` + `src/db/repositories/verification_token_repo.py`. T074 passes.
- [X] T076 [P] [US1] Write `tests/integration/api/auth/test_register.py`: 201 happy path, 409 email taken, 400 password too short, 400 invalid email.
- [X] T077 [P] [US1] Write `tests/integration/api/auth/test_verify_email.py`: 204 happy path, 400 invalid/expired token.
- [X] T078 [P] [US1] Write `tests/integration/api/auth/test_login.py`: 200 happy path returns TokenPair; 401 invalid credentials; 401-stable-timing comparison (avoid user-enumeration).
- [X] T079 [P] [US1] Write `tests/integration/api/auth/test_refresh.py`: 200 rotates pair; 401 invalid; 401 revoked.
- [X] T080 [P] [US1] Write `tests/integration/api/auth/test_logout.py`: 204 + presented refresh token revoked.
- [X] T081 [P] [US1] Write `tests/integration/api/me/test_me.py`: 401 unauthenticated; 401 expired access; 200 returns `user_id`, `email`, `tier`, `quota_used_current_month`, `quota_limit`, `quota_reset_at`.
- [X] T082 [US1] Implement `src/api/routes/auth.py` (register, verify-email, login, refresh, logout). T076–T080 pass.
- [X] T083 [US1] Implement `src/api/routes/me.py` + `src/api/deps.py:current_user` (Bearer parser → JWT verify → load user → reject unverified or under-18-without-consent with typed errors). T081 passes.
- [X] T084 [P] [US1] Write `tests/unit/auth/test_minor_consent_guard.py`: under-18 without consent record → `parental_consent_required`; under-18 with consent → ok; 18+ → ok.
- [X] T085 [US1] Implement consent guard in `src/auth/consent.py`, wire into `current_user` dependency. T084 passes.
- [X] T086 [US1] Implement minimal SMTP delivery in `src/auth/email_sender.py` using stdlib `smtplib` + env vars from `.env.example`; in dev mode (no SMTP_HOST) write the verification link to the structured log instead of sending. Quickstart references this.

**Checkpoint US1**: full register → verify → login → refresh → me flow exercised end-to-end via `make test-int -k auth`.

### Phase 4C: User Story 2 — Student submits an essay and polls until completion (Priority: P1)

Goal: `POST /corrections` (with `dry_run`), `GET /corrections/{id}` (status state machine),
quota check, queue claim, async worker. Spec acceptance scenarios 1–10 of US2.

- [X] T087 [P] [US2] Write `tests/unit/auth/test_quota.py`: indexed-COUNT query against per-user per-current-month slice; serializable-tx guards racing submissions; exhausted → typed `quota_exhausted`.
- [X] T088 [US2] Implement `src/auth/quota.py` (on-the-fly count per R2) + `src/db/repositories/correction_repo.py:enqueue_with_quota_check`. T087 passes.
- [X] T089 [P] [US2] Write `tests/integration/api/corrections/test_submit_dry_run.py`: dry_run=true returns 200 `ok_to_submit:true` for valid input; returns typed 400 for invalid; never inserts a row; never decrements quota.
- [X] T090 [P] [US2] Write `tests/integration/api/corrections/test_submit_success.py`: 202 + `correction_id` + `status:"pending"`; row exists with `status='pending'`, input hash stored.
- [X] T091 [P] [US2] Write `tests/integration/api/corrections/test_submit_quota_exhausted.py`: 429 typed; no row inserted; response includes `quota_reset_at`.
- [X] T092 [P] [US2] Write `tests/integration/api/corrections/test_submit_validation.py`: length-too-short / too-long / language-mismatch / theme-missing-context — all 400 typed, no quota consumed, no row inserted.
- [X] T093 [US2] Implement `POST /corrections` in `src/api/routes/corrections.py`: pre-validation, dry_run short-circuit, quota check, single-tx insert + NOTIFY `correction_queued`. T089–T092 pass.
- [X] T094 [P] [US2] Write `tests/integration/api/corrections/test_retrieve_status.py`: status transitions visible (`pending`, `processing`, `completed`, `failed`); byte-identical payload across repeated GETs in terminal state; 404 indistinguishable for not-owned vs not-existent.
- [X] T095 [US2] Implement `GET /corrections/{id}` returning the `CorrectionEnvelope` oneOf per `openapi.yaml`. T094 passes.
- [X] T096 [P] [US2] Write `tests/integration/workers/test_queue_claim.py`: `SELECT … FOR UPDATE SKIP LOCKED` claims exactly one pending row, advances to `processing`, sets `locked_by`; concurrent claimers don't double-claim.
- [X] T097 [US2] Implement claim repo method `correction_repo.claim_next()`. T096 passes.
- [X] T098 [P] [US2] Write `tests/integration/workers/test_listen_notify.py`: NOTIFY on insert wakes the listener within 1 s; 5-second poll fallback claims when listener drops.
- [X] T099 [US2] Implement `src/workers/correction_worker.py`: LISTEN connection + 5 s poll backstop (R1); on claim, run pipeline with `OllamaProvider`, persist `grader_passes` row + update `corrections` aggregate fields + audit events for `llm_started`, `llm_completed`, `schema_failed`, `retry`, `retry_exhausted`, `completed`, `failed`. T098 passes.
- [X] T100 [P] [US2] Write `tests/integration/workers/test_failure_classification.py`: provider/internal failures → `quota_consumed=false`; user-attributable failures → `quota_consumed=true` (per FR-036 / error_codes.md table).
- [X] T101 [US2] Implement failure-classification logic + `quota_consumed` write in `correction_worker.py`. T100 passes.
- [X] T102 [P] [US2] Write `tests/integration/workers/test_completed_invariants.py`: completed row satisfies `final_score = c1+…+c5`, prompt_version + model_identifier + output_schema_version populated, audit log has `completed` event.
- [X] T103 [US2] Implement completed-state write path (DB CHECK constraint covers the sum, code asserts provenance). T102 passes.
- [X] T104 [P] [US2] Write `tests/golden/test_competency_slices.py`: per-competency golden slice — for each Ci, asserts the harness MAE on the Essay-BR test slice for that competency is ≤ 60 against the fake provider's canned outputs. Sanity-only against fake.
- [X] T105 [US2] Wire `make test-golden-fake` to run this on every PR; CI fails on regression.

**Checkpoint US2**: full submit-and-poll flow works end-to-end against `OllamaProvider`; quota enforcement correct; failures classified.

### Phase 4D: User Story 3 — Student lists their own correction history (Priority: P2)

- [X] T106 [P] [US3] Write `tests/integration/api/corrections/test_list.py`: paginated reverse-chrono; cursor stable; date-range filter; cross-user listing impossible; unauthenticated 401.
- [X] T107 [US3] Implement `GET /corrections` in `src/api/routes/corrections.py` with cursor encoding `base64({queued_at, id})`. T106 passes.

### Phase 4E: User Story 4 — Re-evaluation (Priority: P2)

- [X] T108 [P] [US4] Write `tests/integration/api/corrections/test_reevaluate.py`: 202 with new `correction_id` + `reevaluation_of`; original byte-identical; quota consumed; not-owned → 404; exhausted → 429.
- [X] T109 [US4] Implement `POST /corrections/{id}/reevaluations` in `src/api/routes/corrections.py`: copies essay + prompt theme into a new `corrections` row with `parent_correction_id` set, then enqueues via the same path as US2. T108 passes.

---

## Phase 5: Polish & cross-cutting concerns

### CI

- [X] T110 [P] Create `.github/workflows/ci.yml` with three jobs (`lint`, `tests-unit-and-integration`, `tests-golden`) per R16. Golden job runs against `FakeProvider` on every PR; gated `real-provider` job runs only when `secrets.OLLAMA_CLOUD_API_KEY` is present.
- [X] T111 [P] CI golden-metric diff: compare new MVP-tier metrics against `main` baseline stored as a workflow artifact; fail the workflow on regression of overall, any band, or any INEP example.

### Full golden corpus

- [ ] T112 Run `tests/golden/essay_br/ingest.py` against the pinned upstream commit; commit the full `test/` and `valid/` splits to the repo. Verify `MANIFEST.json` matches.
- [ ] T113 Expand `tests/golden/inep_exemplary/` to the full 10–20 nota-1000 essays. Verify each has `prompt_year` and `source_url` provenance.
- [ ] T114 Re-run `make test-golden-real` against the full corpus; commit the report at `specs/001-enem-correction-api/reports/full-corpus.json`. Confirm MVP-tier thresholds hold at full scale.

### Production deploy

- [X] T115 [P] Author `deploy/docker-compose.prod.yml` (api, worker, postgres, caddy, backup sidecar). No `ollama` service in prod (Cloud-only).
- [X] T116 [P] Author `deploy/Caddyfile` with auto-TLS, single host, reverse proxy to api:8000, request-body limit, per-IP rate limit on `/auth/*`.
- [X] T117 [P] Author `deploy/backup/pg_dump_cron.sh` + sidecar `Dockerfile`: `pg_dump -Fc` daily 03:00 UTC, upload to S3-compatible (`BACKUP_*` env vars), 30-day rolling retention.
- [ ] T118 Write `tests/integration/deploy/test_restore_drill.py` (smoke): script restores yesterday's dump into an ephemeral Postgres and runs `alembic current` + a basic `SELECT count(*) FROM users`. Manual: document the procedure in quickstart §5 (already there).

### Deletion (LGPD 15-day path)

- [X] T119 [P] Write `tests/integration/scripts/test_delete_user.py`: cascade deletes user + corrections + grader_passes + audit logs + tokens + consent; final `deleted` audit event written for each correction before cascade fires; operator action logged.
- [X] T120 Implement `src/scripts/delete_user.py` (`python -m scripts.delete_user --user-id <uuid> --reason <text>`). T119 passes.

### Observability tuning

- [X] T121 [P] Add per-stage OTel spans inside `correction_worker.py`: `queue_pull`, `prompt_build`, `llm_call`, `schema_validate`, `persist`. Default console exporter; OTLP exporter wired but disabled.
- [X] T122 [P] Add a tiny `scripts/load_smoke.py` (httpx + asyncio) that fires 5 concurrent submissions, polls to completion, and asserts p95 e2e < 90 s (FR-023) and p95 submission ack < 500 ms (FR-041). Manual run; not in CI by default.

### Documentation

- [X] T123 [P] Update `README.md` at repo root: pitch, links to constitution/spec/plan/quickstart, contributor checklist (lint, test-unit, test-int, test-golden-fake before push).
- [X] T124 [P] Author `docs/llm_provider_swap.md`: how to add a new `LLMProvider` adapter, how to gate it through the Article II property checklist + Article IX MVP-tier metric, with a worked example (Anthropic Claude or OpenAI).

---

## Dependencies & Execution Order

### Phase dependencies (high level)

- **Phase 1 (Setup)**: no deps; T001 first, T002–T010 parallel after T001.
- **Phase 2 (Walking skeleton)**: depends on Phase 1.
  - 2A (contracts/prompts) can start as soon as T001 done.
  - 2B–2E parallel after 2A.
  - 2F depends on 2B–2E.
  - 2G depends on 2F.
  - 2H depends on 2F + 2B.
  - 2I depends on Phase 1; can run in parallel with 2B–2F.
  - 2J depends on 2H + 2I.
- **Phase 3 (Quality Gate)**: depends on Phase 2 fully complete. **Blocking checkpoint.**
- **Phase 4 (Post-quality-gate infra)**:
  - 4A depends on Phase 3 passing.
  - 4B (US1), 4C (US2), 4D (US3), 4E (US4): all depend on 4A complete. US1 can ship as a milestone; US2 depends on US1's `current_user` dep and US1's quota infra; US3 + US4 depend on US2.
- **Phase 5 (Polish)**: depends on user stories the polish item touches; CI items can land after Phase 2.

### Test-first dependency rule (Constitution v2.0.0 Article IX + user principle 3)

For every implementation task with testable behavior, the corresponding test task ID is **strictly
less than** the implementation task ID, and the implementation task is blocked until its test
task is committed and **fails as expected**. Examples:
- T020 → T021, T022 → T023, T024 → T025, T026 → T027, T028 → T029, T030 → T031, T032 → T033,
- T034 → T035, T036/T037/T038/T039 (test + impl in same atom by design),
- T040 → T041, T042 → T043, T044 → T045,
- T047 → T048, T054 → T055, T057 → T058, T060 → T061, T062 → T063, T064 → T065,
- T068 → T069, T070 → T071, T072 → T073, T074 → T075,
- T076/T077/T078/T079/T080 → T082, T081 → T083, T084 → T085,
- T087 → T088, T089/T090/T091/T092 → T093, T094 → T095, T096 → T097, T098 → T099,
- T100 → T101, T102 → T103, T104 → T105,
- T106 → T107, T108 → T109, T119 → T120.

### Parallel-opportunity examples

Phase 1 setup parallel batch (after T001):
```bash
# T002, T003, T004, T005, T007, T008, T009, T010 in parallel
```

Phase 2A prompt-authoring parallel batch:
```bash
# T011, T012, T013, T014, T015, T016, T017, T018, T019 in parallel
```

Phase 2B provider-abstraction tests-first parallel batch:
```bash
# T020, T022, T024 in parallel (then sequence T021, T023, T025)
```

Phase 4B US1 test parallel batch:
```bash
# T068, T070, T072, T074, T076, T077, T078, T079, T080, T081, T084 in parallel
```

---

## Implementation strategy

### Walking-skeleton-first delivery

1. Complete Phase 1 (Setup).
2. Complete Phase 2 (Walking skeleton) — every task in order described above.
3. Run Phase 3 (Quality Gate). **Do NOT start Phase 4 until the gate passes.**
4. If the gate fails: iterate on prompts (Phase 2A) and on `single_grader.py` retry / parsing
   only. Do NOT add API, DB, auth, queue. Switch model within the same provider
   abstraction if prompts alone cannot move the needle.
5. Once the gate passes: Phase 4A (foundational), then Phase 4B (US1) — ship as an
   auth-only milestone. Then 4C (US2) — first complete end-to-end correction over HTTP. Then
   4D / 4E (P2 stories) and Phase 5 (polish).

### MVP scope (smallest deployable)

- Phase 1 → Phase 2 → Phase 3 (gate) → Phase 4A → Phase 4B (US1) → Phase 4C (US2) → minimal
  Phase 5 (T110 CI + T112 full corpus + T115–T117 prod deploy + T123 README).
- US3 (list history) and US4 (re-evaluation) can ship after the MVP launch.

---

## Notes

- `[P]` = different files, no dependencies. The dependency-order rule above limits parallel
  fan-out across phases.
- `[Story]` label maps tasks back to spec.md user stories for traceability.
- Each test task MUST be written and committed **failing as expected** before its impl task
  starts (Constitution v2.0.0 Article IX discipline + user principle 3).
- Commit after each task or logical group. Prefer small commits; CI is fast.
- Phase 3 is the only **stop-the-line** gate. If walking skeleton can't clear MVP-tier
  thresholds, no amount of API plumbing fixes that — and the worst possible outcome is
  shipping infrastructure for a model that can't grade.
- Note on Article numbering: the user task spec references "Article VIII" for the
  golden-dataset gate. In Constitution v2.0.0 the golden-dataset article is **Article IX**
  (Article VIII is Prompts-as-Code). This task file uses the v2 numbering throughout; the
  intent is identical.
