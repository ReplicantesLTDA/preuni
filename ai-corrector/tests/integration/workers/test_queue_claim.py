"""T096: correction_repo.claim_next — claims oldest pending correction_jobs row.

Constitution v2.1.1 / specs/014-constitution-alignment-refactor: jobs are now
written by the Go monolith directly into `correction.correction_jobs`
(research.md #1); these tests seed that table the same way Go would,
instead of going through a Python-side enqueue function.
"""

from __future__ import annotations

import datetime as dt
import pytest
import uuid
from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncSession

from src.db.models import CorrectionJob
from src.db.models.enums import CorrectionJobStatus, CorrectionStatus
from src.db.repositories.correction_repo import claim_next


async def _clear_pending(session: AsyncSession) -> None:
    """Remove pending jobs left by previous tests (tests share a DB)."""
    await session.execute(text("DELETE FROM correction.correction_jobs WHERE status = 'pending'"))
    await session.commit()


async def _insert_pending_job(session: AsyncSession, essay: str = "essay") -> CorrectionJob:
    job = CorrectionJob(
        id=uuid.uuid4(),
        user_id=uuid.uuid4(),
        essay_text=essay,
        prompt_theme_title="Tema",
        prompt_theme_context="Contexto do tema da redação.",
        status=CorrectionJobStatus.pending,
    )
    session.add(job)
    await session.flush()
    return job


@pytest.mark.asyncio
async def test_claim_advances_status_to_processing(db_session: AsyncSession) -> None:
    await _clear_pending(db_session)
    job = await _insert_pending_job(db_session)
    await db_session.commit()

    async with db_session.bind.connect() as conn, conn.begin():
        session2 = AsyncSession(bind=conn)
        claimed = await claim_next(session2, worker_id="worker-1")
        assert claimed is not None
        assert claimed.job_id == job.id
        assert claimed.status == CorrectionStatus.processing
        assert claimed.locked_by == "worker-1"
        assert claimed.started_at is not None
        await session2.close()


@pytest.mark.asyncio
async def test_claim_returns_none_when_queue_empty(db_session: AsyncSession) -> None:
    await _clear_pending(db_session)
    claimed = await claim_next(db_session, worker_id="worker-1")
    assert claimed is None


@pytest.mark.asyncio
async def test_claim_skips_processing_and_completed_jobs(db_session: AsyncSession) -> None:
    await _clear_pending(db_session)
    processing = await _insert_pending_job(db_session, "essay-proc")
    processing.status = CorrectionJobStatus.processing
    completed = await _insert_pending_job(db_session, "essay-comp")
    completed.status = CorrectionJobStatus.completed
    await db_session.commit()

    claimed = await claim_next(db_session, worker_id="worker-1")
    assert claimed is None


@pytest.mark.asyncio
async def test_claim_oldest_first(db_session: AsyncSession) -> None:
    await _clear_pending(db_session)

    older = await _insert_pending_job(db_session, "essay-old")
    older.queued_at = dt.datetime(2026, 1, 1, tzinfo=dt.UTC)
    newer = await _insert_pending_job(db_session, "essay-new")
    newer.queued_at = dt.datetime(2026, 1, 2, tzinfo=dt.UTC)
    await db_session.commit()

    async with db_session.bind.connect() as conn, conn.begin():
        s = AsyncSession(bind=conn)
        claimed = await claim_next(s, worker_id="w")
        assert claimed is not None
        assert claimed.job_id == older.id
        await s.close()
