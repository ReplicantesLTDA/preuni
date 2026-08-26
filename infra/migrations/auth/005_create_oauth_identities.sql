-- Migration: 005_create_oauth_identities
-- Schema: auth
-- Description: Links a credential to an external OAuth provider identity
-- (e.g. Google). A credential row still exists for OAuth-only signups --
-- it carries an unusable random password_hash, never presented to the user.

CREATE TABLE auth.oauth_identities (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    credential_id    UUID        NOT NULL REFERENCES auth.credentials(id) ON DELETE CASCADE,
    provider         VARCHAR(32) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_oauth_identities_provider_user
    ON auth.oauth_identities (provider, provider_user_id);

CREATE INDEX idx_oauth_identities_credential
    ON auth.oauth_identities (credential_id);
