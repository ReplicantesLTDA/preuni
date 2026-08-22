"""Minimal SMTP email delivery for verification links.

Dev fallback (no SMTP_HOST configured): emit verification link to the
structured log so developers can copy it from `docker compose logs api`.
"""

from __future__ import annotations

import logging
import os
import smtplib
from email.message import EmailMessage

log = logging.getLogger(__name__)


def _smtp_configured() -> bool:
    return bool(os.environ.get("SMTP_HOST"))


def send_verification_email(
    *, to_email: str, verification_token: str, base_url: str | None = None
) -> None:
    """Send (or log) the verification link to `to_email`."""
    link = (base_url or os.environ.get("PUBLIC_BASE_URL") or "http://localhost:8000").rstrip("/")
    verify_url = f"{link}/auth/verify-email?token={verification_token}"

    if not _smtp_configured():
        log.info(
            "auth.verification.dev_fallback",
            extra={"to_email_hash_only": _opaque_email_id(to_email), "verify_url": verify_url},
        )
        return

    msg = EmailMessage()
    msg["Subject"] = "Verifique seu e-mail — Corretor de Redação ENEM"
    msg["From"] = os.environ.get("SMTP_FROM", "noreply@corretor-redacao.example.com")
    msg["To"] = to_email
    msg.set_content(
        f"Bem-vindo ao Corretor de Redação ENEM!\n\n"
        f"Clique no link abaixo para verificar seu e-mail (válido por 24h):\n\n"
        f"{verify_url}\n\n"
        f"Se você não criou esta conta, ignore esta mensagem."
    )

    host = os.environ["SMTP_HOST"]
    port = int(os.environ.get("SMTP_PORT", "587"))
    username = os.environ.get("SMTP_USERNAME", "")
    password = os.environ.get("SMTP_PASSWORD", "")
    use_tls = os.environ.get("SMTP_USE_TLS", "true").lower() == "true"

    with smtplib.SMTP(host, port, timeout=10) as smtp:
        if use_tls:
            smtp.starttls()
        if username and password:
            smtp.login(username, password)
        smtp.send_message(msg)


def _opaque_email_id(email: str) -> str:
    """Stable opaque hash for logging without leaking the address (Constitution VII)."""
    import hashlib

    return hashlib.sha256(email.encode("utf-8")).hexdigest()[:12]


__all__ = ["send_verification_email"]
