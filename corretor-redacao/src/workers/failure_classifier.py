"""Map async error_codes to quota_consumed flag (FR-036 / error_codes.md)."""

from __future__ import annotations

PROVIDER_ERROR_CODES: frozenset[str] = frozenset(
    {
        "provider_rate_limited",
        "provider_timeout",
        "provider_unavailable",
        "schema_violation",
        "internal_error",
    }
)

USER_ERROR_CODES: frozenset[str] = frozenset(
    {
        "language_mismatch",
        "theme_missing_context",
        "length_too_short",
        "length_too_long",
    }
)


def classify_failure(error_code: str) -> bool:
    """Return True if error_code is user-attributable (quota consumed)."""
    return error_code not in PROVIDER_ERROR_CODES


__all__ = ["PROVIDER_ERROR_CODES", "USER_ERROR_CODES", "classify_failure"]
