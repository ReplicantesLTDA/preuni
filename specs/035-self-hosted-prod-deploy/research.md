# Research: Self-Hosted Production Deployment

## R1: Cloudflare Tunnel configuration mode

**Decision**: Use a **remotely-managed tunnel** — create the tunnel and
its public-hostname routing (`preuni.com.br` → the `gateway` container's
internal address/port) entirely through the Cloudflare dashboard, then
run `cloudflared` in the prod compose stack with just a tunnel token
(`cloudflared tunnel run --token $CLOUDFLARE_TUNNEL_TOKEN`), no local
`config.yml` committed or maintained.

**Rationale**: A locally-managed tunnel (`cloudflared tunnel create` +
a checked-in `config.yml` mapping ingress rules) is the "more powerful"
option but adds a config file to keep in sync with the dashboard's DNS
records anyway (Cloudflare still needs a CNAME pointing at the tunnel).
For one hostname routing to one internal service, the token-only mode is
strictly simpler with no loss of capability at this scale, and the token
itself is the only secret involved — fits FR-007/SC-004 (no committed
secrets) cleanly since there's no config file with routing details to
even consider committing.

**Alternatives considered**:
- *Locally-managed tunnel with committed `config.yml`*: rejected —
  more files to maintain for identical behavior at this scale; routing
  changes would need a redeploy instead of a dashboard edit.
- *`cloudflared` running directly on TrueNAS as a native app instead of
  in the compose stack*: rejected — keeps ingress lifecycle tied to the
  same `docker compose up`/`down` as everything else it's tunneling to,
  so it comes up and down atomically with the app instead of being a
  separately-managed piece.

## R2: Self-hosted GitHub Actions runner deployment

**Decision**: Run the GitHub Actions runner as its own long-lived
container on the NAS (e.g. `myoung34/github-runner` or the official
runner image wrapped similarly), registered against the repository with
label `self-hosted-nas`, **separate from** (not inside)
`infra/docker-compose.prod.yml` — the runner must keep running and stay
reachable to GitHub even while the app stack it deploys is being brought
down and back up, so it can't be a service *within* the stack it manages.
It needs the Docker socket mounted (to run `docker compose` commands
against the host's Docker daemon) and a full clone of the repo on each
job (standard runner behavior).

**Rationale**: Directly satisfies FR-004/FR-005 (deploy happens without
manual NAS access, survives temporary unreachability — the runner
reconnects and picks up queued jobs on its own, standard GitHub Actions
runner behavior) and the CGNAT constraint (runner-initiated outbound
connection to GitHub, no inbound needed, same shape as `cloudflared`).

**Alternatives considered**:
- *Push-based deploy (GitHub-hosted runner SSHes into the NAS)*:
  rejected outright — impossible without inbound reachability, which
  CGNAT rules out entirely (spec's core constraint).
- *Tailscale + GitHub-hosted runner reaching the NAS over the mesh
  VPN*: viable alternative achieving the same outbound-only property,
  but adds a second piece of infrastructure (Tailscale) and a second
  account/service dependency for no benefit over a self-hosted runner,
  which needs nothing beyond outbound HTTPS to GitHub — already true of
  every other outbound connection this NAS makes. Rejected as
  unnecessary complexity (Principle VII's "no enterprise security
  theater" reasoning applies to infrastructure choices generally, not
  only secret management).

## R3: `dev` → `main` promotion mechanism

**Decision**: A regular pull request from `dev` into `main`, opened
manually (or via a one-line `gh pr create` helper) when a release is
ready — reusing the exact same PR + required-status-checks + human-
approval gate the constitution already mandates for `main`
(`infra/branch-protection.md`), now actually enforced (see R6). No new
tooling; this is the same promotion pattern already used by every
feature branch merging into `dev`, just one level up.

**Rationale**: Satisfies FR-004a ("defined, repeatable way to promote")
with zero new process to learn or maintain — it's the same mental model
already in daily use. A merge commit (not squash) preserves `dev`'s full
history in `main`, useful for later auditing what shipped when.

**Alternatives considered**:
- *Git tags trigger deploy instead of a `main` merge*: viable, but adds
  a second signal (tag *and* branch state) to reason about for no
  benefit over "the state of `main` is what's in production," which is
  simpler to reason about and matches FR-004's literal wording.
- *Automatic dev→main promotion on every green `dev` build*: rejected —
  contradicts the spec's explicit framing of promotion as "a deliberate
  release step, not an automatic side effect of every `dev` merge"
  (FR-004a).

## R4: Backup strategy for the shared production Postgres

**Decision**: Adapt ai-corrector's existing `deploy/backup/` sidecar
(a `pg_dump` cron container, already built and tested for the
correction-only database) to back up the **whole** shared `preuni`
database (all schemas: `auth`, `essay`, `correction`, `user`, etc.), not
just `correction`. Dumps land on a NAS dataset outside the Postgres
container's own volume — a NAS-level snapshot/backup job (already how
TrueNAS users typically protect data) then covers it like any other
dataset. Daily schedule, 7-day rolling retention as a starting default —
adjustable later, not a hard requirement of this feature (spec
Assumptions explicitly defers the specific policy to this phase, not the
spec itself).

**Rationale**: Reuses already-working code (`pg_dump_cron.sh`,
`deploy/backup/Dockerfile`) instead of writing a new backup mechanism
from scratch — the only change is *what* it dumps (whole database vs.
one schema) and *whose* compose file it lives in
(`infra/docker-compose.prod.yml`, since it's no longer correction-
specific). Satisfies FR-008 ("storage a NAS-level backup process can
reach") without inventing new infrastructure.

**Alternatives considered**:
- *Postgres continuous archiving (WAL-E/WAL-G, point-in-time recovery)*:
  more robust, meaningfully more operational complexity (archive
  storage target, restore tooling) for a single-NAS, no-HA, early-stage
  deployment where daily `pg_dump` is a reasonable and honest tradeoff —
  matches the "upgrade when the product requires it" framing already
  established for the hosting choice itself. Revisit alongside any
  future move off self-hosting.

## R5: Production compose stack shape

**Decision**: `infra/docker-compose.prod.yml` is a sibling file to the
existing `infra/docker-compose.yml`, not a Compose "override" layered on
top of it — the dev file publishes host ports for local access
(`5432:5432`, `8080:8080`, etc.) and uses `build:` contexts pointing at
local source for fast iteration; the prod file publishes **no** host
ports at all (ingress is exclusively through `cloudflared` → `gateway`
inside the compose network) and can still use `build:` contexts too,
since the self-hosted runner has the full repo checked out locally on
every deploy — no image registry needed at this scale.

**Rationale**: A Compose override file (`docker-compose.prod.yml` used
via `-f docker-compose.yml -f docker-compose.prod.yml`) would need to
carefully *remove* the dev file's port publishing, which Compose's
override semantics don't cleanly support (you can add or replace keys,
not subtract a `ports:` list entry). A clean sibling file is simpler to
read as "this is what actually runs in prod," matching this feature's
own goal of no separately-*confusing* production implementation (spec
Edge Cases already commit to it being conceptually identical, not
byte-identical).

**Alternatives considered**:
- *Compose override layering*: rejected for the port-publishing
  conflict above.
- *Building and pushing images to a registry, prod pulls by tag*: more
  conventional at larger scale, but adds a registry dependency (where —
  GHCR? self-hosted?) this single-NAS deployment doesn't need yet since
  the runner already has source access. Revisit if multi-node ever
  happens.

## R6: `main` branch protection

**Decision**: Apply `infra/branch-protection.md`'s already-documented
`gh api` command against `main` for real, updating its one stale
reference (the doc currently names `001-enem-prep-platform` as "this
repo's main branch for PRs," left over from before `dev` became the
actual default) to reflect that `main` is now the release branch this
feature makes meaningful for the first time.

**Rationale**: Constitution Principle VI already requires this
unconditionally ("main is protected... no exceptions"); it was simply
never applied because `main` had no real use until now. Not new scope —
enforcing an existing, already-documented requirement.

**Alternatives considered**: None — this is a compliance gap closure,
not a design choice with tradeoffs.

## R7: Production secret source of truth

**Decision**: GitHub Actions repository secrets are the source of truth
for every production credential (`POSTGRES_PASSWORD`, `JWT_SIGNING_KEY`,
`CORRECTOR_OLLAMA_CLOUD_API_KEY`, `CLOUDFLARE_TUNNEL_TOKEN`,
`BACKUP_S3_*`), not a hand-maintained file on the NAS. `deploy-prod.yml`
writes `infra/.env` on the NAS fresh on every deploy, sourced from those
secrets, before bringing the stack up — rotating a credential is a
GitHub Settings change, not an SSH session.

**Rationale**: Directly requested by the user in preference to the
original plan (T009: SSH into the NAS once and hand-edit `infra/.env`
permanently). GitHub Actions Secrets are encrypted at rest,
access-controlled by repo permissions, auto-masked in workflow logs
(including on self-hosted runners — masking is applied by the runner
process itself, not GitHub-hosted infrastructure specifically), and
cost nothing — proportionate to this project's scale per the
constitution's Security & Secrets principle (a real secret-vault
product like HashiCorp Vault or Azure Key Vault is explicitly rejected
there as unnecessary; GitHub's built-in secret store is the lightweight
equivalent already available for free, not the class of tooling that
principle warns against).

**Consequence, stated plainly**: this does not eliminate the secret
existing as a plaintext file on the NAS's disk — `docker compose` has
no way to read a value directly out of GitHub's vault, only the
workflow step that runs on the NAS can, and it has to materialize the
value into a file the containers read. What changes is where the
*source of truth* lives (GitHub, one auditable place) versus where the
value is merely *materialized* at runtime (the NAS, same as before).

**Alternatives considered**:
- **Hand-maintained `infra/.env` on the NAS, edited over SSH**
  (original plan) — rejected per the user's explicit preference; no
  rotation audit trail, requires NAS access for every credential
  change.
- **A real secret-vault product** (HashiCorp Vault, Doppler,
  Infisical) — rejected as disproportionate to a single-NAS,
  single-operator deployment; same reasoning already applied to
  rejecting a `preuni_monolith` DB role in `032`'s research.md R3.

## R8: Avatar object storage — MinIO, self-hosted

**Decision**: Add a self-hosted MinIO (S3-API-compatible) service to
both `infra/docker-compose.yml` and `infra/docker-compose.prod.yml`,
and implement real presigned-URL avatar upload against it (the Go SDK
`github.com/minio/minio-go/v7`, a new `backend/app/internal/storage`
package). In production, MinIO gets a *second* public hostname
(`storage.preuni.com.br`) routed through the same Cloudflare Tunnel as
`gateway`, since the uploading client PUTs directly to the presigned
URL — the backend never proxies the bytes, so that URL must be reachable
by the client, not just internally.

**Rationale**: Raised by the user mid-implementation, in preference to
AWS S3, for the same self-hosting/cost-minimizing reason as the NAS
itself. Investigating `avatar.go` first revealed it was never a real S3
integration at all — a hardcoded stub URL (`X-Amz-Credential=PRESIGNED`
as a literal string, no AWS SDK dependency anywhere in the codebase),
so this isn't "swap providers," it's implementing avatar upload for the
first time. `minio-go` is S3-API-compatible enough that a future move
to real AWS S3 (if self-hosting is ever outgrown) is an endpoint/config
change, not a rewrite.

**A real finding while implementing**: `minio-go`'s
`PresignedPutObject` is not purely local signing by default — without
an explicit `Region`, it calls `GetBucketLocation` over the network on
first use per bucket, which would have made the whole feature secretly
depend on live MinIO connectivity at request time (and broken unit
tests running with no MinIO available). Fixed by hardcoding
`Region: "us-east-1"` in the client (MinIO's own default region
regardless of where it's actually hosted — no real geographic meaning,
purely there to skip the network call).

**Backup storage note**: the same self-hosted MinIO instance could also
host the `backup` sidecar's `pg_dump` output (a second bucket), removing
the need for any external S3-compatible provider at all. Not done here
— `ai-corrector/deploy/backup/pg_dump_cron.sh`'s `s3cmd` invocation uses
virtual-hosted-style addressing (`%(bucket)s.${endpoint}`), which works
against real AWS S3 but not against MinIO without a real fix (path-style
addressing, `--no-ssl` for the internal in-cluster connection) that this
session had no way to verify end-to-end. `PROD_BACKUP_S3_*` remains an
open, unset secret rather than shipping an unverified change.

**Alternatives considered**:
- **AWS S3** — what the stub code assumed; rejected per the user's
  explicit self-hosting preference.
- **Proxying avatar bytes through the monolith instead of presigned
  direct-upload** — avoids needing a second public hostname/tunnel
  route, but adds backend load and complexity for no benefit; presigned
  direct-upload is the standard pattern the stub code was already
  designed around.
