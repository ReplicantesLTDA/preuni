"""CorrectionRepo — queue operations with quota enforcement.

Uses SERIALIZABLE isolation for quota check + enqueue to prevent concurrent
submissions from racing past the monthly quota boundary (research R2).
"""

from __future__ import annotations

import datetime as dt
import hashlib
import uuid
from sqlalchemy import select, text
from sqlalchemy.ext.asyncio import AsyncSession

from src.auth.quota import check_quota
from src.db.models import Correction, User
from src.db.models.enums import CorrectionStatus


def _input_hash(
    essay_text: str,
    prompt_theme_title: str,
    prompt_theme_context: str,
    motivational_texts: str | None,
) -> bytes:
    content = (
        f"{essay_text}\x00{prompt_theme_title}\x00"
        f"{prompt_theme_context}\x00{motivational_texts or ''}"
    ).encode()
    return hashlib.sha256(content).digest()


async def enqueue_with_quota_check(
    session: AsyncSession,
    *,
    user: User,
    essay_text: str,
    prompt_theme_title: str,
    prompt_theme_context: str,
    motivational_texts: str | None = None,
) -> Correction:
    """Quota check + INSERT in a SERIALIZABLE transaction.

    Raises QuotaExhaustedError when the user is at/over their monthly limit.
    The serializable isolation prevents two concurrent submissions at the
    boundary from both succeeding.
    """
    await check_quota(session, user)

    correction = Correction(
        id=uuid.uuid4(),
        user_id=user.id,
        essay_text=essay_text,
        prompt_theme_title=prompt_theme_title,
        prompt_theme_context=prompt_theme_context,
        motivational_texts=motivational_texts,
        input_hash=_input_hash(
            essay_text, prompt_theme_title, prompt_theme_context, motivational_texts
        ),
        status=CorrectionStatus.pending,
    )
    session.add(correction)
    await session.flush()

    await session.execute(
        text("SELECT pg_notify('correction_queued', :cid)").bindparams(cid=str(correction.id))
    )
    return correction


async def get_for_user(
    session: AsyncSession,
    *,
    correction_id: uuid.UUID,
    user_id: uuid.UUID,
) -> Correction | None:
    """Return a correction owned by user_id, or None (indistinguishable from not-found)."""
    result = await session.execute(
        select(Correction)
        .where(Correction.id == correction_id)
        .where(Correction.user_id == user_id)
    )
    return result.scalar_one_or_none()


async def claim_next(
    session: AsyncSession,
    *,
    worker_id: str,
) -> Correction | None:
    """Claim the oldest pending correction with SELECT … FOR UPDATE SKIP LOCKED.

    Returns the claimed row (now status='processing') or None if the queue is empty.
    The caller is responsible for committing the transaction.
    """
    result = await session.execute(
        select(Correction)
        .where(Correction.status == CorrectionStatus.pending)
        .order_by(Correction.queued_at)
        .limit(1)
        .with_for_update(skip_locked=True)
    )
    correction = result.scalar_one_or_none()
    if correction is None:
        return None

    now = dt.datetime.now(dt.UTC)
    correction.status = CorrectionStatus.processing
    correction.started_at = now
    correction.locked_at = now
    correction.locked_by = worker_id
    return correction


async def mark_completed(
    session: AsyncSession,  # noqa: ARG001
    *,
    correction: Correction,
    final_score: int,
    c1_score: int,
    c2_score: int,
    c3_score: int,
    c4_score: int,
    c5_score: int,
    competencies: dict,
    eliminatory_flags: list,
    prompt_version: str,
    model_identifier: str,
    output_schema_version: str,
) -> None:
    """Advance correction to completed, writing all aggregate fields."""
    now = dt.datetime.now(dt.UTC)
    correction.status = CorrectionStatus.completed
    correction.completed_at = now
    correction.final_score = final_score
    correction.c1_score = c1_score
    correction.c2_score = c2_score
    correction.c3_score = c3_score
    correction.c4_score = c4_score
    correction.c5_score = c5_score
    correction.competencies = competencies
    correction.eliminatory_flags = eliminatory_flags
    correction.prompt_version = prompt_version
    correction.model_identifier = model_identifier
    correction.output_schema_version = output_schema_version
    correction.quota_consumed = True


async def mark_failed(
    session: AsyncSession,  # noqa: ARG001
    *,
    correction: Correction,
    error_code: str,
    error_message_pt_br: str,
    quota_consumed: bool,
) -> None:
    """Advance correction to failed state."""
    now = dt.datetime.now(dt.UTC)
    correction.status = CorrectionStatus.failed
    correction.completed_at = now
    correction.error_code = error_code
    correction.error_message_pt_br = error_message_pt_br
    correction.quota_consumed = quota_consumed


__all__ = [
    "claim_next",
    "enqueue_with_quota_check",
    "get_for_user",
    "mark_completed",
    "mark_failed",
]
