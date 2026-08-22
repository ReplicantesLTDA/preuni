"""T074: email verification token — hashed at rest, single-use, 24h TTL."""

from __future__ import annotations

import datetime as dt
import pytest
import uuid
from sqlalchemy import text

from src.auth.verification import (
    EMAIL_VERIFICATION_TTL_S,
    InvalidVerificationTokenError,
    consume_verification_token,
    issue_verification_token,
)


async def _seed_user(session) -> uuid.UUID:
    uid = uuid.uuid4()
    await session.execute(
        text("""
        INSERT INTO users (id, email, password_hash, tier, created_at, updated_at)
        VALUES (:uid, :email, 'x', 'free', now(), now())
    """),
        {"uid": uid, "email": f"u-{uid}@x.test"},
    )
    return uid


@pytest.mark.asyncio
async def test_issue_returns_cleartext_and_stores_hash(db_session) -> None:
    uid = await _seed_user(db_session)
    cleartext, stored = await issue_verification_token(session=db_session, user_id=uid)
    await db_session.flush()
    assert len(cleartext) >= 40
    assert stored.user_id == uid
    assert len(stored.token_hash) == 32  # sha256


def test_ttl_defaults_to_24h() -> None:
    assert EMAIL_VERIFICATION_TTL_S == 24 * 3600


@pytest.mark.asyncio
async def test_consume_marks_as_used_and_returns_user_id(db_session) -> None:
    uid = await _seed_user(db_session)
    cleartext, _ = await issue_verification_token(session=db_session, user_id=uid)
    await db_session.flush()

    consumed_uid = await consume_verification_token(session=db_session, cleartext=cleartext)
    assert consumed_uid == uid

    # Second consumption rejected.
    with pytest.raises(InvalidVerificationTokenError):
        await consume_verification_token(session=db_session, cleartext=cleartext)


@pytest.mark.asyncio
async def test_consume_rejects_unknown(db_session) -> None:
    with pytest.raises(InvalidVerificationTokenError):
        await consume_verification_token(session=db_session, cleartext="bogus")


@pytest.mark.asyncio
async def test_consume_rejects_expired(db_session) -> None:
    uid = await _seed_user(db_session)
    cleartext, stored = await issue_verification_token(session=db_session, user_id=uid)
    await db_session.flush()

    stored.expires_at = dt.datetime.now(dt.UTC) - dt.timedelta(hours=1)
    await db_session.flush()

    with pytest.raises(InvalidVerificationTokenError):
        await consume_verification_token(session=db_session, cleartext=cleartext)
