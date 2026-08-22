"""Password hashing + OWASP policy validation.

Argon2id is the OWASP-recommended algorithm for new applications (Research R6).
`argon2-cffi` exposes the canonical Python binding with manylinux wheels —
no C compile pain in the docker image.

Policy (Constitution III + FR-028):
- Min length 12 (OWASP).
- Reject against local leaked-password list (small offline corpus; pwned-passwords
  API integration is a post-MVP enhancement).
"""

from __future__ import annotations

from argon2 import PasswordHasher
from argon2.exceptions import InvalidHashError, VerifyMismatchError
from functools import cache
from pathlib import Path

MIN_PASSWORD_LENGTH = 12
LEAKED_PASSWORDS_PATH = Path(__file__).parent / "leaked_passwords.txt"


class PasswordPolicyError(ValueError):
    """Base — raised when a password violates project policy."""


class PasswordTooShortError(PasswordPolicyError):
    """Spec error_code component: password below OWASP minimum length."""


class LeakedPasswordError(PasswordPolicyError):
    """Spec error_code component: password appears in the known-leaked list."""


@cache
def _hasher() -> PasswordHasher:
    # Defaults tuned to ~50 ms on modern CPU per OWASP recommendation.
    return PasswordHasher(
        time_cost=3,
        memory_cost=64 * 1024,
        parallelism=2,
    )


@cache
def _leaked_set() -> frozenset[str]:
    if not LEAKED_PASSWORDS_PATH.is_file():
        return frozenset()
    return frozenset(
        line.strip()
        for line in LEAKED_PASSWORDS_PATH.read_text(encoding="utf-8").splitlines()
        if line.strip() and not line.startswith("#")
    )


def hash_password(plain: str) -> str:
    return _hasher().hash(plain)


def verify_password(plain: str, hashed: str) -> bool:
    try:
        return _hasher().verify(hashed, plain)
    except (VerifyMismatchError, InvalidHashError):
        return False
    except Exception:
        return False


def validate_password_policy(plain: str) -> None:
    """Raise PasswordPolicyError subclass if the password violates project policy."""
    if len(plain) < MIN_PASSWORD_LENGTH:
        raise PasswordTooShortError(
            f"password must be at least {MIN_PASSWORD_LENGTH} characters (OWASP); got {len(plain)}."
        )
    if plain in _leaked_set():
        raise LeakedPasswordError("password appears in known-leaked list; choose a different one.")


__all__ = [
    "MIN_PASSWORD_LENGTH",
    "LeakedPasswordError",
    "PasswordPolicyError",
    "PasswordTooShortError",
    "hash_password",
    "validate_password_policy",
    "verify_password",
]
