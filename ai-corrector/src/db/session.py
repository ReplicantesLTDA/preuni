"""Async session dependency for FastAPI routes and worker code paths.

Usage (FastAPI):
    from fastapi import Depends
    from src.db.session import get_session

    @router.get("/...")
    async def handler(session: AsyncSession = Depends(get_session)):
        ...

Usage (worker, manual):
    async for session in get_session():
        ...
"""

from __future__ import annotations

from collections.abc import AsyncIterator
from sqlalchemy.ext.asyncio import AsyncEngine, AsyncSession

from src.db.engine import session_factory


async def get_session(engine: AsyncEngine | None = None) -> AsyncIterator[AsyncSession]:
    """Yield a transactional AsyncSession.

    Commits on clean exit; rolls back on exception. The session is closed in
    `finally` regardless of outcome.
    """
    Session = session_factory(engine)  # noqa: N806 — sessionmaker is a class-like factory
    session: AsyncSession = Session()
    try:
        yield session
        await session.commit()
    except Exception:
        await session.rollback()
        raise
    finally:
        await session.close()


__all__ = ["get_session"]
