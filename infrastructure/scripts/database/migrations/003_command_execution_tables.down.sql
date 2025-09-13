-- Migration 003 Rollback: Drop Command Execution Tables
-- Story 1.6: Kubernetes Cluster Integration and Safe Command Execution

-- Drop tables in reverse order (to handle foreign key constraints)
DROP TABLE IF EXISTS command_rollback_info;
DROP TABLE IF EXISTS command_approvals;
DROP TABLE IF EXISTS command_executions;