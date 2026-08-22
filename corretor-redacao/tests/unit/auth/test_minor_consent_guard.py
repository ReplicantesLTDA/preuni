"""T084: minor-consent guard (Constitution X — parental consent for under-18)."""

from __future__ import annotations

import datetime as dt
import pytest
import uuid

from src.auth.consent import ParentalConsentRequiredError, ensure_consent_for_minor
from src.db.models import User
from src.db.models.enums import UserTier


def _user(birth_date: dt.date | None, consent_id: uuid.UUID | None = None) -> User:
    return User(
        id=uuid.uuid4(),
        email="test@example.com",
        password_hash="x",
        tier=UserTier.free,
        email_verified_at=dt.datetime.now(dt.UTC),
        birth_date=birth_date,
        consent_record_id=consent_id,
    )


def test_adult_passes_without_consent() -> None:
    adult = _user(birth_date=dt.date(2000, 1, 1))
    ensure_consent_for_minor(adult)


def test_no_birth_date_passes() -> None:
    no_dob = _user(birth_date=None)
    ensure_consent_for_minor(no_dob)


def test_minor_without_consent_rejected() -> None:
    today = dt.date.today()
    minor = _user(birth_date=dt.date(today.year - 15, today.month, today.day))
    with pytest.raises(ParentalConsentRequiredError):
        ensure_consent_for_minor(minor)


def test_minor_with_consent_passes() -> None:
    today = dt.date.today()
    minor = _user(
        birth_date=dt.date(today.year - 15, today.month, today.day),
        consent_id=uuid.uuid4(),
    )
    ensure_consent_for_minor(minor)


def test_user_exactly_18_today_passes() -> None:
    today = dt.date.today()
    eighteen = _user(birth_date=dt.date(today.year - 18, today.month, today.day))
    ensure_consent_for_minor(eighteen)


def test_user_one_day_under_18_rejected() -> None:
    today = dt.date.today()
    # birth_date = (today - 18 years + 1 day) → age 17 today
    target = dt.date(today.year - 18, today.month, today.day) + dt.timedelta(days=1)
    underage = _user(birth_date=target)
    with pytest.raises(ParentalConsentRequiredError):
        ensure_consent_for_minor(underage)
