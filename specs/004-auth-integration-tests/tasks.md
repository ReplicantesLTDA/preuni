# Tasks: Auth Integration & Unit Tests

**Input**: Design documents from `/specs/004-auth-integration-tests/`
**Prerequisites**: plan.md ✓, spec.md ✓

**TDD Approach**: For every phase, test files are written **first** and must produce compile/runtime failures before running the subject. Mark each test task done before its GREEN verification task.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add `ktor-client-mock` to the KMP project — it is not currently in the version catalog or test dependencies.

- [X] T001 Add `ktor-client-mock` alias to `mobile/gradle/libs.versions.toml` under `[libraries]`: `ktor-client-mock = { module = "io.ktor:ktor-client-mock", version.ref = "ktor" }`
- [X] T002 Add `implementation(libs.ktor.client.mock)` to the `commonTest.dependencies` block in `mobile/shared/build.gradle.kts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Go test helper that all US1 handler tests depend on. Must be complete before any Go integration test file can compile.

**⚠️ CRITICAL**: US1 test files cannot compile without this package.

- [X] T003 Create `backend/svc/auth/handler/testhelper/db.go` — package `testhelper`; export `SetupTestDB(t *testing.T) *pgxpool.Pool` that (a) skips via `t.Skip()` when `TEST_DB_URL` is unset, (b) opens a `pgxpool.New` connection, (c) reads and executes each SQL file from `infra/migrations/auth/` in filename order to create the `auth` schema tables, (d) registers `t.Cleanup` to truncate all auth tables after the test; also export `MustEnv(t *testing.T, key string) string` helper.

**Checkpoint**: `go build ./handler/testhelper/...` passes — US1 work can begin.

---

## Phase 3: User Story 1 — Backend Handler Integration Tests (Priority: P1) 🎯 MVP

**Goal**: Every public auth endpoint has at least one passing integration test against a real PostgreSQL instance; the full register → verify-email → login flow is covered by a single E2E test.

**Independent Test**: `go test ./handler/... -run Integration -v` (requires `TEST_DB_URL`)

### Tests for User Story 1 — Write FIRST, confirm RED

> **TDD**: Write all test files below, then run `go test ./handler/... -run Integration`. Tests must compile and either skip (no TEST_DB_URL) or fail before confirming GREEN in T009.

- [X] T004 [US1] Write `backend/svc/auth/handler/auth_flow_integration_test.go` — `func TestIntegration_FullAuthFlow(t *testing.T)`: calls `testhelper.SetupTestDB`; builds `RegisterHandler`, `VerifyEmailHandler`, `LoginHandler` with real repos; uses `httptest.NewRecorder` for each step; asserts register→201, extract OTP from DB, verify→204, login→200 with tokens
- [X] T005 [P] [US1] Write `backend/svc/auth/handler/register_integration_test.go` — two test functions: `TestIntegration_Register_HappyPath` (201 + AuthResponse fields present) and `TestIntegration_Register_DuplicateEmail` (second call → 409)
- [X] T006 [P] [US1] Write `backend/svc/auth/handler/login_integration_test.go` — three test functions: `TestIntegration_Login_UnverifiedEmail` (→ 403), `TestIntegration_Login_InvalidPassword` (→ 401), `TestIntegration_Login_Success` (verified account → 200 + tokens)
- [X] T007 [P] [US1] Write `backend/svc/auth/handler/verify_email_integration_test.go` — two test functions: `TestIntegration_VerifyEmail_ValidOTP` (→ 204), `TestIntegration_VerifyEmail_InvalidOTP` (wrong code → 400)
- [X] T008 [P] [US1] Write `backend/svc/auth/handler/logout_integration_test.go` — two test functions: `TestIntegration_Logout_ValidToken` (→ 204 and token revoked), `TestIntegration_Logout_UnknownToken` (idempotent → 204)

### Confirm GREEN

- [X] T009 [US1] Run `go test ./handler/... -run Integration -v` in `backend/svc/auth` (with `TEST_DB_URL` set) — all 8+ integration test functions must PASS; fix any handler or repository bugs until green

**Checkpoint**: US1 complete — backend auth contract fully verified against real DB.

---

## Phase 4: User Story 2 — Mobile AuthApiClient Contract Tests (Priority: P2)

**Goal**: `AuthApiClient` serializes requests correctly and maps every response/error shape to the right `Result` type, verified without a real backend.

**Independent Test**: `./gradlew :shared:desktopTest --tests "*AuthApiClientTest*"`

### Tests for User Story 2 — Write FIRST, confirm RED

> **TDD**: Write `AuthApiClientTest.kt` first. Run `./gradlew :shared:desktopTest --tests "*AuthApiClientTest*"`. It must fail (class not found or assertion failures) before T011.

- [X] T010 [US2] Write `mobile/shared/src/commonTest/kotlin/com/preuni/shared/data/auth/AuthApiClientTest.kt` — build `AuthApiClient` with `HttpClient(MockEngine { ... })`; five `@Test` functions: (1) `register_success` — mock returns 201 `{"student_id":"x","access_token":"a","refresh_token":"r","expires_in":3600}`, assert `Result.isSuccess` and `AuthSession.userId == "x"`; (2) `register_conflict` — mock returns 409, assert `Result.isFailure` and error `is AppError.Conflict`; (3) `login_forbidden` — mock returns 403, assert error `is AppError.Forbidden`; (4) `verifyEmail_success` — mock returns 204, assert `Result.isSuccess`; (5) `login_unauthorized` — mock returns 401, assert error `is AppError.Unauthorized`

### Confirm GREEN

- [X] T011 [US2] Run `./gradlew :shared:desktopTest --tests "*AuthApiClientTest*"` — all 5 tests must PASS; fix `AuthApiClient` or `toAppError()` mapping if any assertion fails

**Checkpoint**: US2 complete — client serialization and error mapping verified.

---

## Phase 5: User Story 3 — Mobile Store Unit Tests (Priority: P3)

**Goal**: `RegisterStore` and `VerifyEmailStore` intent→state transitions and label emissions are verified with fake collaborators, no coroutine delays, no real network.

**Independent Test**: `./gradlew :shared:desktopTest --tests "*RegisterStoreTest*" "*VerifyEmailStoreTest*"`

### Tests for User Story 3 — Write FIRST, confirm RED

> **TDD**: Write both test files. Run the independent test command. Tests must fail before T014.

- [X] T012 [P] [US3] Write `mobile/shared/src/commonTest/kotlin/com/preuni/shared/presentation/auth/RegisterStoreTest.kt` — fake `AuthRepository` with configurable `register()` result; three `@Test` functions using `runTest`: (1) `submit_validFields_emitsRegisteredLabel` — set email/password/confirmPassword/displayName, collect first label, assert `is RegisterStore.Label.Registered`; (2) `submit_emptyEmail_setsValidationError_withoutCallingRepo` — leave email empty, submit, assert `state.emailError != null` and repo never called; (3) `submit_repositoryReturnsConflict_setsGlobalError` — repo returns `Result.failure(AppError.Conflict("email taken"))`, submit valid fields, assert `state.globalError is AppError.Conflict` and `state.isLoading == false`
- [X] T013 [P] [US3] Write `mobile/shared/src/commonTest/kotlin/com/preuni/shared/presentation/auth/VerifyEmailStoreTest.kt` — create `VerifyEmailStoreFactory` with lambda-based `onSubmit`/`onResend`; two `@Test` functions using `runTest`: (1) `submit_validCode_setsVerifiedTrue` — `onSubmit` returns `Result.success(Unit)`, dispatch `UpdateCode("123456")` + `Submit`, assert `state.verified == true` and `state.isLoading == false`; (2) `submit_repositoryError_setsErrorMessage` — `onSubmit` returns `Result.failure(AppError.Validation("otp","invalid"))`, submit code, assert `state.error != null` and `state.isLoading == false`

### Confirm GREEN

- [X] T014 [US3] Run `./gradlew :shared:desktopTest --tests "*RegisterStoreTest*" "*VerifyEmailStoreTest*"` — all 5 tests must PASS; fix store logic if any test fails

**Checkpoint**: US3 complete — all auth store state machines verified.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T015 Run `go test ./handler/... -short` in `backend/svc/auth` — confirm all integration tests are **skipped** (no output from `TestIntegration_*`), suite completes in under 5 seconds (SC-002)
- [X] T016 Run `./gradlew :shared:desktopTest` — full KMP test suite passes 100% with no network (SC-003); confirm `AuthApiClientTest`, `RegisterStoreTest`, `VerifyEmailStoreTest` all appear in the test report

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — do T001 + T002 immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (build.gradle.kts must be synced before KMP work)
- **US1 — Phase 3**: Depends on Phase 2 (testhelper must exist); Go tests are independent of KMP
- **US2 — Phase 4**: Depends on Phase 1 (ktor-mock in classpath); independent of US1 and US3
- **US3 — Phase 5**: Depends on Phase 1; independent of US1 and US2
- **Polish — Phase 6**: Depends on all story phases complete

### User Story Dependencies

- **US1 (P1)**: Requires testhelper (T003) — no dependency on US2/US3
- **US2 (P2)**: Requires ktor-client-mock (T001, T002) — no dependency on US1/US3
- **US3 (P3)**: Requires ktor-client-mock setup for project sync only — no dependency on US1/US2

### Within Each User Story

1. Test files written and confirmed to **fail/skip** (RED)
2. Subject confirmed or fixed until tests **pass** (GREEN)
3. Checkpoint validation
4. Move to next story

### Parallel Opportunities

- T001 + T002: Sequential (T002 depends on T001)
- T005, T006, T007, T008: All parallel (different files, same phase)
- T012, T013: Parallel (different files)
- US2 and US3 can be worked in parallel once Phase 1 is done

---

## Parallel Example: User Story 1 (Phase 3)

```bash
# After T003 is done, write all handler test files in parallel:
Task A: Write register_integration_test.go     (T005)
Task B: Write login_integration_test.go        (T006)
Task C: Write verify_email_integration_test.go (T007)
Task D: Write logout_integration_test.go       (T008)
# T004 (E2E flow) written sequentially — depends on understanding all handlers

# Then run once all files are written:
cd backend/svc/auth && TEST_DB_URL=... go test ./handler/... -run Integration -v
```

---

## Implementation Strategy

### MVP (US1 Only)

1. Complete Phase 1 (add ktor-mock — quick, unblocks all KMP work)
2. Complete Phase 2 (testhelper — unblocks Go integration tests)
3. Complete Phase 3 (Go handler integration tests)
4. **STOP**: validate with `go test ./handler/... -run Integration`
5. Backend auth contract fully verified — deliverable MVP

### Full Delivery

1. MVP above
2. Phase 4 (AuthApiClient mock tests) — independent, can be done in parallel with Phase 3
3. Phase 5 (Store unit tests) — independent
4. Phase 6 (polish + coverage check)

---

## Notes

- `[P]` = task targets different file from others in same phase — safe to run in parallel
- TDD enforced: every story has a "write test → confirm RED → fix → confirm GREEN" cycle
- Go integration tests require `TEST_DB_URL` env var; they skip gracefully when absent (FR-001)
- KMP tests run on JVM target (`desktopTest`) — no Wasm runner needed (Assumption from spec)
- `ktor-client-mock` version is pinned to the existing `ktor = "3.0.3"` in `libs.versions.toml`
- `VerifyEmailStore` uses lambda injection (`onSubmit`, `onResend`) — no `AuthRepository` fake needed for US3
