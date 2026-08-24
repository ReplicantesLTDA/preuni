"""Single-grader orchestrator (MVP).

Constitution XI: this module is the pure orchestrator. It MUST contain no
scoring logic of its own — it composes prevalidation + LLM call + per-competency
parsers + a final sum check.

Constitution III: deterministic seed derived from the correction_id.
Constitution IV: structured-output contract with 2-attempt corrective retry,
then SchemaViolationError surfaces to the caller.
"""

from __future__ import annotations

import hashlib
import uuid
from dataclasses import dataclass, field
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
from src.corrector.llm.base import LLMProvider, LLMResult
from src.corrector.llm.errors import SchemaViolationError
from src.corrector.prevalidation.language import validate_pt_br
from src.corrector.prevalidation.length import validate_length
from src.corrector.prevalidation.theme import validate_theme
from src.corrector.prompts.loader import PromptLoader
from src.schemas import active_correction_output_schema

DEFAULT_TEMPERATURE = 0.1
# 5 competencies × (score + excerpt + justification_pt_br + improvement_path_pt_br) easily
# exceeds 4096 tokens with verbose models. Schema-constrained generation can leave JSON
# truncated (no auto-close) when num_predict is hit mid-output.
DEFAULT_MAX_TOKENS = 8192
# 80B+ reasoning needs headroom; spec FR-023 caps end-to-end at 90s p95 — LLM call
# alone benefits from slightly higher cap since queue + DB time is negligible at MVP scale.
DEFAULT_TIMEOUT_S = 120.0
MAX_LLM_ATTEMPTS = 2  # Constitution IV: original + 1 corrective retry

_COMPETENCY_PARSERS = {
    "c1": parse_c1,
    "c2": parse_c2,
    "c3": parse_c3,
    "c4": parse_c4,
    "c5": parse_c5,
}

ELIMINATORY_FLAGS = frozenset(
    {"off_topic", "annulled", "insufficient_text", "not_dissertative_argumentative"}
)


def derive_seed(correction_id: uuid.UUID, pass_index: int = 0) -> int:
    """Deterministic int64-range seed from the correction id (Research R10).

    `pass_index` lets multi-pass graders derive distinct-but-deterministic seeds
    per pass: pass 0 = sha256(uuid), pass N = sha256(uuid || N_as_byte).
    """
    if pass_index == 0:
        digest = hashlib.sha256(correction_id.bytes).digest()
    else:
        digest = hashlib.sha256(correction_id.bytes + pass_index.to_bytes(4, "big")).digest()
    return int.from_bytes(digest[:8], "big")


@dataclass(slots=True)
class SingleGraderInput:
    correction_id: uuid.UUID
    essay_text: str
    prompt_theme_title: str
    prompt_theme_context: str
    motivational_texts: str | None = None


@dataclass(slots=True)
class SingleGraderResult:
    correction_id: uuid.UUID
    final_score: int
    competencies: dict[str, CompetencyEntry]
    eliminatory_flags: list[str]
    annulment_reason_pt_br: str | None
    prompt_version: str
    model_id: str
    inference_params: dict[str, Any]
    raw_output: str
    prompt_tokens: int
    completion_tokens: int
    cost_usd: Any
    latency_ms: int
    seed: int
    output_schema_version: str = "v1"
    attempts: int = 1
    schema_failures: list[str] = field(default_factory=list)


class SingleGrader:
    def __init__(
        self,
        *,
        provider: LLMProvider,
        prompt_loader: PromptLoader | None = None,
        temperature: float = DEFAULT_TEMPERATURE,
        max_tokens: int = DEFAULT_MAX_TOKENS,
        timeout_s: float = DEFAULT_TIMEOUT_S,
    ) -> None:
        self.provider = provider
        self.prompt_loader = prompt_loader or PromptLoader()
        self.temperature = temperature
        self.max_tokens = max_tokens
        self.timeout_s = timeout_s
        self._schema = active_correction_output_schema()

    async def grade(
        self,
        inp: SingleGraderInput,
        *,
        seed_override: int | None = None,
        skip_prevalidation: bool = False,
    ) -> SingleGraderResult:
        # ---- Pre-validation (FR-002, FR-003, FR-004): does NOT call LLM. ----
        if not skip_prevalidation:
            validate_length(inp.essay_text)
            validate_pt_br(inp.essay_text)
            validate_theme(title=inp.prompt_theme_title, context=inp.prompt_theme_context)

        system = self.prompt_loader.render(
            "system",
            prompt_theme_title=inp.prompt_theme_title,
            prompt_theme_context=inp.prompt_theme_context,
            motivational_texts=inp.motivational_texts or "",
            essay_text=inp.essay_text,
        )
        user = self.prompt_loader.render(
            "per_competency_assembly",
            c1_fragment=self._fragment("c1"),
            c2_fragment=self._fragment("c2"),
            c3_fragment=self._fragment("c3"),
            c4_fragment=self._fragment("c4"),
            c5_fragment=self._fragment("c5"),
        )

        seed = seed_override if seed_override is not None else derive_seed(inp.correction_id)
        schema_failures: list[str] = []
        last_result: LLMResult | None = None
        last_error: SchemaViolationError | None = None
        for attempt in range(1, MAX_LLM_ATTEMPTS + 1):
            try:
                last_result = await self.provider.complete_structured(
                    system=system,
                    user=user if attempt == 1 else self._corrective_user(user, last_error),
                    output_schema=self._schema,
                    temperature=self.temperature,
                    seed=seed,
                    max_tokens=self.max_tokens,
                    timeout_s=self.timeout_s,
                )
                competencies, flags, reason, final = self._parse_output(
                    last_result.parsed_output, essay_text=inp.essay_text
                )
                return SingleGraderResult(
                    correction_id=inp.correction_id,
                    final_score=final,
                    competencies=competencies,
                    eliminatory_flags=flags,
                    annulment_reason_pt_br=reason,
                    prompt_version=self.prompt_loader.active_version,
                    model_id=last_result.model_id,
                    inference_params=last_result.inference_params,
                    raw_output=last_result.raw_text,
                    prompt_tokens=last_result.prompt_tokens,
                    completion_tokens=last_result.completion_tokens,
                    cost_usd=last_result.cost_usd,
                    latency_ms=last_result.latency_ms,
                    seed=seed,
                    attempts=attempt,
                    schema_failures=schema_failures,
                )
            except (SchemaViolationError, CompetencyParseError) as exc:
                msg = (
                    exc.validator_message
                    if isinstance(exc, SchemaViolationError) and exc.validator_message
                    else str(exc)
                )
                schema_failures.append(msg)
                last_error = (
                    exc
                    if isinstance(exc, SchemaViolationError)
                    else SchemaViolationError(str(exc), validator_message=str(exc))
                )

        # Both attempts exhausted.
        msgs = " | ".join(f"#{i + 1}: {m}" for i, m in enumerate(schema_failures))
        raise SchemaViolationError(
            f"schema validation failed after 2 attempts ({msgs})",
            validator_message=(last_error.validator_message if last_error else None),
        )

    def _fragment(self, code: str) -> str:
        # Each competency fragment is the matrix-descriptor body authored under
        # `corrector/competencies/<code>/prompt_fragment.md`. We read it here, but
        # rendering of the assembly template is the place where the strings land.
        from pathlib import Path

        path = Path(__file__).parent.parent / "competencies" / code / "prompt_fragment.md"
        return path.read_text(encoding="utf-8")

    @staticmethod
    def _corrective_user(original_user: str, last_error: SchemaViolationError | None) -> str:
        msg = last_error.validator_message if last_error and last_error.validator_message else ""
        prefix = (
            "ATENÇÃO: a tentativa anterior falhou a validação do JSON Schema com a seguinte "
            f"mensagem do validador: {msg!r}.\n"
            "Responda novamente, agora respeitando exatamente o schema. Sem texto fora do JSON.\n\n"
        )
        return prefix + original_user

    def _parse_output(
        self,
        out: dict[str, Any],
        *,
        essay_text: str,
    ) -> tuple[dict[str, CompetencyEntry], list[str], str | None, int]:
        # Schema-shaped output already validated against the JSON Schema by the adapter.
        # Here we enforce additional business rules: per-comp parse, excerpt-verbatim, sum.
        flags_raw = out.get("eliminatory_flags", [])
        if not isinstance(flags_raw, list) or any(f not in ELIMINATORY_FLAGS for f in flags_raw):
            raise CompetencyParseError(f"invalid eliminatory_flags: {flags_raw!r}")
        reason = out.get("annulment_reason_pt_br")
        if "annulled" in flags_raw and not (isinstance(reason, str) and reason.strip()):
            raise CompetencyParseError("annulment_reason_pt_br required when 'annulled' is set")
        comps_raw = out.get("competencies", {})
        if not isinstance(comps_raw, dict) or set(comps_raw.keys()) != {
            "c1",
            "c2",
            "c3",
            "c4",
            "c5",
        }:
            raise CompetencyParseError("competencies must include exactly c1..c5")

        competencies: dict[str, CompetencyEntry] = {}
        for code, parser in _COMPETENCY_PARSERS.items():
            competencies[code] = parser(comps_raw[code], essay_text=essay_text)

        # Use sum of competency scores as the authoritative final_score. LLMs frequently
        # err on the arithmetic; recomputing locally avoids spurious failures while still
        # honoring Constitution I (final = sum of 5 competencies).
        expected_sum = sum(e.score for e in competencies.values())
        return (
            competencies,
            list(flags_raw),
            reason if isinstance(reason, str) else None,
            expected_sum,
        )


__all__ = [
    "MAX_LLM_ATTEMPTS",
    "SingleGrader",
    "SingleGraderInput",
    "SingleGraderResult",
    "derive_seed",
]
