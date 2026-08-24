-- Grants the Go monolith's DB role narrow access to the correction
-- service's outbox/bridge table and read-only access to its graded
-- results. Run after ai-corrector's 002_correction_jobs_schema
-- migration has created the `correction` schema and its tables.
--
-- See specs/014-constitution-alignment-refactor/contracts/internal-bridge.md
-- and research.md #1 for why this is a shared-Postgres, schema-scoped grant
-- rather than a second database instance.
--
-- Usage: psql "$DATABASE_URL" -f infra/migrations/grant-correction-jobs.sql
--        -v monolith_role=preuni_monolith
--
-- OPTIONAL, currently unexecuted: this project's dev/CI setup does not
-- provision a separate `preuni_monolith` role — both the monolith and the
-- correction service connect as the same shared Postgres user today
-- (constitution's Security & Secrets principle explicitly rejects
-- RBAC-style access control for this project's current scale). This
-- script is least-privilege hardening for a deployment that has actually
-- created that role; run it by hand only in that case. See
-- specs/032-corrector-service-integration/research.md #R3.

\set monolith_role 'preuni_monolith'

GRANT USAGE ON SCHEMA correction TO :monolith_role;

-- Explicitly deny write access to everything else in the schema first —
-- the monolith enqueues jobs and reads results, it never writes grading
-- output (contracts/internal-bridge.md).
REVOKE ALL ON ALL TABLES IN SCHEMA correction FROM :monolith_role;

-- correction_jobs: the monolith enqueues (INSERT) and polls status (SELECT).
GRANT INSERT, SELECT ON correction.correction_jobs TO :monolith_role;

-- corrections: read-only, so the monolith's reconciler can pull the graded
-- payload (score + competencies) once a job completes.
GRANT SELECT ON correction.corrections TO :monolith_role;
