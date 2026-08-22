"""Refresh-token issue / verify / rotate / revoke.

Auth-layer wrapper over `db.repositories.refresh_token_repo`. Returns typed
errors so the API layer can map them to HTTP error_codes.
"""

from __future__ import annotations

import datetime as dt
import uuid
from dataclasses import dataclass
from sqlalchemy.ext.asyncio import AsyncSession

from src.db.models import RefreshToken
from src.db.repositories import refresh_token_repo

DEFAULT_REFRESH_TTL_S = 30 * 24 * 60 * 60  # 30 days


class InvalidRefreshTokenError(Exception):
    """Spec error_code: invalid_token. Covers unknown, revoked, expired."""


@dataclass(slots=True, frozen=True)
class RefreshTokenPayload:
    user_id: uuid.UUID
    token_id: uuid.UUID
    expires_at: dt.datetime


async def issue_refresh_token(
    *,
    session: AsyncSession,
    user_id: uuid.UUID,
    ttl_s: int = DEFAULT_REFRESH_TTL_S,
) -> tuple[str, RefreshToken]:
    cleartext = refresh_token_repo.generate_cleartext_token()
    expires_at = dt.datetime.now(dt.UTC) + dt.timedelta(seconds=ttl_s)
    row = await refresh_token_repo.insert_refresh_token(
        session=session,
        user_id=user_id,
        cleartext=cleartext,
        expires_at=expires_at,
    )
    return cleartext, row


async def verify_refresh_token(
    *,
    session: AsyncSession,
    cleartext: str,
) -> RefreshTokenPayload:
    row = await refresh_token_repo.get_active_by_cleartext(
        session=session,
        cleartext=cleartext,
    )
    if row is None:
        raise InvalidRefreshTokenError("refresh token unknown, revoked, or expired")
    return RefreshTokenPayload(user_id=row.user_id, token_id=row.id, expires_at=row.expires_at)


async def revoke_refresh_token(
    *,
    session: AsyncSession,
    cleartext: str,
) -> bool:
    return await refresh_token_repo.revoke_token(session=session, cleartext=cleartext)


async def rotate_refresh_token(
    *,
    session: AsyncSession,
    cleartext: str,
    ttl_s: int = DEFAULT_REFRESH_TTL_S,
) -> tuple[str, RefreshToken]:
    """Verify + revoke old + issue new. Atomic at the application layer."""
    payload = await verify_refresh_token(session=session, cleartext=cleartext)
    revoked = await refresh_token_repo.revoke_token(session=session, cleartext=cleartext)
    if not revoked:
        raise InvalidRefreshTokenError("rotation race: token already revoked")
    return await issue_refresh_token(session=session, user_id=payload.user_id, ttl_s=ttl_s)


__all__ = [
    "DEFAULT_REFRESH_TTL_S",
    "InvalidRefreshTokenError",
    "RefreshTokenPayload",
    "issue_refresh_token",
    "revoke_refresh_token",
    "rotate_refresh_token",
    "verify_refresh_token",
]
