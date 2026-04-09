-- +migrate Up
ALTER TABLE users.students
    ADD COLUMN IF NOT EXISTS email VARCHAR(254);

-- +migrate Down
ALTER TABLE users.students DROP COLUMN IF EXISTS email;
