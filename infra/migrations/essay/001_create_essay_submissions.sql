-- Migration: 001_create_essay_submissions
-- Schema: essay
-- Description: Essay submissions (quota gate + streak trigger) and their
--   graded read model. See specs/014-constitution-alignment-refactor/data-model.md.

CREATE SCHEMA IF NOT EXISTS essay;

CREATE TYPE essay.submission_status AS ENUM ('pending', 'graded', 'failed');

CREATE TABLE essay.essay_submissions (
    id                   UUID        PRIMARY KEY,
    user_id              UUID        NOT NULL REFERENCES users.students (id) ON DELETE CASCADE,
    prompt_theme_title   TEXT        NOT NULL,
    prompt_theme_context TEXT        NOT NULL,
    essay_text           TEXT        NOT NULL,
    submission_day       DATE        NOT NULL, -- UTC calendar day; quota key
    status               essay.submission_status NOT NULL DEFAULT 'pending',
    correction_job_id    UUID,                 -- FK lives in the correction service's own schema; opaque reference here
    submitted_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Free-tier quota: at most one submission per user per UTC day. Enforced
-- at the application layer against the caller's subscription_tier (a
-- partial index alone can't see the joined tier), but this index makes the
-- common lookup ("has this user submitted today?") cheap either way.
CREATE INDEX essay_submissions_user_day_idx
    ON essay.essay_submissions (user_id, submission_day);

CREATE INDEX essay_submissions_user_listing_idx
    ON essay.essay_submissions (user_id, submitted_at DESC);

-- Reconciliation poller: find submissions still waiting on a job result.
CREATE INDEX essay_submissions_pending_idx
    ON essay.essay_submissions (correction_job_id)
    WHERE status = 'pending' AND correction_job_id IS NOT NULL;

CREATE TABLE essay.essay_grades (
    submission_id   UUID        PRIMARY KEY REFERENCES essay.essay_submissions (id) ON DELETE CASCADE,
    overall_score   SMALLINT    NOT NULL CHECK (overall_score BETWEEN 0 AND 1000),
    competencies    JSONB       NOT NULL, -- array of 5: {competency, score, justification_pt_br, excerpt}
    graded_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
