# Quickstart: Backend Monolith Refactor

**Feature**: 009-backend-monolith-refactor
**Date**: 2026-05-25

---

## What this feature delivers

1. A single Go backend binary (monolith) that serves the current gateway-facing `/v1/*` HTTP API.
2. Mail delivery rewritten in Go, preserving the existing internal mail send contract.
3. A migration path with rollback: legacy service binaries remain runnable until cutover is verified.

---

## Local setup

```bash
# Infrastructure
make run-infra

# Database migrations
make migrate
```

---

## Run backend (after this feature is implemented)

### Option A: via Docker Compose (recommended)

```bash
# Start gateway + monolith
docker compose -f infra/docker-compose.yml up --build -d monolith gateway
```

### Option B: run monolith directly

```bash
cd backend/svc/monolith && go run ./cmd/server
```

---

## Run tests

### Go workspace tests

```bash
cd backend && go test ./...
```

### Focused auth/user integration tests (example)

```bash
cd backend/svc/auth && go test ./... -run Integration
cd backend/svc/user && go test ./... -run Integration
```

---

## Expected outcomes

- Requests to `/v1/auth/*` and `/v1/students/*` through the gateway are served by the monolith.
- JWT-protected endpoints still require `Authorization: Bearer <access_token>`.
- Internal endpoints still require `Authorization: Bearer <internal_service_token>`.
- `POST /internal/email/send` returns `202` on accepted requests and preserves validation behavior.
