"""c2 parser. Delegates to shared helper (Article XI)."""

from __future__ import annotations

from typing import Any

from src.corrector.competencies._common import CompetencyEntry, parse_competency

CODE = "c2"


def parse(raw: Any, *, essay_text: str) -> CompetencyEntry:
    return parse_competency(CODE, raw, essay_text=essay_text)


__all__ = ["CODE", "parse"]
