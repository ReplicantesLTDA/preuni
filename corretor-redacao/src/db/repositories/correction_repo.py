"""CorrectionRepo — claim-from-bridge-table + result persistence.

Constitution v2.1.1 / specs/014-constitution-alignment-refactor: quota
enforcement and enqueueing now live entirely in the Go monolith
(research.md #2; contracts/internal-bridge.md). This service no longer
accepts end-user submissions — it claims rows the monolith already wrote
to `correction_jobs`, grades them, and writes the result to `corrections`.
"""

from __future__ import annotations

import datetime as dt
import hashlib
import uuid
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from src.db.models import Correction, CorrectionJob
from src.db.models.enums import CorrectionJobStatus, CorrectionStatus


def _input_hash(essay_text: str, prompt_theme_title: str, prompt_theme_context: str) -> bytes:
    content = f"{essay_text}\x00{prompt_theme_title}\x00{prompt_theme_context}".encode()
    return hashlib.sha256(content).digest()


async def claim_next(
    session: AsyncSession,
    *,
    worker_id: str,
) -> Correction | None:
    """Claim the oldest pending job with SELECT ... FOR UPDATE SKIP LOCKED.

    Marks the `correction_jobs` row `processing`, creates the matching
    `corrections` row (also `processing`), and returns that `Correction` —
    the rest of the worker pipeline is unchanged from before this refactor,
    it just no longer finds its row pre-created by an end-user request.
    """
    result = await session.execute(
        select(CorrectionJob)
        .where(CorrectionJob.status == CorrectionJobStatus.pending)
        .order_by(CorrectionJob.queued_at)
        .limit(1)
        .with_for_update(skip_locked=True)
    )
    job = result.scalar_one_or_none()
    if job is None:
        return None

    now = dt.datetime.now(dt.UTC)
    job.status = CorrectionJobStatus.processing
    job.started_at = now

    correction = Correction(
        id=uuid.uuid4(),
        user_id=job.user_id,
        job_id=job.id,
        essay_text=job.essay_text,
        prompt_theme_title=job.prompt_theme_title,
        prompt_theme_context=job.prompt_theme_context,
        input_hash=_input_hash(job.essay_text, job.prompt_theme_title, job.prompt_theme_context),
        status=CorrectionStatus.processing,
        started_at=now,
        locked_at=now,
        locked_by=worker_id,
    )
    session.add(correction)
    await session.flush()
    return correction


async def _mark_job(
    session: AsyncSession,
    *,
    job_id: uuid.UUID | None,
    status: CorrectionJobStatus,
) -> None:
    if job_id is None:
        return
    job = await session.get(CorrectionJob, job_id)
    if job is None:
        return
    job.status = status
    job.completed_at = dt.datetime.now(dt.UTC)


async def mark_completed(
    session: AsyncSession,
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
    await _mark_job(session, job_id=correction.job_id, status=CorrectionJobStatus.completed)


async def mark_failed(
    session: AsyncSession,
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
    await _mark_job(session, job_id=correction.job_id, status=CorrectionJobStatus.failed)


__all__ = ["claim_next", "mark_completed", "mark_failed"]
