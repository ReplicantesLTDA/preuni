"""T076: POST /auth/register."""

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
async def test_register_happy_path_returns_201() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        r = await c.post(
            "/auth/register",
            json={
                "email": f"u-{uuid.uuid4()}@example.com",
                "password": "correctHorseBatteryStaple12!",
            },
        )
    assert r.status_code == 201, r.text
    body = r.json()
    assert "user_id" in body and "email" in body
    assert body["tier"] == "free"


@pytest.mark.asyncio
async def test_register_email_taken_409() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    email = f"u-{uuid.uuid4()}@example.com"
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        await c.post(
            "/auth/register", json={"email": email, "password": "correctHorseBatteryStaple12!"}
        )
        r = await c.post("/auth/register", json={"email": email, "password": "AnotherPassword12!"})
    assert r.status_code == 409
    assert r.json()["error_code"] == "email_already_registered"


@pytest.mark.asyncio
async def test_register_password_too_short_400() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        r = await c.post(
            "/auth/register", json={"email": f"u-{uuid.uuid4()}@example.com", "password": "short"}
        )
    assert r.status_code == 400
    assert r.json()["error_code"] in ("password_too_short", "bad_request")


@pytest.mark.asyncio
async def test_register_invalid_email_422() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        r = await c.post(
            "/auth/register",
            json={"email": "not-an-email", "password": "correctHorseBatteryStaple12!"},
        )
    assert r.status_code == 422
