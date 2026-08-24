"""Orchestrator contract for PerCompetencyGrader.

5 concurrent LLM calls, one focused on each competency -- selectable via
GRADER=per_competency but entirely untested (41% coverage) until now.
"""

from __future__ import annotations

import pytest
import uuid
from typing import Any

from src.corrector.graders.per_competency_grader import PerCompetencyGrader
from src.corrector.graders.single_grader import SingleGraderInput
from src.corrector.llm.errors import SchemaViolationError
from src.corrector.llm.fake import FakeProvider
from src.corrector.prevalidation.length import LengthTooShortError

ESSAY = (
    "A inclusão digital de pessoas idosas é um desafio brasileiro contemporâneo "
    "que demanda ação coordenada do Estado e da sociedade civil. Você pode notar "
    "que muitos avós ainda têm dificuldade para utilizar aplicativos de banco e "
    "plataformas governamentais essenciais para o exercício pleno da cidadania. "
    "Diante disso, propõe-se uma estratégia coordenada de inclusão digital, com "
    "cursos gratuitos em telecentros e parcerias com escolas técnicas para a "
    "mediação digital intergeracional efetiva entre famílias.\n"
    "Portanto, cabe ao Ministério da Educação, em conjunto com as secretarias "
    "estaduais, articular um programa nacional de letramento digital para idosos, "
    "garantindo formação de mediadores nas escolas técnicas e celulares acessíveis "
    "para reduzir o fosso digital geracional no curto prazo."
)
THEME_TITLE = "Inclusão digital de idosos no Brasil"
THEME_CONTEXT = "A democratização das ferramentas digitais não atingiu uniformemente a população idosa, gerando exclusão funcional do exercício de direitos."


def _entry(score: int = 160) -> dict[str, Any]:
    return {
        "score": score,
        "excerpt": "A inclusão digital de pessoas idosas",
        "justification_pt_br": "bom desempenho com lapsos.",
        "improvement_path_pt_br": "revisar pontuação.",
    }


def _make_input() -> SingleGraderInput:
    return SingleGraderInput(
        correction_id=uuid.uuid4(),
        essay_text=ESSAY,
        prompt_theme_title=THEME_TITLE,
        prompt_theme_context=THEME_CONTEXT,
        motivational_texts=None,
    )


@pytest.mark.asyncio
async def test_happy_path_makes_five_calls_and_sums_the_scores() -> None:
    fake = FakeProvider(responses=[_entry() for _ in range(5)])
    grader = PerCompetencyGrader(provider=fake)
    result = await grader.grade(_make_input())
    assert result.final_score == 800
    assert len(result.competencies) == 5
    assert len(fake.calls) == 5
    assert result.eliminatory_flags == []


@pytest.mark.asyncio
async def test_prevalidation_runs_once_before_any_llm_call() -> None:
    fake = FakeProvider(responses=[_entry() for _ in range(5)])
    grader = PerCompetencyGrader(provider=fake)
    bad = SingleGraderInput(
        correction_id=uuid.uuid4(),
        essay_text="muito curto",
        prompt_theme_title=THEME_TITLE,
        prompt_theme_context=THEME_CONTEXT,
        motivational_texts=None,
    )
    with pytest.raises(LengthTooShortError):
        await grader.grade(bad)
    assert fake.calls == []


@pytest.mark.asyncio
async def test_a_schema_failure_on_one_competency_retries_just_that_call() -> None:
    # 5 competencies each get one clean response, except c1's first attempt
    # is a schema violation, so it needs a corrective retry -- 6 calls total.
    responses = [
        SchemaViolationError("bad", validator_message="x"),
        _entry(),
        _entry(),
        _entry(),
        _entry(),
        _entry(),
    ]
    fake = FakeProvider(responses=responses)
    grader = PerCompetencyGrader(provider=fake)
    result = await grader.grade(_make_input())
    assert result.final_score == 800
    assert len(fake.calls) == 6
    assert result.attempts == 2
