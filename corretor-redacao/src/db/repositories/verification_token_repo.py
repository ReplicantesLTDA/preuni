"""Email-verification token repository (SHA-256 hashed at rest, single-use)."""

from __future__ import annotations

import datetime as dt
import hashlib
import secrets
import uuid
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from src.db.models import EmailVerificationToken


def generate_cleartext_token(n_bytes: int = 32) -> str:
    return secrets.token_urlsafe(n_bytes)


def hash_token(cleartext: str) -> bytes:
    return hashlib.sha256(cleartext.encode("utf-8")).digest()


async def insert_token(
    *,
    session: AsyncSession,
    user_id: uuid.UUID,
    cleartext: str,
    expires_at: dt.datetime,
) -> EmailVerificationToken:
    row = EmailVerificationToken(
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
) -> EmailVerificationToken | None:
    """Return row iff: hash matches, not used, not expired."""
    token_hash = hash_token(cleartext)
    stmt = select(EmailVerificationToken).where(EmailVerificationToken.token_hash == token_hash)
    result = await session.execute(stmt)
    row = result.scalar_one_or_none()
    if row is None:
        return None
    if row.used_at is not None:
        return None
    if row.expires_at <= dt.datetime.now(dt.UTC):
        return None
    return row


async def mark_used(*, session: AsyncSession, token: EmailVerificationToken) -> None:  # noqa: ARG001 — session kept for symmetry with sibling repos
    token.used_at = dt.datetime.now(dt.UTC)


__all__ = [
    "generate_cleartext_token",
    "get_active_by_cleartext",
    "hash_token",
    "insert_token",
    "mark_used",
]
