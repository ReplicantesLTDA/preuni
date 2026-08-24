"""SQLAlchemy declarative Base + model registrations.

Importing this package side-effect-loads every model module so the shared
`Base.metadata` is complete for autogenerate (`alembic revision --autogenerate`)
and for unit tests that introspect the schema.

Constitution v2.1.1 / specs/014-constitution-alignment-refactor: this
service's tables live in the `correction` Postgres schema (research.md #1),
not `public` — the monolith owns identity/quota tables elsewhere. `Base`
carries that schema so every model below inherits it without repeating
`__table_args__ = {"schema": ...}` per file.
"""

from __future__ import annotations

from sqlalchemy import MetaData
from sqlalchemy.orm import DeclarativeBase


class Base(DeclarativeBase):
    metadata = MetaData(schema="correction")


# Enums first (referenced by other models).
from src.db.models.correction import Correction  # noqa: E402
from src.db.models.correction_audit_log import CorrectionAuditLog  # noqa: E402
from src.db.models.correction_job import CorrectionJob  # noqa: E402
from src.db.models.enums import (  # noqa: E402
    AuditEventType,
    CorrectionJobStatus,
    CorrectionStatus,
)
from src.db.models.grader_pass import GraderPass  # noqa: E402

__all__ = [
    "AuditEventType",
    "Base",
    "Correction",
    "CorrectionAuditLog",
    "CorrectionJob",
    "CorrectionJobStatus",
    "CorrectionStatus",
    "GraderPass",
]
