"""T060: AuditLogWriter allowlists keys per event_type; rejects PII keys."""

from __future__ import annotations

import pytest
import uuid
from sqlalchemy import text

from src.db.models import AuditEventType
from src.db.repositories.audit_log_repo import (
    FORBIDDEN_KEYS,
    PAYLOAD_ALLOWLIST,
    AuditLogWriter,
    ForbiddenAuditKeyError,
    UnknownAuditKeyError,
)


def test_allowlist_covers_every_event_type() -> None:
    for et in AuditEventType:
        assert et in PAYLOAD_ALLOWLIST, f"event_type {et} missing from PAYLOAD_ALLOWLIST"


def test_forbidden_keys_complete() -> None:
    # Constitution VII forbidden set must include at least:
    required = {"essay_text", "email", "name", "student_id", "user_id", "password"}
    assert required.issubset(FORBIDDEN_KEYS)


@pytest.mark.parametrize("event_type", list(AuditEventType))
def test_writer_rejects_forbidden_keys(event_type: AuditEventType) -> None:
    writer = AuditLogWriter(session=None)  # session not needed for validation-only check
    for forbidden in ("essay_text", "email", "name", "password"):
        with pytest.raises(ForbiddenAuditKeyError):
            writer._validate_payload(event_type, {forbidden: "x"})


def test_writer_rejects_keys_outside_allowlist() -> None:
    writer = AuditLogWriter(session=None)
    with pytest.raises(UnknownAuditKeyError):
        writer._validate_payload(AuditEventType.submitted, {"random_unallowed_key": 1})


def test_writer_accepts_allowlisted_keys() -> None:
    writer = AuditLogWriter(session=None)
    for et, keys in PAYLOAD_ALLOWLIST.items():
        payload = dict.fromkeys(keys, "x")
        writer._validate_payload(et, payload)  # must not raise


def test_writer_rejects_forbidden_keys_nested() -> None:
    writer = AuditLogWriter(session=None)
    with pytest.raises(ForbiddenAuditKeyError):
        writer._validate_payload(
            AuditEventType.submitted,
            {"text_length_chars": 100, "nested": {"email": "x@y"}},
        )


@pytest.mark.asyncio
async def test_writer_persists_row(db_session) -> None:
    # Need a correction row to satisfy FK. Create a user + correction first.
    user_id = uuid.uuid4()
    correction_id = uuid.uuid4()
    await db_session.execute(
        text("""
        INSERT INTO users (id, email, password_hash, tier, created_at, updated_at)
        VALUES (:uid, :email, 'x', 'free', now(), now())
    """),
        {"uid": user_id, "email": f"u-{uuid.uuid4()}@x.test"},
    )
    await db_session.execute(
        text("""
        INSERT INTO corrections (id, user_id, essay_text, prompt_theme_title,
                                  prompt_theme_context, input_hash, status,
                                  queued_at, eliminatory_flags, quota_consumed)
        VALUES (:cid, :uid, 'text', 'title', 'ctx', :hash, 'pending', now(),
                '[]'::jsonb, true)
    """),
        {"cid": correction_id, "uid": user_id, "hash": b"\x00" * 32},
    )
    await db_session.flush()

    writer = AuditLogWriter(session=db_session)
    await writer.write(
        correction_id=correction_id,
        input_hash=b"\x00" * 32,
        event_type=AuditEventType.submitted,
        payload={"text_length_chars": 1234, "text_length_lines": 7, "language_detected": "pt"},
    )
    await db_session.flush()

    row = (
        await db_session.execute(
            text(
                "SELECT event_type, event_payload FROM correction_audit_logs WHERE correction_id = :cid"
            ),
            {"cid": correction_id},
        )
    ).first()
    assert row is not None
    assert row.event_type == "submitted"
    assert row.event_payload["text_length_chars"] == 1234
