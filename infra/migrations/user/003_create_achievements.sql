-- Migration: 003_create_achievements
-- Schema: users

CREATE TYPE users.achievement_condition AS ENUM (
    'STREAK_DAYS',
    'SIMULATIONS_COMPLETED',
    'TRACK_COMPLETED',
    'XP_EARNED',
    'LESSONS_COMPLETED'
);

CREATE TABLE users.achievements (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(64) NOT NULL UNIQUE,   -- e.g., STREAK_7, FIRST_SIMULATION
    name            VARCHAR(128) NOT NULL,
    description     TEXT        NOT NULL,
    icon_url        TEXT        NOT NULL,
    condition_type  users.achievement_condition NOT NULL,
    condition_value INT         NOT NULL
);

CREATE TABLE users.student_achievements (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id      UUID        NOT NULL REFERENCES users.students(id) ON DELETE CASCADE,
    achievement_id  UUID        NOT NULL REFERENCES users.achievements(id),
    unlocked_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (student_id, achievement_id)
);
