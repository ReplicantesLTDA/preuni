"""Unit tests for src.workers.entrypoint.

The __main__ guard itself is not covered: `asyncio.run(_main())` at import
time belongs in a process-level smoke test, not a unit test.
"""

import pytest
from unittest.mock import AsyncMock, patch

from src.workers.entrypoint import _build_provider, _main


def test_build_provider_defaults_to_local_ollama_without_api_key(monkeypatch):
    monkeypatch.delenv("OLLAMA_CLOUD_API_KEY", raising=False)
    monkeypatch.delenv("OLLAMA_BASE_URL", raising=False)
    monkeypatch.delenv("OLLAMA_MODEL", raising=False)

    provider = _build_provider()

    assert provider.base_url == "http://ollama:11434"
    assert provider.model_id == "kimi-k2.6"
    assert provider.api_key is None


def test_build_provider_uses_ollama_cloud_when_api_key_present(monkeypatch):
    monkeypatch.setenv("OLLAMA_CLOUD_API_KEY", "secret")
    monkeypatch.delenv("OLLAMA_CLOUD_BASE_URL", raising=False)
    monkeypatch.setenv("OLLAMA_MODEL", "custom-model")

    provider = _build_provider()

    assert provider.base_url == "https://ollama.com"
    assert provider.model_id == "custom-model"
    assert provider.api_key == "secret"


@pytest.mark.asyncio
async def test_main_wires_up_and_runs_the_worker(monkeypatch):
    monkeypatch.setenv("DATABASE_URL", "postgresql+asyncpg://u:p@localhost/db")
    monkeypatch.delenv("OLLAMA_CLOUD_API_KEY", raising=False)

    with (
        patch("src.workers.entrypoint.configure_logging") as mock_configure_logging,
        patch("src.workers.entrypoint.CorrectionWorker") as mock_worker_cls,
    ):
        mock_worker = mock_worker_cls.return_value
        mock_worker.run = AsyncMock()

        await _main()

        mock_configure_logging.assert_called_once()
        mock_worker_cls.assert_called_once()
        _, kwargs = mock_worker_cls.call_args
        assert kwargs["database_url"] == "postgresql+asyncpg://u:p@localhost/db"
        mock_worker.run.assert_awaited_once()
