"""OpenTelemetry tracing configuration."""

from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import SimpleSpanProcessor


def configure_tracing(
    otlp_endpoint: str = "http://localhost:4317",
    enabled: bool = False,
) -> None:
    """Configure OpenTelemetry tracing."""
    if not enabled:
        # Use no-op tracer
        trace.set_tracer_provider(TracerProvider())
        return

    try:
        # Create OTLP exporter
        otlp_exporter = OTLPSpanExporter(endpoint=otlp_endpoint)

        # Create tracer provider
        tracer_provider = TracerProvider()
        tracer_provider.add_span_processor(SimpleSpanProcessor(otlp_exporter))

        # Set global tracer provider
        trace.set_tracer_provider(tracer_provider)
    except Exception as e:
        # Fall back to no-op tracer
        trace.set_tracer_provider(TracerProvider())
        raise RuntimeError(f"Failed to configure tracing: {e}") from e


def get_tracer(name: str, version: str = "0.1.0") -> trace.Tracer:
    """Get a tracer instance."""
    return trace.get_tracer(name, version)
