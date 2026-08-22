"""T068: password hashing + OWASP rule enforcement."""

from __future__ import annotations

import pytest

from src.auth.passwords import (
    LeakedPasswordError,
    PasswordPolicyError,
    PasswordTooShortError,
    hash_password,
    validate_password_policy,
    verify_password,
)


def test_hash_verify_roundtrip() -> None:
    h = hash_password("correctHorseBatteryStaple12!")
    assert h.startswith("$argon2")
    assert verify_password("correctHorseBatteryStaple12!", h) is True


def test_verify_rejects_wrong_password() -> None:
    h = hash_password("correctHorseBatteryStaple12!")
    assert verify_password("wrong", h) is False


def test_verify_rejects_corrupt_hash() -> None:
    assert verify_password("any", "$argon2id$bogus") is False


def test_owasp_min_length_12() -> None:
    with pytest.raises(PasswordTooShortError):
        validate_password_policy("short")
    with pytest.raises(PasswordTooShortError):
        validate_password_policy("a" * 11)
    validate_password_policy("a" * 12)  # exactly at minimum


def test_leaked_password_rejected() -> None:
    # Common known-leaked passwords from the local list.
    with pytest.raises(LeakedPasswordError):
        validate_password_policy("password123!")
    with pytest.raises(LeakedPasswordError):
        validate_password_policy("123456789012")


def test_policy_accepts_strong_passwords() -> None:
    validate_password_policy("correctHorseBatteryStaple12!")
    validate_password_policy("MinhaSenhaForte#2026")


def test_policy_error_hierarchy() -> None:
    assert issubclass(PasswordTooShortError, PasswordPolicyError)
    assert issubclass(LeakedPasswordError, PasswordPolicyError)
