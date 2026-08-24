"""Per-competency grader (v2.0.0).

5 LLM calls per essay, one focused on each competency. Each call returns ONE
competency entry; orchestrator aggregates into the same `SingleGraderResult`
shape used elsewhere downstream.

Trade-offs vs SingleGrader (one call, all 5 competencies):
- (+) Variance: each competency scored independently, no internal-coherence
  pressure to flatten scores to the same level.
- (+) Focus: prompt is exclusively about ONE dimension; model less likely to
  conflate signals.
- (-) Latency: 5x. Mitigated by running calls concurrently via asyncio.gather.
- (-) Cost: 5x tokens, mostly input (system prompt repeated).
- (-) Failure mode: any of 5 calls failing fails the whole essay.

Constitution Article XI compliance: orchestrator is still empty of scoring
logic; each per-competency call uses its own focused prompt; competency
parsers in `corrector/competencies/<code>/parser.py` validate the returned
entry; aggregation is identity.
"""

from __future__ import annotations

import asyncio
from dataclasses import dataclass
from pathlib import Path
from typing import Any

from src.corrector.competencies._common import (
    CompetencyEntry,
    CompetencyParseError,
)
from src.corrector.competencies.c1 import parse as parse_c1
from src.corrector.competencies.c2 import parse as parse_c2
from src.corrector.competencies.c3 import parse as parse_c3
from src.corrector.competencies.c4 import parse as parse_c4
from src.corrector.competencies.c5 import parse as parse_c5
from src.corrector.graders.single_grader import (
    DEFAULT_MAX_TOKENS,
    DEFAULT_TEMPERATURE,
    DEFAULT_TIMEOUT_S,
    MAX_LLM_ATTEMPTS,
    SingleGraderInput,
    SingleGraderResult,
    derive_seed,
)
from src.corrector.llm.base import LLMProvider, LLMResult
from src.corrector.llm.errors import SchemaViolationError
from src.corrector.prevalidation.language import validate_pt_br
from src.corrector.prevalidation.length import validate_length
from src.corrector.prevalidation.theme import validate_theme

COMPETENCY_NAMES: dict[str, str] = {
    "c1": "Domínio da modalidade escrita formal da língua portuguesa",
    "c2": "Compreensão da proposta + tipo dissertativo-argumentativo",
    "c3": "Seleção, relação, organização e interpretação dos argumentos",
    "c4": "Mecanismos linguísticos para a construção da argumentação",
    "c5": "Proposta de intervenção respeitando os direitos humanos",
}

_PARSERS = {
    "c1": parse_c1,
    "c2": parse_c2,
    "c3": parse_c3,
    "c4": parse_c4,
    "c5": parse_c5,
}

# Schema for a single competency entry returned by one per-competency call.
PER_COMPETENCY_OUTPUT_SCHEMA: dict[str, Any] = {
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "Per-competency LLM output (v2)",
    "type": "object",
    "additionalProperties": True,
    "required": ["score", "excerpt", "justification_pt_br"],
    "properties": {
        "score": {"type": "integer", "enum": [0, 40, 80, 120, 160, 200]},
        "excerpt": {"type": "string", "minLength": 1},
        "justification_pt_br": {"type": "string", "minLength": 1},
        "improvement_path_pt_br": {"type": ["string", "null"]},
    },
}


@dataclass(slots=True)
class _PromptBundle:
    system: str
    user_per_competency: dict[str, str]


class PerCompetencyGrader:
    """5-call grader: one focused LLM call per competency, run concurrently."""

    def __init__(
        self,
        *,
        provider: LLMProvider,
        prompts_dir: Path | None = None,
        version: str = "2.0.0",
        temperature: float = DEFAULT_TEMPERATURE,
        max_tokens: int = DEFAULT_MAX_TOKENS,
        timeout_s: float = DEFAULT_TIMEOUT_S,
    ) -> None:
        self.provider = provider
        self.version = version
        self.temperature = temperature
        self.max_tokens = max_tokens
        self.timeout_s = timeout_s
        self._prompts = self._load_prompts(prompts_dir, version)

    @staticmethod
    def _load_prompts(prompts_dir: Path | None, version: str) -> _PromptBundle:
        root = prompts_dir or Path(__file__).parent.parent / "prompts"
        ver = root / f"v{version}"
        if not ver.is_dir():
            raise FileNotFoundError(f"prompt version dir not found: {ver}")
        system = (ver / "system_per_competency.md").read_text(encoding="utf-8")
        per_comp: dict[str, str] = {}
        for code in COMPETENCY_NAMES:
            per_comp[code] = (ver / "per_competency" / f"{code}.md").read_text(encoding="utf-8")
        return _PromptBundle(system=system, user_per_competency=per_comp)

    async def grade(self, inp: SingleGraderInput) -> SingleGraderResult:
        # Pre-validation (identical to SingleGrader; happens once, not 5x).
        validate_length(inp.essay_text)
        validate_pt_br(inp.essay_text)
        validate_theme(title=inp.prompt_theme_title, context=inp.prompt_theme_context)

        seed = derive_seed(inp.correction_id)

        # 5 concurrent per-competency calls.
        coros = [self._grade_one_competency(code, inp, seed=seed) for code in COMPETENCY_NAMES]
        per_results: list[tuple[str, CompetencyEntry, LLMResult, int]] = await asyncio.gather(
            *coros
        )

        competencies: dict[str, CompetencyEntry] = {
            code: entry for code, entry, _llm, _attempts in per_results
        }
        # Aggregate provenance: take the last call's model + params (all identical).
        last_llm = per_results[-1][2]
        total_latency = sum(r[2].latency_ms for r in per_results)
        total_prompt_tokens = sum(r[2].prompt_tokens for r in per_results)
        total_completion_tokens = sum(r[2].completion_tokens for r in per_results)
        total_cost = sum(
            (r[2].cost_usd for r in per_results), start=last_llm.cost_usd.__class__("0")
        )
        max_attempts_used = max(r[3] for r in per_results)
        schema_failures: list[str] = []  # populated only on failure; per-call diagnostics elsewhere

        final_score = sum(e.score for e in competencies.values())

        return SingleGraderResult(
            correction_id=inp.correction_id,
            final_score=final_score,
            competencies=competencies,
            eliminatory_flags=[],  # per-competency flow does not assess eliminatórios; future hook
            annulment_reason_pt_br=None,
            prompt_version=self.version,
            model_id=last_llm.model_id,
            inference_params=last_llm.inference_params,
            raw_output="",  # per-call raw_text aggregated separately if needed
            prompt_tokens=total_prompt_tokens,
            completion_tokens=total_completion_tokens,
            cost_usd=total_cost,
            latency_ms=total_latency,
            seed=seed,
            attempts=max_attempts_used,
            schema_failures=schema_failures,
        )

    async def _grade_one_competency(
        self,
        code: str,
        inp: SingleGraderInput,
        *,
        seed: int,
    ) -> tuple[str, CompetencyEntry, LLMResult, int]:
        system = self._render_system(code, inp)
        user = self._prompts.user_per_competency[code]
        parser = _PARSERS[code]

        last_error: SchemaViolationError | None = None
        last_result: LLMResult | None = None
        for attempt in range(1, MAX_LLM_ATTEMPTS + 1):
            try:
                last_result = await self.provider.complete_structured(
                    system=system,
                    user=user if attempt == 1 else self._corrective_user(user, last_error),
                    output_schema=PER_COMPETENCY_OUTPUT_SCHEMA,
                    temperature=self.temperature,
                    seed=seed,
                    max_tokens=self.max_tokens,
                    timeout_s=self.timeout_s,
                )
                entry = parser(last_result.parsed_output, essay_text=inp.essay_text)
                return code, entry, last_result, attempt
            except (SchemaViolationError, CompetencyParseError) as exc:
                msg = (
                    exc.validator_message
                    if isinstance(exc, SchemaViolationError) and exc.validator_message
                    else str(exc)
                )
                last_error = (
                    exc
                    if isinstance(exc, SchemaViolationError)
                    else SchemaViolationError(str(exc), validator_message=msg)
                )

        raise SchemaViolationError(
            f"{code}: schema validation failed after {MAX_LLM_ATTEMPTS} attempts",
            validator_message=last_error.validator_message if last_error else None,
        )

    def _render_system(self, code: str, inp: SingleGraderInput) -> str:
        return (
            self._prompts.system.replace("{competency_code}", code.upper())
            .replace("{competency_name}", COMPETENCY_NAMES[code])
            .replace("{prompt_theme_title}", inp.prompt_theme_title)
            .replace("{prompt_theme_context}", inp.prompt_theme_context)
            .replace("{essay_text}", inp.essay_text)
        )

    @staticmethod
    def _corrective_user(original_user: str, last_error: SchemaViolationError | None) -> str:
        msg = last_error.validator_message if last_error and last_error.validator_message else ""
        prefix = (
            "ATENÇÃO: a tentativa anterior falhou a validação com a mensagem: "
            f"{msg!r}. Responda novamente, agora respeitando exatamente o schema. "
            "APENAS JSON, sem texto fora.\n\n"
        )
        return prefix + original_user


__all__ = ["COMPETENCY_NAMES", "PER_COMPETENCY_OUTPUT_SCHEMA", "PerCompetencyGrader"]
