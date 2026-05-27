# preuni Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-05-27

## Active Technologies

**Frontend (012-expo-rn-frontend — current)**

- React Native 0.81 + Expo SDK 54 + TypeScript 5.9 (strict)
- Expo Router 6 (file-based routing, `(auth)/(onboarding)/(tabs)` groups)
- TanStack Query v5 (server cache + mutations), Zustand v5 (client/session state), Zod v3 (schemas)
- `expo-secure-store` (session tokens, web fallback to `localStorage`)
- `expo-image`, `expo-image-picker`, `react-native-reanimated` 4 (+ `react-native-worklets`), `react-native-svg`, `react-native-safe-area-context`, `react-native-screens`, `@expo/vector-icons`
- `@expo-google-fonts/{caveat-brush,patrick-hand,architects-daughter,kalam,caveat}` (hand-drawn family stack)
- Jest 29 + jest-expo 54 + `@testing-library/react-native` 13 (msw deferred; tests use `globalThis.fetch` mock)

**Backend (010-backend-monolith-cleanup)**

- Single Go 1.24 binary at `backend/app/`. Module path `github.com/preuni/app`.
- Domains under `app/internal/<domain>/`: `auth`, `user`, `mail` (live); `content`, `learning`, `simulation`, `dissertation`, `notification` (scaffolding, README-only).
- go-chi/chi v5 (router), pgx/v5 (PostgreSQL), golang-jwt/jwt v5 (JWT), zap (logging), `net/smtp` (mail).
- Shared infra in `backend/pkg/` (config, logger, errors, middleware).
- NGINX gateway routes `/v1/auth/*` + `/v1/students/*` to the monolith.
- Backend wire format is **snake_case**; client uses a case-converter in `mobile/src/lib/api/caseConvert.ts` to round-trip camelCase ↔ snake_case at the network boundary.

**Storage**

- PostgreSQL 16 (one schema per domain, single instance in v1)
- Redis 7 (JWT refresh tokens, rate-limit counters)
- S3-compatible object storage (avatars, essay support media)

## Project Structure

```text
preuni/
├── mobile/                         # React Native + Expo app (replaces the KMP tree retired in 012)
│   ├── app/                        # Expo Router routes
│   │   ├── (auth)/                 # public — login, register, verify-email, otp-login, password-reset, welcome
│   │   ├── (onboarding)/           # authed + !onboardingCompleted — welcome, profile, interests
│   │   └── (tabs)/                 # authed + onboardingCompleted — trilha, redacao, simulado, perfil (each with nested Stack)
│   ├── src/
│   │   ├── theme/                  # design tokens (sourced from wireframe.html)
│   │   ├── components/             # shared UI primitives
│   │   ├── features/<domain>/      # vertical slices: api.ts, hooks.ts, validation.ts
│   │   ├── lib/api/                # typed fetch client + refresh interceptor + case converter + error mapping
│   │   ├── lib/auth/               # tokenStore, bootstrap
│   │   ├── lib/query/              # TanStack Query client + keys
│   │   ├── lib/i18n/               # pt-BR copy
│   │   ├── stores/                 # zustand stores
│   │   └── types/                  # zod schemas + inferred TS
│   ├── assets/fonts/               # placeholders only — actual TTFs pulled from @expo-google-fonts/*
│   └── tests/                      # jest + RTL + globalThis.fetch mock
├── backend/pkg/                    # Shared Go: logger, errors, middleware, config
├── backend/app/                    # Go monolith binary
│   ├── cmd/server/                 # main entrypoint
│   ├── internal/auth/              # auth domain
│   ├── internal/user/              # user domain
│   ├── internal/mail/              # in-process mail
│   ├── internal/{content,learning,simulation,dissertation,notification}/  # scaffolding
│   └── internal/{adapters,config,router}/
├── infra/                          # Docker Compose, NGINX config, migrations
└── specs/                          # Planning documents
```

## Commands

```bash
# Local infra
docker compose -f infra/docker-compose.yml up -d postgres redis

# Backend monolith
cd backend/app && go run ./cmd/server
# or: make run-monolith
make dev                                        # full stack (postgres + redis + monolith + gateway)

# Backend tests
cd backend/app && go test -short ./...
cd backend/app && TEST_DB_URL="postgres://preuni:preuni@localhost:5432/preuni?sslmode=disable" go test ./tests/integration/...

# Mobile app
cd mobile && pnpm install
cd mobile && pnpm start                          # Expo dev server (QR for Expo Go)
cd mobile && pnpm ios | pnpm android | pnpm web
cd mobile && pnpm typecheck && pnpm lint && pnpm test
cd mobile && pnpm build:web                      # static export → mobile/dist/
# Makefile shortcuts: mobile-install, mobile-start, mobile-test, mobile-web, mobile-lint
```

## Code Style

- **Go**: standard `gofmt` + `golangci-lint`; error types from `backend/pkg/errors`.
- **TypeScript**: strict mode, `noUncheckedIndexedAccess`, ESLint flat config; no raw hex/px in component styles (Constitution III enforced by `no-restricted-syntax` rule); design tokens from `mobile/src/theme/tokens.ts`.
- **SQL**: lowercase keywords, snake_case identifiers; new queries need EXPLAIN plan reviewed.

## Recent Changes

- **012-expo-rn-frontend**: Retired the Kotlin Multiplatform frontend. Replaced with React Native + Expo SDK 54 + TypeScript at `mobile/`. Single codebase ships to iOS / Android (via Expo Go) and Web (via `expo export --platform web`). Backend untouched.
- 010-backend-monolith-cleanup: Go 1.24 monolith at `backend/app/`; chi v5 router; pgx/v5 Postgres; jwt v5; zap; `net/smtp`. No new dependencies.

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
