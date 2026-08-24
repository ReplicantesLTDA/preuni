"""T028: length pre-validation (FR-002).

Rules:
- Minimum: shorter of 500 chars OR 7 lines. Below either threshold => too short.
- Maximum: larger of 3500 chars OR 50 lines. Above both => too long.
"""

from __future__ import annotations

import pytest

from src.corrector.prevalidation.length import (
    LengthTooLongError,
    LengthTooShortError,
    validate_length,
)


def _make_lines(n_lines: int, chars_per_line: int) -> str:
    return "\n".join("x" * chars_per_line for _ in range(n_lines))


def test_at_min_threshold_500_chars_7_lines_passes() -> None:
    text = _make_lines(7, 80)  # 7 lines, ~560 chars
    validate_length(text)


def test_exactly_500_chars_one_line_fails_only_when_lines_too_few() -> None:
    text = "x" * 500
    # 500 chars satisfies "shorter of 500 OR 7 lines" — passes
    validate_length(text)


def test_below_min_chars_and_lines_fails() -> None:
    text = _make_lines(3, 50)  # 150 chars, 3 lines
    with pytest.raises(LengthTooShortError):
        validate_length(text)


def test_500_chars_but_one_line_passes() -> None:
    # "shorter of 7 lines or 500 chars" — 500 chars alone is enough.
    text = "x" * 500
    validate_length(text)


def test_7_lines_short_chars_passes() -> None:
    text = _make_lines(7, 5)  # 7 lines, 35 chars
    validate_length(text)


def test_at_max_threshold_3500_chars_passes() -> None:
    text = "x" * 3500
    validate_length(text)


def test_above_max_chars_above_max_lines_fails() -> None:
    text = _make_lines(60, 80)  # 60 lines, ~4800 chars
    with pytest.raises(LengthTooLongError):
        validate_length(text)


def test_above_3500_chars_but_under_50_lines_passes() -> None:
    text = "x" * 4000  # 4000 chars, 1 line — "larger of 3500 or 50 lines" passes
    validate_length(text)


def test_above_50_lines_but_under_3500_chars_passes() -> None:
    text = _make_lines(55, 30)  # 55 lines, 1650 chars — passes
    validate_length(text)


def test_empty_string_fails_short() -> None:
    with pytest.raises(LengthTooShortError):
        validate_length("")
