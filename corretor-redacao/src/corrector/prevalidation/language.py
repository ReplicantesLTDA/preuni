"""pt-BR language detection (FR-003, research R11).

Two-stage check:
1. lingua detects the broad language; rejects non-Portuguese with `language_mismatch`.
2. pt-BR vs pt-PT orthographic / lexical heuristic. Markers favoring pt-BR vs pt-PT inform the
   decision. Borderline cases default to ACCEPT (to keep false-positives low) when text length
   is short.
"""

from __future__ import annotations

from functools import cache
from lingua import Language, LanguageDetectorBuilder


class LanguageMismatchError(ValueError):
    """Spec error_code: language_mismatch."""


# Words/forms strongly more common in pt-BR than pt-PT.
_PT_BR_MARKERS: tuple[str, ...] = (
    "você",
    "vocês",
    "ônibus",
    "trem",
    "bonde",
    "celular",
    "geladeira",
    "açougue",
    "banheiro",
    "calçada",
    "esporte",
    "time",
    "moça",
    "rapaz",
    "menino",
    "grama",
)

# Words/forms strongly more common in pt-PT than pt-BR.
_PT_PT_MARKERS: tuple[str, ...] = (
    "tu próprio",
    "tu próp",
    "autocarro",
    "autocarros",
    "comboio",
    "comboios",
    "telemóvel",
    "frigorífico",
    "talho",
    "casa de banho",
    "passeio",
    "desporto",
    "equipa",
    "rapariga",
    "relva",
    "acção",
    "acções",
    "selecção",
    "óptimo",
    "facto",
    "directo",
    "vais ter de",
    "estás a",
    "está a fazer",
    "junta de freguesia",
    "freguesia",
    "multibanco",
)


@cache
def _detector():  # type: ignore[no-untyped-def]
    return LanguageDetectorBuilder.from_languages(
        Language.PORTUGUESE,
        Language.ENGLISH,
        Language.SPANISH,
    ).build()


def _count_markers(text_lc: str, markers: tuple[str, ...]) -> int:
    return sum(1 for m in markers if m in text_lc)


def validate_pt_br(text: str) -> None:
    lang = _detector().detect_language_of(text)
    if lang != Language.PORTUGUESE:
        raise LanguageMismatchError(
            f"essay language detected as {lang.name if lang else 'unknown'}; expected PORTUGUESE (pt-BR)"
        )

    text_lc = text.lower()
    pt_br_hits = _count_markers(text_lc, _PT_BR_MARKERS)
    pt_pt_hits = _count_markers(text_lc, _PT_PT_MARKERS)

    # Decide: if pt-PT markers outnumber pt-BR markers AND at least one pt-PT marker exists, reject.
    if pt_pt_hits > pt_br_hits and pt_pt_hits >= 1:
        raise LanguageMismatchError(
            f"essay appears to be European Portuguese (pt-PT) rather than Brazilian Portuguese; "
            f"pt-PT markers={pt_pt_hits}, pt-BR markers={pt_br_hits}"
        )


__all__ = ["LanguageMismatchError", "validate_pt_br"]
