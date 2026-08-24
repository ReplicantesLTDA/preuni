"""T062: PII filter strips forbidden keys even when nested in jsonb-like payloads."""

from __future__ import annotations

import pytest

from src.observability.logging import PIIScrubFilter


def test_strips_top_level_forbidden() -> None:
    f = PIIScrubFilter()
    with pytest.raises(ValueError):
        f({}, "info", {"email": "x@y.com", "ok": "stay"})


def test_strips_nested_in_dict() -> None:
    f = PIIScrubFilter()
    with pytest.raises(ValueError):
        f({}, "info", {"event": {"user": {"email": "x@y.com"}}})


def test_strips_nested_in_list_of_dicts() -> None:
    f = PIIScrubFilter()
    with pytest.raises(ValueError):
        f({}, "info", {"items": [{"email": "x@y.com"}, {"ok": 1}]})


def test_strips_nested_in_jsonb_payload() -> None:
    f = PIIScrubFilter()
    with pytest.raises(ValueError):
        f({}, "info", {"event_payload": {"name": "Alice"}})


def test_passes_clean_record() -> None:
    f = PIIScrubFilter()
    result = f({}, "info", {"correction_id": "c", "latency_ms": 42})
    assert result == {"correction_id": "c", "latency_ms": 42}


def test_passes_deeply_nested_clean_record() -> None:
    f = PIIScrubFilter()
    record = {"a": {"b": {"c": [{"d": "ok"}, 1, 2]}}, "ms": 10}
    assert f({}, "info", record) == record


def test_all_pii_keys_caught() -> None:
    f = PIIScrubFilter()
    for forbidden in (
        "essay_text",
        "email",
        "name",
        "student_id",
        "user_id",
        "password",
        "refresh_token",
        "access_token",
    ):
        with pytest.raises(ValueError):
            f({}, "info", {forbidden: "x"})
