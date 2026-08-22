"""T064: Prometheus registry exposes the metrics enumerated in the plan."""

from __future__ import annotations

from prometheus_client import generate_latest

from src.observability import metrics as m


def test_registry_exposes_required_metrics() -> None:
    expected = {
        "correction_throughput",
        "e2e_latency_seconds",
        "llm_latency_seconds",
        "llm_cost_usd_total",
        "quota_rejections_total",
        "errors_total",
    }
    missing = [name for name in expected if not hasattr(m, name)]
    assert not missing, f"metrics module missing attributes: {missing}"


def test_registry_renders_text_format() -> None:
    out = generate_latest(m.REGISTRY).decode("utf-8")
    # Counters in Prometheus exposition get _total suffix on the rendered name.
    must_appear = [
        "correction_throughput_total",
        "correction_e2e_latency_seconds",
        "correction_llm_latency_seconds",
        "correction_llm_cost_usd_total",
        "correction_quota_rejections_total",
        "correction_errors_total",
    ]
    for name in must_appear:
        assert name in out, f"metric {name!r} not exposed by REGISTRY"


def test_metrics_can_be_incremented() -> None:
    m.correction_throughput.labels(status="completed", user_tier="free").inc()
    m.e2e_latency_seconds.observe(12.5)
    m.llm_latency_seconds.observe(8.0)
    m.llm_cost_usd_total.inc(0.06)
    m.quota_rejections_total.labels(user_tier="free", reason="exhausted").inc()
    m.errors_total.labels(error_code="provider_timeout").inc()


def test_e2e_latency_buckets_cover_spec_p95() -> None:
    # Spec FR-023: p95 e2e correction latency under 90s.
    # Bucket array should include 90 so we can measure the gate explicitly.
    assert 90 in m.e2e_latency_seconds._upper_bounds or 90.0 in m.e2e_latency_seconds._upper_bounds
