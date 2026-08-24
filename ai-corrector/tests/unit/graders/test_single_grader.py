"""T040: single-grader orchestrator contract.

Composes: prevalidation → prompt build → LLM call (FakeProvider) → schema
validation → per-competency parse → sum check.
"""

from __future__ import annotations

import pytest
import uuid
from typing import Any

from src.corrector.graders.single_grader import SingleGrader, SingleGraderInput
from src.corrector.llm.errors import (
    SchemaViolationError,
)
from src.corrector.llm.fake import FakeProvider
from src.corrector.prevalidation.language import LanguageMismatchError
from src.corrector.prevalidation.length import LengthTooShortError
from src.corrector.prevalidation.theme import ThemeMissingContextError

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


def _good_output() -> dict[str, Any]:
    return {
        "eliminatory_flags": [],
        "competencies": {
            code: {
                "score": 160,
                "excerpt": "A inclusão digital de pessoas idosas",
                "justification_pt_br": "bom desempenho com lapsos.",
                "improvement_path_pt_br": "revisar pontuação.",
            }
            for code in ("c1", "c2", "c3", "c4", "c5")
        },
        "final_score": 800,
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
async def test_happy_path_completes() -> None:
    fake = FakeProvider(responses=[_good_output()])
    grader = SingleGrader(provider=fake)
    result = await grader.grade(_make_input())
    assert result.final_score == 800
    assert len(result.competencies) == 5
    assert all(c.score == 160 for c in result.competencies.values())
    assert result.eliminatory_flags == []
    assert result.prompt_version
    assert result.model_id == fake.model_id
    assert fake.calls[0]["seed"] is not None


@pytest.mark.asyncio
async def test_prevalidation_short_blocks_llm() -> None:
    fake = FakeProvider(responses=[_good_output()])
    grader = SingleGrader(provider=fake)
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
async def test_prevalidation_language_blocks_llm() -> None:
    fake = FakeProvider(responses=[_good_output()])
    grader = SingleGrader(provider=fake)
    english_essay = "Digital inclusion of elderly people is a pressing contemporary challenge. " * 8
    bad = SingleGraderInput(
        correction_id=uuid.uuid4(),
        essay_text=english_essay,
        prompt_theme_title=THEME_TITLE,
        prompt_theme_context=THEME_CONTEXT,
        motivational_texts=None,
    )
    with pytest.raises(LanguageMismatchError):
        await grader.grade(bad)
    assert fake.calls == []


@pytest.mark.asyncio
async def test_prevalidation_theme_missing_blocks_llm() -> None:
    fake = FakeProvider(responses=[_good_output()])
    grader = SingleGrader(provider=fake)
    bad = SingleGraderInput(
        correction_id=uuid.uuid4(),
        essay_text=ESSAY,
        prompt_theme_title=THEME_TITLE,
        prompt_theme_context="",
        motivational_texts=None,
    )
    with pytest.raises(ThemeMissingContextError):
        await grader.grade(bad)
    assert fake.calls == []


@pytest.mark.asyncio
async def test_schema_violation_triggers_one_retry() -> None:
    fake = FakeProvider(
        responses=[SchemaViolationError("bad", validator_message="x"), _good_output()]
    )
    grader = SingleGrader(provider=fake)
    result = await grader.grade(_make_input())
    assert result.final_score == 800
    assert len(fake.calls) == 2  # original + 1 retry


@pytest.mark.asyncio
async def test_two_schema_failures_exhaust_retry_budget() -> None:
    fake = FakeProvider(
        responses=[
            SchemaViolationError("bad-1", validator_message="x"),
            SchemaViolationError("bad-2", validator_message="y"),
        ]
    )
    grader = SingleGrader(provider=fake)
    with pytest.raises(SchemaViolationError):
        await grader.grade(_make_input())
    assert len(fake.calls) == 2  # 2 attempts total (Constitution IV)


@pytest.mark.asyncio
async def test_sum_mismatch_is_auto_corrected_without_retry() -> None:
    # LLMs frequently mis-add the 5 scores. Pipeline ignores the declared `final_score`
    # and recomputes from competencies; no schema failure, no retry.
    out = _good_output()
    out["final_score"] = 9999  # declared mismatch vs actual 5×160=800
    fake = FakeProvider(responses=[out])
    grader = SingleGrader(provider=fake)
    result = await grader.grade(_make_input())
    assert result.final_score == 800
    assert len(fake.calls) == 1  # no retry; auto-corrected


@pytest.mark.asyncio
async def test_excerpt_not_in_essay_counts_as_schema_failure() -> None:
    out = _good_output()
    out["competencies"]["c1"]["excerpt"] = "trecho fictício inexistente"
    fake = FakeProvider(responses=[out, _good_output()])
    grader = SingleGrader(provider=fake)
    result = await grader.grade(_make_input())
    assert result.final_score == 800
    assert len(fake.calls) == 2
