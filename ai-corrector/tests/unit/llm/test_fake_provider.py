"""T022: FakeProvider behavior contract."""

from __future__ import annotations

import pytest
from decimal import Decimal

from src.corrector.llm import errors
from src.corrector.llm.fake import FakeProvider


@pytest.mark.asyncio
async def test_returns_canned_output() -> None:
    canned = {"final_score": 600}
    fake = FakeProvider(responses=[canned])
    result = await fake.complete_structured(
        system="sys",
        user="usr",
        output_schema={},
        temperature=0.1,
        seed=42,
        max_tokens=128,
        timeout_s=30,
    )
    assert result.parsed_output == canned
    assert result.model_id == fake.model_id
    assert result.inference_params["seed"] == 42
    assert result.inference_params["temperature"] == pytest.approx(0.1)
    assert result.cost_usd == Decimal("0")


@pytest.mark.asyncio
async def test_records_calls() -> None:
    fake = FakeProvider(responses=[{"ok": True}, {"ok": True}])
    await fake.complete_structured(
        system="s1",
        user="u1",
        output_schema={},
        temperature=0.0,
        seed=1,
        max_tokens=10,
        timeout_s=5,
    )
    await fake.complete_structured(
        system="s2",
        user="u2",
        output_schema={},
        temperature=0.0,
        seed=2,
        max_tokens=10,
        timeout_s=5,
    )
    assert len(fake.calls) == 2
    assert fake.calls[0]["seed"] == 1
    assert fake.calls[1]["user"] == "u2"


@pytest.mark.asyncio
async def test_raises_canned_exceptions() -> None:
    fake = FakeProvider(responses=[errors.RateLimitError("rl")])
    with pytest.raises(errors.RateLimitError):
        await fake.complete_structured(
            system="",
            user="",
            output_schema={},
            temperature=0.0,
            seed=None,
            max_tokens=10,
            timeout_s=5,
        )


@pytest.mark.asyncio
async def test_exhaustion_raises_stopiteration_translated_to_fatal() -> None:
    fake = FakeProvider(responses=[])
    with pytest.raises(errors.FatalError):
        await fake.complete_structured(
            system="",
            user="",
            output_schema={},
            temperature=0.0,
            seed=None,
            max_tokens=10,
            timeout_s=5,
        )


@pytest.mark.asyncio
async def test_default_model_id_and_name() -> None:
    fake = FakeProvider(responses=[{"x": 1}])
    assert fake.name == "fake"
    assert fake.model_id == "fake-deterministic"
