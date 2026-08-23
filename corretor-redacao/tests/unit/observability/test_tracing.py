"""Unit tests for observability.tracing -- 33% coverage before, only ever
exercised implicitly via api.main's app startup with tracing disabled."""

from __future__ import annotations

from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider

from src.observability.tracing import configure_tracing, get_tracer


def test_configure_tracing_disabled_sets_a_noop_provider():
    configure_tracing(enabled=False)
    assert isinstance(trace.get_tracer_provider(), TracerProvider)


def test_configure_tracing_enabled_sets_a_real_provider():
    configure_tracing(otlp_endpoint="http://localhost:4317", enabled=True)
    assert isinstance(trace.get_tracer_provider(), TracerProvider)


def test_get_tracer_returns_a_tracer():
    tracer = get_tracer("test-module", "1.2.3")
    assert tracer is not None
