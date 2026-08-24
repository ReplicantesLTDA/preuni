"""Postgres enum types shared across models."""

from __future__ import annotations

import enum


class CorrectionStatus(enum.StrEnum):
    pending = "pending"
    processing = "processing"
    completed = "completed"
    failed = "failed"


class CorrectionJobStatus(enum.StrEnum):
    """Status of a row in the correction_jobs bridge table.

    Mirrors CorrectionStatus but is a distinct DB enum type because it lives
    on a table the Go monolith also reads (contracts/internal-bridge.md) —
    keeping it separate means either side's status vocabulary can evolve
    independently of the other's.
    """

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


__all__ = ["AuditEventType", "CorrectionJobStatus", "CorrectionStatus"]
