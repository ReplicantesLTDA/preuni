"""T046: golden-dataset pytest entrypoint.

Iterates `tests/golden/essay_br/test/` (gated on existence) and
`tests/golden/inep_exemplary/`. Skips gracefully when the ingest script has
not been run yet (Phase 2I tasks T049–T051).

Provider selection:
- Default (CI on every PR + local): `FakeLLMProvider` is too uninformative on
  real-world essays, so the default for this pytest target is the real
  provider iff `OLLAMA_API_KEY` is set, otherwise the suite is skipped.
- Override with `--golden-provider=fake` to run against a deterministic fake
  that always returns the human-reference scores (plumbing sanity only).
"""

from __future__ import annotations

import json
import os
import pytest
from pathlib import Path

from src.corrector.llm.fake import FakeProvider
from tests.golden.harness import Example, run_examples

GOLDEN_ROOT = Path(__file__).parent
ESSAY_BR_TEST = GOLDEN_ROOT / "essay_br" / "test"
INEP_DIR = GOLDEN_ROOT / "inep_exemplary"


def _provider_choice() -> str:
    return os.environ.get("GOLDEN_PROVIDER", "auto").lower()


def _make_real_provider() -> object:
    from src.corrector.llm.ollama import OllamaProvider

    api_key = os.environ.get("OLLAMA_CLOUD_API_KEY") or os.environ.get("OLLAMA_API_KEY")
    base_url = (
        os.environ.get("OLLAMA_CLOUD_BASE_URL")
        or os.environ.get("OLLAMA_BASE_URL")
        or "https://ollama.com"
    )
    model_id = os.environ.get("LLM_MODEL_ID") or os.environ.get("OLLAMA_MODEL") or "kimi-k2:1t"
    if not api_key:
        pytest.skip("OLLAMA_*_API_KEY not set; skipping real-provider golden run")
    return OllamaProvider(base_url=base_url, api_key=api_key, model_id=model_id)


def _make_fake_provider(examples: list[Example]) -> FakeProvider:
    responses: list[dict[str, object]] = []
    for ex in examples:
        comps: dict[str, object] = {}
        for code in ("c1", "c2", "c3", "c4", "c5"):
            score = getattr(ex, f"expected_{code}")
            entry: dict[str, object] = {
                "score": score,
                "excerpt": ex.essay_text[:60],
                "justification_pt_br": "fake reference match.",
            }
            if score < 200:
                entry["improvement_path_pt_br"] = "n/a (fake match)."
            comps[code] = entry
        responses.append(
            {
                "eliminatory_flags": [],
                "competencies": comps,
                "final_score": ex.expected_total,
            }
        )
    return FakeProvider(responses=responses)


@pytest.mark.asyncio
async def test_golden_dataset_mvp_tier_gate(tmp_path: Path) -> None:
    if not ESSAY_BR_TEST.is_dir() or not any(ESSAY_BR_TEST.iterdir()):
        pytest.skip(
            "Essay-BR test split not present. Run `python tests/golden/essay_br/ingest.py` to populate."
        )
    test_examples = Example.iter_dir(ESSAY_BR_TEST)
    inep_examples = Example.iter_dir(INEP_DIR) if INEP_DIR.is_dir() else []
    inep_slugs = {ex.slug for ex in inep_examples}
    all_examples = test_examples + inep_examples

    choice = _provider_choice()
    provider = _make_fake_provider(all_examples) if choice == "fake" else _make_real_provider()

    report = await run_examples(all_examples, provider=provider, inep_slugs=inep_slugs)
    report_path = tmp_path / "golden_report.json"
    report_path.write_text(
        json.dumps(report.as_json(), indent=2, ensure_ascii=False), encoding="utf-8"
    )
    print(f"\nGolden report written to {report_path}\n{json.dumps(report.as_json(), indent=2)}")

    assert report.passes_mvp_tier(), "MVP-tier golden gate failed; see report:\n" + json.dumps(
        report.as_json(), indent=2
    )
