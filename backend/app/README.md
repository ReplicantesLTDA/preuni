# preuni monolith (`backend/app/`)

Single Go binary serving the entire backend HTTP surface.

## Routes

- `GET  /health`
- `/v1/auth/*` — register, login, refresh, email verify, OTP login, password reset, password change, email change, logout, account delete
- `/v1/students/*` — me, onboarding, avatar, data export
- `/v1/essays` — submit (quota-gated), get, list; grading is async via the
  correction service (see below)
- `/v1/streaks/me` — current/longest streak
- `/v1/friends/*` — requests, accept, remove, list (visibility-gated)
- `/v1/ranking/*`, `/v1/medals/me` — weekly leaderboard + earned medals

### Essay grading (the correction service bridge)

The monolith is the only owner of identity, quota, streaks, friends, and
ranking. Essay grading itself is delegated to `ai-corrector/` — an
internal-only Python service with no public API of its own — over a
DB-mediated bridge, not HTTP: `essay.Repository.Submit` inserts a row into
`correction.correction_jobs` (schema owned by `ai-corrector`'s Alembic
migrations) in the same transaction as the submission; the correction
worker claims and grades it; `essay.Repository.ReconcileOnce` (a 5s ticker
in `cmd/server/main.go`) polls for the result and writes `essay_grades`.
See `specs/014-constitution-alignment-refactor/contracts/internal-bridge.md`
for the full contract and `research.md` for why this is DB-mediated rather
than an HTTP call.

Gamification (streaks, weekly ranking, medals) is wired as a best-effort
side effect of that same flow: `essay.Repository` holds an optional
`GamificationHooks` interface that `gamification.Repository` satisfies
structurally (no import needed) — see `internal/essay/repository/essay.go`.
`gamification.Repository.WeekClose` runs on an hourly ticker
(`runWeekCloseJob`) and is idempotent per tier-week.

Mail is **in-process**: auth handlers dispatch via the `EmailSender` port
(`app/internal/adapters/email_sender.go`), which calls
`mail.Validate → Build → Sender.Send` directly. No HTTP mail endpoint.

Student provisioning is also in-process via the `StudentProvisioner` port
(`app/internal/adapters/student_provisioner.go`) — no HTTP `/internal/students`.

## Run locally

### Option A — Docker Compose

```bash
docker compose -f infra/docker-compose.yml up --build -d monolith gateway
# gateway: :8080   monolith (direct): :8088
```

### Option B — Go run

```bash
# Env vars live in infra/.env (same file docker compose uses).
set -a; source infra/.env; set +a
make run-monolith
```

## Tests

```bash
make test-monolith
# or
cd backend/app && go test ./...
```

Integration tests in `tests/integration/` need a running Postgres; they skip
gracefully if unreachable. Handler-level integration tests read `TEST_DB_URL`
from the environment.

## Layout

```
cmd/server/                  main entrypoint
internal/auth/               handlers, repo, router, ports
internal/user/               handlers, repo, router
internal/essay/              submission (quota + streak + correction_jobs enqueue), reconciler, handlers, router
internal/streak/             pure streak state machine (domain/) + persistence + handler
internal/social/             friend requests/accept/remove + visibility-gated list
internal/gamification/       weekly ranking, week-close promotion/demotion, medals
internal/mail/               validator, templates, SMTP sender
internal/adapters/           in-process implementations of auth ports
internal/config/             env loading
internal/router/             top-level chi router that mounts the domains
internal/{content,learning,
          simulation,
          dissertation,
          notification}/     scaffolding (README only — no endpoints yet)
tests/{contract,integration}/
```
