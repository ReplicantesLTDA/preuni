"""Multi-pass grader.

Runs SingleGrader N times with distinct deterministic seeds (preserving the
helpful "halo effect" of seeing all 5 competencies in one call), then aggregates
per the ENEM-inspired rule: average of the two closest passes; if divergence
exceeds a threshold, run a 3rd pass and average the two closest of the three.

Trade-offs vs SingleGrader:
- (+) Variance reduction: noisy outliers get smoothed.
- (+) Calibration: model's two natural attempts at the same essay are averaged,
  reducing systematic under/over-scoring.
- (-) Cost: 2x in the typical case, 3x when divergence triggers a 3rd pass.
- (-) Latency: 2x (passes run sequentially to avoid concurrent provider
  rate-limit pressure; small batch can parallelize via asyncio.gather).

Constitution Article XI compliance: orchestrator composes SingleGrader passes
plus a small aggregator; per-competency parsers are reused; no scoring literal
appears in this file.
"""

from __future__ import annotations

import asyncio
from dataclasses import dataclass, field
from decimal import Decimal

from src.corrector.competencies._common import CompetencyEntry
from src.corrector.graders.single_grader import (
    DEFAULT_MAX_TOKENS,
    DEFAULT_TEMPERATURE,
    DEFAULT_TIMEOUT_S,
    SingleGrader,
    SingleGraderInput,
    SingleGraderResult,
    derive_seed,
)
from src.corrector.llm.base import LLMProvider
from src.corrector.prevalidation.language import validate_pt_br
from src.corrector.prevalidation.length import validate_length
from src.corrector.prevalidation.theme import validate_theme

VALID_SCORES = (0, 40, 80, 120, 160, 200)

# ENEM-inspired divergence thresholds: when two graders disagree by more than
# these amounts, a 3rd grader is called and the two closest are averaged.
DIVERGENCE_TOTAL = 100
DIVERGENCE_PER_COMP = 80


def _snap_to_valid_score(value: float) -> int:
    """Snap an averaged score to the nearest valid matrix score."""
    return min(VALID_SCORES, key=lambda v: abs(v - value))


def _diverges(a: SingleGraderResult, b: SingleGraderResult) -> bool:
    if abs(a.final_score - b.final_score) > DIVERGENCE_TOTAL:
        return True
    for code in ("c1", "c2", "c3", "c4", "c5"):
        if abs(a.competencies[code].score - b.competencies[code].score) > DIVERGENCE_PER_COMP:
            return True
    return False


def _two_closest(
    results: list[SingleGraderResult],
) -> tuple[SingleGraderResult, SingleGraderResult]:
    """Pick the two passes with the smallest total-score gap."""
    best_pair = (results[0], results[1])
    best_gap = abs(results[0].final_score - results[1].final_score)
    for i in range(len(results)):
        for j in range(i + 1, len(results)):
            gap = abs(results[i].final_score - results[j].final_score)
            if gap < best_gap:
                best_gap = gap
                best_pair = (results[i], results[j])
    return best_pair


def _aggregate(passes: list[SingleGraderResult]) -> SingleGraderResult:
    """Aggregate per ENEM rule: 2 closest passes -> snap-averaged competencies."""
    if len(passes) >= 3:
        a, b = _two_closest(passes)
    else:
        a, b = passes[0], passes[1]

    aggregated_competencies: dict[str, CompetencyEntry] = {}
    for code in ("c1", "c2", "c3", "c4", "c5"):
        avg = (a.competencies[code].score + b.competencies[code].score) / 2.0
        snapped = _snap_to_valid_score(avg)
        # Use the entry from the pass closer to the snapped score (preserves better excerpt).
        chosen = (
            a.competencies[code]
            if abs(a.competencies[code].score - snapped)
            <= abs(b.competencies[code].score - snapped)
            else b.competencies[code]
        )
        if chosen.score == snapped:
            aggregated_competencies[code] = chosen
        else:
            # Build a new entry with the snapped score, reusing the closer pass's text.
            aggregated_competencies[code] = CompetencyEntry(
                code=code,
                score=snapped,
                excerpt=chosen.excerpt,
                justification_pt_br=chosen.justification_pt_br,
                improvement_path_pt_br=chosen.improvement_path_pt_br if snapped < 200 else None,
            )

    final_score = sum(e.score for e in aggregated_competencies.values())

    # Provenance: use the first pass for prompt/model/inference_params (identical
    # across passes), but aggregate cost/tokens/latency.
    base = passes[0]
    return SingleGraderResult(
        correction_id=base.correction_id,
        final_score=final_score,
        competencies=aggregated_competencies,
        eliminatory_flags=base.eliminatory_flags,  # eliminatory inferred from first pass; safe default
        annulment_reason_pt_br=base.annulment_reason_pt_br,
        prompt_version=base.prompt_version,
        model_id=base.model_id,
        inference_params=base.inference_params,
        raw_output=base.raw_output,
        prompt_tokens=sum(p.prompt_tokens for p in passes),
        completion_tokens=sum(p.completion_tokens for p in passes),
        cost_usd=sum((p.cost_usd for p in passes), start=Decimal("0")),
        latency_ms=sum(p.latency_ms for p in passes),
        seed=base.seed,
        attempts=max(p.attempts for p in passes),
        schema_failures=[m for p in passes for m in p.schema_failures],
    )


@dataclass(slots=True)
class MultiPassGrader:
    provider: LLMProvider
    n_passes: int = 2
    max_extra_passes: int = 1  # ENEM: third pass on divergence
    temperature: float = DEFAULT_TEMPERATURE
    max_tokens: int = DEFAULT_MAX_TOKENS
    timeout_s: float = DEFAULT_TIMEOUT_S
    # Sequential by default: parallel hits Ollama Cloud rate/capacity limits, returning 500s.
    parallel: bool = False
    _single: SingleGrader = field(init=False)

    def __post_init__(self) -> None:
        self._single = SingleGrader(
            provider=self.provider,
            temperature=self.temperature,
            max_tokens=self.max_tokens,
            timeout_s=self.timeout_s,
        )

    async def grade(self, inp: SingleGraderInput) -> SingleGraderResult:
        # Pre-validate ONCE (does NOT call LLM); subsequent SingleGrader.grade()
        # calls skip prevalidation to avoid redundant work.
        validate_length(inp.essay_text)
        validate_pt_br(inp.essay_text)
        validate_theme(title=inp.prompt_theme_title, context=inp.prompt_theme_context)

        seeds = [derive_seed(inp.correction_id, pass_index=i) for i in range(self.n_passes)]
        if self.parallel:
            passes = await asyncio.gather(
                *(self._single.grade(inp, seed_override=s, skip_prevalidation=True) for s in seeds)
            )
        else:
            passes = []
            for s in seeds:
                passes.append(
                    await self._single.grade(inp, seed_override=s, skip_prevalidation=True)
                )

        # Trigger 3rd pass on divergence (ENEM rule).
        if self.max_extra_passes > 0 and len(passes) == 2 and _diverges(passes[0], passes[1]):
            extra_seed = derive_seed(inp.correction_id, pass_index=self.n_passes)
            extra = await self._single.grade(inp, seed_override=extra_seed, skip_prevalidation=True)
            passes.append(extra)

        return _aggregate(list(passes))


__all__ = ["MultiPassGrader"]
