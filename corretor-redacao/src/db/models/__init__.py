"""SQLAlchemy declarative Base + model registrations.

Importing this package side-effect-loads every model module so the shared
`Base.metadata` is complete for autogenerate (`alembic revision --autogenerate`)
and for unit tests that introspect the schema.
"""

from __future__ import annotations

from sqlalchemy.orm import DeclarativeBase


class Base(DeclarativeBase):
    pass


# Enums first (referenced by other models).
from src.db.models.consent_record import ConsentRecord  # noqa: E402
from src.db.models.correction import Correction  # noqa: E402
from src.db.models.correction_audit_log import CorrectionAuditLog  # noqa: E402
from src.db.models.email_verification_token import EmailVerificationToken  # noqa: E402
from src.db.models.enums import (  # noqa: E402
    AuditEventType,
    CorrectionStatus,
    UserTier,
)
from src.db.models.grader_pass import GraderPass  # noqa: E402
from src.db.models.refresh_token import RefreshToken  # noqa: E402

# Import every model so SQLAlchemy registers them on Base.metadata.
from src.db.models.user import User  # noqa: E402

__all__ = [
    "AuditEventType",
    "Base",
    "ConsentRecord",
    "Correction",
    "CorrectionAuditLog",
    "CorrectionStatus",
    "EmailVerificationToken",
    "GraderPass",
    "RefreshToken",
    "User",
    "UserTier",
]
