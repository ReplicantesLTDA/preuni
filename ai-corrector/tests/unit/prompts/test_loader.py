"""T026: PromptLoader contract."""

from __future__ import annotations

import pytest
from pathlib import Path
from textwrap import dedent

from src.corrector.prompts.loader import PromptLoader, PromptTemplateError


@pytest.fixture
def prompts_dir(tmp_path: Path) -> Path:
    root = tmp_path / "prompts"
    (root / "v1.0.0").mkdir(parents=True)
    (root / "v1.0.0" / "system.md").write_text(
        dedent("""\
            Tema: {prompt_theme_title}
            Contexto: {prompt_theme_context}
            Motivadores: {motivational_texts}

            Redacao:
            {essay_text}
        """),
        encoding="utf-8",
    )
    (root / "v1.1.0").mkdir()
    (root / "v1.1.0" / "system.md").write_text("v1.1.0 stub", encoding="utf-8")
    return root


def test_resolves_newest_version_by_default(
    monkeypatch: pytest.MonkeyPatch, prompts_dir: Path
) -> None:
    monkeypatch.delenv("PROMPT_VERSION", raising=False)
    loader = PromptLoader(prompts_dir=prompts_dir)
    assert loader.active_version == "1.1.0"


def test_explicit_version_env_overrides(monkeypatch: pytest.MonkeyPatch, prompts_dir: Path) -> None:
    monkeypatch.setenv("PROMPT_VERSION", "1.0.0")
    loader = PromptLoader(prompts_dir=prompts_dir)
    assert loader.active_version == "1.0.0"


def test_substitutes_placeholders(prompts_dir: Path) -> None:
    loader = PromptLoader(prompts_dir=prompts_dir, version="1.0.0")
    rendered = loader.render(
        "system",
        prompt_theme_title="Inclusão digital",
        prompt_theme_context="Contexto qualquer.",
        motivational_texts="Textos motivadores.",
        essay_text="Texto da redação.",
    )
    assert "Tema: Inclusão digital" in rendered
    assert "Contexto: Contexto qualquer." in rendered
    assert "Motivadores: Textos motivadores." in rendered
    assert "Texto da redação." in rendered


def test_unknown_placeholder_raises(prompts_dir: Path) -> None:
    (prompts_dir / "v1.0.0" / "broken.md").write_text("Hello {unknown_key}", encoding="utf-8")
    loader = PromptLoader(prompts_dir=prompts_dir, version="1.0.0")
    with pytest.raises(PromptTemplateError):
        loader.render("broken", essay_text="x")


def test_missing_template_raises(prompts_dir: Path) -> None:
    loader = PromptLoader(prompts_dir=prompts_dir, version="1.0.0")
    with pytest.raises(FileNotFoundError):
        loader.render("nonexistent")


def test_caches_loaded_template(prompts_dir: Path) -> None:
    loader = PromptLoader(prompts_dir=prompts_dir, version="1.0.0")
    first = loader._raw("system")
    second = loader._raw("system")
    assert first is second


def test_renders_with_empty_optional_strings(prompts_dir: Path) -> None:
    loader = PromptLoader(prompts_dir=prompts_dir, version="1.0.0")
    rendered = loader.render(
        "system",
        prompt_theme_title="X",
        prompt_theme_context="Y",
        motivational_texts="",
        essay_text="Z",
    )
    assert "Motivadores: " in rendered
