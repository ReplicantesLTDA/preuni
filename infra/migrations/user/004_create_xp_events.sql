-- Migration: 004_create_xp_events
-- Schema: users
-- Description: Append-only XP event log; students.xp_total is the materialized sum

CREATE TYPE users.xp_source AS ENUM (
    'LESSON_COMPLETE',
    'REVIEW_SESSION',
    'SIMULATION_COMPLETE',
    'BONUS'
);

CREATE TABLE users.xp_events (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id  UUID        NOT NULL REFERENCES users.students(id) ON DELETE CASCADE,
    amount      INT         NOT NULL CHECK (amount > 0),
    source      users.xp_source NOT NULL,
    source_id   UUID,                        -- ID of the triggering entity (nullable)
    earned_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_xp_events_student ON users.xp_events (student_id, earned_at DESC);
