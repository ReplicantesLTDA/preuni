-- Downgrade: 005_create_oauth_identities

DROP INDEX IF EXISTS auth.idx_oauth_identities_credential;
DROP INDEX IF EXISTS auth.idx_oauth_identities_provider_user;
DROP TABLE IF EXISTS auth.oauth_identities;
