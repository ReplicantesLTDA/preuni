"""correction_repo._mark_job's two early-return branches (no job_id / a
job_id that doesn't exist) -- mark_completed/mark_failed always pass a real
job_id from a real row, so these edge branches were never hit."""

from __future__ import annotations

import pytest
import uuid
from sqlalchemy.ext.asyncio import AsyncSession

from src.db.models.enums import CorrectionJobStatus
from src.db.repositories.correction_repo import _mark_job


@pytest.mark.asyncio
async def test_mark_job_is_a_noop_when_job_id_is_none(db_session: AsyncSession) -> None:
    await _mark_job(db_session, job_id=None, status=CorrectionJobStatus.completed)


@pytest.mark.asyncio
async def test_mark_job_is_a_noop_when_the_job_does_not_exist(db_session: AsyncSession) -> None:
    await _mark_job(db_session, job_id=uuid.uuid4(), status=CorrectionJobStatus.completed)
