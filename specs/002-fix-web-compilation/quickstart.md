# Quickstart: Fix Web App Runtime

**Branch**: `002-fix-web-compilation`

## What's broken and why

| Symptom | Root cause |
|---------|-----------|
| localhost:8080 shows webpack debug page | `index.html` missing from `webApp/src/wasmJsMain/resources/` |
| UI renders but interactions don't work | `LifecycleRegistry` never called `.create()` / `.resume()` |

## Files to change

1. **CREATE** `mobile/webApp/src/wasmJsMain/resources/index.html` — minimal HTML entry point
2. **EDIT** `mobile/webApp/src/wasmJsMain/kotlin/com/preuni/web/main.kt` — add `lifecycle.create(); lifecycle.resume()` after line 3 of `main()`

## Verify it works

```bash
make run-web
# Browser should open http://localhost:8080 showing the Login screen
```

## JDK requirement

Must use JDK 17, 21, or 23 — JDK 25 is not supported by Kotlin 2.1.20. `make run-web` selects the right JDK automatically. Do NOT run `./gradlew` directly without setting `JAVA_HOME`.
