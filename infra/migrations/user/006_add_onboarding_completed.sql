-- +migrate Up
ALTER TABLE users.students
    ADD COLUMN IF NOT EXISTS onboarding_completed BOOLEAN NOT NULL DEFAULT false;

-- +migrate Down
ALTER TABLE users.students DROP COLUMN IF EXISTS onboarding_completed;
