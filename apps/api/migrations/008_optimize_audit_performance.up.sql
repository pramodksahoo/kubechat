-- Migration 008: Optimize Audit Performance
-- Task 6: Database Performance Optimization for Story 1.8

-- Create optimized indexes for audit queries
-- Index for timestamp-based queries (most common filter)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_timestamp
ON audit_logs (timestamp DESC);

-- Composite index for user-based queries with timestamp
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_user_timestamp
ON audit_logs (user_id, timestamp DESC)
WHERE user_id IS NOT NULL;

-- Index for safety level filtering (dangerous operations)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_safety_level
ON audit_logs (safety_level, timestamp DESC);

-- Index for execution status filtering (failed operations)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_execution_status
ON audit_logs (execution_status, timestamp DESC);

-- Composite index for session-based queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_session_timestamp
ON audit_logs (session_id, timestamp DESC)
WHERE session_id IS NOT NULL;

-- Index for cluster context filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_cluster_context
ON audit_logs (cluster_context, timestamp DESC)
WHERE cluster_context IS NOT NULL;

-- Index for integrity verification queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_integrity
ON audit_logs (id ASC, checksum, previous_checksum);

-- Partial index for suspicious activity detection (failed operations)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_failed_operations
ON audit_logs (user_id, timestamp DESC, execution_status)
WHERE execution_status = 'failed' AND user_id IS NOT NULL;

-- Partial index for dangerous operations monitoring
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_dangerous_ops
ON audit_logs (user_id, timestamp DESC, safety_level)
WHERE safety_level = 'dangerous' AND user_id IS NOT NULL;

-- GIN index for JSON execution_result queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_execution_result_gin
ON audit_logs USING GIN (execution_result);

-- Index for IP address based analysis
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_ip_address
ON audit_logs (ip_address, timestamp DESC)
WHERE ip_address IS NOT NULL;

-- Create partitioning preparation
-- Add partition key constraint to prepare for future partitioning by timestamp
-- This will help with large-scale deployments

-- Create audit metrics materialized view for faster dashboard queries
CREATE MATERIALIZED VIEW IF NOT EXISTS audit_metrics_hourly AS
SELECT
    date_trunc('hour', timestamp) as hour,
    COUNT(*) as total_logs,
    COUNT(CASE WHEN safety_level = 'dangerous' THEN 1 END) as dangerous_ops,
    COUNT(CASE WHEN execution_status = 'failed' THEN 1 END) as failed_ops,
    COUNT(CASE WHEN execution_status = 'success' THEN 1 END) as successful_ops,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(DISTINCT session_id) as unique_sessions,
    AVG(CASE WHEN execution_result->>'duration_ms' IS NOT NULL
        THEN (execution_result->>'duration_ms')::numeric END) as avg_duration_ms
FROM audit_logs
WHERE timestamp >= NOW() - INTERVAL '30 days'
GROUP BY date_trunc('hour', timestamp)
ORDER BY hour DESC;

-- Index on materialized view
CREATE INDEX IF NOT EXISTS idx_audit_metrics_hourly_hour
ON audit_metrics_hourly (hour DESC);

-- Create daily metrics view for compliance reporting
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

-- Index on daily metrics view
CREATE INDEX IF NOT EXISTS idx_audit_metrics_daily_day
ON audit_metrics_daily (day DESC);

-- Create function to refresh materialized views
CREATE OR REPLACE FUNCTION refresh_audit_metrics()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY audit_metrics_hourly;
    REFRESH MATERIALIZED VIEW CONCURRENTLY audit_metrics_daily;
END;
$$ LANGUAGE plpgsql;

-- Create audit log archival table for retention policy
CREATE TABLE IF NOT EXISTS audit_logs_archive (
    LIKE audit_logs INCLUDING ALL
);

-- Add archival metadata
ALTER TABLE audit_logs_archive
ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS archive_batch_id UUID;

-- Index for archive table
CREATE INDEX IF NOT EXISTS idx_audit_logs_archive_archived_at
ON audit_logs_archive (archived_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_archive_batch_id
ON audit_logs_archive (archive_batch_id);

-- Create function for efficient archival
CREATE OR REPLACE FUNCTION archive_audit_logs(cutoff_date TIMESTAMPTZ, batch_id UUID)
RETURNS INTEGER AS $$
DECLARE
    archived_count INTEGER;
BEGIN
    -- Insert into archive table
    INSERT INTO audit_logs_archive
    SELECT *, NOW(), batch_id
    FROM audit_logs
    WHERE timestamp < cutoff_date;

    GET DIAGNOSTICS archived_count = ROW_COUNT;

    -- Delete from main table
    DELETE FROM audit_logs
    WHERE timestamp < cutoff_date;

    RETURN archived_count;
END;
$$ LANGUAGE plpgsql;

-- Create function for suspicious activity detection
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

-- Create function for compliance scoring
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

-- Create indexes for functions
CREATE INDEX IF NOT EXISTS idx_audit_logs_failed_user_timestamp
ON audit_logs (user_id, timestamp DESC)
WHERE execution_status = 'failed' AND user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_audit_logs_dangerous_user_timestamp
ON audit_logs (user_id, timestamp DESC)
WHERE safety_level = 'dangerous' AND user_id IS NOT NULL;

-- Enable query statistics collection
ALTER SYSTEM SET track_activity_query_size = 2048;
ALTER SYSTEM SET log_min_duration_statement = 1000; -- Log slow queries (1 second)
ALTER SYSTEM SET log_statement = 'mod'; -- Log modifications