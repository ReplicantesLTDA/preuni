# Research: Fix Web App Runtime and Compilation Issues

**Branch**: `002-fix-web-compilation` | **Date**: 2026-04-03

---

## Finding 1: Missing `index.html` — root cause of "weird content"

**Decision**: Add `webApp/src/wasmJsMain/resources/index.html`.

**Rationale**:  
The webpack dev server config (at `mobile/build/js/packages/preuni-web/webpack.config.js`) serves static files from `kotlin/` and `../../../../webApp/build/processedResources/wasmJs/main`. The `processedResources` path is populated only from `src/wasmJsMain/resources/`. Since that directory does not exist, no `index.html` is ever placed there. webpack-dev-server, without an HTML entry, shows its own debug/directory page — the "weird content" the user sees.

`CanvasBasedWindow("Preuni")` in Compose Multiplatform creates a `<canvas id="ComposeTarget">` automatically and attaches it to `document.body`, so the HTML does not need to pre-declare the canvas. But it **does** need an HTML document to bootstrap at all.

Required HTML structure for Compose MP 1.8 + Kotlin/Wasm:
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

`skiko.js` is served from the `kotlin/` static directory (built by Compose). `preuni-web.js` is the webpack-bundled entry.

**Alternatives considered**:
- Relying on webpack's `HtmlWebpackPlugin` — not configured and would require adding devDependencies; using a static resource file is the standard Compose MP approach and requires no extra tooling.
- Removing `open: true` from dev server so developer navigates manually — doesn't fix the root cause.

---

## Finding 2: `LifecycleRegistry` never resumed

**Decision**: Call `lifecycle.create()` then `lifecycle.resume()` in `main.kt` before constructing `RootComponent`.

**Rationale**:  
`DefaultComponentContext` accepts a `Lifecycle` to propagate state to child components. In Decompose 3.x, a `LifecycleRegistry` starts in `INITIALIZED` state. If it is never transitioned, the child component tree is created but coroutine scopes are never started:
- `CoroutineExecutor` in MVIKotlin waits for the component lifecycle to reach `STARTED` (or `RESUMED`) before its internal scope is active.
- Without `lifecycle.resume()`, `store.accept(SomeIntent)` dispatches are queued or silently dropped (depending on MVIKotlin version).
- The initial Compose state is rendered correctly (from `initialState`), but all async operations — loading user data, login network calls — never execute.

Decompose 3.x for Wasm does not auto-resume the lifecycle (unlike Android where the Activity handles it). Web lifecycle management is manual. Calling `lifecycle.create()` + `lifecycle.resume()` before entering the Compose rendering loop is the correct pattern.

**Alternatives considered**:
- `lifecycle.attachToDocument()` — available in Decompose's JS/DOM target, NOT available in the Wasm target at this version.
- Starting lifecycle inside `CanvasBasedWindow { }` via `LaunchedEffect` — creates a timing issue where the root component is constructed before its lifecycle is active.

---

## Finding 3: JDK 25 causes `IllegalArgumentException` in Kotlin 2.1.20

**Decision**: No code change needed — the Makefile already handles this correctly.

**Rationale**:  
Kotlin 2.1.20's `JavaVersion.parse()` hard-codes known major versions and throws `IllegalArgumentException` for unknown ones (including 25). The Makefile resolves `JAVA_HOME` to JDK 23, 21, or 17 in order via `/usr/libexec/java_home -v N`. JDK 23 is installed at `/Users/dwbessa/Library/Java/JavaVirtualMachines/openjdk-23.0.2`. Running `make run-web` correctly sets `JAVA_HOME` to JDK 23.

Running Gradle directly without setting `JAVA_HOME` picks up the system default (JDK 25) and fails. **Developers should always use `make run-web`** (or export `JAVA_HOME` manually) — this is documented in `CLAUDE.md`.

**Alternatives considered**:
- Add a Gradle toolchain declaration to pin JDK 17 — avoids the ambient JDK problem entirely but requires all developers to have a toolchain-compatible JDK available via Gradle's auto-provisioning, which adds network dependency.
- Upgrade Kotlin to 2.2+ (when JDK 25 support lands) — out of scope for this fix.

---

## Finding 4: No Kotlin source compilation errors

The `.wasm` and `.mjs` artifacts at `webApp/build/compileSync/wasmJs/main/developmentExecutable/kotlin/` are up-to-date from a previous successful Gradle run with JDK 23. All `expect`/`actual` declarations are fulfilled:
- `SecureStorage` → wasmJs: `sessionStorage`-backed
- `createSqlDriver()` → wasmJs: `WebWorkerDriver` (not yet called by any instantiated component, so worker URL resolution is a future concern)

No compilation issues need to be fixed in this task.
