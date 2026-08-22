"""LLM provider Protocol + result dataclass.

Constitution Article II: every adapter (Ollama, Anthropic, OpenAI, …) implements
this Protocol. The correction pipeline depends on the Protocol; never on a
concrete adapter.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from decimal import Decimal
from typing import Any, Protocol, runtime_checkable


@dataclass(slots=True)
class LLMResult:
    """Outcome of one `complete_structured` call."""

    parsed_output: dict[str, Any]
    raw_text: str
    prompt_tokens: int
    completion_tokens: int
    cost_usd: Decimal
    latency_ms: int
    model_id: str
    inference_params: dict[str, Any] = field(default_factory=dict)


@runtime_checkable
class LLMProvider(Protocol):
    """Provider-agnostic structured-output surface.

    Implementations live in `src/corrector/llm/` (one file per vendor).
    Vendor names MUST NOT appear outside this package (Constitution Article II).
    """

    name: str
    model_id: str

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
        """Run a single structured-output completion. Raises an `LLMError` subclass on failure."""
        ...


__all__ = ["LLMProvider", "LLMResult"]
