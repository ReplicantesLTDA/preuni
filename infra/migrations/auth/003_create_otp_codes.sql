-- Migration: 003_create_otp_codes
-- Schema: auth
-- Description: One-time passwords for email verification, password reset, OTP login, email change

CREATE TYPE auth.otp_purpose AS ENUM (
    'EMAIL_VERIFY',
    'PASSWORD_RESET',
    'LOGIN_OTP',
    'EMAIL_CHANGE'
);

CREATE TABLE auth.otp_codes (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    credential_id   UUID            NOT NULL REFERENCES auth.credentials(id) ON DELETE CASCADE,
    purpose         auth.otp_purpose NOT NULL,
    code_hash       VARCHAR(256)    NOT NULL,   -- SHA-256 of the 6-digit code; never store plaintext
    expires_at      TIMESTAMPTZ     NOT NULL,
    used_at         TIMESTAMPTZ,               -- NULL = not yet consumed
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now()
);

-- Allows fast lookup of an active OTP for a given credential + purpose
CREATE INDEX idx_otp_codes_credential_purpose
    ON auth.otp_codes (credential_id, purpose)
    WHERE used_at IS NULL;
