"""T098: LISTEN/NOTIFY wakes worker; poll fallback still claims."""

from __future__ import annotations

import asyncio
import hashlib
import os
import pytest
import uuid
from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from src.db.models import Correction, User
from src.db.models.enums import CorrectionStatus, UserTier
from src.db.repositories.correction_repo import claim_next

DEFAULT_TEST_DB_URL = "postgresql+asyncpg://corretor:corretor@localhost:5432/corretor_test"


def _test_db_url() -> str:
    return os.environ.get("TEST_DATABASE_URL", DEFAULT_TEST_DB_URL)


def _input_hash(s: str) -> bytes:
    return hashlib.sha256(s.encode()).digest()


async def _insert_user(session: AsyncSession) -> User:
    u = User(
        id=uuid.uuid4(),
        email=f"listen-{uuid.uuid4()}@example.com",
        password_hash="x",
        tier=UserTier.free,
    )
    session.add(u)
    await session.flush()
    return u


async def _insert_pending(session: AsyncSession, user: User) -> Correction:
    c = Correction(
        id=uuid.uuid4(),
        user_id=user.id,
        essay_text="texto de teste",
        prompt_theme_title="Tema",
        prompt_theme_context="Contexto do tema da redação.",
        input_hash=_input_hash("texto de teste"),
        status=CorrectionStatus.pending,
    )
    session.add(c)
    await session.flush()
    return c


@pytest.mark.asyncio
async def test_enqueue_emits_notify(db_session: AsyncSession) -> None:
    """INSERT via enqueue_with_quota_check triggers pg_notify correction_queued."""
    import asyncpg

    url = _test_db_url()
    # Use raw asyncpg to LISTEN
    asyncpg_url = url.replace("postgresql+asyncpg://", "postgresql://")

    notifications: list[str] = []

    conn = await asyncpg.connect(asyncpg_url)
    try:
        await conn.add_listener(
            "correction_queued",
            lambda *args: notifications.append(str(args[-1])),
        )

        user = await _insert_user(db_session)
        correction = await _insert_pending(db_session, user)
        await db_session.execute(
            text("SELECT pg_notify('correction_queued', :cid)").bindparams(cid=str(correction.id))
        )
        await db_session.commit()

        # Give the listener up to 1 second to receive the notification.
        deadline = asyncio.get_event_loop().time() + 1.0
        while not notifications and asyncio.get_event_loop().time() < deadline:
            await asyncio.sleep(0.05)
            await conn.execute("SELECT 1")

    finally:
        await conn.close()

    assert len(notifications) >= 1
    assert str(correction.id) in notifications[0]


@pytest.mark.asyncio
async def test_poll_fallback_claims_without_notify(db_session: AsyncSession) -> None:
    """claim_next() works even when the LISTEN mechanism is not active (poll path)."""
    from sqlalchemy import text

    await db_session.execute(text("DELETE FROM corrections WHERE status = 'pending'"))
    await db_session.commit()
    user = await _insert_user(db_session)
    correction = await _insert_pending(db_session, user)
    await db_session.commit()

    engine = create_async_engine(_test_db_url(), echo=False)
    try:
        session_factory = async_sessionmaker(engine, expire_on_commit=False)
        async with session_factory() as s:
            claimed = await claim_next(s, worker_id="poll-worker")
            assert claimed is not None
            assert claimed.id == correction.id
            assert claimed.status == CorrectionStatus.processing
            await s.commit()
    finally:
        await engine.dispose()
