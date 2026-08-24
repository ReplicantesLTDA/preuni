"""CorrectionWorker.run()'s actual run loop -- previously untested (60%
coverage, missing the run/_run_loop/_listen_task machinery entirely).

This starts a real worker (real asyncpg LISTEN connection, real poll
timer) against the test database, seeds a pending correction_jobs row,
lets the worker's poll backstop pick it up and grade it via FakeProvider,
then stops the worker and asserts the correction actually completed --
not just that lines executed.
"""

from __future__ import annotations

import asyncio
import os
import pytest
import uuid
from sqlalchemy import select, text
from sqlalchemy.ext.asyncio import AsyncSession

from src.corrector.llm.fake import FakeProvider
from src.db.models import Correction, CorrectionJob
from src.db.models.enums import CorrectionJobStatus, CorrectionStatus
from src.workers.correction_worker import CorrectionWorker

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
            "improvement_path_pt_br": "revisar.",
        }
        for code in ("c1", "c2", "c3", "c4", "c5")
    },
}


def _db_url() -> str:
    return os.environ.get(
        "TEST_DATABASE_URL", "postgresql+asyncpg://corretor:corretor@localhost:5432/corretor_test"
    )


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


@pytest.mark.asyncio
async def test_worker_run_polls_and_grades_a_pending_job(db_session: AsyncSession) -> None:
    await db_session.execute(
        text("DELETE FROM correction.correction_jobs WHERE status = 'pending'")
    )
    await db_session.commit()
    job = await _insert_pending_job(db_session)
    await db_session.commit()
    job_id = job.id

    fake = FakeProvider(responses=[_PERFECT_OUTPUT])
    worker = CorrectionWorker(
        database_url=_db_url(),
        provider=fake,
        worker_id="run-loop-test-worker",
        poll_interval_s=0.1,
    )

    run_task = asyncio.create_task(worker.run())
    try:
        deadline = asyncio.get_event_loop().time() + 5.0
        correction = None
        while asyncio.get_event_loop().time() < deadline:
            result = await db_session.execute(
                select(Correction).where(Correction.job_id == job_id)
            )
            correction = result.scalar_one_or_none()
            if correction is not None and correction.status == CorrectionStatus.completed:
                break
            await asyncio.sleep(0.1)
            db_session.expire_all()

        assert correction is not None, "worker never claimed the pending job"
        assert correction.status == CorrectionStatus.completed
        assert correction.final_score == 800
    finally:
        worker.stop()
        await asyncio.wait_for(run_task, timeout=5.0)


@pytest.mark.asyncio
async def test_worker_run_stops_cleanly_with_no_pending_work() -> None:
    """run()/_run_loop()/_listen_task's shutdown path (listen_task.cancel(),
    engine.dispose()) with nothing to process."""
    fake = FakeProvider(responses=[])
    worker = CorrectionWorker(
        database_url=_db_url(),
        provider=fake,
        worker_id="run-loop-idle-worker",
        poll_interval_s=0.1,
    )

    run_task = asyncio.create_task(worker.run())
    await asyncio.sleep(0.3)
    worker.stop()
    await asyncio.wait_for(run_task, timeout=5.0)

    assert run_task.done()
    assert run_task.exception() is None
