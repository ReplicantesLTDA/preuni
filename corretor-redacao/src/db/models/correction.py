"""Corrections table — aggregate result + Postgres-as-queue row.

Constitution I (matrix fidelity) enforced via CHECK constraints:
- per-comp scores must be in {0, 40, 80, 120, 160, 200}
- final_score == sum(c1..c5) when status='completed'
- terminal states have provenance (prompt_version, model_identifier, schema)
- failed state has error_code
"""

from __future__ import annotations

import datetime as dt
import uuid
from sqlalchemy import (
    Boolean,
    CheckConstraint,
    DateTime,
    Enum,
    ForeignKey,
    Index,
    LargeBinary,
    SmallInteger,
    Text,
    func,
)
from sqlalchemy.dialects.postgresql import JSONB, UUID
from sqlalchemy.orm import Mapped, mapped_column
from typing import Any

from src.db.models import Base
from src.db.models.enums import CorrectionStatus

_SCORE_SET = "(0, 40, 80, 120, 160, 200)"


class Correction(Base):
    __tablename__ = "corrections"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True)
    # Opaque reference — the Go monolith owns identity, this service does
    # not (constitution Architecture section; research.md #2). No local FK.
    user_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), nullable=False)
    # The correction_jobs row (schema `correction`, same DB) that this
    # result was produced for. No FK constraint: correction_jobs is the
    # Go-writable bridge table and the two are linked by application code,
    # not referential integrity, to keep the bridge table's contract narrow
    # (contracts/internal-bridge.md).
    job_id: Mapped[uuid.UUID | None] = mapped_column(UUID(as_uuid=True), nullable=True)

    # Input.
    essay_text: Mapped[str] = mapped_column(Text, nullable=False)
    prompt_theme_title: Mapped[str] = mapped_column(Text, nullable=False)
    prompt_theme_context: Mapped[str] = mapped_column(Text, nullable=False)
    motivational_texts: Mapped[str | None] = mapped_column(Text, nullable=True)
    input_hash: Mapped[bytes] = mapped_column(LargeBinary(32), nullable=False)

    # Status / queue.
    status: Mapped[CorrectionStatus] = mapped_column(
        Enum(CorrectionStatus, name="correction_status", native_enum=True),
        nullable=False,
        default=CorrectionStatus.pending,
    )
    queued_at: Mapped[dt.datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, server_default=func.now()
    )
    started_at: Mapped[dt.datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    completed_at: Mapped[dt.datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    locked_at: Mapped[dt.datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    locked_by: Mapped[str | None] = mapped_column(Text, nullable=True)

    # Aggregate result (copied from sole grader pass in MVP).
    final_score: Mapped[int | None] = mapped_column(SmallInteger, nullable=True)
    c1_score: Mapped[int | None] = mapped_column(SmallInteger, nullable=True)
    c2_score: Mapped[int | None] = mapped_column(SmallInteger, nullable=True)
    c3_score: Mapped[int | None] = mapped_column(SmallInteger, nullable=True)
    c4_score: Mapped[int | None] = mapped_column(SmallInteger, nullable=True)
    c5_score: Mapped[int | None] = mapped_column(SmallInteger, nullable=True)
    competencies: Mapped[dict[str, Any] | None] = mapped_column(JSONB, nullable=True)
    eliminatory_flags: Mapped[list[Any]] = mapped_column(JSONB, nullable=False, server_default="[]")

    # Provenance.
    prompt_version: Mapped[str | None] = mapped_column(Text, nullable=True)
    model_identifier: Mapped[str | None] = mapped_column(Text, nullable=True)
    output_schema_version: Mapped[str | None] = mapped_column(Text, nullable=True)

    # Failure.
    error_code: Mapped[str | None] = mapped_column(Text, nullable=True)
    error_message_pt_br: Mapped[str | None] = mapped_column(Text, nullable=True)
    quota_consumed: Mapped[bool] = mapped_column(Boolean, nullable=False, server_default="true")

    # Lineage.
    parent_correction_id: Mapped[uuid.UUID | None] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("corrections.id", ondelete="SET NULL"),
        nullable=True,
    )

    __table_args__ = (
        CheckConstraint(
            f"c1_score IS NULL OR c1_score IN {_SCORE_SET}",
            name="corrections_c1_score_scale",
        ),
        CheckConstraint(
            f"c2_score IS NULL OR c2_score IN {_SCORE_SET}",
            name="corrections_c2_score_scale",
        ),
        CheckConstraint(
            f"c3_score IS NULL OR c3_score IN {_SCORE_SET}",
            name="corrections_c3_score_scale",
        ),
        CheckConstraint(
            f"c4_score IS NULL OR c4_score IN {_SCORE_SET}",
            name="corrections_c4_score_scale",
        ),
        CheckConstraint(
            f"c5_score IS NULL OR c5_score IN {_SCORE_SET}",
            name="corrections_c5_score_scale",
        ),
        CheckConstraint(
            "final_score IS NULL OR (final_score BETWEEN 0 AND 1000)",
            name="corrections_final_score_range",
        ),
        CheckConstraint(
            "status <> 'completed' OR final_score = c1_score + c2_score + c3_score + c4_score + c5_score",
            name="corrections_final_score_sum",
        ),
        CheckConstraint(
            "status <> 'completed' OR (prompt_version IS NOT NULL AND model_identifier IS NOT NULL AND output_schema_version IS NOT NULL)",
            name="corrections_completed_has_provenance",
        ),
        CheckConstraint(
            "status <> 'failed' OR error_code IS NOT NULL",
            name="corrections_failed_has_error",
        ),
        Index(
            "corrections_queue_idx",
            "queued_at",
            postgresql_where=(status == CorrectionStatus.pending),
        ),
        Index("corrections_user_queued_idx", "user_id", "queued_at"),
        Index(
            "corrections_user_listing_idx",
            "user_id",
            "queued_at",
            postgresql_using="btree",
            postgresql_ops={"queued_at": "DESC"},
        ),
        Index(
            "corrections_parent_idx",
            "parent_correction_id",
            postgresql_where=(parent_correction_id.is_not(None)),
        ),
        Index("corrections_job_idx", "job_id"),
    )
