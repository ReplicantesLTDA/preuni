# Implementation Plan: Fix Web App Runtime and Compilation Issues

**Branch**: `002-fix-web-compilation` | **Date**: 2026-04-03 | **Spec**: `spec.md`  
**Input**: Feature specification from `/specs/002-fix-web-compilation/spec.md`

## Summary

The web app compiles successfully but renders a webpack debug page instead of the Compose UI, because (1) `webApp/src/wasmJsMain/resources/index.html` is missing, so webpack dev server has no HTML entry point, and (2) `LifecycleRegistry` in `main.kt` is never transitioned to RESUMED, so Decompose component coroutines and MVIKotlin store executors never activate. No Kotlin compilation errors exist. Fix is two targeted file changes.

## Technical Context

**Language/Version**: Kotlin 2.1.20 + Compose Multiplatform 1.8.0  
**Primary Dependencies**: Decompose 3.3.0, MVIKotlin 4.2.0, Compose canvas rendering (Skiko), webpack 5  
**Storage**: sessionStorage (web), SQLDelight/WebWorkerDriver (not yet activated)  
**Testing**: `./gradlew :shared:desktopTest` (unit); manual browser test for runtime  
**Target Platform**: Kotlin/Wasm (browser), webpack-dev-server 4.x  
**Project Type**: Compose Multiplatform web app  
**Performance Goals**: First-render under 3 s on dev build  
**Constraints**: JDK 23 required (JDK 25 not supported by Kotlin 2.1.20)  
**Scale/Scope**: 2 file changes

## Constitution Check

| Gate | Status | Notes |
|------|--------|-------|
| CI green (lint, type-check, unit tests) | PASS (after fix) | No new business logic; existing tests unaffected |
| Coverage not decreased | PASS | No logic changes |
| Lighthouse CI | N/A for dev fix | Dev build only |
| Bundle size budget | N/A | No new dependencies |
| Design system tokens | N/A | No UI changes |
| Accessibility | N/A | No UI changes |
| Mobile layout | N/A | Fix is structural, not visual |

## Project Structure

### Documentation (this feature)

```text
specs/002-fix-web-compilation/
├── plan.md              ✅ This file
├── research.md          ✅ Phase 0 output
├── data-model.md        ✅ Phase 1 output
└── tasks.md             (Phase 2 — /speckit.tasks command)
```

### Source Code changes

```text
mobile/webApp/src/wasmJsMain/
├── kotlin/com/preuni/web/main.kt          ← add lifecycle.create() + lifecycle.resume()
└── resources/
    └── index.html                         ← CREATE (missing entry point)
```

## Implementation Steps

### Step 1 — Create `index.html`

Create `mobile/webApp/src/wasmJsMain/resources/index.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Preuni</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        html, body { width: 100%; height: 100%; overflow: hidden; }
        canvas { display: block; }
    </style>
</head>
<body>
    <script src="skiko.js"></script>
    <script src="preuni-web.js"></script>
</body>
</html>
```

Why `skiko.js` first: Skiko provides the Wasm/canvas rendering layer that Compose builds on; it must be loaded before the app script instantiates any Compose composables.

### Step 2 — Fix lifecycle in `main.kt`

In `mobile/webApp/src/wasmJsMain/kotlin/com/preuni/web/main.kt`, after `val lifecycle = LifecycleRegistry()`, add:

```kotlin
lifecycle.create()
lifecycle.resume()
```

This transitions the lifecycle through INITIALIZED → CREATED → RESUMED before `DefaultComponentContext` is built, ensuring that all child components inherit the RESUMED state and their coroutine scopes start immediately.

### Step 3 — Verify

```bash
make run-web
```

Expected:
- Browser opens to `http://localhost:8080/`
- Login screen renders (email field, password field, sign-in button)
- Typing into the email field updates the UI
- Clicking Sign In shows a loading spinner / network error (confirming store executor is active)
- No browser console errors about missing assets

## Complexity Tracking

No constitution violations. This is a two-line code fix and one new static file.
