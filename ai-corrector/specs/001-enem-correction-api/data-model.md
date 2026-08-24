# Phase 1 Data Model: ENEM Essay Correction API (MVP)

**Spec**: [spec.md](./spec.md) v2.0.0
**Plan**: [plan.md](./plan.md)
**Research**: [research.md](./research.md)
**Constitution**: v2.0.0
**Date**: 2026-05-28

All tables live in PostgreSQL 16+, owned by Alembic migrations under `src/db/migrations/`.
SQLAlchemy 2.x async ORM. Identifiers are UUIDv7 (time-ordered) generated app-side; this lets
the queue's `ORDER BY id` correlate with `queued_at` cheaply.

Convention: timestamps are `TIMESTAMPTZ`, stored UTC. JSON columns are `JSONB`. Soft delete is
not used; deletion is hard (Constitution X) with audit-trail purge handled at the AuditLogWriter
boundary.

---

## Entity 1 — `users`

```sql
CREATE TYPE user_tier AS ENUM ('free', 'premium');

CREATE TABLE users (
    id                  UUID PRIMARY KEY,
    email               CITEXT NOT NULL UNIQUE,
    password_hash       TEXT NOT NULL,             -- Argon2id encoded hash
    tier                user_tier NOT NULL DEFAULT 'free',
    email_verified_at   TIMESTAMPTZ,                -- NULL when unverified
    birth_date          DATE,                       -- nullable; required only when minor
    consent_record_id   UUID REFERENCES consent_records(id),  -- required when birth_date < 18y old
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX users_email_idx ON users (email);     -- redundant with UNIQUE; explicit for clarity
```

**Notes**:
- `email_verified_at IS NULL` ⇒ account is in `unverified` state; correction endpoints reject.
- Consent FK is enforced at the application layer when `birth_date` puts the user under 18; we
  do not enforce it as a DB constraint because the rule depends on "today".
- `email` uses `CITEXT` (case-insensitive) — no case-fold bugs in login.

---

## Entity 2 — `consent_records`

```sql
CREATE TABLE consent_records (
    id                  UUID PRIMARY KEY,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    guardian_name       TEXT NOT NULL,
    guardian_relation   TEXT NOT NULL,             -- e.g. "mãe", "pai", "responsável legal"
    consent_text_sha256 BYTEA NOT NULL,            -- hash of the consent statement shown
    accepted_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at          TIMESTAMPTZ
);

CREATE INDEX consent_records_user_idx ON consent_records (user_id) WHERE revoked_at IS NULL;
```

**Notes**:
- Parental consent record per Constitution X. Soft revocation is supported here (audit trail)
  rather than hard delete, because consent withdrawal MUST persist as a fact while the
  underlying user data is purged on deletion.

---

## Entity 3 — `refresh_tokens`

```sql
CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      BYTEA NOT NULL UNIQUE,         -- sha256 of the cleartext token
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX refresh_tokens_user_active_idx
    ON refresh_tokens (user_id)
    WHERE revoked_at IS NULL;
```

**Notes**:
- 30-day TTL set at insert.
- Rotation: refresh-endpoint revokes presented token, inserts a new one.
- Logout: revoke presented token.

---

## Entity 4 — `email_verification_tokens`

```sql
CREATE TABLE email_verification_tokens (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      BYTEA NOT NULL UNIQUE,
    expires_at      TIMESTAMPTZ NOT NULL,           -- 24h after creation
    used_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX email_verification_tokens_user_idx
    ON email_verification_tokens (user_id)
    WHERE used_at IS NULL;
```

---

## Entity 5 — `corrections` (aggregate + queue row)

```sql
CREATE TYPE correction_status AS ENUM ('pending', 'processing', 'completed', 'failed');

CREATE TABLE corrections (
    id                      UUID PRIMARY KEY,        -- UUIDv7
    user_id                 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- input
    essay_text              TEXT NOT NULL,
    prompt_theme_title      TEXT NOT NULL,
    prompt_theme_context    TEXT NOT NULL,
    motivational_texts      TEXT,
    input_hash              BYTEA NOT NULL,           -- sha256 of canonicalized input

    -- status
    status                  correction_status NOT NULL DEFAULT 'pending',
    queued_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at              TIMESTAMPTZ,
    completed_at            TIMESTAMPTZ,
    locked_at               TIMESTAMPTZ,              -- worker claim timestamp
    locked_by               TEXT,                     -- worker instance id

    -- aggregate result (copied from sole grader_pass in MVP)
    final_score             SMALLINT,
    c1_score                SMALLINT,
    c2_score                SMALLINT,
    c3_score                SMALLINT,
    c4_score                SMALLINT,
    c5_score                SMALLINT,
    competencies            JSONB,                    -- full per-competency object (excerpt, justification, improvement_path)
    eliminatory_flags       JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- provenance
    prompt_version          TEXT,                     -- SemVer, NULL until completed
    model_identifier        TEXT,                     -- e.g. "kimi-k2:1t"
    output_schema_version   TEXT,                     -- e.g. "v1"

    -- failure
    error_code              TEXT,                     -- typed: provider_rate_limited, ...
    error_message_pt_br     TEXT,
    quota_consumed          BOOLEAN NOT NULL DEFAULT TRUE,

    -- lineage
    parent_correction_id    UUID REFERENCES corrections(id),  -- for re-evaluations

    CONSTRAINT corrections_score_scale CHECK (
        c1_score IS NULL OR c1_score IN (0,40,80,120,160,200)
    ),
    CONSTRAINT corrections_c2_score_scale CHECK (
        c2_score IS NULL OR c2_score IN (0,40,80,120,160,200)
    ),
    CONSTRAINT corrections_c3_score_scale CHECK (
        c3_score IS NULL OR c3_score IN (0,40,80,120,160,200)
    ),
    CONSTRAINT corrections_c4_score_scale CHECK (
        c4_score IS NULL OR c4_score IN (0,40,80,120,160,200)
    ),
    CONSTRAINT corrections_c5_score_scale CHECK (
        c5_score IS NULL OR c5_score IN (0,40,80,120,160,200)
    ),
    CONSTRAINT corrections_final_score_range CHECK (
        final_score IS NULL OR (final_score BETWEEN 0 AND 1000)
    ),
    CONSTRAINT corrections_final_score_sum CHECK (
        status <> 'completed' OR final_score = c1_score + c2_score + c3_score + c4_score + c5_score
    ),
    CONSTRAINT corrections_completed_has_provenance CHECK (
        status <> 'completed' OR (prompt_version IS NOT NULL AND model_identifier IS NOT NULL AND output_schema_version IS NOT NULL)
    ),
    CONSTRAINT corrections_failed_has_error CHECK (
        status <> 'failed' OR error_code IS NOT NULL
    )
);

-- queue claim index (worker SELECT ... FOR UPDATE SKIP LOCKED)
CREATE INDEX corrections_queue_idx
    ON corrections (queued_at)
    WHERE status = 'pending';

-- quota-counter index (per-user, per-month COUNT)
CREATE INDEX corrections_user_queued_idx
    ON corrections (user_id, queued_at);

-- listing endpoint (per-user, reverse chronological)
CREATE INDEX corrections_user_listing_idx
    ON corrections (user_id, queued_at DESC);

-- re-evaluation lineage walk
CREATE INDEX corrections_parent_idx
    ON corrections (parent_correction_id)
    WHERE parent_correction_id IS NOT NULL;
```

**State machine**:

```
        register submission (FR-038, quota ok)
                    │
                    ▼
                 pending ────► failed (typed error; quota_consumed = false)
                    │             ▲
                    │             │ pre-claim error (rare; worker crash before update)
                    ▼             │
              processing ─────────┼────► failed (provider/schema/internal; quota_consumed false)
                    │             │
                    │             └─► failed (user-attributable; quota_consumed true)
                    ▼
                completed
```

Only `pending → processing → {completed, failed}` and `pending → failed`. No backward
transitions. No `processing → pending`. Terminal states are immutable (FR-009, FR-039).

---

## Entity 6 — `grader_passes` (multi-grader-ready)

```sql
CREATE TABLE grader_passes (
    id                      UUID PRIMARY KEY,
    correction_id           UUID NOT NULL REFERENCES corrections(id) ON DELETE CASCADE,
    pass_index              SMALLINT NOT NULL,                -- 0 in MVP; 0..N in future
    c1_score                SMALLINT NOT NULL,
    c2_score                SMALLINT NOT NULL,
    c3_score                SMALLINT NOT NULL,
    c4_score                SMALLINT NOT NULL,
    c5_score                SMALLINT NOT NULL,
    competencies            JSONB NOT NULL,                   -- per-competency excerpt + justification + improvement_path
    eliminatory_flags       JSONB NOT NULL DEFAULT '[]'::jsonb,
    seed                    BIGINT,                           -- the deterministic seed used
    prompt_version          TEXT NOT NULL,                    -- SemVer
    model_identifier        TEXT NOT NULL,
    output_schema_version   TEXT NOT NULL,
    inference_params        JSONB NOT NULL,                   -- temperature, top_p, max_tokens, ...
    raw_output              TEXT NOT NULL,                    -- exact LLM reply (post-parse)
    prompt_tokens           INT,
    completion_tokens       INT,
    latency_ms              INT NOT NULL,
    cost_usd                NUMERIC(10, 6),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (correction_id, pass_index),
    CONSTRAINT grader_passes_c1_scale CHECK (c1_score IN (0,40,80,120,160,200)),
    CONSTRAINT grader_passes_c2_scale CHECK (c2_score IN (0,40,80,120,160,200)),
    CONSTRAINT grader_passes_c3_scale CHECK (c3_score IN (0,40,80,120,160,200)),
    CONSTRAINT grader_passes_c4_scale CHECK (c4_score IN (0,40,80,120,160,200)),
    CONSTRAINT grader_passes_c5_scale CHECK (c5_score IN (0,40,80,120,160,200))
);

CREATE INDEX grader_passes_correction_idx ON grader_passes (correction_id);
```

**MVP behavior**: exactly one `grader_passes` row per `corrections` row, `pass_index = 0`.
Aggregate columns on `corrections` are copies of this row.

**Future multi-grader (post-MVP)**: insert N rows (`pass_index = 0..N-1`), then run the ENEM
aggregation rule (average of the two closest passes; trigger 3rd pass when divergence > 100
total or > 80 on any competency) inside `corrector/graders/aggregator.py`, and write the
aggregate back to `corrections`. **No schema change required.**

---

## Entity 7 — `correction_audit_logs`

```sql
CREATE TYPE audit_event_type AS ENUM (
    'submitted',
    'llm_started',
    'llm_completed',
    'schema_failed',
    'retry',
    'retry_exhausted',
    'completed',
    'failed',
    'deleted'
);

CREATE TABLE correction_audit_logs (
    id              UUID PRIMARY KEY,
    correction_id   UUID NOT NULL REFERENCES corrections(id) ON DELETE CASCADE,
    input_hash      BYTEA NOT NULL,                  -- sha256 of canonicalized input; redundant w/ corrections.input_hash, kept for post-delete trace
    event_type      audit_event_type NOT NULL,
    event_payload   JSONB NOT NULL DEFAULT '{}'::jsonb,    -- allowlisted keys ONLY (enforced by AuditLogWriter)
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX correction_audit_logs_correction_idx
    ON correction_audit_logs (correction_id, created_at);
```

**Allowlist by event_type** (enforced by `db/repositories/audit_log_repo.py:AuditLogWriter`):

| event_type | allowed keys in `event_payload` |
|---|---|
| `submitted` | `text_length_chars`, `text_length_lines`, `language_detected` |
| `llm_started` | `prompt_version`, `model_identifier`, `seed`, `attempt` |
| `llm_completed` | `prompt_tokens`, `completion_tokens`, `latency_ms`, `cost_usd`, `attempt` |
| `schema_failed` | `attempt`, `validator_message` (truncated 500 chars; PII filter applied) |
| `retry` | `attempt`, `reason` |
| `retry_exhausted` | `final_error_code` |
| `completed` | `final_score`, `eliminatory_flags`, `prompt_version`, `model_identifier` |
| `failed` | `error_code`, `quota_consumed` |
| `deleted` | `deletion_reason` |

**Forbidden keys anywhere**: `essay_text`, `email`, `name`, `student_id`, `user_id`, `password`.
The writer raises at runtime; unit test enumerates every event type to lock the allowlist.

**On user deletion (Constitution X, 15-day window)**:
1. `DELETE FROM users WHERE id = $1` cascades to `corrections`, `grader_passes`,
   `correction_audit_logs`, `refresh_tokens`, `email_verification_tokens`,
   `consent_records`.
2. **Exception**: a single `audit_event_type = 'deleted'` row may be retained per correction in
   a separate `deletion_log` table outside this entity model (TBD post-MVP). The MVP cascades
   everything and relies on Postgres backups for forensic recovery if a dispute arises.

---

## Cross-cutting

### Deletion semantics

- User deletion deletes everything tied to the user (Constitution X).
- Correction deletion (future self-service, Open Question Q4) deletes essay text +
  justifications + raw_output + audit rows except the deletion event.

### Concurrency

- Quota enforcement on submission runs inside a single transaction:
  `BEGIN; SELECT count … FOR UPDATE NOWAIT (on users row); INSERT corrections; COMMIT;` to
  serialize racing quota-spending submissions per user. NOWAIT returns immediately so the API
  surface stays responsive.

### Queue claim

- Worker claim query:
  ```sql
  WITH claimed AS (
      SELECT id FROM corrections
      WHERE status = 'pending'
      ORDER BY queued_at
      FOR UPDATE SKIP LOCKED
      LIMIT 1
  )
  UPDATE corrections c
  SET status = 'processing', started_at = now(), locked_at = now(), locked_by = $worker_id
  FROM claimed
  WHERE c.id = claimed.id
  RETURNING c.*;
  ```
- LISTEN channel: `correction_queued` (NOTIFY emitted on submission insert via app code, not
  trigger — keeps DB schema simple).

### Indexes summary

| Table | Index | Purpose |
|---|---|---|
| `users` | UNIQUE on `email` (CITEXT) | login |
| `consent_records` | partial on `user_id` WHERE not revoked | active-consent lookup |
| `refresh_tokens` | UNIQUE on `token_hash`; partial on `user_id` WHERE active | token validation; logout-all-sessions |
| `email_verification_tokens` | UNIQUE on `token_hash`; partial on `user_id` WHERE unused | verification |
| `corrections` | partial on `queued_at` WHERE pending | worker queue claim |
| `corrections` | composite `(user_id, queued_at)` | quota count |
| `corrections` | composite `(user_id, queued_at DESC)` | listing |
| `corrections` | partial on `parent_correction_id` | re-evaluation lineage |
| `grader_passes` | UNIQUE `(correction_id, pass_index)` | multi-grader integrity |
| `grader_passes` | on `correction_id` | per-correction join |
| `correction_audit_logs` | composite `(correction_id, created_at)` | per-correction timeline |

---

## Mapping to spec entities

| Spec entity | Tables |
|---|---|
| User | `users` |
| Session / Tokens | `refresh_tokens` (+ stateless JWT access tokens) |
| Quota Ledger | derived from `corrections` (on-the-fly count, R2) |
| Essay Submission | `corrections.{essay_text, prompt_theme_*, motivational_texts, input_hash}` |
| Prompt Theme | `corrections.{prompt_theme_title, prompt_theme_context}` |
| Correction Job | `corrections` (+ `grader_passes` for the underlying pass(es)) |
| Competency Entry | `corrections.competencies` jsonb + `grader_passes.competencies` jsonb |
| Audit Record | `correction_audit_logs` |
| Consent Record | `consent_records` (linked from `users.consent_record_id`) |
