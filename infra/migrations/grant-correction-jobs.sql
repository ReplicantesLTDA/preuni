-- Grants the Go monolith's DB role narrow, enqueue/poll-only access to the
-- correction service's outbox/bridge table. Run after
-- corretor-redacao's 002_correction_jobs_schema migration has created the
-- `correction` schema and `correction.correction_jobs` table.
--
-- See specs/014-constitution-alignment-refactor/contracts/internal-bridge.md
-- and research.md #1 for why this is a shared-Postgres, schema-scoped grant
-- rather than a second database instance.
--
-- Usage: psql "$DATABASE_URL" -f infra/migrations/grant-correction-jobs.sql
--        -v monolith_role=preuni_monolith

\set monolith_role 'preuni_monolith'

GRANT USAGE ON SCHEMA correction TO :monolith_role;
GRANT INSERT, SELECT ON correction.correction_jobs TO :monolith_role;

-- Explicitly deny write access to everything else in the schema — the
-- monolith must never touch `corrections`, `grader_passes`, or
-- `correction_audit_logs` directly (contracts/internal-bridge.md).
REVOKE ALL ON ALL TABLES IN SCHEMA correction FROM :monolith_role;
GRANT INSERT, SELECT ON correction.correction_jobs TO :monolith_role;
