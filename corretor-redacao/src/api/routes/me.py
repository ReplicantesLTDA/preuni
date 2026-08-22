"""GET /me — profile + quota."""

from __future__ import annotations

import datetime as dt
from fastapi import APIRouter, Depends
from pydantic import BaseModel, EmailStr
from sqlalchemy import func, select
from sqlalchemy.ext.asyncio import AsyncSession

from src.api.deps import current_user, get_session_dep
from src.db.models import Correction, User
from src.db.models.enums import UserTier

# Quota per tier (Constitution X / spec FR-034). Initial estimates pending unit
# economics validation; live in env when product wants to A/B them.
QUOTA_BY_TIER: dict[UserTier, int] = {
    UserTier.free: 3,
    UserTier.premium: 30,
}

router = APIRouter(tags=["me"])


class ProfileResponse(BaseModel):
    user_id: str
    email: EmailStr
    tier: str
    quota_used_current_month: int
    quota_limit: int
    quota_reset_at: dt.datetime


def _month_window(now: dt.datetime) -> tuple[dt.datetime, dt.datetime]:
    start = now.replace(day=1, hour=0, minute=0, second=0, microsecond=0)
    next_month = start.month + 1
    next_year = start.year + (1 if next_month > 12 else 0)
    next_month = 1 if next_month > 12 else next_month
    end = start.replace(year=next_year, month=next_month)
    return start, end


@router.get("/me", response_model=ProfileResponse)
async def me(
    user: User = Depends(current_user),
    session: AsyncSession = Depends(get_session_dep),
) -> ProfileResponse:
    now = dt.datetime.now(dt.UTC)
    start, end = _month_window(now)
    used = (
        await session.execute(
            select(func.count(Correction.id))
            .where(Correction.user_id == user.id)
            .where(Correction.queued_at >= start)
            .where(Correction.queued_at < end)
        )
    ).scalar_one()

    return ProfileResponse(
        user_id=str(user.id),
        email=user.email,
        tier=user.tier.value,
        quota_used_current_month=int(used),
        quota_limit=QUOTA_BY_TIER[user.tier],
        quota_reset_at=end,
    )
