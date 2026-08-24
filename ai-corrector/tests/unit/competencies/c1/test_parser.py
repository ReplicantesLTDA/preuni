"""T034: per-competency parser contract (C1 representative; C2..C5 use same code)."""

from __future__ import annotations

import pytest

from src.corrector.competencies._common import CompetencyParseError, parse_competency

ESSAY = (
    "A inclusão digital de pessoas idosas exige ação coordenada do Estado.\n"
    "Por meio de cursos gratuitos em telecentros e parcerias com escolas técnicas,\n"
    "será possível mitigar o letramento digital insuficiente entre os mais velhos."
)


def _valid_below_ceiling() -> dict[str, str | int]:
    return {
        "score": 160,
        "excerpt": "A inclusão digital de pessoas idosas",
        "justification_pt_br": "Domínio bom da norma culta, com poucos desvios.",
        "improvement_path_pt_br": "Revisar pontuação em períodos compostos.",
    }


def _valid_at_ceiling() -> dict[str, str | int]:
    return {
        "score": 200,
        "excerpt": "A inclusão digital de pessoas idosas",
        "justification_pt_br": "Excelente domínio.",
    }


def test_valid_below_ceiling_parses() -> None:
    entry = parse_competency("c1", _valid_below_ceiling(), essay_text=ESSAY)
    assert entry.code == "c1"
    assert entry.score == 160
    assert entry.improvement_path_pt_br is not None


def test_valid_at_ceiling_parses() -> None:
    entry = parse_competency("c1", _valid_at_ceiling(), essay_text=ESSAY)
    assert entry.score == 200
    assert entry.improvement_path_pt_br is None


def test_invalid_score_rejected() -> None:
    bad = _valid_below_ceiling() | {"score": 100}
    with pytest.raises(CompetencyParseError):
        parse_competency("c1", bad, essay_text=ESSAY)


def test_score_string_rejected() -> None:
    bad = _valid_below_ceiling() | {"score": "160"}
    with pytest.raises(CompetencyParseError):
        parse_competency("c1", bad, essay_text=ESSAY)


def test_missing_excerpt_rejected() -> None:
    bad = _valid_below_ceiling()
    bad.pop("excerpt")
    with pytest.raises(CompetencyParseError):
        parse_competency("c1", bad, essay_text=ESSAY)


def test_excerpt_not_in_essay_rejected() -> None:
    bad = _valid_below_ceiling() | {"excerpt": "trecho inexistente na redação"}
    with pytest.raises(CompetencyParseError):
        parse_competency("c1", bad, essay_text=ESSAY)


def test_improvement_path_required_when_below_200() -> None:
    bad = _valid_below_ceiling()
    bad.pop("improvement_path_pt_br")
    with pytest.raises(CompetencyParseError):
        parse_competency("c1", bad, essay_text=ESSAY)


def test_improvement_path_forbidden_at_200() -> None:
    bad = _valid_at_ceiling() | {"improvement_path_pt_br": "vazio"}
    with pytest.raises(CompetencyParseError):
        parse_competency("c1", bad, essay_text=ESSAY)


def test_excerpt_normalization_tolerates_whitespace() -> None:
    # excerpt has collapsed whitespace; essay has the original line breaks.
    raw = _valid_below_ceiling() | {
        "excerpt": "letramento digital insuficiente entre os mais velhos",
    }
    parse_competency("c1", raw, essay_text=ESSAY)
