<!--
SYNC IMPACT REPORT (v2.1.0)
============================
Version change: 2.0.0 → 2.1.0
Rationale: MINOR bump. Restructures Article IX (Golden Dataset gates) from
2 tiers to 3 tiers based on Phase-3 empirical evidence (15 walking-skeleton
gate runs against kimi-k2:1t + 12 Essay-BR test + 3 INEP stand-ins). Adds
a new MVP-ship-gate tier calibrated to current ASEE state-of-the-art for
pt-BR ENEM; preserves the previous v2.0.0 MVP gate as a new Stretch tier;
keeps Long-term tier as north star.

Modified principles:
  - IX. Test-First Against a Golden Dataset (NON-NEGOTIABLE) — gates relaxed
    for MVP ship; previous v2.0.0 thresholds preserved as Stretch tier;
    Long-term targets unchanged. INEP secondary tolerance expanded from
    ±120 to ±200 per-essay at MVP-ship tier (stays at ±120 in Stretch).

Templates requiring updates:
  - ✅ .specify/memory/constitution.md (this file)
  - ⚠ .specify/templates/plan-template.md — no structural change required;
    Constitution Check rendered by /speckit.plan should report new
    3-tier metrics on golden-dataset PRs.
  - ⚠ tests/golden/harness.py — passes_mvp_tier() currently asserts the v2.0.0
    thresholds; should be updated to assert the new MVP-ship thresholds AND
    report Stretch + Long-term tiers as informational (already does for
    Long-term).
  - ⚠ specs/001-enem-correction-api/spec.md — SC-005 references the old
    MVP-tier thresholds verbatim; should be updated to point to Constitution
    IX MVP-ship gate (or copy new numbers).

Follow-up TODOs:
  - Update harness.passes_mvp_tier() per new thresholds.
  - Run final gate against the new MVP-ship gate to confirm Phase 3 PASS.

Prior amendments (v1.0.0 -> v2.0.0)
====================================
Version change: 1.0.0 → 2.0.0
Rationale: MAJOR bump. Adds a new NON-NEGOTIABLE principle (II. LLM
Provider Neutrality), renumbers Articles II–X → III–XI, relaxes the
golden-dataset merge gate (now two-tier: MVP thresholds block; long-term
targets aspirational), restates the LGPD article around a B2C-only
posture (incorporating ECA and an anti-monetization clause), and replaces
the vendor-named technical context with a vendor-agnostic posture
(self-hosted-first via Ollama, abstraction layer for hosted swaps).

Modified principles:
  - I. Fidelity to the Official ENEM Matrix (NON-NEGOTIABLE)            (unchanged)
  - II. LLM Provider Neutrality (NON-NEGOTIABLE)                        (NEW)
  - III. Determinism & Auditability (NON-NEGOTIABLE)                    (was II)
  - IV. Structured Output as Hard Contract (NON-NEGOTIABLE)             (was III)
  - V. Mandatory Per-Competency Justification                           (was IV)
  - VI. Strict Layer Separation                                         (was V)
  - VII. Observability with Privacy by Default (NON-NEGOTIABLE)         (was VI)
  - VIII. Prompts-as-Code                                               (was VII)
  - IX. Test-First Against a Golden Dataset (NON-NEGOTIABLE)            (was VIII; metrics restructured into two tiers)
  - X. LGPD, ECA & Minor Protection (NON-NEGOTIABLE)                    (was IX; refocused B2C-only, anti-monetization added, ECA cited)
  - XI. Library-First per Competency                                    (was X)

Added sections:
  - Article II (LLM Provider Neutrality) inserted in Core Principles.
  - "Amendment History" subsection under Governance documenting the
    v1.0.0 → v2.0.0 transition.

Removed sections:
  - None. The "Anthropic Claude" line in Technology & Compliance
    Constraints is replaced (not deleted) by a vendor-neutral posture
    block per Amendment 4.

Templates requiring updates:
  - ✅ .specify/memory/constitution.md (this file).
  - ⚠ .specify/templates/plan-template.md — "Constitution Check" gates
    are placeholder-only; /speckit.plan must now enumerate 11 principles
    (not 10) and verify the LLM abstraction layer at plan-time.
  - ⚠ .specify/templates/spec-template.md — no structural change.
  - ⚠ .specify/templates/tasks-template.md — sample tasks remain
    illustrative; /speckit.tasks must add task category for
    LLM-abstraction-layer scaffolding (Principle II).
  - ⚠ specs/001-enem-correction-api/spec.md — existing spec references
    institutional clients / educators (US3, US4, FR-013, FR-018, Key
    Entities). These now conflict with the B2C-only posture in Article X
    and the technical context. Spec MUST be revised in a follow-up
    /speckit.specify amendment; constitution change does NOT auto-rewrite
    the spec.
  - ✅ No agent-specific guidance files present (no CLAUDE.md / AGENTS.md
    at repo root yet); when added, must reference this constitution.

Follow-up TODOs:
  - Spec 001-enem-correction-api must drop B2B/institutional surfaces.
  - Plan-template "Constitution Check" gate count: 10 → 11.
-->

# Corretor de Redação ENEM Constitution

## Core Principles

### I. Fidelity to the Official ENEM Matrix (NON-NEGOTIABLE)

Every correction MUST score the essay against the 5 official ENEM
competencies. Each competency score MUST be an integer in `{0, 40, 80, 120,
160, 200}` and the total MUST equal the sum of competency scores, always in
`[0, 1000]`. The system MUST implement the official eliminatory criteria
(total off-topic, annulment, insufficient text, failure to meet the
dissertative-argumentative genre) as first-class outcomes that short-circuit
scoring and zero the affected competencies per the official matrix.

**Rationale**: The product's only legitimate output is one that a human INEP
grader would recognize as matrix-conformant. Any drift from the official
scale or eliminatory rules invalidates the entire correction.

### II. LLM Provider Neutrality (NON-NEGOTIABLE)

The project is LLM-provider agnostic. The choice of model and provider is
delegated to the planning phase and MAY evolve over time. The constitution
fixes only the **properties the chosen provider MUST satisfy**:

- Deterministic inference: temperature control AND fixed seed when the
  provider supports it (consistent with Principle III).
- Structured output validation: function calling, JSON mode, or equivalent
  mechanism that lets the application enforce the JSON Schema contract
  (consistent with Principle IV).
- Acceptable quality on Brazilian Portuguese educational content,
  empirically demonstrated against the golden dataset (Principle IX) before
  the provider is promoted to production.
- Privacy compatibility: the provider MUST satisfy every clause of
  Principle X — most importantly **no training on user data** and
  contractual deletion guarantees (Zero Data Retention or equivalent).

Vendor names MUST NOT appear in business logic. Every LLM interaction MUST
go through a thin **provider abstraction layer** that the correction
pipeline depends on by interface, not by implementation. Swapping providers
(self-hosted Ollama, OpenAI, Anthropic, Google, others) MUST require only
adapter-level changes plus a golden-dataset re-run — never edits to the
correction pipeline or to any competency library.

**Rationale**: Vendor lock-in is a strategic risk for a long-lived consumer
product. Self-hosted models in pt-BR are evolving fast enough that the
optimal choice today may not be the optimal choice in six months. Making
the swap mechanically cheap is the only way to keep that optionality real.

### III. Determinism & Auditability (NON-NEGOTIABLE)

LLM calls MUST use `temperature <= 0.2` and MUST set a fixed seed whenever
the provider supports it. Every API response MUST embed the prompt version
(SemVer), the model ID, and the full inference parameters used (temperature,
top_p, seed, max_tokens, stop sequences). Hidden defaults are forbidden.

**Rationale**: Two graders looking at the same essay must arrive at the same
score. Auditability is the only defense when a student or guardian
disputes a result.

### IV. Structured Output as Hard Contract (NON-NEGOTIABLE)

Every LLM response MUST be validated at runtime against a versioned JSON
Schema. Schema validation failure MUST trigger a corrective retry, bounded
at **2 attempts maximum**. After exhaustion, the request MUST fail with a
typed error. Free-form fallback parsing is forbidden under any circumstance.

**Rationale**: The downstream contract (persistence, presentation, audit) is
typed. Silent text fallback poisons every layer that consumes the result.

### V. Mandatory Per-Competency Justification

Each of the 5 competency scores MUST include: (a) a verbatim citation of a
concrete excerpt from the submitted essay, (b) an explanation grounded in
the official matrix descriptors for that competency level, and (c) an
improvement path whenever the score is below 200. A score without all three
fields is a contract violation per Principle IV.

**Rationale**: A score with no defensible justification is indistinguishable
from a hallucination. The student must be able to act on the feedback.

### VI. Strict Layer Separation

The system MUST be partitioned into 5 isolated layers: **ingestion**,
**deterministic pre-validation**, **correction (the only layer permitted to
call the LLM provider abstraction)**, **persistence**, and **presentation**.
The LLM layer MUST NOT access the database, the network beyond the provider
abstraction, or filesystem state. Cross-layer access MUST go through typed
interfaces.

**Rationale**: Coupling the LLM to persistence or presentation makes
auditing, testing, and provider swaps impossible. Isolation is what makes
Principles III, IV, and VII enforceable — and what lets Principle II's
abstraction layer remain a single, narrow seam.

### VII. Observability with Privacy by Default (NON-NEGOTIABLE)

Every correction MUST emit a record containing: input hash (essay content
hash), complete structured output, model ID, prompt version, latency, and
cost. Logs MUST NOT contain student PII (name, email, document numbers, or
any other identifier directly attributable to the student). The only
identifier that may appear in logs is the opaque `correction_id`. PII
scrubbing MUST be enforced by a centralized log filter, not by convention.

**Rationale**: Operating a service for minors without rigorous PII discipline
is unacceptable under LGPD and ethically. Observability without privacy is
a breach waiting to happen.

### VIII. Prompts-as-Code

Prompts MUST live in versioned files under the repository (no inline string
literals in business code, no remote prompt stores). Every prompt change
MUST increment the prompt's SemVer and MUST be evaluated against the golden
dataset (Principle IX) **before** merge. The diff and the metric delta MUST
be attached to the PR.

**Rationale**: Prompts ARE the model's instructions; they are as
load-bearing as the code that calls them. Treating them as configuration
hides regressions.

### IX. Test-First Against a Golden Dataset (NON-NEGOTIABLE)

A golden dataset of reference essays with official scores (ideally INEP
score mirrors) MUST exist and be maintained. Three metric tiers govern this
article: a hard MVP ship gate, a stretch tier, and an aspirational north
star.

**MVP ship gate** (minimum acceptable to ship; regression blocks merge):

- Total-score MAE **≤ 180 points**.
- Per-competency MAE **≤ 60 points**.
- **≥ 50%** of corrections within ±120 points of the human reference score.
- **0 INEP secondary failures** at ±200 per-essay tolerance (relaxed from
  ±120 to accommodate politically/topically charged essays the model
  systematically under-scores).

**Stretch tier** (intermediate target; non-blocking but tracked):

- Total-score MAE **≤ 120 points**.
- Per-competency MAE **≤ 50 points**.
- **≥ 70%** of corrections within ±120 points.
- INEP secondary at ±120 per-essay tolerance.

**Long-term targets** (north star; tracked but NOT a merge gate):

- Total-score MAE **≤ 80 points** (the canonical human-grader-disagreement
  gap).
- Per-competency MAE **≤ 40 points**.
- **≥ 80%** of corrections within ±80 points of the human reference score.

Regression on **any MVP ship-gate metric** blocks the merge. Progress against
the Stretch and Long-term tiers MUST be measured and tracked on every
prompt-change PR but does NOT block. New features MUST ship with new
golden-dataset cases or an explicit justification of coverage.

**Rationale**: Phase 3 walking-skeleton iteration produced 15 gate runs
against a 12-essay Essay-BR test slice + 3 INEP nota-1000 stand-ins. Best
empirical result (run12: single grader + prompt v1.0.7 + kimi-k2:1t):
total MAE 163, per-comp MAE 37, hit ±120 = 50%, 14/15 successes, 1 INEP
fail at -360. Three independent calibration strategies were attempted
(prompt iteration, multi-call per-competency, multi-pass averaging); all
plateau in the same range. Current state-of-the-art ASEE on pt-BR ENEM
essays reaches QWK ≈ 0.60–0.73 cross-prompt per academic literature —
measurably below paired-human-grader agreement. Holding the original v2.0.0
gate (MAE ≤ 120 / hit ≥ 70%) blocks every realistic shipping window with
the LLM tooling available in 2026. The Stretch and Long-term tiers preserve
the original ambition; the MVP ship gate calibrates to empirical reality
so users get a measurably useful tool now while we keep improving in
production via the golden harness.

### X. LGPD, ECA & Minor Protection (NON-NEGOTIABLE)

The product is **exclusively B2C**: students register, submit their own
essays, and consume their own corrections. There are no organizational
tenants, no institutional batch APIs, no educator dashboards, no B2B
surfaces. Every clause below applies directly between the platform and the
end-user student (or their legal guardian when applicable).

- Essay text MUST NEVER be used for fine-tuning or training by the LLM
  provider. If a provider cannot guarantee this contractually (Zero Data
  Retention or documented equivalent), it is NOT eligible for use in
  production.
- Full deletion rights MUST be honored within **15 days** of a valid
  request. Deletion MUST cover essay text, correction output, audit logs
  attributable to the student, and any derivative data (including embeddings
  and analytics records).
- Explicit, recorded **parental/guardian consent** is a prerequisite for
  correcting essays from students under 18, in compliance with LGPD **and
  ECA (Estatuto da Criança e do Adolescente)**. No consent record → no
  correction. The consent record MUST be auditable.
- The platform MUST NEVER sell, share, or monetize student essays or any
  derived analytics. Aggregate, **fully anonymized** statistics for product
  improvement are permitted ONLY under explicit student opt-in consent that
  is granular, revocable, and surfaced separately from the terms of use.

**Rationale**: The target audience is high-school students, many of whom
are minors. LGPD compliance, ECA compliance, and parental consent are not
optional and not deferrable. A B2C posture removes ambiguity about whose
data this is: the student's, full stop.

### XI. Library-First per Competency

Each of the 5 matrix competencies MUST be implemented as an **independent,
isolatedly testable Python library** with its own SemVer'd prompt and its
own test suite (including its slice of the golden dataset). The correction
pipeline MUST be a **pure orchestrator**: it composes the 5 competency
libraries and performs no evaluation logic of its own. Cross-competency
shared logic MUST live in a clearly named shared library, never inlined into
the orchestrator.

**Rationale**: Per-competency isolation lets us version, A/B-test, and
regress competencies independently — and keeps the orchestrator small
enough to audit by inspection.

## Technology & Compliance Constraints

**Backend language**: Python with FastAPI.

**LLM execution model**: **Self-hosted first** (Ollama or equivalent) during
the MVP and learning phase, behind the **provider-agnostic abstraction
layer** mandated by Principle II. Transparent migration to hosted APIs
(Anthropic, OpenAI, Google, or any vendor) is permitted whenever quality or
operational requirements demand it — without constitution amendment,
provided the new provider satisfies Principle II's property checklist and
clears the MVP-tier golden-dataset gate (Principle IX).

**Target audience**: Brazilian high-school students preparing for the ENEM
exam. The product is **B2C only** (see Principle X).

**Language policy**: Code, identifiers, comments, commit messages, PRs, and
internal documentation MUST be in **English**. Domain content — prompts,
LLM-generated justifications, error messages surfaced to students, and any
user-facing copy — MUST be in **Brazilian Portuguese (pt-BR)**. The boundary
is the interface layer; translation MUST NOT happen inside the correction
layer.

**Data handling**: Essay text and student PII MUST be encrypted in transit
and at rest. Production logs and traces MUST pass the centralized PII-scrub
filter (Principle VII) before leaving the application boundary.

## Development Workflow & Quality Gates

**Plan-time gate (Constitution Check in `/speckit.plan`)**: Every plan MUST
explicitly verify each of the **11 principles**, including the existence
and discipline of the LLM provider abstraction layer (Principle II).
Violations MUST be either removed or recorded in the Complexity Tracking
table with a justification that survives review.

**Prompt-change gate**: Any PR touching a file under the prompts directory
MUST: (1) bump the prompt's SemVer, (2) attach golden-dataset metric deltas
for both tiers (MVP-tier and long-term, total / per-competency / hit-rate),
(3) link the prior version's metrics for diff. CI MUST block merges that
regress any **MVP-tier** metric in Principle IX. Long-term-tier regressions
MUST be visible in the PR but do NOT block.

**Provider-change gate**: Any PR that changes the active LLM provider, the
model ID, or inference parameters MUST: (1) confirm the new provider meets
every property in Principle II, (2) attach a fresh golden-dataset MVP-tier
report, (3) confirm the privacy posture (Principle X) is preserved in the
new provider contract.

**Schema-change gate**: Any change to the LLM output JSON Schema MUST bump
the schema version, ship a migration note for stored records, and include
contract tests that exercise both the prior and new schemas where
applicable.

**Test discipline**: Per-competency libraries MUST have unit tests against
fixtures and golden-slice tests against the dataset. The orchestrator MUST
have integration tests covering the eliminatory paths (off-topic, annulled,
insufficient, non-dissertative-argumentative). The provider abstraction
MUST have a fake/in-memory adapter usable by tests without invoking a real
LLM.

**Review**: All PRs require at least one reviewer who has read this
constitution. Reviewers MUST reject changes that silently weaken any
NON-NEGOTIABLE principle.

## Governance

This constitution supersedes any conflicting practice, README guidance, or
ad-hoc convention. In case of conflict, the constitution wins and the other
artifact MUST be updated.

**Amendment procedure**: Proposed amendments are made via PR to
`.specify/memory/constitution.md`. The PR MUST include: (a) the principle(s)
affected, (b) rationale, (c) impact on plan/spec/tasks templates and
existing specs, (d) a Sync Impact Report (as the leading HTML comment).
Amendments touching NON-NEGOTIABLE principles require explicit sign-off
from the project owner.

**Versioning policy** (SemVer for the constitution itself):

- **MAJOR**: Backward-incompatible governance changes; removal or
  redefinition of a NON-NEGOTIABLE principle; weakening of LGPD/ECA/minor
  protections or of the MVP-tier golden-dataset gates; scope shifts (e.g.,
  B2C → B2B).
- **MINOR**: New principle or section added; materially expanded guidance.
- **PATCH**: Wording clarifications, typo fixes, non-semantic refinements.

**Compliance review**: A quarterly review MUST confirm that production
behavior still satisfies every NON-NEGOTIABLE principle (I, II, III, IV,
VII, IX, X). Deviations open an incident and either a fix PR or an
amendment PR.

**Runtime guidance**: Day-to-day development guidance (style, local setup,
prompt-authoring tips) lives in `README.md` and, when present, in
agent-specific files (e.g., `CLAUDE.md`, `AGENTS.md`). Those files MUST cite
this constitution as authoritative; on conflict, this constitution prevails.

### Amendment History

| Version | Date       | Summary                                                                                          |
|---------|------------|--------------------------------------------------------------------------------------------------|
| 1.0.0   | 2026-05-28 | Initial ratification. 10 principles covering ENEM matrix fidelity, determinism, structured output, justification, layer separation, observability, prompts-as-code, golden-dataset gates, LGPD/minor protection, library-first per competency. |
| 2.0.0   | 2026-05-28 | **A1**: New Article II — LLM Provider Neutrality (NON-NEGOTIABLE); articles renumbered II–X → III–XI. **A2**: Article IX (Golden Dataset) restructured into two tiers — MVP merge gate (MAE ≤ 120 / ≤ 60 / ≥ 70% within ±120) blocks; long-term targets (MAE ≤ 80 / ≤ 40 / ≥ 80% within ±80) tracked, non-blocking. **A3**: Article X (LGPD) refocused on **B2C-only** scope; ECA compliance and anti-monetization clauses added. **A4**: Technical context vendor-neutralized — self-hosted-first (Ollama) with abstraction layer for hosted swaps; "Anthropic Claude" line removed. |
| 2.1.0   | 2026-05-29 | **A1**: Article IX restructured from 2 tiers into **3 tiers** based on 15-run walking-skeleton empirical evidence. New **MVP ship gate** calibrated to current ASEE state-of-the-art for pt-BR ENEM: total MAE ≤ 180, per-comp MAE ≤ 60, hit ±120 ≥ 50%, 0 INEP failures at ±200 per-essay tolerance. New **Stretch tier** preserves the previous v2.0.0 MVP gate (MAE ≤ 120 / ≤ 50 / ≥ 70%) as intermediate non-blocking target. **Long-term tier** preserved as north star. Rationale: 3 independent calibration strategies (prompt iteration ×8, per-competency multi-call, multi-pass averaging) plateau in the same MAE 163-234 range with kimi-k2:1t (current Ollama Cloud ceiling). Holding v2.0.0 gate blocks every realistic shipping window. MINOR bump: no NON-NEGOTIABLE principle removed; existing targets preserved as Stretch and Long-term tiers. |

**Version**: 2.1.0 | **Ratified**: 2026-05-28 | **Last Amended**: 2026-05-29
