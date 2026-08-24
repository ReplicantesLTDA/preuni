"""Async SQLAlchemy engine + sessionmaker factory.

Layer-separation rule (Constitution VI): this module is the ONLY place that
constructs an AsyncEngine. Callers in `api/` and `workers/` get sessions via
`src.db.session.get_session` — never instantiate engines directly.
"""

from __future__ import annotations

import os
from functools import lru_cache
from sqlalchemy.ext.asyncio import (
    AsyncEngine,
    AsyncSession,
    async_sessionmaker,
    create_async_engine,
)

DEFAULT_DATABASE_URL = "postgresql+asyncpg://corretor:corretor@localhost:5432/corretor_db"


def _resolve_database_url(override: str | None = None) -> str:
    if override:
        return override
    return os.environ.get("DATABASE_URL") or DEFAULT_DATABASE_URL


@lru_cache(maxsize=4)
def _engine_singleton(url: str, *, echo: bool, pool_size: int) -> AsyncEngine:
    return create_async_engine(
        url,
        echo=echo,
        pool_size=pool_size,
        pool_pre_ping=True,
        pool_recycle=int(os.environ.get("DATABASE_POOL_RECYCLE", "3600")),
        future=True,
    )


def get_engine(url: str | None = None) -> AsyncEngine:
    """Process-wide async engine, memoized per URL."""
    resolved = _resolve_database_url(url)
    echo = os.environ.get("DATABASE_ECHO", "false").lower() == "true"
    pool_size = int(os.environ.get("DATABASE_POOL_SIZE", "10"))
    return _engine_singleton(resolved, echo=echo, pool_size=pool_size)


def session_factory(engine: AsyncEngine | None = None) -> async_sessionmaker[AsyncSession]:
    """Build an AsyncSession factory bound to the given engine (or default)."""
    return async_sessionmaker(
        engine or get_engine(),
        expire_on_commit=False,
        autoflush=False,
        autocommit=False,
    )


__all__ = [
    "DEFAULT_DATABASE_URL",
    "get_engine",
    "session_factory",
]
