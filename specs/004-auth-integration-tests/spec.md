# Feature Specification: Auth Integration & Unit Tests

**Feature Branch**: `004-auth-integration-tests`
**Created**: 2026-04-04
**Status**: Draft
**Input**: "Integration tests to test frontend/backend communication and be certain that all communication is set. Unit tests at the end. TDD focus."

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Backend Handler Integration Tests (Priority: P1)

A developer running `go test ./handler/...` in `backend/svc/auth` sees the full register → verify-email → login flow pass against a real PostgreSQL instance. Every public endpoint of the auth service is covered by at least one positive and one negative test case that hits the database.

**Why this priority**: The backend contract (HTTP status codes, JSON shape, error codes) is the ground truth that the mobile client depends on. Verifying it first gives confidence that the API is correct before testing the client side.

**Independent Test**: Run `go test ./handler/... -run Integration` in `backend/svc/auth` with `TEST_DB_URL` pointing to a running Postgres. Delivers standalone value — the entire server-side auth flow is verified.

**Acceptance Scenarios**:

1. **Given** a fresh test database, **When** `POST /v1/auth/register` is called with valid body, **Then** HTTP 201 is returned with `student_id`, `access_token`, `refresh_token`, and `expires_in` fields.
2. **Given** a registered but unverified account, **When** `POST /v1/auth/login` is called, **Then** HTTP 403 Forbidden is returned.
3. **Given** a valid OTP was generated during registration, **When** `POST /v1/auth/email/verify` is called with the correct code, **Then** HTTP 204 No Content is returned.
4. **Given** an email-verified account, **When** `POST /v1/auth/login` is called with correct credentials, **Then** HTTP 200 is returned with valid JWT tokens.
5. **Given** duplicate email registration, **When** `POST /v1/auth/register` is called again, **Then** HTTP 409 Conflict is returned.
6. **Given** a logged-in session, **When** `POST /v1/auth/logout` is called with the refresh token, **Then** HTTP 204 is returned and the token is invalidated.

---

### User Story 2 — Mobile AuthApiClient Contract Tests (Priority: P2)

A developer running `./gradlew :shared:desktopTest` sees that `AuthApiClient` sends the correct JSON payloads to each endpoint and correctly deserialises every response and error shape. Tests use a ktor `MockEngine` — no real network or backend required.

**Why this priority**: The client contract (field names, serialization, error mapping) must match the backend contract. Catching mismatches here is faster than running a full E2E test. This is independent of US1.

**Independent Test**: Run `./gradlew :shared:desktopTest --tests "*AuthApiClientTest*"`. Delivers standalone value — the client serialization layer is fully verified.

**Acceptance Scenarios**:

1. **Given** a mock engine returning `{"student_id":"x","access_token":"a","refresh_token":"r","expires_in":3600}`, **When** `register()` is called, **Then** `AuthSession(userId="x", accessToken="a")` is returned.
2. **Given** a mock engine returning HTTP 409, **When** `register()` is called, **Then** the result is `Failure(AppError.Conflict(...))`.
3. **Given** a mock engine returning HTTP 403, **When** `login()` is called, **Then** the result is `Failure(AppError.Forbidden(...))`.
4. **Given** a mock engine returning HTTP 204, **When** `verifyEmail()` is called, **Then** the result is `Success(Unit)`.
5. **Given** a mock engine returning HTTP 401, **When** `login()` is called, **Then** the result is `Failure(AppError.Unauthorized(...))`.

---

### User Story 3 — Mobile Store Unit Tests (Priority: P3)

A developer running `./gradlew :shared:desktopTest` sees `RegisterStore` and `VerifyEmailStore` unit tests pass, covering all intent → state transitions and label emissions. Tests use a fake `AuthRepository`.

**Why this priority**: Store tests verify that the UI state machine (loading states, error propagation, navigation labels) works correctly independently of the API layer. Catches regressions in the presentation logic.

**Independent Test**: Run `./gradlew :shared:desktopTest --tests "*RegisterStoreTest*" "*VerifyEmailStoreTest*"`. Delivers standalone value — all auth store logic is verified.

**Acceptance Scenarios**:

1. **Given** a RegisterStore, **When** `Submit` is dispatched with valid fields, **Then** `Label.Registered` is emitted with the correct email.
2. **Given** a RegisterStore, **When** `Submit` is dispatched with empty email, **Then** `AppError.Validation` is set on state without calling the repository.
3. **Given** a RegisterStore, **When** the repository returns `Conflict`, **Then** `globalError` is set to `AppError.Conflict` and `isLoading = false`.
4. **Given** a VerifyEmailStore, **When** `Submit` is dispatched with a 6-digit code, **Then** `state.verified = true` and `onVerified` is called.
5. **Given** a VerifyEmailStore, **When** the repository returns a validation error, **Then** `state.error` is non-null and `isLoading = false`.

---

### Edge Cases

- What happens when the PostgreSQL test database is unavailable? Integration tests must skip gracefully with a clear message rather than panicking.
- What happens when the OTP has expired? Verify-email must return a validation error, not a 500.
- What happens when the refresh token is already used? Logout must be idempotent (204 even if token is not found).
- What happens when `AuthApiClient` receives an unexpected JSON shape? It must map to `AppError.NetworkError`, not crash.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Integration tests for `backend/svc/auth/handler/` MUST use `net/http/httptest` and a real PostgreSQL connection (via `TEST_DB_URL` env var); tests MUST be skipped if `TEST_DB_URL` is not set.
- **FR-002**: Each handler MUST have at least one test for the happy path and one for each documented error response (409, 403, 401, 400).
- **FR-003**: The full register → verify-email → login sequence MUST be covered by a single end-to-end integration test in `handler/auth_flow_integration_test.go`.
- **FR-004**: KMP `AuthApiClient` tests MUST use `io.ktor.client.engine.mock.MockEngine` and MUST NOT require a running backend.
- **FR-005**: KMP store tests MUST use a fake `AuthRepository` interface implementation and MUST NOT use real coroutine delays.
- **FR-006**: All tests MUST follow the TDD Red-Green-Refactor cycle: test file is written first and must fail before the subject is added/modified.
- **FR-007**: Go integration tests MUST run a schema migration (creating `auth.*` tables) at test setup using the existing SQL from `infra/migrations/`.
- **FR-008**: Test helpers (DB setup, fake repositories) MUST be in dedicated helper files, not inlined in test cases.

### Key Entities

- **TestDB**: Shared PostgreSQL connection pool used across handler integration tests; set up once per `TestMain`, torn down after all tests.
- **MockAuthRepository**: Kotlin interface implementation with configurable return values; reused across store tests.
- **MockEngine**: Ktor client mock engine with pre-configured response handlers per URL path.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `go test ./handler/... -run Integration` passes 100% in under 30 seconds with a running PostgreSQL.
- **SC-002**: `go test ./handler/... -short` skips all integration tests and runs in under 5 seconds.
- **SC-003**: `./gradlew :shared:desktopTest` passes 100% with no network connectivity.
- **SC-004**: Every auth API endpoint documented in `backend/svc/auth/cmd/server/main.go` has at least one test case.
- **SC-005**: Test coverage for `backend/svc/auth/handler/` reaches ≥ 80% line coverage.

---

## Assumptions

- A PostgreSQL instance accessible at `TEST_DB_URL` is required to run integration tests; `make run-backend` starts one.
- `testcontainers-go` is NOT used — tests rely on an externally running database to keep the test setup simple and consistent with the existing `make run-backend` workflow.
- The `infra/migrations/` SQL files are the canonical schema source; test setup runs them against the test DB.
- KMP mock engine tests run on the JVM target (`desktopTest`) — no Wasm test runner is required.
- `ktor-client-mock` is already in the KMP test dependencies; if not, it must be added to `mobile/shared/build.gradle.kts`.
