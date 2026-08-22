"""T057: ORM models smoke — every entity from data-model.md is importable,
has the right tablename, and exposes the columns the data model specifies.
"""

from __future__ import annotations

import pytest


def test_base_exposes_metadata() -> None:
    from src.db.models import Base

    assert Base.metadata is not None
    assert len(Base.metadata.tables) > 0


@pytest.mark.parametrize(
    "module_path, class_name, tablename, required_columns",
    [
        (
            "src.db.models.user",
            "User",
            "users",
            {
                "id",
                "email",
                "password_hash",
                "tier",
                "email_verified_at",
                "created_at",
                "updated_at",
            },
        ),
        (
            "src.db.models.consent_record",
            "ConsentRecord",
            "consent_records",
            {
                "id",
                "user_id",
                "guardian_name",
                "guardian_relation",
                "consent_text_sha256",
                "accepted_at",
            },
        ),
        (
            "src.db.models.refresh_token",
            "RefreshToken",
            "refresh_tokens",
            {"id", "user_id", "token_hash", "expires_at", "created_at"},
        ),
        (
            "src.db.models.email_verification_token",
            "EmailVerificationToken",
            "email_verification_tokens",
            {"id", "user_id", "token_hash", "expires_at", "created_at"},
        ),
        (
            "src.db.models.correction",
            "Correction",
            "corrections",
            {
                "id",
                "user_id",
                "essay_text",
                "prompt_theme_title",
                "prompt_theme_context",
                "input_hash",
                "status",
                "queued_at",
                "final_score",
                "c1_score",
                "c2_score",
                "c3_score",
                "c4_score",
                "c5_score",
                "competencies",
                "eliminatory_flags",
                "prompt_version",
                "model_identifier",
                "output_schema_version",
                "quota_consumed",
            },
        ),
        (
            "src.db.models.grader_pass",
            "GraderPass",
            "grader_passes",
            {
                "id",
                "correction_id",
                "pass_index",
                "c1_score",
                "c2_score",
                "c3_score",
                "c4_score",
                "c5_score",
                "competencies",
                "prompt_version",
                "model_identifier",
                "output_schema_version",
                "raw_output",
                "latency_ms",
                "created_at",
            },
        ),
        (
            "src.db.models.correction_audit_log",
            "CorrectionAuditLog",
            "correction_audit_logs",
            {"id", "correction_id", "input_hash", "event_type", "event_payload", "created_at"},
        ),
    ],
)
def test_model_shape(
    module_path: str, class_name: str, tablename: str, required_columns: set[str]
) -> None:
    module = __import__(module_path, fromlist=[class_name])
    cls = getattr(module, class_name)
    assert cls.__tablename__ == tablename, f"{class_name}.__tablename__ = {cls.__tablename__!r}"
    actual = {col.name for col in cls.__table__.columns}
    missing = required_columns - actual
    assert not missing, f"{class_name} missing columns: {missing}"


def test_all_models_registered_under_base() -> None:
    from src.db.models import Base

    expected = {
        "users",
        "consent_records",
        "refresh_tokens",
        "email_verification_tokens",
        "corrections",
        "grader_passes",
        "correction_audit_logs",
    }
    actual = set(Base.metadata.tables.keys())
    missing = expected - actual
    assert not missing, f"missing tables in Base.metadata: {missing}"
