"""Observability and structured logging setup."""

import contextvars
import logging
import structlog
import sys
from typing import Any

# Context variables for correlation
correction_id_var: contextvars.ContextVar[str | None] = contextvars.ContextVar(
    "correction_id", default=None
)
user_id_var: contextvars.ContextVar[str | None] = contextvars.ContextVar("user_id", default=None)

# PII forbidden keys that must never appear in logs
PII_FORBIDDEN_KEYS = {
    "essay_text",
    "email",
    "name",
    "student_id",
    "user_id",
    "password",
    "refresh_token",
    "access_token",
}


class PIIScrubFilter:
    """Filter that scrubs PII from log records."""

    def __call__(self, logger: Any, method_name: str, event_dict: dict[str, Any]) -> dict[str, Any]:  # noqa: ARG002 — structlog filter signature
        """Scrub PII from event dictionary."""
        return self._scrub_dict(event_dict)

    def _scrub_dict(self, data: dict[str, Any]) -> dict[str, Any]:
        """Recursively scrub forbidden keys from a dictionary."""
        scrubbed = {}
        for key, value in data.items():
            if key in PII_FORBIDDEN_KEYS:
                raise ValueError(
                    f"Forbidden PII key in log payload: {key}. "
                    f"This key must not appear in audit logs per Constitution Article VII."
                )
            if isinstance(value, dict):
                scrubbed[key] = self._scrub_dict(value)
            elif isinstance(value, (list, tuple)):
                scrubbed[key] = [
                    self._scrub_dict(item) if isinstance(item, dict) else item for item in value
                ]
            else:
                scrubbed[key] = value
        return scrubbed


class ContextVarsFilter:
    """Add context variables to log records."""

    def __call__(self, logger: Any, method_name: str, event_dict: dict[str, Any]) -> dict[str, Any]:  # noqa: ARG002 — structlog filter signature
        """Add correction_id and user_id from context if available."""
        correction_id = correction_id_var.get()
        user_id = user_id_var.get()

        if correction_id:
            event_dict["correction_id"] = correction_id
        if user_id:
            event_dict["user_id"] = user_id

        return event_dict


def set_correction_id(correction_id: str | None) -> None:
    """Set the current correction ID in context."""
    correction_id_var.set(correction_id)


def set_user_id(user_id: str | None) -> None:
    """Set the current user ID in context."""
    user_id_var.set(user_id)


def get_correction_id() -> str | None:
    """Get the current correction ID from context."""
    return correction_id_var.get()


def get_user_id() -> str | None:
    """Get the current user ID from context."""
    return user_id_var.get()


def configure_logging(
    log_format: str = "json", log_level: str = "INFO", disable_existing: bool = True
) -> None:
    """Configure structured logging with structlog."""
    # Configure Python's standard logging
    logging.basicConfig(
        format="%(message)s",
        stream=sys.stdout,
        level=getattr(logging, log_level.upper()),
        force=True,
    )
    if disable_existing:
        for name in list(logging.root.manager.loggerDict):
            logging.getLogger(
                name
            ).disabled = False  # leave intact; force above already reset handlers

    # Configure structlog
    processors: list[Any] = [
        ContextVarsFilter(),
        PIIScrubFilter(),
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.dev.set_exc_info,
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
    ]

    if log_format == "json":
        processors.append(structlog.processors.JSONRenderer())
    else:
        processors.extend(
            [
                structlog.dev.ConsoleRenderer(),
            ]
        )

    structlog.configure(
        processors=processors,
        context_class=dict,
        logger_factory=structlog.PrintLoggerFactory(),
        cache_logger_on_first_use=False,
    )


def get_logger(name: str) -> structlog.BoundLogger:
    """Get a configured logger instance."""
    return structlog.get_logger(name)


# Initialize logger
logger = get_logger(__name__)
