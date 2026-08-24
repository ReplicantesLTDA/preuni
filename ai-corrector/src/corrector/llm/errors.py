"""LLM provider error taxonomy.

Research R7: five-class hierarchy that maps 1:1 to the spec's typed failure
codes. The worker catches `LLMError` once at the boundary and routes to the
correct `error_code` + `quota_consumed` value (FR-020, FR-036).
"""

from __future__ import annotations


class LLMError(Exception):
    """Base class for every LLM-provider-attributable failure."""


class RateLimitError(LLMError):
    """Provider rate-limited the call. Spec error_code: provider_rate_limited."""


class TimeoutError(LLMError):
    """Provider exceeded the worker's timeout budget. Spec: provider_timeout."""


class TransientError(LLMError):
    """Transient provider failure (5xx, connection reset). Spec: provider_unavailable."""


class SchemaViolationError(LLMError):
    """LLM output failed JSON Schema validation. Spec: schema_violation.

    Raised by the provider adapter after schema validation fails. The pipeline
    catches it, applies the 2-attempt corrective-retry budget (Constitution IV),
    then re-raises if the budget is exhausted.
    """

    def __init__(self, message: str, *, validator_message: str | None = None) -> None:
        super().__init__(message)
        self.validator_message = validator_message


class FatalError(LLMError):
    """Non-retryable provider error (auth failure, unsupported model, etc.). Spec: internal_error."""


__all__ = [
    "FatalError",
    "LLMError",
    "RateLimitError",
    "SchemaViolationError",
    "TimeoutError",
    "TransientError",
]
