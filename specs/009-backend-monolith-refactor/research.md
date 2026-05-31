# Research: Backend Monolith Refactor

**Feature**: 009-backend-monolith-refactor
**Date**: 2026-05-25
**Status**: Complete

---

## Finding 1: The backend is already organized as a Go workspace of service modules

**Decision**: Introduce the monolith as a new Go module (`backend/svc/monolith`) and add it to `backend/go.work`, instead of immediately collapsing all services into one module.

**Rationale**:
- `backend/go.work` currently composes `pkg` plus multiple `svc/*` Go modules.
- A new module can reuse existing packages/handlers with minimal disruption.
- Keeping service binaries runnable during the migration reduces cutover risk and enables rollback.

**Alternatives considered**:
- Collapse all Go code into a single `backend/` module immediately: rejected; high churn and a harder rollback story.
- Keep services split and only add a “gateway service”: rejected; does not satisfy the single deployable unit requirement.

---

## Finding 2: Route compatibility is best preserved by extracting router builders

**Decision**: Refactor `auth` and `user` services to expose router construction functions (router builders) that can be mounted by both their current `cmd/server/main.go` and the monolith.

**Rationale**:
- Today, route definitions live in each service’s `main.go`.
- Exported router builders allow the monolith to assemble a unified router without duplicating route definitions.
- This preserves existing behavior (middlewares, status codes, error envelopes) while enabling consolidation.

**Alternatives considered**:
- Copy/paste route definitions into monolith: rejected; guarantees drift.
- Rewrite handlers into a new framework: rejected; unnecessary scope expansion.

---

## Finding 3: The API gateway is the compatibility surface for clients

**Decision**: Preserve the gateway-facing HTTP surface (paths, methods, auth requirements, request-id propagation) and cut over NGINX to route `/v1/*` prefixes to the monolith.

**Rationale**:
- All clients reach the backend via the NGINX gateway.
- NGINX also owns rate limiting and request-id header propagation.
- A monolith cutover can be done by changing upstream routing, without changing client behavior.

**Notable current gateway expectations**:
- The gateway routes refresh traffic under `/v1/auth/refresh`, so the monolith must serve that endpoint.
- If the cutover switches to routing all `/v1/auth/*` traffic to the monolith, endpoints like `/v1/auth/account` become gateway-reachable even if current regex routing does not include them.

**Alternatives considered**:
- Expose the monolith directly to clients and remove NGINX: rejected; changes deployment topology and rate limiting behavior.

---

## Finding 4: Internal HTTP endpoints are a real contract (even in a monolith)

**Decision**: Keep internal endpoints as HTTP routes with the same internal-token protection, even if the monolith can also call the same logic in-process.

**Rationale**:
- `POST /internal/students` is used by auth during registration.
- `POST /internal/email/send` is used for transactional emails.
- Preserving these endpoints maintains the current “service boundary” contract and simplifies incremental migration.

**Alternatives considered**:
- Remove internal endpoints and only do in-process calls: rejected; breaks existing split-service mode and makes rollback harder.

---

## Finding 5: Mail sending contract must be preserved while rewriting to Go

**Decision**: Re-implement the current mail contract in Go and preserve:
- `POST /internal/email/send` payload shape
- supported `type` values: `WELCOME`, `EMAIL_VERIFY`, `OTP_LOGIN`, `EMAIL_CHANGE`, `PASSWORD_RESET`
- parameter validation rules (422 on missing required params)
- success semantics (`202 Accepted` + `{ "status": "queued" }`)

**Rationale**:
- auth-svc relies on this internal endpoint for verification OTP, OTP login, email change, password reset, and welcome email.
- Current behavior is fire-and-forget; failures should not block primary flows unless explicitly designed to.

**Notable current behavior**:
- `WELCOME` requires `display_name` and `verification_link` params, but auth currently sends only `display_name`.
- `PASSWORD_RESET` currently uses the same underlying builder as verification emails; preserving behavior avoids template drift during refactor.

**Alternatives considered**:
- Keep the Elixir mail service as a sidecar: rejected; violates “mail is rewritten to Go” requirement for the monolith.
- Change email param requirements during migration: rejected; risks unplanned UX changes.

---

## Finding 6: Inter-service HTTP calls should become interface-driven dependencies

**Decision**: Introduce interface boundaries for cross-service actions (student provisioning, email sending) so the same handlers can run:
- in split-service mode (HTTP adapters)
- in monolith mode (in-process adapters)

**Rationale**:
- auth currently calls user/mail via HTTP using `USER_SERVICE_URL` / `MAIL_SERVICE_URL`.
- In monolith mode, self-HTTP is unnecessary overhead and complicates observability.
- Interface-based adapters allow incremental migration without forcing an all-at-once rewrite.

**Alternatives considered**:
- Keep self-HTTP calls in monolith: rejected; adds latency and makes internal auth token logic awkward.

---

## Finding 7: Shared middleware + error envelope should remain the global standard

**Decision**: Use `backend/pkg/middleware` consistently in monolith routing to preserve:
- JWT auth behavior (`RequireAuth`)
- internal token behavior (`InternalAuth`)
- standard JSON error envelope (`{ "error": { ... } }`)

**Rationale**:
- Multiple services already depend on this shared behavior.
- Consistency at the boundary reduces client regression risk during consolidation.
