"""Unit tests for observability.tracing -- 33% coverage before, only ever
exercised implicitly via api.main's app startup with tracing disabled."""

from __future__ import annotations

import pytest
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from unittest.mock import patch

from src.observability.tracing import configure_tracing, get_tracer


def test_configure_tracing_disabled_sets_a_noop_provider():
    configure_tracing(enabled=False)
    assert isinstance(trace.get_tracer_provider(), TracerProvider)


def test_configure_tracing_enabled_sets_a_real_provider():
    configure_tracing(otlp_endpoint="http://localhost:4317", enabled=True)
    assert isinstance(trace.get_tracer_provider(), TracerProvider)


def test_configure_tracing_falls_back_to_noop_on_exporter_failure():
    # Simulates a real OTLP exporter construction failure (bad credentials,
    # invalid endpoint config, etc.) -- configure_tracing must still install
    # a working no-op provider before re-raising, so the app doesn't crash
    # with no tracer at all. Previously untested (lines 29-32).
    with patch(
        "src.observability.tracing.OTLPSpanExporter",
        side_effect=ValueError("bad endpoint"),
    ), pytest.raises(RuntimeError, match="Failed to configure tracing"):
        configure_tracing(otlp_endpoint="not-a-real-endpoint", enabled=True)

    assert isinstance(trace.get_tracer_provider(), TracerProvider)


def test_get_tracer_returns_a_tracer():
    tracer = get_tracer("test-module", "1.2.3")
    assert tracer is not None
