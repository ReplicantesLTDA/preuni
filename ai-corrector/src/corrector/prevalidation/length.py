"""Length pre-validation (FR-002).

Min: shorter of 500 chars or 7 lines (whichever-shorter rule). Below either =>
too short.
Max: larger of 3500 chars or 50 lines (whichever-larger rule). Above both =>
too long.
"""

from __future__ import annotations

MIN_CHARS = 500
MIN_LINES = 7
MAX_CHARS = 3500
MAX_LINES = 50


class LengthTooShortError(ValueError):
    """Spec error_code: length_too_short."""


class LengthTooLongError(ValueError):
    """Spec error_code: length_too_long."""


def _line_count(text: str) -> int:
    if not text:
        return 0
    # newline-separated lines; trailing empty line not counted
    parts = text.split("\n")
    while parts and not parts[-1].strip():
        parts.pop()
    return len(parts)


def validate_length(text: str) -> None:
    n_chars = len(text)
    n_lines = _line_count(text)

    # Min: passes if chars >= MIN_CHARS OR lines >= MIN_LINES (shorter-of rule).
    if n_chars < MIN_CHARS and n_lines < MIN_LINES:
        raise LengthTooShortError(
            f"essay below minimum length (got {n_chars} chars / {n_lines} lines; "
            f"need ≥ {MIN_CHARS} chars or ≥ {MIN_LINES} lines)"
        )

    # Max: fails only if BOTH chars > MAX_CHARS AND lines > MAX_LINES (larger-of rule).
    if n_chars > MAX_CHARS and n_lines > MAX_LINES:
        raise LengthTooLongError(
            f"essay above maximum length (got {n_chars} chars / {n_lines} lines; "
            f"limit {MAX_CHARS} chars or {MAX_LINES} lines)"
        )


__all__ = [
    "MAX_CHARS",
    "MAX_LINES",
    "MIN_CHARS",
    "MIN_LINES",
    "LengthTooLongError",
    "LengthTooShortError",
    "validate_length",
]
