"""Worker-level test: process_one() claims a real correction_jobs row and
grades it end to end via FakeProvider -- exercises the orchestration in
src/workers/correction_worker.py that claim/mark_completed unit tests
alone don't reach (audit writes, grader_pass persistence, competency dict
building)."""

from __future__ import annotations

import os
import pytest
import uuid
from sqlalchemy import select, text
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from src.corrector.llm.errors import SchemaViolationError
from src.corrector.llm.fake import FakeProvider
from src.db.models import Correction, CorrectionJob, GraderPass
from src.db.models.enums import CorrectionJobStatus, CorrectionStatus
from src.workers.correction_worker import process_one

ESSAY_TEXT = (
    "A inclusão digital de pessoas idosas demanda ação coordenada do Estado.\n"
    "Você pode notar, no cotidiano, que muitos avós têm dificuldade para utilizar\n"
    "aplicativos de banco e plataformas governamentais essenciais para a cidadania.\n"
    "Diante disso, propõe-se uma estratégia coordenada de inclusão, com cursos\n"
    "gratuitos em telecentros e parcerias com escolas técnicas para mediação\n"
    "digital intergeracional efetiva entre famílias brasileiras hoje em dia.\n"
    "Portanto, cabe ao MEC articular um programa nacional de letramento digital."
)

_PERFECT_OUTPUT = {
    "eliminatory_flags": [],
    "final_score": 800,
    "competencies": {
        code: {
            "score": 160,
            "excerpt": "A inclusão digital de pessoas idosas",
            "justification_pt_br": "ok",
            "improvement_path_pt_br": "revisar.",  # required by the schema when score < 200
        }
        for code in ("c1", "c2", "c3", "c4", "c5")
    },
}


async def _insert_pending_job(session: AsyncSession) -> CorrectionJob:
    job = CorrectionJob(
        id=uuid.uuid4(),
        user_id=uuid.uuid4(),
        essay_text=ESSAY_TEXT,
        prompt_theme_title="Inclusão digital de idosos no Brasil",
        prompt_theme_context="A democratização das ferramentas digitais não atingiu uniformemente a população idosa.",
        status=CorrectionJobStatus.pending,
    )
    session.add(job)
    await session.flush()
    return job


def _db_url() -> str:
    return os.environ.get(
        "TEST_DATABASE_URL", "postgresql+asyncpg://corretor:corretor@localhost:5432/corretor_test"
    )


@pytest.mark.asyncio
async def test_process_one_grades_a_real_job_end_to_end(db_session: AsyncSession) -> None:
    await db_session.execute(
        text("DELETE FROM correction.correction_jobs WHERE status = 'pending'")
    )
    await db_session.commit()

    job = await _insert_pending_job(db_session)
    await db_session.commit()

    fake = FakeProvider(responses=[_PERFECT_OUTPUT])
    engine = create_async_engine(_db_url(), echo=False)
    try:
        async with async_sessionmaker(engine, expire_on_commit=False)() as s:
            processed = await process_one(s, provider=fake, worker_id="test-worker")
            assert processed is True
            await s.commit()
    finally:
        await engine.dispose()

    correction = (
        await db_session.execute(select(Correction).where(Correction.job_id == job.id))
    ).scalar_one()
    assert correction.status == CorrectionStatus.completed
    assert correction.final_score == 800
    assert correction.c1_score == 160

    grader_pass = (
        await db_session.execute(
            select(GraderPass).where(GraderPass.correction_id == correction.id)
        )
    ).scalar_one()
    assert grader_pass.c1_score == 160

    job_row = (
        await db_session.execute(
            select(CorrectionJob)
            .where(CorrectionJob.id == job.id)
            .execution_options(populate_existing=True)
        )
    ).scalar_one()
    assert job_row.status == CorrectionJobStatus.completed


@pytest.mark.asyncio
async def test_process_one_marks_failed_on_provider_error(db_session: AsyncSession) -> None:
    await db_session.execute(
        text("DELETE FROM correction.correction_jobs WHERE status = 'pending'")
    )
    await db_session.commit()

    job = await _insert_pending_job(db_session)
    await db_session.commit()

    fake = FakeProvider(
        responses=[SchemaViolationError("bad output"), SchemaViolationError("bad output")]
    )
    engine = create_async_engine(_db_url(), echo=False)
    try:
        async with async_sessionmaker(engine, expire_on_commit=False)() as s:
            processed = await process_one(s, provider=fake, worker_id="test-worker")
            assert processed is True
            await s.commit()
    finally:
        await engine.dispose()

    correction = (
        await db_session.execute(select(Correction).where(Correction.job_id == job.id))
    ).scalar_one()
    assert correction.status == CorrectionStatus.failed
    assert correction.error_code == "schema_violation"
    assert correction.quota_consumed is False


@pytest.mark.asyncio
async def test_process_one_returns_false_when_queue_empty(db_session: AsyncSession) -> None:
    await db_session.execute(
        text("DELETE FROM correction.correction_jobs WHERE status = 'pending'")
    )
    await db_session.commit()

    fake = FakeProvider(responses=[])
    processed = await process_one(db_session, provider=fake, worker_id="test-worker")
    assert processed is False
