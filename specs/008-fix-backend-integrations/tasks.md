# Tasks: Fix Backend Integration Issues

**Input**: Design documents from `specs/008-fix-backend-integrations/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Included (explicitly requested in spec via "with e2e tests" and SC-004).

**Organization**: Tasks are grouped by user story so each story can be implemented and validated independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency conflicts)
- **[Story]**: User story label (`US1`, `US2`, `US3`)
- Every task includes a concrete file path

---

## Phase 1: Setup

**Purpose**: Establish a clean baseline before modifications.

- [X] T001 Run baseline mobile tests for `mobile/shared/src/commonTest/kotlin/com/preuni/shared/presentation/home/HomeStoreTest.kt` and `mobile/shared/src/commonTest/kotlin/com/preuni/shared/presentation/profile/ProfileStoreTest.kt`
- [X] T002 Run baseline backend handler tests in `backend/svc/user/handler/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared network/bootstrap and backend test infrastructure required by all stories.

**⚠️ CRITICAL**: User story tasks begin only after this phase is complete.

- [X] T003 [P] Create reusable integration DB helper in `backend/svc/user/handler/testhelper/db.go`
- [X] T004 Write failing bootstrap coverage tests in `mobile/shared/src/commonTest/kotlin/com/preuni/shared/data/network/NetworkStackFactoryTest.kt`
- [X] T005 Implement shared authenticated client bootstrap in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/data/network/NetworkStackFactory.kt`
- [X] T006 [P] Migrate Android bootstrap to shared factory in `mobile/androidApp/src/androidMain/kotlin/com/preuni/android/MainActivity.kt`
- [X] T007 [P] Migrate iOS bootstrap to shared factory in `mobile/shared/src/iosMain/kotlin/com/preuni/shared/MainViewController.kt`
- [X] T008 [P] Migrate Web bootstrap to shared factory in `mobile/webApp/src/wasmJsMain/kotlin/com/preuni/web/main.kt`

**Checkpoint**: Shared bootstrap and test infrastructure are in place.

---

## Phase 3: User Story 1 - Home section loads backend data (Priority: P1) 🎯 MVP

**Goal**: Authenticated users consistently see Home dashboard/profile summary data and can refresh it.

**Independent Test**: Open Home on a fresh authenticated session, verify data loads in <=3s, then trigger refresh and confirm fresh data fetch.

### Tests for User Story 1 (write first, confirm failing)

- [X] T009 [P] [US1] Add failing GET `/v1/students/me` auth header + schema tests in `mobile/shared/src/commonTest/kotlin/com/preuni/shared/data/user/UserApiClientTest.kt`
- [X] T010 [P] [US1] Add failing Home load/retry/refresh tests in `mobile/shared/src/commonTest/kotlin/com/preuni/shared/presentation/home/HomeStoreTest.kt`
- [X] T011 [P] [US1] Add failing GET `/v1/students/me` integration tests in `backend/svc/user/handler/get_student_integration_test.go`

### Implementation for User Story 1

- [X] T012 [US1] Implement strict `StudentResponse` validation + mapping for `getMe()` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/data/user/UserApiClient.kt`
- [X] T013 [US1] Implement 10s timeout + transient exponential retry for Home fetch flow in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/home/HomeStore.kt`
- [X] T014 [US1] Implement Home refresh affordance and user-friendly error copy in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/home/HomeScreen.kt`
- [X] T015 [US1] Wire Home refresh intent handling in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/PreuniApp.kt`

**Checkpoint**: US1 is independently functional and testable.

---

## Phase 4: User Story 2 - Profile section loads and updates backend data (Priority: P1)

**Goal**: Authenticated users consistently see complete profile information and receive validated update responses.

**Independent Test**: Open Profile, verify profile payload renders in <=3s, update profile data, and confirm new values appear on subsequent fetch.

### Tests for User Story 2 (write first, confirm failing)

- [X] T016 [P] [US2] Add failing Profile load/retry/update tests in `mobile/shared/src/commonTest/kotlin/com/preuni/shared/presentation/profile/ProfileStoreTest.kt`
- [X] T017 [P] [US2] Add failing PATCH `/v1/students/me` integration tests in `backend/svc/user/handler/update_student_integration_test.go`
- [X] T018 [US2] Add failing profile-update schema contract tests in `mobile/shared/src/commonTest/kotlin/com/preuni/shared/data/user/UserApiClientTest.kt`

### Implementation for User Story 2

- [X] T019 [US2] Reuse strict `StudentResponse` validation for `updateProfile()` in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/data/user/UserApiClient.kt`
- [X] T020 [US2] Implement Profile fetch timeout/retry/refresh flow in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileStore.kt`
- [X] T021 [US2] Implement Profile refresh/retry UX and friendly error states in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileScreen.kt`
- [X] T022 [US2] Wire Profile refresh callbacks and labels in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/PreuniApp.kt`

**Checkpoint**: US2 is independently functional and testable.

---

## Phase 5: User Story 3 - Resilience under poor network + observability (Priority: P2)

**Goal**: Home/Profile handle latency and transient failures gracefully, while logging timing/attempt telemetry with PII redaction.

**Independent Test**: Simulate slow/failing network, verify timeout at 10s, retry only on transient failures, and confirm retry action works without full reload.

### Tests for User Story 3 (write first, confirm failing)

- [X] T023 [P] [US3] Add failing telemetry + PII-redaction tests in `mobile/shared/src/commonTest/kotlin/com/preuni/shared/data/network/ApiClientTelemetryTest.kt`
- [X] T024 [P] [US3] Add failing transient-vs-non-transient retry classification tests in `mobile/shared/src/commonTest/kotlin/com/preuni/shared/presentation/home/HomeStoreTest.kt`
- [X] T025 [P] [US3] Add failing transient-vs-non-transient retry classification tests in `mobile/shared/src/commonTest/kotlin/com/preuni/shared/presentation/profile/ProfileStoreTest.kt`

### Implementation for User Story 3

- [X] T026 [US3] Add structured request telemetry and redacted logging hooks in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/data/network/ApiClient.kt`
- [X] T027 [US3] Create shared retry classification helper in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/data/network/RetryPolicy.kt`
- [X] T028 [US3] Apply `RetryPolicy` to Home request flow in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/home/HomeStore.kt`
- [X] T029 [US3] Apply `RetryPolicy` to Profile request flow in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileStore.kt`
- [X] T030 [US3] Update resilience acceptance matrix and telemetry contract in `specs/008-fix-backend-integrations/contracts/home-profile-api-contract.md`
- [X] T031 [US3] Update degraded-network verification steps in `specs/008-fix-backend-integrations/quickstart.md`

**Checkpoint**: US3 is independently functional and testable.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, cleanup, and regression hardening across stories.

- [X] T032 [P] Run KMP feature suites for `mobile/shared/src/commonTest/kotlin/com/preuni/shared/data/user/UserApiClientTest.kt`, `mobile/shared/src/commonTest/kotlin/com/preuni/shared/presentation/home/HomeStoreTest.kt`, `mobile/shared/src/commonTest/kotlin/com/preuni/shared/presentation/profile/ProfileStoreTest.kt`, and `mobile/shared/src/commonTest/kotlin/com/preuni/shared/data/network/ApiClientTelemetryTest.kt`
- [X] T033 [P] Run backend integration suites for `backend/svc/user/handler/get_student_integration_test.go` and `backend/svc/user/handler/update_student_integration_test.go`
- [X] T034 Resolve final regressions in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/data/user/UserApiClient.kt`, `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/home/HomeStore.kt`, and `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/profile/ProfileStore.kt`
- [X] T035 [P] Remove any sensitive/debug logging leftovers in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/data/network/ApiClient.kt` and `mobile/shared/src/commonMain/kotlin/com/preuni/shared/data/network/NetworkStackFactory.kt`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies.
- **Phase 2 (Foundational)**: Depends on Phase 1, blocks all user stories.
- **Phase 3 (US1)**: Depends on Phase 2.
- **Phase 4 (US2)**: Depends on Phase 2.
- **Phase 5 (US3)**: Depends on Phase 2 and integrates behavior across US1 + US2 flows.
- **Phase 6 (Polish)**: Depends on all user story phases.

### User Story Dependencies

- **US1 (P1)**: Starts after Foundation; no dependency on US2.
- **US2 (P1)**: Starts after Foundation; no dependency on US1.
- **US3 (P2)**: Starts after Foundation, then validates resilience across both US1 and US2 surfaces.

### Within Each User Story

- Tests first and failing before implementation.
- API/contract handling before store wiring.
- Store wiring before screen behavior.
- Story-specific checkpoint before moving to next phase.

---

## Parallel Opportunities

- **Foundational**: `T003`, `T006`, `T007`, `T008`
- **US1**: `T009`, `T010`, `T011`
- **US2**: `T016`, `T017`
- **US3**: `T023`, `T024`, `T025`
- **Polish**: `T032`, `T033`, `T035`

---

## Parallel Example: User Story 1

```bash
# Run US1 test authoring in parallel:
Task: "T009 [US1] UserApiClient GET /v1/students/me tests in .../UserApiClientTest.kt"
Task: "T010 [US1] Home retry/refresh tests in .../HomeStoreTest.kt"
Task: "T011 [US1] Backend GET /v1/students/me integration tests in .../get_student_integration_test.go"
```

## Parallel Example: User Story 2

```bash
# Run US2 test authoring in parallel:
Task: "T016 [US2] Profile store tests in .../ProfileStoreTest.kt"
Task: "T017 [US2] Backend PATCH /v1/students/me integration tests in .../update_student_integration_test.go"
```

## Parallel Example: User Story 3

```bash
# Run US3 test authoring in parallel:
Task: "T023 [US3] ApiClient telemetry/redaction tests in .../ApiClientTelemetryTest.kt"
Task: "T024 [US3] Home retry classification tests in .../HomeStoreTest.kt"
Task: "T025 [US3] Profile retry classification tests in .../ProfileStoreTest.kt"
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 (US1).
3. Validate US1 independently (Home load + refresh + contract tests).
4. Demo/deploy MVP.

### Incremental Delivery

1. Foundation complete (Phases 1-2).
2. Deliver US1 (Home integration).
3. Deliver US2 (Profile integration).
4. Deliver US3 (network resilience + telemetry).
5. Finish with Phase 6 polish.

### Parallel Team Strategy

1. Team completes Setup + Foundational together.
2. Then split:
   - Dev A: US1
   - Dev B: US2
   - Dev C: US3 test scaffolding
3. Rejoin for cross-cutting polish and regression validation.

---

## Notes

- `[P]` tasks are safe parallel candidates because they target different files and concerns.
- Keep task updates in this file as progress markers (`[ ]` -> `[X]`).
- Preserve strict error typing (`AppError`) and avoid broad catch/silent fallback behavior.
