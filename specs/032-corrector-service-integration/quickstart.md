# Quickstart: Verifying Corrector Service Integration

This is the manual/scripted verification for spec User Story 1's
acceptance scenarios once this feature is implemented — a single
bring-up command producing a working, end-to-end essay-grading pipeline.

## Bring up the stack

```bash
make dev
```

Expected: postgres + redis + monolith + gateway (existing behavior),
**plus** `corrector-api` and `corrector-worker` now start automatically,
both connecting to the same `postgres` service. No second `docker
compose` invocation, no separate `cd ai-corrector`.

## Verify no standalone footprint remains

```bash
docker compose -f infra/docker-compose.yml ps
# Expect: postgres, redis, gateway, monolith, corrector-api, corrector-worker
# Expect NOT to see: any "corretor-postgres" / "corretor-network" container

docker network ls | grep corretor
# Expect: no output — the old isolated network no longer exists
```

## Verify the schema lives in the shared database

```bash
docker compose -f infra/docker-compose.yml exec -T postgres \
  psql -U preuni -d preuni -c "\dn" | grep -E "correction|essay"
# Expect both `correction` and `essay` schemas listed, same database
```

## End-to-end grading smoke test

1. Register/login a test user via the monolith (existing auth flow).
2. Submit an essay via the monolith's essay-submission endpoint.
3. Poll the submission's status.

```bash
curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"quickstart@example.com","password":"Passw0rd!","display_name":"Quickstart"}'
# -> {"student_id": "...", "access_token": "...", ...} (token usable immediately, no verify-email gate on this endpoint)

curl -s -X POST http://localhost:8080/v1/essays \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"prompt_theme_title": "...", "prompt_theme_context": "...", "essay_text": "..."}'
# -> {"id": "...", "status": "pending", "submitted_at": "..."}

# Poll:
curl -s http://localhost:8080/v1/essays/$ID \
  -H "Authorization: Bearer $TOKEN"
```

Note: without a real LLM provider running (no `ollama` container in this
compose stack by default), expect the terminal state to be a typed
`failed` with `error_code: provider_unavailable` rather than `graded` —
that still fully validates the bridge (enqueue → claim → typed failure →
reconcile all completing in well under a second). `graded` requires
`CORRECTOR_OLLAMA_BASE_URL` (or the Ollama Cloud vars) to actually point
at a running provider.

Expected: status transitions from `pending` to `graded` (or a typed
`failed`) within the reconciler's `GradingTimeout` (10 minutes, almost
always seconds in practice), with zero manual steps — no manually
starting a worker container, no manually running a migration.

## Health checks

```bash
curl -s http://localhost:8000/healthz  # if corrector-api's port is
                                        # exposed for local debugging;
                                        # otherwise check via
                                        # `docker compose exec corrector-api curl ...`
                                        # — it must NOT be reachable from
                                        # outside the Docker network in
                                        # any non-local environment (FR-004)
```

## Rollback / cutover note

This is a one-way infra cutover (spec Edge Cases): `ai-corrector/deploy/`
is removed, not deprecated-in-place. There is no dual-running mode where
both the old standalone stack and the new shared stack are expected to
work simultaneously.
