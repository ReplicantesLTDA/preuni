"""Liveness + readiness probes."""

from __future__ import annotations

from fastapi import APIRouter, Response, status
from sqlalchemy import text

from src.db.engine import get_engine

router = APIRouter(tags=["health"])


@router.get("/healthz", include_in_schema=False)
async def healthz() -> dict[str, str]:
    """Liveness: returns 200 if the process is up. No external deps checked."""
    return {"status": "ok"}


@router.get("/readyz", include_in_schema=False)
async def readyz(response: Response) -> dict[str, str]:
    """Readiness: returns 200 only if Postgres is reachable."""
    try:
        engine = get_engine()
        async with engine.connect() as conn:
            await conn.execute(text("SELECT 1"))
        return {"status": "ready"}
    except Exception as exc:
        response.status_code = status.HTTP_503_SERVICE_UNAVAILABLE
        return {"status": "unavailable", "reason": type(exc).__name__}
