# Corretor de Redação ENEM

**Internal-only** essay-grading service for the [preuni](../README.md) monorepo — no public API surface. Originally a standalone B2C product (history preserved at github.com/dwbessa/redacao-enem); imported and trimmed down by `specs/014-constitution-alignment-refactor/` (see `.specify/memory/constitution.md` there — that's this service's *own* constitution, v2.0.0, describing its correction-pipeline principles; preuni's top-level constitution at `../.specify/memory/constitution.md` v2.1.1 is the one that actually governs this refactor).

## What it does

The preuni Go monolith (`../backend/app/`) owns identity, quota, and submission intake. It enqueues a job into `correction.correction_jobs` — the *only* table this boundary is DB-mediated through (see `../specs/014-constitution-alignment-refactor/contracts/internal-bridge.md`), not an HTTP call. This service's worker claims the job, runs the LLM correction pipeline, and writes the graded result back to `correction.corrections`; the monolith polls and reconciles it. Each correction breaks the essay into the 5 official ENEM competencies with verbatim excerpts, pt-BR justifications, and improvement paths.

Its own `/auth`, `/me`, and `POST /corrections` endpoints were removed in that refactor — this service no longer has end users of its own, and no longer owns any identity/quota data (`users`, `refresh_tokens`, `consent_records`, `email_verification_tokens` were dropped). `/healthz`, `/readyz`, and `/metrics` are all that remain public.

## Key documents

| Document | Purpose |
|---|---|
| [Constitution](.specify/memory/constitution.md) | 11 non-negotiable governing principles |
| [Spec](specs/001-enem-correction-api/spec.md) | B2C product requirements, user stories, quotas |
| [Plan](specs/001-enem-correction-api/plan.md) | Architecture, tech stack, project structure |
| [Quickstart](specs/001-enem-correction-api/quickstart.md) | Bring up dev env, run e2e smoke |
| [OpenAPI contract](specs/001-enem-correction-api/contracts/openapi.yaml) | HTTP API contract |
| [Error codes](specs/001-enem-correction-api/contracts/error_codes.md) | Typed error taxonomy |
| [Data model](specs/001-enem-correction-api/data-model.md) | DB schema, indexes, constraints |

## Quick start

```bash
cp .env.example .env          # edit DATABASE_URL, JWT_SECRET_KEY, OLLAMA_CLOUD_API_KEY
make up                       # postgres + api + worker via docker-compose
make db-migrate               # alembic upgrade head
curl http://localhost:8000/healthz
```

## Development

```bash
make fmt          # ruff format
make lint         # ruff check + mypy + import-linter
make test-unit    # fast, no DB
make test-int     # requires TEST_DATABASE_URL
make test-golden-fake  # golden harness against FakeProvider (every PR)
```

## Contributor checklist (before push)

- [ ] `make lint` clean
- [ ] `make test-unit` green
- [ ] `make test-int` green (requires local Postgres)
- [ ] `make test-golden-fake` green
- [ ] Prompt changes: bump SemVer in `src/corrector/prompts/`, include metric delta in PR description

## Architecture overview

```
api/          FastAPI surface (thin): health/readiness + metrics only, no public API
workers/      Async worker: LISTEN/NOTIFY + 5 s poll + pipeline execution
corrector/    Pure correction logic: prompts, LLM abstraction, per-competency parsers
db/           SQLAlchemy models (schema `correction`), Alembic migrations, repositories
observability/ structlog JSON, Prometheus metrics, OTel spans
```

The `auth/` package (JWT, argon2id, quota enforcement, email verification)
was removed — the Go monolith owns all of that now.

Import-linter enforces: `corrector/` must not import `db/`, `api/`, or `workers/`.

## License

AGPL-3.0. See [NOTICE](NOTICE) for upstream attributions (Essay-BR dataset).
