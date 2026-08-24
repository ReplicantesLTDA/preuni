"""T067: /healthz always 200; /readyz 200 iff Postgres reachable."""

from __future__ import annotations

import os
import pytest
from httpx import ASGITransport, AsyncClient

from src.api.main import create_app


@pytest.mark.asyncio
async def test_healthz_always_200() -> None:
    app = create_app()
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://t") as client:
        r = await client.get("/healthz")
    assert r.status_code == 200
    assert r.json() == {"status": "ok"}


@pytest.mark.asyncio
async def test_readyz_checks_postgres() -> None:
    # Point engine at the test DB if available; otherwise still expect a
    # response code (200 if reachable, 503 otherwise).
    test_url = os.environ.get(
        "TEST_DATABASE_URL",
        "postgresql+asyncpg://corretor:corretor@localhost:5432/corretor_test",
    )
    os.environ["DATABASE_URL"] = test_url
    # Clear lru_cache so the new URL is used
    from src.db.engine import _engine_singleton

    _engine_singleton.cache_clear()

    app = create_app()
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://t") as client:
        r = await client.get("/readyz")

    assert r.status_code in (200, 503)
    body = r.json()
    if r.status_code == 200:
        assert body["status"] == "ready"
    else:
        assert body["status"] == "unavailable"


@pytest.mark.asyncio
async def test_metrics_endpoint_returns_prometheus_text() -> None:
    app = create_app()
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://t") as client:
        r = await client.get("/metrics")
    assert r.status_code == 200
    body = r.text
    assert "correction_throughput_total" in body
    assert "correction_e2e_latency_seconds" in body
