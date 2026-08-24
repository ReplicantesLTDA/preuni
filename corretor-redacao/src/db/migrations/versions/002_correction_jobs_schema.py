"""correction_jobs_schema

Constitution v2.1.1 / specs/014-constitution-alignment-refactor: the correction
service stops owning identity/quota (research.md #2) and moves its pipeline
tables into a dedicated `correction` schema behind one bridge table,
`correction.correction_jobs`, that the Go monolith is granted narrow
INSERT/SELECT access to (see contracts/internal-bridge.md and
infra/migrations/grant-correction-jobs.sql for the grant itself).

Revision ID: 002_correction_jobs_schema
Revises: 001_init
Create Date: 2026-08-22

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

revision: str = '002_correction_jobs_schema'
down_revision: Union[str, Sequence[str], None] = '001_init'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    """Upgrade schema."""
    op.execute("CREATE SCHEMA IF NOT EXISTS correction")

    # Drop identity/quota tables — the Go monolith is now the sole owner of
    # this data (constitution Architecture section; research.md #2).
    op.drop_index('email_verification_tokens_user_idx', table_name='email_verification_tokens', postgresql_where=sa.text('used_at IS NULL'))
    op.drop_table('email_verification_tokens')
    op.drop_index('refresh_tokens_user_active_idx', table_name='refresh_tokens', postgresql_where=sa.text('revoked_at IS NULL'))
    op.drop_table('refresh_tokens')
    op.drop_index('consent_records_user_idx', table_name='consent_records', postgresql_where=sa.text('revoked_at IS NULL'))
    op.drop_table('consent_records')

    # corrections.user_id becomes an opaque reference — no local FK to a
    # users table this service no longer owns.
    op.drop_constraint('corrections_user_id_fkey', 'corrections', type_='foreignkey')
    op.drop_index('users_email_idx', table_name='users')
    op.drop_table('users')

    # The outbox/bridge table — the only surface the Go monolith touches
    # in this schema (contracts/internal-bridge.md).
    op.create_table(
        'correction_jobs',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('essay_text', sa.Text(), nullable=False),
        sa.Column('prompt_theme_title', sa.Text(), nullable=False),
        sa.Column('prompt_theme_context', sa.Text(), nullable=False),
        sa.Column('status', sa.Enum('pending', 'processing', 'completed', 'failed', name='correction_job_status'), server_default='pending', nullable=False),
        sa.Column('queued_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
        sa.Column('started_at', sa.DateTime(timezone=True), nullable=True),
        sa.Column('completed_at', sa.DateTime(timezone=True), nullable=True),
        sa.PrimaryKeyConstraint('id'),
        schema='correction',
    )
    op.create_index('correction_jobs_queue_idx', 'correction_jobs', ['queued_at'], unique=False, schema='correction', postgresql_where=sa.text("status = 'pending'"))

    # The worker's LISTEN/NOTIFY wake-up (src/workers/correction_worker.py)
    # used to fire from the Python INSERT path. Now that the Go monolith
    # writes rows here directly (no Python code in that path), the NOTIFY
    # moves to a DB trigger so the worker still wakes up immediately instead
    # of relying solely on its 5s poll backstop.
    op.execute("""
        CREATE OR REPLACE FUNCTION correction.notify_correction_queued()
        RETURNS trigger AS $$
        BEGIN
            PERFORM pg_notify('correction_queued', NEW.id::text);
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;
    """)
    op.execute("""
        CREATE TRIGGER correction_jobs_notify_queued
        AFTER INSERT ON correction.correction_jobs
        FOR EACH ROW EXECUTE FUNCTION correction.notify_correction_queued();
    """)

    # Move the existing grading-pipeline tables into the `correction` schema.
    op.execute("ALTER TABLE corrections SET SCHEMA correction")
    op.execute("ALTER TABLE correction_audit_logs SET SCHEMA correction")
    op.execute("ALTER TABLE grader_passes SET SCHEMA correction")

    # Link a completed correction back to the job that produced it.
    op.add_column('corrections', sa.Column('job_id', postgresql.UUID(), nullable=True), schema='correction')
    op.create_index('corrections_job_idx', 'corrections', ['job_id'], unique=False, schema='correction')


def downgrade() -> None:
    """Downgrade schema."""
    op.drop_index('corrections_job_idx', table_name='corrections', schema='correction')
    op.drop_column('corrections', 'job_id', schema='correction')

    op.execute("ALTER TABLE correction.grader_passes SET SCHEMA public")
    op.execute("ALTER TABLE correction.correction_audit_logs SET SCHEMA public")
    op.execute("ALTER TABLE correction.corrections SET SCHEMA public")

    op.execute("DROP TRIGGER IF EXISTS correction_jobs_notify_queued ON correction.correction_jobs")
    op.execute("DROP FUNCTION IF EXISTS correction.notify_correction_queued()")
    op.drop_index('correction_jobs_queue_idx', table_name='correction_jobs', schema='correction', postgresql_where=sa.text("status = 'pending'"))
    op.drop_table('correction_jobs', schema='correction')
    op.execute("DROP TYPE IF EXISTS correction.correction_job_status")

    op.create_table(
        'users',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('email', postgresql.CITEXT(), nullable=False),
        sa.Column('password_hash', sa.Text(), nullable=False),
        sa.Column('tier', postgresql.ENUM('free', 'premium', name='user_tier', create_type=False), nullable=False),
        sa.Column('email_verified_at', sa.DateTime(timezone=True), nullable=True),
        sa.Column('birth_date', sa.Date(), nullable=True),
        sa.Column('consent_record_id', sa.UUID(), nullable=True),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
        sa.Column('updated_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
        sa.PrimaryKeyConstraint('id'),
        sa.UniqueConstraint('email'),
    )
    op.create_index('users_email_idx', 'users', ['email'], unique=False)
    op.create_foreign_key('corrections_user_id_fkey', 'corrections', 'users', ['user_id'], ['id'], ondelete='CASCADE')

    op.create_table(
        'consent_records',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('guardian_name', sa.Text(), nullable=False),
        sa.Column('guardian_relation', sa.Text(), nullable=False),
        sa.Column('consent_text_sha256', sa.LargeBinary(length=32), nullable=False),
        sa.Column('accepted_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
        sa.Column('revoked_at', sa.DateTime(timezone=True), nullable=True),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.PrimaryKeyConstraint('id'),
    )
    op.create_index('consent_records_user_idx', 'consent_records', ['user_id'], unique=False, postgresql_where=sa.text('revoked_at IS NULL'))

    op.create_table(
        'refresh_tokens',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('token_hash', sa.LargeBinary(length=32), nullable=False),
        sa.Column('expires_at', sa.DateTime(timezone=True), nullable=False),
        sa.Column('revoked_at', sa.DateTime(timezone=True), nullable=True),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.PrimaryKeyConstraint('id'),
        sa.UniqueConstraint('token_hash'),
    )
    op.create_index('refresh_tokens_user_active_idx', 'refresh_tokens', ['user_id'], unique=False, postgresql_where=sa.text('revoked_at IS NULL'))

    op.create_table(
        'email_verification_tokens',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('token_hash', sa.LargeBinary(length=32), nullable=False),
        sa.Column('expires_at', sa.DateTime(timezone=True), nullable=False),
        sa.Column('used_at', sa.DateTime(timezone=True), nullable=True),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.PrimaryKeyConstraint('id'),
        sa.UniqueConstraint('token_hash'),
    )
    op.create_index('email_verification_tokens_user_idx', 'email_verification_tokens', ['user_id'], unique=False, postgresql_where=sa.text('used_at IS NULL'))
