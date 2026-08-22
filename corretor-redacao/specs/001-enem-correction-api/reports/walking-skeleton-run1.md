# Walking-Skeleton Gate Run 1 — FAIL

**Date**: 2026-05-29
**Model**: `qwen3-next:80b` (Ollama Cloud)
**Prompt version**: 1.0.0
**Wall time**: 17 min (1024s) for 15 examples
**Phase 3 verdict**: **NO-GO — do not start Phase 4**

## Per-essay results

| slug | expected | predicted | delta | error |
|---|---|---|---|---|
| 0001 | 400 | 0 | -400 | false `insufficient_text` |
| 0002 | 120 | None | — | timeout 75s |
| 0003 | 400 | 680 | +280 | — |
| 0004 | 240 | 80 | -160 | — |
| 0005 | 600 | None | — | schema_violation (2 attempts exhausted) |
| 0006 | 560 | 640 | +80 | — |
| 0007 | 440 | 0 | -440 | false eliminatory |
| 0008 | 560 | 0 | -560 | false eliminatory |
| 0009 | 800 | 640 | -160 | — |
| 0010 | 1000 | 960 | -40 | — |
| 0011 | 880 | 440 | -440 | — |
| 0012 | 840 | None | — | timeout 75s |
| inep-001 | 1000 | 1000 | 0 | ✓ |
| inep-002 | 1000 | 680 | -320 | **INEP fail** (>±120) |
| inep-003 | 1000 | 960 | -40 | — |

## Metrics

- Successes: 11/15 (3 timeouts, 1 schema-exhaustion, 1 returned 0)
- Total-score MAE (successes only): **~229**. Gate: ≤120. **FAIL.**
- INEP per-essay gate: inep-002 fails (-320 > ±120). **FAIL.**

## Root causes

1. **Operational — false timeouts** (essays 0002, 0012). LLM call clock cut at 75s. Default
   too tight for 80B reasoning + long prompts. **Fixed**: `DEFAULT_TIMEOUT_S` 75 → 120.

2. **Schema violation** (essay 0005). 2 attempts exhausted. Need to inspect raw output —
   likely the model produced extra prose around the JSON. May warrant adding `format=json`
   reinforcement to the corrective-retry user message.

3. **Over-application of eliminatory criteria** (0001, 0007, 0008 all scored 0). Model uses
   `insufficient_text` as escape hatch on essays it finds confusing. Root cause: the
   eliminatory_check.md prompt lists the criteria too prominently and the model picks them
   defensively. Fix: rewrite eliminatory_check.md to be much more restrictive ("apply only
   when evidence is overwhelming and a human grader would also apply").

4. **Systematic under-scoring** (0011 -440, 0009 -160, inep-002 -320). Model is conservative
   relative to human reference. Two hypotheses:
   - prompt fragments copy literally the matrix descriptors → model regurgitates them →
     no real evaluation, just template selection
   - 80B param model genuinely under-calibrated for ENEM

## Action plan (Phase 3 loop — no Phase 4 code)

- [X] Bump `DEFAULT_TIMEOUT_S` 75 → 120 in `single_grader.py`.
- [ ] Iterate prompt to v1.0.1 in `src/corrector/prompts/v1.0.1/`:
  - Rewrite `system.md` with 3 few-shot examples (one nota-1000, one ~600, one ~200).
  - Rewrite `eliminatory_check.md` with explicit "apply ONLY if [hard criterion]" framing.
  - Reduce verbosity of `per_competency_assembly.md`; trim per-competency fragments.
- [ ] Re-run gate. Target MAE ≤ 120 + zero INEP failures.
- [ ] If still failing after 2 prompt iterations: switch to `qwen3.5:397b` (more reasoning
  headroom) OR `gpt-oss:120b` (different reasoning lineage).

## Phase 4 gate

Phase 4 (auth/DB/API) starts **only after** this report shows PASS on:
- overall MVP-tier (MAE ≤ 120, per-comp MAE ≤ 60, ≥ 70% within ±120)
- per-band MVP-tier (each of 3 bands clears overall thresholds)
- zero INEP failures
