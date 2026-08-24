"""correction_audit_logs — Constitution VII (observability + PII privacy).

event_payload allowlist enforced by `db/repositories/audit_log_repo.py:AuditLogWriter`.
Schema is permissive at the DB level (JSONB); enforcement lives in the writer.
"""

from __future__ import annotations

import datetime as dt
import uuid
from sqlalchemy import DateTime, Enum, ForeignKey, Index, LargeBinary, func
from sqlalchemy.dialects.postgresql import JSONB, UUID
from sqlalchemy.orm import Mapped, mapped_column
from typing import Any

from src.db.models import Base
from src.db.models.enums import AuditEventType


class CorrectionAuditLog(Base):
    __tablename__ = "correction_audit_logs"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True)
    correction_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("corrections.id", ondelete="CASCADE"),
        nullable=False,
    )
    input_hash: Mapped[bytes] = mapped_column(LargeBinary(32), nullable=False)
    event_type: Mapped[AuditEventType] = mapped_column(
        Enum(AuditEventType, name="audit_event_type", native_enum=True),
        nullable=False,
    )
    event_payload: Mapped[dict[str, Any]] = mapped_column(JSONB, nullable=False, server_default="{}")
    created_at: Mapped[dt.datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, server_default=func.now()
    )

    __table_args__ = (
        Index(
            "correction_audit_logs_correction_idx",
            "correction_id",
            "created_at",
        ),
    )
