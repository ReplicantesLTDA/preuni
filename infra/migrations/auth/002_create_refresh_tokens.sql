-- Migration: 002_create_refresh_tokens
-- Schema: auth
-- Description: Opaque refresh token store (stores SHA-256 hash, never plaintext)

CREATE TABLE auth.refresh_tokens (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    credential_id   UUID        NOT NULL REFERENCES auth.credentials(id) ON DELETE CASCADE,
    token_hash      VARCHAR(256) NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,             -- NULL = still valid
    issued_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_refresh_tokens_hash ON auth.refresh_tokens (token_hash);
CREATE INDEX idx_refresh_tokens_credential ON auth.refresh_tokens (credential_id);
