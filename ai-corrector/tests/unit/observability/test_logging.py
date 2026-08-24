"""Unit tests for observability.logging's context-var helpers and
configure_logging -- the PII scrubber already had test_pii_filter.py, but
ContextVarsFilter, the correction_id/user_id accessors, and
configure_logging itself (62% coverage) were untested."""

from __future__ import annotations

from src.observability.logging import (
    ContextVarsFilter,
    configure_logging,
    get_correction_id,
    get_user_id,
    set_correction_id,
    set_user_id,
)


def test_correction_id_roundtrips_through_context():
    set_correction_id("corr-123")
    try:
        assert get_correction_id() == "corr-123"
    finally:
        set_correction_id(None)
    assert get_correction_id() is None


def test_user_id_roundtrips_through_context():
    set_user_id("user-456")
    try:
        assert get_user_id() == "user-456"
    finally:
        set_user_id(None)
    assert get_user_id() is None


def test_context_vars_filter_adds_ids_when_set():
    set_correction_id("corr-abc")
    set_user_id("user-xyz")
    try:
        event = ContextVarsFilter()(None, "info", {"event": "test"})
        assert event["correction_id"] == "corr-abc"
        assert event["user_id"] == "user-xyz"
    finally:
        set_correction_id(None)
        set_user_id(None)


def test_context_vars_filter_omits_ids_when_unset():
    set_correction_id(None)
    set_user_id(None)
    event = ContextVarsFilter()(None, "info", {"event": "test"})
    assert "correction_id" not in event
    assert "user_id" not in event


def test_configure_logging_json_format_does_not_raise():
    configure_logging(log_format="json", log_level="INFO")


def test_configure_logging_console_format_does_not_raise():
    configure_logging(log_format="console", log_level="DEBUG", disable_existing=False)
