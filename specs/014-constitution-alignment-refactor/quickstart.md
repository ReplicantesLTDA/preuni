# Quickstart: Constitution Alignment Refactor

## Bring up the full stack locally

```bash
# 1. Shared Postgres + Redis (single instance; correction service gets its own schema)
docker compose -f infra/docker-compose.yml up -d postgres redis

# 2. Apply migrations — both sets, against the same database (research.md #1).
#    infra/migrations/<schema>/*.sql are plain numbered SQL files, applied in
#    order with psql (not golang-migrate — see backend-ci.yml's comment on
#    why infra/scripts/migrate.sh's docstring doesn't match reality).
for schema in auth user content learning simulation dissertation essay social gamification; do
  for f in infra/migrations/"$schema"/*.sql; do
    [ -e "$f" ] && psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$f"
  done
done
(cd corretor-redacao && alembic upgrade head)   # creates the `correction` schema + correction_jobs bridge table

# 3. Go monolith (auth, user, essay, streak, social, gamification)
cd backend/app && go run ./cmd/server

# 4. Correction service (grading engine only — no public auth surface after this refactor)
cd corretor-redacao && make up   # api + worker

# 5. Mobile app
cd mobile && pnpm install && pnpm start
```

## Smoke-test the essay → grade → streak → ranking loop end to end

```bash
# Submit (as an authenticated free-tier test user)
curl -X POST http://localhost:8080/v1/essays \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"prompt_theme_title":"...","prompt_theme_context":"...","essay_text":"..."}'
# → 202 { "id": "...", "status": "pending" }

# Poll until graded (the essay reconciler ticks every 5s — cmd/server's
# runEssayReconciler; the correction worker itself must be running, step 4 above)
curl http://localhost:8080/v1/essays/<id> -H "Authorization: Bearer $TOKEN"
# → 200 { "id": "...", "status": "graded", "grade": { "overall_score": ..., "competencies": [...] } }

# Confirm second same-day submission is blocked on free tier
curl -X POST http://localhost:8080/v1/essays -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" -d '{"prompt_theme_title":"...","prompt_theme_context":"...","essay_text":"..."}'
# → 429 { "error": { "code": "QUOTA_EXCEEDED", "message": "..." } }

# Confirm streak advanced
curl http://localhost:8080/v1/streaks/me -H "Authorization: Bearer $TOKEN"
# → 200 { "current_streak": 1, "longest_streak": 1, ... }

# Friends: send a request, accept it as the other user, then list friends
curl -X POST http://localhost:8080/v1/friends/requests -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" -d '{"addressee_id":"<other-user-id>"}'
curl -X POST http://localhost:8080/v1/friends/requests/<id>/accept -H "Authorization: Bearer $OTHER_TOKEN"
curl http://localhost:8080/v1/friends -H "Authorization: Bearer $TOKEN"

# Ranking + medals (weekly score updates after each grade; WeekClose runs
# hourly via cmd/server's runWeekCloseJob and only acts once a week is over)
curl http://localhost:8080/v1/ranking/me -H "Authorization: Bearer $TOKEN"
curl "http://localhost:8080/v1/ranking/weekly?tier=bronze" -H "Authorization: Bearer $TOKEN"
curl http://localhost:8080/v1/medals/me -H "Authorization: Bearer $TOKEN"
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
