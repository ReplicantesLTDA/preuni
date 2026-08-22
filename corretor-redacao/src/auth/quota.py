"""Quota enforcement — Constitution X + spec FR-034/FR-035."""

from __future__ import annotations

import datetime as dt
import uuid
from sqlalchemy import func, select
from sqlalchemy.ext.asyncio import AsyncSession

from src.db.models import Correction, User
from src.db.models.enums import UserTier

QUOTA_LIMIT: dict[UserTier, int] = {
    UserTier.free: 3,
    UserTier.premium: 30,
}


class QuotaExhaustedError(Exception):
    def __init__(self, *, quota_reset_at: dt.datetime, used: int, limit: int) -> None:
        self.quota_reset_at = quota_reset_at
        self.used = used
        self.limit = limit
        super().__init__(f"Quota exhausted: {used}/{limit}, resets at {quota_reset_at}")


def _month_window(now: dt.datetime) -> tuple[dt.datetime, dt.datetime]:
    start = now.replace(day=1, hour=0, minute=0, second=0, microsecond=0)
    next_month = start.month + 1
    next_year = start.year + (1 if next_month > 12 else 0)
    next_month = 1 if next_month > 12 else next_month
    end = start.replace(year=next_year, month=next_month)
    return start, end


async def count_corrections_this_month(
    session: AsyncSession,
    user_id: uuid.UUID,
    now: dt.datetime | None = None,
) -> int:
    """COUNT quota-consuming corrections for user in the current calendar month (UTC)."""
    now = now or dt.datetime.now(dt.UTC)
    start, end = _month_window(now)
    result = await session.execute(
        select(func.count(Correction.id))
        .where(Correction.user_id == user_id)
        .where(Correction.queued_at >= start)
        .where(Correction.queued_at < end)
        .where(Correction.quota_consumed.is_(True))
    )
    return int(result.scalar_one())


async def check_quota(
    session: AsyncSession,
    user: User,
    now: dt.datetime | None = None,
) -> int:
    """Return current-month usage. Raise QuotaExhaustedError when at/over limit."""
    now = now or dt.datetime.now(dt.UTC)
    _, reset_at = _month_window(now)
    limit = QUOTA_LIMIT[user.tier]
    used = await count_corrections_this_month(session, user.id, now=now)
    if used >= limit:
        raise QuotaExhaustedError(quota_reset_at=reset_at, used=used, limit=limit)
    return used


__all__ = [
    "QUOTA_LIMIT",
    "QuotaExhaustedError",
    "_month_window",
    "check_quota",
    "count_corrections_this_month",
]
