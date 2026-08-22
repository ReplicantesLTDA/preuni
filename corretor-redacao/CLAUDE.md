# corretor-redacao Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-05-28

## Active Technologies

- Python 3.12+. + FastAPI, SQLAlchemy 2.x (async), Alembic, asyncpg, pydantic 2.x, httpx, (001-enem-correction-api)

## Project Structure

```text
src/
tests/
```

## Commands

cd src [ONLY COMMANDS FOR ACTIVE TECHNOLOGIES][ONLY COMMANDS FOR ACTIVE TECHNOLOGIES] pytest [ONLY COMMANDS FOR ACTIVE TECHNOLOGIES][ONLY COMMANDS FOR ACTIVE TECHNOLOGIES] ruff check .

## Code Style

Python 3.12+.: Follow standard conventions

## Recent Changes

- 001-enem-correction-api: Added Python 3.12+. + FastAPI, SQLAlchemy 2.x (async), Alembic, asyncpg, pydantic 2.x, httpx,

<!-- MANUAL ADDITIONS START -->

## Authoritative documents (read before changing code)

- **Constitution** (`.specify/memory/constitution.md`, v2.0.0): 11 NON-NEGOTIABLE / governing
  principles. Highlights: provider neutrality via `LLMProvider` Protocol (Art. II), determinism
  (`temperature ≤ 0.2`, fixed seed) + auditability (Art. III), JSON-Schema-validated structured
  output with 2-attempt corrective retry then typed failure (Art. IV), strict layer separation
  — LLM never touches DB (Art. VI), PII-scrubbed structured logs only (Art. VII), prompts as
  versioned files (Art. VIII), golden-dataset MVP-tier merge gate (Art. IX: total MAE ≤ 120,
  per-comp MAE ≤ 60, ≥ 70% within ±120), LGPD/ECA/B2C + ZDR + parental consent for minors
  (Art. X), library-first per competency (Art. XI).
- **Spec** (`specs/001-enem-correction-api/spec.md`, v2.0.0): B2C only, async correction flow
  (202 + poll), JWT auth, Free (3/mo) and Premium (30/mo) quotas reset 00:00 UTC on the 1st.
- **Plan** (`specs/001-enem-correction-api/plan.md`): architecture, project structure,
  Constitution gate analysis.
- **Research** (`specs/001-enem-correction-api/research.md`): resolved decisions
  (LISTEN/NOTIFY + 5s poll backstop, on-the-fly quota count, Caddy, uv, PyJWT, argon2-cffi,
  `kimi-k2:1t` on Ollama Cloud).
- **Data model** (`specs/001-enem-correction-api/data-model.md`): 7 tables, multi-grader-ready
  via `grader_passes`.
- **Contracts** (`specs/001-enem-correction-api/contracts/`): `openapi.yaml`, the versioned
  LLM output JSON Schema (`correction_output.schema.json`), and the typed error taxonomy
  (`error_codes.md`).

## Hard rules a code reviewer must enforce

- `corrector/` MUST NOT import `db/` or `api/`. Import-linter contract enforces this.
- Scoring literals `{40, 80, 120, 160, 200}` MUST live under `corrector/competencies/` only.
- Free-form fallback parsing of LLM output is FORBIDDEN. JSON Schema or typed failure, period.
- Vendor names (Ollama, Anthropic, OpenAI, …) MUST NOT appear in `corrector/` business logic;
  they live in `corrector/llm/<adapter>.py` and in env config only.
- Audit log payloads MUST go through `AuditLogWriter`; forbidden keys (`essay_text`, `email`,
  `name`, `student_id`, `user_id`, `password`) raise at write time.
- All correction endpoints require a verified-account access token.
- Prompt changes require SemVer bump and golden-dataset metric delta on the PR. CI blocks
  MVP-tier regression.

## Languages and naming

- Code, identifiers, comments, commits, PRs, docs: **English**.
- Prompts, justifications, improvement paths, pt-BR error messages, user-facing copy: **pt-BR**.
- Translation never happens inside `corrector/`; it happens at the interface layer.

<!-- MANUAL ADDITIONS END -->
