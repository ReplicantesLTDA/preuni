"""T087: quota counting and exhaustion guard."""

from __future__ import annotations

import datetime as dt
import pytest
import uuid
from unittest.mock import AsyncMock, MagicMock

from src.auth.quota import (
    QUOTA_LIMIT,
    QuotaExhaustedError,
    _month_window,
    check_quota,
    count_corrections_this_month,
)
from src.db.models import User
from src.db.models.enums import UserTier


def _user(tier: UserTier = UserTier.free) -> User:
    return User(
        id=uuid.uuid4(),
        email="test@example.com",
        password_hash="x",
        tier=tier,
        email_verified_at=dt.datetime.now(dt.UTC),
    )


def _mock_session(count: int) -> AsyncMock:
    session = AsyncMock()
    result = MagicMock()
    result.scalar_one.return_value = count
    session.execute.return_value = result
    return session


def test_quota_limits() -> None:
    assert QUOTA_LIMIT[UserTier.free] == 3
    assert QUOTA_LIMIT[UserTier.premium] == 30


def test_month_window_start_is_first_day() -> None:
    now = dt.datetime(2026, 6, 15, 12, 30, 0, tzinfo=dt.UTC)
    start, _ = _month_window(now)
    assert start == dt.datetime(2026, 6, 1, 0, 0, 0, tzinfo=dt.UTC)


def test_month_window_end_is_first_of_next_month() -> None:
    now = dt.datetime(2026, 6, 15, 12, 30, 0, tzinfo=dt.UTC)
    _, end = _month_window(now)
    assert end == dt.datetime(2026, 7, 1, 0, 0, 0, tzinfo=dt.UTC)


def test_month_window_december_rolls_to_january() -> None:
    now = dt.datetime(2026, 12, 15, tzinfo=dt.UTC)
    _, end = _month_window(now)
    assert end == dt.datetime(2027, 1, 1, 0, 0, 0, tzinfo=dt.UTC)


def test_month_window_january() -> None:
    now = dt.datetime(2026, 1, 1, tzinfo=dt.UTC)
    start, end = _month_window(now)
    assert start == dt.datetime(2026, 1, 1, 0, 0, 0, tzinfo=dt.UTC)
    assert end == dt.datetime(2026, 2, 1, 0, 0, 0, tzinfo=dt.UTC)


@pytest.mark.asyncio
async def test_count_corrections_calls_db() -> None:
    session = _mock_session(2)
    uid = uuid.uuid4()
    now = dt.datetime(2026, 6, 15, tzinfo=dt.UTC)
    count = await count_corrections_this_month(session, uid, now=now)
    assert count == 2
    session.execute.assert_called_once()


@pytest.mark.asyncio
async def test_check_quota_passes_when_under_limit() -> None:
    user = _user(UserTier.free)
    session = _mock_session(2)
    now = dt.datetime(2026, 6, 15, tzinfo=dt.UTC)
    used = await check_quota(session, user, now=now)
    assert used == 2


@pytest.mark.asyncio
async def test_check_quota_raises_at_limit_free() -> None:
    user = _user(UserTier.free)
    session = _mock_session(3)
    now = dt.datetime(2026, 6, 15, tzinfo=dt.UTC)
    with pytest.raises(QuotaExhaustedError) as exc:
        await check_quota(session, user, now=now)
    assert exc.value.used == 3
    assert exc.value.limit == 3
    assert exc.value.quota_reset_at == dt.datetime(2026, 7, 1, 0, 0, 0, tzinfo=dt.UTC)


@pytest.mark.asyncio
async def test_check_quota_raises_over_limit_free() -> None:
    user = _user(UserTier.free)
    session = _mock_session(5)
    now = dt.datetime(2026, 6, 15, tzinfo=dt.UTC)
    with pytest.raises(QuotaExhaustedError) as exc:
        await check_quota(session, user, now=now)
    assert exc.value.used == 5
    assert exc.value.limit == 3


@pytest.mark.asyncio
async def test_premium_user_under_quota() -> None:
    user = _user(UserTier.premium)
    session = _mock_session(29)
    now = dt.datetime(2026, 6, 15, tzinfo=dt.UTC)
    used = await check_quota(session, user, now=now)
    assert used == 29


@pytest.mark.asyncio
async def test_premium_user_exhausted() -> None:
    user = _user(UserTier.premium)
    session = _mock_session(30)
    now = dt.datetime(2026, 6, 15, tzinfo=dt.UTC)
    with pytest.raises(QuotaExhaustedError) as exc:
        await check_quota(session, user, now=now)
    assert exc.value.limit == 30
    assert exc.value.quota_reset_at == dt.datetime(2026, 7, 1, 0, 0, 0, tzinfo=dt.UTC)


@pytest.mark.asyncio
async def test_check_quota_zero_used_passes() -> None:
    user = _user(UserTier.free)
    session = _mock_session(0)
    now = dt.datetime(2026, 6, 15, tzinfo=dt.UTC)
    used = await check_quota(session, user, now=now)
    assert used == 0
