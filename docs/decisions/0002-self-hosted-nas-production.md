# ADR-0002: Self-hosted TrueNAS production deployment via Cloudflare Tunnel

**Status**: accepted
**Date**: 2026-08-24

## Context

preuni had never been deployed anywhere (GitHub issue #68) — `main` was
still literally the Specify template's initial commit, `infra/docker-
compose.yml` was explicitly dev-only, and no hosting decision had been
made. Separately, ai-corrector's `deploy/docker-compose.prod.yml` (issue
#69) still assumed it was a standalone product with its own Postgres and
JWT auth, because it had never been rewritten against a real production
target — because none existed. Resolving #69 properly required #68 to
be resolved first.

The user's own constraint set: minimize cost at this pre-launch stage
("when the product requires it, I'll upgrade to a more robust
environment" — an explicitly temporary choice, not a permanent
architecture bet), self-host on hardware already owned (a TrueNAS SCALE
NAS), and that NAS sits behind CGNAT with no option to forward inbound
ports from its home router.

## Decision

1. **Host on the user's TrueNAS SCALE NAS**, running the same Docker
   Compose stack shape already proven in local dev
   (`032-corrector-service-integration`) via a new
   `infra/docker-compose.prod.yml`.
2. **Cloudflare Tunnel** (`cloudflared`, token-only/remotely-managed
   mode) for ingress — the NAS makes an outbound-only connection to
   Cloudflare's edge, so CGNAT and the lack of port forwarding are
   irrelevant. `preuni.com.br` (already owned) routes through it.
3. **A self-hosted GitHub Actions runner on the NAS itself** for CD —
   same outbound-only property, so a push to `main` can trigger a
   deploy without the NAS ever needing to accept an inbound connection
   from anywhere, including from GitHub.
4. **`main` revived as the release branch** that triggers deploys;
   `dev` remains the ordinary integration branch every feature PR
   already targets. Promotion is a plain PR from `dev` into `main`, no
   new tooling.
5. ai-corrector's production deployment is rewritten to match — no
   standalone Postgres, no JWT auth, no public port, sharing the
   monolith's Postgres via the same `correction`-schema bridge already
   proven correct.

## Alternatives considered

- **A managed cloud VM/PaaS (Railway, Fly.io, a small VPS)** — the
  conventional choice, and one the user may move to later. Rejected for
  *now* specifically because it costs money at a stage where the product
  hasn't validated it needs to spend any — the NAS is already owned and
  otherwise idle capacity. This is an explicitly reversible choice, not
  a permanent architectural commitment.
- **Port forwarding instead of a tunnel** — impossible; the ISP's CGNAT
  gives the NAS no public IP of its own to forward a port on.
- **Tailscale mesh VPN + a GitHub-hosted runner reaching the NAS over
  it, instead of a self-hosted runner** — achieves the same outbound-
  only property but adds a second infrastructure dependency (a VPN
  service and its own account) for no benefit over a self-hosted
  runner, which needs nothing beyond outbound HTTPS to GitHub — already
  true of every other outbound connection this NAS makes.
- **A locally-managed Cloudflare Tunnel with a committed `config.yml`**
  — more powerful, but for one hostname routing to one internal service
  it's strictly more to maintain than the dashboard-managed, token-only
  mode for no capability gained at this scale.
- **Automatic `dev`→`main` promotion on every green `dev` build** —
  rejected; a production release should be a deliberate step, not an
  automatic side effect of ordinary feature-branch merges.

## Consequences

**Easier**: preuni finally has a real, reachable production instance;
ai-corrector's deployment config finally matches its actual architecture
instead of resurrecting a second identity system; deploys are automatic
and don't require the user to operate the NAS by hand; the whole
production stack is defined in one compose file that mirrors dev,
keeping "what actually runs in prod" legible.

**Harder / accepted tradeoffs**: no high availability — the NAS is a
single point of failure, a home-network or hardware outage takes
preuni down, and there is no automatic failover. Backups rely on a
`pg_dump` sidecar plus NAS-level snapshotting, not continuous
archiving/point-in-time recovery — acceptable for a pre-validation
stage, revisit once real user data volume justifies stronger
guarantees. Cloudflare's tunnel service itself becomes a dependency and
single point of failure for ingress specifically — an accepted tradeoff
of the cost-minimizing, self-hosted approach, not something this
decision attempts to mitigate.

**Forecloses, for now**: horizontal scaling (single NAS, single Postgres
instance) and true zero-downtime deploys (a health-check-gated deploy
with no automatic rollback is "doesn't get worse," not "self-heals").
Neither is a hard architectural wall — both are addressed by migrating
off self-hosting later, which this decision explicitly anticipates
rather than forecloses permanently.
