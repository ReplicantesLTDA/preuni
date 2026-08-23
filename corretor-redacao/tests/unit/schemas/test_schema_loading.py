"""Unit tests for src.schemas' load_schema -- the not-found branch was
untested (only the happy path is exercised elsewhere via
active_correction_output_schema())."""

from __future__ import annotations

import pytest

from src.schemas import load_schema


def test_load_schema_raises_for_a_missing_version_or_name():
    with pytest.raises(FileNotFoundError):
        load_schema(version="v999-does-not-exist", name="correction_output")
