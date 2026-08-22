"""T080: POST /auth/logout."""

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
async def test_logout_revokes_refresh_token() -> None:
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

        r = await c.post("/auth/logout", json={"refresh_token": login["refresh_token"]})
        assert r.status_code == 204

        # Subsequent refresh with revoked token must fail.
        r2 = await c.post("/auth/refresh", json={"refresh_token": login["refresh_token"]})
        assert r2.status_code == 401
