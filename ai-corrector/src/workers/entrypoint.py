"""Worker process entrypoint — instantiates CorrectionWorker and runs it."""

from __future__ import annotations

import asyncio
import logging
import os

from src.corrector.llm.ollama import OllamaProvider
from src.observability.logging import configure_logging
from src.workers.correction_worker import CorrectionWorker

log = logging.getLogger(__name__)


def _build_provider() -> OllamaProvider:
    api_key = os.environ.get("OLLAMA_CLOUD_API_KEY", "")
    base_url = os.environ.get(
        "OLLAMA_CLOUD_BASE_URL" if api_key else "OLLAMA_BASE_URL",
        "https://ollama.com" if api_key else "http://ollama:11434",
    )
    model = os.environ.get("OLLAMA_MODEL", "kimi-k2.6")
    return OllamaProvider(base_url=base_url, model_id=model, api_key=api_key or None)


async def _main() -> None:
    configure_logging()
    database_url = os.environ["DATABASE_URL"]
    provider = _build_provider()
    worker = CorrectionWorker(database_url=database_url, provider=provider)
    log.info("worker.entrypoint.start")
    await worker.run()


if __name__ == "__main__":
    asyncio.run(_main())
