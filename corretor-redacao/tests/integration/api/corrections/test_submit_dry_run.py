"""T089: POST /corrections?dry_run=true."""

from __future__ import annotations

import os
import pytest
import uuid
from httpx import ASGITransport, AsyncClient

VALID_ESSAY = (
    "A educação brasileira enfrenta desafios históricos profundos que exigem reflexão crítica. "
    "Primeiramente, a desigualdade no acesso ao ensino de qualidade perpetua ciclos de pobreza. "
    "Em segundo lugar, a formação docente ainda carece de investimentos adequados em todo o país. "
    "Além disso, a infraestrutura das escolas públicas é frequentemente precária e inadequada. "
    "Nesse sentido, políticas públicas eficazes são fundamentais para transformar essa realidade. "
    "Portanto, é urgente que o Estado, a sociedade civil e as famílias se mobilizem juntos. "
    "Somente com articulação coletiva será possível garantir educação de qualidade para todos. "
) * 3

VALID_THEME = {"title": "Educação no Brasil", "context": "Desafios do ensino público brasileiro."}


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
async def test_dry_run_valid_input_returns_200_ok_to_submit() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        r = await c.post(
            "/corrections?dry_run=true",
            json={"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME},
            headers={"Authorization": f"Bearer {token}"},
        )
    assert r.status_code == 200, r.text
    body = r.json()
    assert body["ok_to_submit"] is True


@pytest.mark.asyncio
async def test_dry_run_does_not_insert_row() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        r = await c.post(
            "/corrections?dry_run=true",
            json={"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME},
            headers={"Authorization": f"Bearer {token}"},
        )
        assert r.status_code == 200

        me_r = await c.get("/me", headers={"Authorization": f"Bearer {token}"})
        assert me_r.json()["quota_used_current_month"] == 0


@pytest.mark.asyncio
async def test_dry_run_essay_too_short_returns_400() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        short_essay = "Curta " * 5
        r = await c.post(
            "/corrections?dry_run=true",
            json={"essay_text": short_essay, "prompt_theme": VALID_THEME},
            headers={"Authorization": f"Bearer {token}"},
        )
    assert r.status_code == 400, r.text
    assert r.json()["error_code"] == "length_too_short"


@pytest.mark.asyncio
async def test_dry_run_theme_missing_context_returns_400() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        bad_theme = {"title": "Título sem contexto", "context": ""}
        r = await c.post(
            "/corrections?dry_run=true",
            json={"essay_text": VALID_ESSAY, "prompt_theme": bad_theme},
            headers={"Authorization": f"Bearer {token}"},
        )
    assert r.status_code == 400, r.text
    assert r.json()["error_code"] in ("theme_missing_context", "theme_missing_title")


@pytest.mark.asyncio
async def test_dry_run_unauthenticated_returns_401() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        r = await c.post(
            "/corrections?dry_run=true",
            json={"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME},
        )
    assert r.status_code == 401
