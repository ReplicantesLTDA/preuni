# Corretor de Redação ENEM

B2C API for automated ENEM essay correction via LLM — Constitution v2.1.0.

## What it does

A Brazilian high-school student submits an essay + prompt theme, receives a 202, and polls until a structured correction is ready. Each correction breaks the essay into the 5 official ENEM competencies with verbatim excerpts, pt-BR justifications, and improvement paths.

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
api/          FastAPI surface (thin): auth, corrections, me
workers/      Async worker: LISTEN/NOTIFY + 5 s poll + pipeline execution
corrector/    Pure correction logic: prompts, LLM abstraction, per-competency parsers
db/           SQLAlchemy models, Alembic migrations, repositories
auth/         JWT, argon2id passwords, quota enforcement, email verification
observability/ structlog JSON, Prometheus metrics, OTel spans
```

Import-linter enforces: `corrector/` must not import `db/`, `api/`, or `workers/`.

## License

AGPL-3.0. See [NOTICE](NOTICE) for upstream attributions (Essay-BR dataset).
