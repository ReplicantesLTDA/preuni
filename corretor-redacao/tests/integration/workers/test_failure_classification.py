"""T100: failure classification — quota_consumed per error_codes.md FR-036.

Constitution v2.1.1 / specs/014-constitution-alignment-refactor: quota
itself is enforced by the Go monolith now, but this service still reports
whether a failure is provider/internal (quota_consumed=false) or
user-attributable (quota_consumed=true) so Go's reconciler can decide
whether to refund the day's submission.
"""

from __future__ import annotations

import os
import pytest
import uuid
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from src.db.models import Correction, CorrectionJob
from src.db.models.enums import CorrectionJobStatus, CorrectionStatus
from src.db.repositories.correction_repo import claim_next, mark_failed
from src.workers.failure_classifier import PROVIDER_ERROR_CODES, USER_ERROR_CODES, classify_failure


async def _insert_pending_job(session: AsyncSession, n: int = 0) -> CorrectionJob:
    job = CorrectionJob(
        id=uuid.uuid4(),
        user_id=uuid.uuid4(),
        essay_text=f"essay text {n}",
        prompt_theme_title="Tema",
        prompt_theme_context="Contexto do tema da redação.",
        status=CorrectionJobStatus.pending,
    )
    session.add(job)
    await session.flush()
    return job


@pytest.mark.parametrize(
    "error_code",
    [
        "provider_rate_limited",
        "provider_timeout",
        "provider_unavailable",
        "schema_violation",
        "internal_error",
    ],
)
def test_provider_errors_not_quota_consumed(error_code: str) -> None:
    quota_consumed = classify_failure(error_code)
    assert quota_consumed is False, f"{error_code} should not consume quota"


@pytest.mark.parametrize(
    "error_code",
    [
        "language_mismatch",
        "theme_missing_context",
        "length_too_short",
        "length_too_long",
    ],
)
def test_user_errors_quota_consumed(error_code: str) -> None:
    quota_consumed = classify_failure(error_code)
    assert quota_consumed is True, f"{error_code} should consume quota"


def test_provider_error_codes_set() -> None:
    assert "provider_rate_limited" in PROVIDER_ERROR_CODES
    assert "schema_violation" in PROVIDER_ERROR_CODES
    assert "internal_error" in PROVIDER_ERROR_CODES


def test_user_error_codes_set() -> None:
    assert "language_mismatch" in USER_ERROR_CODES
    assert "length_too_short" in USER_ERROR_CODES


@pytest.mark.asyncio
async def test_mark_failed_provider_error_quota_false(db_session: AsyncSession) -> None:
    job = await _insert_pending_job(db_session)
    await db_session.commit()

    _db = os.environ.get(
        "TEST_DATABASE_URL", "postgresql+asyncpg://corretor:corretor@localhost:5432/corretor_test"
    )
    engine = create_async_engine(_db, echo=False)
    try:
        async with async_sessionmaker(engine, expire_on_commit=False)() as s:
            claimed = await claim_next(s, worker_id="w")
            assert claimed is not None
            await mark_failed(
                s,
                correction=claimed,
                error_code="schema_violation",
                error_message_pt_br="O modelo não produziu saída válida após duas tentativas.",
                quota_consumed=False,
            )
            await s.commit()
            claimed_id = claimed.id
    finally:
        await engine.dispose()

    row = (
        await db_session.execute(select(Correction).where(Correction.id == claimed_id))
    ).scalar_one()
    assert row.status == CorrectionStatus.failed
    assert row.quota_consumed is False
    assert row.error_code == "schema_violation"

    job_row = (
        await db_session.execute(select(CorrectionJob).where(CorrectionJob.id == job.id))
    ).scalar_one()
    assert job_row.status == CorrectionJobStatus.failed


@pytest.mark.asyncio
async def test_mark_failed_user_error_quota_true(db_session: AsyncSession) -> None:
    job = await _insert_pending_job(db_session, n=1)
    await db_session.commit()

    _db = os.environ.get(
        "TEST_DATABASE_URL", "postgresql+asyncpg://corretor:corretor@localhost:5432/corretor_test"
    )
    engine = create_async_engine(_db, echo=False)
    try:
        async with async_sessionmaker(engine, expire_on_commit=False)() as s:
            claimed = await claim_next(s, worker_id="w")
            assert claimed is not None
            await mark_failed(
                s,
                correction=claimed,
                error_code="language_mismatch",
                error_message_pt_br="A redação não está em português brasileiro.",
                quota_consumed=True,
            )
            await s.commit()
            claimed_id = claimed.id
    finally:
        await engine.dispose()

    row = (
        await db_session.execute(select(Correction).where(Correction.id == claimed_id))
    ).scalar_one()
    assert row.status == CorrectionStatus.failed
    assert row.quota_consumed is True
    assert job.id is not None
