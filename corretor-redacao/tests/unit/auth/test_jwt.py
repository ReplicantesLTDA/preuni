"""T070: JWT HS256 issue + verify; expiry honored; tampered tokens rejected."""

from __future__ import annotations

import jwt as pyjwt
import pytest
import time
import uuid

from src.auth.jwt import (
    JWTConfig,
    JWTExpiredError,
    JWTInvalidError,
    issue_access_token,
    verify_access_token,
)

CFG = JWTConfig(secret="test-secret-do-not-use-in-prod", algorithm="HS256", access_ttl_s=900)


def test_issue_then_verify_roundtrip() -> None:
    uid = uuid.uuid4()
    token = issue_access_token(user_id=uid, tier="free", config=CFG)
    payload = verify_access_token(token, config=CFG)
    assert payload["sub"] == str(uid)
    assert payload["tier"] == "free"
    assert "exp" in payload
    assert "iat" in payload


def test_tampered_token_rejected() -> None:
    token = issue_access_token(user_id=uuid.uuid4(), tier="premium", config=CFG)
    bad = token[:-3] + ("AAA" if token[-3:] != "AAA" else "BBB")
    with pytest.raises(JWTInvalidError):
        verify_access_token(bad, config=CFG)


def test_wrong_secret_rejected() -> None:
    token = issue_access_token(user_id=uuid.uuid4(), tier="free", config=CFG)
    other_cfg = JWTConfig(secret="different-secret", algorithm="HS256", access_ttl_s=900)
    with pytest.raises(JWTInvalidError):
        verify_access_token(token, config=other_cfg)


def test_expired_token_rejected() -> None:
    short_cfg = JWTConfig(secret=CFG.secret, algorithm="HS256", access_ttl_s=1)
    token = issue_access_token(user_id=uuid.uuid4(), tier="free", config=short_cfg)
    time.sleep(1.2)
    with pytest.raises(JWTExpiredError):
        verify_access_token(token, config=short_cfg)


def test_payload_carries_required_claims() -> None:
    uid = uuid.uuid4()
    token = issue_access_token(user_id=uid, tier="premium", config=CFG)
    decoded = pyjwt.decode(token, CFG.secret, algorithms=[CFG.algorithm])
    assert decoded["sub"] == str(uid)
    assert decoded["tier"] == "premium"
    assert decoded["typ"] == "access"


def test_invalid_string_rejected() -> None:
    with pytest.raises(JWTInvalidError):
        verify_access_token("not-a-jwt", config=CFG)


def test_tier_round_trips() -> None:
    for tier in ("free", "premium"):
        token = issue_access_token(user_id=uuid.uuid4(), tier=tier, config=CFG)
        payload = verify_access_token(token, config=CFG)
        assert payload["tier"] == tier
