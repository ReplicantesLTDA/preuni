"""FastAPI app factory.

Phase 4A scope: app boots, JSON logging middleware, Prometheus `/metrics`,
OpenTelemetry span wrap, health routes. Auth + correction routers populated
in subsequent phases (US1, US2, etc).
"""

from __future__ import annotations

import logging
import os
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException, Request, Response
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from prometheus_client import CONTENT_TYPE_LATEST, generate_latest

from src.api.routes import auth, corrections, health, me
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
        title="Corretor de Redação ENEM",
        version="0.1.0",
        description="B2C API for automated ENEM essay correction.",
        lifespan=lifespan,
    )

    @app.middleware("http")
    async def _structured_log_mw(request: Request, call_next):
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
    async def _typed_http_exc(_: Request, exc: HTTPException):
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
    async def _validation_exc(_: Request, exc: RequestValidationError):
        return JSONResponse(
            status_code=422,
            content={
                "error_code": "bad_request",
                "message": "Corpo da requisição inválido.",
                "details": exc.errors(),
            },
        )

    app.include_router(health.router)
    app.include_router(auth.router)
    app.include_router(me.router)
    app.include_router(corrections.router)

    return app


app = create_app()


__all__ = ["app", "create_app"]
