# Tasks: Fix Web App Runtime and Compilation Issues

**Input**: Design documents from `/specs/002-fix-web-compilation/`  
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, quickstart.md ✅

**Organization**: Tasks are grouped by user story. Two user stories, two targeted file changes — no setup phase needed.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel
- **[Story]**: User story (US1, US2)

---

## Phase 1: User Story 1 — Developer sees Login screen at localhost:8080 (Priority: P1) 🎯 MVP

**Goal**: `make run-web` opens the browser and renders the Compose Login screen instead of the webpack debug page.

**Independent Test**: Run `make run-web` → browser opens `http://localhost:8080` → canvas renders Login form (email field, password field, sign-in button).

### Implementation for User Story 1

- [x] T001 [P] [US1] Create `mobile/webApp/src/wasmJsMain/resources/index.html` — minimal HTML entry point that loads `skiko.js` then `preuni-web.js` with full-viewport CSS reset
- [x] T002 [P] [US1] Edit `mobile/webApp/src/wasmJsMain/kotlin/com/preuni/web/main.kt` — add `lifecycle.create()` and `lifecycle.resume()` immediately after `val lifecycle = LifecycleRegistry()` (before `DefaultComponentContext` is constructed)

**Checkpoint**: After T001 + T002, run `make run-web`. The Login screen should render on the canvas and text input events should update the UI (confirms Compose event loop + store executors are active).

---

## Phase 2: User Story 2 — `make run-web` succeeds on machines with JDK 25 as system default (Priority: P2)

**Goal**: Confirm the Makefile's JDK selection logic protects the build from JDK 25 incompatibility.

**Independent Test**: On a machine where `java -version` reports `25.0.2`, running `make run-web` selects JDK 23 and completes without `IllegalArgumentException: 25.0.2`.

**Note**: No code change is required (the Makefile already handles this). This phase is a verification task only.

### Verification for User Story 2

- [x] T003 [US2] Verify that `make run-web` selects JDK 23 — check output for `JAVA_HOME_CANDIDATE` pointing to the JDK 23 path (`/Users/dwbessa/Library/Java/JavaVirtualMachines/openjdk-23.0.2/Contents/Home`), and that the Gradle build does not throw `IllegalArgumentException: 25.0.2`

**Checkpoint**: Build succeeds and JDK warning is absent from Gradle output.

---

## Final Phase: Polish & Cross-Cutting Concerns

**Purpose**: End-to-end validation of all acceptance scenarios from spec.md.

- [ ] T004 **[MANUAL]** Run all quickstart.md verification scenarios: `make run-web` → confirm Login screen renders, typing updates the email field, and submitting the form produces a loading or error state (SC-001 through SC-004)
- [ ] T005 [P] **[MANUAL]** Check browser console for errors after `make run-web` — confirm no missing asset errors (`skiko.js`, `preuni-web.js`, `preuni-web.wasm`), no Wasm instantiation errors, and no `ComposeTarget` errors

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (US1)**: No dependencies — start immediately. T001 and T002 touch different files and can run in parallel.
- **Phase 2 (US2)**: No code dependencies. Can run in parallel with or after Phase 1 since it is pure verification.
- **Final Phase**: Depends on T001 + T002 being complete and `make run-web` running successfully.

### User Story Dependencies

- **US1 (P1)**: Fully independent — two file changes, no dependencies on other stories.
- **US2 (P2)**: Fully independent — verification only, no file changes.

### Within Each Phase

- T001 and T002 are independent (different files) — mark both `[P]`, run in parallel.
- T003 requires T001 + T002 complete (needs the dev server running).
- T004 + T005 require T001 + T002 complete.

---

## Parallel Execution Example: Phase 1

```bash
# T001 and T002 touch different files — run simultaneously:
Task: "Create mobile/webApp/src/wasmJsMain/resources/index.html"
Task: "Edit mobile/webApp/src/wasmJsMain/kotlin/com/preuni/web/main.kt"
```

---

## Implementation Strategy

### MVP (User Story 1 Only)

1. Complete T001 + T002 in parallel
2. Run `make run-web` and validate Login screen
3. **STOP and VALIDATE** — this is the entire fix for the reported issue

### Full Delivery (Both Stories)

1. T001 + T002 in parallel → verify US1
2. T003 → verify US2
3. T004 + T005 → full sign-off

---

## Notes

- Total tasks: 5 (2 implementation, 1 verification, 2 polish)
- Parallel opportunities: T001 ‖ T002, T004 ‖ T005
- No new dependencies, no schema changes, no backend changes
- JDK requirement: `make run-web` only (not bare `./gradlew`)
- Commit after T002 with both files together (they are a single logical fix)
