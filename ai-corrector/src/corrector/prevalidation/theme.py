"""Prompt-theme pre-validation (FR-004)."""

from __future__ import annotations

MIN_CONTEXT_CHARS = 20


class ThemeMissingTitleError(ValueError):
    """Spec error_code: theme_missing_title."""


class ThemeMissingContextError(ValueError):
    """Spec error_code: theme_missing_context."""


def validate_theme(*, title: str, context: str) -> None:
    if not title or not title.strip():
        raise ThemeMissingTitleError("prompt theme requires a non-empty title")
    if not context or len(context.strip()) < MIN_CONTEXT_CHARS:
        raise ThemeMissingContextError(
            f"prompt theme requires a contextualization of at least {MIN_CONTEXT_CHARS} characters"
        )


__all__ = ["ThemeMissingContextError", "ThemeMissingTitleError", "validate_theme"]
