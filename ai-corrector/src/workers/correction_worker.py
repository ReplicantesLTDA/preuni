"""Async correction worker — LISTEN/NOTIFY + 5-second poll backstop (research R1).

Lifecycle:
  1. Opens a dedicated asyncpg LISTEN connection for correction_queued channel.
  2. On every wake (NOTIFY received OR 5 s poll timer fires), calls claim_next().
  3. Runs the correction pipeline on the claimed row.
  4. Persists grader_pass + updates corrections aggregate + writes audit events.
  5. On provider/internal failure: status=failed, quota_consumed=false.
  6. On user-attributable failure: status=failed, quota_consumed=true.
"""

from __future__ import annotations

import asyncio
import decimal
import logging
import uuid
from contextlib import suppress
from opentelemetry import trace as otel_trace
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine
from typing import Any

from src.corrector.graders.single_grader import SingleGraderInput, SingleGraderResult, derive_seed
from src.corrector.llm.errors import (
    LLMError,
    RateLimitError,
    SchemaViolationError,
    TimeoutError,
    TransientError,
)
from src.corrector.pipeline import correct_essay
from src.db.models import GraderPass
from src.db.models.enums import AuditEventType
from src.db.repositories.audit_log_repo import AuditLogWriter
from src.db.repositories.correction_repo import claim_next, mark_completed, mark_failed
from src.workers.failure_classifier import classify_failure

log = logging.getLogger(__name__)
_tracer = otel_trace.get_tracer(__name__, "0.1.0")

POLL_INTERVAL_S = 5.0
LISTEN_CHANNEL = "correction_queued"


def _llm_error_to_code(exc: LLMError) -> str:
    if isinstance(exc, RateLimitError):
        return "provider_rate_limited"
    if isinstance(exc, TimeoutError):
        return "provider_timeout"
    if isinstance(exc, TransientError):
        return "provider_unavailable"
    if isinstance(exc, SchemaViolationError):
        return "schema_violation"
    return "internal_error"


def _llm_error_message_pt_br(error_code: str) -> str:
    messages = {
        "provider_rate_limited": "O servidor de correção está sobrecarregado. Tente novamente em breve.",
        "provider_timeout": "O servidor de correção demorou demais para responder. Tente novamente.",
        "provider_unavailable": "O servidor de correção está indisponível no momento.",
        "schema_violation": "O modelo não produziu uma saída válida após duas tentativas.",
        "internal_error": "Ocorreu um erro interno. Tente novamente em breve.",
        "language_mismatch": "A redação não está em português brasileiro.",
        "theme_missing_context": "O tema não contém contextualização suficiente.",
        "length_too_short": "A redação está abaixo do tamanho mínimo.",
        "length_too_long": "A redação ultrapassa o tamanho máximo.",
    }
    return messages.get(error_code, "Ocorreu um erro durante a correção.")


async def _persist_grader_pass(
    session: AsyncSession,
    result: SingleGraderResult,
    correction_id: uuid.UUID,
) -> None:
    competencies_dict = {
        k: {
            "score": v.score,
            "excerpt": v.excerpt,
            "justification_pt_br": v.justification_pt_br,
            **(
                {"improvement_path_pt_br": v.improvement_path_pt_br}
                if v.improvement_path_pt_br
                else {}
            ),
        }
        for k, v in result.competencies.items()
    }
    grader_pass = GraderPass(
        id=uuid.uuid4(),
        correction_id=correction_id,
        pass_index=0,
        c1_score=result.competencies["c1"].score,
        c2_score=result.competencies["c2"].score,
        c3_score=result.competencies["c3"].score,
        c4_score=result.competencies["c4"].score,
        c5_score=result.competencies["c5"].score,
        competencies=competencies_dict,
        eliminatory_flags=result.eliminatory_flags,
        seed=result.seed,
        prompt_version=result.prompt_version,
        model_identifier=result.model_id,
        output_schema_version=result.output_schema_version,
        inference_params=result.inference_params,
        raw_output=result.raw_output,
        prompt_tokens=result.prompt_tokens,
        completion_tokens=result.completion_tokens,
        latency_ms=result.latency_ms,
        cost_usd=decimal.Decimal(str(result.cost_usd)) if result.cost_usd is not None else None,
    )
    session.add(grader_pass)


async def process_one(
    session: AsyncSession,
    *,
    provider: Any,
    worker_id: str,
) -> bool:
    """Claim and process one pending correction. Returns True if a job was processed."""
    with _tracer.start_as_current_span("queue_pull"):
        correction = await claim_next(session, worker_id=worker_id)
    if correction is None:
        return False

    correction_id = correction.id
    audit = AuditLogWriter(session)

    try:
        inp = SingleGraderInput(
            correction_id=correction_id,
            essay_text=correction.essay_text,
            prompt_theme_title=correction.prompt_theme_title,
            prompt_theme_context=correction.prompt_theme_context,
            motivational_texts=correction.motivational_texts,
        )

        await audit.write(
            correction_id=correction_id,
            input_hash=correction.input_hash,
            event_type=AuditEventType.llm_started,
            payload={
                "prompt_version": "unknown",
                "model_identifier": "unknown",
                "seed": derive_seed(correction_id),
                "attempt": 1,
            },
        )

        with _tracer.start_as_current_span("llm_call") as llm_span:
            llm_span.set_attribute("correction_id", str(correction_id))
            result: SingleGraderResult = await correct_essay(inp, provider=provider)

        with _tracer.start_as_current_span("schema_validate"):
            pass  # validation happens inside correct_essay; span marks the boundary

        competencies_dict = {
            k: {
                "score": v.score,
                "excerpt": v.excerpt,
                "justification_pt_br": v.justification_pt_br,
                **(
                    {"improvement_path_pt_br": v.improvement_path_pt_br}
                    if v.improvement_path_pt_br
                    else {}
                ),
            }
            for k, v in result.competencies.items()
        }

        await audit.write(
            correction_id=correction_id,
            input_hash=correction.input_hash,
            event_type=AuditEventType.llm_completed,
            payload={
                "prompt_tokens": result.prompt_tokens,
                "completion_tokens": result.completion_tokens,
                "latency_ms": result.latency_ms,
                "cost_usd": float(result.cost_usd) if result.cost_usd is not None else None,
                "attempt": result.attempts,
            },
        )

        with _tracer.start_as_current_span("persist"):
            await _persist_grader_pass(session, result, correction_id)

        await mark_completed(
            session,
            correction=correction,
            final_score=result.final_score,
            c1_score=result.competencies["c1"].score,
            c2_score=result.competencies["c2"].score,
            c3_score=result.competencies["c3"].score,
            c4_score=result.competencies["c4"].score,
            c5_score=result.competencies["c5"].score,
            competencies=competencies_dict,
            eliminatory_flags=result.eliminatory_flags,
            prompt_version=result.prompt_version,
            model_identifier=result.model_id,
            output_schema_version=result.output_schema_version,
        )

        await audit.write(
            correction_id=correction_id,
            input_hash=correction.input_hash,
            event_type=AuditEventType.completed,
            payload={
                "final_score": result.final_score,
                "eliminatory_flags": result.eliminatory_flags,
                "prompt_version": result.prompt_version,
                "model_identifier": result.model_id,
            },
        )

    except LLMError as exc:
        error_code = _llm_error_to_code(exc)
        quota_consumed = classify_failure(error_code)
        await mark_failed(
            session,
            correction=correction,
            error_code=error_code,
            error_message_pt_br=_llm_error_message_pt_br(error_code),
            quota_consumed=quota_consumed,
        )
        await audit.write(
            correction_id=correction_id,
            input_hash=correction.input_hash,
            event_type=AuditEventType.failed,
            payload={"error_code": error_code, "quota_consumed": quota_consumed},
        )
        log.warning(
            "worker.correction_failed",
            extra={"correction_id": str(correction_id), "error_code": error_code},
        )

    except Exception:
        error_code = "internal_error"
        await mark_failed(
            session,
            correction=correction,
            error_code=error_code,
            error_message_pt_br=_llm_error_message_pt_br(error_code),
            quota_consumed=False,
        )
        await audit.write(
            correction_id=correction_id,
            input_hash=correction.input_hash,
            event_type=AuditEventType.failed,
            payload={"error_code": error_code, "quota_consumed": False},
        )
        log.exception("worker.unexpected_error", extra={"correction_id": str(correction_id)})

    return True


class CorrectionWorker:
    """Long-running worker: LISTEN/NOTIFY + 5 s poll backstop."""

    def __init__(
        self,
        *,
        database_url: str,
        provider: Any,
        worker_id: str | None = None,
        poll_interval_s: float = POLL_INTERVAL_S,
    ) -> None:
        self.database_url = database_url
        self.provider = provider
        self.worker_id = worker_id or f"worker-{uuid.uuid4().hex[:8]}"
        self.poll_interval_s = poll_interval_s
        self._stop = asyncio.Event()

    def stop(self) -> None:
        self._stop.set()

    async def run(self) -> None:
        engine = create_async_engine(self.database_url, echo=False)
        session_factory = async_sessionmaker(engine, expire_on_commit=False)

        log.info("worker.started", extra={"worker_id": self.worker_id})

        try:
            await self._run_loop(session_factory)
        finally:
            await engine.dispose()
            log.info("worker.stopped", extra={"worker_id": self.worker_id})

    async def _run_loop(self, session_factory: async_sessionmaker[AsyncSession]) -> None:
        wake = asyncio.Event()

        async def _listen_task() -> None:
            try:
                import asyncpg

                asyncpg_url = self.database_url.replace("postgresql+asyncpg://", "postgresql://")
                conn = await asyncpg.connect(asyncpg_url)
                try:
                    await conn.add_listener(LISTEN_CHANNEL, lambda *_: wake.set())
                    while not self._stop.is_set():
                        await asyncio.sleep(1)
                finally:
                    with suppress(Exception):
                        await conn.close()
            except Exception:
                log.exception("worker.listen_failed")

        listen_task = asyncio.create_task(_listen_task())

        try:
            while not self._stop.is_set():
                with suppress(asyncio.TimeoutError):
                    await asyncio.wait_for(wake.wait(), timeout=self.poll_interval_s)
                wake.clear()

                async with session_factory() as session:
                    while await process_one(
                        session, provider=self.provider, worker_id=self.worker_id
                    ):
                        await session.commit()
                        async with session_factory() as s2:
                            if not await process_one(
                                s2, provider=self.provider, worker_id=self.worker_id
                            ):
                                break
                            await s2.commit()

        finally:
            listen_task.cancel()
            with suppress(asyncio.CancelledError):
                await listen_task


__all__ = ["CorrectionWorker", "process_one"]
