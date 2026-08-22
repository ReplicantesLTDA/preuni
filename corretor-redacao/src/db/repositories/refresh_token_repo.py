"""Refresh-token repository — low-level CRUD for the auth layer.

Tokens are SHA-256-hashed at rest (Research R14). The cleartext is returned to
the caller and never persisted.
"""

from __future__ import annotations

import datetime as dt
import hashlib
import secrets
import uuid
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from src.db.models import RefreshToken


def generate_cleartext_token(n_bytes: int = 32) -> str:
    """Cryptographically random, URL-safe token (length ~43 chars for 32 bytes)."""
    return secrets.token_urlsafe(n_bytes)


def hash_token(cleartext: str) -> bytes:
    return hashlib.sha256(cleartext.encode("utf-8")).digest()


async def insert_refresh_token(
    *,
    session: AsyncSession,
    user_id: uuid.UUID,
    cleartext: str,
    expires_at: dt.datetime,
) -> RefreshToken:
    row = RefreshToken(
        id=uuid.uuid4(),
        user_id=user_id,
        token_hash=hash_token(cleartext),
        expires_at=expires_at,
    )
    session.add(row)
    return row


async def get_active_by_cleartext(
    *,
    session: AsyncSession,
    cleartext: str,
) -> RefreshToken | None:
    """Return token row iff: hash matches, not revoked, not expired."""
    token_hash = hash_token(cleartext)
    stmt = select(RefreshToken).where(RefreshToken.token_hash == token_hash)
    result = await session.execute(stmt)
    row = result.scalar_one_or_none()
    if row is None:
        return None
    if row.revoked_at is not None:
        return None
    if row.expires_at <= dt.datetime.now(dt.UTC):
        return None
    return row


async def revoke_token(*, session: AsyncSession, cleartext: str) -> bool:
    """Mark token as revoked. Returns True if a token was revoked."""
    token_hash = hash_token(cleartext)
    stmt = select(RefreshToken).where(RefreshToken.token_hash == token_hash)
    result = await session.execute(stmt)
    row = result.scalar_one_or_none()
    if row is None or row.revoked_at is not None:
        return False
    row.revoked_at = dt.datetime.now(dt.UTC)
    return True


__all__ = [
    "generate_cleartext_token",
    "get_active_by_cleartext",
    "hash_token",
    "insert_refresh_token",
    "revoke_token",
]
