# Contract: Production Deploy Pipeline

Not an HTTP/API contract — the interface here is the trigger condition,
inputs, and guarantees of the CD workflow, since that's the thing other
parts of the process (the promotion step in R3, the runner in R2) depend
on having a stable shape.

## Trigger

`.github/workflows/deploy-prod.yml` runs on `push` to `main` only. It
does **not** re-run CI (lint/test/type-check) — those already ran and
were required to pass before the PR that produced this push was
mergeable (both the `dev`-integration PR and the `dev`→`main` promotion
PR go through the same required-status-checks gate once R6's branch
protection is applied). The deploy workflow's only job is bringing the
already-validated code up on the NAS.

## Runner

Targets `runs-on: [self-hosted, self-hosted-nas]` — the label the NAS's
registered runner carries (research.md R2). If no runner with that label
is online, the job queues (GitHub Actions default behavior) rather than
failing outright — satisfies FR-005 (temporary NAS unreachability
doesn't lose the deploy).

## Steps (contract, not final implementation)

1. Checkout `main` at the pushed commit.
2. Write `infra/.env` fresh from GitHub Actions repository secrets
   (research.md R7) — the NAS's `.env` is never hand-edited; this step
   is its only source. `chmod 600` afterward.
3. Build (or rebuild changed) images via
   `docker compose -f infra/docker-compose.prod.yml build`.
4. Bring up the stack:
   `docker compose -f infra/docker-compose.prod.yml up -d`.
5. Health-check the `gateway`/`monolith` health endpoint from inside the
   compose network (`docker compose exec gateway wget ...` — `gateway`
   publishes no host port, so this can't be a plain `curl` against the
   runner's own `localhost`) before declaring success.
6. On health-check failure: leave the previous containers' images
   available (do not prune) so a manual `docker compose up -d` against
   the last-known-good build can recover without a rebuild — satisfies
   FR-006 (a bad deploy doesn't take down a working one). Automatic
   rollback is explicitly not required by the spec at this stage; the
   guarantee is "doesn't get worse," not "self-heals to the exact prior
   state."

## Inputs the workflow depends on existing already (not created by it)

- Every `PROD_*` GitHub Actions repository secret listed in
  `data-model.md`'s secrets table — the workflow's step 2 is what turns
  these into the NAS's real `infra/.env`; nothing on the NAS itself
  needs provisioning by hand (research.md R7).
- The registered, online self-hosted runner (research.md R2).
- The Cloudflare Tunnel already routing `preuni.com.br` to the
  `gateway` service (research.md R1) — this workflow doesn't touch
  tunnel configuration at all; it only affects what's running behind it.

## Guarantees

- A push to `main` results in that commit's code running on the NAS
  without any person operating the NAS directly (FR-004).
- A transient NAS/runner outage delays but does not lose a deploy
  (FR-005).
- A deploy that fails its health check does not remove the
  previously-working containers (FR-006).
