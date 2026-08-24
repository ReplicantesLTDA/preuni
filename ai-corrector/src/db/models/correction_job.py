"""correction_jobs — the Go monolith <-> correction-worker bridge/outbox table.

The Go monolith is the only writer of new rows (INSERT-only grant; see
infra/migrations/grant-correction-jobs.sql and
contracts/internal-bridge.md). This service's worker claims rows here,
does the grading work, and writes the result to `corrections`
(schema `correction`, keyed by `Correction.job_id`), then flips this row's
status so the monolith's poller can reconcile it.
"""

from __future__ import annotations

import datetime as dt
import uuid
from sqlalchemy import DateTime, Enum, Index, Text, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column

from src.db.models import Base
from src.db.models.enums import CorrectionJobStatus


class CorrectionJob(Base):
    __tablename__ = "correction_jobs"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True)
    user_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), nullable=False)

    essay_text: Mapped[str] = mapped_column(Text, nullable=False)
    prompt_theme_title: Mapped[str] = mapped_column(Text, nullable=False)
    prompt_theme_context: Mapped[str] = mapped_column(Text, nullable=False)

    status: Mapped[CorrectionJobStatus] = mapped_column(
        Enum(CorrectionJobStatus, name="correction_job_status", native_enum=True),
        nullable=False,
        default=CorrectionJobStatus.pending,
    )
    queued_at: Mapped[dt.datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, server_default=func.now()
    )
    started_at: Mapped[dt.datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    completed_at: Mapped[dt.datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)

    __table_args__ = (
        Index(
            "correction_jobs_queue_idx",
            "queued_at",
            postgresql_where=(status == CorrectionJobStatus.pending),
        ),
    )
