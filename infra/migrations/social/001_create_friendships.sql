-- Migration: 001_create_friendships
-- Schema: social
-- Description: Friend requests/connections gating streak + latest-grade
--   visibility. See specs/014-constitution-alignment-refactor/data-model.md.

CREATE SCHEMA IF NOT EXISTS social;

CREATE TYPE social.friendship_status AS ENUM ('pending', 'accepted', 'removed');

CREATE TABLE social.friendships (
    id            UUID        PRIMARY KEY,
    requester_id  UUID        NOT NULL REFERENCES users.students (id) ON DELETE CASCADE,
    addressee_id  UUID        NOT NULL REFERENCES users.students (id) ON DELETE CASCADE,
    status        social.friendship_status NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    responded_at  TIMESTAMPTZ,
    CONSTRAINT friendships_no_self CHECK (requester_id <> addressee_id)
);

-- One relationship per unordered pair — re-requesting after removal creates
-- a new row only if no pending/accepted row already exists for the pair
-- (enforced at the application layer, since the pair is unordered and a
-- plain UNIQUE constraint can't express that without a canonical ordering).
CREATE INDEX friendships_requester_idx ON social.friendships (requester_id, status);
CREATE INDEX friendships_addressee_idx ON social.friendships (addressee_id, status);
