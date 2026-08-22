"""POST /auth/{register, verify-email, login, refresh, logout} routes."""

from __future__ import annotations

import datetime as dt
import os
import uuid
from fastapi import APIRouter, Depends, Response, status
from pydantic import BaseModel, EmailStr, Field
from sqlalchemy import select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.ext.asyncio import AsyncSession

from src.api.deps import get_session_dep
from src.api.errors import typed_error_response
from src.auth.email_sender import send_verification_email
from src.auth.jwt import default_config, issue_access_token
from src.auth.passwords import (
    LeakedPasswordError,
    PasswordTooShortError,
    hash_password,
    validate_password_policy,
    verify_password,
)
from src.auth.refresh_tokens import (
    DEFAULT_REFRESH_TTL_S,
    InvalidRefreshTokenError,
    issue_refresh_token,
    revoke_refresh_token,
    rotate_refresh_token,
)
from src.auth.verification import (
    InvalidVerificationTokenError,
    consume_verification_token,
    issue_verification_token,
)
from src.db.models import User
from src.db.models.enums import UserTier

router = APIRouter(prefix="/auth", tags=["auth"])

_DEV_MODE = lambda: os.environ.get("APP_ENV", "dev") == "dev"  # noqa: E731


class RegisterRequest(BaseModel):
    email: EmailStr
    password: str = Field(min_length=1)  # full policy in service layer
    birth_date: dt.date | None = None


class LoginRequest(BaseModel):
    email: EmailStr
    password: str


class RefreshRequest(BaseModel):
    refresh_token: str


class LogoutRequest(BaseModel):
    refresh_token: str


class VerifyEmailRequest(BaseModel):
    token: str


class TokenPair(BaseModel):
    access_token: str
    refresh_token: str
    access_token_expires_in: int
    refresh_token_expires_in: int
    token_type: str = "Bearer"


@router.post("/register", status_code=status.HTTP_201_CREATED)
async def register(
    body: RegisterRequest,
    session: AsyncSession = Depends(get_session_dep),
) -> Response:
    try:
        validate_password_policy(body.password)
    except PasswordTooShortError:
        return typed_error_response(
            status_code=400,
            error_code="password_too_short",
            message="Senha abaixo do mínimo de 12 caracteres.",
        )
    except LeakedPasswordError:
        return typed_error_response(
            status_code=400,
            error_code="password_leaked",
            message="Senha aparece em listas de vazamento conhecidas.",
        )

    user = User(
        id=uuid.uuid4(),
        email=str(body.email),
        password_hash=hash_password(body.password),
        tier=UserTier.free,
        birth_date=body.birth_date,
    )
    session.add(user)
    try:
        await session.flush()
    except IntegrityError:
        await session.rollback()
        return typed_error_response(
            status_code=409,
            error_code="email_already_registered",
            message="E-mail já cadastrado.",
        )

    # Issue email-verification token + log/email link.
    verify_clear, _ = await issue_verification_token(session=session, user_id=user.id)
    import contextlib

    with contextlib.suppress(Exception):
        # Don't fail registration if SMTP is misconfigured; dev mode logs the link.
        send_verification_email(to_email=str(body.email), verification_token=verify_clear)

    payload: dict[str, object] = {
        "user_id": str(user.id),
        "email": user.email,
        "tier": user.tier.value,
        "verification_email_sent": True,
    }
    if _DEV_MODE():
        payload["verification_token_dev"] = verify_clear
    return Response(
        content=__import__("json").dumps(payload),
        media_type="application/json",
        status_code=201,
    )


@router.post("/verify-email", status_code=status.HTTP_204_NO_CONTENT)
async def verify_email(
    body: VerifyEmailRequest,
    session: AsyncSession = Depends(get_session_dep),
) -> Response:
    try:
        await consume_verification_token(session=session, cleartext=body.token)
    except InvalidVerificationTokenError:
        return typed_error_response(
            status_code=400,
            error_code="invalid_token",
            message="Token de verificação inválido ou expirado.",
        )
    return Response(status_code=204)


@router.post("/login", response_model=None)
async def login(
    body: LoginRequest,
    session: AsyncSession = Depends(get_session_dep),
) -> TokenPair | Response:
    user = (
        await session.execute(select(User).where(User.email == str(body.email)))
    ).scalar_one_or_none()
    # Always do a hash verify to keep timing consistent (anti user-enumeration).
    if user is None:
        # Dummy verify against a fixed hash so wrong-email and wrong-password take similar time.
        verify_password(body.password, "$argon2id$v=19$m=65536,t=3,p=2$dummy$dummy")
        return typed_error_response(
            status_code=401,
            error_code="invalid_credentials",
            message="E-mail ou senha incorretos.",
        )
    if not verify_password(body.password, user.password_hash):
        return typed_error_response(
            status_code=401,
            error_code="invalid_credentials",
            message="E-mail ou senha incorretos.",
        )

    cfg = default_config()
    access = issue_access_token(user_id=user.id, tier=user.tier.value, config=cfg)
    refresh_clear, _ = await issue_refresh_token(session=session, user_id=user.id)

    return TokenPair(
        access_token=access,
        refresh_token=refresh_clear,
        access_token_expires_in=cfg.access_ttl_s,
        refresh_token_expires_in=DEFAULT_REFRESH_TTL_S,
    )


@router.post("/refresh", response_model=None)
async def refresh(
    body: RefreshRequest,
    session: AsyncSession = Depends(get_session_dep),
) -> TokenPair | Response:
    try:
        new_clear, new_row = await rotate_refresh_token(
            session=session, cleartext=body.refresh_token
        )
    except InvalidRefreshTokenError:
        return typed_error_response(
            status_code=401,
            error_code="invalid_token",
            message="Refresh token inválido, revogado ou expirado.",
        )

    user = (
        await session.execute(select(User).where(User.id == new_row.user_id))
    ).scalar_one_or_none()
    if user is None:
        return typed_error_response(
            status_code=401,
            error_code="invalid_token",
            message="Usuário não encontrado.",
        )

    cfg = default_config()
    access = issue_access_token(user_id=user.id, tier=user.tier.value, config=cfg)
    return TokenPair(
        access_token=access,
        refresh_token=new_clear,
        access_token_expires_in=cfg.access_ttl_s,
        refresh_token_expires_in=DEFAULT_REFRESH_TTL_S,
    )


@router.post("/logout", status_code=status.HTTP_204_NO_CONTENT)
async def logout(
    body: LogoutRequest,
    session: AsyncSession = Depends(get_session_dep),
) -> Response:
    await revoke_refresh_token(session=session, cleartext=body.refresh_token)
    return Response(status_code=204)
