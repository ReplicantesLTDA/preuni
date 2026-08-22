"""T119: delete_user script — cascade delete + audit trail."""

from __future__ import annotations

import hashlib
import pytest
import uuid
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from src.db.models import (
    Correction,
    CorrectionAuditLog,
    User,
)
from src.db.models.enums import AuditEventType, CorrectionStatus, UserTier
from src.scripts.delete_user import delete_user


async def _insert_user(session: AsyncSession) -> User:
    u = User(
        id=uuid.uuid4(),
        email=f"del-{uuid.uuid4()}@example.com",
        password_hash="x",
        tier=UserTier.free,
    )
    session.add(u)
    await session.flush()
    return u


async def _insert_correction(session: AsyncSession, user: User) -> Correction:
    text = "essay del test"
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


@pytest.mark.asyncio
async def test_delete_user_removes_user_row(db_session: AsyncSession) -> None:
    user = await _insert_user(db_session)
    user_id = user.id
    await db_session.commit()

    await delete_user(db_session, user_id=user_id, reason="user_requested", operator="test")
    await db_session.commit()

    result = await db_session.execute(select(User).where(User.id == user_id))
    assert result.scalar_one_or_none() is None


@pytest.mark.asyncio
async def test_delete_user_cascades_corrections(db_session: AsyncSession) -> None:
    user = await _insert_user(db_session)
    correction = await _insert_correction(db_session, user)
    correction_id = correction.id
    await db_session.commit()

    await delete_user(db_session, user_id=user.id, reason="user_requested", operator="test")
    await db_session.commit()

    result = await db_session.execute(select(Correction).where(Correction.id == correction_id))
    assert result.scalar_one_or_none() is None


@pytest.mark.asyncio
async def test_delete_user_writes_deleted_audit_event(db_session: AsyncSession) -> None:
    user = await _insert_user(db_session)
    correction = await _insert_correction(db_session, user)
    correction_id = correction.id
    await db_session.commit()

    await delete_user(db_session, user_id=user.id, reason="user_requested", operator="test")

    # Check audit log in session.new BEFORE commit: after commit the cascade
    # deletes the correction and its audit logs (LGPD full erasure).
    pending_logs = [
        obj
        for obj in db_session.new
        if isinstance(obj, CorrectionAuditLog)
        and obj.correction_id == correction_id
        and obj.event_type == AuditEventType.deleted
    ]
    assert len(pending_logs) >= 1
    await db_session.commit()
