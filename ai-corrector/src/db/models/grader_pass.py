"""grader_passes table — multi-grader-ready (MVP uses pass_index=0)."""

from __future__ import annotations

import datetime as dt
import decimal
import uuid
from sqlalchemy import (
    BigInteger,
    CheckConstraint,
    DateTime,
    ForeignKey,
    Index,
    Integer,
    Numeric,
    SmallInteger,
    Text,
    UniqueConstraint,
    func,
)
from sqlalchemy.dialects.postgresql import JSONB, UUID
from sqlalchemy.orm import Mapped, mapped_column
from typing import Any

from src.db.models import Base

_SCORE_SET = "(0, 40, 80, 120, 160, 200)"


class GraderPass(Base):
    __tablename__ = "grader_passes"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True)
    correction_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("corrections.id", ondelete="CASCADE"),
        nullable=False,
    )
    pass_index: Mapped[int] = mapped_column(SmallInteger, nullable=False)

    c1_score: Mapped[int] = mapped_column(SmallInteger, nullable=False)
    c2_score: Mapped[int] = mapped_column(SmallInteger, nullable=False)
    c3_score: Mapped[int] = mapped_column(SmallInteger, nullable=False)
    c4_score: Mapped[int] = mapped_column(SmallInteger, nullable=False)
    c5_score: Mapped[int] = mapped_column(SmallInteger, nullable=False)
    competencies: Mapped[dict[str, Any]] = mapped_column(JSONB, nullable=False)
    eliminatory_flags: Mapped[list[Any]] = mapped_column(JSONB, nullable=False, server_default="[]")

    seed: Mapped[int | None] = mapped_column(BigInteger, nullable=True)
    prompt_version: Mapped[str] = mapped_column(Text, nullable=False)
    model_identifier: Mapped[str] = mapped_column(Text, nullable=False)
    output_schema_version: Mapped[str] = mapped_column(Text, nullable=False)
    inference_params: Mapped[dict[str, Any]] = mapped_column(JSONB, nullable=False)
    raw_output: Mapped[str] = mapped_column(Text, nullable=False)
    prompt_tokens: Mapped[int | None] = mapped_column(Integer, nullable=True)
    completion_tokens: Mapped[int | None] = mapped_column(Integer, nullable=True)
    latency_ms: Mapped[int] = mapped_column(Integer, nullable=False)
    cost_usd: Mapped[decimal.Decimal | None] = mapped_column(Numeric(10, 6), nullable=True)
    created_at: Mapped[dt.datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, server_default=func.now()
    )

    __table_args__ = (
        UniqueConstraint("correction_id", "pass_index", name="grader_passes_correction_pass_uq"),
        CheckConstraint(f"c1_score IN {_SCORE_SET}", name="grader_passes_c1_scale"),
        CheckConstraint(f"c2_score IN {_SCORE_SET}", name="grader_passes_c2_scale"),
        CheckConstraint(f"c3_score IN {_SCORE_SET}", name="grader_passes_c3_scale"),
        CheckConstraint(f"c4_score IN {_SCORE_SET}", name="grader_passes_c4_scale"),
        CheckConstraint(f"c5_score IN {_SCORE_SET}", name="grader_passes_c5_scale"),
        Index("grader_passes_correction_idx", "correction_id"),
    )
