# preuni

preuni is an essay-challenge app for ENEM (Brazilian university entrance exam)
prep: students write essays as often as they can, get them graded by an AI
corrector, and build a daily submission streak. Gamification — streaks,
friends, medals, weekly ranking — is a core mechanic, not a bolted-on
feature, the way Duolingo reinforces its daily-habit loop.

- **Free tier**: one graded essay submission per day.
- **Pro tier**: multiple submissions per day.
- **Social**: add friends, see their streaks and latest grades.
- **Ranking**: a weekly leaderboard aggregates grades; users rank up or down
  week over week across league tiers.
- **Medals**: awarded for streak milestones and ranking achievements.

## Architecture

Two backend systems with a clear ownership boundary, one mobile client:

- **`backend/app/`** — a single Go monolith, system of record for auth, user
  profiles, streaks, friends, medals, ranking, and quota enforcement.
  Domains live under `internal/<domain>/`. Shared infra (logger, errors,
  middleware, config) is in `backend/pkg/`.
- **`corretor-redacao/`** — a Python/FastAPI correction service. Owns
  AI-driven essay grading only: takes an essay + prompt theme, runs the LLM
  correction pipeline, returns the structured per-competency result. Never
  talks to end users directly; the monolith calls it internally and the two
  share no identity/streak/ranking data.
- **`mobile/`** — React Native + Expo (SDK 54) + TypeScript, one codebase
  shipping to iOS, Android, and Web.
- NGINX is the external gateway. PostgreSQL 16 and Redis 7 are the only
  stateful stores.

Full detail: [`.specify/memory/constitution.md`](.specify/memory/constitution.md).

## Project layout

```text
preuni/
├── mobile/              # Expo Router app — (auth)/(onboarding)/(tabs)
├── backend/
│   ├── pkg/              # shared Go: logger, errors, middleware, config
│   └── app/               # the Go monolith binary
│       ├── cmd/server/     # entrypoint
│       └── internal/        # auth, user, mail, essay, streak, social, gamification
├── corretor-redacao/    # Python correction service (internal-only, no public API)
├── infra/               # docker-compose, NGINX config, SQL migrations
├── e2e/                 # Playwright E2E suite (network-level, real backend)
└── specs/               # spec-kit planning docs (one dir per feature)
```

## Getting started

Prerequisites: Docker, Go 1.24+ (go.work pins toolchain 1.26.6, auto-fetched), Node 20+ with `pnpm`, Python 3.12+ with `uv`.

```bash
# Local infra (Postgres + Redis)
docker compose -f infra/docker-compose.yml up -d postgres redis

# Backend monolith
cd backend/app && go run ./cmd/server
# or: make run-monolith
make dev                # full stack: postgres + redis + monolith + gateway + corrector-api + corrector-worker

# Mobile app
cd mobile && pnpm install
cd mobile && pnpm start                  # Expo dev server (QR for Expo Go)
cd mobile && pnpm ios | pnpm android | pnpm web

# Correction service
cd corretor-redacao && uv pip install -e ".[dev]"
```

## Testing

Every codebase carries a CI-enforced coverage floor (constitution Principle
II: 90%) plus dedicated E2E/race/soak suites for the backend↔client seam and
concurrency-sensitive paths.

```bash
# Backend unit + integration
cd backend/app && go test -p 1 ./...
cd backend/app && TEST_DB_URL="postgres://preuni:preuni@localhost:5432/preuni?sslmode=disable" go test ./tests/integration/...

# Mobile
cd mobile && pnpm typecheck && pnpm lint && pnpm test

# Correction service
cd corretor-redacao && pytest

# E2E / race / soak (see specs/023-e2e-race-soak-tests/quickstart.md)
cd e2e && pnpm install && pnpm exec playwright install --with-deps chromium && pnpm test
cd backend/app && go test -race ./tests/race/...
cd backend/app && SOAK=1 go test -timeout 30m ./tests/soak/...
```

## Development workflow

This project follows [spec-kit](https://github.com/github/spec-kit): every
non-trivial feature gets a spec, plan, and task list under `specs/` before
implementation, generated via the `/speckit.*` slash commands. Architectural
decisions that aren't obvious from the code are recorded as ADRs under
[`docs/decisions/`](docs/decisions/).

See [`CLAUDE.md`](CLAUDE.md) for the full command reference and code style
rules, and [`.specify/memory/constitution.md`](.specify/memory/constitution.md)
for the non-negotiable engineering principles (testing standards, migration
reversibility, security & secrets, CI gates).
