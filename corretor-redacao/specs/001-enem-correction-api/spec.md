# Feature Specification: Automated ENEM Essay Correction API (MVP)

**Feature Branch**: `001-enem-correction-api`
**Created**: 2026-05-28
**Last Amended**: 2026-05-28
**Spec Version**: 2.0.0 (breaking: sync → async, auth added, B2C-only)
**Status**: Draft
**Input**: User description: "Build an MVP API for automated ENEM essay correction." (v1.0.0) amended by v2.0.0 amendment set (B2C scope, async flow, auth as first-class, MVP-tier golden-dataset gates per Constitution v2.0.0).

<!--
AMENDMENT LOG
=============
v1.0.0 → v2.0.0 (2026-05-28). Breaking change.

A1 — B2C scope: institutional clients / educator surface dropped. Users are
     end-user students only, segmented into Free (3/mo) and Premium (30/mo)
     tiers. Quotas reset 1st of month at 00:00 UTC. Re-evaluation consumes
     quota.
A2 — Core user story rewritten around the async flow (register → submit →
     poll).
A3 — Authentication added as first-class: register (email + password,
     verification, OWASP), login (JWT access + refresh), refresh, logout,
     /me profile. Auth required for all correction endpoints. Out of MVP:
     social login, 2FA, self-service password reset, self-service account
     deletion.
A4 — Correction flow async: submit returns 202 + correction_id + status
     "pending"; retrieve adds status state machine (pending, processing,
     completed, failed). List summaries include status. Re-evaluation uses
     same async flow and consumes quota. Quota check pre-queue, blocking 429
     on exhaustion.
A5 — NFRs split: submission-ack p95 < 500ms; end-to-end correction p95 <
     90s; throughput 5 concurrent in-flight; 99% API uptime decoupled from
     worker uptime (queue absorbs worker downtime).
A6 — Acceptance criteria realigned to constitution v2.0.0 MVP-tier
     thresholds (total MAE ≤ 120, per-comp MAE ≤ 60, ≥ 70% within ±120).
A7 — Out-of-scope extended: institutional/B2B, social login, 2FA,
     self-service password reset, completion webhooks, billing/payments,
     refunds/top-ups.
A8 — Open questions refreshed: quota unit-economics validation, dry-run
     mode, retention window for completed corrections, self-service
     pre-LGPD-window history deletion.
-->

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Student registers, verifies email, and authenticates (Priority: P1)

A Brazilian high-school student preparing for ENEM creates an account with email + password,
verifies the email address, logs in, and obtains an access token (JWT) + refresh token. They can
fetch their own profile, which exposes their tier (Free or Premium), current month's quota usage,
and quota reset date.

**Why this priority**: No correction can happen without an authenticated, verified account. Quota
enforcement and B2C identity scoping (Constitution X) depend entirely on this story.

**Independent Test**: Register a new account, verify the email via the verification token, log in
with valid credentials, and call the profile endpoint with the returned access token; verify the
profile exposes tier, quota used (0 for a fresh account), and quota reset date. Attempt to call any
correction endpoint without a token or with an unverified account; verify 401 (no token) or
typed `email_unverified` error.

**Acceptance Scenarios**:

1. **Given** a valid unused email and an OWASP-compliant password (≥ 12 chars, not in known leaked
   lists),
   **When** the student registers,
   **Then** the account is created in an `unverified` state and a verification token is delivered
   via email. The student cannot submit corrections until verification completes.
2. **Given** a valid verification token,
   **When** the student submits it,
   **Then** the account transitions to `active` and correction submission becomes permitted.
3. **Given** an active account with correct credentials,
   **When** the student logs in,
   **Then** the API returns a short-lived access token and a long-lived refresh token. The access
   token carries the user ID and tier (Free or Premium).
4. **Given** a valid refresh token,
   **When** the student requests a refresh,
   **Then** the API returns a new access token; the prior access token MAY be invalidated by expiry
   but the refresh token remains valid until logout.
5. **Given** a valid refresh token,
   **When** the student logs out,
   **Then** the refresh token is revoked and subsequent refresh attempts with it fail.
6. **Given** a valid access token,
   **When** the student calls the profile endpoint,
   **Then** the API returns user ID, email, tier, current month's quota usage, and quota reset date.
7. **Given** a request to a correction endpoint with no Authorization header,
   **When** the request arrives,
   **Then** the API returns 401 with a typed `unauthenticated` error code.
8. **Given** a request with an unverified account's token,
   **When** the request targets a correction endpoint,
   **Then** the API returns a typed `email_unverified` error.

---

### User Story 2 - Student submits an essay and polls until the correction completes (Priority: P1)

A registered, verified, authenticated student submits their essay text, the official prompt theme,
and optional motivational texts. The API checks quota, accepts the submission (HTTP 202) with a
`correction_id` and status `pending`, and enqueues the job. The student polls the retrieve endpoint
until status reaches `completed` (full correction returned) or `failed` (typed error + flag stating
whether the failure consumed quota).

**Why this priority**: The reason the product exists. Without an asynchronous submit-poll loop that
produces a matrix-conformant correction, no other story has value.

**Independent Test**: With a valid token and remaining quota, submit a known reference essay (from
the golden dataset) with its official prompt theme; verify 202 + `correction_id` + status `pending`,
and that `prompt_version` and `model_id` are recorded against the job. Poll the retrieval endpoint;
verify status progresses through `processing` to `completed` within the end-to-end latency target;
verify the completed payload validates against the published JSON Schema, each competency score is
in `{0, 40, 80, 120, 160, 200}`, the total equals the sum, every competency cites at least one
verbatim excerpt of the submitted essay, every score below 200 carries an improvement path, and
`prompt_version` + `model_id` are present.

**Acceptance Scenarios**:

1. **Given** an authenticated student with remaining quota and a valid essay + prompt theme,
   **When** the student submits,
   **Then** the API responds with 202 Accepted, a `correction_id`, and initial status `pending`.
   The submission is enqueued.
2. **Given** an authenticated student whose monthly quota is exhausted,
   **When** the student submits,
   **Then** the API responds with 429 and a typed `quota_exhausted` error including the quota reset
   date. No work is enqueued. No LLM call is made.
3. **Given** a previously submitted `correction_id` whose worker has not yet started,
   **When** the student polls,
   **Then** the API returns status `pending` and no correction payload.
4. **Given** a `correction_id` whose worker is mid-flight,
   **When** the student polls,
   **Then** the API returns status `processing`.
5. **Given** a `correction_id` whose worker completed successfully,
   **When** the student polls,
   **Then** the API returns status `completed` and the full correction object (final score,
   5 competency entries with scores in `{0, 40, 80, 120, 160, 200}`, justifications citing essay
   excerpts, improvement paths where score < 200, eliminatory flags if applicable, `prompt_version`,
   `model_id`, inference parameters).
6. **Given** a `correction_id` whose worker failed due to user-attributable input (e.g., language
   detected as non-pt-BR after pre-validation slipped, malformed structure),
   **When** the student polls,
   **Then** the API returns status `failed` with a typed error code and `quota_consumed: true`.
7. **Given** a `correction_id` whose worker failed due to provider/internal causes (rate limit,
   timeout, schema-violation-after-retries, internal error),
   **When** the student polls,
   **Then** the API returns status `failed` with a typed error code (`provider_rate_limited`,
   `provider_timeout`, `schema_violation`, `provider_unavailable`, or `internal_error`) and
   `quota_consumed: false`.
8. **Given** an essay totally off-topic relative to the prompt theme,
   **When** the worker completes,
   **Then** the completed correction sets the `off_topic` eliminatory flag, zeroes the affected
   competencies per the matrix, and includes a pt-BR justification.
9. **Given** an essay shorter than the minimum length (Validation),
   **When** the student submits it,
   **Then** the API rejects synchronously with a typed validation error, **without enqueueing** and
   **without consuming quota**.
10. **Given** the LLM provider returns malformed output twice in a row,
    **When** the correction layer exhausts the retry budget,
    **Then** the job transitions to `failed` with typed `schema_violation` error and
    `quota_consumed: false`.

---

### User Story 3 - Student lists their own correction history (Priority: P2)

The student lists their prior corrections to track progress over time. Listing is paginated, may be
filtered by date range, and each summary includes the status field so in-flight corrections appear
alongside completed ones.

**Why this priority**: Progress tracking is the second most cited use case. The system is still
useful for one-off self-assessment without it.

**Independent Test**: Submit N essays for the same authenticated student, poll a subset to
completion and leave others in flight; call the list endpoint; verify N summaries are returned with
correct `correction_id`, `issued_at`, `final_score` (null when not yet completed),
`prompt_theme_title`, and `status`, in reverse-chronological order, paginated according to the
requested cursor and page size.

**Acceptance Scenarios**:

1. **Given** an authenticated student with prior corrections,
   **When** the student requests their listing,
   **Then** the API returns paginated summaries (`correction_id`, `issued_at`, `final_score` or
   null, `prompt_theme_title`, `status`, eliminatory flag if any) in reverse-chronological order by
   default.
2. **Given** a date-range filter,
   **When** the student requests the listing,
   **Then** only corrections whose `issued_at` falls inside the range appear.
3. **Given** a pagination cursor returned by the prior page,
   **When** the student passes it in the next request,
   **Then** the next page returns with no duplicates and no gaps.
4. **Given** an unauthenticated request,
   **When** it arrives at the list endpoint,
   **Then** the API responds 401 with typed `unauthenticated` error.

---

### User Story 4 - Student requests a re-evaluation of their own prior correction (Priority: P2)

A student wants a second opinion or wants to re-run an old essay against an updated prompt version
or model. The student references one of their own `correction_id`s; the API enqueues a **new**
asynchronous correction job linked to the original via `reevaluation_of`. Re-evaluation consumes
quota the same as a fresh submission. The original correction remains accessible and unchanged.

**Why this priority**: Real users iterate after a prompt SemVer bump (Constitution VIII) or after
disagreement. Re-evaluation makes iteration practical but is not strictly required to ship the MVP.

**Independent Test**: Submit an essay, poll to completion, take the `correction_id`, call the
re-evaluation endpoint; verify (a) the response returns 202 + new `correction_id` + status
`pending`; (b) the original correction remains retrievable byte-for-byte; (c) the new correction
links back via `reevaluation_of`; (d) the new correction's `prompt_version` and `model_id` are
recorded independently; (e) quota usage on the profile increased by 1.

**Acceptance Scenarios**:

1. **Given** a `correction_id` belonging to the authenticated student, with remaining quota,
   **When** the student requests a re-evaluation,
   **Then** the API responds 202 with a new `correction_id` and status `pending`. The new job is
   linked via `reevaluation_of` to the original; the original remains retrievable unchanged.
2. **Given** a `correction_id` not owned by the caller,
   **When** they request a re-evaluation,
   **Then** the API returns a not-found / not-authorized response that does not leak existence.
3. **Given** the student's quota is exhausted,
   **When** they request a re-evaluation,
   **Then** the API responds 429 with typed `quota_exhausted` error. No job is enqueued.

---

### Edge Cases

- Essay exactly at the minimum threshold (500 chars / 7 lines, whichever is shorter): accepted.
- Essay exactly at the maximum threshold (3500 chars / 50 lines, whichever is larger): accepted.
- Essay contains the prompt theme text copied verbatim (suspected cópia da proposta motivadora):
  the worker raises the matrix's corresponding eliminatory flag on the completed correction.
- Essay is dissertative-argumentative but completely off-topic: `off_topic` eliminatory flag, total 0.
- Essay is on-topic but not dissertative-argumentative (narrative, poem):
  `not_dissertative_argumentative` flag; affected competencies zeroed.
- Essay is in Portuguese from Portugal (pt-PT): rejected at pre-validation as language mismatch,
  before queueing, without consuming quota.
- Prompt theme is provided but lacks contextualization (title only): rejected at pre-validation,
  before queueing, without consuming quota.
- Worker fails twice on JSON Schema after corrective retries: job → `failed`, typed
  `schema_violation`, `quota_consumed: false`.
- LLM provider unavailable / timing out: job → `failed`, typed `provider_unavailable` /
  `provider_timeout`, `quota_consumed: false`. Per Constitution-aligned posture: fail-fast for the
  user; the queue may retain the queued job for transparent worker retry on transient errors at
  the worker's discretion, but the user-visible status is `failed` once the user-visible deadline
  is breached.
- Worker downtime: API surface remains up; pending submissions accumulate in the queue and are
  drained when worker capacity returns. Status remains `pending` while queued.
- Quota tipping at month boundary: a submission at 23:59:59 UTC on the last day of month M counts
  against month M's quota; a submission at 00:00:00 UTC on day 1 of month M+1 counts against
  month M+1. The reset is atomic at the boundary.
- Concurrent submissions racing the last quota slot: at most one wins; the others receive 429.
- Re-evaluation of a correction whose prompt version has since been retired: uses the **current**
  active prompt/model; the completed payload makes the version delta explicit.
- Token expiry mid-poll: the API returns 401 `token_expired`; the client refreshes and continues
  polling.

## Requirements *(mandatory)*

### Functional Requirements

#### Authentication & Account Management

- **FR-028**: The system MUST support account registration with email + password. Passwords MUST
  satisfy OWASP guidance: minimum 12 characters and rejection against a known leaked-password list.
- **FR-029**: A new account MUST start in an `unverified` state. Email verification MUST be
  required and MUST be a prerequisite for any correction-endpoint access. Verification tokens MUST
  be single-use and time-bounded.
- **FR-030**: The system MUST support login with email + password and MUST return a short-lived
  access token (JWT) and a long-lived refresh token. The access token MUST carry user ID and tier.
- **FR-031**: The system MUST support refresh-token exchange returning a new access token, MUST
  treat refresh tokens as revocable, and MUST support an explicit logout that revokes the active
  refresh token.
- **FR-032**: The system MUST expose a profile endpoint returning user ID, email, tier (Free /
  Premium), current month's quota usage, and quota reset date.
- **FR-033**: All correction endpoints (submit, retrieve, list, re-evaluate) MUST require a valid
  access token. Unauthenticated requests MUST return 401 with a typed `unauthenticated` error.
  Authenticated requests from unverified accounts MUST return a typed `email_unverified` error.

#### Quotas & Tiering

- **FR-034**: The system MUST track each user's monthly correction quota, segmented by tier (Free
  and Premium). The MVP initial values are Free = 3 corrections/month and Premium = 30
  corrections/month, subject to unit-economics validation before launch.
- **FR-035**: Quotas MUST reset atomically at 00:00 UTC on the first day of each calendar month.
- **FR-036**: Every successful enqueue of a correction job (new submission or re-evaluation) MUST
  consume exactly 1 quota unit. A job that fails for user-attributable reasons MUST consume quota;
  a job that fails for provider/internal reasons MUST NOT consume quota (the failed response MUST
  carry `quota_consumed: false`). A submission rejected at pre-validation (length, language,
  missing prompt fields) MUST NOT consume quota.
- **FR-037**: Quota exhaustion MUST be checked **before** any work is enqueued. Exhausted
  submissions MUST return 429 with typed `quota_exhausted` including the quota reset date. No LLM
  call is made.

#### Submission (asynchronous)

- **FR-001**: The system MUST accept an essay submission containing: essay text (pt-BR), a prompt
  theme block (title + brief contextualization, mandatory), and optional motivational texts
  (`textos motivadores`). Student-provided metadata (preparation level, grade) MAY be accepted but
  is not required.
- **FR-002**: The system MUST validate essay text length **before** queueing. Minimum: shorter of
  7 lines or 500 characters. Maximum: larger of 50 lines or 3500 characters. Violations return a
  typed validation error and do NOT consume quota.
- **FR-003**: The system MUST validate that the essay language is Brazilian Portuguese before
  queueing and reject other languages (including pt-PT) with a typed `language_mismatch` error
  without consuming quota.
- **FR-004**: The system MUST reject submissions whose prompt theme block lacks title or
  contextualization, before queueing, without consuming quota.
- **FR-038**: A valid submission MUST be enqueued and the API MUST respond with HTTP 202 Accepted,
  a `correction_id`, and an initial `status` of `pending`. The submission response MUST NOT block
  on the LLM.

#### Correction Result (worker-produced)

- **FR-005**: The worker MUST score each of the 5 official ENEM competencies (C1–C5) on the set
  `{0, 40, 80, 120, 160, 200}`. The final score MUST equal the sum of the 5 competency scores and
  MUST lie in `[0, 1000]`.
- **FR-006**: Every competency entry in the completed correction MUST include: the integer score,
  a pt-BR justification grounded in the official matrix descriptors for that level, at least one
  verbatim excerpt cited from the submitted essay, and — whenever the score is below 200 — a pt-BR
  improvement path.
- **FR-007**: The worker MUST detect and surface the official eliminatory criteria as typed flags
  on the completed correction: `off_topic`, `annulled` (with reason), `insufficient_text`, and
  `not_dissertative_argumentative`. When an eliminatory flag fires, the affected competencies MUST
  be zeroed per the official matrix.
- **FR-008**: Every completed correction MUST embed the `prompt_version` (SemVer) and `model_id`
  used to produce it, plus the inference parameters (temperature, seed when applicable, top_p,
  max_tokens) — per Constitution III.
- **FR-009**: Every correction job (regardless of outcome) MUST be persisted as an immutable
  record keyed by `correction_id`. Once `status` reaches a terminal value (`completed` or
  `failed`), the record's payload MUST NOT change.
- **FR-010**: Worker output MUST be validated against a versioned JSON Schema. On validation
  failure the worker MUST retry at most twice with a corrective prompt; on exhaustion the job
  MUST transition to `status = failed` with typed `schema_violation` error and
  `quota_consumed: false` (Constitution IV).

#### Retrieval (status-aware)

- **FR-011**: The system MUST allow the authenticated owner to retrieve a correction by
  `correction_id`. The response MUST include a `status` field with one of `pending`, `processing`,
  `completed`, `failed`.
- **FR-039**: When `status = completed`, the response MUST include the full correction object
  exactly as produced by the worker, and repeated retrievals of the same `correction_id` MUST
  return byte-for-byte identical payloads.
- **FR-040**: When `status = failed`, the response MUST include a typed error code, a pt-BR
  message, and a `quota_consumed` boolean indicating whether the failure consumed the owner's
  quota.
- **FR-012**: Retrieval of a `correction_id` that does not exist or is not owned by the caller
  MUST return a not-found response that does not allow callers to distinguish "does not exist"
  from "exists but not yours".

#### Listing (own corrections only)

- **FR-013**: The system MUST allow the authenticated student to list **their own** corrections.
  Cross-user listing is not supported (no institutional surface).
- **FR-014**: Listing MUST support optional date-range filtering and cursor-based pagination.
- **FR-015**: Listing summaries MUST include at minimum: `correction_id`, `issued_at`, `status`,
  `final_score` (null when not in a terminal completed state), `prompt_theme_title`, and any
  eliminatory flag set. Summaries MUST NOT include full justifications.

#### Re-evaluation

- **FR-016**: The system MUST allow the authenticated owner to re-evaluate one of their own prior
  corrections by `correction_id`. The result MUST be a new asynchronous job (status `pending`)
  with a new `correction_id`, linked to the original via a `reevaluation_of` field. Re-evaluation
  MUST consume 1 quota unit on successful enqueue.
- **FR-017**: Re-evaluation MUST always use the **current** active prompt version and model. The
  completed payload MUST surface the version delta when it differs from the original.

#### Errors & Observability

- **FR-020**: LLM-provider-attributable failures MUST surface as distinct typed error codes inside
  the `failed` status response: `provider_rate_limited`, `provider_timeout`, `schema_violation`,
  `provider_unavailable`. These MUST NOT collapse into generic 500s on the HTTP layer either.
- **FR-021**: Every correction job (successful or failed) MUST emit an audit record containing:
  input hash, complete structured output (when produced), model ID, prompt version, latency, and
  cost. Student PII MUST NOT appear in logs; only `correction_id` and the opaque user ID
  (Constitution VII).

#### Privacy & LGPD

- **FR-024**: The system MUST refuse to correct essays from students under 18 unless an explicit
  parental/guardian consent record is on file for that user, per Constitution X. Consent state is
  captured during account onboarding for minors.
- **FR-025**: The system MUST support a deletion request that removes the essay text, justifications,
  and derived records within 15 days of the request, per Constitution X. In the MVP, deletion is
  routed via support (not self-service). Audit input hash and minimal metadata may be retained for
  integrity; essay content MUST be purged.
- **FR-026**: Submitted essay text MUST NOT be used by the LLM provider for training or
  fine-tuning; provider configuration MUST enforce Zero Data Retention (or equivalent) per
  Constitution X. The platform MUST NOT sell, share, or monetize student essays or derived
  analytics.

#### Non-Functional

- **FR-041**: Submission acknowledgement latency (POST returning 202) MUST be under **500 ms at
  p95**.
- **FR-023**: End-to-end correction latency (from queueing to `status = completed`) MUST be under
  **90 seconds at p95** in the MVP, assuming self-hosted or managed open-source inference (Ollama
  or equivalent). This target MAY be revised when production traffic patterns emerge.
- **FR-022**: The system MUST sustain at least **5 concurrent in-flight corrections** in the MVP
  without breaching the end-to-end latency target.
- **FR-027**: Monthly API-surface availability MUST be ≥ 99%. Inference-worker downtime degrades
  end-to-end latency but MUST NOT bring the API surface down; pending jobs remain in queue and
  resume when worker capacity returns.

### Key Entities

- **User**: The end-user student account. Carries opaque user ID, email, password hash,
  verification state, tier (Free or Premium), and the consent record reference when applicable.
- **Session / Tokens**: Access token (JWT, short-lived, carries user ID + tier) and refresh token
  (long-lived, revocable). Logout revokes the refresh token.
- **Quota Ledger**: Per-user, per-month counter of consumed corrections. Reset atomically at
  00:00 UTC on the first day of each calendar month.
- **Essay Submission**: The text plus prompt theme block + optional motivational texts. Tied to
  the submitting user. Immutable post-submission.
- **Prompt Theme**: Title + contextualization describing the ENEM prompt the essay addresses.
- **Correction Job**: Carries `correction_id`, `owner_user_id`, `status` (`pending` |
  `processing` | `completed` | `failed`), `issued_at`, the full correction payload when
  completed, the typed error and `quota_consumed` flag when failed, `prompt_version`, `model_id`,
  inference parameters, and optional `reevaluation_of`.
- **Competency Entry**: One of C1–C5. Score in `{0, 40, 80, 120, 160, 200}`, justification (pt-BR),
  ≥ 1 verbatim excerpt from the essay, improvement path (pt-BR) when score < 200.
- **Audit Record**: Per-job operational record with input hash, full output (when produced), model,
  prompt version, latency, cost, `correction_id`, and opaque user ID. No PII.
- **Consent Record**: Parental/guardian consent on file for users under 18; prerequisite for
  correction.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of completed corrections present 5 competency scores from
  `{0, 40, 80, 120, 160, 200}` summing exactly to a final score in `[0, 1000]`.
- **SC-002**: 100% of competency entries cite at least one excerpt that appears verbatim in the
  submitted essay text.
- **SC-003**: 100% of competency entries with score below 200 include a non-empty improvement
  path in pt-BR.
- **SC-004**: A correction retrieved by ID in `completed` status is byte-for-byte identical to
  any prior retrieval of the same correction in 100% of calls.
- **SC-005**: On the golden dataset, total-score MAE is **≤ 120 points**, per-competency MAE is
  **≤ 60**, and **≥ 70%** of corrections fall within ±120 points of the human reference score
  (Constitution IX MVP-tier merge gate).
- **SC-006**: Eliminatory criteria detection (off-topic, annulment, insufficient text,
  not-dissertative-argumentative) is correct on ≥ 95% of golden-dataset eliminatory cases.
- **SC-007**: p95 submission-acknowledgement latency is under **500 ms**; p95 end-to-end
  correction latency is under **90 seconds**; the system sustains 5 concurrent in-flight
  corrections without breaching the end-to-end target.
- **SC-008**: 0% of `completed` correction payloads fail JSON Schema validation. Schema violations
  from the LLM, after exhausting the 2-attempt retry budget, surface as `failed` corrections with
  typed `schema_violation` error in 100% of cases (never as a 500, never as free-form text).
- **SC-009**: 0 occurrences of student PII (name, email, document numbers) in production logs
  across any rolling 30-day audit window.
- **SC-010**: Re-evaluations preserve the original correction unmodified in 100% of cases, link
  back via `reevaluation_of`, and consume exactly 1 quota unit per successful enqueue.
- **SC-011**: ≥ 95% of input-validation rejections (length, language, missing prompt fields)
  return without queueing and without consuming quota.
- **SC-012**: Monthly API-surface uptime ≥ 99% over any rolling 30-day window for the MVP.
- **SC-013**: 100% of quota-exhausted submission attempts return 429 with typed `quota_exhausted`
  and do NOT enqueue work.
- **SC-014**: 100% of `failed` job responses include a `quota_consumed` boolean accurately
  reflecting the user-attributable vs. provider/internal classification of the failure.
- **SC-015**: 100% of correction endpoints reject unauthenticated requests with 401 and
  unverified-account requests with `email_unverified`.

## Out of Scope (MVP)

- OCR of handwritten essays (plain text only).
- Grammar-level inline annotations / span-level markup.
- Comparison reports against other students or cohort analytics.
- Multi-essay batch submission in a single request.
- Real-time streaming of the correction as it is generated.
- Direct user-facing UI (this is the API; frontend is a separate product).
- **Institutional or B2B usage of any kind** (no organizational accounts, no API keys for schools,
  no educator dashboards or views).
- Social login providers (Google, Apple, etc.).
- Two-factor authentication.
- Self-service password reset (handled by support in the MVP).
- Self-service account deletion (right-to-erasure routed via support per the LGPD 15-day window).
- Webhook notifications for correction completion (polling only).
- Payment integration and billing for Premium tier (the tier distinction exists in the data model
  but the upgrade path is out of scope; Premium accounts are provisioned manually in the MVP).
- Refund or quota top-up flows.

## Assumptions

- Target audience is Brazilian high-school students (B2C only, Constitution X). User-facing
  strings (justifications, improvement paths, error messages) are in pt-BR. Code, identifiers,
  and internal documentation are in English.
- The LLM execution backend is provider-agnostic (Constitution II); self-hosted-first (Ollama or
  equivalent) is the MVP default, accessed through the provider abstraction. The 90-second p95
  end-to-end latency target assumes that posture.
- Free-tier quota = 3/month and Premium-tier quota = 30/month are initial estimates pending
  unit-economics validation before launch (see Open Questions).
- A `dry_run` mode on the submission endpoint executes all deterministic pre-validation (length,
  language, theme completeness, schema availability) and returns either a validation response or
  an `ok_to_submit` acknowledgement **without invoking the LLM or consuming quota**. Useful for
  frontend development and CI smoke tests. (Carried forward from v1 as an open question; see
  Open Questions for confirmation prompt.)
- Authentication is JWT-based with a short-lived access token and a revocable, long-lived refresh
  token. Exact lifetimes are an implementation concern.
- Persistence technology is a project-internal concern; the API exposes IDs but does not promise
  any specific storage technology in the contract.
- The golden dataset (Constitution IX) is curated out-of-band; this spec consumes it as a CI gate.
- Worker concurrency, queue technology, and retry/backoff policy are implementation concerns
  bounded by the latency, availability, and quota-consumption rules above.

## Open Questions

- **Q1 — Quota unit economics**: The Free-tier quota of 3/month and Premium-tier quota of 30/month
  are initial estimates. Final values depend on inference cost per correction × target margin and
  MUST be validated before launch.
- **Q2 — Dry-run mode confirmation**: Should the API formally expose a `dry_run` mode on the
  submission endpoint that validates input but skips the LLM call, useful for frontend development
  against realistic responses? (Carried over from v1; assumption above treats it as supported, but
  the formal contract surface is unconfirmed.)
- **Q3 — Retention window for completed corrections**: Should completed corrections be retrievable
  forever, or for a fixed window (e.g., 12 months) after which they are archived or deleted? This
  intersects LGPD storage minimization (Constitution X).
- **Q4 — Self-service pre-window deletion**: Should the student be able to delete their own
  correction history before the LGPD-mandated 15-day deletion window, as a self-service action?
  If yes, deletion is immediate from the user's perspective and the audit trail retains only the
  input hash + metadata.
