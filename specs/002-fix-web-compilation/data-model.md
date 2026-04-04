# Data Model: Fix Web App Runtime and Compilation Issues

**Branch**: `002-fix-web-compilation` | **Date**: 2026-04-03

No new data entities are introduced by this fix.

## Affected Files (not a data model — structural change map)

| File | Change | Purpose |
|------|--------|---------|
| `mobile/webApp/src/wasmJsMain/resources/index.html` | **Create** | HTML entry point for webpack dev server |
| `mobile/webApp/src/wasmJsMain/kotlin/com/preuni/web/main.kt` | **Modify** | Add `lifecycle.create()` + `lifecycle.resume()` |

## State Transitions (Lifecycle)

```
Before fix:
  LifecycleRegistry → INITIALIZED (stuck)
  ↓
  ComponentContext propagates INITIALIZED to all children
  ↓
  CoroutineExecutor scopes never start → stores don't process intents

After fix:
  LifecycleRegistry → INITIALIZED → CREATED → RESUMED
  ↓
  ComponentContext propagates RESUMED to children
  ↓
  CoroutineExecutor scopes start → stores process intents → data loads
```
