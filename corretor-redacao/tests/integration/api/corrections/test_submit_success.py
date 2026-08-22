"""T090: POST /corrections — happy path enqueue."""

from __future__ import annotations

import os
import pytest
import uuid
from httpx import ASGITransport, AsyncClient

VALID_ESSAY = (
    "A desigualdade social no Brasil constitui um problema estrutural de grande magnitude. "
    "Historicamente, a concentração de renda impossibilitou o acesso igualitário a direitos. "
    "Educação, saúde e moradia são negados sistematicamente às camadas mais vulneráveis. "
    "O mercado de trabalho informal absorve boa parte dos trabalhadores sem proteção social. "
    "Políticas redistributivas, como o Bolsa Família, atenuam mas não resolvem o problema. "
    "A reforma tributária progressiva é apontada por economistas como medida estrutural. "
    "É indispensável que o Estado amplie programas de habitação e educação de qualidade. "
) * 3

VALID_THEME = {
    "title": "Desigualdade Social no Brasil",
    "context": "A concentração de renda e seus impactos na cidadania.",
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
async def test_submit_returns_202_with_correction_id() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        r = await c.post(
            "/corrections",
            json={"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME},
            headers={"Authorization": f"Bearer {token}"},
        )
    assert r.status_code == 202, r.text
    body = r.json()
    assert "correction_id" in body
    assert body["status"] == "pending"
    uuid.UUID(body["correction_id"])


@pytest.mark.asyncio
async def test_submit_increments_quota() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        me_before = (await c.get("/me", headers={"Authorization": f"Bearer {token}"})).json()
        await c.post(
            "/corrections",
            json={"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME},
            headers={"Authorization": f"Bearer {token}"},
        )
        me_after = (await c.get("/me", headers={"Authorization": f"Bearer {token}"})).json()
    assert me_after["quota_used_current_month"] == me_before["quota_used_current_month"] + 1


@pytest.mark.asyncio
async def test_submit_unauthenticated_returns_401() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        r = await c.post(
            "/corrections",
            json={"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME},
        )
    assert r.status_code == 401
