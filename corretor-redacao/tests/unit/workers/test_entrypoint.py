"""Unit tests for src.workers.entrypoint._build_provider.

_main()/the __main__ guard are not covered here: they wire up a real DB
connection and run CorrectionWorker.run() forever, which belongs in a
process-level smoke test, not a unit test.
"""

from src.workers.entrypoint import _build_provider


def test_build_provider_defaults_to_local_ollama_without_api_key(monkeypatch):
    monkeypatch.delenv("OLLAMA_CLOUD_API_KEY", raising=False)
    monkeypatch.delenv("OLLAMA_BASE_URL", raising=False)
    monkeypatch.delenv("OLLAMA_MODEL", raising=False)

    provider = _build_provider()

    assert provider.base_url == "http://ollama:11434"
    assert provider.model_id == "kimi-k2:1t"
    assert provider.api_key is None


def test_build_provider_uses_ollama_cloud_when_api_key_present(monkeypatch):
    monkeypatch.setenv("OLLAMA_CLOUD_API_KEY", "secret")
    monkeypatch.delenv("OLLAMA_CLOUD_BASE_URL", raising=False)
    monkeypatch.setenv("OLLAMA_MODEL", "custom-model")

    provider = _build_provider()

    assert provider.base_url == "https://api.ollama.ai"
    assert provider.model_id == "custom-model"
    assert provider.api_key == "secret"
