# Phase 1 Data Model: Constitution Alignment Refactor

Two ownership zones, one shared Postgres instance (per research.md §1):

- **Go monolith** owns `public`/domain schemas: identity, quota, streaks,
  friends, ranking, medals.
- **Correction service** owns schema `correction`: the grading pipeline
  only, reachable by the monolith through exactly one bridge table.

Timestamps are `TIMESTAMPTZ` stored UTC. IDs are UUIDs.

## Go monolith entities

### `users` (existing table, gains columns)

- `subscription_tier` — `'free' | 'pro'` (existing tier concept, renamed if
  needed to match constitution's Free/Pro language)
- `current_streak` — integer, ≥ 0
- `longest_streak` — integer, ≥ 0
- `last_submission_day` — `DATE` (UTC calendar day), used to detect streak
  breaks and same-day quota checks without a table scan

### `essay_submissions` (new — `essay` domain)

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `user_id` | UUID FK → `users.id` | |
| `prompt_theme_title` | TEXT | |
| `prompt_theme_context` | TEXT | |
| `essay_text` | TEXT | |
| `submission_day` | DATE (UTC) | quota key: one free row per (`user_id`, `submission_day`) unless Pro |
| `status` | ENUM `pending, graded, failed` | mirrors the bridged correction job |
| `correction_job_id` | UUID, nullable FK → `correction.correction_jobs.id` | set once enqueued |
| `submitted_at` | TIMESTAMPTZ | |

**Rule**: streak increment and quota consumption happen in the same
transaction that inserts this row — per research.md §4, *before* grading
completes.

**Validation**: a free-tier user cannot insert a second row with the same
(`user_id`, `submission_day`) — enforced by a partial unique index scoped to
free-tier submissions (checked against the user's tier at submission time).

### `essay_grades` (new — `essay` domain, denormalized read model)

Populated when the monolith reconciles a completed correction job back to
its submission (research.md §1). One row per graded `essay_submissions`.

| Column | Type | Notes |
|---|---|---|
| `submission_id` | UUID PK/FK → `essay_submissions.id` | |
| `overall_score` | SMALLINT, 0–1000 | |
| `competencies` | JSONB | array of 5: `{competency, score, justification_pt_br, excerpt}` — reuses `ai-corrector`'s existing structured shape |
| `graded_at` | TIMESTAMPTZ | |

### `friendships` (new — `social` domain)

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `requester_id` | UUID FK → `users.id` | |
| `addressee_id` | UUID FK → `users.id` | |
| `status` | ENUM `pending, accepted, removed` | |
| `created_at` / `responded_at` | TIMESTAMPTZ | |

**Visibility rule**: a user's `current_streak` and most recent
`essay_grades` row are only queryable by another user through a join gated
on an `accepted` friendship row existing between them (enforced in the
`social` domain's query layer, not left to the client).

### `weekly_ranking_entries` (new — `gamification` domain)

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `user_id` | UUID FK → `users.id` | |
| `week_start` | DATE (UTC, Monday) | |
| `weekly_score` | INTEGER | sum of that week's `essay_grades.overall_score` |
| `league_tier` | ENUM (e.g. `bronze, silver, gold, ...`) | |
| `rank_in_tier` | INTEGER, computed at week close | |

**State transition**: at week close, a scheduled job computes `rank_in_tier`
per `league_tier`, promotes the top band / demotes the bottom band into a
new `weekly_ranking_entries` row for the next `week_start` with an updated
`league_tier`, per the deterministic rule from research.md §4. Tie-break:
earlier `essay_grades.graded_at` among tied scores wins the higher slot
(spec Assumptions).

### `medals` (new — `gamification` domain)

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `user_id` | UUID FK → `users.id` | |
| `type` | ENUM (streak milestones, tier promotions, top finishes — enumerated at tasks stage) | |
| `earned_at` | TIMESTAMPTZ | |

## Correction service entities (schema `correction`)

### `correction_jobs` (new — the outbox/bridge table, replaces direct writes to `corrections`)

The **only** table the Go monolith is granted access to in this schema.

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `user_id` | UUID | opaque reference, no local FK (research.md §2) |
| `essay_text` | TEXT | |
| `prompt_theme_title` / `prompt_theme_context` | TEXT | |
| `status` | ENUM `pending, processing, completed, failed` | claimed via existing `SELECT ... FOR UPDATE SKIP LOCKED` worker pattern |
| `queued_at` / `started_at` / `completed_at` | TIMESTAMPTZ | |

Go: `INSERT` (enqueue) + `SELECT` (poll for `completed`/`failed` status and
read the result) only. Python: full read/write, including the `UPDATE`
that claims and completes a row.

### `corrections` (existing table, trimmed)

Kept as-is structurally (5-competency JSONB, score-scale CHECK constraints,
typed `error_code`/`error_message_pt_br`, `prompt_version` /
`model_identifier` / `output_schema_version` provenance — all reused
unchanged from `ai-corrector`'s existing schema) but:
- `user_id` FK to the local `users` table is **dropped** (table itself is
  dropped per research.md §2); column becomes an opaque UUID matching
  `correction_jobs.user_id`.
- Row is now created by the worker when it claims a `correction_jobs` row,
  not by an end-user-facing `POST /corrections`.
- `quota_consumed` column becomes informational only — actual quota
  enforcement has moved to the Go `essay` domain (Principle V).

### Dropped tables

`users`, `refresh_tokens`, `consent_records`, `email_verification_tokens` —
removed via a new Alembic migration; their responsibilities move to the Go
monolith's existing `auth`/`user` domains, which already implement
equivalent concepts (see `backend/app/internal/auth/`, `internal/user/`).

## Cross-system reconciliation flow

1. `essay` domain (Go) checks quota, opens a transaction, inserts
   `essay_submissions` (status `pending`), increments the user's streak,
   inserts into `correction_jobs` (schema `correction`), commits.
2. Correction worker (Python) claims the `correction_jobs` row via its
   existing LISTEN/NOTIFY + poll consumer, runs the grading pipeline,
   writes a `corrections` row, marks `correction_jobs.status = completed`
   (or `failed` with a typed `error_code`, reusing the existing taxonomy).
3. `essay` domain (Go) — via poll or a lightweight reconciliation worker —
   reads completed/failed `correction_jobs` rows, writes `essay_grades` (on
   success) or marks `essay_submissions.status = failed` (on failure,
   surfaced per spec Edge Cases: user can resubmit without losing the
   day's streak credit), and updates `weekly_ranking_entries` for the
   current week.
