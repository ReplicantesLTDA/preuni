"""T078: POST /auth/login."""

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


async def _register_and_get_email(
    client: AsyncClient, password: str = "correctHorseBatteryStaple12!"
) -> str:
    email = f"u-{uuid.uuid4()}@example.com"
    r = await client.post("/auth/register", json={"email": email, "password": password})
    assert r.status_code == 201
    return email


@pytest.mark.asyncio
async def test_login_happy_path_returns_token_pair() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        email = await _register_and_get_email(c)
        r = await c.post(
            "/auth/login", json={"email": email, "password": "correctHorseBatteryStaple12!"}
        )
    assert r.status_code == 200, r.text
    body = r.json()
    assert "access_token" in body
    assert "refresh_token" in body
    assert body["token_type"] == "Bearer"
    assert isinstance(body["access_token_expires_in"], int)
    assert isinstance(body["refresh_token_expires_in"], int)


@pytest.mark.asyncio
async def test_login_invalid_password_401() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        email = await _register_and_get_email(c)
        r = await c.post("/auth/login", json={"email": email, "password": "wrongPassword12345"})
    assert r.status_code == 401
    assert r.json()["error_code"] == "invalid_credentials"


@pytest.mark.asyncio
async def test_login_unknown_email_401_same_message() -> None:
    """Avoid user-enumeration: unknown email returns same shape as wrong password."""
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        r = await c.post(
            "/auth/login", json={"email": "nobody@example.com", "password": "AnyPassword12!"}
        )
    assert r.status_code == 401
    assert r.json()["error_code"] == "invalid_credentials"
