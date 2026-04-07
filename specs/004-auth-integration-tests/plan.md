# Implementation Plan: Auth Integration & Unit Tests

**Branch**: `004-auth-integration-tests` | **Date**: 2026-04-05 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-auth-integration-tests/spec.md`

## Summary

Write a TDD test suite covering the full auth layer end-to-end: Go handler integration tests against a real PostgreSQL instance (httptest + TEST_DB_URL), KMP `AuthApiClient` contract tests using ktor `MockEngine`, and KMP `RegisterStore`/`VerifyEmailStore` unit tests with a fake `AuthRepository`. Tests are written first (Red), then made green by confirming or adjusting existing implementation.

## Technical Context

**Language/Version**: Go 1.24 (backend tests) + Kotlin 2.1.20 / KMP (mobile tests)
**Primary Dependencies**:
- Go: `net/http/httptest`, `github.com/stretchr/testify v1.10.0`, `github.com/jackc/pgx/v5`, `github.com/google/uuid`
- KMP: `io.ktor.client.engine.mock.MockEngine`, `kotlinx.coroutines.test`, `kotlin.test`, `com.arkivanov.mvikotlin.main.store.DefaultStoreFactory`
**Storage**: PostgreSQL 16 (integration tests via `TEST_DB_URL`); N/A for mock tests
**Testing**: `go test ./handler/...`, `./gradlew :shared:desktopTest`
**Target Platform**: Linux/macOS (CI + local); tests run on JVM target for KMP
**Project Type**: test suite for existing microservice + KMP shared module
**Performance Goals**: Integration suite under 30s; unit suite under 5s
**Constraints**: No testcontainers-go; no real network in KMP tests; TDD (tests written before subject)
**Scale/Scope**: ~6 Go handler test cases + 5 AuthApiClient mock cases + 5 store unit test cases = ~16 test functions total across 5 new test files

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| Code Quality — no raw values | ✓ PASS | SQL loaded from `infra/migrations/`; test helpers in dedicated files |
| Code Quality — single responsibility | ✓ PASS | Each test file covers one subject |
| Testing — TDD | ✓ PASS | Tests written first, must fail before implementation confirmed |
| Testing — DB isolation | ✓ PASS | TestMain creates schema; each test cleans up its own rows |
| No new dependencies | ✓ PASS | testify + ktor-mock already in go.sum / build.gradle.kts |

## Project Structure

### Documentation (this feature)

```text
specs/004-auth-integration-tests/
├── plan.md              # This file
├── spec.md              # Feature specification
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code

```text
backend/svc/auth/
├── handler/
│   ├── testhelper/
│   │   └── db.go                           # NEW — TestDB setup/teardown + schema migration
│   ├── auth_flow_integration_test.go        # NEW [US1] — full register→verify→login E2E
│   ├── register_integration_test.go         # NEW [US1] — POST /v1/auth/register cases
│   ├── login_integration_test.go            # NEW [US1] — POST /v1/auth/login cases
│   ├── verify_email_integration_test.go     # NEW [US1] — POST /v1/auth/email/verify cases
│   └── logout_integration_test.go           # NEW [US1] — POST /v1/auth/logout cases
│
mobile/shared/src/
├── commonTest/kotlin/com/preuni/shared/
│   ├── data/auth/
│   │   └── AuthApiClientTest.kt            # NEW [US2] — MockEngine contract tests
│   └── presentation/auth/
│       ├── RegisterStoreTest.kt            # NEW [US3] — RegisterStore unit tests
│       └── VerifyEmailStoreTest.kt         # NEW [US3] — VerifyEmailStore unit tests
```

**Structure Decision**: Test files co-located with their subjects per existing project convention (see `LoginStoreTest.kt`, `password_test.go`). Go test helpers extracted to `handler/testhelper/` package per FR-008.
