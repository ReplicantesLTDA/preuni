"""T098: LISTEN/NOTIFY wakes worker; poll fallback still claims.

Constitution v2.1.1 / specs/014-constitution-alignment-refactor: the NOTIFY
now fires from a DB trigger on `correction.correction_jobs` INSERT
(002_correction_jobs_schema migration) rather than from a Python enqueue
function, since the Go monolith is the one inserting rows.
"""

from __future__ import annotations

import asyncio
import os
import pytest
import uuid
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from src.db.models import CorrectionJob
from src.db.models.enums import CorrectionJobStatus, CorrectionStatus
from src.db.repositories.correction_repo import claim_next

DEFAULT_TEST_DB_URL = "postgresql+asyncpg://corretor:corretor@localhost:5432/corretor_test"


def _test_db_url() -> str:
    return os.environ.get("TEST_DATABASE_URL", DEFAULT_TEST_DB_URL)


async def _insert_pending_job(
    session: AsyncSession, essay: str = "texto de teste"
) -> CorrectionJob:
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
async def test_insert_emits_notify(db_session: AsyncSession) -> None:
    """INSERT into correction_jobs triggers pg_notify('correction_queued', ...) via the DB trigger."""
    import asyncpg

    url = _test_db_url()
    asyncpg_url = url.replace("postgresql+asyncpg://", "postgresql://")

    notifications: list[str] = []

    conn = await asyncpg.connect(asyncpg_url)
    try:
        await conn.add_listener(
            "correction_queued",
            lambda *args: notifications.append(str(args[-1])),
        )

        job = await _insert_pending_job(db_session)
        await db_session.commit()

        deadline = asyncio.get_event_loop().time() + 1.0
        while not notifications and asyncio.get_event_loop().time() < deadline:
            await asyncio.sleep(0.05)
            await conn.execute("SELECT 1")

    finally:
        await conn.close()

    assert len(notifications) >= 1
    assert str(job.id) in notifications[0]


@pytest.mark.asyncio
async def test_poll_fallback_claims_without_notify(db_session: AsyncSession) -> None:
    """claim_next() works even when the LISTEN mechanism is not active (poll path)."""
    from sqlalchemy import text

    await db_session.execute(
        text("DELETE FROM correction.correction_jobs WHERE status = 'pending'")
    )
    await db_session.commit()
    job = await _insert_pending_job(db_session)
    await db_session.commit()

    engine = create_async_engine(_test_db_url(), echo=False)
    try:
        session_factory = async_sessionmaker(engine, expire_on_commit=False)
        async with session_factory() as s:
            claimed = await claim_next(s, worker_id="poll-worker")
            assert claimed is not None
            assert claimed.job_id == job.id
            assert claimed.status == CorrectionStatus.processing
            await s.commit()
    finally:
        await engine.dispose()
