# Walking-Skeleton Gate Run 2 — FAIL (improved)

**Date**: 2026-05-29
**Model**: `qwen3-next:80b` (Ollama Cloud)
**Prompt version**: 1.0.1
**Wall time**: ~17 min
**Phase 3 verdict**: **NO-GO**, but materially improved vs run 1.

## Aggregate

| metric | run 1 (v1.0.0) | run 2 (v1.0.1) | gate |
|---|---|---|---|
| successes | 11/15 | 13/15 | — |
| overall MAE total | ~229 | **157** | ≤ 120 |
| overall MAE per-comp | — | **34** | ≤ 60 ✓ |
| overall hit ±120 | — | **54%** | ≥ 70% |
| band 0-400 MAE | — | **320** | ≤ 120 |
| band 401-700 MAE | — | **60** | ≤ 120 ✓ |
| band 701-1000 MAE | — | **104** | ≤ 120 ✓ |
| INEP failures | 1 | **1** (inep-002) | 0 |

## Per-essay

| slug | expected | run1 | run2 | observation |
|---|---|---|---|---|
| 0001 | 400 | 0 | 600 | false `insufficient_text` gone; now +200 over |
| 0002 | 120 | timeout | 680 | timeout gone; gross over-score |
| 0003 | 400 | 680 | 760 | got worse |
| 0004 | 240 | 80 | 400 | same magnitude, sign flipped |
| 0005 | 600 | schema | 520 | schema gone ✓ |
| 0006 | 560 | 640 | **560** | perfect ✓ |
| 0007 | 440 | 0 | 600 | false eliminatory gone |
| 0008 | 560 | 0 | **560** | perfect ✓ |
| 0009 | 800 | 640 | **800** | perfect ✓ |
| 0010 | 1000 | 960 | 920 | slightly worse |
| 0011 | 880 | 440 | schema | regressed |
| 0012 | 840 | timeout | schema | regressed |
| inep-001 | 1000 | 1000 | 960 | slightly worse |
| inep-002 | 1000 | 680 | 640 | slightly worse |
| inep-003 | 1000 | 960 | 960 | same |

## What worked

- Anti-eliminatory framing removed false `insufficient_text` on 0001, 0007, 0008.
- Calibration anchors brought 0006/0008/0009 to perfect.
- Timeout 75 → 120 s eliminated false-timeouts (0002, 0012).
- Compact tabular fragments seem to help mid/upper-band reasoning.

## What broke

- Calibration sentence "nota inicial mental ~520-560" biased low-band essays upward.
  0002 (expected 120) jumped to 680; 0003 (expected 400) jumped to 760.
- Schema-violation regression on 0011, 0012 — model emits prose around JSON when given a
  verbose prompt.
- `c5` of inep-002 stays at 0 — model may be flagging Direitos-Humanos incorrectly.

## v1.0.2 changes (in flight)

- Remove "nota mental 520-560" anchor; replace with explicit "no typical starting point;
  use the 4 anchors".
- Add anchor D (nota 240) to combat low-band over-scoring.
- Hard-stop JSON-only language: "no markdown, no ```json, no prose before/after, just `{`
  to `}`".
- Loosen DH zero-trigger in C5: only literal pena de morte / tortura / segregação obrigatória;
  critiques of public programs (Escola Sem Partido, MBL, etc.) are NOT DH violations.
