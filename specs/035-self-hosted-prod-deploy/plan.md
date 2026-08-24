# Implementation Plan: Self-Hosted Production Deployment

**Branch**: `035-self-hosted-prod-deploy` | **Date**: 2026-08-24 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/035-self-hosted-prod-deploy/spec.md`

## Summary

preuni has never been deployed anywhere (#68) and ai-corrector's production
deploy config still assumes it's a standalone product with its own
Postgres/JWT/API (#69). This plan stands up a real production target — the
user's own TrueNAS SCALE NAS, unreachable by direct inbound connection
(CGNAT) — using Cloudflare Tunnel for ingress against the already-owned
`preuni.com.br`, and a self-hosted GitHub Actions runner on the NAS itself
for CD (outbound-only, so CGNAT doesn't block it either). `main` is revived
as the release branch that triggers deploys; `dev` remains the ordinary
integration branch. The production compose stack is the same
`032-corrector-service-integration` shape already proven in local dev
(shared Postgres, correction-schema bridge, no public port on the
correction service) — which is what finally gives #69 a real target to be
rewritten against instead of inventing one speculatively.

## Technical Context

**Language/Version**: No new application code — infra/ops only (Docker Compose, GitHub Actions YAML, Cloudflare Tunnel config)
**Primary Dependencies**: Docker Compose (already used), `cloudflared` (Cloudflare Tunnel daemon), a self-hosted GitHub Actions runner (official `actions/runner` or an unofficial Docker-wrapped equivalent)
**Storage**: PostgreSQL 16 on a TrueNAS SCALE dataset (persistent, NAS-backed volume) — same single-instance/schema-isolated shape as dev, now with real persistent storage and NAS-level backup reach (FR-008)
**Testing**: No new automated tests (infra/ops feature); verification is the quickstart.md manual/scripted checklist (external reachability, deploy round-trip, essay-grading smoke test, reboot recovery) matching spec's Success Criteria
**Target Platform**: TrueNAS SCALE (Linux-based, native Docker/Apps support — confirmed with user), self-hosted at home, behind CGNAT
**Project Type**: Deployment/CD pipeline addition to an existing web-service stack — no new service, no new UI
**Performance Goals**: N/A — no request-path change; this is where the existing stack runs, not how fast it runs
**Constraints**: No forwardable inbound port (CGNAT) — ingress MUST be an outbound-initiated tunnel; deploy MUST also work outbound-only (self-hosted runner, not push-based SSH); single-node, no HA requirement (spec Assumptions)
**Scale/Scope**: One NAS, one production compose stack (monolith, gateway, shared Postgres, correction-service worker/api, cloudflared), one release branch (`main`)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principle VI (`main` is protected)**: `main` has never actually had
  branch protection applied in GitHub despite `infra/branch-protection.md`
  already documenting the exact `gh api` command to do it (its own text
  notes the command needs the branch name updated from a stale reference).
  This feature is what finally makes `main` meaningful (it becomes the
  deploy trigger), so applying that protection is in scope here — not
  scope creep, it's the same gate the constitution already requires,
  previously unenforced because `main` was unused. Flagged as a task, not
  a violation.
- **Principle VII (Security & Secrets)**: this feature adds a materially
  new secret surface — Cloudflare Tunnel token, GitHub Actions runner
  registration token, production `JWT_SIGNING_KEY`/`POSTGRES_PASSWORD`/
  `OLLAMA_CLOUD_API_KEY`, distinct from their dev-only counterparts. All
  MUST live in gitignored files on the NAS or in GitHub Actions Secrets
  (for whatever the runner itself needs to authenticate), never in the
  repo — matches FR-007/SC-004 exactly, no new exception needed.
  `gitleaks` (already CI-enforced) continues to be the backstop.
  **No enterprise-specific security theater** (Principle VII's closing
  bullet, and the same reasoning `032`'s research.md R3 already applied
  to the unused `preuni_monolith` role): a single self-hosted runner with
  filesystem access to a `.env` file is proportionate to a one-NAS, one-
  operator deployment. No secret-vault product, no SSO, introduced here.
- **Principle II (Testing Standards)**: no new business logic, so no new
  unit/integration test obligation. The "third-party service boundary"
  language extends naturally to Cloudflare Tunnel and the runner, but
  those are ops infrastructure verified by the quickstart checklist
  (external reachability, deploy round-trip), not unit-testable code.
- **Governance (ADR discipline)**: choosing a hosting target (self-hosted
  NAS vs. cloud), an ingress mechanism (Cloudflare Tunnel), and a CD
  mechanism (self-hosted runner) are exactly the "hard-to-reverse,
  security-posture" class of decision the constitution says gets an ADR.
  This plan's Phase 1 includes writing `docs/decisions/0002-*.md` (the
  next available number) capturing this — not deferred to a follow-up.

**Gate result**: PASS. One task generated directly from an existing,
previously-unenforced constitutional requirement (branch protection on
`main`); one new ADR required by Governance, included in this plan's scope
rather than deferred.

## Project Structure

### Documentation (this feature)

```text
specs/035-self-hosted-prod-deploy/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md         # Phase 1 output
├── quickstart.md         # Phase 1 output
├── contracts/
│   └── deploy-pipeline.md   # CD workflow's trigger/secrets/runner contract
└── tasks.md              # Phase 2 output (/speckit.tasks — not created here)
```

### Source Code (repository root)

```text
infra/
├── docker-compose.yml           # unchanged — this is the dev stack
├── docker-compose.prod.yml      # NEW — prod stack: monolith, gateway,
│                                 #       postgres (NAS-backed volume),
│                                 #       corrector-api, corrector-worker,
│                                 #       cloudflared. No dev conveniences
│                                 #       (no host port publishing beyond
│                                 #       what cloudflared needs internally,
│                                 #       which is none — it's outbound-only)
├── .env.prod.example             # NEW — production secret shape (values
│                                 #       never committed; documents what
│                                 #       must exist in the NAS-local .env)
└── nginx/nginx.conf              # unchanged — same gateway config, prod
                                  #       compose reuses it as-is

ai-corrector/
└── deploy/
    ├── docker-compose.prod.yml  # DELETED — superseded by
    │                             #   infra/docker-compose.prod.yml
    │                             #   (closes #69)
    ├── Caddyfile                 # DELETED — cloudflared/the gateway
    │                             #   handle ingress now, no second proxy
    └── backup/                   # kept, generalized: still the pg_dump
                                  #   sidecar pattern, now backing up the
                                  #   shared prod Postgres (all schemas),
                                  #   not a correction-only database

.github/workflows/
└── deploy-prod.yml               # NEW — runs on push to `main` (i.e.
                                  #       after CI already gated the
                                  #       merge into dev/main), targets
                                  #       the self-hosted runner, pulls
                                  #       latest, brings up
                                  #       docker-compose.prod.yml

docs/decisions/
└── 0002-self-hosted-nas-production.md  # NEW ADR (Governance requirement)

infra/branch-protection.md        # updated — apply main's protection for
                                  #   real (script already exists, was
                                  #   never actually run against a live repo)
```

**Structure Decision**: New `infra/docker-compose.prod.yml` (sibling to
the existing dev-only compose, not a replacement — the two audiences
differ) plus one new GitHub Actions workflow and one new ADR. No new
application source directories.

## Complexity Tracking

*No constitution violations requiring justification — table omitted.*
