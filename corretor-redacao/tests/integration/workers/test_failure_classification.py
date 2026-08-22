"""T100: failure classification — quota_consumed per error_codes.md FR-036."""

from __future__ import annotations

import hashlib
import os
import pytest
import uuid
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from src.db.models import Correction, User
from src.db.models.enums import CorrectionStatus, UserTier
from src.db.repositories.correction_repo import claim_next, mark_failed
from src.workers.failure_classifier import PROVIDER_ERROR_CODES, USER_ERROR_CODES, classify_failure


async def _insert_user(session: AsyncSession) -> User:
    u = User(
        id=uuid.uuid4(),
        email=f"fc-{uuid.uuid4()}@example.com",
        password_hash="x",
        tier=UserTier.free,
    )
    session.add(u)
    await session.flush()
    return u


async def _insert_pending(session: AsyncSession, user: User, n: int = 0) -> Correction:
    text = f"essay text {n}"
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
    user = await _insert_user(db_session)
    correction = await _insert_pending(db_session, user)
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
    finally:
        await engine.dispose()

    row = (
        await db_session.execute(select(Correction).where(Correction.id == correction.id))
    ).scalar_one()
    assert row.status == CorrectionStatus.failed
    assert row.quota_consumed is False
    assert row.error_code == "schema_violation"


@pytest.mark.asyncio
async def test_mark_failed_user_error_quota_true(db_session: AsyncSession) -> None:
    user = await _insert_user(db_session)
    correction = await _insert_pending(db_session, user, n=1)
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
    finally:
        await engine.dispose()

    row = (
        await db_session.execute(select(Correction).where(Correction.id == correction.id))
    ).scalar_one()
    assert row.status == CorrectionStatus.failed
    assert row.quota_consumed is True
