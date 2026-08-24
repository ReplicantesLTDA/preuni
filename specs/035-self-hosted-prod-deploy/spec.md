# Feature Specification: Self-Hosted Production Deployment

**Feature Branch**: `035-self-hosted-prod-deploy`
**Created**: 2026-08-24
**Status**: Draft
**Input**: User description: "First production deployment of preuni: self-hosted on the user's own TrueNAS NAS (behind CGNAT, no port forwarding), reached via Cloudflare Tunnel using the already-owned domain preuni.com.br, deployed by a self-hosted GitHub Actions runner running on the NAS itself. Closes GitHub issue #68 (no real deploy target exists yet, main branch never shipped) and #69 (ai-corrector/deploy/docker-compose.prod.yml is stale and architecturally wrong -- stands up its own Postgres and JWT auth instead of sharing the monolith's Postgres via the DB-mediated correction schema bridge already proven working in local dev)."

## Investigation Summary

Two open GitHub issues describe the same underlying gap from different angles:

- **#68**: preuni has never actually been deployed anywhere. `infra/docker-compose.yml` is explicitly dev-only (build contexts, no TLS, no secrets injection). No IaC, no CD, no hosting decision made. This is the real launch gate.
- **#69**: ai-corrector's `deploy/docker-compose.prod.yml` (written before ai-corrector was merged into this monorepo) still assumes it's a standalone product — its own Postgres container, its own JWT auth, its own public API, its own Caddy TLS proxy. Deploying from it today would resurrect a second identity system the constitution explicitly killed off. It needs a real rewrite against wherever preuni actually deploys, or deletion.

Resolving #69 correctly requires #68 to exist first — there is no "real shared Postgres in production" to point ai-corrector's worker at until a production target exists at all. This feature defines that target and, in the same pass, gives ai-corrector's production deployment its correct shape.

**Chosen target** (confirmed with the user): the user's own TrueNAS NAS, run at home/self-hosted for cost reasons at this early stage (explicitly a temporary choice — "when the product requires, I'll upgrade to a more robust environment"). The NAS sits behind CGNAT with no port-forwarding option, so inbound traffic cannot reach it directly. `preuni.com.br` is already owned and ready to point at whatever solves that.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - preuni is reachable on the internet, for real (Priority: P1)

Someone visits `preuni.com.br` (or the API/app endpoints under it) from anywhere on the internet and reaches a running instance of the app — not a dev laptop, not `localhost`. This is the first time that has ever been true.

**Why this priority**: This is the actual launch gate (#68's own words). Nothing else about "production" matters until this is real.

**Independent Test**: From a network outside the NAS's home network (e.g. mobile data, a friend's wifi), hit a health endpoint under `preuni.com.br` and get a valid response over HTTPS.

**Acceptance Scenarios**:

1. **Given** the production stack is running on the NAS, **When** a request is made to `preuni.com.br` from the public internet, **Then** it reaches the running application over a valid HTTPS connection, without the NAS having any port directly forwarded from its router.
2. **Given** the NAS's home internet connection drops and reconnects (CGNAT can reassign its address at any time), **When** connectivity returns, **Then** `preuni.com.br` becomes reachable again without any manual reconfiguration.
3. **Given** the production stack, **When** its component processes (app, database, correction service) are inspected, **Then** they are the same containerized services already proven working together in local development — not a separately-maintained "production" implementation.

---

### User Story 2 - Shipping code means it goes live, not just merges (Priority: P2)

A developer merges a change to the project's integration branch. Some time later, without anyone touching the NAS by hand, that change is running in production.

**Why this priority**: Directly follows from P1 — a production target nobody can deploy to without manual, undocumented server work isn't meaningfully different from having none. This is what makes "launched" sustainable rather than a one-time miracle.

**Independent Test**: Merge a trivial, observable change (e.g. a version string or health-check response) and confirm it appears at `preuni.com.br` without any manual step on the NAS.

**Acceptance Scenarios**:

1. **Given** a change is merged to the project's integration branch, **When** the deployment process runs, **Then** the new version is running in production without anyone logging into the NAS to run a command by hand.
2. **Given** the NAS is offline when a deploy would normally trigger, **When** it comes back online, **Then** the pending deployment completes on its own rather than being silently lost.
3. **Given** a deploy introduces a container that fails to start, **When** this happens, **Then** the previously-running, working version keeps serving traffic rather than the whole stack going down.

---

### User Story 3 - Essay grading actually works in production (Priority: P1)

A real user, in production, submits an essay and receives a real graded correction — not a typed failure caused by the correction service being unreachable, misconfigured, or running against the wrong database.

**Why this priority**: This is the MVP's core loop (per the user's own repeated framing across this project). Equal priority to User Story 1 — a production deployment that doesn't actually grade essays isn't a working production deployment of preuni specifically.

**Independent Test**: In the production environment, submit an essay through the real flow and observe it reach a graded state.

**Acceptance Scenarios**:

1. **Given** the production stack is running, **When** an essay is submitted, **Then** the correction service processes it against the same shared database the rest of the app uses — not a separate, standalone database of its own.
2. **Given** the production correction service, **When** its deployment configuration is inspected, **Then** it has no identity/auth system, no public API, and no database of its own — matching the DB-mediated bridge architecture already proven correct in local development (`032-corrector-service-integration`).
3. **Given** the production stack, **When** its network exposure is inspected, **Then** the correction service is reachable only from other containers in the same stack, never from the internet directly (same requirement as local dev, now also true in production).

---

### Edge Cases

- What happens to in-flight requests during a deploy? A brief interruption is acceptable at this stage (single-node, no high-availability requirement yet) as long as the stack recovers to a fully working state without manual intervention.
- What happens if the NAS itself loses power or needs a reboot? The full stack (tunnel, app, database, correction service) must come back up on its own when the NAS restarts, without requiring anyone to manually start containers.
- What happens to submitted-but-not-yet-graded essays if the correction service container restarts during a deploy? Must resolve the same way local dev already proven it does (`032`'s worker-restart resilience) — no resubmission required, no silently lost job.
- What happens to the production database if the NAS's storage fails? Out of scope for this feature to solve fully (that's a hardware/backup-strategy question), but the deployment must not make backups any harder than they'd otherwise be — data lives on a persistent volume a NAS-level backup can reach.
- What happens if Cloudflare's tunnel service itself has an outage? Out of scope to mitigate (single point of failure inherent to the chosen self-hosted, cost-minimizing approach) — accepted tradeoff for this early stage, revisit when upgrading past self-hosting.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The production application MUST be reachable over HTTPS at `preuni.com.br` (and its relevant subdomains/paths) from the public internet, without any inbound port forwarded on the NAS's router.
- **FR-002**: The production deployment MUST run the same containerized services already proven correct in local development (monolith, gateway, shared database, correction service worker) rather than a parallel, separately-maintained production implementation.
- **FR-003**: The correction service in production MUST share the same database instance as the rest of the application (schema-isolated, DB-mediated bridge) and MUST NOT run its own database, its own authentication system, or its own public-facing API — resolving #69.
- **FR-004**: Merging a change into `main` (the release branch) MUST result in that change running in production without a person manually operating the NAS. `dev` remains the ordinary integration branch every feature PR targets and does not itself trigger a deploy.
- **FR-004a**: There MUST be a defined, repeatable way to promote a ready state of `dev` into `main` (the trigger for FR-004) — this is a deliberate release step, not an automatic side effect of every `dev` merge.
- **FR-005**: The deployment process MUST be resilient to the NAS or its network connection being temporarily unreachable — a pending deploy completes once connectivity returns rather than being lost.
- **FR-006**: A deploy that fails to bring up a working container MUST NOT take down an already-working production instance — the prior working version continues serving traffic.
- **FR-007**: All production secrets (database credentials, signing keys, tunnel credentials, LLM provider keys) MUST be injected at deploy/runtime and MUST NOT be committed to the repository at any point.
- **FR-008**: The production database MUST persist across container restarts and NAS reboots, on storage a NAS-level backup process can reach.
- **FR-009**: The correction service's production deployment configuration (replacing the current `ai-corrector/deploy/docker-compose.prod.yml`) MUST NOT expose a public port, a reverse proxy, or TLS termination of its own — the single public entry point remains the same gateway already used in local development.

### Key Entities

- **Production stack**: the running set of containers (monolith, gateway, shared database, correction service worker/api) deployed to the NAS — conceptually identical to the local dev stack proven in `032-corrector-service-integration`, just running in a different location with real secrets and real persistent storage.
- **Deployment pipeline**: the automated process that takes a merged change and results in it running on the NAS, without manual server access.
- **Tunnel/ingress**: whatever mechanism makes the NAS's stack reachable from `preuni.com.br` despite having no forwardable public port.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `preuni.com.br` responds successfully over HTTPS from an arbitrary external network within the time it takes to run a single health-check request — no manual step required to "turn it on."
- **SC-002**: A merged code change is running in production without any person accessing the NAS directly, end to end, in under the time of one normal deploy cycle (no fixed number mandated at this stage — the requirement is "automatic," not "fast").
- **SC-003**: An essay submitted against the production deployment reaches a graded (or typed-failure) terminal state under the same conditions and timeout already proven in local development — not blocked by a misconfigured or unreachable correction service.
- **SC-004**: Zero secrets (database passwords, signing keys, API keys, tunnel tokens) appear anywhere in the git history of this repository as a result of this feature.
- **SC-005**: The NAS can reboot (simulating a power loss) and the full stack returns to a working, reachable state with no manual intervention.

## Assumptions

- **TrueNAS SCALE**, not TrueNAS CORE — SCALE has native Docker/container-app support; CORE (FreeBSD-based) does not. If the NAS is actually running CORE, the plan phase needs to revisit this (e.g. a VM running Linux + Docker inside CORE).
- **Cloudflare Tunnel** (`cloudflared`) is the ingress mechanism — the only realistic option for a CGNAT'd, non-port-forwardable host reaching a domain the user already controls, and it requires no change to the router at all (outbound-only connection).
- **`main` is revived as the release branch that triggers production deploys** (confirmed with the user) — `dev` remains the integration branch every feature PR targets, unchanged; this feature adds a promotion step (`dev` → `main`, e.g. a PR or fast-forward merge once `dev` is in a state worth shipping) as new process. `main` is currently stale ("Initial commit from Specify template," per #68) and needs to be brought up to date with `dev` as part of this feature, not left behind.
- **Self-hosted GitHub Actions runner on the NAS** is the deploy mechanism, rather than a push-based approach (e.g. SSH from GitHub-hosted runners) — a self-hosted runner only needs outbound connectivity to GitHub, consistent with the CGNAT constraint, and needs no additional tunneling infrastructure beyond what already exists for the app's own ingress.
- **Single-node, no high-availability** — acceptable for this stage per the user's own framing ("when the product requires, I'll upgrade"). Brief downtime during deploys or NAS reboots is acceptable; multi-node failover is explicitly out of scope.
- **Mobile app store / Expo release distribution is out of scope** — tracked separately by GitHub issue #70 ("No mobile release config"). This feature covers the backend/correction-service deployment target only.
- **Backup strategy** (schedule, retention, off-NAS copy) is acknowledged as necessary (FR-008, SC edge case) but the specific policy is a plan-phase detail, not a scope-defining decision for this spec.
- **Secrets management**: a `.env`-style file local to the NAS (populated once, outside git), consistent with how this project already avoids secret-vault tooling (constitution's Security & Secrets principle explicitly rejects that complexity class at current scale) — GitHub Actions Secrets may additionally hold whatever the deploy pipeline itself needs to authenticate (e.g. a tunnel token), but application runtime secrets live on the NAS, not in CI.
