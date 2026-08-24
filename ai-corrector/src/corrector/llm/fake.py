"""Deterministic in-memory LLMProvider for tests.

Used by `tests/unit/` and by the golden harness's fake-provider job to exercise
the full pipeline without burning Ollama credits.
"""

from __future__ import annotations

import json
from collections.abc import Iterable
from decimal import Decimal
from typing import Any

from src.corrector.llm.base import LLMProvider, LLMResult
from src.corrector.llm.errors import FatalError, LLMError


class FakeProvider(LLMProvider):
    """Returns pre-seeded responses or raises pre-seeded exceptions in order."""

    name = "fake"
    model_id = "fake-deterministic"

    def __init__(
        self,
        *,
        responses: Iterable[dict[str, Any] | LLMError],
        latency_ms: int = 1,
        cost_per_call_usd: Decimal | str = "0",
    ) -> None:
        self._responses = list(responses)
        self._index = 0
        self._latency_ms = latency_ms
        self._cost = Decimal(cost_per_call_usd)
        self.calls: list[dict[str, Any]] = []

    async def complete_structured(
        self,
        *,
        system: str,
        user: str,
        output_schema: dict[str, Any],  # noqa: ARG002 — Protocol signature; FakeProvider does not validate
        temperature: float,
        seed: int | None,
        max_tokens: int,
        timeout_s: float,
    ) -> LLMResult:
        self.calls.append(
            {
                "system": system,
                "user": user,
                "temperature": temperature,
                "seed": seed,
                "max_tokens": max_tokens,
                "timeout_s": timeout_s,
            }
        )
        if self._index >= len(self._responses):
            raise FatalError(
                f"FakeProvider exhausted after {self._index} call(s); seed more responses."
            )
        item = self._responses[self._index]
        self._index += 1
        if isinstance(item, LLMError):
            raise item
        return LLMResult(
            parsed_output=item,
            raw_text=json.dumps(item, ensure_ascii=False),
            prompt_tokens=len(system) + len(user),
            completion_tokens=len(json.dumps(item)),
            cost_usd=self._cost,
            latency_ms=self._latency_ms,
            model_id=self.model_id,
            inference_params={
                "temperature": temperature,
                "seed": seed,
                "max_tokens": max_tokens,
            },
        )


__all__ = ["FakeProvider"]
