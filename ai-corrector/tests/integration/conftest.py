"""Shared fixtures for integration tests against real Postgres.

Skips integration tests gracefully when TEST_DATABASE_URL is not set or the
target Postgres is unreachable. CI / local dev provides Postgres via
docker-compose (TEST_DATABASE_URL defaults to a `corretor_test` database on
the dev compose stack).
"""

from __future__ import annotations

import os
import pytest
import pytest_asyncio
from collections.abc import AsyncIterator
from sqlalchemy.ext.asyncio import (
    AsyncEngine,
    AsyncSession,
    async_sessionmaker,
    create_async_engine,
)

DEFAULT_TEST_DB_URL = "postgresql+asyncpg://corretor:corretor@localhost:5432/corretor_test"


def _test_db_url() -> str:
    return os.environ.get("TEST_DATABASE_URL", DEFAULT_TEST_DB_URL)


async def _ensure_test_db_reachable(url: str) -> bool:
    try:
        engine = create_async_engine(url, pool_pre_ping=True)
        async with engine.connect() as conn:
            await conn.execute(__import__("sqlalchemy").text("SELECT 1"))
        await engine.dispose()
        return True
    except Exception:
        return False


# Function-scoped: pytest-asyncio creates a fresh event loop per test, and
# asyncpg connections refuse to be reused across loops. Session scope on the
# engine would crash on the 2nd test; function scope is correct here.
@pytest_asyncio.fixture
async def test_engine() -> AsyncIterator[AsyncEngine]:
    url = _test_db_url()
    reachable = await _ensure_test_db_reachable(url)
    if not reachable:
        pytest.skip(f"Postgres unreachable at {url}; skipping integration tests")
    engine = create_async_engine(url, echo=False, pool_pre_ping=True)
    yield engine
    await engine.dispose()


@pytest_asyncio.fixture
async def db_session(test_engine: AsyncEngine) -> AsyncIterator[AsyncSession]:
    """Per-test session that rolls back at teardown (no test pollution)."""
    Session = async_sessionmaker(test_engine, expire_on_commit=False)  # noqa: N806
    async with Session() as session:
        try:
            yield session
        finally:
            await session.rollback()
            await session.close()
