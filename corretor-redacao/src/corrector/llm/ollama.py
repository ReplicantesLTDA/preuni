"""Ollama provider adapter (Cloud + local).

Constitution II: this is the ONLY module that mentions the vendor by name. The
correction pipeline depends on the `LLMProvider` Protocol, not this class.

Supports:
- Ollama Cloud (HTTPS, Bearer API key)
- Local Ollama (http://localhost:11434, no auth)

A single `OllamaProvider` instance is configured by base URL + optional API key.
JSON-mode is requested via `format: "json"`; the parsed JSON is validated against
the caller-provided JSON Schema. Validation failure raises SchemaViolationError;
the pipeline owns the 2-attempt corrective-retry budget (Constitution IV).
"""

from __future__ import annotations

import httpx
import json
import jsonschema
import re
import time
from decimal import Decimal
from jsonschema.exceptions import ValidationError
from typing import Any

from src.corrector.llm.base import LLMResult
from src.corrector.llm.errors import (
    FatalError,
    RateLimitError,
    SchemaViolationError,
    TimeoutError,
    TransientError,
)


class OllamaProvider:
    """LLMProvider adapter for Ollama Cloud and local Ollama."""

    name = "ollama"

    def __init__(
        self,
        *,
        base_url: str,
        api_key: str | None,
        model_id: str,
        transport: httpx.BaseTransport | None = None,
    ) -> None:
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.model_id = model_id
        self._transport = transport

    def _client(self, timeout_s: float) -> httpx.AsyncClient:
        headers: dict[str, str] = {"content-type": "application/json"}
        if self.api_key:
            headers["authorization"] = f"Bearer {self.api_key}"
        return httpx.AsyncClient(
            base_url=self.base_url,
            headers=headers,
            timeout=timeout_s,
            transport=self._transport,
        )

    async def complete_structured(
        self,
        *,
        system: str,
        user: str,
        output_schema: dict[str, Any],
        temperature: float,
        seed: int | None,
        max_tokens: int,
        timeout_s: float,
    ) -> LLMResult:
        body: dict[str, Any] = {
            "model": self.model_id,
            "messages": [
                {"role": "system", "content": system},
                {"role": "user", "content": user},
            ],
            # Schema-constrained generation: forces Ollama to emit JSON matching the schema,
            # eliminating delimiter / missing-field errors at the source.
            "format": output_schema,
            "stream": False,
            "options": {
                "temperature": temperature,
                "num_predict": max_tokens,
            },
        }
        if seed is not None:
            body["options"]["seed"] = seed

        start = time.perf_counter()
        try:
            async with self._client(timeout_s) as client:
                resp = await client.post("/api/chat", json=body)
        except (
            httpx.ReadTimeout,
            httpx.WriteTimeout,
            httpx.ConnectTimeout,
            httpx.PoolTimeout,
        ) as exc:
            raise TimeoutError(f"Ollama timed out after {timeout_s}s") from exc
        except (httpx.NetworkError, httpx.RemoteProtocolError) as exc:
            raise TransientError(f"Ollama network error: {exc!s}") from exc

        latency_ms = int((time.perf_counter() - start) * 1000)

        self._raise_for_status(resp)

        payload = resp.json()
        raw_text = payload.get("message", {}).get("content", "")
        try:
            parsed = json.loads(raw_text)
        except json.JSONDecodeError:
            # Model may have wrapped JSON in markdown fences or added prose. Try to extract.
            try:
                parsed = _extract_json_object(raw_text)
            except (json.JSONDecodeError, ValueError) as exc:
                snippet = raw_text[:200].replace("\n", " ")
                raise SchemaViolationError(
                    "Ollama returned non-JSON content",
                    validator_message=f"{exc!s} | raw[:200]={snippet!r}",
                ) from exc

        try:
            jsonschema.validate(parsed, output_schema)
        except ValidationError as exc:
            raise SchemaViolationError(
                "Ollama output failed JSON Schema validation",
                validator_message=exc.message,
            ) from exc

        total_duration_ns = int(payload.get("total_duration", 0))
        reported_ms = total_duration_ns // 1_000_000 if total_duration_ns else 0
        return LLMResult(
            parsed_output=parsed,
            raw_text=raw_text,
            prompt_tokens=int(payload.get("prompt_eval_count", 0)),
            completion_tokens=int(payload.get("eval_count", 0)),
            cost_usd=Decimal("0"),  # Cost accounting wired later; Ollama Cloud bills out-of-band
            latency_ms=max(latency_ms, reported_ms),
            model_id=self.model_id,
            inference_params={
                "temperature": temperature,
                "seed": seed,
                "max_tokens": max_tokens,
                "provider": self.name,
            },
        )

    @staticmethod
    def _raise_for_status(resp: httpx.Response) -> None:
        if resp.is_success:
            return
        code = resp.status_code
        try:
            detail = resp.json()
        except json.JSONDecodeError:
            detail = {"text": resp.text}
        if code == 429:
            raise RateLimitError(f"Ollama 429: {detail}")
        if code in (401, 403):
            raise FatalError(f"Ollama auth failure ({code}): {detail}")
        if code in (400, 404):
            raise FatalError(f"Ollama {code}: {detail}")
        if 500 <= code < 600:
            raise TransientError(f"Ollama {code}: {detail}")
        raise TransientError(f"Ollama unexpected {code}: {detail}")


_FENCE_RE = re.compile(r"```(?:json)?\s*\n?(.+?)\n?```", re.DOTALL)


def _extract_json_object(raw: str) -> dict[str, Any]:
    """Pull a JSON object out of text that may have markdown fences or prose.

    Strategy: strip ```json fences if present; otherwise locate the first balanced
    `{...}` block. Raises json.JSONDecodeError or ValueError on failure.
    """
    s = raw.strip()
    m = _FENCE_RE.search(s)
    if m:
        s = m.group(1).strip()
    start = s.find("{")
    end = s.rfind("}")
    if start < 0 or end <= start:
        raise ValueError("no JSON object found in raw text")
    return json.loads(s[start : end + 1])


__all__ = ["OllamaProvider"]
