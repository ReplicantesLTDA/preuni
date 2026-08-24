"""Orchestrator contract for MultiPassGrader.

Runs SingleGrader N times with distinct seeds and aggregates per the
ENEM-inspired closest-two rule; a 3rd pass runs only on divergence. This
grader is selectable via GRADER=multi_pass but was otherwise entirely
untested (40% coverage) -- these mirror the SingleGrader contract tests.
"""

from __future__ import annotations

import pytest
import uuid
from typing import Any

from src.corrector.graders.multi_pass_grader import MultiPassGrader
from src.corrector.graders.single_grader import SingleGraderInput
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


def _output(score: int) -> dict[str, Any]:
    return {
        "eliminatory_flags": [],
        "competencies": {
            code: {
                "score": score,
                "excerpt": "A inclusão digital de pessoas idosas",
                "justification_pt_br": "bom desempenho com lapsos.",
                "improvement_path_pt_br": "revisar pontuação.",
            }
            for code in ("c1", "c2", "c3", "c4", "c5")
        },
        "final_score": score * 5,
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
async def test_two_close_passes_are_averaged_without_a_third_pass() -> None:
    fake = FakeProvider(responses=[_output(160), _output(160)])
    grader = MultiPassGrader(provider=fake)
    result = await grader.grade(_make_input())
    assert result.final_score == 800
    assert len(fake.calls) == 2


@pytest.mark.asyncio
async def test_divergent_passes_trigger_a_third_pass_and_average_closest_two() -> None:
    # Pass 1 scores everything low, pass 2 scores everything high -- divergence
    # exceeds DIVERGENCE_TOTAL/DIVERGENCE_PER_COMP, so a 3rd pass runs. An
    # extra spare response covers a possible retry on the 3rd pass.
    fake = FakeProvider(responses=[_output(40), _output(200), _output(160), _output(160)])
    grader = MultiPassGrader(provider=fake)
    result = await grader.grade(_make_input())
    assert len(fake.calls) >= 3
    # Closest pair among the resulting passes is averaged and snapped to the
    # nearest valid matrix score.
    assert result.final_score == 800


@pytest.mark.asyncio
async def test_prevalidation_runs_once_before_any_llm_call() -> None:
    fake = FakeProvider(responses=[_output(160), _output(160)])
    grader = MultiPassGrader(provider=fake)
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
