-- PostgreSQL initialization script for Corretor Redação
-- This script is run automatically when the postgres container starts

-- Create UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create ENUM types
CREATE TYPE user_tier AS ENUM ('free', 'premium');
CREATE TYPE correction_status AS ENUM ('pending', 'processing', 'completed', 'failed');
CREATE TYPE audit_event_type AS ENUM (
    'user_created',
    'user_verified',
    'user_deleted',
    'correction_submitted',
    'correction_started',
    'llm_started',
    'llm_completed',
    'schema_failed',
    'retry',
    'retry_exhausted',
    'correction_completed',
    'correction_failed',
    'reevaluation_submitted'
);

-- Set connection settings
ALTER DATABASE corretor_db SET max_connections = 200;
ALTER DATABASE corretor_db SET shared_buffers = '256MB';
ALTER DATABASE corretor_db SET effective_cache_size = '1GB';

COMMIT;
