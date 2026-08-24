"""Shared per-competency parsing primitives.

Used by `c1/parser.py`..`c5/parser.py`. The pipeline never imports this directly;
only competency modules do (Constitution Article XI: orchestrator is empty of
scoring logic).
"""

from __future__ import annotations

import re
import unicodedata
from dataclasses import dataclass
from typing import Any

VALID_SCORES: frozenset[int] = frozenset({0, 40, 80, 120, 160, 200})


class CompetencyParseError(ValueError):
    """Raised when a competency entry violates the v1 contract.

    The pipeline catches this and counts it as a schema failure, triggering the
    Constitution-IV-mandated corrective retry.
    """


@dataclass(slots=True, frozen=True)
class CompetencyEntry:
    code: str  # "c1".."c5"
    score: int
    excerpt: str
    justification_pt_br: str
    improvement_path_pt_br: str | None


_PUNCT_RE = re.compile(r"[‘’“”\"'`\(\)\[\]\{\}.,;:!?\-—–…]")
_WHITESPACE_RE = re.compile(r"\s+")


def normalize_for_substring(text: str) -> str:
    """Tolerant normalization for excerpt verbatim check.

    Folds case, strips most punctuation, NFC-normalizes accents, collapses whitespace.
    LLMs frequently emit minor variations (smart quotes, dropped punctuation, capitalized
    sentence start) — substring check should tolerate those without losing the "verbatim"
    Constitution V intent (no rewording, no paraphrase).
    """
    s = unicodedata.normalize("NFC", text).casefold()
    s = _PUNCT_RE.sub(" ", s)
    s = _WHITESPACE_RE.sub(" ", s).strip()
    return s


def parse_competency(code: str, raw: Any, *, essay_text: str) -> CompetencyEntry:
    if not isinstance(raw, dict):
        raise CompetencyParseError(f"{code}: not an object")
    score = raw.get("score")
    if not isinstance(score, int) or isinstance(score, bool) or score not in VALID_SCORES:
        raise CompetencyParseError(
            f"{code}: score must be int in {sorted(VALID_SCORES)}, got {score!r}"
        )
    excerpt = raw.get("excerpt")
    if not isinstance(excerpt, str) or not excerpt.strip():
        raise CompetencyParseError(f"{code}: excerpt must be a non-empty string")
    justification = raw.get("justification_pt_br")
    if not isinstance(justification, str) or not justification.strip():
        raise CompetencyParseError(f"{code}: justification_pt_br must be a non-empty string")
    improvement = raw.get("improvement_path_pt_br")
    if score == 200:
        if improvement is not None:
            raise CompetencyParseError(
                f"{code}: improvement_path_pt_br must be omitted when score == 200"
            )
    else:
        if not isinstance(improvement, str) or not improvement.strip():
            raise CompetencyParseError(f"{code}: improvement_path_pt_br required when score < 200")
    if not _excerpt_is_verbatim(excerpt, essay_text):
        raise CompetencyParseError(
            f"{code}: excerpt is not a verbatim substring of the submitted essay"
        )
    return CompetencyEntry(
        code=code,
        score=score,
        excerpt=excerpt,
        justification_pt_br=justification,
        improvement_path_pt_br=improvement if score < 200 else None,
    )


def _excerpt_is_verbatim(excerpt: str, essay_text: str) -> bool:
    needle = normalize_for_substring(excerpt)
    haystack = normalize_for_substring(essay_text)
    if needle in haystack:
        return True
    # Fallback: allow excerpt if any sentence-sized chunk of it appears verbatim,
    # OR if 90% of its tokens appear in haystack in same order. Combats LLMs that
    # ellipsize ("(...)") or drop one word from a long citation.
    # Sub-chunk: split needle by ellipsis / "..." / multiple spaces.
    chunks = [
        c.strip()
        for c in re.split(r"\s*\.{2,}\s*|\s+\(\.{3}\)\s+|\s+\.\.\.\s+", needle)
        if c.strip()
    ]
    for chunk in chunks:
        if len(chunk) >= 20 and chunk in haystack:
            return True
    # 90% in-order token overlap
    needle_tokens = needle.split()
    if len(needle_tokens) < 4:
        return False
    haystack_tokens = haystack.split()
    matched = 0
    j = 0
    for tok in needle_tokens:
        while j < len(haystack_tokens) and haystack_tokens[j] != tok:
            j += 1
        if j < len(haystack_tokens):
            matched += 1
            j += 1
    return matched / len(needle_tokens) >= 0.9


__all__ = [
    "VALID_SCORES",
    "CompetencyEntry",
    "CompetencyParseError",
    "normalize_for_substring",
    "parse_competency",
]
