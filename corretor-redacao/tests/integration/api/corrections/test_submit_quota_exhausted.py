"""T091: POST /corrections — quota exhausted path."""

from __future__ import annotations

import os
import pytest
import uuid
from httpx import ASGITransport, AsyncClient

VALID_ESSAY = (
    "A violência urbana no Brasil é reflexo de desigualdades históricas profundas. "
    "A ausência do Estado em periferias cria um vácuo preenchido pela criminalidade. "
    "Programas sociais integrados são fundamentais para reverter este ciclo vicioso. "
    "A educação de qualidade representa a principal ferramenta de transformação social. "
    "Investir em segurança pública sem tratar as causas é insuficiente e ineficaz. "
    "O diálogo entre comunidades e autoridades fortalece a cidadania e a democracia. "
    "Somente com políticas integradas o Brasil poderá superar este grave problema social. "
) * 3

VALID_THEME = {
    "title": "Violência Urbana no Brasil",
    "context": "Causas e consequências da violência nas cidades brasileiras.",
}


def _setup_env() -> None:
    os.environ.setdefault("JWT_SECRET_KEY", "test-secret-key-with-enough-bytes-for-hs256")
    os.environ["DATABASE_URL"] = os.environ.get(
        "TEST_DATABASE_URL",
        "postgresql+asyncpg://corretor:corretor@localhost:5432/corretor_test",
    )
    from src.db.engine import _engine_singleton

    _engine_singleton.cache_clear()


async def _register_verify_login(client: AsyncClient) -> str:
    email = f"u-{uuid.uuid4()}@example.com"
    reg = await client.post(
        "/auth/register", json={"email": email, "password": "correctHorseBatteryStaple12!"}
    )
    verify_token = reg.json().get("verification_token_dev")
    if verify_token:
        await client.post("/auth/verify-email", json={"token": verify_token})
    token_r = await client.post(
        "/auth/login", json={"email": email, "password": "correctHorseBatteryStaple12!"}
    )
    return token_r.json()["access_token"]


@pytest.mark.asyncio
async def test_quota_exhausted_returns_429() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        payload = {"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME}
        headers = {"Authorization": f"Bearer {token}"}

        for _ in range(3):
            r = await c.post("/corrections", json=payload, headers=headers)
            assert r.status_code == 202, r.text

        r = await c.post("/corrections", json=payload, headers=headers)

    assert r.status_code == 429, r.text
    body = r.json()
    assert body["error_code"] == "quota_exhausted"
    assert "quota_reset_at" in body


@pytest.mark.asyncio
async def test_quota_exhausted_response_includes_reset_at() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        payload = {"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME}
        headers = {"Authorization": f"Bearer {token}"}

        for _ in range(3):
            await c.post("/corrections", json=payload, headers=headers)

        r = await c.post("/corrections", json=payload, headers=headers)

    body = r.json()
    import datetime as dt

    reset_at = dt.datetime.fromisoformat(body["quota_reset_at"].replace("Z", "+00:00"))
    assert reset_at.day == 1
    assert reset_at.hour == 0


@pytest.mark.asyncio
async def test_quota_exhausted_no_row_inserted() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        payload = {"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME}
        headers = {"Authorization": f"Bearer {token}"}

        for _ in range(3):
            await c.post("/corrections", json=payload, headers=headers)

        me_before = (await c.get("/me", headers=headers)).json()
        r = await c.post("/corrections", json=payload, headers=headers)
        assert r.status_code == 429
        me_after = (await c.get("/me", headers=headers)).json()

    assert me_after["quota_used_current_month"] == me_before["quota_used_current_month"]
