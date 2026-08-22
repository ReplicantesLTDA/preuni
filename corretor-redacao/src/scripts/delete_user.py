"""LGPD 15-day deletion path — cascade delete user + write audit trail.

Usage:
    python -m src.scripts.delete_user --user-id <uuid> --reason <text> [--operator <str>]

Cascade order (before DB ON DELETE CASCADE fires):
  1. Write `deleted` audit event for each owned correction.
  2. Session commit → DB cascades: corrections, grader_passes, audit_logs,
     refresh_tokens, verification_tokens, consent_records, then user row.
"""

from __future__ import annotations

import argparse
import asyncio
import logging
import os
import uuid
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from src.db.models import Correction, User
from src.db.models.enums import AuditEventType
from src.db.repositories.audit_log_repo import AuditLogWriter

log = logging.getLogger(__name__)


async def delete_user(
    session: AsyncSession,
    *,
    user_id: uuid.UUID,
    reason: str,
    operator: str = "system",
) -> None:
    """Delete a user and all their data. Writes audit events before cascade."""
    user = (await session.execute(select(User).where(User.id == user_id))).scalar_one_or_none()

    if user is None:
        raise ValueError(f"User {user_id} not found")

    corrections = (
        (await session.execute(select(Correction).where(Correction.user_id == user_id)))
        .scalars()
        .all()
    )

    writer = AuditLogWriter(session)

    for correction in corrections:
        await writer.write(
            correction_id=correction.id,
            input_hash=correction.input_hash,
            event_type=AuditEventType.deleted,
            payload={"deletion_reason": reason},
        )

    log.info(
        "delete_user.cascade",
        extra={
            "user_id_hash": str(user_id)[:8],
            "correction_count": len(corrections),
            "operator": operator,
            "reason": reason,
        },
    )

    await session.delete(user)


async def _main() -> None:
    parser = argparse.ArgumentParser(description="Delete a user and all their data (LGPD)")
    parser.add_argument("--user-id", required=True, help="UUID of the user to delete")
    parser.add_argument("--reason", required=True, help="Deletion reason (logged)")
    parser.add_argument(
        "--operator", default="operator", help="Identity of operator performing deletion"
    )
    args = parser.parse_args()

    from src.observability.logging import configure_logging

    configure_logging()

    database_url = os.environ["DATABASE_URL"]
    engine = create_async_engine(database_url)
    session_factory = async_sessionmaker(engine, expire_on_commit=False)

    try:
        async with session_factory() as session:
            await delete_user(
                session,
                user_id=uuid.UUID(args.user_id),
                reason=args.reason,
                operator=args.operator,
            )
            await session.commit()
        print(f"Deleted user {args.user_id}")
    finally:
        await engine.dispose()


if __name__ == "__main__":
    asyncio.run(_main())
