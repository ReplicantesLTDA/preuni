"""POST /corrections, GET /corrections/{id}, GET /corrections (list), POST reevaluations."""

from __future__ import annotations

import base64
import datetime as dt
import json
import uuid
from fastapi import APIRouter, Depends, HTTPException, Query, status
from fastapi.responses import JSONResponse
from pydantic import BaseModel
from sqlalchemy import and_, select
from sqlalchemy.ext.asyncio import AsyncSession

from src.api.deps import current_user_correction_ready, get_session_dep
from src.auth.quota import QuotaExhaustedError
from src.corrector.prevalidation.language import LanguageMismatchError
from src.corrector.prevalidation.length import (
    LengthTooLongError,
    LengthTooShortError,
    validate_length,
)
from src.corrector.prevalidation.theme import (
    ThemeMissingContextError,
    ThemeMissingTitleError,
    validate_theme,
)
from src.db.models import Correction, User
from src.db.models.enums import CorrectionStatus
from src.db.repositories.correction_repo import enqueue_with_quota_check, get_for_user

router = APIRouter(tags=["corrections"])


class PromptTheme(BaseModel):
    title: str
    context: str


class SubmitCorrectionRequest(BaseModel):
    essay_text: str
    prompt_theme: PromptTheme
    motivational_texts: str | None = None


class CorrectionAccepted(BaseModel):
    correction_id: str
    status: str
    reevaluation_of: str | None = None


class DryRunOk(BaseModel):
    ok_to_submit: bool = True


def _run_prevalidation(req: SubmitCorrectionRequest) -> None:
    """Run synchronous pre-validation; raises typed errors on failure."""
    validate_length(req.essay_text)
    validate_theme(title=req.prompt_theme.title, context=req.prompt_theme.context)


def _prevalidation_error_response(exc: Exception) -> HTTPException:
    if isinstance(exc, LengthTooShortError):
        return HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={
                "error_code": "length_too_short",
                "message": "A redação está abaixo do tamanho mínimo (500 caracteres ou 7 linhas).",
            },
        )
    if isinstance(exc, LengthTooLongError):
        return HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={
                "error_code": "length_too_long",
                "message": "A redação ultrapassa o tamanho máximo (3500 caracteres ou 50 linhas).",
            },
        )
    if isinstance(exc, ThemeMissingTitleError):
        return HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={
                "error_code": "theme_missing_title",
                "message": "O tema deve incluir um título.",
            },
        )
    if isinstance(exc, ThemeMissingContextError):
        return HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={
                "error_code": "theme_missing_context",
                "message": "O tema deve incluir título e uma breve contextualização.",
            },
        )
    if isinstance(exc, LanguageMismatchError):
        return HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={
                "error_code": "language_mismatch",
                "message": "A redação deve estar em português brasileiro.",
            },
        )
    raise exc


@router.post("/corrections", status_code=status.HTTP_202_ACCEPTED)
async def submit_correction(
    req: SubmitCorrectionRequest,
    dry_run: bool = Query(default=False),
    user: User = Depends(current_user_correction_ready),
    session: AsyncSession = Depends(get_session_dep),
) -> CorrectionAccepted | DryRunOk:
    try:
        _run_prevalidation(req)
    except (
        LengthTooShortError,
        LengthTooLongError,
        ThemeMissingTitleError,
        ThemeMissingContextError,
        LanguageMismatchError,
    ) as exc:
        raise _prevalidation_error_response(exc) from exc

    if dry_run:
        return JSONResponse(status_code=200, content={"ok_to_submit": True})

    try:
        correction = await enqueue_with_quota_check(
            session,
            user=user,
            essay_text=req.essay_text,
            prompt_theme_title=req.prompt_theme.title,
            prompt_theme_context=req.prompt_theme.context,
            motivational_texts=req.motivational_texts,
        )
    except QuotaExhaustedError as exc:
        raise HTTPException(
            status_code=status.HTTP_429_TOO_MANY_REQUESTS,
            detail={
                "error_code": "quota_exhausted",
                "message": "Você atingiu o limite mensal de correções.",
                "quota_reset_at": exc.quota_reset_at.isoformat(),
            },
        ) from exc

    return CorrectionAccepted(
        correction_id=str(correction.id),
        status="pending",
    )


def _correction_to_envelope(correction: Correction) -> dict:
    base = {
        "correction_id": str(correction.id),
        "status": correction.status.value,
        "issued_at": correction.queued_at.isoformat(),
    }
    if correction.parent_correction_id:
        base["reevaluation_of"] = str(correction.parent_correction_id)

    if correction.status == CorrectionStatus.pending:
        return base

    if correction.status == CorrectionStatus.processing:
        if correction.started_at:
            base["started_at"] = correction.started_at.isoformat()
        return base

    if correction.status == CorrectionStatus.completed:
        base.update(
            {
                "completed_at": correction.completed_at.isoformat()
                if correction.completed_at
                else None,
                "prompt_version": correction.prompt_version,
                "model_identifier": correction.model_identifier,
                "output_schema_version": correction.output_schema_version,
                "inference_params": {},
                "final_score": correction.final_score,
                "competencies": correction.competencies,
                "eliminatory_flags": correction.eliminatory_flags or [],
            }
        )
        return base

    if correction.status == CorrectionStatus.failed:
        base.update(
            {
                "completed_at": correction.completed_at.isoformat()
                if correction.completed_at
                else None,
                "error_code": correction.error_code,
                "message_pt_br": correction.error_message_pt_br or "",
                "quota_consumed": correction.quota_consumed,
            }
        )
        return base

    return base


def _encode_cursor(queued_at: dt.datetime, correction_id: uuid.UUID) -> str:
    payload = json.dumps({"q": queued_at.isoformat(), "id": str(correction_id)})
    return base64.urlsafe_b64encode(payload.encode()).decode()


def _decode_cursor(cursor: str) -> tuple[dt.datetime, uuid.UUID]:
    payload = json.loads(base64.urlsafe_b64decode(cursor.encode()).decode())
    return dt.datetime.fromisoformat(payload["q"]), uuid.UUID(payload["id"])


def _correction_to_summary(correction: Correction) -> dict:
    return {
        "correction_id": str(correction.id),
        "issued_at": correction.queued_at.isoformat(),
        "status": correction.status.value,
        "final_score": correction.final_score,
        "prompt_theme_title": correction.prompt_theme_title,
        "eliminatory_flags": correction.eliminatory_flags or [],
    }


@router.get("/corrections")
async def list_corrections(
    limit: int = Query(default=20, ge=1, le=100),
    cursor: str | None = Query(default=None),
    from_dt: dt.datetime | None = Query(default=None, alias="from"),
    to_dt: dt.datetime | None = Query(default=None, alias="to"),
    user: User = Depends(current_user_correction_ready),
    session: AsyncSession = Depends(get_session_dep),
) -> dict:
    filters = [Correction.user_id == user.id]

    if cursor:
        try:
            cursor_queued_at, cursor_id = _decode_cursor(cursor)
            filters.append(
                and_(
                    Correction.queued_at <= cursor_queued_at,
                    Correction.id != cursor_id,
                )
            )
        except Exception:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail={"error_code": "bad_request", "message": "Cursor inválido."},
            ) from None

    if from_dt:
        filters.append(Correction.queued_at >= from_dt)
    if to_dt:
        filters.append(Correction.queued_at <= to_dt)

    result = await session.execute(
        select(Correction)
        .where(*filters)
        .order_by(Correction.queued_at.desc(), Correction.id.desc())
        .limit(limit + 1)
    )
    rows = result.scalars().all()

    has_more = len(rows) > limit
    items = rows[:limit]

    next_cursor: str | None = None
    if has_more:
        last = items[-1]
        next_cursor = _encode_cursor(last.queued_at, last.id)

    return {
        "items": [_correction_to_summary(c) for c in items],
        **({"next_cursor": next_cursor} if next_cursor else {}),
    }


@router.get("/corrections/{correction_id}")
async def get_correction(
    correction_id: uuid.UUID,
    user: User = Depends(current_user_correction_ready),
    session: AsyncSession = Depends(get_session_dep),
) -> dict:
    correction = await get_for_user(session, correction_id=correction_id, user_id=user.id)
    if correction is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={"error_code": "not_found", "message": "Correção não encontrada."},
        )
    return _correction_to_envelope(correction)


@router.post("/corrections/{correction_id}/reevaluations", status_code=status.HTTP_202_ACCEPTED)
async def reevaluate_correction(
    correction_id: uuid.UUID,
    user: User = Depends(current_user_correction_ready),
    session: AsyncSession = Depends(get_session_dep),
) -> dict:
    original = await get_for_user(session, correction_id=correction_id, user_id=user.id)
    if original is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={"error_code": "not_found", "message": "Correção não encontrada."},
        )

    try:
        new_correction = await enqueue_with_quota_check(
            session,
            user=user,
            essay_text=original.essay_text,
            prompt_theme_title=original.prompt_theme_title,
            prompt_theme_context=original.prompt_theme_context,
            motivational_texts=original.motivational_texts,
        )
    except QuotaExhaustedError as exc:
        raise HTTPException(
            status_code=status.HTTP_429_TOO_MANY_REQUESTS,
            detail={
                "error_code": "quota_exhausted",
                "message": "Você atingiu o limite mensal de correções.",
                "quota_reset_at": exc.quota_reset_at.isoformat(),
            },
        ) from exc

    new_correction.parent_correction_id = original.id

    return {
        "correction_id": str(new_correction.id),
        "status": "pending",
        "reevaluation_of": str(original.id),
    }


__all__ = ["router"]
