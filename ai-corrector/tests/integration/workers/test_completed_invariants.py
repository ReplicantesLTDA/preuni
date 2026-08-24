"""T102: completed correction row satisfies all invariants.

Constitution v2.1.1 / specs/014-constitution-alignment-refactor: seeds a
`correction_jobs` row (as Go would) instead of a `User` + pending
`Correction` row.
"""

from __future__ import annotations

import os
import pytest
import uuid
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from src.db.models import Correction, CorrectionAuditLog, CorrectionJob
from src.db.models.enums import AuditEventType, CorrectionJobStatus, CorrectionStatus
from src.db.repositories.audit_log_repo import AuditLogWriter
from src.db.repositories.correction_repo import claim_next, mark_completed

_DB_URL = os.environ.get(
    "TEST_DATABASE_URL",
    "postgresql+asyncpg://corretor:corretor@localhost:5432/corretor_test",
)


async def _insert_pending_job(session: AsyncSession) -> CorrectionJob:
    job = CorrectionJob(
        id=uuid.uuid4(),
        user_id=uuid.uuid4(),
        essay_text="essay invariant test",
        prompt_theme_title="Tema",
        prompt_theme_context="Contexto do tema da redação.",
        status=CorrectionJobStatus.pending,
    )
    session.add(job)
    await session.flush()
    return job


COMPETENCIES = {
    "c1": {"score": 160, "excerpt": "essay", "justification_pt_br": "Bom."},
    "c2": {"score": 160, "excerpt": "essay", "justification_pt_br": "Bom."},
    "c3": {"score": 160, "excerpt": "essay", "justification_pt_br": "Bom."},
    "c4": {"score": 160, "excerpt": "essay", "justification_pt_br": "Bom."},
    "c5": {"score": 160, "excerpt": "essay", "justification_pt_br": "Bom."},
}


@pytest.mark.asyncio
async def test_completed_final_score_equals_sum(db_session: AsyncSession) -> None:
    await _insert_pending_job(db_session)
    await db_session.commit()

    engine = create_async_engine(_DB_URL, echo=False)
    try:
        factory = async_sessionmaker(engine, expire_on_commit=False)
        async with factory() as s:
            claimed = await claim_next(s, worker_id="w")
            assert claimed is not None
            await mark_completed(
                s,
                correction=claimed,
                final_score=800,
                c1_score=160,
                c2_score=160,
                c3_score=160,
                c4_score=160,
                c5_score=160,
                competencies=COMPETENCIES,
                eliminatory_flags=[],
                prompt_version="1.0.0",
                model_identifier="kimi-k2:1t",
                output_schema_version="1",
            )
            await s.commit()
            correction_id = claimed.id
    finally:
        await engine.dispose()

    row = (
        await db_session.execute(select(Correction).where(Correction.id == correction_id))
    ).scalar_one()
    assert row.status == CorrectionStatus.completed
    assert (
        row.final_score == row.c1_score + row.c2_score + row.c3_score + row.c4_score + row.c5_score
    )


@pytest.mark.asyncio
async def test_completed_has_provenance(db_session: AsyncSession) -> None:
    await _insert_pending_job(db_session)
    await db_session.commit()

    engine = create_async_engine(_DB_URL, echo=False)
    try:
        factory = async_sessionmaker(engine, expire_on_commit=False)
        async with factory() as s:
            claimed = await claim_next(s, worker_id="w")
            assert claimed is not None
            await mark_completed(
                s,
                correction=claimed,
                final_score=800,
                c1_score=160,
                c2_score=160,
                c3_score=160,
                c4_score=160,
                c5_score=160,
                competencies=COMPETENCIES,
                eliminatory_flags=[],
                prompt_version="1.0.7",
                model_identifier="kimi-k2:1t",
                output_schema_version="v1",
            )
            await s.commit()
            correction_id = claimed.id
    finally:
        await engine.dispose()

    row = (
        await db_session.execute(select(Correction).where(Correction.id == correction_id))
    ).scalar_one()
    assert row.prompt_version == "1.0.7"
    assert row.model_identifier == "kimi-k2:1t"
    assert row.output_schema_version == "v1"


@pytest.mark.asyncio
async def test_completed_audit_log_event(db_session: AsyncSession) -> None:
    await _insert_pending_job(db_session)
    await db_session.commit()

    engine = create_async_engine(_DB_URL, echo=False)
    try:
        factory = async_sessionmaker(engine, expire_on_commit=False)
        async with factory() as s:
            claimed = await claim_next(s, worker_id="w")
            assert claimed is not None
            await mark_completed(
                s,
                correction=claimed,
                final_score=800,
                c1_score=160,
                c2_score=160,
                c3_score=160,
                c4_score=160,
                c5_score=160,
                competencies=COMPETENCIES,
                eliminatory_flags=[],
                prompt_version="1.0.0",
                model_identifier="kimi-k2:1t",
                output_schema_version="v1",
            )
            writer = AuditLogWriter(s)
            await writer.write(
                correction_id=claimed.id,
                input_hash=claimed.input_hash,
                event_type=AuditEventType.completed,
                payload={
                    "final_score": 800,
                    "eliminatory_flags": [],
                    "prompt_version": "1.0.0",
                    "model_identifier": "kimi-k2:1t",
                },
            )
            await s.commit()
            correction_id = claimed.id
    finally:
        await engine.dispose()

    logs = (
        (
            await db_session.execute(
                select(CorrectionAuditLog).where(CorrectionAuditLog.correction_id == correction_id)
            )
        )
        .scalars()
        .all()
    )
    assert any(log.event_type == AuditEventType.completed for log in logs)
