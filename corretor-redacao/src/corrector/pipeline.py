"""Public correction pipeline entrypoint.

The orchestrator the workers and CLI both call. Pure composition: no DB, no
HTTP, no logging side-effects (callers wrap their own observability around it).

Grader selection (env `GRADER`):
- `single` (default): SingleGrader, 1 LLM call evaluating all 5 competencies.
- `per_competency`: PerCompetencyGrader, 5 concurrent LLM calls, one per competency.
"""

from __future__ import annotations

import os
from typing import Protocol

from src.corrector.graders.aggregator import aggregate_single
from src.corrector.graders.multi_pass_grader import MultiPassGrader
from src.corrector.graders.per_competency_grader import PerCompetencyGrader
from src.corrector.graders.single_grader import (
    SingleGrader,
    SingleGraderInput,
    SingleGraderResult,
)
from src.corrector.llm.base import LLMProvider


class _Grader(Protocol):
    async def grade(self, inp: SingleGraderInput) -> SingleGraderResult: ...


def _default_grader(provider: LLMProvider) -> _Grader:
    mode = os.environ.get("GRADER", "single").lower()
    if mode == "per_competency":
        return PerCompetencyGrader(provider=provider)
    if mode == "multi_pass":
        return MultiPassGrader(provider=provider)
    return SingleGrader(provider=provider)


async def correct_essay(
    inp: SingleGraderInput,
    *,
    provider: LLMProvider,
    grader: _Grader | None = None,
) -> SingleGraderResult:
    """Run the full correction pipeline for one essay.

    Default: SingleGrader (1 call). Override via env `GRADER=per_competency`
    to use PerCompetencyGrader (5 concurrent calls). Multi-grader (multiple
    full passes) is a separate future swap of the grader+aggregator pair.
    """
    g = grader or _default_grader(provider)
    pass_result = await g.grade(inp)
    return aggregate_single(pass_result)


__all__ = ["correct_essay"]
