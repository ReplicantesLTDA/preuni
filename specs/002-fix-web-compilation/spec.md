# Feature Specification: Fix Web App Runtime and Compilation Issues

**Feature Branch**: `002-fix-web-compilation`  
**Created**: 2026-04-03  
**Status**: Draft  
**Input**: Developer cannot verify the web app works; `make run-web` opens localhost:8080 with unexpected content instead of the application UI.

## User Scenarios & Testing

### User Story 1 - Developer can run the web app and see the Login screen (Priority: P1)

As a developer, when I run `make run-web`, I want the browser to open localhost:8080 and display the Preuni app (starting at the Login screen) rendered inside a Compose canvas, so I can verify the UI and behavior.

**Why this priority**: Without this, no feature work on the web target can be verified. It is the baseline for all other web QA.

**Independent Test**: Run `make run-web`, confirm browser shows a canvas-based login form with email/password fields and a submit button.

**Acceptance Scenarios**:

1. **Given** no session token in sessionStorage, **When** `make run-web` is executed and the browser opens, **Then** the Login screen is rendered on the canvas at localhost:8080.
2. **Given** the Login screen is visible, **When** the user types into the email field, **Then** the field updates (confirming the app's event loop is active).
3. **Given** the Login screen is visible, **When** the user submits the form, **Then** the LoginStore dispatches an intent and shows a loading or error state (confirming store coroutines are running).

---

### User Story 2 - Developer can rebuild the project without JDK errors (Priority: P2)

As a developer, when I run `make run-web` on a machine with both JDK 23 and JDK 25, the build uses JDK 23 and succeeds without `IllegalArgumentException: 25.0.2`.

**Why this priority**: Builds failing silently due to JDK resolution block other developers.

**Independent Test**: Check that `make run-web` prints no JDK version parse error and the Gradle build completes.

**Acceptance Scenarios**:

1. **Given** JDK 25 is the system default but JDK 23 is installed, **When** `make run-web` runs, **Then** Gradle uses JDK 23 (as selected by the Makefile) and the build succeeds.

---

### Edge Cases

- What happens when the browser tab is closed while the webpack dev server runs? (Hot reload / reconnect handled by webpack-dev-server, out of scope.)
- What if sessionStorage already has a stale token? The app routes to Onboarding/Main instead of Login — this is expected behavior, not a bug.

## Requirements

### Functional Requirements

- **FR-001**: The `webApp` module MUST include an `index.html` resource at `src/wasmJsMain/resources/index.html` that loads the Compose Wasm bundle and sets a full-viewport canvas.
- **FR-002**: `main.kt` MUST call `lifecycle.create()` and `lifecycle.resume()` before constructing `RootComponent`, ensuring Decompose component coroutines and MVIKotlin store executors are active.
- **FR-003**: The build MUST succeed with JDK 23 (via Makefile's `JAVA_HOME` selection).
- **FR-004**: The app MUST render the Login screen when no session token exists in sessionStorage.
- **FR-005**: The Login form MUST accept user input (confirming the Compose event loop is running).

### Key Entities

- **`index.html`**: Static HTML entry point served by webpack dev server; sets viewport, loads `preuni-web.js`.
- **`LifecycleRegistry`**: Decompose lifecycle object in `main.kt`; MUST be transitioned to RESUMED state for component coroutines to run.

## Success Criteria

### Measurable Outcomes

- **SC-001**: `make run-web` completes without errors and browser opens to a canvas-rendered Preuni Login screen.
- **SC-002**: Typing in the email field on the Login screen triggers visible state changes in the UI (Compose event loop is active).
- **SC-003**: Submitting the Login form dispatches to LoginStore and produces a loading or error state (store executor is running).
- **SC-004**: No browser console errors about missing Wasm module, failed asset loads, or missing `ComposeTarget`.

## Assumptions

- Developer has JDK 23 installed (confirmed: `/Users/dwbessa/Library/Java/JavaVirtualMachines/openjdk-23.0.2`).
- Node.js 20+ is installed for webpack dev server.
- No backend services need to be running to verify the web UI renders correctly (network calls will fail gracefully and show error states).
- `CanvasBasedWindow` in Compose Multiplatform 1.8 / Kotlin 2.1.20 uses `#ComposeTarget` canvas ID.
