"""T032: prompt-theme pre-validation (FR-004)."""

from __future__ import annotations

import pytest

from src.corrector.prevalidation.theme import (
    ThemeMissingContextError,
    ThemeMissingTitleError,
    validate_theme,
)


def test_both_fields_pass() -> None:
    validate_theme(title="Inclusão digital de idosos", context="Texto contextualizando.")


def test_missing_title_fails() -> None:
    with pytest.raises(ThemeMissingTitleError):
        validate_theme(title="", context="contexto suficiente.")


def test_whitespace_only_title_fails() -> None:
    with pytest.raises(ThemeMissingTitleError):
        validate_theme(title="   ", context="contexto suficiente.")


def test_missing_context_fails() -> None:
    with pytest.raises(ThemeMissingContextError):
        validate_theme(title="Tema válido", context="")


def test_short_context_fails_as_missing() -> None:
    # Less than ~20 chars of meaningful context => treat as missing.
    with pytest.raises(ThemeMissingContextError):
        validate_theme(title="Tema válido", context="curto")
