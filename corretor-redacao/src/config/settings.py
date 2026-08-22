"""Application settings and configuration management."""

from pydantic_settings import BaseSettings, SettingsConfigDict
from typing import Literal


class DatabaseSettings(BaseSettings):
    """Database configuration."""

    url: str
    pool_size: int = 20
    pool_recycle: int = 3600
    pool_pre_ping: bool = True
    echo: bool = False

    model_config = SettingsConfigDict(env_prefix="DATABASE_", case_sensitive=False)


class LLMSettings(BaseSettings):
    """LLM provider configuration."""

    provider: Literal["ollama", "anthropic", "openai"] = "ollama"
    ollama_base_url: str = "http://ollama:11434"
    ollama_model: str = "kimi-k2:1t"
    ollama_api_key: str = ""
    ollama_cloud_api_key: str = ""
    ollama_cloud_base_url: str = "https://api.ollama.ai"
    temperature: float = 0.1

    model_config = SettingsConfigDict(env_prefix="", case_sensitive=False, extra="ignore")


class PromptSettings(BaseSettings):
    """Prompt configuration."""

    version: str = "v1.0.0"
    temperature: float = 0.1

    model_config = SettingsConfigDict(env_prefix="PROMPT_", case_sensitive=False)


class EssayValidationSettings(BaseSettings):
    """Essay validation settings."""

    min_length: int = 500
    max_length: int = 3500
    min_lines: int = 7
    max_lines: int = 50

    model_config = SettingsConfigDict(env_prefix="ESSAY_", case_sensitive=False)


class ObservabilitySettings(BaseSettings):
    """Observability and logging configuration."""

    log_level: str = "INFO"
    log_format: Literal["json", "text"] = "json"
    otel_enabled: bool = False
    otel_exporter_otlp_endpoint: str = "http://localhost:4317"
    prometheus_enabled: bool = True
    pii_scrub_enabled: bool = True
    audit_log_enabled: bool = True
    metrics_enabled: bool = True

    model_config = SettingsConfigDict(case_sensitive=False)


class AppSettings(BaseSettings):
    """Main application settings.

    Constitution v2.1.1 / specs/014-constitution-alignment-refactor: this
    service is internal-only now (research.md #2, #3) — JWT, SMTP, quota,
    and end-user-facing feature flags/CORS/rate-limiting were all removed
    as dead config once the Go monolith took over identity/quota/email.
    """

    app_name: str = "Corretor Redação"
    app_version: str = "0.1.0"
    debug: bool = False
    testing: bool = False

    # Sub-configurations
    database: DatabaseSettings = DatabaseSettings()
    llm: LLMSettings = LLMSettings()
    prompt: PromptSettings = PromptSettings()
    essay_validation: EssayValidationSettings = EssayValidationSettings()
    observability: ObservabilitySettings = ObservabilitySettings()

    # API Configuration
    api_host: str = "0.0.0.0"
    api_port: int = 8000
    api_workers: int = 4
    api_log_level: str = "INFO"

    feature_reevaluation: bool = True
    feature_golden_metrics_only: bool = False

    # Worker Configuration
    worker_enabled: bool = True
    worker_poll_interval_seconds: int = 5
    worker_batch_size: int = 5
    worker_log_level: str = "INFO"

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
        extra="ignore",
    )

    @property
    def database_url(self) -> str:
        """Get the database URL."""
        return self.database.url

    @property
    def is_production(self) -> bool:
        """Check if running in production."""
        return not self.debug and not self.testing


# Global settings instance
settings = AppSettings()
