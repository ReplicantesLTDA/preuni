# Quickstart: Verifying the Production Deployment

Manual/scripted verification against spec's Success Criteria, once this
feature is implemented (runner registered, tunnel configured, prod
compose running on the NAS).

## SC-001: External reachability

From a network that is NOT the NAS's home network (mobile data, a
different wifi):

```bash
curl -s -o /dev/null -w "%{http_code}\n" https://preuni.com.br/health
```

Expect `200`, valid TLS certificate (issued via Cloudflare, not
self-signed), and confirm no port is forwarded on the home router at
all — the tunnel is the only path in.

## SC-002: Deploy round-trip

1. Make a trivial, observable change (e.g. bump a version string
   surfaced by the health endpoint).
2. Open a PR into `dev`, merge once green.
3. Open the `dev` → `main` promotion PR (research.md R3), merge once
   green and approved.
4. Watch `.github/workflows/deploy-prod.yml` run against the
   `self-hosted-nas` runner in the Actions tab.
5. Re-run the `curl` from SC-001 and confirm the version string changed
   — with zero manual commands run on the NAS itself.

## SC-003: Essay grading works in production

Using a real (or disposable test) account against `https://preuni.com.br`:

1. Register/login.
2. Submit an essay.
3. Poll until it reaches `graded` (or a typed `failed` if intentionally
   testing failure handling) — same flow already proven in
   `032`/`034`'s local verification, now against the production
   deployment.

```bash
curl -s -X POST https://preuni.com.br/v1/essays \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"prompt_theme_title": "...", "prompt_theme_context": "...", "essay_text": "..."}'
```

## SC-004: No secrets in git history

```bash
gitleaks detect --source . --log-opts="--all" 2>&1 | tail -20
```

Expect clean against the existing baseline — this feature introduces no
new secret category into the repo (all production secrets live on the
NAS or in GitHub Actions Secrets, per data-model.md).

## SC-005: Reboot recovery

1. Reboot the NAS (simulating power loss).
2. Wait for it to come back online.
3. Re-run the SC-001 `curl` — expect the stack to have restarted on its
   own (compose `restart: unless-stopped` / TrueNAS SCALE's own
   app-restart-on-boot behavior) with no manual container start.

## Rollback note

Per contracts/deploy-pipeline.md: a failed deploy leaves the prior
working containers' images available. Manual recovery in the worst case
is `docker compose -f infra/docker-compose.prod.yml up -d` against the
last-known-good build on the NAS — not a full redeploy from GitHub.
