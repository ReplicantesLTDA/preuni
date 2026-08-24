"""Unit tests for correction_worker's pure error-mapping helpers.

Only exercised indirectly before (via one error case in the integration
suite), leaving most branches of _llm_error_to_code untested.
"""

import pytest

from src.corrector.llm.errors import (
    FatalError,
    RateLimitError,
    SchemaViolationError,
    TimeoutError,
    TransientError,
)
from src.workers.correction_worker import _llm_error_message_pt_br, _llm_error_to_code


@pytest.mark.parametrize(
    ("exc", "expected_code"),
    [
        (RateLimitError("rate limited"), "provider_rate_limited"),
        (TimeoutError("timed out"), "provider_timeout"),
        (TransientError("transient"), "provider_unavailable"),
        (SchemaViolationError("bad schema"), "schema_violation"),
        (FatalError("fatal"), "internal_error"),
    ],
)
def test_llm_error_to_code_maps_each_error_type(exc, expected_code):
    assert _llm_error_to_code(exc) == expected_code


@pytest.mark.parametrize(
    "code",
    [
        "provider_rate_limited",
        "provider_timeout",
        "provider_unavailable",
        "schema_violation",
        "internal_error",
        "language_mismatch",
        "theme_missing_context",
        "length_too_short",
        "length_too_long",
    ],
)
def test_llm_error_message_pt_br_has_a_message_for_every_known_code(code):
    message = _llm_error_message_pt_br(code)
    assert message
    assert message != "Ocorreu um erro durante a correção."


def test_llm_error_message_pt_br_falls_back_for_unknown_codes():
    assert _llm_error_message_pt_br("something_unmapped") == "Ocorreu um erro durante a correção."
