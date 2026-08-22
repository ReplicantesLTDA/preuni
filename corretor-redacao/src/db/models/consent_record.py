"""Parental / guardian consent record (Constitution X — minor protection)."""

from __future__ import annotations

import datetime as dt
import uuid
from sqlalchemy import DateTime, ForeignKey, Index, LargeBinary, Text, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column

from src.db.models import Base


class ConsentRecord(Base):
    __tablename__ = "consent_records"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True)
    user_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("users.id", ondelete="CASCADE"),
        nullable=False,
    )
    guardian_name: Mapped[str] = mapped_column(Text, nullable=False)
    guardian_relation: Mapped[str] = mapped_column(Text, nullable=False)
    consent_text_sha256: Mapped[bytes] = mapped_column(LargeBinary(32), nullable=False)
    accepted_at: Mapped[dt.datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, server_default=func.now()
    )
    revoked_at: Mapped[dt.datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)

    __table_args__ = (
        Index(
            "consent_records_user_idx",
            "user_id",
            postgresql_where=(revoked_at.is_(None)),
        ),
    )
