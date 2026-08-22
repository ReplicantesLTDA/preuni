"""FastAPI dependencies — DB session, current_user (Bearer JWT)."""

from __future__ import annotations

import uuid
from collections.abc import AsyncIterator
from fastapi import Depends, Header, HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from src.auth.consent import ParentalConsentRequiredError, ensure_consent_for_minor
from src.auth.jwt import JWTExpiredError, JWTInvalidError, default_config, verify_access_token
from src.db.engine import session_factory
from src.db.models import User


async def get_session_dep() -> AsyncIterator[AsyncSession]:
    Session = session_factory()  # noqa: N806
    session: AsyncSession = Session()
    try:
        yield session
        await session.commit()
    except Exception:
        await session.rollback()
        raise
    finally:
        await session.close()


def _bearer_token(authorization: str | None) -> str:
    if not authorization:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail={"error_code": "unauthenticated", "message": "Token de acesso ausente."},
        )
    if not authorization.lower().startswith("bearer "):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail={
                "error_code": "unauthenticated",
                "message": "Cabeçalho Authorization inválido.",
            },
        )
    return authorization[7:].strip()


async def current_user(
    authorization: str | None = Header(default=None, alias="Authorization"),
    session: AsyncSession = Depends(get_session_dep),
) -> User:
    token = _bearer_token(authorization)
    try:
        payload = verify_access_token(token, config=default_config())
    except JWTExpiredError as exc:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail={"error_code": "token_expired", "message": "Token de acesso expirado."},
        ) from exc
    except JWTInvalidError as exc:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail={"error_code": "unauthenticated", "message": "Token de acesso inválido."},
        ) from exc

    try:
        user_id = uuid.UUID(payload["sub"])
    except (KeyError, ValueError) as exc:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail={"error_code": "unauthenticated", "message": "Token de acesso inválido."},
        ) from exc

    user = (await session.execute(select(User).where(User.id == user_id))).scalar_one_or_none()
    if user is None:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail={"error_code": "unauthenticated", "message": "Usuário não encontrado."},
        )
    return user


async def current_user_correction_ready(
    user: User = Depends(current_user),
) -> User:
    """current_user + enforces email-verified + minor-consent guards.

    Use on correction endpoints; do NOT use on /me (which is reachable by an
    unverified user so they can see their profile + verification status)."""
    if user.email_verified_at is None:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail={
                "error_code": "email_unverified",
                "message": "Confirme seu e-mail antes de enviar redações.",
            },
        )
    try:
        ensure_consent_for_minor(user)
    except ParentalConsentRequiredError as exc:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail={
                "error_code": "parental_consent_required",
                "message": "É necessário o consentimento do responsável para usuários menores de 18 anos.",
            },
        ) from exc
    return user


__all__ = ["current_user", "current_user_correction_ready", "get_session_dep"]
