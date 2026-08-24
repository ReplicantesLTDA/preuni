<!--
Sync Impact Report
==================
Version change: 2.3.1 → 2.3.2 (PATCH — wording only: corrected the
stale "Go 1.24" claim to describe the actual toolchain pin
(go1.26.6, auto-fetched) accurately; no principle, gate, or rule changed)

Prior: 2.3.0 → 2.3.1 (PATCH — wording only: the correction
service's directory was renamed corretor-redacao/ → ai-corrector/
throughout; no principle, gate, or rule changed)

Prior: 2.2.0 → 2.3.0 (MINOR — new Principle VII: Security &
Secrets added; Principle II strengthened with mandatory, blocking
per-language type-checking; Principle VI's migration-reversibility bullet
strengthened from "ship a downgrade" to "prove the downgrade works";
Governance gained ADR discipline)

Modified:
- II. Testing Standards — added a mandatory, blocking type-check bullet
  per language (Go compiler, `tsc --noEmit`, mypy strict) with an
  explicit "no `Any`/`interface{}` across an API boundary" rule.
- VI. Engineering Workflow — the migration-reversibility bullet now
  requires the downgrade to be proven (round-trip executed), not just
  present. Python: `alembic downgrade base && upgrade head` in CI. Go:
  adopted a `NNN_name.down.sql` sibling-file convention (no migration
  tool in use) with a grandfather list
  (`infra/migrations/.grandfathered`) for pre-existing forward-only
  migrations, a presence-check gate, and a round-trip runner that
  activates automatically per schema once that schema's whole chain has
  downgrades.
- Quality Gates — gained matching checklist items for the type-check and
  round-trip gates, plus new Security & Secrets items.
- Governance — added ADR discipline: non-trivial architectural decisions
  get a numbered file under `docs/decisions/`, not just constitution
  prose or PR descriptions.

Added sections:
- VII. Security & Secrets (NON-NEGOTIABLE) — secrets never in the repo,
  gitleaks secret-scanning (pre-commit + CI, baseline-aware), and
  per-language dependency vulnerability audits.

Removed sections: none.

Context: prompted by comparing against a reference constitution from
another (work) project; SSO/Key Vault/RBAC/Teams-specific pieces were
deliberately not adopted (no multi-user staff access to gate). CD/deploy
principles were deliberately deferred — no real deploy target exists yet.
Full rationale and alternatives considered: docs/decisions/0001-*.md.

Templates requiring updates:
- .specify/templates/plan-template.md — ⚠ pending (verify Constitution
  Check gate references principle VII once a plan touches secrets/deps)
- .specify/templates/spec-template.md — ✅ no changes needed (generic)
- .specify/templates/tasks-template.md — ✅ no changes needed (generic)
- .specify/templates/commands/*.md — ✅ no agent-specific references found

Follow-up TODOs:
- backend/app's go.mod declares `go 1.25.0` while backend/pkg's declares
  `go 1.24` — both pin `toolchain go1.26.6` (auto-fetched) for security
  patches, so the actual build/CI toolchain is correct and consistent;
  only the two modules' `go` directives themselves still differ. Docs
  (CLAUDE.md, README.md) updated 2026-08-24 to describe this accurately
  instead of the old blanket "Go 1.24" claim. Unifying the two `go`
  directives is still open, low-priority cleanup.
- mobile's `pnpm audit` gate blocks on critical only (34 high/moderate
  findings are transitive Expo/Metro build-tooling debt, tracked but not
  yet fixed) — revisit as Expo SDK releases catch up.
-->

# preuni.com.br Constitution

## Product

preuni is an **essay-challenge app**: students write ENEM-style essays as often
as they can, get them graded by an AI corrector, and build a daily submission
streak. Gamification (streaks, friends, medals, weekly ranking) is a core
product mechanic, not a bolted-on feature — every user-facing decision should
default to reinforcing the daily-habit loop, the way Duolingo does for
language practice.

- **Free tier**: one graded essay submission per day.
- **Pro tier**: multiple submissions per day.
- **Social**: users can add friends, see friends' streaks and latest grades.
- **Ranking**: a weekly leaderboard aggregates grades; users rank up or down
  week over week.
- **Medals**: awarded for streak milestones and ranking achievements.

## Architecture

The preuni backend is split across two systems with a clear ownership boundary:

- **Go monolith** (`backend/app/`) remains the system of record for auth, user
  profiles, streaks, friends, medals, ranking, and quota enforcement (free vs.
  Pro submission limits). Domains are organized as internal subpackages under
  `backend/app/internal/<domain>/`. Shared infrastructure (logger, errors,
  middleware, config) lives in `backend/pkg/`.
- **Correction service** (`ai-corrector/`, Python/FastAPI) owns AI-driven
  essay grading only: accepting an essay + prompt theme, running the LLM
  correction pipeline, and returning the structured per-competency result.
  Python is kept deliberately — the correction pipeline is LLM-integration-heavy
  and Python's ecosystem is the better fit; it is **not** to be rewritten to Go.
  The Go monolith calls this service over internal HTTP; the correction service
  never talks to end users directly and never owns auth, streaks, or ranking
  data.
- NGINX remains the external gateway. Postgres and Redis remain the only
  stateful stores; the correction service may use its own schema/tables but
  must not duplicate identity, streak, or ranking data owned by the monolith.

References to "the auth service" or "the mail service" elsewhere in this
document or in older specs refer to domain packages within the Go monolith,
not to separate processes. "The correction service" refers specifically to
`ai-corrector`.

## Core Principles

### I. Code Quality (NON-NEGOTIABLE)

Every piece of code merged to `main` must meet these standards:

- **Readability first**: Code is written once but read many times; clarity beats cleverness
- **Single responsibility**: Functions and modules do one thing well; no god objects or catch-all utilities
- **No dead code**: Unused imports, variables, functions, and feature flags must be removed before merge
- **Explicit over implicit**: Avoid magic values, undocumented side effects, and surprise mutations
- **Consistent naming**: Follow the conventions already present in the file/module; do not mix styles
- **No premature abstraction**: Three similar lines of code are better than a wrong abstraction; extract only when a third use case appears

### II. Testing Standards (NON-NEGOTIABLE)

- **Test-first for new features**: Write failing tests → get approval → implement → pass tests (Red-Green-Refactor)
- **Unit tests required** for all business logic, utilities, and pure functions
- **Integration tests required** for: API endpoints, database interactions, authentication flows, and third-party service boundaries (including the Go monolith ↔ correction service boundary)
- **No test skipping**: `skip`, `xit`, `xtest`, or equivalent are forbidden in CI — comment them with a tracked issue instead
- **Test names must describe behavior**: `it("returns 404 when user does not exist")` not `it("works")`
- **Coverage floor**: Maintain ≥ 90% line coverage across backend (Go), correction service (Python), and mobile (TypeScript); new code must not lower the project average; CI fails the build below the floor
- **Real dependencies over mocks** at integration boundaries: mock only what you own or what is external and unreliable
- **Type-checking is mandatory and blocking, per language**: Go's compiler enforces this by construction; the correction service runs `mypy` in `strict` mode (`ai-corrector/pyproject.toml`'s `[tool.mypy]`) and mobile runs `tsc --noEmit` — both block CI and pre-commit, no `continue-on-error` exceptions. No `Any` (Python) or an untyped/`any`-shaped payload crosses an API boundary (HTTP request/response bodies, the Go↔correction-service bridge table, the mobile↔backend wire format) without a concrete type or schema on both sides

### III. Gamification & UX Consistency

- **Streak integrity**: Streak, medal, and ranking calculations must be deterministic, covered by tests, and computed server-side only — the client never computes or self-reports gamification state
- **Design system adherence**: All UI components must use tokens from the design system (colors, spacing, typography); raw hex/px values are not allowed in component styles
- **Consistent feedback patterns**: Loading states, error messages, and empty states must follow the established patterns for the application; no one-off spinners or ad-hoc error strings
- **Accessible by default**: Every interactive element requires keyboard navigability and ARIA labels; WCAG AA is the minimum bar
- **Mobile-first**: Layouts are designed and reviewed on mobile before desktop; no feature ships without a responsive implementation
- **Copy consistency**: User-facing strings follow the established tone (friendly, direct, encouraging for a pre-university audience); no technical jargon exposed to end users
- **Predictable navigation**: URLs are stable and bookmarkable; back-button behavior must work as expected
- **No dark-pattern gamification**: Streak-loss messaging, Pro upsells, and ranking notifications must be honest and never manipulate users with false urgency or hidden mechanics

### IV. Performance Requirements

- **Page load (LCP)**: ≤ 2.5 s on a simulated mid-tier mobile device (Lighthouse, throttled 4G)
- **Interaction responsiveness (INP)**: ≤ 200 ms for all user interactions
- **Bundle size budget**: No single route chunk > 150 kB (gzipped); new dependencies must be evaluated for size impact before adoption
- **Image optimization**: All images served in next-gen formats (WebP/AVIF) with explicit `width`/`height` to prevent layout shift (CLS ≤ 0.1)
- **API response time**: p95 < 500 ms for all user-facing endpoints, excluding the async essay-correction pipeline itself (correction is submit-then-poll, not synchronous)
- **No N+1 queries**: Every new data-fetching path must be reviewed for N+1 patterns; use batching or eager loading where appropriate

### V. AI Correction Integrity (NON-NEGOTIABLE)

- **Structured, explainable grading**: Every correction must break the essay into the 5 official ENEM competencies with verbatim excerpts and pt-BR justifications — a bare numeric score with no explanation is not a valid correction
- **Async by default**: Essay submission returns immediately (202-style); grading happens out-of-band and the client polls or is notified — the user-facing API must never block on LLM latency
- **Quota enforcement at the boundary**: Free-tier (1/day) vs. Pro (multi/day) submission limits are enforced by the Go monolith before a request reaches the correction service, not inside the correction pipeline
- **Prompt/version discipline**: Changes to correction prompts or grading logic bump a tracked version and report a metric delta (golden-set comparison) before merge — grading behavior must not silently drift
- **Provider swappability**: The LLM provider integration must stay isolated behind an abstraction; correction logic must not import transport/API details of a specific provider directly

### VI. Engineering Workflow (NON-NEGOTIABLE)

preuni is built XP-style, adapted for an LLM pair: tests, clean code, continuous
integration, pair programming, and continuous deploy — with the AI driving
the keyboard and the human directing, reviewing, and correcting cheaply and
early, the way you'd drive a very fast pair. The AI's code is the 10%; the
other 90% — deciding what to build, reviewing it, deciding what merges — stays
ordinary, disciplined software engineering.

- **TDD is enforced, not aspirational**: the Red-Green-Refactor loop from
  Principle II runs locally via **pre-commit** (fast unit tests + lint +
  type-check block the commit) and again in full in **CI/CD** (unit +
  integration + golden-set suites) on every push
- **`main` is protected**: no direct pushes, no force-push, no bypassing
  status checks — the only way code reaches `main` is a pull request with all
  required CI checks green
- **Nothing broken merges**: a red CI run blocks merge unconditionally; there
  is no "merge anyway" override for a failing required check
- **Human-in-the-loop review is mandatory**: every PR requires human approval
  before merge, even when the AI authored the change — "the tests pass" is
  necessary but not sufficient, the same as Principle II already states for
  "it works"
- **Pair programming is human + AI**: the human plays navigator/reviewer —
  states intent, watches execution, interrupts and corrects while the error
  is still cheap to fix — rather than reviewing only a finished diff after
  the fact
- **Small, continuous integration**: work lands in small PRs merged
  frequently against an up-to-date `main`, not long-lived branches that
  diverge for weeks
- **Every migration is reversible, and CI proves it, starting now**: any
  migration that adds an `upgrade`/forward step (Alembic revisions,
  `infra/migrations/*.sql`, or equivalent) MUST ship with a working
  `downgrade`/rollback counterpart in the same PR — no migration gaps —
  and CI MUST actually execute the round-trip, not just check the
  downgrade exists. Correction service: `alembic downgrade base &&
  alembic upgrade head` runs in CI on every push. Go monolith:
  `infra/migrations/<schema>/NNN_name.sql` pairs with a sibling
  `NNN_name.down.sql`; `backend/scripts/check-migration-downgrades.sh`
  gates on the pairing existing, and `backend/scripts/migrate-round-trip.sh`
  actually executes the round-trip for any schema whose entire migration
  chain has downgrades. Applies prospectively from the date this bullet
  was ratified; pre-existing forward-only migrations are grandfathered
  (`infra/migrations/.grandfathered`) — not retroactively blocked, but
  MUST gain a downgrade (and be removed from that list) before they are
  ever edited again.

### VII. Security & Secrets (NON-NEGOTIABLE)

- **Secrets never reach the repo**: no API key, DB credential, JWT
  signing key, or LLM provider token is ever committed, in code, config,
  `.env` files, or fixtures — real config lives in local `.env` files
  (gitignored) or CI/CD secret stores, never in git history
- **Secret scanning runs on every commit and every push**: `gitleaks`
  runs in pre-commit (`gitleaks protect --staged`) and in CI
  (`.github/workflows/security-ci.yml`), both baseline-aware
  (`.gitleaks-baseline.json` — the pre-existing, confirmed-harmless
  test-fixture findings frozen at v2.3.0's ratification). A real secret
  is never added to the baseline — it's rotated and the leak is treated
  as an incident, not silenced
- **Dependency vulnerability audits run per language, every PR**:
  `govulncheck` (Go), `pip-audit` (correction service), `pnpm audit`
  (mobile) — in `security-ci.yml`. A newly-introduced dependency with a
  known critical vulnerability blocks merge; exceptions for pre-existing,
  not-safely-fixable transitive findings must be documented inline in
  the workflow file with the specific reason (see mobile's `pnpm audit`
  step for the current example) — the same "non-blocking with a
  documented plan" pattern this repo already uses for pre-existing
  gofmt/golangci-lint debt, never a silent gap
- **No enterprise-specific security theater**: SSO, RBAC, and centralized
  secret-vault tooling (Key Vault, etc.) are deliberately NOT required
  here — preuni has no multi-user staff access to gate. Re-evaluate if
  that changes

## Quality Gates

Every pull request must pass all of the following before merge:

- [ ] CI is green (lint, type-check, unit tests, integration tests)
- [ ] Coverage has not decreased
- [ ] Lighthouse CI score does not regress on LCP, INP, or CLS
- [ ] Bundle size budget is not exceeded
- [ ] Design system tokens used (no raw values in styles)
- [ ] Accessibility: no new axe-core violations introduced
- [ ] Reviewer has verified mobile layout
- [ ] Gamification-affecting changes (streaks, medals, ranking) have deterministic tests
- [ ] Correction-pipeline changes include a golden-set metric delta in the PR description
- [ ] Pre-commit hooks ran clean (unit tests + lint + type-check)
- [ ] Full CI/CD pipeline is green on the PR's latest commit
- [ ] At least one human approval is recorded on the PR
- [ ] Any new migration includes a working downgrade/rollback, and CI's round-trip step covers it once its schema's whole chain is paired
- [ ] Type-checks pass and block (Go compiler, `mypy` strict, `tsc --noEmit`) — no `continue-on-error` on any of them
- [ ] Secret scanning (`gitleaks`) is clean against the current baseline
- [ ] Dependency vulnerability audits (`govulncheck`, `pip-audit`, `pnpm audit`) show no new critical/high finding
- [ ] A non-trivial architectural decision in this PR has a corresponding `docs/decisions/NNNN-*.md` entry

## Governance

- This Constitution supersedes all other practices and informal agreements
- **Architectural decisions get an ADR**: any non-trivial, hard-to-reverse
  choice (a new dependency category, a schema/ownership boundary, a
  security posture trade-off, adopting or rejecting a tool) gets a
  numbered entry in `docs/decisions/` (template: `docs/decisions/template.md`)
  — context, decision, alternatives considered, consequences — instead of
  living only in constitution prose, a PR description, or nowhere. Small,
  easily-reversed choices don't need one; when in doubt, write it
- Amendments require: written proposal, team discussion, documented rationale, and update to this file
- All code reviews must verify compliance with these principles — "it works" is not sufficient approval
- Exceptions must be documented inline with a comment referencing a tracked issue and an expiry plan
- Complexity must be justified; the burden of proof is on the author adding it
- `main` is a protected branch: all changes land via pull request, all required
  CI checks must be green, and at least one human reviewer must approve —
  no exceptions, including for AI-authored changes

**Version**: 2.3.2 | **Ratified**: 2026-04-03 | **Last Amended**: 2026-08-24
