-- Migration 008 Down: Remove Audit Performance Optimizations

-- Drop performance monitoring settings
ALTER SYSTEM RESET track_activity_query_size;
ALTER SYSTEM RESET log_min_duration_statement;
ALTER SYSTEM RESET log_statement;

-- Drop function-specific indexes
DROP INDEX IF EXISTS idx_audit_logs_failed_user_timestamp;
DROP INDEX IF EXISTS idx_audit_logs_dangerous_user_timestamp;

-- Drop functions
DROP FUNCTION IF EXISTS calculate_compliance_score(TEXT, TIMESTAMPTZ, TIMESTAMPTZ);
DROP FUNCTION IF EXISTS detect_suspicious_activity(INTERVAL);
DROP FUNCTION IF EXISTS archive_audit_logs(TIMESTAMPTZ, UUID);
DROP FUNCTION IF EXISTS refresh_audit_metrics();

-- Drop archive table
DROP TABLE IF EXISTS audit_logs_archive;

-- Drop materialized views
DROP MATERIALIZED VIEW IF EXISTS audit_metrics_daily;
DROP MATERIALIZED VIEW IF EXISTS audit_metrics_hourly;

-- Drop performance indexes
DROP INDEX IF EXISTS idx_audit_logs_ip_address;
DROP INDEX IF EXISTS idx_audit_logs_execution_result_gin;
DROP INDEX IF EXISTS idx_audit_logs_dangerous_ops;
DROP INDEX IF EXISTS idx_audit_logs_failed_operations;
DROP INDEX IF EXISTS idx_audit_logs_integrity;
DROP INDEX IF EXISTS idx_audit_logs_cluster_context;
DROP INDEX IF EXISTS idx_audit_logs_session_timestamp;
DROP INDEX IF EXISTS idx_audit_logs_execution_status;
DROP INDEX IF EXISTS idx_audit_logs_safety_level;
DROP INDEX IF EXISTS idx_audit_logs_user_timestamp;
DROP INDEX IF EXISTS idx_audit_logs_timestamp;