-- Migration: 001_create_students
-- Schema: users
-- Description: Student profile table

CREATE SCHEMA IF NOT EXISTS users;

CREATE TABLE users.students (
    id                      UUID        PRIMARY KEY,  -- same UUID as auth.credentials.id
    display_name            VARCHAR(64) NOT NULL,
    username                VARCHAR(30),              -- set during onboarding
    avatar_url              TEXT,
    xp_total                INT         NOT NULL DEFAULT 0 CHECK (xp_total >= 0),
    streak_count            INT         NOT NULL DEFAULT 0 CHECK (streak_count >= 0),
    streak_last_active_date DATE,
    readiness_score         SMALLINT    CHECK (readiness_score BETWEEN 0 AND 100),
    onboarding_completed    BOOLEAN     NOT NULL DEFAULT false,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Unique username index (case-sensitive; usernames are always lowercase by domain rule)
CREATE UNIQUE INDEX idx_students_username ON users.students (username)
    WHERE username IS NOT NULL;

CREATE OR REPLACE FUNCTION users.set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_students_updated_at
    BEFORE UPDATE ON users.students
    FOR EACH ROW EXECUTE FUNCTION users.set_updated_at();
