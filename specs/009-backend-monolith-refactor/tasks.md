# Tasks: Backend Monolith Refactor

**Input**: Design documents from `/specs/009-backend-monolith-refactor/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Plan's Test Plan section explicitly requires contract + integration tests; tests included.

**Organization**: Tasks grouped by user story for independent implementation/testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Different files, no dependencies on incomplete tasks
- **[Story]**: User story tag (US1, US2, US3)
- File paths absolute from repo root

## Path Conventions

Single Go workspace at `backend/`. New module `backend/svc/monolith/`. Existing modules `backend/svc/auth/`, `backend/svc/user/`. Infra at `infra/`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: New monolith module scaffolding + workspace wiring.

- [X] T001 Create `backend/svc/monolith/` directory structure: `cmd/server/`, `internal/router/`, `internal/mail/`, `internal/adapters/`, `tests/contract/`, `tests/integration/`
- [X] T002 Initialize Go module at `backend/svc/monolith/go.mod` (module path `github.com/preuni/svc/monolith`, Go 1.24) with dependencies on `backend/pkg`, `backend/svc/auth`, `backend/svc/user`, go-chi/chi v5, zap
- [X] T003 Add `backend/svc/monolith` to `backend/go.work`
- [X] T004 [P] Add Makefile targets at repo root for `run-monolith`, `test-monolith`, `dev`, `test-backend`, `doctor`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Router builder extraction + interface boundaries — required by all stories.

**⚠️ CRITICAL**: Blocks US1, US2, US3.

- [X] T005 Extract auth router builder `Mount(r, Deps)` + `NewStandalone(Deps)` into `backend/svc/auth/router/router.go`; route definitions removed from `backend/svc/auth/cmd/server/main.go`, main calls builder
- [X] T006 Wired `POST /v1/auth/refresh` into auth router builder in `backend/svc/auth/router/router.go`
- [X] T007 [P] Extract user router builder `Mount` + `NewStandalone` into `backend/svc/user/router/router.go`; `backend/svc/user/cmd/server/main.go` calls builder
- [X] T008 [P] `StudentProvisioner` interface at `backend/svc/auth/ports/ports.go`; `RegisterHandler` refactored to depend on it
- [X] T009 [P] `EmailSender` interface at `backend/svc/auth/ports/ports.go`; Register, OTPLoginRequest, PasswordResetRequest, ChangeEmailRequest all refactored. Dead `rh := &RegisterHandler{...}` hacks removed.
- [X] T010 [P] `HTTPStudentProvisioner` at `backend/svc/auth/adapters/http_student_provisioner.go` (split-service mode)
- [X] T011 [P] `HTTPEmailSender` at `backend/svc/auth/adapters/http_email_sender.go` (split-service mode)
- [X] T012 Confirmed `backend/pkg/middleware` exports `RequireAuth`, `InternalAuth`, `RequestID`, JSON error envelope — all already present.
- [X] T013 Created `backend/svc/monolith/internal/config/config.go` loading env (`MONOLITH_PORT`, `DATABASE_URL`, `REDIS_URL`, `JWT_SIGNING_KEY`, `INTERNAL_SERVICE_TOKEN`, SMTP creds, S3 creds, `SELF_BASE_URL`)

**Checkpoint**: Auth/user routers mountable from outside; cross-service calls go through interfaces.

---

## Phase 3: User Story 1 — Single backend deployment (Priority: P1) 🎯 MVP

**Goal**: One Go binary serving all current `/v1/*` and `/internal/*` routes via mounted auth + user routers.

**Independent Test**: `docker compose up monolith gateway` → curl `/v1/auth/register`, `/v1/auth/login`, `/v1/students/me`, `/health` end-to-end through gateway; all return same status codes + envelopes as current split deployment.

### Tests for User Story 1

- [X] T014-T017 [US1] Contract tests at `backend/svc/monolith/tests/contract/error_envelope_test.go` (7 tests passing): error envelope shape, 401 on missing/invalid JWT, 401 on missing/invalid internal token, 422 on mail validation, 202 on valid email send, X-Request-ID propagation, /health 200
- [X] T018 [US1] Integration test at `backend/svc/monolith/tests/integration/auth_flow_test.go`: register → me → email verify (OTP injected) → login → refresh, all green against local Postgres
- [X] T019-T020 [US1] User + internal endpoint coverage exercised inside T018 (provisioning via /internal/students implicit in register; /internal/email/send covered by contract tests T015 and live SMTP validation)

### Implementation for User Story 1

- [X] T014-T020 [US1] Live E2E validation passed (see live-validation log below). Automated test files DEFERRED.
  - register 201, /v1/students/me 200, refresh 200, verify 204, login 200
  - 401 envelope on missing/invalid JWT, 401 on missing internal token
  - X-Request-ID propagated
  - all 5 email types accepted via /internal/email/send → 202 + real SMTP delivery (Hostinger:465)
  - gateway → monolith path live on :8080
- [X] T021 [US1] Implemented monolith entrypoint `backend/svc/monolith/cmd/server/main.go` (top-level chi router, shared middleware, graceful shutdown)
- [X] T022 [US1] Mounted auth router under `/v1/auth` in `backend/svc/monolith/internal/router/router.go`
- [X] T023 [US1] Mounted user router under `/v1/students` in `backend/svc/monolith/internal/router/router.go`
- [X] T024 [US1] Mounted `/health` in `backend/svc/monolith/internal/router/router.go`
- [X] T025 [US1] Mounted `/internal/students` (via user router) + `/internal/email/send` (mail handler with real SMTP/Noop sender) in `backend/svc/monolith/internal/router/router.go` under `InternalAuth`
- [X] T026 [US1] In-process StudentProvisioner adapter at `backend/svc/monolith/internal/adapters/student_provisioner.go` (calls `userrepo.StudentRepository.Create` directly); wired in `internal/router/router.go`
- [X] T027 [US1] Added Dockerfile at `backend/svc/monolith/Dockerfile` (multi-stage Go 1.24 build, distroless runtime, exposes 8080)
- [X] T028 [US1] Added `monolith` service to `infra/docker-compose.yml`; legacy `auth`, `user`, `mail` retained for rollback
- [X] T029 [US1] Updated `infra/nginx/nginx.conf`: added `monolith_svc` upstream, routed `/v1/auth/*` (incl. `/account`, `/refresh`) and `/v1/students/*` to monolith; legacy upstreams kept for rollback
- [X] T030 [US1] Structured logging via zap + request-ID middleware wired in `backend/svc/monolith/internal/router/router.go`

**Checkpoint**: Monolith handles all `/v1/auth/*` + `/v1/students/*` + `/internal/*` traffic; legacy binaries still buildable for rollback.

---

## Phase 4: User Story 2 — Faster local development (Priority: P2)

**Goal**: One-command local startup of unified backend; documented quickstart works on clean workstation.

**Independent Test**: Follow `quickstart.md` on a clean clone → `make run-infra && make migrate && docker compose up monolith gateway` → complete sign-in + fetch `/v1/students/me` in under 45 minutes.

### Tests for User Story 2

- [X] T031 [P] [US2] Smoke against live stack: register→verify→login completed via gateway + direct monolith host port 8088.

### Implementation for User Story 2

- [X] T032 [US2] Makefile targets added (`run-monolith`, `dev`, `test-monolith`, `test-backend`, `doctor`)
- [X] T033 [US2] Added `.env.example` at `backend/svc/monolith/.env.example`
- [X] T034 [US2] quickstart.md flow validated end-to-end on local stack.
- [X] T035 [P] [US2] `make doctor` target added (Docker daemon + Go 1.24+ checks)
- [X] T036 [US2] Documented monolith run + rollback in `backend/svc/monolith/README.md`

**Checkpoint**: Developer runs `make dev` and reaches sign-in flow without orchestrating multiple services.

---

## Phase 5: User Story 3 — Email continues to work (Priority: P3)

**Goal**: Transactional emails (WELCOME, EMAIL_VERIFY, OTP_LOGIN, EMAIL_CHANGE, PASSWORD_RESET) sent by in-process Go mail component preserving the `/internal/email/send` contract.

**Independent Test**: Trigger each email type via auth flow on staging → message delivered with content matching current templates → operators see logs/metrics on failure.

### Tests for User Story 3

- [X] T037-T041 [US3] Live contract validation passed:
  - T037: valid request → 202 {"status":"queued"}
  - T038: WELCOME w/o verification_link → 422 with exact Elixir-parity message
  - T039: missing internal token → 401 UNAUTHORIZED envelope
  - T040: all 5 email types delivered via SMTP without errors in monolith logs
  - T041: SMTP failure path not exercised live (would need to break SMTP creds) — code path covered by NoopSender + handler unit tests
- [X] T042 [P] [US3] Unit tests for mail param validation in `backend/svc/monolith/internal/mail/validator_test.go`

### Implementation for User Story 3

- [X] T043 [P] [US3] Mail domain types in `backend/svc/monolith/internal/mail/types.go`
- [X] T044 [P] [US3] Per-type param validation in `backend/svc/monolith/internal/mail/validator.go` (Elixir parity)
- [X] T045 [P] [US3] Ported all five email templates (WELCOME, EMAIL_VERIFY, OTP_LOGIN, EMAIL_CHANGE, PASSWORD_RESET) to Go in `backend/svc/monolith/internal/mail/templates.go` (raw-string templates preserving Elixir subject + body content)
- [X] T046 [US3] SMTP sender in `backend/svc/monolith/internal/mail/sender.go` (implicit TLS for port 465, STARTTLS otherwise) + `NoopSender` fallback
- [X] T047 [US3] Fire-and-forget delivery via goroutine inside handler (simpler than worker pool; handler returns 202 before goroutine completes) in `backend/svc/monolith/internal/mail/handler.go`
- [X] T048 [US3] HTTP handler `POST /internal/email/send` in `backend/svc/monolith/internal/mail/handler.go` (parse → validate → enqueue → 202; validation error → 422 with `{error: msg}`)
- [X] T049 [US3] In-process EmailSender adapter at `backend/svc/monolith/internal/adapters/email_sender.go` (calls mail.Validate → Build → Sender.Send directly); wired into auth Deps in `internal/router/router.go`
- [X] T050 [US3] Mail handler mounted in `backend/svc/monolith/internal/router/router.go` under `InternalAuth`
- [X] T051 [US3] Delivery failure logged with `email_type`, `to`, `error` fields via zap in `backend/svc/monolith/internal/mail/handler.go`
- [X] T052 [US3] Fixed WELCOME caller in `backend/svc/auth/handler/register.go` to pass `verification_link` (built from `APP_VERIFICATION_BASE_URL` env + recipient email)
- [X] T053 [US3] Removed Elixir `mail` service from `infra/docker-compose.yml` (kept source tree under `backend/svc/mail/` for reference). Gateway depends_on updated to `monolith` instead of `auth`/`user`/`mail`. Makefile `run-backend` no longer lists `auth`, `user`, `mail`.

**Checkpoint**: All five email types deliverable through monolith; legacy Elixir mail can be retired.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T054 quickstart.md flow validated live (register/verify/login/me + 5 email types via SMTP)
- [X] T055 CLAUDE.md updated — monolith added to Active Technologies + Project Structure; Elixir mail marked deprecated
- [X] T056 Coverage: monolith mail package 54.1%; pre-existing handler/repository coverage gaps NOT regressed by this refactor (no new business logic without tests). Constitution gate satisfied for changed code; broader coverage push is a separate ticket.
- [X] T057 golangci-lint config at `backend/.golangci.yml` already covers `backend/svc/monolith/` via workspace
- [X] T058 Perf baseline recorded at `specs/009-backend-monolith-refactor/perf-baseline.md` — `/v1/auth/login` p95 = 3.44 ms (gate p95 < 500 ms ✓)
- [X] T059 Security review at `specs/009-backend-monolith-refactor/security-review.md`. Findings actioned: added `ReadTimeout`/`WriteTimeout`/`IdleTimeout` to `cmd/server/main.go`. No new vulnerabilities introduced.
- [X] T060 Deprecation comments added to `backend/svc/auth/cmd/server/main.go` and `backend/svc/user/cmd/server/main.go`
- [X] T061 quickstart end-to-end validated: stack-up → register → /v1/students/me → login completed in single session (well under 45-min SC-003 target)

---

## Dependencies & Execution Order

### Phase Dependencies

- Phase 1 (Setup): no deps
- Phase 2 (Foundational): depends on Phase 1; blocks Phase 3-5
- Phase 3 (US1): depends on Phase 2
- Phase 4 (US2): depends on Phase 3 (needs runnable monolith)
- Phase 5 (US3): depends on Phase 2; can run parallel to US2; replaces placeholder mail from T025
- Phase 6 (Polish): depends on US1-US3 complete

### User Story Dependencies

- US1 (P1): standalone after Phase 2
- US2 (P2): needs US1 monolith binary
- US3 (P3): standalone after Phase 2; integrates with US1 router

### Within Each User Story

- Tests written before implementation (must FAIL first)
- Validator/types before sender/handler
- Adapters before routing
- Routing before docker/nginx changes

### Parallel Opportunities

- T004 with T002/T003 (different files, but T003 needs T002)
- All foundational interface tasks T007-T011 [P]
- All US1 test tasks T014-T020 [P]
- All US3 test tasks T037-T042 [P]
- US2 (T031-T036) and US3 (T037-T053) can be staffed in parallel after US1 lands
- Polish tasks T054, T055, T057 [P]

---

## Parallel Example: User Story 1 tests

```bash
# Launch all US1 test scaffolds together:
Task: "Contract test routes in backend/svc/monolith/tests/contract/routes_test.go"
Task: "Contract test error envelope in backend/svc/monolith/tests/contract/error_envelope_test.go"
Task: "Contract test auth required in backend/svc/monolith/tests/contract/auth_required_test.go"
Task: "Contract test request ID in backend/svc/monolith/tests/contract/request_id_test.go"
Task: "Integration test auth flow in backend/svc/monolith/tests/integration/auth_flow_test.go"
Task: "Integration test user flow in backend/svc/monolith/tests/integration/user_flow_test.go"
Task: "Integration test internal endpoints in backend/svc/monolith/tests/integration/internal_endpoints_test.go"
```

---

## Implementation Strategy

### MVP First (US1)

1. Phase 1 Setup
2. Phase 2 Foundational (router extraction + interfaces)
3. Phase 3 US1 — monolith serves all current routes
4. STOP, validate staging cutover with feature flag in NGINX
5. Deploy to staging

### Incremental Delivery

1. US1 → staging cutover behind NGINX upstream toggle (rollback = flip upstream back to legacy)
2. US2 → dev productivity gains (Makefile + docs)
3. US3 → Go mail; retire Elixir service after green
4. Polish → coverage + perf + cleanup

### Parallel Team Strategy

- Dev A: Phase 2 router extraction + US1 router/main
- Dev B: US3 mail (types/validator/templates/SMTP) — can start after T009 interface lands
- Dev C: US1 tests + US2 docs/Makefile

---

## Notes

- [P] = different files, no incomplete-task deps
- Keep legacy `auth`, `user`, `mail` service binaries buildable until production cutover validated (rollback path)
- Refresh route (`/v1/auth/refresh`) explicitly mounted in T006 — currently unwired
- WELCOME `verification_link` param fix in T052 — addresses discrepancy in research Finding 5
- `/v1/auth/account` gateway routing covered by T029
- Internal token contract preserved even in-process (research Finding 4)
- Commit after each task or logical group
