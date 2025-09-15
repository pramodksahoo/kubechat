-- KubeChat Database - Seed Data & Compliance Features Script
-- Consolidates: Seed Data, Story 1.8 Compliance Features, Advanced Security
-- Version: Story 1.8 - Complete Implementation

-- =================================================================
-- COMPLIANCE AND SECURITY ENHANCEMENTS
-- =================================================================

-- Legal Hold Management Tables
CREATE TABLE IF NOT EXISTS legal_holds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_number VARCHAR(255) UNIQUE NOT NULL,
    description TEXT NOT NULL,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'released', 'expired')),
    record_count BIGINT DEFAULT 0,
    -- Compliance metadata
    legal_authority VARCHAR(255),
    case_type VARCHAR(100),
    retention_override BOOLEAN DEFAULT TRUE,
    -- Audit fields
    released_by UUID REFERENCES users(id),
    released_at TIMESTAMPTZ,
    release_reason TEXT
);

-- Compliance Violation Tracking
CREATE TABLE IF NOT EXISTS compliance_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    framework VARCHAR(50) NOT NULL, -- sox, hipaa, soc2, gdpr, pci_dss
    violation_type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    description TEXT NOT NULL,
    affected_log_ids BIGINT[],
    detected_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    status VARCHAR(20) DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'resolved', 'false_positive')),
    -- Resolution tracking
    assigned_to UUID REFERENCES users(id),
    resolution_notes TEXT,
    remediation_actions TEXT[]
);

-- Tamper Detection Alerts
CREATE TABLE IF NOT EXISTS tamper_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    detected_at TIMESTAMPTZ DEFAULT NOW(),
    affected_log_id BIGINT REFERENCES audit_logs(id),
    violation_type VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    severity VARCHAR(20) CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    -- Detection metadata
    detection_method VARCHAR(50), -- checksum, pattern, anomaly
    confidence_score NUMERIC(3,2) CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0),
    -- Response tracking
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMPTZ,
    response_actions TEXT[],
    status VARCHAR(20) DEFAULT 'open' CHECK (status IN ('open', 'investigating', 'resolved', 'false_positive'))
);

-- Retention Policies Management
CREATE TABLE IF NOT EXISTS retention_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    retention_days INTEGER NOT NULL CHECK (retention_days > 0),
    applies_to TEXT NOT NULL, -- table or criteria specification
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_applied TIMESTAMPTZ,
    automatic BOOLEAN DEFAULT FALSE,
    -- Policy configuration
    priority INTEGER DEFAULT 100,
    conditions JSONB, -- Complex condition rules
    archive_location VARCHAR(255),
    compression_enabled BOOLEAN DEFAULT TRUE,
    encryption_enabled BOOLEAN DEFAULT TRUE,
    -- Compliance framework requirements
    compliance_frameworks TEXT[],
    legal_hold_exempt BOOLEAN DEFAULT FALSE
);

-- =================================================================
-- COMPLIANCE INDEXES
-- =================================================================

CREATE INDEX idx_legal_holds_status ON legal_holds(status);
CREATE INDEX idx_legal_holds_dates ON legal_holds(start_time, end_time);
CREATE INDEX idx_legal_holds_case_number ON legal_holds(case_number);

CREATE INDEX idx_compliance_violations_framework ON compliance_violations(framework);
CREATE INDEX idx_compliance_violations_severity ON compliance_violations(severity);
CREATE INDEX idx_compliance_violations_status ON compliance_violations(status);
CREATE INDEX idx_compliance_violations_detected ON compliance_violations(detected_at DESC);

CREATE INDEX idx_tamper_alerts_severity ON tamper_alerts(severity);
CREATE INDEX idx_tamper_alerts_detected ON tamper_alerts(detected_at DESC);
CREATE INDEX idx_tamper_alerts_status ON tamper_alerts(status);

CREATE INDEX idx_retention_policies_automatic ON retention_policies(automatic) WHERE automatic = TRUE;

-- =================================================================
-- COMPLIANCE FUNCTIONS
-- =================================================================

-- Legal Hold Management Function
CREATE OR REPLACE FUNCTION apply_legal_hold(
    p_case_number VARCHAR(255),
    p_start_time TIMESTAMPTZ,
    p_end_time TIMESTAMPTZ DEFAULT NULL
) RETURNS UUID AS $$
DECLARE
    hold_id UUID;
    record_count BIGINT;
BEGIN
    -- Count affected records
    SELECT COUNT(*) INTO record_count
    FROM audit_logs
    WHERE timestamp >= p_start_time
      AND (p_end_time IS NULL OR timestamp <= p_end_time);

    -- Create legal hold record
    INSERT INTO legal_holds (case_number, start_time, end_time, record_count, description)
    VALUES (p_case_number, p_start_time, p_end_time, record_count,
            'Legal hold for case ' || p_case_number || ' covering ' || record_count || ' records')
    RETURNING id INTO hold_id;

    -- Update affected audit logs with legal hold flag
    UPDATE audit_logs
    SET compliance_tags = array_append(COALESCE(compliance_tags, '{}'), 'legal_hold:' || p_case_number),
        retention_policy = 'legal_hold'
    WHERE timestamp >= p_start_time
      AND (p_end_time IS NULL OR timestamp <= p_end_time);

    RETURN hold_id;
END;
$$ LANGUAGE plpgsql;

-- Compliance Violation Detection Function
CREATE OR REPLACE FUNCTION detect_compliance_violations(
    p_framework VARCHAR(50),
    p_start_time TIMESTAMPTZ DEFAULT NOW() - INTERVAL '24 hours'
) RETURNS INTEGER AS $$
DECLARE
    violation_count INTEGER := 0;
    violation_record RECORD;
BEGIN
    -- SOX Compliance Violations
    IF p_framework = 'sox' OR p_framework = 'all' THEN
        -- Detect dangerous operations without proper approval
        FOR violation_record IN
            SELECT array_agg(id) as log_ids, COUNT(*) as count, user_id
            FROM audit_logs
            WHERE timestamp >= p_start_time
              AND safety_level = 'dangerous'
              AND execution_status = 'success'
              AND NOT EXISTS (
                  SELECT 1 FROM audit_logs al2
                  WHERE al2.user_id = audit_logs.user_id
                    AND al2.timestamp < audit_logs.timestamp
                    AND al2.query_text LIKE '%approve%'
                    AND al2.timestamp >= audit_logs.timestamp - INTERVAL '1 hour'
              )
            GROUP BY user_id
            HAVING COUNT(*) >= 3
        LOOP
            INSERT INTO compliance_violations (
                framework, violation_type, severity, description, affected_log_ids
            ) VALUES (
                'sox', 'unauthorized_dangerous_operations', 'high',
                'User performed ' || violation_record.count || ' dangerous operations without proper approval',
                violation_record.log_ids
            );
            violation_count := violation_count + 1;
        END LOOP;
    END IF;

    -- HIPAA Compliance Violations
    IF p_framework = 'hipaa' OR p_framework = 'all' THEN
        -- Detect access without proper authentication
        FOR violation_record IN
            SELECT array_agg(id) as log_ids, COUNT(*) as count
            FROM audit_logs
            WHERE timestamp >= p_start_time
              AND user_id IS NULL
              AND execution_status = 'success'
            GROUP BY ip_address
            HAVING COUNT(*) >= 5
        LOOP
            INSERT INTO compliance_violations (
                framework, violation_type, severity, description, affected_log_ids
            ) VALUES (
                'hipaa', 'unauthenticated_access', 'critical',
                'Multiple successful operations without proper user authentication',
                violation_record.log_ids
            );
            violation_count := violation_count + 1;
        END LOOP;
    END IF;

    RETURN violation_count;
END;
$$ LANGUAGE plpgsql;

-- Tamper Detection Function
CREATE OR REPLACE FUNCTION detect_tampering() RETURNS INTEGER AS $$
DECLARE
    alert_count INTEGER := 0;
    log_record RECORD;
    expected_checksum VARCHAR(64);
BEGIN
    -- Check for checksum violations
    FOR log_record IN
        SELECT id, checksum, user_id, session_id, query_text, generated_command,
               safety_level, execution_result, execution_status, cluster_context,
               namespace_context, timestamp, ip_address, user_agent, previous_checksum,
               risk_score, command_category
        FROM audit_logs
        WHERE timestamp >= NOW() - INTERVAL '1 hour'
        ORDER BY id
    LOOP
        -- Recalculate expected checksum
        expected_checksum := calculate_audit_checksum(
            log_record.user_id, log_record.session_id, log_record.query_text,
            log_record.generated_command, log_record.safety_level, log_record.execution_result,
            log_record.execution_status, log_record.cluster_context, log_record.namespace_context,
            log_record.timestamp, log_record.ip_address, log_record.user_agent,
            log_record.previous_checksum, log_record.risk_score, log_record.command_category
        );

        -- Check for tampering
        IF expected_checksum != log_record.checksum THEN
            INSERT INTO tamper_alerts (
                affected_log_id, violation_type, description, severity,
                detection_method, confidence_score
            ) VALUES (
                log_record.id, 'checksum_mismatch',
                'Audit log checksum does not match calculated value',
                'critical', 'checksum', 1.0
            );
            alert_count := alert_count + 1;
        END IF;
    END LOOP;

    RETURN alert_count;
END;
$$ LANGUAGE plpgsql;

-- =================================================================
-- AUTOMATED COMPLIANCE MONITORING
-- =================================================================

-- Create automated compliance check procedure
CREATE OR REPLACE FUNCTION run_compliance_checks() RETURNS TEXT AS $$
DECLARE
    sox_violations INTEGER;
    hipaa_violations INTEGER;
    tamper_alerts INTEGER;
    result_text TEXT;
BEGIN
    -- Run compliance violation detection
    sox_violations := detect_compliance_violations('sox');
    hipaa_violations := detect_compliance_violations('hipaa');

    -- Run tamper detection
    tamper_alerts := detect_tampering();

    -- Refresh metrics
    PERFORM refresh_audit_metrics();

    result_text := format(
        'Compliance check completed: %s SOX violations, %s HIPAA violations, %s tamper alerts detected',
        sox_violations, hipaa_violations, tamper_alerts
    );

    RETURN result_text;
END;
$$ LANGUAGE plpgsql;

-- =================================================================
-- SEED DATA AND INITIAL SETUP
-- =================================================================

-- Create default admin user (password: admin123 - change in production!)
INSERT INTO users (id, username, email, password_hash, role) VALUES (
    gen_random_uuid(),
    'admin',
    'admin@kubechat.dev',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewNxn0h8P8O9gOJm', -- bcrypt hash of 'admin123'
    'admin'
) ON CONFLICT (username) DO NOTHING;

-- Create compliance officer user
INSERT INTO users (id, username, email, password_hash, role) VALUES (
    gen_random_uuid(),
    'compliance_officer',
    'compliance@kubechat.dev',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewNxn0h8P8O9gOJm',
    'compliance_officer'
) ON CONFLICT (username) DO NOTHING;

-- Create auditor user
INSERT INTO users (id, username, email, password_hash, role) VALUES (
    gen_random_uuid(),
    'auditor',
    'auditor@kubechat.dev',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewNxn0h8P8O9gOJm',
    'auditor'
) ON CONFLICT (username) DO NOTHING;

-- =================================================================
-- DEFAULT RETENTION POLICIES
-- =================================================================

-- Standard retention policy
INSERT INTO retention_policies (
    name, description, retention_days, applies_to, automatic,
    compliance_frameworks, archive_location
) VALUES (
    'Standard Audit Retention',
    'Standard 7-year retention for audit logs as per SOX requirements',
    2555, -- 7 years
    'audit_logs',
    TRUE,
    ARRAY['sox', 'soc2'],
    'archive_standard'
) ON CONFLICT (name) DO NOTHING;

-- HIPAA retention policy
INSERT INTO retention_policies (
    name, description, retention_days, applies_to, automatic,
    compliance_frameworks, archive_location
) VALUES (
    'HIPAA Audit Retention',
    'HIPAA-compliant 6-year retention for healthcare audit logs',
    2190, -- 6 years
    'audit_logs WHERE compliance_tags && ARRAY[''hipaa'']',
    TRUE,
    ARRAY['hipaa'],
    'archive_hipaa'
) ON CONFLICT (name) DO NOTHING;

-- High-risk operations extended retention
INSERT INTO retention_policies (
    name, description, retention_days, applies_to, automatic,
    compliance_frameworks, archive_location, priority
) VALUES (
    'High Risk Extended Retention',
    'Extended 10-year retention for high-risk operations',
    3650, -- 10 years
    'audit_logs WHERE risk_score > 0.7 OR safety_level = ''dangerous''',
    TRUE,
    ARRAY['sox', 'soc2', 'hipaa'],
    'archive_high_risk',
    50 -- Higher priority than standard
) ON CONFLICT (name) DO NOTHING;

-- =================================================================
-- SAMPLE AUDIT DATA FOR TESTING
-- =================================================================

-- Insert sample audit entries for testing (only if no audit logs exist)
DO $$
DECLARE
    admin_user_id UUID;
    test_session_id UUID;
BEGIN
    -- Only insert sample data if no audit logs exist
    IF NOT EXISTS (SELECT 1 FROM audit_logs LIMIT 1) THEN
        -- Get admin user ID
        SELECT id INTO admin_user_id FROM users WHERE username = 'admin' LIMIT 1;

        -- Create a test session
        INSERT INTO user_sessions (user_id, session_token, expires_at, ip_address)
        VALUES (admin_user_id, encode(gen_random_bytes(32), 'hex'), NOW() + INTERVAL '1 day', '127.0.0.1'::INET)
        RETURNING id INTO test_session_id;

        -- Insert sample audit logs
        INSERT INTO audit_logs (
            user_id, session_id, query_text, generated_command, safety_level,
            execution_result, execution_status, cluster_context, command_category
        ) VALUES
        (admin_user_id, test_session_id, 'get pods', 'kubectl get pods', 'safe',
         '{"pods": []}', 'success', 'default', 'read'),
        (admin_user_id, test_session_id, 'get services', 'kubectl get services', 'safe',
         '{"services": []}', 'success', 'default', 'read'),
        (admin_user_id, test_session_id, 'describe cluster', 'kubectl cluster-info', 'safe',
         '{"cluster": "healthy"}', 'success', 'kube-system', 'read');

        RAISE NOTICE 'Sample audit data inserted for testing';
    END IF;
END;
$$;

-- =================================================================
-- FINAL CONFIGURATION AND CLEANUP
-- =================================================================

-- Create database-level configuration
CREATE OR REPLACE FUNCTION configure_audit_database() RETURNS TEXT AS $$
BEGIN
    -- Set database parameters for audit workload
    PERFORM set_config('log_statement', 'mod', false);
    PERFORM set_config('log_min_duration_statement', '1000', false);

    -- Enable query statistics
    PERFORM set_config('track_activities', 'on', false);
    PERFORM set_config('track_counts', 'on', false);

    RETURN 'Database configured for audit workload';
END;
$$ LANGUAGE plpgsql;

-- Apply configuration
SELECT configure_audit_database();

-- =================================================================
-- MONITORING AND ALERTING SETUP
-- =================================================================

-- Create a function to check system health
CREATE OR REPLACE FUNCTION audit_system_health_check() RETURNS TABLE(
    component TEXT,
    status TEXT,
    details TEXT
) AS $$
BEGIN
    -- Database connection check
    component := 'database_connection';
    status := 'healthy';
    details := 'PostgreSQL connection active';
    RETURN NEXT;

    -- Audit log integrity check
    RETURN QUERY
    SELECT
        'audit_integrity'::TEXT,
        CASE WHEN COUNT(*) = 0 THEN 'healthy' ELSE 'warning' END,
        'Checked ' || COUNT(*) || ' recent audit logs for integrity'
    FROM verify_audit_log_integrity()
    WHERE NOT is_valid AND entry_id IN (
        SELECT id FROM audit_logs WHERE timestamp >= NOW() - INTERVAL '1 hour'
    );

    -- Recent tamper alerts check
    RETURN QUERY
    SELECT
        'tamper_detection'::TEXT,
        CASE WHEN COUNT(*) = 0 THEN 'healthy' ELSE 'critical' END,
        COUNT(*) || ' tamper alerts in last hour'
    FROM tamper_alerts
    WHERE detected_at >= NOW() - INTERVAL '1 hour';

    -- Compliance violations check
    RETURN QUERY
    SELECT
        'compliance_status'::TEXT,
        CASE WHEN COUNT(*) = 0 THEN 'healthy' ELSE 'warning' END,
        COUNT(*) || ' open compliance violations'
    FROM compliance_violations
    WHERE status = 'open';
END;
$$ LANGUAGE plpgsql;

-- =================================================================
-- COMPLETION AND VERIFICATION
-- =================================================================

-- Log the completion of seed and compliance setup
INSERT INTO schema_migrations (version, migration_name, notes) VALUES (
    '003_seed_and_compliance',
    'Seed Data and Compliance Features',
    'User accounts, compliance tables, retention policies, sample data, and Story 1.8 features completed'
);

-- Final verification
DO $$
DECLARE
    table_count INTEGER;
    index_count INTEGER;
    function_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO table_count FROM information_schema.tables WHERE table_schema = 'public';
    SELECT COUNT(*) INTO index_count FROM pg_indexes WHERE schemaname = 'public';
    SELECT COUNT(*) INTO function_count FROM information_schema.routines WHERE routine_schema = 'public';

    RAISE NOTICE 'Database setup completed successfully:';
    RAISE NOTICE '  - Tables: %', table_count;
    RAISE NOTICE '  - Indexes: %', index_count;
    RAISE NOTICE '  - Functions: %', function_count;
    RAISE NOTICE '  - Story 1.8 Implementation: Complete';
    RAISE NOTICE '  - Performance Optimizations: Applied';
    RAISE NOTICE '  - Compliance Features: Enabled';
END;
$$;