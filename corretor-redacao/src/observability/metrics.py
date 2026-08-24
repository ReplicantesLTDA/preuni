"""Prometheus metrics and monitoring."""

from prometheus_client import CollectorRegistry, Counter, Histogram, Info

# Create a global registry
REGISTRY = CollectorRegistry()

# Application info
app_info = Info(
    "corretor_app_info",
    "Corretor Redação application information",
    registry=REGISTRY,
)

# Correction pipeline metrics
correction_throughput = Counter(
    "correction_throughput_total",
    "Total number of corrections processed",
    labelnames=["status", "user_tier"],
    registry=REGISTRY,
)

e2e_latency_seconds = Histogram(
    "correction_e2e_latency_seconds",
    "End-to-end correction latency in seconds",
    buckets=(5, 10, 30, 60, 90, 120),
    registry=REGISTRY,
)

llm_latency_seconds = Histogram(
    "correction_llm_latency_seconds",
    "LLM call latency in seconds",
    buckets=(1, 5, 10, 30, 60),
    registry=REGISTRY,
)

llm_cost_usd_total = Counter(
    "correction_llm_cost_usd_total",
    "Total cost of LLM calls in USD",
    registry=REGISTRY,
)

quota_rejections_total = Counter(
    "correction_quota_rejections_total",
    "Total quota rejections",
    labelnames=["user_tier", "reason"],
    registry=REGISTRY,
)

errors_total = Counter(
    "correction_errors_total",
    "Total errors by error code",
    labelnames=["error_code"],
    registry=REGISTRY,
)

prevalidation_failures = Counter(
    "correction_prevalidation_failures_total",
    "Pre-validation failures by reason",
    labelnames=["reason"],
    registry=REGISTRY,
)

schema_violations = Counter(
    "correction_schema_violations_total",
    "Schema validation violations",
    labelnames=["attempt"],
    registry=REGISTRY,
)

db_query_latency_seconds = Histogram(
    "db_query_latency_seconds",
    "Database query latency in seconds",
    labelnames=["query_type"],
    buckets=(0.001, 0.01, 0.05, 0.1, 0.5, 1.0),
    registry=REGISTRY,
)


def initialize_app_info(version: str, commit: str = "unknown") -> None:
    """Initialize application info metric."""
    app_info.info(
        {
            "version": version,
            "commit": commit,
        }
    )
