"""T079: POST /auth/refresh."""

from __future__ import annotations

import os
import pytest
import uuid
from httpx import ASGITransport, AsyncClient


def _setup_env() -> None:
    os.environ.setdefault("JWT_SECRET_KEY", "test-secret-key-with-enough-bytes-for-hs256")
    os.environ["DATABASE_URL"] = os.environ.get(
        "TEST_DATABASE_URL",
        "postgresql+asyncpg://corretor:corretor@localhost:5432/corretor_test",
    )
    from src.db.engine import _engine_singleton

    _engine_singleton.cache_clear()


async def _register_login(client: AsyncClient) -> dict:
    email = f"u-{uuid.uuid4()}@example.com"
    await client.post(
        "/auth/register", json={"email": email, "password": "correctHorseBatteryStaple12!"}
    )
    r = await client.post(
        "/auth/login", json={"email": email, "password": "correctHorseBatteryStaple12!"}
    )
    return r.json()


@pytest.mark.asyncio
async def test_refresh_rotates_pair() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        pair = await _register_login(c)
        r = await c.post("/auth/refresh", json={"refresh_token": pair["refresh_token"]})
    assert r.status_code == 200, r.text
    new_pair = r.json()
    # Refresh tokens MUST rotate (old revoked). Access tokens may match when
    # issued within the same second (JWT iat is integer seconds); we only
    # assert what's load-bearing.
    assert new_pair["refresh_token"] != pair["refresh_token"]
    assert "access_token" in new_pair


@pytest.mark.asyncio
async def test_refresh_with_invalid_token_401() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        r = await c.post("/auth/refresh", json={"refresh_token": "bogus"})
    assert r.status_code == 401
    assert r.json()["error_code"] == "invalid_token"


@pytest.mark.asyncio
async def test_refresh_rejects_rotated_token() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        pair = await _register_login(c)
        # First rotation succeeds.
        r1 = await c.post("/auth/refresh", json={"refresh_token": pair["refresh_token"]})
        assert r1.status_code == 200
        # Second rotation with the OLD (now revoked) token fails.
        r2 = await c.post("/auth/refresh", json={"refresh_token": pair["refresh_token"]})
    assert r2.status_code == 401
