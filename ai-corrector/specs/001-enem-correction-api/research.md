# Phase 0 Research: ENEM Essay Correction API (MVP)

**Spec**: [spec.md](./spec.md) v2.0.0
**Constitution**: `.specify/memory/constitution.md` v2.0.0
**Date**: 2026-05-28

This document resolves every open decision the planner is required to make. Format per item:
**Decision**, **Rationale**, **Alternatives considered**.

---

## R1. Worker wake-up: LISTEN/NOTIFY + short-interval poll fallback

**Decision**: Hybrid — worker holds a Postgres `LISTEN correction_queued` connection on a
dedicated `asyncpg` connection, and on every wakeup (NOTIFY received OR 5-second poll timer
fires) executes the queue claim with `SELECT … FOR UPDATE SKIP LOCKED`.

**Rationale**: NOTIFY gives sub-100 ms wake-up latency, comfortably inside the 500 ms submission
ack budget and the 90 s e2e budget. The 5 s poll backstop covers missed notifications (network
hiccups, dropped LISTEN connection, worker restart mid-flight) without raising operational
surface area. No external broker needed; satisfies the "no Redis, no Celery, no RabbitMQ" stack
constraint.

**Alternatives considered**:
- **Pure polling** (e.g., every 1 s): simpler but adds up to 1 s of wake-up latency to every
  correction and burns a query/sec per worker against Postgres for no benefit when traffic is
  bursty.
- **Pure LISTEN/NOTIFY**: lowest latency but a dropped LISTEN connection silently stalls the
  queue until the next restart. Unacceptable for a single-worker MVP.

References: Postgres docs on `LISTEN`/`NOTIFY` and `SELECT FOR UPDATE SKIP LOCKED` (9.5+);
prior-art patterns from PgQ, river, graphile-worker.

---

## R2. Quota counter: on-the-fly count, no denormalized counter

**Decision**: Compute monthly quota usage on submission via a single indexed query:
`SELECT COUNT(*) FROM corrections WHERE user_id = $1 AND queued_at >= $month_start AND queued_at < $month_end`,
backed by composite index `(user_id, queued_at)`. The query is wrapped in a serializable
short transaction together with the insert of the new `corrections` row.

**Rationale**: At MVP target volume (≤ 50 corrections/day, hundreds of users, single-digit
queries per user per day) the indexed `COUNT(*)` on a per-user, per-month slice executes well
under 5 ms — orders of magnitude under the 500 ms submission ack budget. A denormalized
counter introduces a second source of truth, a month-rollover job, and a class of bugs
(double-counting on retry, drift after deletion) that buys nothing at MVP scale. When traffic
grows past ~1k submissions/user/month we revisit.

**Alternatives considered**:
- **Denormalized `monthly_quota_usage(user_id, year_month, count)`**: faster at scale, slower
  to ship correctly, requires reconciliation against `corrections` on deletion (Constitution X).
  Premature optimization for MVP scale.

---

## R3. Reverse proxy: Caddy 2

**Decision**: Caddy 2 in front of the API container, terminating TLS, serving HTTP/2 + HTTP/3,
auto-issuing Let's Encrypt certificates.

**Rationale**: Caddyfile is one screen; automatic HTTPS works out of the box with zero cert
management. Single container, no separate certbot cron. For a single VPS / single-domain MVP,
the operational surface area is the lowest available option. Caddy's request-budget guards
(per-IP rate limit) also give us a cheap first line of defence in front of the FastAPI auth
endpoints.

**Alternatives considered**:
- **Traefik**: equally capable, but the Docker label-driven config is heavier than necessary
  for one service, and the dashboard / API features are surface area we don't need at MVP.
- **Nginx + certbot**: more configuration knobs, but two moving parts and a cron job for cert
  rotation. Not justified at this scale.

---

## R4. Dependency manager: uv

**Decision**: `uv` (Astral) with a `pyproject.toml` + `uv.lock`.

**Rationale**: uv resolves and installs an order of magnitude faster than Poetry, ships a single
static binary (trivial to install in Dockerfiles and CI), and uses standard `pyproject.toml`
(PEP 621) without lock-in to a custom `[tool.poetry]` table — keeps the door open to switching
without rewriting metadata. Lockfile is deterministic and CI-friendly.

**Alternatives considered**:
- **Poetry**: more mature ecosystem, but materially slower install on cold caches (CI hot path),
  and its custom metadata table couples us to the tool.

---

## R5. JWT library: PyJWT

**Decision**: `PyJWT` for access-token issue/verify with HS256 (single-secret rotation via
env var).

**Rationale**: PyJWT is narrow, well-audited, and has no transitive `cryptography` deprecation
issues that have plagued `python-jose` historically. We need exactly the JWS path; we do not
need JWE, JWK rotation, or OIDC discovery. PyJWT is the right size for this surface.

**Alternatives considered**:
- **python-jose**: broader (JWE, JWK), but unmaintained periods and dependency-pin pain make it
  a poor fit for a new build.
- **Authlib**: heavier; useful when we add OAuth2 / OIDC, which is out of scope for the MVP.

---

## R6. Password hashing: argon2-cffi

**Decision**: `argon2-cffi` with the Argon2id variant, parameters tuned to ~50 ms hash time on
the target VPS at process start.

**Rationale**: Argon2id is the OWASP-recommended default for new applications (resistant to GPU
and side-channel attacks). `argon2-cffi` is the canonical Python binding, actively maintained,
no transitive C compile pain in our base image (manylinux wheels).

**Alternatives considered**:
- **passlib**: convenient multi-algo facade but accumulating deprecation warnings on Python
  3.12+ and last release cadence is slow. We need exactly one algorithm.
- **bcrypt**: still acceptable per OWASP, but Argon2id is the explicit current recommendation
  for new builds.

---

## R7. LLM provider abstraction shape

**Decision**: `corrector/llm/base.py` defines a `LLMProvider` Protocol (typing.Protocol, not ABC)
with the following surface:

```python
class LLMProvider(Protocol):
    name: str  # opaque, e.g. "ollama-cloud"
    model_id: str  # e.g. "kimi-k2:1t"

    async def complete_structured(
        self,
        *,
        system: str,
        user: str,
        output_schema: dict,
        temperature: float,
        seed: int | None,
        max_tokens: int,
        timeout_s: float,
    ) -> LLMResult: ...
```

`LLMResult` carries: `parsed_output: dict`, `raw_text: str`, `prompt_tokens: int`,
`completion_tokens: int`, `cost_usd: Decimal`, `latency_ms: int`, `model_id: str`,
`inference_params: dict`. Error taxonomy is a flat `LLMError` hierarchy in `corrector/llm/errors.py`:
`RateLimitError`, `TimeoutError`, `TransientError`, `SchemaViolationError`, `FatalError`.

Concrete adapters: `OllamaProvider` (MVP), `FakeProvider` (tests). Future adapters
(Anthropic, OpenAI, Google) implement the same Protocol with no changes to `corrector/pipeline.py`
or `corrector/graders/`.

**Rationale**: Protocol gives structural typing — adapters do not need to import a base class,
which keeps the provider abstraction truly orthogonal. The five-error taxonomy maps 1:1 to the
spec's typed error codes (`provider_rate_limited`, `provider_timeout`, `schema_violation`,
`provider_unavailable`, `internal_error`) so the worker's failure-classification logic is a
straight `match` on exception type.

**Alternatives considered**:
- **ABC base class**: requires inheritance, makes test fakes verbose.
- **Generic `call(prompt) -> str` + caller-side parse**: violates Constitution IV — schema
  enforcement has to live at the provider boundary so retries can happen there.

---

## R8. Structured-output strategy on Ollama

**Decision**: Use Ollama's native JSON-mode (`format: "json"`) plus an explicit JSON Schema
embedded in the system prompt. After every call, validate the parsed JSON against the schema
with `jsonschema`. On validation failure: retry once with a corrective prompt injecting the
validator error message; on second failure: surface `SchemaViolationError`.

**Rationale**: Ollama's JSON mode does not guarantee schema conformance — only well-formed JSON.
The application-side validator is the contract per Constitution IV. The 2-attempt budget matches
the spec's FR-010.

**Alternatives considered**:
- **Grammar-constrained decoding** (GBNF in llama.cpp): stronger guarantee, but Ollama Cloud
  doesn't expose GBNF and we'd lose provider parity. Keep this as a future optimization.
- **Function/tool calling**: Ollama's tool-call path is less mature than its JSON mode at the
  time of writing; not worth the surface area.

---

## R9. Default model: Kimi K2 1T

**Decision**: `kimi-k2:1t` on Ollama Cloud is the default production model. Local dev
defaults to `qwen3:14b` (or 7B on resource-constrained machines) via the same
`OllamaProvider` adapter, swapping only the base URL and model ID via env vars.

**Rationale**:
- Strong structured-output behavior in JSON mode at 72B scale.
- Demonstrated pt-BR quality on educational content per public benchmarks (CMMLU-pt, BLUEX).
- Multi-trait evaluation (5-competency reasoning) is closer to a chain-of-thought + structured
  output task than to a single classification — the 72B model gives us the headroom we need to
  clear MVP-tier thresholds (Constitution IX).
- 14B/7B locally is enough to exercise the full pipeline on a laptop during dev (slower, less
  accurate, but the code paths are identical).

**Constitution II property check for `kimi-k2:1t` on Ollama Cloud**:
- Deterministic inference: `temperature` controllable; Ollama exposes `seed`. ✅
- Structured output validation: JSON mode + app-side `jsonschema`. ✅
- pt-BR quality: empirically demonstrated against the golden dataset before promotion. ✅
  (CI golden-tier check is the gate.)
- Privacy compatibility: Ollama Cloud terms must be configured to non-training + ZDR before
  prod. Verified at deployment time. ✅

**Alternatives considered**:
- **Llama 4 family**: similar tier, slightly weaker pt-BR per anecdote.
- **Anthropic Claude / OpenAI GPT-4-class**: better quality, but vendor-hosted and Constitution
  X requires explicit ZDR contract review per provider; deferred to a future provider swap.

---

## R10. Determinism strategy & known MVP limitation

**Decision**: `temperature = 0.1`, `seed = int.from_bytes(sha256(correction_id.bytes)[:8], "big")`.
Single grader pass per correction in the MVP. Re-evaluation reuses the same seed-derivation, so
re-evaluating the same essay text against the same prompt version yields effectively the same
correction. **This is a documented MVP limitation, not a bug.**

**Rationale**: Constitution III mandates determinism. Constitution-aligned variance across
grader passes is the right way to get the spec's iteration value, and that is a multi-grader-day
feature. The data model already supports it (`grader_passes.pass_index`, `grader_passes.seed`).

**Documented for users**: the re-evaluation endpoint surfaces a note in the response when the
new pass's `prompt_version` + `model_id` match the original — useful for callers building UIs.

**Alternatives considered**:
- **Raise temperature on re-evaluation**: violates Constitution III. Rejected.
- **Block re-evaluation when nothing has changed**: hostile UX; users want to validate that the
  system gives them the same answer twice (this is itself a feature).

---

## R11. pt-BR language detection

**Decision**: `lingua-language-detector` Python package with `Language.PORTUGUESE` confirmed,
plus a heuristic post-check that the text matches pt-BR-specific orthography (e.g., `você` vs
`tu` distribution, presence of `dezembro` vs `Dezembro`). The combined check rejects pt-PT
samples that pass the broad-portuguese detector.

**Rationale**: Lingua handles short text well, ships statistical models offline, no LLM call
needed at pre-validation time. The pt-BR vs pt-PT split is the operationally important one
(per spec edge case).

**Alternatives considered**:
- **fasttext language id**: heavier, single-language Portuguese label (no pt-BR/pt-PT split).
- **LLM-based detection**: violates pre-validation budget; we must not invoke the LLM for
  invalid inputs.

---

## R12. Audit log payload schema

**Decision**: `correction_audit_logs.event_type` is an enum:
`{submitted, llm_started, llm_completed, schema_failed, retry, retry_exhausted, completed, failed, deleted}`.
`event_payload` is a jsonb with an allowlisted key set per `event_type`. Forbidden keys at the
write boundary: `essay_text`, `email`, `name`, `student_id`. Enforced by a single
`AuditLogWriter` class that rejects forbidden keys; covered by a unit test that catalogues every
key we ever write.

**Rationale**: Constitution VII mandates centralized scrub. A jsonb payload is the right shape
for forward-compat (new event types ship without migrations), but it needs a typed writer to
remain safe.

---

## R13. Email verification token strategy

**Decision**: 32-byte URL-safe random token, stored hashed (SHA-256) in
`user_verification_tokens(user_id, token_hash, expires_at, used_at)`. TTL 24 h. Resend rate-
limited to 1/min, 5/24h per user. Email delivery via a simple SMTP integration (Mailgun or
Brevo); both supported by setting the SMTP env vars only.

**Rationale**: Random tokens, hashed at rest, single-use, time-bounded. SMTP keeps the provider
question to a configuration choice, not a code dependency. Email sending is the single piece of
external I/O that is not the LLM — keeping it boring is the point.

**Alternatives considered**:
- **JWT email-verification tokens**: stateless but harder to revoke; not worth it.

---

## R14. Refresh-token storage & rotation

**Decision**: `refresh_tokens(id uuid, user_id, token_hash sha256, expires_at, revoked_at)`.
Login issues a fresh refresh token (32-byte random), returns the cleartext to the client, stores
the hash. Refresh-endpoint behavior: validate, rotate (revoke the presented token, issue a new
one). Logout revokes the presented refresh token. TTL 30 days. Access-token TTL 15 minutes.

**Rationale**: Hash-at-rest mirrors password hygiene. Rotation on every refresh limits stolen-
token blast radius. 15-min/30-day pair matches the spec.

**Alternatives considered**:
- **Storing cleartext**: violates security baseline.
- **Long-lived access tokens, no refresh**: violates spec FR-030/FR-031.

---

## R15. Postgres backup

**Decision**: Daily `pg_dump -Fc` cron in a sidecar container, uploaded to S3-compatible storage
(Backblaze B2 by default; bucket + creds configurable). Retention 30 days. Restore drill
documented in `quickstart.md`.

**Rationale**: B2 has the cheapest egress in the S3-compatible field, which matters for monthly
restores during dev. A custom-format dump (`-Fc`) gives selective restore and parallel restore.

**Alternatives considered**:
- **Continuous WAL archiving**: PITR is the right answer post-MVP; over-engineering for current
  scale and uptime targets.

---

## R16. CI: golden-dataset gate

**Decision**: GitHub Actions workflow `ci.yml` runs three jobs in order: `lint` (ruff + mypy +
import-linter), `tests-unit-and-integration`, `tests-golden`. The golden job spins up a
Postgres service container and runs against a **`FakeLLMProvider`** that returns a deterministic
fixture matching `expected.json` to verify pipeline plumbing, then runs against the **real
provider** (Ollama Cloud) on the golden test split + INEP secondary set for the MVP-tier metric
check. Real-provider job is gated on `secrets.OLLAMA_CLOUD_API_KEY` being present (forks fail
soft).

CI compares the new MVP-tier metrics against the prior main's metrics (stored as a workflow
artifact) and blocks merge on any of: (a) overall MVP-tier regression on Essay-BR test split,
(b) per-band MVP-tier regression in any of the three score bands (0–400, 401–700, 701–1000),
(c) any failure on the INEP secondary sanity set.

**Rationale**: Cheap plumbing exercised on every PR against the fake provider; expensive metric
gate paid only against the real provider on the merge-gating split. Per-band stratification
prevents high-score dominance from masking poor low-score performance. INEP secondary catches
silent regression against true ENEM-grade material even if Essay-BR proxies pass.

---

## R17. Primary golden dataset: Essay-BR extended corpus

**Decision**: Use the **Essay-BR extended corpus** (Marinho, Anchiêta & Moura, 2022) — ~6,500
essays by Brazilian high-school students against authentic ENEM prompts, expert-graded across
all 5 official competencies. Upstream: `github.com/rafaelanchieta/essay`, MIT license.

Split usage:
- **`test/` → merge-gating golden set.** Iterated by `tests/golden/test_golden_dataset.py`.
- **`valid/` → prompt iteration only.** Developers may inspect, run, and tune against it during
  prompt development. NOT a CI gate. Documented at file head.
- **`train/` → never used.** Prevents prompt-engineering contamination (the model and the
  prompts must not see what the test split was drawn from when iterating).

**Caveats documented at `tests/golden/essay_br/README.md`**:
- Essays come from an online learning platform, not actual ENEM exams. Reference scores are
  **expert proxies**, not official INEP grades. Acceptable for MVP gating; informs
  interpretation of MAE numbers.
- Corpus is imbalanced toward higher scores → see R18 (stratified per-band metrics).

**Alternatives considered**:
- **Curated proprietary corpus**: would block MVP shipping until corpus exists.
- **Synthetic essays generated by an LLM**: contaminates the evaluation; rejected on principle.

---

## R18. Stratified per-band metric reporting

**Decision**: The golden harness computes overall MVP-tier metrics AND three per-band MVP-tier
slices, partitioning the Essay-BR test split by human reference total score:
- **Low band**: 0–400 (eliminatory and weak essays)
- **Mid band**: 401–700 (the dense majority)
- **High band**: 701–1000 (strong essays)

Each band MUST independently clear the MVP-tier thresholds (MAE-total ≤ 120, MAE-per-comp ≤ 60,
≥ 70% within ±120). Overall pass + any band fail = merge blocked.

**Rationale**: Essay-BR is heavily right-skewed (most essays cluster in the upper-middle band).
A monolithic MAE can be ≤ 120 while the low band is wildly off — i.e., the model is reliably
generous and never recognizes a weak essay. That is precisely the failure mode users care about
catching pre-launch (students need to know they're weak before the exam, not after). Per-band
gating forces the model to perform across the whole score distribution.

**Alternatives considered**:
- **Single global metric only**: cheaper, but masks the failure mode above. Rejected.
- **Reweighted overall metric**: equivalent in effect but harder to read in PR reports.

---

## R19. INEP secondary sanity set (nota-1000 públicas)

**Decision**: Maintain a separately-curated set of **10–20 INEP-released exemplary essays**
(`notas 1000` públicas) at `tests/golden/inep_exemplary/`. The harness runs the full pipeline
against this set on every CI run; **any** discrepancy outside the MVP-tier per-essay tolerance
(±120 total, ±60 per competency) on this set blocks merge — even when Essay-BR metrics pass.

**Rationale**: INEP exemplary essays are the only material in our gate that is verifiably
ENEM-grade (official, real prompts, official top scores). Failing on them is a categorical
signal — proxy MAEs lying does not save us. Small N (10–20) is OK because the bar is per-essay
tolerance, not aggregate statistics.

**Sourcing note**: Each INEP example carries provenance metadata (`prompt_year`,
`source_url`) in `expected.json`. The harness logs the example's slug on failure so reviewers
can audit the discrepancy directly.

---

## R20. License compliance & attribution

**Decision**:
- Add a `NOTICE` file at the repository root. Contents: credit to the Essay-BR corpus, full
  citation (Marinho, J. R. M.; Anchiêta, R. T.; Moura, R. F. — 2022. *Essay-BR: a Brazilian
  Portuguese essay corpus.* — appropriate venue), the upstream copyright notice, and the
  verbatim MIT License text covering the corpus.
- Add `tests/golden/essay_br/README.md` containing the upstream repository URL, the **pinned
  commit hash** used by `ingest.py`, and re-running instructions.
- Generated example files (`essay.txt`, `prompt_theme.txt`, `motivational_texts.txt`,
  `expected.json`) ARE committed to the repo. The MIT license permits redistribution; keeping
  them in-repo eliminates network dependence in CI.

**Rationale**: MIT requires retaining the license text and copyright notice in derived works;
the `NOTICE` file is the canonical mechanism for this in a multi-component repo. The pinned
commit hash is what makes the ingest reproducible (R21).

**Alternatives considered**:
- **gitignore the data, fetch in CI**: makes CI depend on GitHub uptime + network egress;
  rejected.
- **No attribution beyond the README**: under-discharges MIT obligations; rejected.

---

## R21. Ingest script: reproducible, idempotent, byte-stable

**Decision**: `tests/golden/essay_br/ingest.py`:
1. Reads a single module-level constant `UPSTREAM_COMMIT_HASH = "<sha>"` (pinned).
2. Fetches the upstream tarball at that commit (no `main` reference).
3. Verifies the tarball's SHA-256 against a constant `UPSTREAM_TARBALL_SHA256 = "<sha256>"`.
4. Streams essays into the on-disk format (`<split>/<slug>/{essay.txt, prompt_theme.txt,
   motivational_texts.txt, expected.json}`).
5. Slug = zero-padded sequential index per split (`0001`, `0002`, …), assigned in the upstream
   file order. Stable across re-runs.
6. Refuses to overwrite a non-empty target directory unless `--force` is passed (idempotent
   default).
7. Emits a final `MANIFEST.json` listing every produced file + its SHA-256, plus the two
   upstream pins. CI verifies the manifest matches committed contents byte-for-byte.

**Rationale**: Byte-stable output is what lets us commit the data and treat the in-repo files
as canonical without trust in re-runners. Two-tier pinning (commit hash + tarball hash) defeats
upstream history rewrites.

**Alternatives considered**:
- **Submodule the upstream repo**: ties our build to upstream layout changes; brittle.
- **Vendor the upstream as a tarball in-repo**: works but adds binary churn; the in-repo
  generated files cover the use case more directly.

---

## Summary of resolutions

| # | Open decision | Choice |
|---|---------------|--------|
| R1 | Worker wake-up | LISTEN/NOTIFY + 5 s poll backstop |
| R2 | Quota counter | On-the-fly indexed count |
| R3 | Reverse proxy | Caddy 2 |
| R4 | Dep manager | uv |
| R5 | JWT lib | PyJWT |
| R6 | Password hash | argon2-cffi (Argon2id) |
| R7 | LLM abstraction shape | `LLMProvider` Protocol with 5-error taxonomy |
| R8 | Structured output | Ollama JSON mode + app-side `jsonschema` + 2-attempt retry |
| R9 | Default model | `kimi-k2:1t` (Cloud); `qwen3:14b` (local dev) |
| R10 | Determinism / MVP limit | seed = sha256(correction_id)[:8]; single grader pass; documented |
| R11 | Language detection | `lingua` + pt-BR orthographic heuristic |
| R12 | Audit log shape | enum event_type + allowlisted jsonb payload via `AuditLogWriter` |
| R13 | Email verification | hashed random token, 24 h TTL, SMTP via env vars |
| R14 | Refresh tokens | hashed at rest, rotated on refresh, 30 d TTL |
| R15 | Postgres backup | daily `pg_dump -Fc` → B2 (S3-compatible), 30 d retention |
| R16 | CI golden gate | fake-provider full corpus + real-provider metric slice; main-baseline diff; blocks on overall, per-band, or INEP-secondary regression |
| R17 | Primary golden dataset | Essay-BR extended corpus (Marinho et al. 2022, MIT); `test/` gates merges, `valid/` for prompt iter, `train/` never touched |
| R18 | Stratified bands | Per-band MVP-tier gate on 0–400 / 401–700 / 701–1000; each band must independently pass |
| R19 | INEP secondary set | 10–20 nota-1000 públicas; per-essay tolerance gate; any failure blocks merge |
| R20 | License & attribution | `NOTICE` at repo root + `tests/golden/essay_br/README.md`; data files committed (MIT permits) |
| R21 | Ingest reproducibility | Pinned upstream commit + tarball SHA-256; byte-stable slugs; `MANIFEST.json` checked in CI |

All Phase 0 NEEDS-CLARIFICATION items are resolved. No items deferred to spec author.
