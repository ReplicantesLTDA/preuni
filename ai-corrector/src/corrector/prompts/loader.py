"""Versioned prompt loader (Constitution Article VIII: prompts-as-code).

Layout: `<root>/v<semver>/<name>.md`. Loader picks the newest version by
default; `PROMPT_VERSION` env var or constructor arg overrides.
"""

from __future__ import annotations

import os
import re
from functools import cache
from pathlib import Path

_PROMPT_ROOT = Path(__file__).parent
_SEMVER_DIR = re.compile(r"^v(\d+\.\d+\.\d+)$")


class PromptTemplateError(Exception):
    """Raised when a template references a placeholder not supplied by the caller."""


class PromptLoader:
    def __init__(
        self,
        *,
        prompts_dir: Path | None = None,
        version: str | None = None,
    ) -> None:
        self.prompts_dir = prompts_dir or _PROMPT_ROOT
        if version is None:
            version = os.environ.get("PROMPT_VERSION") or self._newest_version_on_disk()
        if not version:
            raise FileNotFoundError(f"No prompt version directories found in {self.prompts_dir}")
        # Tolerate both "1.0.0" and "v1.0.0" forms.
        if version.startswith("v"):
            version = version[1:]
        self.active_version = version
        self._version_dir = self.prompts_dir / f"v{version}"
        if not self._version_dir.is_dir():
            raise FileNotFoundError(f"Prompt version dir not found: {self._version_dir}")

    def _newest_version_on_disk(self) -> str | None:
        versions: list[tuple[int, int, int]] = []
        if not self.prompts_dir.is_dir():
            return None
        for child in self.prompts_dir.iterdir():
            m = _SEMVER_DIR.match(child.name)
            if m and child.is_dir():
                versions.append(tuple(int(p) for p in m.group(1).split(".")))  # type: ignore[arg-type]
        if not versions:
            return None
        versions.sort()
        return ".".join(str(p) for p in versions[-1])

    @cache  # noqa: B019 — per-instance cache fine; loader is itself memoized at app level
    def _raw(self, name: str) -> str:
        path = self._version_dir / f"{name}.md"
        if not path.is_file():
            raise FileNotFoundError(f"Prompt template not found: {path}")
        return path.read_text(encoding="utf-8")

    def render(self, name: str, **values: str) -> str:
        raw = self._raw(name)
        try:
            return _strict_format(raw, values)
        except KeyError as exc:
            raise PromptTemplateError(
                f"Template '{name}' references unknown placeholder {exc.args[0]!r}"
            ) from exc


_PLACEHOLDER = re.compile(r"\{([A-Za-z_][A-Za-z0-9_]*)\}")


def _strict_format(template: str, values: dict[str, str]) -> str:
    """Replace `{identifier}` placeholders. Non-identifier `{...}` is treated as
    literal text. Raises KeyError when an identifier-shaped placeholder is not
    supplied.
    """

    def _sub(m: re.Match[str]) -> str:
        key = m.group(1)
        if key not in values:
            raise KeyError(key)
        return values[key]

    return _PLACEHOLDER.sub(_sub, template)


__all__ = ["PromptLoader", "PromptTemplateError"]
