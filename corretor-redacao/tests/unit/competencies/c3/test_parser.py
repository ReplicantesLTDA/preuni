"""Smoke test for c3 parser (logic shared with C1)."""

from __future__ import annotations

import pytest

from src.corrector.competencies._common import CompetencyParseError
from src.corrector.competencies.c3 import parse

ESSAY = "Trecho exemplo da redação enviada pelo aluno."


def test_happy_path_at_200() -> None:
    entry = parse(
        {"score": 200, "excerpt": "Trecho exemplo", "justification_pt_br": "ok"},
        essay_text=ESSAY,
    )
    assert entry.code == "c3"
    assert entry.score == 200
    assert entry.improvement_path_pt_br is None


def test_below_ceiling_requires_improvement() -> None:
    entry = parse(
        {
            "score": 160,
            "excerpt": "Trecho exemplo",
            "justification_pt_br": "bom mas com lapsos.",
            "improvement_path_pt_br": "Revisar pontuação.",
        },
        essay_text=ESSAY,
    )
    assert entry.improvement_path_pt_br == "Revisar pontuação."


def test_invalid_score_rejected() -> None:
    with pytest.raises(CompetencyParseError):
        parse(
            {"score": 100, "excerpt": "Trecho exemplo", "justification_pt_br": "x"},
            essay_text=ESSAY,
        )
