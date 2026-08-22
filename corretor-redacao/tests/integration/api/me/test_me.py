"""T081: GET /me."""

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


@pytest.mark.asyncio
async def test_me_without_auth_401() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        r = await c.get("/me")
    assert r.status_code == 401
    assert r.json()["error_code"] == "unauthenticated"


@pytest.mark.asyncio
async def test_me_with_invalid_token_401() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        r = await c.get("/me", headers={"Authorization": "Bearer not.a.jwt"})
    assert r.status_code == 401


@pytest.mark.asyncio
async def test_me_happy_path_returns_profile() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        email = f"u-{uuid.uuid4()}@example.com"
        await c.post(
            "/auth/register", json={"email": email, "password": "correctHorseBatteryStaple12!"}
        )
        login = (
            await c.post(
                "/auth/login", json={"email": email, "password": "correctHorseBatteryStaple12!"}
            )
        ).json()

        r = await c.get("/me", headers={"Authorization": f"Bearer {login['access_token']}"})
    assert r.status_code == 200, r.text
    profile = r.json()
    assert profile["email"] == email
    assert profile["tier"] == "free"
    assert "user_id" in profile
    assert profile["quota_used_current_month"] == 0
    assert profile["quota_limit"] == 3
    assert "quota_reset_at" in profile
