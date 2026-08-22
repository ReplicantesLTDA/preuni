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


class JWTSettings(BaseSettings):
    """JWT authentication configuration."""

    secret_key: str
    algorithm: str = "HS256"
    access_token_expire_minutes: int = 15
    refresh_token_expire_days: int = 7

    model_config = SettingsConfigDict(env_prefix="JWT_", case_sensitive=False)


class SMTPSettings(BaseSettings):
    """SMTP email configuration."""

    host: str = ""
    port: int = 587
    username: str = ""
    password: str = ""
    from_email: str = "noreply@corretor-redacao.example.com"
    use_tls: bool = True

    model_config = SettingsConfigDict(env_prefix="SMTP_", case_sensitive=False)

    @property
    def is_configured(self) -> bool:
        """Check if SMTP is configured."""
        return bool(self.host and self.username)


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


class QuotaSettings(BaseSettings):
    """Quota configuration."""

    free_tier_monthly_quota: int = 3
    premium_tier_monthly_quota: int = 30
    quota_reset_utc_day: int = 1
    quota_reset_utc_hour: int = 0

    model_config = SettingsConfigDict(env_prefix="", case_sensitive=False)


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
    """Main application settings."""

    app_name: str = "Corretor Redação"
    app_version: str = "0.1.0"
    debug: bool = False
    testing: bool = False

    # Sub-configurations
    database: DatabaseSettings = DatabaseSettings()
    jwt: JWTSettings = JWTSettings()
    smtp: SMTPSettings = SMTPSettings()
    llm: LLMSettings = LLMSettings()
    prompt: PromptSettings = PromptSettings()
    essay_validation: EssayValidationSettings = EssayValidationSettings()
    quota: QuotaSettings = QuotaSettings()
    observability: ObservabilitySettings = ObservabilitySettings()

    # API Configuration
    api_host: str = "0.0.0.0"
    api_port: int = 8000
    api_workers: int = 4
    api_log_level: str = "INFO"

    # Features
    feature_email_verification: bool = True
    feature_parental_consent: bool = True
    feature_reevaluation: bool = True
    feature_golden_metrics_only: bool = False

    # LGPD & Privacy
    zero_data_retention: bool = True
    allow_data_training: bool = False
    user_deletion_grace_period_days: int = 15

    # Worker Configuration
    worker_enabled: bool = True
    worker_poll_interval_seconds: int = 5
    worker_batch_size: int = 5
    worker_log_level: str = "INFO"

    # CORS
    cors_origins: list[str] = [
        "http://localhost:3000",
        "http://localhost:8080",
    ]
    cors_allow_credentials: bool = True
    cors_allow_methods: list[str] = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    cors_allow_headers: list[str] = ["Authorization", "Content-Type"]

    # Rate Limiting
    rate_limit_enabled: bool = True
    rate_limit_requests_per_minute: int = 60
    rate_limit_auth_requests_per_minute: int = 5

    # Backup
    backup_enabled: bool = False
    backup_s3_endpoint: str = ""
    backup_s3_access_key: str = ""
    backup_s3_secret_key: str = ""
    backup_s3_bucket: str = ""
    backup_s3_region: str = "us-east-1"

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
