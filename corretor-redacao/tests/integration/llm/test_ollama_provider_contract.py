"""T024: contract for OllamaProvider against a mocked httpx transport.

NO live LLM call. Verifies:
- request shape (model, options.temperature, options.seed, format=json)
- response parsing into LLMResult
- error-taxonomy mapping (429 → RateLimitError, 5xx → TransientError, timeout → TimeoutError,
  401 → FatalError)
- schema validation post-parse (invalid JSON / invalid against schema → SchemaViolationError)
"""

from __future__ import annotations

import httpx
import json
import pytest
from typing import Any

from src.corrector.llm import errors
from src.corrector.llm.ollama import OllamaProvider


def _valid_correction() -> dict[str, Any]:
    return {
        "eliminatory_flags": [],
        "competencies": {
            "c1": {
                "score": 200,
                "excerpt": "exemplo de trecho citado",
                "justification_pt_br": "ok",
            },
            "c2": {
                "score": 200,
                "excerpt": "exemplo de trecho citado",
                "justification_pt_br": "ok",
            },
            "c3": {
                "score": 200,
                "excerpt": "exemplo de trecho citado",
                "justification_pt_br": "ok",
            },
            "c4": {
                "score": 200,
                "excerpt": "exemplo de trecho citado",
                "justification_pt_br": "ok",
            },
            "c5": {
                "score": 200,
                "excerpt": "exemplo de trecho citado",
                "justification_pt_br": "ok",
            },
        },
        "final_score": 1000,
    }


def _ollama_chat_response(body: dict[str, Any]) -> dict[str, Any]:
    return {
        "model": "kimi-k2:1t",
        "created_at": "2026-05-28T20:00:00Z",
        "message": {"role": "assistant", "content": json.dumps(body, ensure_ascii=False)},
        "done": True,
        "prompt_eval_count": 100,
        "eval_count": 200,
        "total_duration": 50_000_000,
    }


SCHEMA: dict[str, Any] = __import__(
    "src.schemas", fromlist=["active_correction_output_schema"]
).active_correction_output_schema()


@pytest.mark.asyncio
async def test_request_shape() -> None:
    captured: dict[str, Any] = {}

    def handler(req: httpx.Request) -> httpx.Response:
        captured["url"] = str(req.url)
        captured["headers"] = dict(req.headers)
        captured["body"] = json.loads(req.content)
        return httpx.Response(200, json=_ollama_chat_response(_valid_correction()))

    transport = httpx.MockTransport(handler)
    provider = OllamaProvider(
        base_url="https://ollama.cloud",
        api_key="sk-test",
        model_id="kimi-k2:1t",
        transport=transport,
    )
    await provider.complete_structured(
        system="sys",
        user="usr",
        output_schema=SCHEMA,
        temperature=0.1,
        seed=42,
        max_tokens=2048,
        timeout_s=10,
    )
    assert captured["url"].endswith("/api/chat")
    assert captured["headers"].get("authorization") == "Bearer sk-test"
    assert captured["body"]["model"] == "kimi-k2:1t"
    # Constrained-generation mode: format carries the JSON schema, not literal "json".
    assert isinstance(captured["body"]["format"], dict)
    assert captured["body"]["format"].get("type") == "object"
    assert captured["body"]["stream"] is False
    assert captured["body"]["options"]["temperature"] == pytest.approx(0.1)
    assert captured["body"]["options"]["seed"] == 42
    assert captured["body"]["options"]["num_predict"] == 2048
    # messages: [system, user]
    msgs = captured["body"]["messages"]
    assert msgs[0]["role"] == "system" and msgs[0]["content"] == "sys"
    assert msgs[1]["role"] == "user" and msgs[1]["content"] == "usr"


@pytest.mark.asyncio
async def test_local_ollama_skips_authorization_header() -> None:
    captured: dict[str, Any] = {}

    def handler(req: httpx.Request) -> httpx.Response:
        captured["headers"] = dict(req.headers)
        return httpx.Response(200, json=_ollama_chat_response(_valid_correction()))

    provider = OllamaProvider(
        base_url="http://localhost:11434",
        api_key=None,
        model_id="qwen3:14b",
        transport=httpx.MockTransport(handler),
    )
    await provider.complete_structured(
        system="s",
        user="u",
        output_schema=SCHEMA,
        temperature=0.0,
        seed=None,
        max_tokens=512,
        timeout_s=10,
    )
    assert "authorization" not in captured["headers"]


@pytest.mark.asyncio
async def test_result_fields_populated() -> None:
    transport = httpx.MockTransport(
        lambda _: httpx.Response(200, json=_ollama_chat_response(_valid_correction()))
    )
    provider = OllamaProvider(
        base_url="https://ollama.cloud",
        api_key="sk",
        model_id="kimi-k2:1t",
        transport=transport,
    )
    result = await provider.complete_structured(
        system="s",
        user="u",
        output_schema=SCHEMA,
        temperature=0.0,
        seed=1,
        max_tokens=1024,
        timeout_s=10,
    )
    assert result.parsed_output["final_score"] == 1000
    assert result.prompt_tokens == 100
    assert result.completion_tokens == 200
    assert result.latency_ms > 0
    assert result.model_id == "kimi-k2:1t"
    assert result.inference_params["seed"] == 1


@pytest.mark.asyncio
async def test_429_maps_to_rate_limit() -> None:
    transport = httpx.MockTransport(lambda _: httpx.Response(429, json={"error": "rate limit"}))
    provider = OllamaProvider(
        base_url="https://ollama.cloud",
        api_key="sk",
        model_id="m",
        transport=transport,
    )
    with pytest.raises(errors.RateLimitError):
        await provider.complete_structured(
            system="",
            user="",
            output_schema=SCHEMA,
            temperature=0.0,
            seed=None,
            max_tokens=10,
            timeout_s=5,
        )


@pytest.mark.asyncio
async def test_5xx_maps_to_transient() -> None:
    transport = httpx.MockTransport(lambda _: httpx.Response(503, json={"error": "down"}))
    provider = OllamaProvider(
        base_url="https://ollama.cloud",
        api_key="sk",
        model_id="m",
        transport=transport,
    )
    with pytest.raises(errors.TransientError):
        await provider.complete_structured(
            system="",
            user="",
            output_schema=SCHEMA,
            temperature=0.0,
            seed=None,
            max_tokens=10,
            timeout_s=5,
        )


@pytest.mark.asyncio
async def test_401_maps_to_fatal() -> None:
    transport = httpx.MockTransport(lambda _: httpx.Response(401, json={"error": "auth"}))
    provider = OllamaProvider(
        base_url="https://ollama.cloud",
        api_key="bad",
        model_id="m",
        transport=transport,
    )
    with pytest.raises(errors.FatalError):
        await provider.complete_structured(
            system="",
            user="",
            output_schema=SCHEMA,
            temperature=0.0,
            seed=None,
            max_tokens=10,
            timeout_s=5,
        )


@pytest.mark.asyncio
async def test_timeout_maps_to_timeout_error() -> None:
    def handler(req: httpx.Request) -> httpx.Response:
        raise httpx.ReadTimeout("boom", request=req)

    provider = OllamaProvider(
        base_url="https://ollama.cloud",
        api_key="sk",
        model_id="m",
        transport=httpx.MockTransport(handler),
    )
    with pytest.raises(errors.TimeoutError):
        await provider.complete_structured(
            system="",
            user="",
            output_schema=SCHEMA,
            temperature=0.0,
            seed=None,
            max_tokens=10,
            timeout_s=5,
        )


@pytest.mark.asyncio
async def test_malformed_json_raises_schema_violation() -> None:
    bad = {
        "model": "m",
        "created_at": "x",
        "message": {"role": "assistant", "content": "not json{"},
        "done": True,
        "prompt_eval_count": 1,
        "eval_count": 1,
        "total_duration": 1,
    }
    provider = OllamaProvider(
        base_url="https://ollama.cloud",
        api_key="sk",
        model_id="m",
        transport=httpx.MockTransport(lambda _: httpx.Response(200, json=bad)),
    )
    with pytest.raises(errors.SchemaViolationError):
        await provider.complete_structured(
            system="",
            user="",
            output_schema=SCHEMA,
            temperature=0.0,
            seed=None,
            max_tokens=10,
            timeout_s=5,
        )


@pytest.mark.asyncio
async def test_schema_invalid_output_raises_schema_violation() -> None:
    invalid = {"competencies": {}, "final_score": 5000, "eliminatory_flags": []}
    provider = OllamaProvider(
        base_url="https://ollama.cloud",
        api_key="sk",
        model_id="m",
        transport=httpx.MockTransport(
            lambda _: httpx.Response(200, json=_ollama_chat_response(invalid))
        ),
    )
    with pytest.raises(errors.SchemaViolationError) as exc:
        await provider.complete_structured(
            system="",
            user="",
            output_schema=SCHEMA,
            temperature=0.0,
            seed=None,
            max_tokens=10,
            timeout_s=5,
        )
    assert exc.value.validator_message is not None
