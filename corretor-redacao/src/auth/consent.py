"""Parental-consent guard (Constitution X — minor protection).

Refuses correction-endpoint access for users under 18 without an active
parental consent record. Consent state is captured during account onboarding;
this module only checks it at request time.
"""

from __future__ import annotations

import datetime as dt

from src.db.models import User


class ParentalConsentRequiredError(Exception):
    """Spec error_code: parental_consent_required."""


def _age_in_years(birth_date: dt.date, today: dt.date | None = None) -> int:
    today = today or dt.date.today()
    years = today.year - birth_date.year
    # Subtract a year if birthday hasn't happened yet this year.
    if (today.month, today.day) < (birth_date.month, birth_date.day):
        years -= 1
    return years


def ensure_consent_for_minor(user: User, *, today: dt.date | None = None) -> None:
    """Raise if the user is under 18 and has no consent record on file."""
    if user.birth_date is None:
        return  # Birth date unknown → trust onboarding flow handled it.
    if _age_in_years(user.birth_date, today=today) >= 18:
        return
    if user.consent_record_id is None:
        raise ParentalConsentRequiredError(
            "user is under 18 and no parental consent record is on file"
        )


__all__ = ["ParentalConsentRequiredError", "ensure_consent_for_minor"]
