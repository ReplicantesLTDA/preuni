-- Migration: 001_create_credentials
-- Schema: auth
-- Description: Core authentication credentials table

CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE auth.credentials (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(320) NOT NULL,
    password_hash   VARCHAR(256) NOT NULL,
    email_verified  BOOLEAN     NOT NULL DEFAULT false,
    email_verified_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Unique index on normalized lowercase email
CREATE UNIQUE INDEX idx_credentials_email ON auth.credentials (lower(email));

-- Trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION auth.set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_credentials_updated_at
    BEFORE UPDATE ON auth.credentials
    FOR EACH ROW EXECUTE FUNCTION auth.set_updated_at();
