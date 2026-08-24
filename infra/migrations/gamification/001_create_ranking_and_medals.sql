-- Migration: 001_create_ranking_and_medals
-- Schema: gamification
-- Description: Weekly global leaderboard with league tiers, and medals.
--   See specs/014-constitution-alignment-refactor/data-model.md.

CREATE SCHEMA IF NOT EXISTS gamification;

CREATE TYPE gamification.league_tier AS ENUM ('bronze', 'silver', 'gold', 'platinum', 'diamond');
CREATE TYPE gamification.medal_type AS ENUM (
    'streak_7_day', 'streak_30_day', 'streak_100_day',
    'tier_promotion', 'weekly_top_finish'
);

CREATE TABLE gamification.weekly_ranking_entries (
    id             UUID        PRIMARY KEY,
    user_id        UUID        NOT NULL REFERENCES users.students (id) ON DELETE CASCADE,
    week_start     DATE        NOT NULL, -- UTC Monday
    weekly_score   INT         NOT NULL DEFAULT 0 CHECK (weekly_score >= 0),
    league_tier    gamification.league_tier NOT NULL DEFAULT 'bronze',
    rank_in_tier   INT,                  -- computed at week close; NULL until then
    UNIQUE (user_id, week_start)
);

CREATE INDEX weekly_ranking_leaderboard_idx
    ON gamification.weekly_ranking_entries (week_start, league_tier, weekly_score DESC);

CREATE TABLE gamification.medals (
    id         UUID        PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users.students (id) ON DELETE CASCADE,
    type       gamification.medal_type NOT NULL,
    earned_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX medals_user_idx ON gamification.medals (user_id, earned_at DESC);
