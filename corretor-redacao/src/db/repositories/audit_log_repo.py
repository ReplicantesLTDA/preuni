"""AuditLogWriter — Constitution VII enforcement at the database boundary.

Allowlists payload keys per event_type and rejects forbidden PII keys at any
nesting depth. The Postgres column is permissive jsonb; the writer is the
only sanctioned path for inserts.
"""

from __future__ import annotations

import uuid
from sqlalchemy.ext.asyncio import AsyncSession
from typing import Any

from src.db.models import AuditEventType, CorrectionAuditLog

# Constitution VII: these keys must never appear in audit-log payloads, at any
# nesting depth. Centralized for grep-ability.
FORBIDDEN_KEYS: frozenset[str] = frozenset(
    {
        "essay_text",
        "email",
        "name",
        "student_id",
        "user_id",
        "password",
        "refresh_token",
        "access_token",
    }
)

# Per event_type, the explicit allowlist of payload keys. New event types MUST
# extend this map; the writer rejects unknown keys.
PAYLOAD_ALLOWLIST: dict[AuditEventType, frozenset[str]] = {
    AuditEventType.submitted: frozenset(
        {
            "text_length_chars",
            "text_length_lines",
            "language_detected",
        }
    ),
    AuditEventType.llm_started: frozenset(
        {
            "prompt_version",
            "model_identifier",
            "seed",
            "attempt",
        }
    ),
    AuditEventType.llm_completed: frozenset(
        {
            "prompt_tokens",
            "completion_tokens",
            "latency_ms",
            "cost_usd",
            "attempt",
        }
    ),
    AuditEventType.schema_failed: frozenset(
        {
            "attempt",
            "validator_message",
        }
    ),
    AuditEventType.retry: frozenset(
        {
            "attempt",
            "reason",
        }
    ),
    AuditEventType.retry_exhausted: frozenset(
        {
            "final_error_code",
        }
    ),
    AuditEventType.completed: frozenset(
        {
            "final_score",
            "eliminatory_flags",
            "prompt_version",
            "model_identifier",
        }
    ),
    AuditEventType.failed: frozenset(
        {
            "error_code",
            "quota_consumed",
        }
    ),
    AuditEventType.deleted: frozenset(
        {
            "deletion_reason",
        }
    ),
}


class ForbiddenAuditKeyError(ValueError):
    """Raised when an audit payload contains a Constitution-VII-forbidden key."""


class UnknownAuditKeyError(ValueError):
    """Raised when an audit payload contains a key not in PAYLOAD_ALLOWLIST."""


def _walk_for_forbidden(value: Any) -> None:
    """Recursively scan dicts/lists for forbidden keys; raise on first hit."""
    if isinstance(value, dict):
        for k, v in value.items():
            if k in FORBIDDEN_KEYS:
                raise ForbiddenAuditKeyError(
                    f"Forbidden PII key {k!r} in audit payload "
                    f"(Constitution VII). Strip before passing to AuditLogWriter."
                )
            _walk_for_forbidden(v)
    elif isinstance(value, (list, tuple)):
        for item in value:
            _walk_for_forbidden(item)


class AuditLogWriter:
    """Sole sanctioned writer for `correction_audit_logs` rows."""

    def __init__(self, session: AsyncSession | None) -> None:
        self.session = session

    def _validate_payload(self, event_type: AuditEventType, payload: dict[str, Any]) -> None:
        if not isinstance(payload, dict):
            raise TypeError(f"payload must be a dict, got {type(payload).__name__}")
        # First: forbidden keys at any depth.
        _walk_for_forbidden(payload)
        # Then: top-level allowlist.
        allowed = PAYLOAD_ALLOWLIST.get(event_type)
        if allowed is None:
            raise UnknownAuditKeyError(f"unknown event_type: {event_type}")
        extra = set(payload.keys()) - allowed
        if extra:
            raise UnknownAuditKeyError(
                f"payload keys {extra} not in allowlist for {event_type.value}; "
                f"allowed: {sorted(allowed)}"
            )

    async def write(
        self,
        *,
        correction_id: uuid.UUID,
        input_hash: bytes,
        event_type: AuditEventType,
        payload: dict[str, Any] | None = None,
    ) -> CorrectionAuditLog:
        payload = payload or {}
        self._validate_payload(event_type, payload)
        if self.session is None:
            raise RuntimeError("AuditLogWriter requires a session to persist")
        row = CorrectionAuditLog(
            id=uuid.uuid4(),
            correction_id=correction_id,
            input_hash=input_hash,
            event_type=event_type,
            event_payload=payload,
        )
        self.session.add(row)
        return row


__all__ = [
    "FORBIDDEN_KEYS",
    "PAYLOAD_ALLOWLIST",
    "AuditLogWriter",
    "ForbiddenAuditKeyError",
    "UnknownAuditKeyError",
]
