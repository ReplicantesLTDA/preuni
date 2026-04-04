# Implementation Plan: preuni.com.br – ENEM Prep Platform

**Branch**: `001-enem-prep-platform` | **Date**: 2026-04-03 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/001-enem-prep-platform/spec.md`

## Summary

Build a Duolingo-inspired, mobile-first ENEM preparation platform targeting Brazilian students seeking higher education access. The system uses FSRS v6 for spaced repetition, delivers bite-sized subject lessons across all five ENEM areas, supports full ENEM mock simulations with dissertation, and generates AI-powered feedback aligned to the five official ENEM competency criteria.

**Frontend**: Kotlin Multiplatform + Compose Multiplatform (Android/iOS/Web) with Decompose navigation and MVIKotlin state management.
**Backend**: Go microservices (REST, NGINX gateway) + Elixir mailing service (Phoenix + Swoosh).

---

## Technical Context

**Language/Version**:
- Frontend: Kotlin 2.x / Compose Multiplatform 1.8+
- Backend (primary): Go 1.23+
- Mail service: Elixir 1.17+ / Phoenix 1.7+

**Primary Dependencies**:
- KMP: Compose Multiplatform, Decompose, MVIKotlin, Ktor Client, SQLDelight, Koin
- Go services: chi (router), pgx v5 (PostgreSQL), zap (logging), golang-migrate, golang-jwt
- go-fsrs v3 (`github.com/open-spaced-repetition/go-fsrs/v3`)
- FSRS Kotlin (`github.com/open-spaced-repetition/FSRS-Kotlin`) — for offline FSRS rating in shared module
- Dissertation AI: Anthropic Claude API (claude-haiku-4-5-20251001 primary, claude-sonnet-4-6 fallback)
- Elixir: Phoenix 1.7, Swoosh, ExUnit

**Storage**:
- PostgreSQL 16 (primary; one schema per service, single instance in v1)
- Redis 7 (JWT refresh token store, rate-limit counters)
- S3-compatible object storage (avatars, essay support images)

**Testing**:
- Go: `testing` stdlib + testify + testcontainers-go (real PostgreSQL in integration tests)
- Elixir: ExUnit
- KMP: kotlin.test (commonMain), Paparazzi (Android screenshot), KMP ViewModel tests via JVM target

**Target Platform**: Android 8+ (API 26), iOS 16+, Web (Wasm-capable modern browsers)

**Project Type**: Multi-platform mobile app + microservices backend

**Performance Goals**:
- LCP ≤ 2.5 s on mobile (4G throttled)
- INP ≤ 200 ms
- API p95 < 500 ms for all user-facing endpoints
- Dissertation feedback delivered within 60 s of submission (SC-006)

**Constraints**:
- KMP Wasm bundle will exceed the 150 kB per-chunk constitution rule (see Complexity Tracking below)
- Offline support deferred to v2
- FSRS per-user parameter optimization deferred to v2 (requires ≥ 1,000 reviews)
- No teacher/admin roles in v1

**Scale/Scope**:
- 5 subject tracks × ~100 lessons × ~500 concepts initially
- ~180 simulation questions per full ENEM exam
- Target: ~50,000 students at launch, designed to scale to 500,000

---

## Constitution Check

*GATE: Must pass before implementation. Re-checked after design.*

### I. Code Quality ✅

- Microservice boundaries enforce single responsibility by domain
- Each service has one clear owner (auth, user, content, learning, simulation, dissertation, notification, mail)
- FSRS library (`go-fsrs`) used as-is — no premature custom implementation
- Magic values (FSRS parameters, XP amounts, session batch size) moved to configuration, not hardcoded

### II. Testing Standards ✅

- TDD approach required per constitution: tests written before implementation for each service
- Integration tests use `testcontainers-go` (real PostgreSQL) — no mocked databases
- 80% coverage floor enforced per service; CI fails below threshold
- KMP shared module tested via JVM target for fast iteration + device targets for platform-specific code
- `ExUnit` covers Elixir mail service; real Swoosh adapters used (Local in dev, provider in CI)

### III. User Experience Consistency ✅

- Compose Multiplatform `MaterialTheme` + custom color tokens for subject tracks (no raw hex values in composables)
- All interactive Compose elements use `semantics {}` modifier for WCAG AA accessibility
- Loading, error, and empty states use shared composable components from `shared/presentation/components/`
- Mobile-first: all layouts reviewed on 360dp screen before desktop/tablet variants

### IV. Performance ✅ (with one tracked exception)

- LCP ≤ 2.5 s: achievable on native Android/iOS; Kotlin/Wasm web requires active monitoring
- INP ≤ 200 ms: Compose native handles this; web WASM performance is tracked
- API p95 < 500 ms: Go + pgx + indexed queries on `concept_states(student_id, fsrs_due)` is well within target
- No N+1: all review session queries use `WHERE student_id = ? AND fsrs_due <= now()` with covering index; no per-concept individual fetches

### ⚠ Constitution Violation: Bundle Size (Kotlin/Wasm)

**Violation**: Kotlin/Wasm + Compose Multiplatform web produces a monolithic WASM binary of 2–5 MB gzipped, exceeding the 150 kB per-chunk rule.

**Justification**: The 150 kB rule was designed to prevent route-level JS chunk bloat in web applications. WASM is a single binary that:
- Is cached by the browser beyond session (served with aggressive cache headers)
- Cannot be split across routes with current Kotlin/Wasm tooling (JetBrains roadmap item)
- Achieves ~3× faster runtime performance than an equivalent Kotlin/JS bundle

**Mitigation**:
- Enable Kotlin IR + release optimizations (tree-shaking eliminates unused code)
- Minimize third-party dependencies pulled into the WASM binary
- Defer non-critical initialization to post-load
- Apply a loading skeleton so perceived startup is fast even if WASM takes >2.5 s on first load (cached loads are fast)
- Re-evaluate when JetBrains ships WASM module splitting (tracked in Kotlin roadmap)

---

## Project Structure

### Documentation (this feature)

```text
specs/001-enem-prep-platform/
├── plan.md              # This file
├── research.md          # Phase 0 decisions
├── data-model.md        # Phase 1 entity definitions
├── quickstart.md        # Phase 1 dev setup guide
├── contracts/
│   └── rest-api.md      # External REST API contract
└── tasks.md             # Phase 2 output (/speckit.tasks — not yet created)
```

### Source Code (repository root)

```text
preuni.com.br/
├── mobile/
│   ├── shared/
│   │   └── src/
│   │       ├── commonMain/
│   │       │   ├── domain/         # Entities, repo interfaces, use cases
│   │       │   ├── data/           # Repo implementations, API client, local DB
│   │       │   └── presentation/   # ViewModels, MVIKotlin stores, UI components
│   │       ├── androidMain/
│   │       ├── iosMain/
│   │       └── wasmJsMain/
│   ├── androidApp/
│   ├── iosApp/
│   └── webApp/
│
├── backend/
│   ├── pkg/                        # Shared: logger, errors, middleware, config
│   ├── svc/
│   │   ├── auth/                   # Go: credentials, tokens, OTP
│   │   ├── user/                   # Go: profiles, XP, streaks, achievements
│   │   ├── content/                # Go: tracks, lessons, exercises
│   │   ├── learning/               # Go: FSRS, lesson progress, review sessions
│   │   ├── simulation/             # Go: exams, questions, results
│   │   ├── dissertation/           # Go: prompts, submissions, Claude API
│   │   ├── notification/           # Go: FCM/APNs push
│   │   └── mail/                   # Elixir: Phoenix + Swoosh
│   └── go.work
│
├── infra/
│   ├── docker-compose.yml
│   ├── nginx/
│   └── migrations/                 # Per-service SQL migrations
│       ├── auth/ user/ content/ learning/ simulation/ dissertation/
│
└── specs/
```

**Structure Decision**: Multi-root monorepo. `mobile/` contains the KMP project; `backend/` is a Go workspace with one module per service plus an Elixir service. This avoids cross-language tooling conflicts while keeping all code in one repository for atomic changes across boundaries.

---

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| Kotlin/Wasm bundle > 150 kB | KMP web target has no code-splitting support; monolithic WASM binary is a current technology constraint | A separate React/SvelteKit web app would duplicate all domain logic and UI components, doubling frontend codebase; rejected until Kotlin/Wasm splitting ships |
| Elixir mail service alongside Go | Email reliability and OTP retry logic benefit from Elixir/OTP fault tolerance and Swoosh's adapter pattern | Go-based email (gomail) is functional but lacks the OTP retry/supervision trees Elixir makes trivial; operator expressed a preference for Elixir here |
| 8 microservices (not a monolith) | FSRS, AI dissertation, and email are naturally isolated domains with very different scaling profiles | A Go monolith was considered for v1; rejected because FSRS CPU spike (scheduling) and Claude API latency (dissertation) would block unrelated features in a synchronous monolith |

---

## Service Interface Summary

All services expose **HTTP REST** internally and externally (v1). The NGINX gateway handles:
- TLS termination
- JWT signature validation (rejects invalid tokens before forwarding)
- `X-User-ID` / `X-User-Email` header injection from decoded JWT
- Rate limiting (per IP: 100 req/min for auth endpoints, 1000 req/min for others)
- Routing rules: `/v1/auth/*` → auth-svc, `/v1/students/*` → user-svc, `/v1/tracks/*` + `/v1/lessons/*` → content-svc, `/v1/learning/*` → learning-svc, `/v1/simulations/*` → simulation-svc, `/v1/dissertation/*` → dissertation-svc

Internal service-to-service calls (e.g., learning-svc notifying user-svc of XP to award) use a shared `INTERNAL_SERVICE_TOKEN` in the `Authorization: Bearer {token}` header. The gateway does not validate internal routes.

Proto definitions for potential future gRPC migration are authored in `/backend/proto/` from day one, using the REST contracts as the source of truth for message shapes.

---

## Post-Design Constitution Re-Check ✅

After Phase 1 design:
- Data model confirmed: no cross-service FK dependencies (all cross-service references are UUIDs passed via API)
- All PostgreSQL queries on the hot path (`concept_states` review lookup, `simulation_answers` resume) have covering indexes defined in `data-model.md`
- FSRS state machine fully defined in `data-model.md` — no ambiguous transitions
- REST API contract (`contracts/rest-api.md`) uses consistent error envelope and no implementation leakage
- Dissertation feedback JSON is a fixed schema — no raw LLM text surfaces to the client as unstructured output
