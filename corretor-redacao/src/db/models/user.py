"""User table — end-user student account.

Constitution X (B2C only): the User is the platform's only consumer class.
Tier governs quotas; consent_record_id is required at the app layer when
birth_date implies the user is under 18.
"""

from __future__ import annotations

import datetime as dt
import uuid
from sqlalchemy import (
    Date,
    DateTime,
    Enum,
    ForeignKey,
    Index,
    Text,
    func,
)
from sqlalchemy.dialects.postgresql import CITEXT, UUID
from sqlalchemy.orm import Mapped, mapped_column

from src.db.models import Base
from src.db.models.enums import UserTier


class User(Base):
    __tablename__ = "users"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True)
    email: Mapped[str] = mapped_column(CITEXT(), nullable=False, unique=True)
    password_hash: Mapped[str] = mapped_column(Text, nullable=False)
    tier: Mapped[UserTier] = mapped_column(
        Enum(UserTier, name="user_tier", native_enum=True),
        nullable=False,
        default=UserTier.free,
    )
    email_verified_at: Mapped[dt.datetime | None] = mapped_column(
        DateTime(timezone=True), nullable=True
    )
    birth_date: Mapped[dt.date | None] = mapped_column(Date, nullable=True)
    consent_record_id: Mapped[uuid.UUID | None] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey(
            "consent_records.id",
            ondelete="SET NULL",
            use_alter=True,
            name="users_consent_record_id_fkey",
        ),
        nullable=True,
    )
    created_at: Mapped[dt.datetime] = mapped_column(
        DateTime(timezone=True), nullable=False, server_default=func.now()
    )
    updated_at: Mapped[dt.datetime] = mapped_column(
        DateTime(timezone=True),
        nullable=False,
        server_default=func.now(),
        onupdate=func.now(),
    )

    __table_args__ = (Index("users_email_idx", "email"),)
