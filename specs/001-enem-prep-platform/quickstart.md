# Quickstart: preuni.com.br – ENEM Prep Platform

**Branch**: `001-enem-prep-platform` | **Date**: 2026-04-03

## Repository Layout

```
preuni.com.br/
├── mobile/                        # Kotlin Multiplatform (KMP)
│   ├── shared/                    # Shared business logic & domain
│   │   ├── build.gradle.kts
│   │   └── src/
│   │       ├── commonMain/        # 100% shared: domain, data, presentation logic
│   │       │   ├── domain/        # Entities, repositories (interfaces), use cases
│   │       │   ├── data/          # Repository implementations, DTOs, network, DB
│   │       │   └── presentation/  # ViewModels (KMP ViewModel), MVIKotlin stores
│   │       ├── androidMain/       # Android-specific implementations (e.g., notifications)
│   │       ├── iosMain/           # iOS-specific implementations
│   │       └── wasmJsMain/        # Web-specific implementations
│   ├── androidApp/                # Android host app (Compose UI)
│   ├── iosApp/                    # iOS host app (Compose Multiplatform UI)
│   └── webApp/                    # Web host (Kotlin/Wasm + Compose)
│
├── backend/
│   ├── pkg/                       # Shared Go modules
│   │   ├── logger/                # zap-based structured logging
│   │   ├── errors/                # Typed error definitions
│   │   ├── middleware/            # HTTP middleware (auth header parsing, request ID, CORS)
│   │   └── config/                # Environment-based config loader
│   ├── svc/                       # Microservices (one module each)
│   │   ├── gateway/               # NGINX config + optional Go custom handlers
│   │   ├── auth/                  # Go: registration, login, tokens
│   │   ├── user/                  # Go: profiles, XP, streaks, achievements
│   │   ├── content/               # Go: tracks, lessons, exercises
│   │   ├── learning/              # Go: FSRS engine, lesson progress, reviews
│   │   ├── simulation/            # Go: ENEM simulations, questions, results
│   │   ├── dissertation/          # Go: prompts, submissions, Claude API integration
│   │   ├── notification/          # Go: FCM/APNs push notifications
│   │   └── mail/                  # Elixir: Phoenix + Swoosh transactional email
│   └── go.work                    # Go workspace for multi-module monorepo
│
├── infra/
│   ├── docker-compose.yml         # Local dev stack (PostgreSQL, Redis, all services)
│   ├── nginx/                     # NGINX gateway configuration
│   └── migrations/                # Per-service SQL migrations (golang-migrate)
│       ├── auth/
│       ├── user/
│       ├── content/
│       ├── learning/
│       ├── simulation/
│       └── dissertation/
│
└── specs/                         # SpecKit planning documents
    └── 001-enem-prep-platform/
```

---

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Docker + Docker Compose | 24+ | docker.com |
| Go | 1.23+ | go.dev |
| Elixir | 1.17+ | elixir-lang.org |
| Kotlin / Android Studio | Kotlin 2.x, Android Studio Ladybug+ | developer.android.com |
| Xcode | 16+ (for iOS) | App Store |
| Node.js | 20+ (for Kotlin/Wasm web tooling) | nodejs.org |

---

## Local Development Setup

### 1. Start infrastructure services

```bash
docker compose -f infra/docker-compose.yml up -d postgres redis
```

This starts:
- PostgreSQL 16 on port `5432`
- Redis 7 on port `6379`

### 2. Run database migrations

```bash
# Install golang-migrate once
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run all service migrations
for svc in auth user content learning simulation dissertation; do
  migrate -path infra/migrations/$svc -database "postgres://preuni:preuni@localhost:5432/${svc}?sslmode=disable" up
done
```

### 3. Start backend services

```bash
# Start all Go services via Docker Compose (recommended for development)
docker compose -f infra/docker-compose.yml up -d auth user content learning simulation dissertation notification

# Or run a single service locally for active development
cd backend/svc/learning
go run ./cmd/server
```

### 4. Start the mail service

```bash
cd backend/svc/mail
mix deps.get
mix phx.server
```
Mail service starts on port `4000` by default.

### 5. Start NGINX gateway

```bash
docker compose -f infra/docker-compose.yml up -d gateway
```
API gateway available at `http://localhost:8080/v1`.

### 6. Run the mobile app

**Android** (from Android Studio):
```
Open mobile/ → Select androidApp run configuration → Run on device or emulator
```

**iOS** (from Xcode):
```bash
cd mobile/iosApp
open iosApp.xcworkspace  # Opens Xcode
# Select a simulator and press Run
```

**Web** (Kotlin/Wasm):
```bash
cd mobile/webApp
./gradlew wasmJsBrowserDevelopmentRun
# Opens at http://localhost:8080
```

---

## Environment Variables

Each service reads from environment variables. A `.env.example` file in each service directory documents required variables.

**Common across Go services**:
```
DATABASE_URL=postgres://preuni:preuni@localhost:5432/{service-name}
REDIS_URL=redis://localhost:6379
LOG_LEVEL=debug
PORT=808x
INTERNAL_SERVICE_TOKEN=dev-shared-secret
```

**auth-svc extras**:
```
JWT_SIGNING_KEY=dev-signing-key-min-32-chars
JWT_ACCESS_EXPIRY_SECONDS=3600
JWT_REFRESH_EXPIRY_DAYS=30
MAIL_SERVICE_URL=http://localhost:4000
```

**dissertation-svc extras**:
```
ANTHROPIC_API_KEY=sk-ant-...
CLAUDE_MODEL=claude-haiku-4-5-20251001
```

**mail-svc (Elixir)**:
```
SWOOSH_ADAPTER=Swoosh.Adapters.Local   # Dev: uses local mailbox at /dev/mailbox
SWOOSH_FROM_EMAIL=noreply@preuni.com.br
INTERNAL_TOKEN=dev-shared-secret
```
In development, Swoosh Local adapter stores emails in memory. Open `http://localhost:4000/dev/mailbox` to inspect sent emails.

---

## Running Tests

### Go services

```bash
# Unit tests (no external dependencies)
cd backend/svc/learning
go test ./... -short

# Integration tests (requires Docker for testcontainers)
go test ./... -run Integration
```

### Elixir mail service

```bash
cd backend/svc/mail
mix test
```

### KMP shared module

```bash
cd mobile/shared
./gradlew allTests          # Runs on all targets
./gradlew desktopTest       # Fast unit tests via JVM target
```

### Coverage check (must stay ≥ 80%)

```bash
# Go
go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out

# Kotlin
./gradlew koverReport
```

---

## Key Architecture Decisions (summary)

| Decision | Choice | Where documented |
|----------|--------|-----------------|
| Spaced repetition | FSRS v6 (`go-fsrs/v3`) | research.md |
| KMP navigation | Decompose + MVIKotlin | research.md |
| Service communication | REST/JSON (v1) | research.md |
| API gateway | NGINX | research.md |
| AI feedback | Claude API (haiku-4-5 / sonnet-4-6) | research.md |
| Database | PostgreSQL 16 per service schema | data-model.md |
| Email | Elixir + Phoenix + Swoosh | research.md |

---

## Common Development Workflows

### Adding a new exercise type

1. Add the new `type` value to the `exercises.type` CHECK constraint via migration in `infra/migrations/content/`
2. Add the type to the `ExerciseType` enum in `backend/svc/content/domain/exercise.go`
3. Add the type to the shared KMP `ExerciseType` enum in `mobile/shared/src/commonMain/domain/Exercise.kt`
4. Implement rendering in `mobile/shared/src/commonMain/presentation/` (shared) and platform-specific UI if needed
5. Write unit tests for any new answer-evaluation logic in `content-svc`

### Adding a new achievement

1. Insert a row into `achievements` table via a content migration in `infra/migrations/user/`
2. Add the trigger condition to `user-svc`'s achievement evaluator (called after XP events, lesson completions, simulation completions)
3. Write a unit test asserting the condition fires at the correct threshold

### Updating the FSRS parameters per user (v2 prep)

The `review_logs` table in `learning-svc` captures every review with full FSRS state. When a student accumulates ≥ 1,000 reviews, a background job can run gradient-descent optimization using `go-fsrs` and update the student's personal FSRS parameters in a `student_fsrs_params` table (not yet created — defer to v2).

---

## Release Process

### Bumping the version

The project uses semantic versioning (`MAJOR.MINOR.PATCH`) stored in `version.txt`. The `Makefile` automates the entire release cycle:

```bash
# Patch release (bug fixes)
make bump-version patch

# Minor release (new backward-compatible features)
make bump-version minor

# Major release (breaking changes)
make bump-version major
```

`bump-version` will:
1. Read the current version from `version.txt`
2. Increment the requested component
3. Write the new version back to `version.txt`
4. Update `appVersion` in `mobile/gradle/libs.versions.toml` to keep the Android/iOS build in sync
5. Create a signed git commit `chore(release): bump version to X.Y.Z`
6. Create an annotated tag `vX.Y.Z`

### Generating the CHANGELOG

[git-cliff](https://git-cliff.org/) reads the conventional commit history and produces a structured `CHANGELOG.md`:

```bash
# (Re)generate CHANGELOG.md from the full history
make changelog

# Preview the next release entry without writing the file
git cliff --unreleased
```

The `.cliff.toml` at the repo root groups commits by type:

| Commit type | CHANGELOG section |
|-------------|-------------------|
| `feat`      | Features          |
| `fix`       | Bug Fixes         |
| `refactor`  | Refactoring       |
| `chore`     | Maintenance       |

Breaking changes (commits with `!` or a `BREAKING CHANGE:` footer) are marked with ⚠️ and listed at the top of the entry.

### Pushing a release

```bash
# Push commits and the new tag together
git push origin main --follow-tags
```

CI will pick up the tag and trigger the release pipeline (Docker image build, Play Store / App Store submission, etc.).

### Hotfix workflow

1. Branch off the tag: `git checkout -b hotfix/v1.0.1 v1.0.0`
2. Apply the fix and commit with `fix(scope): ...`
3. Run `make bump-version patch` to cut `v1.0.1`
4. Cherry-pick the fix commit to `main` if needed
5. Push with `--follow-tags`
