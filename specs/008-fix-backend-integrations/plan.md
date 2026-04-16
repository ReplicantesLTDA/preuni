# Implementation Plan: Fix Backend Integration Issues

**Branch**: `008-fix-backend-integrations` | **Date**: 2026-04-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/008-fix-backend-integrations/spec.md`

## Summary

Fix Home (Inicio) and Profile (Perfil) backend data integration issues by standardizing authenticated client bootstrap across platforms, tightening request timeout/retry + schema validation behavior, and adding end-to-end/integration tests that detect contract regressions before merge.

Key outcome: data loading and refresh in Home/Profile should work consistently on Android, iOS, and Web with measurable telemetry and deterministic failure handling.

## Technical Context

**Language/Version**: Kotlin 2.1.20 (KMP shared + Android/iOS/Web entrypoints), Go 1.23 (user-svc integration tests)
**Primary Dependencies**: Compose Multiplatform 1.8.0, Decompose 3.3.0, MVIKotlin 4.2.0, Ktor Client, kotlinx.serialization, go-chi + pgx (existing backend stack)
**Storage**: `SecureStorage` / `TokenStore` for session tokens; PostgreSQL 16 only for backend integration test fixture
**Testing**: KMP `desktopTest` (`kotlin.test`, coroutines test, Ktor `MockEngine`) + Go handler integration tests (`go test`, `httptest`, `TEST_DB_URL`)
**Target Platform**: Android, iOS, Web (Wasm) app clients + user-svc backend contract
**Project Type**: Mobile app + backend service contract hardening
**Performance Goals**: Home/Profile data visible <= 3s on stable network (p98), request timeout capped at 10s for these fetch paths, backend endpoint p95 < 500ms (existing constitution target)
**Constraints**: No new backend endpoints; preserve current screen IA/navigation; retry only transient failures (network + 5xx), redact PII in request/response logs
**Scale/Scope**: ~12-18 Kotlin files + ~2-4 new/updated Kotlin test files + ~2 Go integration test files in user-svc

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Readability / single responsibility | PASS | Shared network bootstrap extracted instead of per-platform copy-paste wiring |
| No dead code | PASS | Deprecated duplicated bootstrap code in iOS/Web entrypoints removed after extraction |
| Test-first for new behavior | PASS | Add failing integration tests for Home/Profile fetch + retry + schema/contract handling before implementation |
| Unit tests for business logic | PASS | Home/Profile retry + error mapping behavior covered in store tests |
| Integration tests for API boundaries | PASS | User-svc `/v1/students/me` integration tests + KMP `UserApiClient` contract tests |
| UX consistency | PASS | Error/loading/retry states follow existing Material3 patterns already used in Home/Profile |
| Performance requirements | PASS | Timeout and request timing instrumentation added for Home/Profile fetch path |
| Coverage floor >= 80% | PASS | New tests are additive and focused on critical regressions |

Post-design re-check: still PASS, no constitution violations introduced.

## Project Structure

### Documentation (this feature)

```text
specs/008-fix-backend-integrations/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── home-profile-api-contract.md
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
mobile/
├── shared/src/commonMain/kotlin/com/preuni/shared/data/network/
│   ├── ApiClient.kt                        # timeout + telemetry/redaction hooks
│   └── NetworkStackFactory.kt              # NEW shared authenticated client bootstrap
├── shared/src/commonMain/kotlin/com/preuni/shared/data/user/
│   └── UserApiClient.kt                    # response validation + Home/Profile request behavior
├── shared/src/commonMain/kotlin/com/preuni/shared/presentation/home/
│   ├── HomeStore.kt                        # transient retry policy + timing capture
│   └── HomeScreen.kt                       # user-facing retry/error refresh behavior
├── shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/
│   ├── ProfileStore.kt                     # profile fetch resilience + retry path
│   └── ProfileScreen.kt                    # error/retry states for profile data
├── androidApp/src/androidMain/kotlin/com/preuni/android/MainActivity.kt
├── shared/src/iosMain/kotlin/com/preuni/shared/MainViewController.kt
├── webApp/src/wasmJsMain/kotlin/com/preuni/web/main.kt
└── shared/src/commonTest/kotlin/com/preuni/shared/
    ├── data/network/NetworkStackFactoryTest.kt      # NEW
    ├── data/user/UserApiClientTest.kt               # extend with integration edge cases
    ├── presentation/home/HomeStoreTest.kt           # extend retry/timing/error scenarios
    └── presentation/profile/ProfileStoreTest.kt     # extend profile fetch/retry scenarios

backend/svc/user/
└── handler/
    ├── testhelper/db.go                             # NEW (mirrors auth test helper pattern)
    ├── get_student_integration_test.go              # NEW
    └── update_student_integration_test.go           # NEW
```

**Structure Decision**: Keep a single KMP shared domain/data/presentation core and add one shared network bootstrap abstraction consumed by all platform entrypoints. Backend changes are test-only in user-svc to validate existing contracts.

## Implementation Reference

### US1 - Home loads backend dashboard/profile summary data reliably

1. Extract common authenticated HTTP bootstrap (token injection + refresh interceptor) into shared code and consume it from Android/iOS/Web entrypoints to remove platform wiring drift.
2. Keep Home source of truth as `GET /v1/students/me` (existing contract), but enforce:
   - per-request timeout aligned to 10 seconds for Home/Profile data fetch path,
   - transient-only exponential backoff (network/5xx),
   - schema validation before mapping to `Student` domain model.
3. Update Home UI messaging to surface actionable retry state instead of generic failure text.

### US2 - Profile loads and updates backend data consistently

4. Apply the same transport + validation contract to Profile load (`LoadProfile`) and update flow responses.
5. Add explicit retry action for profile fetch failures without forcing full app reload.
6. Ensure successful updates immediately rehydrate screen state from backend response object.

### US3 - Network degradation behavior + observability

7. Add request/response telemetry capture in network layer:
   - method, route, status, duration, attempt index, outcome category,
   - no request/response body logging for PII fields,
   - preserve current debugging utility while redacting credentials/tokens.
8. Emit timing metrics consumable by tests and future analytics plumbing.

### US4 - End-to-end/integration test coverage

9. KMP integration tests (`MockEngine`) for Home/Profile API contract:
   - authorization header present,
   - timeout handling,
   - retry behavior across transient failures,
   - malformed/missing required fields rejected.
10. user-svc integration tests for `/v1/students/me` read/update contract:
   - full response payload shape,
   - validation error envelope behavior,
   - auth-required boundary conditions.

## Test Plan

### Mobile (KMP shared)

- `UserApiClientTest`:
  - success payload maps all required fields without loss
  - malformed/missing required field returns `AppError`
  - unauthorized/5xx/network timeout map to expected `AppError` categories
- `HomeStoreTest`:
  - transient failure retries with backoff then succeeds
  - persistent failure stops after max retries and exposes retry action state
  - manual retry re-requests backend data
- `ProfileStoreTest`:
  - profile load success and failure states
  - retry path after timeout/network failure
  - update flow returns refreshed profile fields

### Backend (Go user-svc)

- `TestIntegration_GetStudent_*`:
  - returns `200` with expected JSON schema and values
  - returns `401` when bearer token missing/invalid
- `TestIntegration_UpdateStudent_*`:
  - valid patch returns updated student payload
  - invalid username returns validation envelope (`422`, `error.field = "username"`)

## Complexity Tracking

No constitution violations requiring justification.
