"""Email-verification token issue + consume.

Also marks the verified user's `email_verified_at` on successful consumption.
"""

from __future__ import annotations

import datetime as dt
import uuid
from sqlalchemy import update
from sqlalchemy.ext.asyncio import AsyncSession

from src.db.models import EmailVerificationToken, User
from src.db.repositories import verification_token_repo

EMAIL_VERIFICATION_TTL_S = 24 * 3600  # 24h per FR-029


class InvalidVerificationTokenError(Exception):
    """Spec error_code: invalid_token. Covers unknown, used, expired."""


async def issue_verification_token(
    *,
    session: AsyncSession,
    user_id: uuid.UUID,
    ttl_s: int = EMAIL_VERIFICATION_TTL_S,
) -> tuple[str, EmailVerificationToken]:
    cleartext = verification_token_repo.generate_cleartext_token()
    expires_at = dt.datetime.now(dt.UTC) + dt.timedelta(seconds=ttl_s)
    row = await verification_token_repo.insert_token(
        session=session,
        user_id=user_id,
        cleartext=cleartext,
        expires_at=expires_at,
    )
    return cleartext, row


async def consume_verification_token(
    *,
    session: AsyncSession,
    cleartext: str,
) -> uuid.UUID:
    """Mark token used and stamp the user's email_verified_at; return user_id."""
    row = await verification_token_repo.get_active_by_cleartext(
        session=session,
        cleartext=cleartext,
    )
    if row is None:
        raise InvalidVerificationTokenError("verification token unknown, used, or expired")
    await verification_token_repo.mark_used(session=session, token=row)
    await session.execute(
        update(User).where(User.id == row.user_id).values(email_verified_at=dt.datetime.now(dt.UTC))
    )
    return row.user_id


__all__ = [
    "EMAIL_VERIFICATION_TTL_S",
    "InvalidVerificationTokenError",
    "consume_verification_token",
    "issue_verification_token",
]
