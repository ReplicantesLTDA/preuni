"""T072: refresh-token repo — hashed at rest, rotated on refresh, revocable."""

from __future__ import annotations

import pytest
import uuid
from sqlalchemy import text

from src.auth.refresh_tokens import (
    InvalidRefreshTokenError,
    issue_refresh_token,
    revoke_refresh_token,
    rotate_refresh_token,
    verify_refresh_token,
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
async def test_issue_returns_cleartext_and_persists_hash(db_session) -> None:
    uid = await _seed_user(db_session)
    cleartext, stored = await issue_refresh_token(session=db_session, user_id=uid, ttl_s=86400)
    await db_session.flush()

    assert isinstance(cleartext, str) and len(cleartext) >= 40
    assert stored.user_id == uid
    assert stored.token_hash != cleartext.encode()  # hashed, not cleartext
    assert len(stored.token_hash) == 32  # sha256

    # Persisted lookup
    row = (
        await db_session.execute(
            text("SELECT token_hash FROM refresh_tokens WHERE id = :tid"),
            {"tid": stored.id},
        )
    ).first()
    assert row is not None


@pytest.mark.asyncio
async def test_verify_accepts_valid_token(db_session) -> None:
    uid = await _seed_user(db_session)
    cleartext, _ = await issue_refresh_token(session=db_session, user_id=uid, ttl_s=86400)
    await db_session.flush()

    payload = await verify_refresh_token(session=db_session, cleartext=cleartext)
    assert payload.user_id == uid


@pytest.mark.asyncio
async def test_verify_rejects_unknown_token(db_session) -> None:
    with pytest.raises(InvalidRefreshTokenError):
        await verify_refresh_token(session=db_session, cleartext="nonexistent")


@pytest.mark.asyncio
async def test_verify_rejects_revoked_token(db_session) -> None:
    uid = await _seed_user(db_session)
    cleartext, _ = await issue_refresh_token(session=db_session, user_id=uid, ttl_s=86400)
    await db_session.flush()

    await revoke_refresh_token(session=db_session, cleartext=cleartext)
    await db_session.flush()

    with pytest.raises(InvalidRefreshTokenError):
        await verify_refresh_token(session=db_session, cleartext=cleartext)


@pytest.mark.asyncio
async def test_verify_rejects_expired_token(db_session) -> None:
    uid = await _seed_user(db_session)
    cleartext, stored = await issue_refresh_token(session=db_session, user_id=uid, ttl_s=86400)
    # Force expiry in the past by mutating the ORM object directly (simplest path).
    import datetime as _dt

    stored.expires_at = _dt.datetime.now(_dt.UTC) - _dt.timedelta(days=1)
    await db_session.flush()

    with pytest.raises(InvalidRefreshTokenError):
        await verify_refresh_token(session=db_session, cleartext=cleartext)


@pytest.mark.asyncio
async def test_rotate_revokes_old_and_issues_new(db_session) -> None:
    uid = await _seed_user(db_session)
    old_clear, _ = await issue_refresh_token(session=db_session, user_id=uid, ttl_s=86400)
    await db_session.flush()

    new_clear, new_stored = await rotate_refresh_token(
        session=db_session,
        cleartext=old_clear,
        ttl_s=86400,
    )
    await db_session.flush()

    # Old token now revoked.
    with pytest.raises(InvalidRefreshTokenError):
        await verify_refresh_token(session=db_session, cleartext=old_clear)

    # New token valid.
    payload = await verify_refresh_token(session=db_session, cleartext=new_clear)
    assert payload.user_id == uid
    assert new_stored.user_id == uid
