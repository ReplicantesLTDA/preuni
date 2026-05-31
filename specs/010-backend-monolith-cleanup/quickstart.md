# Quickstart: Backend Folder Reorganization

**Feature**: 010-backend-monolith-cleanup
**Date**: 2026-05-25

---

## What this feature delivers

1. `backend/svc/` directory gone. Backend tree is `pkg/` + `app/` + `go.work`.
2. Deprecated Elixir mail tree deleted.
3. Five stub services folded as empty `app/internal/<domain>/` packages with README markers.
4. Constitution + `CLAUDE.md` aligned with the actual architecture.

---

## Verify the reorg locally

### Build + test

```bash
cd backend/app
go build ./...
go test ./...

# integration test needs a running Postgres:
docker compose -f ../../infra/docker-compose.yml up -d postgres redis
make -C ../.. migrate
TEST_DB_URL="postgres://preuni:preuni@localhost:5432/preuni?sslmode=disable" go test ./tests/integration/...
```

Expected: all tests green. No reference to `backend/svc/` in any error.

### Layout check

```bash
ls backend/
# expected: go.work  pkg/  app/

ls backend/svc 2>/dev/null
# expected: ls: no such file or directory

find backend -name "*.ex" -o -name "*.exs"
# expected: (empty)

grep -r "backend/svc/" Makefile infra/ backend/ 2>/dev/null | grep -v 'specs/'
# expected: (empty)
```

### Live HTTP smoke

```bash
docker compose -f infra/docker-compose.yml up --build -d monolith gateway
sleep 3

# health
curl -sS http://localhost:8088/health
# → ok

# register through gateway
EMAIL="qs+$(date +%s)@preuni.com.br"
curl -sS -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"P@ssw0rd123\",\"display_name\":\"Reorg\"}"
# → 201 + AuthResponse

# /v1/students/me with returned access_token
TOKEN=...   # paste from previous response
curl -sS -H "Authorization: Bearer $TOKEN" http://localhost:8088/v1/students/me
# → 200 + StudentResponse
```

---

## Rollback

If the reorg breaks something not caught by tests:

```bash
git checkout pre-monolith-reorg   # tag created in Phase A
```

The pre-reorg tree builds and deploys exactly as it did before the change.

---

## Expected outcomes

- `make dev` + `make test-backend` work without changes.
- HTTP surface unchanged. Mobile + web clients require no updates.
- Onboarding doc reading time drops because there is one place to look for backend code.
