"""T096: correction_repo.claim_next — SELECT FOR UPDATE SKIP LOCKED."""

from __future__ import annotations

import datetime as dt
import hashlib
import pytest
import uuid
from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncSession

from src.db.models import Correction, User
from src.db.models.enums import CorrectionStatus, UserTier
from src.db.repositories.correction_repo import claim_next


def _input_hash(essay: str) -> bytes:
    return hashlib.sha256(essay.encode()).digest()


async def _clear_pending(session: AsyncSession) -> None:
    """Remove pending corrections left by previous tests (tests share a DB)."""
    await session.execute(text("DELETE FROM corrections WHERE status = 'pending'"))
    await session.commit()


async def _insert_pending(session: AsyncSession, user: User, essay: str = "essay") -> Correction:
    c = Correction(
        id=uuid.uuid4(),
        user_id=user.id,
        essay_text=essay,
        prompt_theme_title="Tema",
        prompt_theme_context="Contexto do tema da redação.",
        input_hash=_input_hash(essay),
        status=CorrectionStatus.pending,
    )
    session.add(c)
    await session.flush()
    return c


async def _insert_user(session: AsyncSession) -> User:
    u = User(
        id=uuid.uuid4(),
        email=f"worker-{uuid.uuid4()}@example.com",
        password_hash="x",
        tier=UserTier.free,
    )
    session.add(u)
    await session.flush()
    return u


@pytest.mark.asyncio
async def test_claim_advances_status_to_processing(db_session: AsyncSession) -> None:
    await _clear_pending(db_session)
    user = await _insert_user(db_session)
    correction = await _insert_pending(db_session, user)
    await db_session.commit()

    async with db_session.bind.connect() as conn, conn.begin():
        session2 = AsyncSession(bind=conn)
        claimed = await claim_next(session2, worker_id="worker-1")
        assert claimed is not None
        assert claimed.id == correction.id
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
async def test_claim_skips_processing_and_completed(db_session: AsyncSession) -> None:
    await _clear_pending(db_session)
    user = await _insert_user(db_session)
    processing = await _insert_pending(db_session, user, "essay-proc")
    processing.status = CorrectionStatus.processing
    completed = await _insert_pending(db_session, user, "essay-comp")
    completed.status = CorrectionStatus.completed
    completed.final_score = 800
    completed.c1_score = 160
    completed.c2_score = 160
    completed.c3_score = 160
    completed.c4_score = 160
    completed.c5_score = 160
    completed.prompt_version = "1.0.0"
    completed.model_identifier = "kimi-k2:1t"
    completed.output_schema_version = "1"
    await db_session.commit()

    claimed = await claim_next(db_session, worker_id="worker-1")
    assert claimed is None


@pytest.mark.asyncio
async def test_claim_oldest_first(db_session: AsyncSession) -> None:
    await _clear_pending(db_session)
    user = await _insert_user(db_session)

    older = await _insert_pending(db_session, user, "essay-old")
    older.queued_at = dt.datetime(2026, 1, 1, tzinfo=dt.UTC)
    newer = await _insert_pending(db_session, user, "essay-new")
    newer.queued_at = dt.datetime(2026, 1, 2, tzinfo=dt.UTC)
    await db_session.commit()

    async with db_session.bind.connect() as conn, conn.begin():
        s = AsyncSession(bind=conn)
        claimed = await claim_next(s, worker_id="w")
        assert claimed is not None
        assert claimed.id == older.id
        await s.close()
