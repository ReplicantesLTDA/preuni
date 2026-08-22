"""T094: GET /corrections/{id} — status-aware retrieval."""

from __future__ import annotations

import os
import pytest
import uuid
from httpx import ASGITransport, AsyncClient

VALID_ESSAY = (
    "O meio ambiente brasileiro sofre com desmatamento e poluição em larga escala. "
    "A Amazônia perde milhares de hectares anualmente por ação humana predatória. "
    "O agronegócio, apesar de importante, deve adotar práticas mais sustentáveis. "
    "Políticas ambientais eficazes são urgentes para preservar nossa biodiversidade. "
    "A educação ambiental desde cedo forma cidadãos mais conscientes e responsáveis. "
    "O Brasil tem potencial enorme para liderar a transição para energia renovável. "
    "É preciso unir desenvolvimento econômico com preservação ambiental de forma eficaz. "
) * 3

VALID_THEME = {
    "title": "Crise Ambiental no Brasil",
    "context": "Desafios para a preservação do meio ambiente e desenvolvimento sustentável.",
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
async def test_get_pending_correction_returns_pending_status() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        headers = {"Authorization": f"Bearer {token}"}

        submit_r = await c.post(
            "/corrections",
            json={"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME},
            headers=headers,
        )
        assert submit_r.status_code == 202
        correction_id = submit_r.json()["correction_id"]

        get_r = await c.get(f"/corrections/{correction_id}", headers=headers)

    assert get_r.status_code == 200, get_r.text
    body = get_r.json()
    assert body["correction_id"] == correction_id
    assert body["status"] == "pending"
    assert "issued_at" in body


@pytest.mark.asyncio
async def test_get_correction_repeated_get_stable_payload() -> None:
    """Same correction returns byte-identical payload across repeated GETs."""
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        headers = {"Authorization": f"Bearer {token}"}

        submit_r = await c.post(
            "/corrections",
            json={"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME},
            headers=headers,
        )
        correction_id = submit_r.json()["correction_id"]

        get1 = await c.get(f"/corrections/{correction_id}", headers=headers)
        get2 = await c.get(f"/corrections/{correction_id}", headers=headers)

    assert get1.json() == get2.json()


@pytest.mark.asyncio
async def test_get_nonexistent_correction_returns_404() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        headers = {"Authorization": f"Bearer {token}"}
        fake_id = uuid.uuid4()

        r = await c.get(f"/corrections/{fake_id}", headers=headers)

    assert r.status_code == 404, r.text
    assert r.json()["error_code"] == "not_found"


@pytest.mark.asyncio
async def test_get_correction_other_user_returns_404() -> None:
    """Cross-user access is indistinguishable from not-found (spec §404 policy)."""
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token_a = await _register_verify_login(c)
        token_b = await _register_verify_login(c)

        submit_r = await c.post(
            "/corrections",
            json={"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME},
            headers={"Authorization": f"Bearer {token_a}"},
        )
        correction_id = submit_r.json()["correction_id"]

        r = await c.get(
            f"/corrections/{correction_id}",
            headers={"Authorization": f"Bearer {token_b}"},
        )

    assert r.status_code == 404, r.text
    assert r.json()["error_code"] == "not_found"


@pytest.mark.asyncio
async def test_get_correction_unauthenticated_returns_401() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        r = await c.get(f"/corrections/{uuid.uuid4()}")
    assert r.status_code == 401
