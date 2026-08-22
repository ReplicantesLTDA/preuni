"""Postgres enum types shared across models."""

from __future__ import annotations

import enum


class UserTier(enum.StrEnum):
    free = "free"
    premium = "premium"


class CorrectionStatus(enum.StrEnum):
    pending = "pending"
    processing = "processing"
    completed = "completed"
    failed = "failed"


class AuditEventType(enum.StrEnum):
    submitted = "submitted"
    llm_started = "llm_started"
    llm_completed = "llm_completed"
    schema_failed = "schema_failed"
    retry = "retry"
    retry_exhausted = "retry_exhausted"
    completed = "completed"
    failed = "failed"
    deleted = "deleted"


__all__ = ["AuditEventType", "CorrectionStatus", "UserTier"]
