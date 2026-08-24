# ADR-0001: Security & Secrets, mandatory type-checking, migration reversibility proof, and ADR discipline

**Status**: accepted
**Date**: 2026-08-23

## Context

The constitution (v2.2.0) covered code quality, testing, gamification/UX,
performance, AI correction integrity, and engineering workflow — but had
no principle for secrets/dependency security, no gate proving Python's
type annotations are actually correct (mypy ran non-blocking), and no way
to prove a migration's downgrade actually works (the v2.2.0 rule required
one to exist, not that it functions). Architectural decisions like "Python
stays for the correction service" lived only in the constitution's prose,
with no separate trail for smaller decisions going forward.

Comparing against a reference constitution from another project (work,
not personal — Azure/Entra/SSO/Key Vault/RBAC pieces don't apply to a
solo project with no deploy pipeline yet) surfaced four gaps worth
closing now: secrets/dependency scanning, blocking type-checks, proven
(not asserted) migration reversibility, and ADR discipline. CD/deploy
gates were explicitly deferred — no real deploy target exists yet.

## Decision

1. **New Principle VII: Security & Secrets** (NON-NEGOTIABLE) — secrets
   never in the repo, gitleaks secret-scanning pre-commit + CI
   (baseline-aware, see `.gitleaks-baseline.json`), and per-language
   dependency vulnerability audits (`govulncheck`, `pip-audit`, `pnpm
   audit`) in a new `security-ci.yml` workflow.
2. **mypy is now blocking** in `correction-service-ci.yml` (was
   `continue-on-error: true`). Fixed the 22 pre-existing errors first
   rather than flip a broken gate on.
3. **Migration reversibility is now proven, not just asserted.** Python:
   `alembic downgrade base && alembic upgrade head` runs in CI — this
   caught and fixed two real bugs (orphaned Postgres ENUM types from
   `001_init`/`002_correction_jobs_schema`'s downgrades). Go: adopted a
   `NNN_name.down.sql` sibling-file convention (no migration tool is in
   use, just a numbered-file bash loop) with a grandfather list
   (`infra/migrations/.grandfathered`) for the ~15 pre-existing
   forward-only migrations, a presence-check gate
   (`check-migration-downgrades.sh`), and a round-trip runner
   (`migrate-round-trip.sh`) that activates automatically per schema once
   that schema's whole chain has downgrades (all schemas currently skip —
   expected, not a failure).
4. **ADR discipline**: `docs/decisions/`, this file being the first entry.
   Architectural decisions get their own numbered file going forward
   instead of living only in constitution prose or PR descriptions.

Also (uncovered while wiring the security gates, fixed as part of making
them real rather than left as new debt): removed a fully-dead
`pydantic-settings` dependency, bumped 3 vulnerable Go modules
(`chi` v5.2.1→v5.3.0, `pgx/v5` v5.7.2→v5.9.2, `x/text` v0.21.0→v0.39.0)
and pinned the Go toolchain to `1.26.6` (was silently running a
vulnerable stdlib patch version), and 2 vulnerable transitive
mobile/Python packages (`h2`, `starlette`, both pulled in current via
normal resolution — no override needed).

## Alternatives considered

- **Adopt the reference project's SSO/Key Vault/RBAC pieces wholesale** —
  rejected, no multi-user staff access to gate; would be pure overhead.
- **Rewrite `infra/migrations/*.sql` history into golang-migrate's
  `.up.sql`/`.down.sql` naming** — rejected for now (more rework, only
  useful if actually adopting golang-migrate later); the sibling-file
  convention gets the same reversibility guarantee without a rename.
- **Block mobile's `pnpm audit` on `--audit-level=high`** — rejected: 34
  findings are transitive, deep inside Expo/Metro's own build-tooling
  chain (dev/build-time code, not shipped in the app bundle), not safely
  fixable without risking Expo's own dependency resolution. Blocking only
  on `critical` (now clean) and tracking high/moderate as debt matches
  this repo's existing golangci-lint/gofmt precedent for exactly this
  situation.
- **Build the full CD/deploy principle from the reference project now** —
  rejected, no real deploy target exists yet; revisit when one does.

## Consequences

- New PRs touching `ai-corrector/src/` must pass mypy — a real,
  if occasionally annoying, constraint that already caught genuinely
  wrong type annotations (e.g. `httpx.BaseTransport` where
  `AsyncBaseTransport` was actually required).
- A new Go migration without a `.down.sql` sibling now fails CI/pre-commit
  instead of merging silently. Editing a grandfathered migration requires
  removing it from the list and adding a downgrade in the same PR.
- The mobile `pnpm audit` gate is honest but incomplete — it does not
  currently catch high/moderate transitive findings. This is a known,
  documented gap, not a silent one.
- Every future non-trivial architectural choice should get an ADR. This
  file is the pattern to follow; keep them short.
