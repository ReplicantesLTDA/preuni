-- Migration: 002_create_track_enrollments
-- Schema: users

CREATE TABLE users.track_enrollments (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id      UUID        NOT NULL REFERENCES users.students(id) ON DELETE CASCADE,
    track_id        UUID        NOT NULL,  -- references content-svc; no FK (cross-service)
    enrolled_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    unenrolled_at   TIMESTAMPTZ           -- NULL = currently enrolled
);

-- Only one active enrollment per track per student
CREATE UNIQUE INDEX idx_track_enrollments_active
    ON users.track_enrollments (student_id, track_id)
    WHERE unenrolled_at IS NULL;

CREATE INDEX idx_track_enrollments_student ON users.track_enrollments (student_id);
