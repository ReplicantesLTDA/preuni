"""T106: GET /corrections — paginated correction list."""

from __future__ import annotations

import os
import pytest
import uuid
from httpx import ASGITransport, AsyncClient

VALID_ESSAY = (
    "A falta de saneamento básico no Brasil afeta milhões de cidadãos vulneráveis. "
    "Sem água tratada e esgoto adequado, doenças se proliferam em regiões carentes. "
    "O investimento público em infraestrutura sanitária é direito social básico. "
    "Países desenvolvidos alcançaram saúde pública avançada via saneamento universal. "
    "O Brasil precisa ampliar acesso ao saneamento especialmente em áreas periféricas. "
    "Parcerias entre governo e iniciativa privada podem acelerar esta transformação. "
    "Só com saneamento básico universal o Brasil garantirá dignidade a todos os cidadãos. "
) * 3

VALID_THEME = {
    "title": "Saneamento Básico no Brasil",
    "context": "Desafios e soluções para universalizar o acesso ao saneamento básico.",
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
async def test_list_empty_returns_empty_items() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        r = await c.get("/corrections", headers={"Authorization": f"Bearer {token}"})
    assert r.status_code == 200, r.text
    body = r.json()
    assert "items" in body
    assert isinstance(body["items"], list)


@pytest.mark.asyncio
async def test_list_shows_submitted_corrections() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        headers = {"Authorization": f"Bearer {token}"}
        payload = {"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME}

        await c.post("/corrections", json=payload, headers=headers)
        await c.post("/corrections", json=payload, headers=headers)

        r = await c.get("/corrections", headers=headers)

    assert r.status_code == 200, r.text
    body = r.json()
    assert len(body["items"]) == 2
    for item in body["items"]:
        assert "correction_id" in item
        assert "status" in item
        assert "issued_at" in item
        assert "prompt_theme_title" in item


@pytest.mark.asyncio
async def test_list_reverse_chronological_order() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        headers = {"Authorization": f"Bearer {token}"}
        payload = {"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME}

        r1 = await c.post("/corrections", json=payload, headers=headers)
        r2 = await c.post("/corrections", json=payload, headers=headers)
        id_first = r1.json()["correction_id"]
        id_second = r2.json()["correction_id"]

        list_r = await c.get("/corrections", headers=headers)

    items = list_r.json()["items"]
    assert len(items) == 2
    # Most recent first
    assert items[0]["correction_id"] == id_second
    assert items[1]["correction_id"] == id_first


@pytest.mark.asyncio
async def test_list_cross_user_impossible() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token_a = await _register_verify_login(c)
        token_b = await _register_verify_login(c)

        payload = {"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME}
        await c.post("/corrections", json=payload, headers={"Authorization": f"Bearer {token_a}"})

        r = await c.get("/corrections", headers={"Authorization": f"Bearer {token_b}"})

    assert r.status_code == 200
    assert r.json()["items"] == []


@pytest.mark.asyncio
async def test_list_unauthenticated_returns_401() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        r = await c.get("/corrections")
    assert r.status_code == 401


@pytest.mark.asyncio
async def test_list_limit_parameter() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        headers = {"Authorization": f"Bearer {token}"}
        payload = {"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME}

        await c.post("/corrections", json=payload, headers=headers)
        await c.post("/corrections", json=payload, headers=headers)
        await c.post("/corrections", json=payload, headers=headers)

        r = await c.get("/corrections?limit=2", headers=headers)

    assert r.status_code == 200, r.text
    body = r.json()
    assert len(body["items"]) == 2
    assert "next_cursor" in body
