"""T092: POST /corrections — pre-validation rejection paths."""

from __future__ import annotations

import os
import pytest
import uuid
from httpx import ASGITransport, AsyncClient

VALID_ESSAY = (
    "A saúde pública no Brasil enfrenta crises profundas e estruturais que exigem atenção. "
    "O subfinanciamento do SUS compromete o atendimento de milhões de brasileiros carentes. "
    "Longas filas, falta de médicos e equipamentos obsoletos são realidade cotidiana. "
    "A pandemia evidenciou tanto as fragilidades quanto a capacidade de resposta do sistema. "
    "A medicina preventiva ainda é pouco valorizada frente à medicina curativa no Brasil. "
    "Investimentos em atenção básica reduzem custos hospitalares e melhoram qualidade de vida. "
    "Somente com vontade política e recursos adequados o SUS poderá cumprir seu papel. "
) * 3

VALID_THEME = {
    "title": "Saúde Pública no Brasil",
    "context": "Desafios e oportunidades do Sistema Único de Saúde.",
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


async def _quota_unchanged(client: AsyncClient, token: str, before: int) -> bool:
    me = (await client.get("/me", headers={"Authorization": f"Bearer {token}"})).json()
    return me["quota_used_current_month"] == before


@pytest.mark.asyncio
async def test_essay_too_short_returns_400() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        r = await c.post(
            "/corrections",
            json={"essay_text": "Curto demais.", "prompt_theme": VALID_THEME},
            headers={"Authorization": f"Bearer {token}"},
        )
    assert r.status_code == 400, r.text
    assert r.json()["error_code"] == "length_too_short"


@pytest.mark.asyncio
async def test_essay_too_short_no_quota_consumed() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        me_before = (await c.get("/me", headers={"Authorization": f"Bearer {token}"})).json()
        await c.post(
            "/corrections",
            json={"essay_text": "Curto demais.", "prompt_theme": VALID_THEME},
            headers={"Authorization": f"Bearer {token}"},
        )
        assert await _quota_unchanged(c, token, me_before["quota_used_current_month"])


@pytest.mark.asyncio
async def test_essay_too_long_returns_400() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        long_essay = "\n".join(["Texto muito longo. " * 20] * 55)  # >3500 chars AND >50 lines
        r = await c.post(
            "/corrections",
            json={"essay_text": long_essay, "prompt_theme": VALID_THEME},
            headers={"Authorization": f"Bearer {token}"},
        )
    assert r.status_code == 400, r.text
    assert r.json()["error_code"] == "length_too_long"


@pytest.mark.asyncio
async def test_theme_missing_context_returns_400() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        bad_theme = {"title": "Título", "context": ""}
        r = await c.post(
            "/corrections",
            json={"essay_text": VALID_ESSAY, "prompt_theme": bad_theme},
            headers={"Authorization": f"Bearer {token}"},
        )
    assert r.status_code == 400, r.text
    assert r.json()["error_code"] in ("theme_missing_context", "theme_missing_title")


@pytest.mark.asyncio
async def test_theme_missing_context_no_quota_consumed() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        me_before = (await c.get("/me", headers={"Authorization": f"Bearer {token}"})).json()
        bad_theme = {"title": "Título", "context": ""}
        await c.post(
            "/corrections",
            json={"essay_text": VALID_ESSAY, "prompt_theme": bad_theme},
            headers={"Authorization": f"Bearer {token}"},
        )
        assert await _quota_unchanged(c, token, me_before["quota_used_current_month"])


@pytest.mark.asyncio
async def test_missing_essay_text_returns_422() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        r = await c.post(
            "/corrections",
            json={"prompt_theme": VALID_THEME},
            headers={"Authorization": f"Bearer {token}"},
        )
    assert r.status_code == 422
