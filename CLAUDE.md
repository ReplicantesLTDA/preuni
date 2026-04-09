# preuni.com.br Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-04-08

## Active Technologies
- Kotlin 2.1.20 + Compose Multiplatform 1.8.0 + Decompose 3.3.0, MVIKotlin 4.2.0, Compose canvas rendering (Skiko), webpack 5 (002-fix-web-compilation)
- sessionStorage (web), SQLDelight/WebWorkerDriver (not yet activated) (002-fix-web-compilation)
- Kotlin 2.1.20 (Compose Multiplatform 1.8.0) + Material Design 3 (already in classpath), Decompose 3.3.0, MVIKotlin 4.2.0 (003-ui-polish)
- N/A (UI layer only — no persistence changes) (003-ui-polish)
- Kotlin 2.1.20 / Compose Multiplatform 1.8.0 + Decompose 3.3.0 (navigation), MVIKotlin 4.2.0 (state), Ktor Client (HTTP) (005-onboarding-ux-fixes)
- `SecureStorage` (expect/actual, already exists) — key-value local store (005-onboarding-ux-fixes)
- Kotlin 2.1.20 / Compose Multiplatform 1.8.0 + Decompose 3.3.0 (navigation), MVIKotlin 4.2.0 (state), Compose Canvas (path drawing) (006-subject-track-path)
- `SecureStorage` (expect/actual, already exists) — add `active_track_id` key to `TokenStore` (006-subject-track-path)

**Frontend (001-enem-prep-platform)**
- Kotlin 2.x + Compose Multiplatform 1.8+ (Android/iOS/Web)
- Decompose (navigation), MVIKotlin (state management)
- Ktor Client (HTTP), SQLDelight (local DB), Koin (DI)
- FSRS-Kotlin (`github.com/open-spaced-repetition/FSRS-Kotlin`)
- Kotlin/Wasm for web target (Beta)

**Backend Go services (001-enem-prep-platform)**
- Go 1.23+: auth, user, content, learning, simulation, dissertation, notification
- go-chi (router), pgx v5 (PostgreSQL), zap (logging), golang-migrate
- go-fsrs v3 (`github.com/open-spaced-repetition/go-fsrs/v3`) in learning-svc
- Anthropic Claude API in dissertation-svc (haiku-4-5 primary, sonnet-4-6 fallback)
- NGINX API gateway (JWT validation + routing)

**Mail service (001-enem-prep-platform)**
- Elixir 1.17+ / Phoenix 1.7 / Swoosh

**Storage**
- PostgreSQL 16 (one schema per service, single instance in v1)
- Redis 7 (JWT refresh tokens, rate-limit counters)
- S3-compatible object storage (avatars, essay support media)

## Project Structure

```text
preuni.com.br/
├── mobile/shared/src/commonMain/   # KMP shared domain, data, presentation
├── mobile/androidApp/              # Android host
├── mobile/iosApp/                  # iOS host
├── mobile/webApp/                  # Kotlin/Wasm web host
├── backend/pkg/                    # Shared Go: logger, errors, middleware, config
├── backend/svc/{auth,user,content,learning,simulation,dissertation,notification}/
├── backend/svc/mail/               # Elixir Phoenix service
├── infra/                          # Docker Compose, NGINX config, migrations
└── specs/001-enem-prep-platform/   # Planning documents
```

## Commands

```bash
# Start local infra
docker compose -f infra/docker-compose.yml up -d postgres redis

# Run a Go service
cd backend/svc/learning && go run ./cmd/server

# Run Elixir mail service
cd backend/svc/mail && mix phx.server

# Run KMP Android app (from Android Studio)
# Run KMP web: cd mobile/webApp && ./gradlew wasmJsBrowserDevelopmentRun

# Run Go tests (unit)
cd backend/svc/learning && go test ./... -short

# Run Go tests (integration, requires Docker)
cd backend/svc/learning && go test ./... -run Integration

# Run KMP shared module tests (fast, JVM target)
cd mobile/shared && ./gradlew desktopTest
```

## Code Style

**Go**: Follow standard `gofmt` + `golangci-lint` conventions; error types from `backend/pkg/errors`
**Kotlin/KMP**: Kotlin coding conventions; composables in PascalCase; coroutines not threads
**Elixir**: `mix format`; ExUnit for tests; pattern match over if/else
**SQL**: lowercase keywords, snake_case identifiers; all new queries need EXPLAIN plan reviewed

## Recent Changes
- 006-subject-track-path: Added Kotlin 2.1.20 / Compose Multiplatform 1.8.0 + Decompose 3.3.0 (navigation), MVIKotlin 4.2.0 (state), Compose Canvas (path drawing)
- 005-onboarding-ux-fixes: Added Kotlin 2.1.20 / Compose Multiplatform 1.8.0 + Decompose 3.3.0 (navigation), MVIKotlin 4.2.0 (state), Ktor Client (HTTP)
- 003-ui-polish: Added Kotlin 2.1.20 (Compose Multiplatform 1.8.0) + Material Design 3 (already in classpath), Decompose 3.3.0, MVIKotlin 4.2.0


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
