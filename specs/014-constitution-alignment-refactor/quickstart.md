# Quickstart: Constitution Alignment Refactor

## Bring up the full stack locally

```bash
# 1. Shared Postgres + Redis (single instance; correction service gets its own schema)
docker compose -f infra/docker-compose.yml up -d postgres redis

# 2. Go monolith (auth, user, essay, streak, social, gamification)
cd backend/app && go run ./cmd/server

# 3. Correction service (grading engine only — no public auth surface after this refactor)
cd corretor-redacao && make db-migrate && make up   # api + worker

# 4. Mobile app
cd mobile && pnpm install && pnpm start
```

## Smoke-test the essay → grade loop end to end

```bash
# Submit (as an authenticated free-tier test user)
curl -X POST http://localhost:8080/v1/essays \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"prompt_theme_title":"...","prompt_theme_context":"...","essay_text":"..."}'
# → 202 { "id": "...", "status": "pending" }

# Poll until graded
curl http://localhost:8080/v1/essays/<id> -H "Authorization: Bearer $TOKEN"
# → 200 { "status": "graded", "overall_score": ..., "competencies": [...] }

# Confirm second same-day submission is blocked on free tier
curl -X POST http://localhost:8080/v1/essays -H "Authorization: Bearer $TOKEN" ... 
# → 429 { "error_code": "quota_exhausted" }

# Confirm streak advanced
curl http://localhost:8080/v1/streaks/me -H "Authorization: Bearer $TOKEN"
# → 200 { "current_streak": 1, ... }
```

## Run the full test/coverage suite locally (mirrors CI)

```bash
# Go — same check CI runs (backend-ci.yml), floor tracked in backend/scripts/check-coverage.sh
bash backend/scripts/check-coverage.sh

# Python correction service — floor tracked in corretor-redacao/pyproject.toml [tool.coverage.report]
cd corretor-redacao && pytest tests/unit tests/integration --cov=src --cov-report=term-missing

# Mobile — floor tracked in mobile/jest.config.js coverageThreshold
cd mobile && pnpm test -- --coverage --runInBand
```

All three floors are provisional baselines as of the CI/CD rollout (specs/014-constitution-alignment-refactor/tasks.md US4), not yet the constitution's 90% target — see the comments in each config file. Polish task T060 raises them to 90% once US1–US3 land their test suites.

## Install the pre-commit gate

```bash
pip install pre-commit   # or: pipx install pre-commit
pre-commit install       # wires .pre-commit-config.yaml into .git/hooks/pre-commit
```

After this, every local commit runs `gofmt`/`golangci-lint`, `ruff`/`mypy`,
and `eslint`/`tsc` against changed files, and blocks the commit on failure
(Principle VI). The full test + coverage suites above still run in CI on
every pull request, independent of what pre-commit catches locally.
