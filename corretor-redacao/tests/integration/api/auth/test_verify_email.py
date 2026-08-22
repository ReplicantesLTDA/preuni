"""T077: POST /auth/verify-email."""

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


async def _register_and_extract_token(client: AsyncClient) -> tuple[str, str]:
    """Register, then read the verification token from the DB (dev fallback path
    does not actually email)."""
    email = f"u-{uuid.uuid4()}@example.com"
    r = await client.post(
        "/auth/register", json={"email": email, "password": "correctHorseBatteryStaple12!"}
    )
    assert r.status_code == 201
    user_id = r.json()["user_id"]

    # Pull the most-recent active verification token for that user.
    from sqlalchemy import text

    from src.db.engine import get_engine

    engine = get_engine()
    async with engine.connect() as conn:
        row = (
            await conn.execute(
                text(
                    "SELECT token_hash FROM email_verification_tokens "
                    "WHERE user_id = :uid AND used_at IS NULL ORDER BY created_at DESC LIMIT 1"
                ),
                {"uid": user_id},
            )
        ).first()
    assert row is not None
    # We only have the hash; for testing, register sets the cleartext on the response.
    return email, user_id


@pytest.mark.asyncio
async def test_verify_email_happy_path_returns_204() -> None:
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
        assert r.status_code == 201
        token = r.json().get("verification_token_dev")
        assert token, "register response must include verification_token_dev in dev mode"

        r = await c.post("/auth/verify-email", json={"token": token})
    assert r.status_code == 204


@pytest.mark.asyncio
async def test_verify_email_invalid_token_400() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        r = await c.post("/auth/verify-email", json={"token": "totally-bogus-token"})
    assert r.status_code == 400
    assert r.json()["error_code"] == "invalid_token"
