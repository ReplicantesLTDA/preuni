"""T108: POST /corrections/{id}/reevaluations."""

from __future__ import annotations

import os
import pytest
import uuid
from httpx import ASGITransport, AsyncClient

VALID_ESSAY = (
    "A mobilidade urbana no Brasil enfrenta sérios desafios que afetam qualidade de vida. "
    "O transporte público precário força muitos cidadãos a usar veículos particulares. "
    "O aumento de automóveis nas cidades intensifica o caos no trânsito cotidianamente. "
    "Investimentos em transporte coletivo reduzem emissões e melhoram a mobilidade urbana. "
    "Cidades como Curitiba mostram que planejamento urbano eficaz é possível no Brasil. "
    "A integração entre meios de transporte é fundamental para solucionar este problema. "
    "Só com políticas públicas integradas será possível garantir mobilidade digna a todos. "
) * 3

VALID_THEME = {
    "title": "Mobilidade Urbana no Brasil",
    "context": "Desafios do transporte público e soluções para cidades mais sustentáveis.",
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
async def test_reevaluate_returns_202_with_new_id() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        headers = {"Authorization": f"Bearer {token}"}

        original_r = await c.post(
            "/corrections",
            json={"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME},
            headers=headers,
        )
        original_id = original_r.json()["correction_id"]

        reeval_r = await c.post(f"/corrections/{original_id}/reevaluations", headers=headers)

    assert reeval_r.status_code == 202, reeval_r.text
    body = reeval_r.json()
    assert "correction_id" in body
    assert body["correction_id"] != original_id
    assert body["reevaluation_of"] == original_id
    assert body["status"] == "pending"


@pytest.mark.asyncio
async def test_reevaluate_original_stays_unchanged() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        headers = {"Authorization": f"Bearer {token}"}

        original_r = await c.post(
            "/corrections",
            json={"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME},
            headers=headers,
        )
        original_id = original_r.json()["correction_id"]

        await c.post(f"/corrections/{original_id}/reevaluations", headers=headers)

        get_r = await c.get(f"/corrections/{original_id}", headers=headers)

    assert get_r.status_code == 200
    assert get_r.json()["status"] == "pending"
    assert get_r.json()["correction_id"] == original_id


@pytest.mark.asyncio
async def test_reevaluate_consumes_quota() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        headers = {"Authorization": f"Bearer {token}"}

        original_r = await c.post(
            "/corrections",
            json={"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME},
            headers=headers,
        )
        original_id = original_r.json()["correction_id"]

        me_before = (await c.get("/me", headers=headers)).json()
        await c.post(f"/corrections/{original_id}/reevaluations", headers=headers)
        me_after = (await c.get("/me", headers=headers)).json()

    assert me_after["quota_used_current_month"] == me_before["quota_used_current_month"] + 1


@pytest.mark.asyncio
async def test_reevaluate_not_owned_returns_404() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token_a = await _register_verify_login(c)
        token_b = await _register_verify_login(c)

        original_r = await c.post(
            "/corrections",
            json={"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME},
            headers={"Authorization": f"Bearer {token_a}"},
        )
        original_id = original_r.json()["correction_id"]

        r = await c.post(
            f"/corrections/{original_id}/reevaluations",
            headers={"Authorization": f"Bearer {token_b}"},
        )

    assert r.status_code == 404
    assert r.json()["error_code"] == "not_found"


@pytest.mark.asyncio
async def test_reevaluate_quota_exhausted_returns_429() -> None:
    _setup_env()
    from src.api.main import create_app

    app = create_app()
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://t") as c:
        token = await _register_verify_login(c)
        headers = {"Authorization": f"Bearer {token}"}
        payload = {"essay_text": VALID_ESSAY, "prompt_theme": VALID_THEME}

        original_r = await c.post("/corrections", json=payload, headers=headers)
        original_id = original_r.json()["correction_id"]

        # Use up remaining quota (started with 1, now 2 more = 3 total)
        for _ in range(2):
            await c.post("/corrections", json=payload, headers=headers)

        r = await c.post(f"/corrections/{original_id}/reevaluations", headers=headers)

    assert r.status_code == 429, r.text
    assert r.json()["error_code"] == "quota_exhausted"
