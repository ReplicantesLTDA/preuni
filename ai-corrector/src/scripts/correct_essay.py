"""Walking-skeleton CLI.

Reads one essay + theme + (optional) motivational texts and prints the
correction JSON to stdout. No DB. No auth. No API. The only purpose is to
exercise the riskiest path: provider abstraction → structured output → schema.
"""

from __future__ import annotations

import argparse
import asyncio
import json
import os
import sys
import uuid
from typing import Any

from src.corrector.competencies._common import CompetencyParseError
from src.corrector.graders.single_grader import SingleGraderInput
from src.corrector.llm.base import LLMProvider
from src.corrector.llm.errors import LLMError, SchemaViolationError
from src.corrector.llm.fake import FakeProvider
from src.corrector.llm.ollama import OllamaProvider
from src.corrector.pipeline import correct_essay
from src.corrector.prevalidation.language import LanguageMismatchError
from src.corrector.prevalidation.length import LengthTooLongError, LengthTooShortError
from src.corrector.prevalidation.theme import (
    ThemeMissingContextError,
    ThemeMissingTitleError,
)


def _build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(prog="correct-essay")
    p.add_argument("--essay", required=True, help="Path to essay text file.")
    p.add_argument("--theme-title", required=True)
    p.add_argument("--theme-context", required=True)
    p.add_argument("--motivational", help="Path to motivational texts file.")
    p.add_argument(
        "--provider",
        default="ollama",
        choices=("ollama", "fake-perfect"),
        help="LLM provider. 'fake-perfect' bypasses LLM with a canned 1000/1000 output.",
    )
    return p


def _perfect_fake(essay_text: str) -> FakeProvider:
    excerpt = essay_text[:60]
    canned = {
        "eliminatory_flags": [],
        "competencies": {
            code: {
                "score": 200,
                "excerpt": excerpt,
                "justification_pt_br": "Resposta canônica do FakeProvider.",
            }
            for code in ("c1", "c2", "c3", "c4", "c5")
        },
        "final_score": 1000,
    }
    return FakeProvider(responses=[canned])


def _make_provider(name: str, env: dict[str, str], essay_text: str) -> LLMProvider:
    if name == "fake-perfect":
        return _perfect_fake(essay_text)
    # Cloud-specific vars take priority; falls back to generic OLLAMA_*.
    base_url = (
        env.get("OLLAMA_CLOUD_BASE_URL") or env.get("OLLAMA_BASE_URL") or "https://ollama.com"
    )
    api_key = env.get("OLLAMA_CLOUD_API_KEY") or env.get("OLLAMA_API_KEY")
    model_id = env.get("LLM_MODEL_ID") or env.get("OLLAMA_MODEL") or "kimi-k2.6"
    return OllamaProvider(base_url=base_url, api_key=api_key, model_id=model_id)


_ERROR_CODES = {
    LengthTooShortError: "length_too_short",
    LengthTooLongError: "length_too_long",
    LanguageMismatchError: "language_mismatch",
    ThemeMissingTitleError: "theme_missing_title",
    ThemeMissingContextError: "theme_missing_context",
    CompetencyParseError: "schema_violation",
    SchemaViolationError: "schema_violation",
}


def _typed_error(exc: BaseException) -> tuple[str, str]:
    for cls, code in _ERROR_CODES.items():
        if isinstance(exc, cls):
            return code, str(exc) or cls.__name__
    if isinstance(exc, LLMError):
        return type(exc).__name__.lower().replace("error", "_error").strip("_"), str(exc)
    return "internal_error", str(exc)


async def run_cli(argv: list[str], *, env: dict[str, str] | None = None) -> int:
    args = _build_parser().parse_args(argv)
    if env is None:
        try:
            import dotenv

            dotenv.load_dotenv()
        except ImportError:
            pass
    env = dict(env if env is not None else os.environ)

    with open(args.essay, encoding="utf-8") as fh:
        essay = fh.read()
    motivational = None
    if args.motivational:
        with open(args.motivational, encoding="utf-8") as fh:
            motivational = fh.read()

    provider = _make_provider(args.provider, env, essay)

    try:
        result = await correct_essay(
            SingleGraderInput(
                correction_id=uuid.uuid4(),
                essay_text=essay,
                prompt_theme_title=args.theme_title,
                prompt_theme_context=args.theme_context,
                motivational_texts=motivational,
            ),
            provider=provider,
        )
    except Exception as exc:
        code, msg = _typed_error(exc)
        json.dump({"error_code": code, "message": msg}, sys.stdout, ensure_ascii=False)
        sys.stdout.write("\n")
        return 1

    payload: dict[str, Any] = {
        "correction_id": str(result.correction_id),
        "final_score": result.final_score,
        "eliminatory_flags": result.eliminatory_flags,
        "annulment_reason_pt_br": result.annulment_reason_pt_br,
        "competencies": {
            code: {
                "score": e.score,
                "excerpt": e.excerpt,
                "justification_pt_br": e.justification_pt_br,
                "improvement_path_pt_br": e.improvement_path_pt_br,
            }
            for code, e in result.competencies.items()
        },
        "prompt_version": result.prompt_version,
        "model_identifier": result.model_id,
        "inference_params": result.inference_params,
        "output_schema_version": result.output_schema_version,
        "seed": result.seed,
        "attempts": result.attempts,
        "latency_ms": result.latency_ms,
    }
    json.dump(payload, sys.stdout, ensure_ascii=False)
    sys.stdout.write("\n")
    return 0


def main() -> None:
    sys.exit(asyncio.run(run_cli(sys.argv[1:])))


if __name__ == "__main__":
    main()
