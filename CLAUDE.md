# preuni.com.br Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-05-25

## Active Technologies
- Kotlin 2.1.20 + Compose Multiplatform 1.8.0 + Decompose 3.3.0, MVIKotlin 4.2.0, Compose canvas rendering (Skiko), webpack 5 (002-fix-web-compilation)
- sessionStorage (web), SQLDelight/WebWorkerDriver (not yet activated) (002-fix-web-compilation)
- Kotlin 2.1.20 (Compose Multiplatform 1.8.0) + Material Design 3 (already in classpath), Decompose 3.3.0, MVIKotlin 4.2.0 (003-ui-polish)
- N/A (UI layer only — no persistence changes) (003-ui-polish)
- Kotlin 2.1.20 / Compose Multiplatform 1.8.0 + Decompose 3.3.0 (navigation), MVIKotlin 4.2.0 (state), Ktor Client (HTTP) (005-onboarding-ux-fixes)
- `SecureStorage` (expect/actual, already exists) — key-value local store (005-onboarding-ux-fixes)
- Kotlin 2.1.20 / Compose Multiplatform 1.8.0 + Decompose 3.3.0 (navigation), MVIKotlin 4.2.0 (state), Compose Canvas (path drawing) (006-subject-track-path)
- `SecureStorage` (expect/actual, already exists) — add `active_track_id` key to `TokenStore` (006-subject-track-path)
- Kotlin 2.1.20 (KMP shared module), Go 1.23 (backend — no changes) + Compose Multiplatform 1.8.0, Decompose 3.3.0, MVIKotlin 4.2.0, Ktor Client (007-profile-mgmt-fixes)
- `SecureStorage` / `TokenStore` (local session only); no new persistent storage (007-profile-mgmt-fixes)
- Kotlin 2.1.20 (KMP shared + Android/iOS/Web entrypoints), Go 1.23 (user-svc integration tests) + Compose Multiplatform 1.8.0, Decompose 3.3.0, MVIKotlin 4.2.0, Ktor Client, kotlinx.serialization, go-chi + pgx (existing backend stack) (008-fix-backend-integrations)
- `SecureStorage` / `TokenStore` for session tokens; PostgreSQL 16 only for backend integration test fixture (008-fix-backend-integrations)
- Go 1.24 (per `go.work`/`go.mod`); Elixir mail service currently exists and will be replaced for this feature + go-chi/chi (HTTP routing), pgx/v5 (PostgreSQL), golang-jwt/jwt (JWT), zap (logging), testify (tests), NGINX (gateway routing + rate limiting) (009-backend-monolith-refactor)
- PostgreSQL 16 (auth + users schemas), Redis 7 (existing infra), S3-compatible object storage (avatar flows) (009-backend-monolith-refactor)
- Go 1.24 (per `backend/go.work`) + go-chi/chi v5 (router), pgx/v5 (Postgres), golang-jwt/jwt v5 (JWT), zap (logging), testify (tests), `net/smtp` (mail). No new dependencies introduced. (010-backend-monolith-cleanup)
- PostgreSQL 16 (`auth.*`, `users.*` schemas — unchanged). Redis 7 (existing). S3 (avatars). (010-backend-monolith-cleanup)

**Frontend (001-enem-prep-platform)**
- Kotlin 2.x + Compose Multiplatform 1.8+ (Android/iOS/Web)
- Decompose (navigation), MVIKotlin (state management)
- Ktor Client (HTTP), SQLDelight (local DB), Koin (DI)
- FSRS-Kotlin (`github.com/open-spaced-repetition/FSRS-Kotlin`)
- Kotlin/Wasm for web target (Beta)

**Backend (010-backend-monolith-cleanup)**
- Single Go 1.24 binary at `backend/app/`. Module path `github.com/preuni/app`.
- Domains under `app/internal/<domain>/`: `auth`, `user`, `mail` (live); `content`, `learning`, `simulation`, `dissertation`, `notification` (scaffolding, README-only).
- go-chi/chi v5 (router), pgx/v5 (PostgreSQL), golang-jwt/jwt v5 (JWT), zap (logging), `net/smtp` (mail).
- Shared infra in `backend/pkg/` (config, logger, errors, middleware).
- Mail handled in-process: SMTP via `net/smtp` (implicit TLS 465 / STARTTLS 587), fire-and-forget delivery.
- NGINX gateway routes `/v1/auth/*` + `/v1/students/*` to the monolith.

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
├── backend/app/                    # Go monolith binary (module: github.com/preuni/app)
│   ├── cmd/server/                 # main entrypoint
│   ├── internal/auth/              # auth domain (handlers, repo, router, ports, adapters)
│   ├── internal/user/              # user domain
│   ├── internal/mail/              # in-process mail (validator, templates, SMTP)
│   ├── internal/{content,learning,simulation,dissertation,notification}/  # scaffolding
│   ├── internal/{adapters,config,router}/
│   └── tests/{contract,integration}/
├── infra/                          # Docker Compose, NGINX config, migrations
└── specs/                          # Planning documents
```

## Commands

```bash
# Start local infra
docker compose -f infra/docker-compose.yml up -d postgres redis

# Run the backend monolith locally
cd backend/app && go run ./cmd/server
# or: make run-monolith

# Bring up monolith + gateway via compose
make dev

# Run Go tests (unit + contract)
cd backend/app && go test -short ./...

# Run Go integration tests (requires running Postgres)
cd backend/app && TEST_DB_URL="postgres://preuni:preuni@localhost:5432/preuni?sslmode=disable" go test ./tests/integration/...

# Run KMP Android app (from Android Studio)
# Run KMP web: cd mobile/webApp && ./gradlew wasmJsBrowserDevelopmentRun

# Run KMP shared module tests (fast, JVM target)
cd mobile/shared && ./gradlew desktopTest
```

## Code Style

**Go**: Follow standard `gofmt` + `golangci-lint` conventions; error types from `backend/pkg/errors`
**Kotlin/KMP**: Kotlin coding conventions; composables in PascalCase; coroutines not threads
**Elixir**: `mix format`; ExUnit for tests; pattern match over if/else
**SQL**: lowercase keywords, snake_case identifiers; all new queries need EXPLAIN plan reviewed

## Recent Changes
- 010-backend-monolith-cleanup: Added Go 1.24 (per `backend/go.work`) + go-chi/chi v5 (router), pgx/v5 (Postgres), golang-jwt/jwt v5 (JWT), zap (logging), testify (tests), `net/smtp` (mail). No new dependencies introduced.
- 009-backend-monolith-refactor: Added Go 1.24 (per `go.work`/`go.mod`); Elixir mail service currently exists and will be replaced for this feature + go-chi/chi (HTTP routing), pgx/v5 (PostgreSQL), golang-jwt/jwt (JWT), zap (logging), testify (tests), NGINX (gateway routing + rate limiting)
- 008-fix-backend-integrations: Added Kotlin 2.1.20 (KMP shared + Android/iOS/Web entrypoints), Go 1.23 (user-svc integration tests) + Compose Multiplatform 1.8.0, Decompose 3.3.0, MVIKotlin 4.2.0, Ktor Client, kotlinx.serialization, go-chi + pgx (existing backend stack)


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
