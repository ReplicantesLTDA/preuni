"""FastAPI app factory.

Constitution v2.1.1 / specs/014-constitution-alignment-refactor: this
service is internal-only now (research.md #2, #3). Essay submission,
quota, and identity live in the Go monolith; the Go<->worker hand-off is
DB-mediated via the `correction.correction_jobs` table (research.md #1),
not HTTP, so there is no user-facing submission endpoint here. This app
now serves only liveness/readiness probes and metrics.
"""

from __future__ import annotations

import logging
import os
from collections.abc import AsyncIterator, Awaitable, Callable
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException, Request, Response
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from prometheus_client import CONTENT_TYPE_LATEST, generate_latest

from src.api.routes import health
from src.observability import logging as obs_logging
from src.observability import metrics as obs_metrics
from src.observability import tracing as obs_tracing

log = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    obs_logging.configure_logging()
    obs_tracing.configure_tracing(
        otlp_endpoint=os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
        enabled=os.environ.get("OTEL_ENABLED", "false").lower() == "true",
    )
    log.info("api.startup", extra={"version": app.version})
    yield
    log.info("api.shutdown")


def create_app() -> FastAPI:
    app = FastAPI(
        title="preuni correction service (internal)",
        version="0.1.0",
        description="Internal essay-grading worker process for preuni. No public API surface.",
        lifespan=lifespan,
    )

    @app.middleware("http")
    async def _structured_log_mw(request: Request, call_next: Callable[[Request], Awaitable[Response]]) -> Response:
        from time import perf_counter

        t0 = perf_counter()
        response = await call_next(request)
        elapsed_ms = int((perf_counter() - t0) * 1000)
        log.info(
            "http.request",
            extra={
                "method": request.method,
                "path": request.url.path,
                "status_code": response.status_code,
                "latency_ms": elapsed_ms,
            },
        )
        return response

    @app.get("/metrics", include_in_schema=False)
    async def metrics_endpoint() -> Response:
        return Response(
            content=generate_latest(obs_metrics.REGISTRY),
            media_type=CONTENT_TYPE_LATEST,
        )

    @app.exception_handler(HTTPException)
    async def _typed_http_exc(_: Request, exc: HTTPException) -> JSONResponse:
        # If detail is already a typed-error dict, pass it through; otherwise wrap.
        if isinstance(exc.detail, dict) and "error_code" in exc.detail:
            return JSONResponse(status_code=exc.status_code, content=exc.detail)
        return JSONResponse(
            status_code=exc.status_code,
            content={
                "error_code": "bad_request"
                if exc.status_code == 400
                else f"http_{exc.status_code}",
                "message": str(exc.detail),
            },
        )

    @app.exception_handler(RequestValidationError)
    async def _validation_exc(_: Request, exc: RequestValidationError) -> JSONResponse:
        return JSONResponse(
            status_code=422,
            content={
                "error_code": "bad_request",
                "message": "Corpo da requisição inválido.",
                "details": exc.errors(),
            },
        )

    app.include_router(health.router)

    return app


app = create_app()


__all__ = ["app", "create_app"]
