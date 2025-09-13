-- Down Migration: 003_rollback_tables.down.sql
-- Description: Drop rollback functionality tables
-- Story: 1.6 - Kubernetes Cluster Integration and Safe Command Execution

-- Drop security policies first
DROP POLICY IF EXISTS rollback_executions_user_access ON rollback_executions;
DROP POLICY IF EXISTS rollback_plans_user_access ON rollback_plans;

-- Drop indexes
DROP INDEX IF EXISTS idx_rollback_executions_completed_at;
DROP INDEX IF EXISTS idx_rollback_executions_started_at;
DROP INDEX IF EXISTS idx_rollback_executions_status;
DROP INDEX IF EXISTS idx_rollback_executions_user;
DROP INDEX IF EXISTS idx_rollback_executions_plan;

DROP INDEX IF EXISTS idx_rollback_plans_created_at;
DROP INDEX IF EXISTS idx_rollback_plans_expires_at;
DROP INDEX IF EXISTS idx_rollback_plans_status;
DROP INDEX IF EXISTS idx_rollback_plans_session;
DROP INDEX IF EXISTS idx_rollback_plans_user;
DROP INDEX IF EXISTS idx_rollback_plans_command_execution;

-- Drop tables (child table first due to foreign key constraints)
DROP TABLE IF EXISTS rollback_executions;
DROP TABLE IF EXISTS rollback_plans;