-- Migration 001 DOWN: Rollback Initial Database Schema
-- Rollback script for the initial database schema
-- Date: 2025-01-11
-- Author: James (Full Stack Developer Agent)

-- Drop triggers first
DROP TRIGGER IF EXISTS set_audit_log_checksum_trigger ON audit_logs;
DROP TRIGGER IF EXISTS prevent_audit_log_updates ON audit_logs;
DROP TRIGGER IF EXISTS prevent_audit_log_deletes ON audit_logs;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_cluster_configs_updated_at ON cluster_configs;

-- Drop functions
DROP FUNCTION IF EXISTS calculate_audit_checksum(UUID, UUID, TEXT, TEXT, VARCHAR(20), JSONB, VARCHAR(20), VARCHAR(255), VARCHAR(255), TIMESTAMP, INET, TEXT, VARCHAR(64));
DROP FUNCTION IF EXISTS set_audit_log_checksum();
DROP FUNCTION IF EXISTS prevent_audit_log_changes();
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS verify_audit_log_integrity(BIGINT);

-- Drop indexes
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_created_at;

DROP INDEX IF EXISTS idx_user_sessions_user_id;
DROP INDEX IF EXISTS idx_user_sessions_token;
DROP INDEX IF EXISTS idx_user_sessions_expires_at;
DROP INDEX IF EXISTS idx_user_sessions_created_at;

DROP INDEX IF EXISTS idx_audit_logs_user_id;
DROP INDEX IF EXISTS idx_audit_logs_session_id;
DROP INDEX IF EXISTS idx_audit_logs_timestamp;
DROP INDEX IF EXISTS idx_audit_logs_safety_level;
DROP INDEX IF EXISTS idx_audit_logs_execution_status;
DROP INDEX IF EXISTS idx_audit_logs_cluster_context;
DROP INDEX IF EXISTS idx_audit_logs_checksum;

DROP INDEX IF EXISTS idx_cluster_configs_user_id;
DROP INDEX IF EXISTS idx_cluster_configs_name;
DROP INDEX IF EXISTS idx_cluster_configs_active;
DROP INDEX IF EXISTS idx_cluster_configs_created_at;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS cluster_configs CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop extensions (only if no other objects depend on them)
-- Note: These are commonly used extensions, so we don't drop them
-- DROP EXTENSION IF EXISTS "pgcrypto";
-- DROP EXTENSION IF EXISTS "uuid-ossp";