# Adding a New LLM Provider

This guide explains how to add a new `LLMProvider` adapter and gate it through the Constitution Article II + Article IX checks.

## Overview

The correction pipeline depends only on the `LLMProvider` Protocol (`src/corrector/llm/base.py`). Vendors live in `src/corrector/llm/<adapter>.py`. No vendor name or SDK import may appear in `src/corrector/` business logic — import-linter enforces this at CI.

## Step 1: Implement the adapter

Create `src/corrector/llm/<vendor>.py`. Implement two things:

```python
from src.corrector.llm.base import LLMProvider, LLMResult
from src.corrector.llm.errors import RateLimitError, TimeoutError, TransientError, SchemaViolationError, FatalError

class AnthropicProvider:
    name = "anthropic"
    model_id: str  # set from config

    async def complete_structured(
        self,
        *,
        system: str,
        user: str,
        output_schema: dict,
        temperature: float,
        seed: int | None,
        max_tokens: int,
        timeout_s: float,
    ) -> LLMResult:
        ...
```

**Error mapping** (Constitution IV):

| HTTP / SDK signal | Raise |
|---|---|
| 429 | `RateLimitError` |
| Timeout / read timeout | `TimeoutError` |
| 5xx / connection error | `TransientError` |
| JSON Schema validation fail | `SchemaViolationError(validator_message=...)` |
| Auth error, unsupported model | `FatalError` |

**Constitution III requirements:**
- Pass `temperature` through (caller passes ≤ 0.2).
- Accept `seed` parameter. If the provider doesn't support seed, log a warning and proceed — do not silently ignore it.
- Persist `model_id`, `inference_params` in `LLMResult.inference_params`.

## Step 2: Write a contract test

Create `tests/integration/llm/test_<vendor>_provider_contract.py` modelled on `test_ollama_provider_contract.py`. Use a mock httpx transport — no live call.

Test checklist:
- [ ] Request shape includes temperature and seed
- [ ] 429 → `RateLimitError`
- [ ] 5xx → `TransientError`
- [ ] Timeout → `TimeoutError`
- [ ] Schema-invalid JSON → `SchemaViolationError`
- [ ] `LLMResult` fields populated

## Step 3: Wire into config

In `src/config/settings.py`, add the new provider string to `LLMSettings.provider` Literal.

In `src/workers/entrypoint.py`, extend `_build_provider()` to instantiate the new adapter.

## Step 4: Article II property checklist

Before merging, confirm:
- [ ] Provider name (`"anthropic"`, `"openai"`, …) appears **only** in `corrector/llm/<adapter>.py` and env config — never in business logic.
- [ ] `correct_essay(inp, provider=provider)` call-site in `correction_worker.py` is unchanged.
- [ ] Import-linter still passes: `uv run lint-imports --root src`

## Step 5: Article IX MVP-tier gate

Before merging a new default provider, run the golden harness against the full corpus:

```bash
OLLAMA_MODEL=<new_model> make test-golden-real
```

The PR description must include a metric delta table:

| Metric | Baseline | New |
|---|---|---|
| mae_total | 163 | … |
| mae_per_comp | 37 | … |
| hit_rate_±120 | 50% | … |

CI blocks if any threshold regresses. See `scripts/compare_golden_metrics.py`.

## Worked example: Anthropic Claude

```python
# src/corrector/llm/anthropic.py
import anthropic
from src.corrector.llm.base import LLMProvider, LLMResult
from src.corrector.llm.errors import RateLimitError, TransientError, SchemaViolationError, FatalError
import time, json, jsonschema

class AnthropicProvider:
    name = "anthropic"

    def __init__(self, *, model: str = "claude-sonnet-4-6", api_key: str) -> None:
        self._client = anthropic.AsyncAnthropic(api_key=api_key)
        self.model_id = model

    async def complete_structured(self, *, system, user, output_schema, temperature, seed, max_tokens, timeout_s):
        t0 = time.monotonic()
        try:
            response = await self._client.messages.create(
                model=self.model_id,
                max_tokens=max_tokens,
                temperature=temperature,
                system=system,
                messages=[{"role": "user", "content": user}],
            )
        except anthropic.RateLimitError as exc:
            raise RateLimitError(str(exc)) from exc
        except anthropic.APIStatusError as exc:
            raise TransientError(str(exc)) from exc

        raw = response.content[0].text
        latency_ms = int((time.monotonic() - t0) * 1000)

        try:
            parsed = json.loads(raw)
            jsonschema.validate(parsed, output_schema)
        except (json.JSONDecodeError, jsonschema.ValidationError) as exc:
            raise SchemaViolationError(str(exc), validator_message=str(exc)) from exc

        return LLMResult(
            raw=raw,
            parsed=parsed,
            model_id=self.model_id,
            prompt_tokens=response.usage.input_tokens,
            completion_tokens=response.usage.output_tokens,
            latency_ms=latency_ms,
            cost_usd=None,
            inference_params={"temperature": temperature, "max_tokens": max_tokens},
        )
```

Key differences from `OllamaProvider`:
- Anthropic doesn't support `seed` — log a warning.
- Use `anthropic.AsyncAnthropic`, not httpx directly.
- Add `anthropic` to `pyproject.toml` dependencies.
