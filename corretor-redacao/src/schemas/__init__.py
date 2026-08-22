"""Versioned JSON Schemas for LLM structured output.

Constitution Article IV: every LLM response is validated against a versioned
schema. Free-form fallback parsing is forbidden.
"""

from __future__ import annotations

import json
from functools import lru_cache
from pathlib import Path
from typing import Any

ACTIVE_VERSION = "v1"
SCHEMAS_DIR = Path(__file__).parent


@lru_cache(maxsize=8)
def load_schema(version: str = ACTIVE_VERSION, name: str = "correction_output") -> dict[str, Any]:
    """Load a versioned JSON Schema from disk."""
    path = SCHEMAS_DIR / version / f"{name}.schema.json"
    if not path.is_file():
        raise FileNotFoundError(f"Schema not found: {path}")
    with path.open("r", encoding="utf-8") as fh:
        return json.load(fh)


def active_correction_output_schema() -> dict[str, Any]:
    """Current correction-output schema."""
    return load_schema(ACTIVE_VERSION, "correction_output")


__all__ = ["ACTIVE_VERSION", "active_correction_output_schema", "load_schema"]
