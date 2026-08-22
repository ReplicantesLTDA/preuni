-- Migration: 009_add_streak_gamification_columns
-- Schema: users
-- Description: Constitution v2.1.1 pivot (specs/014-constitution-alignment-refactor).
--   Reuses the existing streak_count / streak_last_active_date columns as the
--   UTC-day-boundary streak (data-model.md "users" entity); adds the two
--   fields that table didn't already have.

ALTER TABLE users.students
    ADD COLUMN longest_streak     INT         NOT NULL DEFAULT 0 CHECK (longest_streak >= 0),
    ADD COLUMN subscription_tier  VARCHAR(10) NOT NULL DEFAULT 'free' CHECK (subscription_tier IN ('free', 'pro'));
