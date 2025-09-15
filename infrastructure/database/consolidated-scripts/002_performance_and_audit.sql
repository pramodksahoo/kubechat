-- KubeChat Database - Performance & Audit Enhancement Script
-- Consolidates: Performance Optimizations, Audit Functions, Triggers, Advanced Indexes
-- Version: Story 1.8 - Task 6 Performance Optimizations Included

-- =================================================================
-- PERFORMANCE OPTIMIZATION INDEXES
-- =================================================================

-- Composite indexes for audit queries (from Task 6)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_user_timestamp
ON audit_logs (user_id, timestamp DESC) WHERE user_id IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_safety_timestamp
ON audit_logs (safety_level, timestamp DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_execution_timestamp
ON audit_logs (execution_status, timestamp DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_session_timestamp
ON audit_logs (session_id, timestamp DESC) WHERE session_id IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_cluster_timestamp
ON audit_logs (cluster_context, timestamp DESC) WHERE cluster_context IS NOT NULL;

-- Partial indexes for specific query patterns
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_failed_operations
ON audit_logs (user_id, timestamp DESC, execution_status)
WHERE execution_status = 'failed' AND user_id IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_dangerous_ops
ON audit_logs (user_id, timestamp DESC, safety_level)
WHERE safety_level = 'dangerous' AND user_id IS NOT NULL;

-- GIN index for JSON execution_result queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_execution_result_gin
ON audit_logs USING GIN (execution_result);

-- IP address analysis index
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_ip_timestamp
ON audit_logs (ip_address, timestamp DESC) WHERE ip_address IS NOT NULL;

-- Compliance and risk analysis indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_risk_score
ON audit_logs (risk_score DESC, timestamp DESC) WHERE risk_score > 0.5;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_compliance_tags
ON audit_logs USING GIN (compliance_tags) WHERE compliance_tags IS NOT NULL;

-- Session security indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_sessions_security_level
ON user_sessions (security_level, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_sessions_device
ON user_sessions (device_fingerprint, user_id) WHERE device_fingerprint IS NOT NULL;

-- =================================================================
-- AUDIT INTEGRITY FUNCTIONS
-- =================================================================

-- Enhanced audit checksum calculation function
CREATE OR REPLACE FUNCTION calculate_audit_checksum(
    p_user_id UUID,
    p_session_id UUID,
    p_query_text TEXT,
    p_generated_command TEXT,
    p_safety_level VARCHAR(20),
    p_execution_result JSONB,
    p_execution_status VARCHAR(20),
    p_cluster_context VARCHAR(255),
    p_namespace_context VARCHAR(255),
    p_timestamp TIMESTAMPTZ,
    p_ip_address INET,
    p_user_agent TEXT,
    p_previous_checksum VARCHAR(64),
    p_risk_score NUMERIC DEFAULT 0.0,
    p_command_category VARCHAR(50) DEFAULT NULL
) RETURNS VARCHAR(64) AS $$
DECLARE
    checksum_input TEXT;
BEGIN
    checksum_input := CONCAT(
        COALESCE(p_user_id::text, ''),
        '|',
        COALESCE(p_session_id::text, ''),
        '|',
        COALESCE(p_query_text, ''),
        '|',
        COALESCE(p_generated_command, ''),
        '|',
        COALESCE(p_safety_level, ''),
        '|',
        COALESCE(p_execution_result::text, ''),
        '|',
        COALESCE(p_execution_status, ''),
        '|',
        COALESCE(p_cluster_context, ''),
        '|',
        COALESCE(p_namespace_context, ''),
        '|',
        COALESCE(p_timestamp::text, ''),
        '|',
        COALESCE(p_ip_address::text, ''),
        '|',
        COALESCE(p_user_agent, ''),
        '|',
        COALESCE(p_previous_checksum, ''),
        '|',
        COALESCE(p_risk_score::text, '0'),
        '|',
        COALESCE(p_command_category, '')
    );

    RETURN encode(digest(checksum_input, 'sha256'), 'hex');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Chain integrity verification function
CREATE OR REPLACE FUNCTION verify_audit_log_integrity(p_entry_id BIGINT DEFAULT NULL)
RETURNS TABLE(
    entry_id BIGINT,
    is_valid BOOLEAN,
    calculated_checksum VARCHAR(64),
    stored_checksum VARCHAR(64),
    error_message TEXT
) AS $$
DECLARE
    rec RECORD;
    calc_checksum VARCHAR(64);
BEGIN
    FOR rec IN
        SELECT * FROM audit_logs
        WHERE (p_entry_id IS NULL OR id = p_entry_id)
        ORDER BY id
    LOOP
        BEGIN
            calc_checksum := calculate_audit_checksum(
                rec.user_id,
                rec.session_id,
                rec.query_text,
                rec.generated_command,
                rec.safety_level,
                rec.execution_result,
                rec.execution_status,
                rec.cluster_context,
                rec.namespace_context,
                rec.timestamp,
                rec.ip_address,
                rec.user_agent,
                rec.previous_checksum,
                rec.risk_score,
                rec.command_category
            );

            entry_id := rec.id;
            calculated_checksum := calc_checksum;
            stored_checksum := rec.checksum;
            is_valid := (calc_checksum = rec.checksum);
            error_message := CASE
                WHEN NOT is_valid THEN 'Checksum mismatch detected'
                ELSE NULL
            END;

            RETURN NEXT;
        EXCEPTION WHEN OTHERS THEN
            entry_id := rec.id;
            calculated_checksum := NULL;
            stored_checksum := rec.checksum;
            is_valid := FALSE;
            error_message := SQLERRM;
            RETURN NEXT;
        END;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- =================================================================
-- PERFORMANCE MONITORING FUNCTIONS (Task 6)
-- =================================================================

-- Function for suspicious activity detection
CREATE OR REPLACE FUNCTION detect_suspicious_activity(time_window INTERVAL DEFAULT INTERVAL '24 hours')
RETURNS TABLE(
    user_id UUID,
    activity_type TEXT,
    risk_score NUMERIC,
    event_count BIGINT,
    latest_event TIMESTAMPTZ,
    description TEXT
) AS $$
BEGIN
    -- Multiple failed operations by same user
    RETURN QUERY
    SELECT
        al.user_id,
        'multiple_failures'::TEXT as activity_type,
        COUNT(*)::NUMERIC * 10 as risk_score,
        COUNT(*) as event_count,
        MAX(al.timestamp) as latest_event,
        'User has ' || COUNT(*) || ' failed operations in ' || time_window as description
    FROM audit_logs al
    WHERE al.timestamp >= NOW() - time_window
        AND al.execution_status = 'failed'
        AND al.user_id IS NOT NULL
    GROUP BY al.user_id
    HAVING COUNT(*) >= 5;

    -- High rate of dangerous operations
    RETURN QUERY
    SELECT
        al.user_id,
        'high_risk_operations'::TEXT as activity_type,
        COUNT(*)::NUMERIC * 15 as risk_score,
        COUNT(*) as event_count,
        MAX(al.timestamp) as latest_event,
        'User has ' || COUNT(*) || ' dangerous operations in ' || time_window as description
    FROM audit_logs al
    WHERE al.timestamp >= NOW() - time_window
        AND al.safety_level = 'dangerous'
        AND al.user_id IS NOT NULL
    GROUP BY al.user_id
    HAVING COUNT(*) >= 10;

    -- Unusual access patterns (multiple IPs)
    RETURN QUERY
    SELECT
        al.user_id,
        'multiple_ip_access'::TEXT as activity_type,
        COUNT(DISTINCT al.ip_address)::NUMERIC * 20 as risk_score,
        COUNT(*) as event_count,
        MAX(al.timestamp) as latest_event,
        'User accessed from ' || COUNT(DISTINCT al.ip_address) || ' different IP addresses' as description
    FROM audit_logs al
    WHERE al.timestamp >= NOW() - time_window
        AND al.user_id IS NOT NULL
        AND al.ip_address IS NOT NULL
    GROUP BY al.user_id
    HAVING COUNT(DISTINCT al.ip_address) >= 3;
END;
$$ LANGUAGE plpgsql;

-- Compliance score calculation function
CREATE OR REPLACE FUNCTION calculate_compliance_score(
    framework TEXT,
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ
) RETURNS NUMERIC AS $$
DECLARE
    total_events BIGINT;
    violations BIGINT;
    score NUMERIC;
BEGIN
    SELECT COUNT(*) INTO total_events
    FROM audit_logs
    WHERE timestamp BETWEEN start_date AND end_date;

    IF total_events = 0 THEN
        RETURN 100.0;
    END IF;

    -- Framework-specific violation counting
    IF framework = 'sox' THEN
        -- SOX compliance: dangerous operations without proper controls
        SELECT COUNT(*) INTO violations
        FROM audit_logs
        WHERE timestamp BETWEEN start_date AND end_date
            AND safety_level = 'dangerous'
            AND execution_status = 'failed';

    ELSIF framework = 'hipaa' THEN
        -- HIPAA compliance: unauthorized access attempts
        SELECT COUNT(*) INTO violations
        FROM audit_logs
        WHERE timestamp BETWEEN start_date AND end_date
            AND execution_status = 'failed'
            AND user_id IS NULL;

    ELSIF framework = 'soc2' THEN
        -- SOC2 compliance: security control failures
        SELECT COUNT(*) INTO violations
        FROM audit_logs
        WHERE timestamp BETWEEN start_date AND end_date
            AND (safety_level = 'dangerous' OR execution_status = 'failed');
    ELSE
        violations := 0;
    END IF;

    -- Calculate score (100% - violation percentage)
    score := 100.0 - (violations::NUMERIC / total_events::NUMERIC * 100.0);

    RETURN GREATEST(0.0, score);
END;
$$ LANGUAGE plpgsql;

-- =================================================================
-- MATERIALIZED VIEWS FOR PERFORMANCE (Task 6)
-- =================================================================

-- Hourly audit metrics view
CREATE MATERIALIZED VIEW IF NOT EXISTS audit_metrics_hourly AS
SELECT
    date_trunc('hour', timestamp) as hour,
    COUNT(*) as total_logs,
    COUNT(CASE WHEN safety_level = 'dangerous' THEN 1 END) as dangerous_ops,
    COUNT(CASE WHEN execution_status = 'failed' THEN 1 END) as failed_ops,
    COUNT(CASE WHEN execution_status = 'success' THEN 1 END) as successful_ops,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(DISTINCT session_id) as unique_sessions,
    AVG(execution_duration_ms) as avg_duration_ms
FROM audit_logs
WHERE timestamp >= NOW() - INTERVAL '30 days'
GROUP BY date_trunc('hour', timestamp)
ORDER BY hour DESC;

-- Daily audit metrics view
CREATE MATERIALIZED VIEW IF NOT EXISTS audit_metrics_daily AS
SELECT
    date_trunc('day', timestamp) as day,
    COUNT(*) as total_logs,
    COUNT(CASE WHEN safety_level = 'dangerous' THEN 1 END) as dangerous_ops,
    COUNT(CASE WHEN execution_status = 'failed' THEN 1 END) as failed_ops,
    COUNT(CASE WHEN execution_status = 'success' THEN 1 END) as successful_ops,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(DISTINCT session_id) as unique_sessions,
    COUNT(DISTINCT cluster_context) as unique_clusters,
    -- Compliance metrics
    COUNT(CASE WHEN safety_level = 'dangerous' AND execution_status = 'success' THEN 1 END) as successful_dangerous_ops,
    COUNT(CASE WHEN safety_level = 'dangerous' AND execution_status = 'failed' THEN 1 END) as failed_dangerous_ops,
    -- Integrity metrics
    COUNT(CASE WHEN checksum IS NOT NULL THEN 1 END) as checksummed_logs,
    COUNT(CASE WHEN previous_checksum IS NOT NULL THEN 1 END) as chained_logs
FROM audit_logs
WHERE timestamp >= NOW() - INTERVAL '2 years'
GROUP BY date_trunc('day', timestamp)
ORDER BY day DESC;

-- Indexes on materialized views
CREATE INDEX IF NOT EXISTS idx_audit_metrics_hourly_hour
ON audit_metrics_hourly (hour DESC);

CREATE INDEX IF NOT EXISTS idx_audit_metrics_daily_day
ON audit_metrics_daily (day DESC);

-- Function to refresh materialized views
CREATE OR REPLACE FUNCTION refresh_audit_metrics()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY audit_metrics_hourly;
    REFRESH MATERIALIZED VIEW CONCURRENTLY audit_metrics_daily;
END;
$$ LANGUAGE plpgsql;

-- =================================================================
-- AUDIT INTEGRITY TRIGGERS
-- =================================================================

-- Enhanced trigger function for audit log integrity with risk scoring
CREATE OR REPLACE FUNCTION audit_log_integrity_trigger() RETURNS TRIGGER AS $$
DECLARE
    prev_checksum VARCHAR(64);
    calculated_risk NUMERIC;
BEGIN
    -- Get the previous checksum from the last audit log entry
    SELECT checksum INTO prev_checksum
    FROM audit_logs
    ORDER BY id DESC
    LIMIT 1;

    -- Calculate risk score based on operation characteristics
    calculated_risk := 0.0;

    -- Increase risk for dangerous operations
    IF NEW.safety_level = 'dangerous' THEN
        calculated_risk := calculated_risk + 0.5;
    ELSIF NEW.safety_level = 'warning' THEN
        calculated_risk := calculated_risk + 0.2;
    END IF;

    -- Increase risk for failed operations
    IF NEW.execution_status = 'failed' THEN
        calculated_risk := calculated_risk + 0.3;
    END IF;

    -- Increase risk for admin/write operations
    IF NEW.command_category IN ('admin', 'delete', 'write') THEN
        calculated_risk := calculated_risk + 0.2;
    END IF;

    -- Set calculated values
    NEW.previous_checksum := prev_checksum;
    NEW.risk_score := LEAST(calculated_risk, 1.0);

    -- Generate correlation IDs if not provided
    IF NEW.request_id IS NULL THEN
        NEW.request_id := gen_random_uuid();
    END IF;

    IF NEW.correlation_id IS NULL THEN
        NEW.correlation_id := gen_random_uuid();
    END IF;

    -- Calculate checksum for the new record
    NEW.checksum := calculate_audit_checksum(
        NEW.user_id,
        NEW.session_id,
        NEW.query_text,
        NEW.generated_command,
        NEW.safety_level,
        NEW.execution_result,
        NEW.execution_status,
        NEW.cluster_context,
        NEW.namespace_context,
        NEW.timestamp,
        NEW.ip_address,
        NEW.user_agent,
        prev_checksum,
        NEW.risk_score,
        NEW.command_category
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create/Replace the trigger
DROP TRIGGER IF EXISTS audit_log_checksum_trigger ON audit_logs;
CREATE TRIGGER audit_log_checksum_trigger
    BEFORE INSERT ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION audit_log_integrity_trigger();

-- =================================================================
-- ARCHIVAL AND RETENTION FUNCTIONS (Task 6)
-- =================================================================

-- Function for efficient archival
CREATE OR REPLACE FUNCTION archive_audit_logs(cutoff_date TIMESTAMPTZ, batch_id UUID)
RETURNS INTEGER AS $$
DECLARE
    archived_count INTEGER;
BEGIN
    -- Insert into archive table
    INSERT INTO audit_logs_archive
    SELECT *, NOW(), batch_id, 'audit_logs', 'retention_policy'
    FROM audit_logs
    WHERE timestamp < cutoff_date;

    GET DIAGNOSTICS archived_count = ROW_COUNT;

    -- Delete from main table
    DELETE FROM audit_logs
    WHERE timestamp < cutoff_date;

    RETURN archived_count;
END;
$$ LANGUAGE plpgsql;

-- =================================================================
-- PERFORMANCE CONFIGURATION
-- =================================================================

-- Optimize PostgreSQL settings for audit workload
DO $$
BEGIN
    -- Only execute if user has sufficient privileges
    IF current_setting('is_superuser')::boolean THEN
        -- Query performance optimizations
        PERFORM set_config('work_mem', '256MB', false);
        PERFORM set_config('maintenance_work_mem', '512MB', false);
        PERFORM set_config('max_parallel_workers_per_gather', '4', false);
        PERFORM set_config('random_page_cost', '1.1', false);

        -- Audit-specific optimizations
        PERFORM set_config('autovacuum_analyze_scale_factor', '0.05', false);
        PERFORM set_config('autovacuum_vacuum_scale_factor', '0.1', false);

        -- Logging optimizations
        PERFORM set_config('track_activities', 'on', false);
        PERFORM set_config('track_counts', 'on', false);
        PERFORM set_config('track_io_timing', 'on', false);
    END IF;
EXCEPTION WHEN insufficient_privilege THEN
    -- Log the warning but continue
    RAISE NOTICE 'Skipping PostgreSQL configuration - insufficient privileges';
END;
$$;

-- =================================================================
-- PERFORMANCE SCRIPT COMPLETION
-- =================================================================

-- Log the completion of performance and audit setup
INSERT INTO schema_migrations (version, migration_name, notes) VALUES (
    '002_performance_and_audit',
    'Performance Optimizations and Audit Enhancements',
    'Advanced indexes, audit functions, triggers, materialized views, and Task 6 optimizations applied'
);