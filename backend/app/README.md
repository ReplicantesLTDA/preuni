# preuni monolith (`backend/app/`)

Single Go binary serving the entire backend HTTP surface.

## Routes

- `GET  /health`
- `/v1/auth/*` — register, login, refresh, email verify, OTP login, password reset, password change, email change, logout, account delete
- `/v1/students/*` — me, onboarding, avatar, data export

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
