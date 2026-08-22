"""T102: completed correction row satisfies all invariants."""

from __future__ import annotations

import hashlib
import os
import pytest
import uuid
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from src.db.models import Correction, CorrectionAuditLog, User
from src.db.models.enums import AuditEventType, CorrectionStatus, UserTier
from src.db.repositories.audit_log_repo import AuditLogWriter
from src.db.repositories.correction_repo import claim_next, mark_completed

_DB_URL = os.environ.get(
    "TEST_DATABASE_URL",
    "postgresql+asyncpg://corretor:corretor@localhost:5432/corretor_test",
)


async def _worker_session():
    engine = create_async_engine(_DB_URL, echo=False)
    factory = async_sessionmaker(engine, expire_on_commit=False)
    try:
        yield factory
    finally:
        await engine.dispose()


async def _insert_user(session: AsyncSession) -> User:
    u = User(
        id=uuid.uuid4(),
        email=f"inv-{uuid.uuid4()}@example.com",
        password_hash="x",
        tier=UserTier.free,
    )
    session.add(u)
    await session.flush()
    return u


async def _insert_pending(session: AsyncSession, user: User) -> Correction:
    text = "essay invariant test"
    c = Correction(
        id=uuid.uuid4(),
        user_id=user.id,
        essay_text=text,
        prompt_theme_title="Tema",
        prompt_theme_context="Contexto do tema da redação.",
        input_hash=hashlib.sha256(text.encode()).digest(),
        status=CorrectionStatus.pending,
    )
    session.add(c)
    await session.flush()
    return c


COMPETENCIES = {
    "c1": {"score": 160, "excerpt": "essay", "justification_pt_br": "Bom."},
    "c2": {"score": 160, "excerpt": "essay", "justification_pt_br": "Bom."},
    "c3": {"score": 160, "excerpt": "essay", "justification_pt_br": "Bom."},
    "c4": {"score": 160, "excerpt": "essay", "justification_pt_br": "Bom."},
    "c5": {"score": 160, "excerpt": "essay", "justification_pt_br": "Bom."},
}


@pytest.mark.asyncio
async def test_completed_final_score_equals_sum(db_session: AsyncSession) -> None:
    user = await _insert_user(db_session)
    correction = await _insert_pending(db_session, user)
    correction_id = correction.id
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
    user = await _insert_user(db_session)
    correction = await _insert_pending(db_session, user)
    correction_id = correction.id
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
    user = await _insert_user(db_session)
    correction = await _insert_pending(db_session, user)
    correction_id = correction.id
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
