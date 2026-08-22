"""JWT access-token issue + verify (PyJWT HS256).

Constitution III: access tokens carry user_id + tier; short TTL (default 15 min)
so revocation is cheap (just wait); refresh tokens (separate module) handle
long-lived auth state.

Research R5: PyJWT chosen for narrow surface + active maintenance.
"""

from __future__ import annotations

import jwt as pyjwt
import os
import time
import uuid
from dataclasses import dataclass
from jwt.exceptions import (
    ExpiredSignatureError,
    InvalidSignatureError,
    InvalidTokenError,
)


class JWTError(Exception):
    """Base — wraps any PyJWT verification failure under our taxonomy."""


class JWTInvalidError(JWTError):
    """Invalid signature, malformed token, wrong algorithm, or tampered payload."""


class JWTExpiredError(JWTError):
    """`exp` claim is in the past."""


@dataclass(slots=True, frozen=True)
class JWTConfig:
    secret: str
    algorithm: str = "HS256"
    access_ttl_s: int = 900  # 15 min default per FR-030 / quickstart


def default_config() -> JWTConfig:
    secret = os.environ.get("JWT_SECRET") or os.environ.get("JWT_SECRET_KEY")
    if not secret:
        raise RuntimeError("JWT_SECRET (or JWT_SECRET_KEY) must be set")
    return JWTConfig(
        secret=secret,
        algorithm=os.environ.get("JWT_ALGORITHM", "HS256"),
        access_ttl_s=int(os.environ.get("JWT_ACCESS_TTL_S", "900")),
    )


def issue_access_token(
    *,
    user_id: uuid.UUID,
    tier: str,
    config: JWTConfig | None = None,
) -> str:
    cfg = config or default_config()
    now = int(time.time())
    payload = {
        "sub": str(user_id),
        "tier": tier,
        "typ": "access",
        "iat": now,
        "exp": now + cfg.access_ttl_s,
    }
    return pyjwt.encode(payload, cfg.secret, algorithm=cfg.algorithm)


def verify_access_token(token: str, *, config: JWTConfig | None = None) -> dict:
    cfg = config or default_config()
    try:
        payload = pyjwt.decode(token, cfg.secret, algorithms=[cfg.algorithm])
    except ExpiredSignatureError as exc:
        raise JWTExpiredError("access token expired") from exc
    except (InvalidSignatureError, InvalidTokenError) as exc:
        raise JWTInvalidError(f"invalid access token: {exc!s}") from exc
    if payload.get("typ") != "access":
        raise JWTInvalidError("not an access token")
    return payload


__all__ = [
    "JWTConfig",
    "JWTError",
    "JWTExpiredError",
    "JWTInvalidError",
    "default_config",
    "issue_access_token",
    "verify_access_token",
]
