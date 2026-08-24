"""T020: contract assertions for the LLMProvider abstraction surface."""

from __future__ import annotations

import inspect
from dataclasses import fields
from decimal import Decimal
from typing import Protocol, runtime_checkable

from src.corrector.llm import base, errors


def test_llm_provider_is_a_protocol() -> None:
    assert issubclass(type(base.LLMProvider), type(Protocol))
    # Constitution II: Protocol over ABC so adapters need no inheritance
    assert getattr(base.LLMProvider, "_is_protocol", False)


def test_llm_provider_required_attrs() -> None:
    annotations = base.LLMProvider.__annotations__
    assert "name" in annotations
    assert "model_id" in annotations


def test_llm_provider_complete_structured_signature() -> None:
    sig = inspect.signature(base.LLMProvider.complete_structured)
    expected_kw = {
        "system",
        "user",
        "output_schema",
        "temperature",
        "seed",
        "max_tokens",
        "timeout_s",
    }
    actual = set(sig.parameters.keys()) - {"self"}
    assert expected_kw.issubset(actual), f"missing kwargs: {expected_kw - actual}"


def test_llm_result_dataclass_fields() -> None:
    expected = {
        "parsed_output",
        "raw_text",
        "prompt_tokens",
        "completion_tokens",
        "cost_usd",
        "latency_ms",
        "model_id",
        "inference_params",
    }
    actual = {f.name for f in fields(base.LLMResult)}
    assert expected.issubset(actual), f"missing fields: {expected - actual}"


def test_llm_result_can_be_constructed() -> None:
    r = base.LLMResult(
        parsed_output={"ok": True},
        raw_text='{"ok": true}',
        prompt_tokens=10,
        completion_tokens=20,
        cost_usd=Decimal("0.001"),
        latency_ms=42,
        model_id="kimi-k2:1t",
        inference_params={"temperature": 0.1, "seed": 42},
    )
    assert r.parsed_output["ok"] is True
    assert r.model_id == "kimi-k2:1t"


def test_error_hierarchy() -> None:
    # Constitution IV + error taxonomy in research.md R7
    assert issubclass(errors.RateLimitError, errors.LLMError)
    assert issubclass(errors.TimeoutError, errors.LLMError)
    assert issubclass(errors.TransientError, errors.LLMError)
    assert issubclass(errors.SchemaViolationError, errors.LLMError)
    assert issubclass(errors.FatalError, errors.LLMError)
    # All distinct
    classes = {
        errors.RateLimitError,
        errors.TimeoutError,
        errors.TransientError,
        errors.SchemaViolationError,
        errors.FatalError,
    }
    assert len(classes) == 5


def test_schema_violation_carries_validator_message() -> None:
    e = errors.SchemaViolationError("bad output", validator_message="missing 'final_score'")
    assert e.validator_message == "missing 'final_score'"


@runtime_checkable
class _FakeAdapter(Protocol):
    """Adapters do not need to inherit; structural typing must work."""

    name: str
    model_id: str
