"""T047: walking-skeleton CLI behavior."""

from __future__ import annotations

import json
import pytest
from pathlib import Path

from src.corrector.llm.errors import RateLimitError
from src.corrector.llm.ollama import OllamaProvider
from src.scripts.correct_essay import _make_provider, _typed_error, run_cli

ESSAY_BODY = (
    "A inclusão digital de pessoas idosas exige ação coordenada do Estado brasileiro.\n"
    "Você pode notar, no cotidiano, que muitos avós ainda têm dificuldade para utilizar\n"
    "aplicativos de banco e plataformas governamentais essenciais para a cidadania.\n"
    "Diante disso, propõe-se uma estratégia coordenada de inclusão digital intergeracional.\n"
    "Cursos gratuitos em telecentros e parcerias com escolas técnicas formam a base do plano.\n"
    "Portanto, cabe ao MEC articular um programa nacional de letramento digital para idosos.\n"
    "Assim, o fosso digital geracional será reduzido substancialmente no curto prazo brasileiro."
)


def _write_essay(tmp: Path, body: str = ESSAY_BODY) -> Path:
    p = tmp / "essay.txt"
    p.write_text(body, encoding="utf-8")
    return p


@pytest.mark.asyncio
async def test_happy_path_prints_correction_json(
    tmp_path: Path, capsys: pytest.CaptureFixture[str]
) -> None:
    essay_path = _write_essay(tmp_path)
    code = await run_cli(
        [
            "--essay",
            str(essay_path),
            "--theme-title",
            "Inclusão digital de idosos",
            "--theme-context",
            "A democratização digital exclui idosos. Discutir caminhos para reverter o quadro.",
            "--provider",
            "fake-perfect",
        ],
        env={},
    )
    assert code == 0
    out = capsys.readouterr().out
    payload = json.loads(out)
    assert payload["final_score"] == 1000
    assert set(payload["competencies"].keys()) == {"c1", "c2", "c3", "c4", "c5"}


@pytest.mark.asyncio
async def test_length_too_short_emits_typed_error(
    tmp_path: Path, capsys: pytest.CaptureFixture[str]
) -> None:
    essay_path = _write_essay(tmp_path, body="muito curto")
    code = await run_cli(
        [
            "--essay",
            str(essay_path),
            "--theme-title",
            "Tema",
            "--theme-context",
            "Contexto qualquer com mais de vinte caracteres.",
            "--provider",
            "fake-perfect",
        ],
        env={},
    )
    assert code != 0
    payload = json.loads(capsys.readouterr().out)
    assert payload["error_code"] == "length_too_short"


@pytest.mark.asyncio
async def test_language_mismatch_emits_typed_error(
    tmp_path: Path, capsys: pytest.CaptureFixture[str]
) -> None:
    english = "Digital inclusion of elderly people is a pressing contemporary challenge. " * 8
    essay_path = _write_essay(tmp_path, body=english)
    code = await run_cli(
        [
            "--essay",
            str(essay_path),
            "--theme-title",
            "Tema",
            "--theme-context",
            "Contexto qualquer com mais de vinte caracteres válido.",
            "--provider",
            "fake-perfect",
        ],
        env={},
    )
    assert code != 0
    payload = json.loads(capsys.readouterr().out)
    assert payload["error_code"] == "language_mismatch"


@pytest.mark.asyncio
async def test_theme_missing_context_emits_typed_error(
    tmp_path: Path, capsys: pytest.CaptureFixture[str]
) -> None:
    essay_path = _write_essay(tmp_path)
    code = await run_cli(
        [
            "--essay",
            str(essay_path),
            "--theme-title",
            "Tema válido",
            "--theme-context",
            "",
            "--provider",
            "fake-perfect",
        ],
        env={},
    )
    assert code != 0
    payload = json.loads(capsys.readouterr().out)
    assert payload["error_code"] == "theme_missing_context"


@pytest.mark.asyncio
async def test_motivational_texts_file_is_read_and_passed_through(tmp_path: Path) -> None:
    essay_path = _write_essay(tmp_path)
    motivational_path = tmp_path / "motivational.txt"
    motivational_path.write_text("Texto motivador de apoio.", encoding="utf-8")
    code = await run_cli(
        [
            "--essay",
            str(essay_path),
            "--theme-title",
            "Inclusão digital de idosos",
            "--theme-context",
            "A democratização digital exclui idosos. Discutir caminhos para reverter o quadro.",
            "--motivational",
            str(motivational_path),
            "--provider",
            "fake-perfect",
        ],
        env={},
    )
    assert code == 0


def test_make_provider_builds_an_ollama_provider_from_env() -> None:
    provider = _make_provider(
        "ollama",
        {
            "OLLAMA_CLOUD_BASE_URL": "https://ollama.cloud.test",
            "OLLAMA_CLOUD_API_KEY": "secret",
            "LLM_MODEL_ID": "custom-model",
        },
        "essay text",
    )
    assert isinstance(provider, OllamaProvider)
    assert provider.base_url == "https://ollama.cloud.test"
    assert provider.api_key == "secret"
    assert provider.model_id == "custom-model"


def test_make_provider_falls_back_to_defaults_with_empty_env() -> None:
    provider = _make_provider("ollama", {}, "essay text")
    assert isinstance(provider, OllamaProvider)
    assert provider.base_url == "https://ollama.com"
    assert provider.api_key is None
    assert provider.model_id == "kimi-k2.6"


def test_typed_error_maps_a_generic_llm_error_by_class_name() -> None:
    code, msg = _typed_error(RateLimitError("too many requests"))
    assert code == "ratelimit_error"
    assert msg == "too many requests"


def test_typed_error_falls_back_to_internal_error() -> None:
    code, msg = _typed_error(ValueError("something else"))
    assert code == "internal_error"
    assert msg == "something else"
