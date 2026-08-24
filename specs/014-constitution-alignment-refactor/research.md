# Phase 0 Research: Constitution Alignment Refactor

All Technical Context items were resolved from repo inspection and prior
`/speckit.constitution` clarifications; no items were left as
NEEDS CLARIFICATION going into Phase 1. This document records the decisions
that required judgment.

## 1. Go monolith ↔ correction service transport

**Decision**: Shared Postgres instance, correction service owns a
`correction_jobs` outbox table (its own migration, its own schema
`correction`); the Go monolith is granted narrow `INSERT`/`SELECT` privileges
on that one table via a thin repository, and does not otherwise read the
`correction` schema. The correction service's existing worker (LISTEN/NOTIFY
+ 5s poll fallback, already built in `ai-corrector/src/workers/`) is
reused unchanged — it just claims rows from `correction_jobs` instead of
`corrections` directly, then writes results into its own `corrections`-style
result table, and the Go monolith polls/reads *that* result table (also
narrow grant) to reconcile back to the originating essay submission.

**Rationale**: The user's prior clarification selected "async via DB/queue"
over a synchronous HTTP call. A shared Postgres instance (already the only
Postgres the constitution allows as a stateful store) with schema-level
ownership gives durability and the existing LISTEN/NOTIFY consumer without
standing up a second broker (Kafka/SQS/etc.), and keeps `corretor/` isolated
from `db`/`api` per the service's own import-linter rule. It also avoids the
"own separate DB instance" reading of the prior clarification turning into
extra infra to operate, while still satisfying "own schema" from the
constitution's Architecture section.

**Alternatives considered**:
- *Synchronous internal HTTP call* — rejected per explicit prior decision
  (couples monolith request latency to LLM latency, violates Principle V's
  "must never block on LLM latency").
- *Separate Postgres instance for the correction service* — rejected for
  MVP: doubles ops surface (backups, connection pooling, migrations
  runners) for no benefit once schema-level isolation already satisfies the
  constitution; can be split later without an application-level rewrite
  since the boundary is already schema-clean.
- *Message broker (Redis Streams, since Redis is already provisioned)* —
  viable, rejected only because it throws away the correction service's
  already-working, already-tested LISTEN/NOTIFY implementation for no
  functional gain; revisit if correction volume ever needs true horizontal
  worker fan-out beyond what Postgres advisory-lock/poll claiming supports.

## 2. Identity and quota ownership (removing duplication)

**Decision**: `ai-corrector` loses its `users`, `refresh_tokens`,
`consent_records`, and `email_verification_tokens` tables and every endpoint
built on them (`/auth/*`, parts of `/me`). Its `corrections` table keeps a
`user_id` column but it becomes an **opaque UUID reference** with no local
foreign key — the Go monolith is the only source of truth for who that user
is. Quota enforcement (free 1/day, Pro multi/day) is fully removed from the
Python side and re-implemented in the new Go `essay` domain, which checks
quota *before* writing a `correction_jobs` row.

**Rationale**: Directly required by the constitution's Architecture section
("must not duplicate identity ... data owned by the monolith") and Principle
V ("quota enforcement at the boundary ... enforced by the Go monolith ...
not inside the correction pipeline"). `ai-corrector` was built as a
standalone B2C product with its own auth; that auth becomes dead weight once
it's an internal service behind the Go monolith.

**Alternatives considered**:
- *Keep Python auth as-is, have Go proxy through it* — rejected: doubles the
  auth surface area to maintain and directly violates the constitution.
- *Sync users table from Go to Python via replication* — rejected: exactly
  the "duplication" the constitution forbids, adds a consistency-lag failure
  mode for no requirement that needs it (the correction service never talks
  to end users directly).

## 3. Service-to-service authentication (Go → correction service)

**Decision**: Internal-network trust boundary — the correction service's
remaining endpoints (health, and any endpoint the worker/API still expose
internally) are reachable only from the Go monolith's network segment, using
a shared internal credential (e.g., a service-level bearer token or mTLS,
finalized at the tasks/implementation stage) instead of end-user JWTs.

**Rationale**: With end-user auth removed from Python, calls arriving at the
correction service are monolith-to-service, not user-to-service. NGINX
(the constitution's only external gateway) does not route directly to the
correction service.

**Alternatives considered**: End-user JWT forwarded and re-verified by
Python — rejected, since Python no longer has the users table to verify
against, and it would reintroduce an identity dependency the constitution
forbids.

## 4. Streak, ranking, and medal computation

**Decision**: All three are computed and persisted server-side in the new
Go `streak` and `gamification` domains, using a fixed UTC day boundary for
streaks (per prior clarification) and a global weekly leaderboard with
league tiers that promote/demote at week boundaries (per prior
clarification). Streak increments on *submission acceptance*, not on
completed grading (per prior clarification) — this is enforced in the
`essay` domain at the same transaction that writes the `correction_jobs`
row and consumes the day's quota.

**Rationale**: Directly required by Principle III ("computed server-side
only — the client never computes or self-reports gamification state") and
matches the already-resolved product clarifications from the constitution
session.

**Alternatives considered**: None — these were closed decisions carried
forward from `/speckit.constitution`, not re-opened here.

## 5. CI/CD shape (pre-commit + pipeline)

**Decision**: A single repo-root `pre-commit` config (the `pre-commit`
framework, language-agnostic) runs per-codebase fast hooks: `gofmt`
`golangci-lint` for Go, `ruff` + `mypy` for Python (both already used by
`ai-corrector`, per its existing `Makefile`), and `eslint`/`tsc` for
mobile (already used, per its `pnpm lint`/`pnpm typecheck` scripts). Full
CI runs as three parallel GitHub Actions jobs (`backend-ci.yml`,
`correction-service-ci.yml`, extending existing `mobile-ci.yml`), each
running that codebase's full test suite with coverage measured and a 90%
floor enforced (`go test -coverprofile` + `go tool cover`; `pytest --cov`
with `--cov-fail-under=90`; `jest --coverage` with a
`coverageThreshold` of 90). Branch protection on `main` requires all three
jobs green plus one human review before merge.

**Rationale**: Directly required by this feature's own FR-009 through
FR-013 and Principle VI. Reuses each codebase's already-established
tool choice (nothing new introduced: `golangci-lint`, `ruff`/`mypy`,
`eslint`/`tsc` are all already present in the repo's Makefiles/configs) so
the pipeline enforces existing conventions rather than inventing new ones.

**Alternatives considered**: A single monolithic CI job running everything
serially — rejected: slower feedback, and a failure in one codebase would
be harder to attribute at a glance; three parallel jobs give clear
per-codebase pass/fail and let contributors who touch only one codebase get
faster signal.
