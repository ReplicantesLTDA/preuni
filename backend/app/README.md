# preuni monolith

Unified backend binary. Serves the gateway-facing `/v1/*` HTTP API plus
internal endpoints, replacing `auth-svc`, `user-svc`, and the Elixir
`mail-svc` with a single Go process.

## Routes mounted

- `GET  /health`
- `/v1/auth/*` — register, login, refresh, email verify, OTP login,
  password reset, password change, email change, logout, account delete
- `/v1/students/*` — me, onboarding, avatar, data export
- `/internal/students` — provisioning (internal-token protected)
- `/internal/email/send` — transactional email (internal-token protected)

## Run locally

### Option A — Docker Compose (recommended)

```bash
docker compose -f infra/docker-compose.yml up --build -d monolith gateway
```

Gateway listens on `:8080`. Routes `/v1/auth/*` and `/v1/students/*` go to
the monolith. Monolith itself is exposed on host port `:8088`.

### Option B — Go run

```bash
cp backend/svc/monolith/.env.example backend/svc/monolith/.env
# fill in values, then:
make run-monolith
```

## Tests

```bash
make test-monolith        # unit + handler tests
cd backend/svc/monolith && go test ./...
```

## Rollback to legacy split services

The legacy `auth`, `user`, and `mail` containers remain in
`infra/docker-compose.yml` and the legacy upstream blocks in
`infra/nginx/nginx.conf`. To roll back gateway routing:

1. In `infra/nginx/nginx.conf`, change `proxy_pass http://monolith_svc`
   back to `http://auth_svc` (for `/v1/auth/`) and `http://user_svc`
   (for `/v1/students/`).
2. `docker compose up -d --no-deps gateway` to reload nginx.
3. `docker compose up -d --build auth user mail` to bring legacy back.

## Self-HTTP loopback

In monolith mode, auth handlers still self-HTTP to `/internal/students`
and `/internal/email/send` via `SELF_BASE_URL`. This preserves the
existing service contract and means internal endpoints remain
internal-token protected. Direct in-process adapters are deferred —
see specs/009-backend-monolith-refactor/tasks.md (T008–T011, T026, T049).
