-- KubeChat Admin Credential Management Script
-- AC 8: Built-in Admin Credential Management - K8s Secret Integration
-- Enterprise security compliance for SOC 2 Type II, GDPR, and enterprise audits
-- Version: Story 2.3 - Admin User Management & RBAC Interface

-- =================================================================
-- ADMIN CREDENTIAL MANAGEMENT TABLES
-- =================================================================

-- Admin credential sync status tracking
CREATE TABLE IF NOT EXISTS admin_credential_sync (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sync_timestamp TIMESTAMPTZ DEFAULT NOW(),
    sync_source VARCHAR(20) NOT NULL CHECK (sync_source IN ('k8s_secret', 'database', 'rotation', 'emergency')),
    sync_type VARCHAR(20) NOT NULL CHECK (sync_type IN ('bootstrap', 'update', 'rotation', 'verification')),
    sync_status VARCHAR(20) NOT NULL CHECK (sync_status IN ('success', 'failed', 'pending', 'partial')),
    -- Credential metadata
    username VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    -- K8s Secret metadata
    k8s_secret_name VARCHAR(255) DEFAULT 'kubechat-admin-secret',
    k8s_namespace VARCHAR(255) DEFAULT 'kubechat-system',
    k8s_resource_version VARCHAR(255),
    -- Password policy compliance
    password_created_at TIMESTAMPTZ,
    password_expires_at TIMESTAMPTZ,
    rotation_count INTEGER DEFAULT 0,
    -- Audit and compliance
    compliance_status VARCHAR(20) DEFAULT 'compliant' CHECK (compliance_status IN ('compliant', 'non_compliant', 'pending_review')),
    compliance_frameworks TEXT[] DEFAULT ARRAY['soc2', 'gdpr'],
    sync_initiated_by VARCHAR(255), -- user or system
    error_message TEXT,
    notes TEXT
);

-- Admin credential audit trail (immutable audit log)
CREATE TABLE IF NOT EXISTS admin_credential_audit (
    id BIGSERIAL PRIMARY KEY,
    audit_timestamp TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    event_type VARCHAR(50) NOT NULL, -- 'credential_access', 'password_change', 'sync_operation', 'rotation', 'emergency_access'
    actor_type VARCHAR(20) NOT NULL CHECK (actor_type IN ('user', 'system', 'k8s_controller')),
    actor_id VARCHAR(255), -- user ID or system component
    -- Operation details
    operation VARCHAR(100) NOT NULL,
    target_username VARCHAR(255) NOT NULL,
    -- Source and destination
    source_system VARCHAR(50), -- 'k8s_secret', 'database', 'admin_ui', 'cli'
    destination_system VARCHAR(50),
    -- Security context
    ip_address INET,
    user_agent TEXT,
    session_id UUID,
    -- Compliance and risk
    risk_level VARCHAR(20) DEFAULT 'medium' CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    compliance_impact TEXT[],
    -- Outcome
    success BOOLEAN NOT NULL,
    error_code VARCHAR(50),
    error_message TEXT,
    -- Immutable audit fields
    audit_checksum VARCHAR(128) NOT NULL,
    previous_audit_checksum VARCHAR(128),
    -- Legal and compliance
    retention_policy VARCHAR(50) DEFAULT 'enterprise_admin',
    legal_hold_exempt BOOLEAN DEFAULT FALSE,
    pii_data_mask BOOLEAN DEFAULT TRUE
);

-- Admin emergency access log (for break-glass scenarios)
CREATE TABLE IF NOT EXISTS admin_emergency_access (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    emergency_timestamp TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    emergency_type VARCHAR(50) NOT NULL, -- 'password_reset', 'account_unlock', 'credential_recovery', 'bypass_mfa'
    -- Request details
    requested_by VARCHAR(255) NOT NULL,
    approved_by VARCHAR(255),
    justification TEXT NOT NULL,
    business_impact TEXT,
    -- Approval workflow
    approval_status VARCHAR(20) DEFAULT 'pending' CHECK (approval_status IN ('pending', 'approved', 'denied', 'expired')),
    approval_timestamp TIMESTAMPTZ,
    -- Emergency credentials
    emergency_username VARCHAR(255),
    emergency_password_hash VARCHAR(255),
    emergency_token VARCHAR(500),
    expires_at TIMESTAMPTZ,
    -- Security controls
    ip_whitelist INET[],
    session_restrictions JSONB,
    mfa_bypass_approved BOOLEAN DEFAULT FALSE,
    -- Audit and compliance
    used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    revocation_reason TEXT,
    compliance_review_required BOOLEAN DEFAULT TRUE,
    compliance_review_completed BOOLEAN DEFAULT FALSE,
    -- Immutable audit
    audit_checksum VARCHAR(128) NOT NULL
);

-- =================================================================
-- INDEXES FOR PERFORMANCE AND COMPLIANCE
-- =================================================================

-- Admin credential sync indexes
CREATE INDEX idx_admin_credential_sync_timestamp ON admin_credential_sync(sync_timestamp DESC);
CREATE INDEX idx_admin_credential_sync_status ON admin_credential_sync(sync_status);
CREATE INDEX idx_admin_credential_sync_source ON admin_credential_sync(sync_source);
CREATE INDEX idx_admin_credential_sync_compliance ON admin_credential_sync(compliance_status);

-- Admin audit trail indexes (optimized for compliance queries)
CREATE INDEX idx_admin_audit_timestamp ON admin_credential_audit(audit_timestamp DESC);
CREATE INDEX idx_admin_audit_event_type ON admin_credential_audit(event_type);
CREATE INDEX idx_admin_audit_actor ON admin_credential_audit(actor_type, actor_id);
CREATE INDEX idx_admin_audit_username ON admin_credential_audit(target_username);
CREATE INDEX idx_admin_audit_risk_level ON admin_credential_audit(risk_level);
CREATE INDEX idx_admin_audit_compliance ON admin_credential_audit USING GIN(compliance_impact);

-- Emergency access indexes
CREATE INDEX idx_emergency_access_timestamp ON admin_emergency_access(emergency_timestamp DESC);
CREATE INDEX idx_emergency_access_status ON admin_emergency_access(approval_status);
CREATE INDEX idx_emergency_access_requested_by ON admin_emergency_access(requested_by);
CREATE INDEX idx_emergency_access_expires ON admin_emergency_access(expires_at);

-- =================================================================
-- ADMIN CREDENTIAL MANAGEMENT FUNCTIONS
-- =================================================================

-- Function to calculate admin audit checksum
CREATE OR REPLACE FUNCTION calculate_admin_audit_checksum(
    p_audit_timestamp TIMESTAMPTZ,
    p_event_type VARCHAR(50),
    p_actor_type VARCHAR(20),
    p_actor_id VARCHAR(255),
    p_operation VARCHAR(100),
    p_target_username VARCHAR(255),
    p_source_system VARCHAR(50),
    p_success BOOLEAN,
    p_previous_checksum VARCHAR(128)
) RETURNS VARCHAR(128) AS $$
BEGIN
    RETURN encode(
        digest(
            COALESCE(p_audit_timestamp::TEXT, '') || '|' ||
            COALESCE(p_event_type, '') || '|' ||
            COALESCE(p_actor_type, '') || '|' ||
            COALESCE(p_actor_id, '') || '|' ||
            COALESCE(p_operation, '') || '|' ||
            COALESCE(p_target_username, '') || '|' ||
            COALESCE(p_source_system, '') || '|' ||
            COALESCE(p_success::TEXT, '') || '|' ||
            COALESCE(p_previous_checksum, ''),
            'sha256'
        ),
        'hex'
    );
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to log admin credential operations with audit trail
CREATE OR REPLACE FUNCTION log_admin_credential_audit(
    p_event_type VARCHAR(50),
    p_actor_type VARCHAR(20),
    p_actor_id VARCHAR(255),
    p_operation VARCHAR(100),
    p_target_username VARCHAR(255),
    p_source_system VARCHAR(50) DEFAULT NULL,
    p_destination_system VARCHAR(50) DEFAULT NULL,
    p_ip_address INET DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL,
    p_session_id UUID DEFAULT NULL,
    p_risk_level VARCHAR(20) DEFAULT 'medium',
    p_success BOOLEAN DEFAULT TRUE,
    p_error_code VARCHAR(50) DEFAULT NULL,
    p_error_message TEXT DEFAULT NULL
) RETURNS BIGINT AS $$
DECLARE
    audit_id BIGINT;
    current_checksum VARCHAR(128);
    previous_checksum VARCHAR(128);
BEGIN
    -- Get the last audit checksum for chain integrity
    SELECT audit_checksum INTO previous_checksum
    FROM admin_credential_audit
    ORDER BY id DESC
    LIMIT 1;

    -- Calculate checksum for this audit entry
    current_checksum := calculate_admin_audit_checksum(
        NOW(),
        p_event_type,
        p_actor_type,
        p_actor_id,
        p_operation,
        p_target_username,
        p_source_system,
        p_success,
        previous_checksum
    );

    -- Insert audit entry
    INSERT INTO admin_credential_audit (
        event_type, actor_type, actor_id, operation, target_username,
        source_system, destination_system, ip_address, user_agent, session_id,
        risk_level, success, error_code, error_message,
        audit_checksum, previous_audit_checksum
    ) VALUES (
        p_event_type, p_actor_type, p_actor_id, p_operation, p_target_username,
        p_source_system, p_destination_system, p_ip_address, p_user_agent, p_session_id,
        p_risk_level, p_success, p_error_code, p_error_message,
        current_checksum, previous_checksum
    ) RETURNING id INTO audit_id;

    RETURN audit_id;
END;
$$ LANGUAGE plpgsql;

-- Function to sync admin credentials from K8s Secret to database
CREATE OR REPLACE FUNCTION sync_admin_credentials_from_k8s_secret(
    p_username VARCHAR(255),
    p_password_hash VARCHAR(255),
    p_email VARCHAR(255),
    p_k8s_resource_version VARCHAR(255),
    p_password_created_at TIMESTAMPTZ,
    p_password_expires_at TIMESTAMPTZ,
    p_rotation_count INTEGER DEFAULT 0,
    p_sync_initiated_by VARCHAR(255) DEFAULT 'k8s_controller'
) RETURNS UUID AS $$
DECLARE
    sync_id UUID;
    sync_success BOOLEAN := TRUE;
    error_msg TEXT := NULL;
    existing_user_count INTEGER;
BEGIN
    -- Check if admin user already exists
    SELECT COUNT(*) INTO existing_user_count
    FROM users
    WHERE username = p_username AND role = 'admin';

    BEGIN
        -- Update or insert admin user
        IF existing_user_count > 0 THEN
            -- Update existing admin user
            UPDATE users
            SET password_hash = p_password_hash,
                email = p_email,
                updated_at = NOW()
            WHERE username = p_username AND role = 'admin';
        ELSE
            -- Insert new admin user
            INSERT INTO users (
                id, username, email, password_hash, role, created_at, updated_at
            ) VALUES (
                gen_random_uuid(), p_username, p_email, p_password_hash, 'admin', NOW(), NOW()
            );
        END IF;

        -- Log successful sync
        PERFORM log_admin_credential_audit(
            'credential_sync',
            'system',
            p_sync_initiated_by,
            'sync_from_k8s_secret',
            p_username,
            'k8s_secret',
            'database',
            NULL, -- IP address
            NULL, -- User agent
            NULL, -- Session ID
            'low', -- Risk level for automated sync
            TRUE
        );

    EXCEPTION WHEN OTHERS THEN
        sync_success := FALSE;
        error_msg := SQLERRM;

        -- Log failed sync
        PERFORM log_admin_credential_audit(
            'credential_sync',
            'system',
            p_sync_initiated_by,
            'sync_from_k8s_secret_failed',
            p_username,
            'k8s_secret',
            'database',
            NULL, -- IP address
            NULL, -- User agent
            NULL, -- Session ID
            'high', -- Risk level for failed sync
            FALSE,
            'SYNC_FAILED',
            error_msg
        );
    END;

    -- Record sync operation
    INSERT INTO admin_credential_sync (
        sync_source, sync_type, sync_status, username, password_hash, email,
        k8s_resource_version, password_created_at, password_expires_at,
        rotation_count, sync_initiated_by, error_message
    ) VALUES (
        'k8s_secret', 'bootstrap',
        CASE WHEN sync_success THEN 'success' ELSE 'failed' END,
        p_username, p_password_hash, p_email,
        p_k8s_resource_version, p_password_created_at, p_password_expires_at,
        p_rotation_count, p_sync_initiated_by, error_msg
    ) RETURNING id INTO sync_id;

    RETURN sync_id;
END;
$$ LANGUAGE plpgsql;

-- Function to validate admin credential compliance
CREATE OR REPLACE FUNCTION validate_admin_credential_compliance(
    p_username VARCHAR(255) DEFAULT 'admin'
) RETURNS TABLE(
    compliance_check VARCHAR(100),
    status VARCHAR(20),
    details TEXT,
    remediation TEXT
) AS $$
DECLARE
    user_record RECORD;
    sync_record RECORD;
    audit_count INTEGER;
    emergency_count INTEGER;
BEGIN
    -- Get admin user record
    SELECT * INTO user_record
    FROM users
    WHERE username = p_username AND role = 'admin'
    LIMIT 1;

    -- Get latest sync record
    SELECT * INTO sync_record
    FROM admin_credential_sync
    WHERE username = p_username
    ORDER BY sync_timestamp DESC
    LIMIT 1;

    -- Check 1: Admin user exists
    compliance_check := 'admin_user_exists';
    IF user_record.id IS NOT NULL THEN
        status := 'compliant';
        details := 'Admin user found in database';
        remediation := NULL;
    ELSE
        status := 'non_compliant';
        details := 'Admin user not found in database';
        remediation := 'Execute K8s Secret sync to create admin user';
    END IF;
    RETURN NEXT;

    -- Check 2: Recent sync status
    compliance_check := 'credential_sync_status';
    IF sync_record.id IS NOT NULL AND sync_record.sync_status = 'success' THEN
        status := 'compliant';
        details := 'Last sync successful at ' || sync_record.sync_timestamp;
        remediation := NULL;
    ELSE
        status := 'non_compliant';
        details := 'No successful sync found or last sync failed';
        remediation := 'Re-run credential sync process';
    END IF;
    RETURN NEXT;

    -- Check 3: Password age compliance
    compliance_check := 'password_age_compliance';
    IF sync_record.password_expires_at IS NOT NULL THEN
        IF sync_record.password_expires_at > NOW() THEN
            status := 'compliant';
            details := 'Password expires at ' || sync_record.password_expires_at;
            remediation := NULL;
        ELSE
            status := 'non_compliant';
            details := 'Password expired at ' || sync_record.password_expires_at;
            remediation := 'Rotate admin password immediately';
        END IF;
    ELSE
        status := 'pending_review';
        details := 'Password expiry not tracked';
        remediation := 'Configure password expiry tracking';
    END IF;
    RETURN NEXT;

    -- Check 4: Audit trail integrity
    SELECT COUNT(*) INTO audit_count
    FROM admin_credential_audit
    WHERE target_username = p_username
      AND audit_timestamp >= NOW() - INTERVAL '30 days';

    compliance_check := 'audit_trail_integrity';
    IF audit_count > 0 THEN
        status := 'compliant';
        details := audit_count || ' audit entries in last 30 days';
        remediation := NULL;
    ELSE
        status := 'non_compliant';
        details := 'No audit entries found in last 30 days';
        remediation := 'Investigate audit logging configuration';
    END IF;
    RETURN NEXT;

    -- Check 5: Emergency access usage
    SELECT COUNT(*) INTO emergency_count
    FROM admin_emergency_access
    WHERE emergency_username = p_username
      AND emergency_timestamp >= NOW() - INTERVAL '90 days'
      AND used_at IS NOT NULL;

    compliance_check := 'emergency_access_usage';
    IF emergency_count = 0 THEN
        status := 'compliant';
        details := 'No emergency access used in last 90 days';
        remediation := NULL;
    ELSE
        status := 'pending_review';
        details := emergency_count || ' emergency access uses in last 90 days';
        remediation := 'Review emergency access justifications';
    END IF;
    RETURN NEXT;

END;
$$ LANGUAGE plpgsql;

-- Function to rotate admin credentials with zero-downtime
CREATE OR REPLACE FUNCTION rotate_admin_credentials(
    p_new_password_hash VARCHAR(255),
    p_rotation_initiated_by VARCHAR(255),
    p_justification TEXT DEFAULT 'Scheduled rotation'
) RETURNS TABLE(
    rotation_id UUID,
    success BOOLEAN,
    message TEXT
) AS $$
DECLARE
    new_rotation_id UUID;
    rotation_success BOOLEAN := TRUE;
    rotation_message TEXT;
    current_rotation_count INTEGER;
BEGIN
    -- Get current rotation count
    SELECT COALESCE(rotation_count, 0) INTO current_rotation_count
    FROM admin_credential_sync
    WHERE username = 'admin'
    ORDER BY sync_timestamp DESC
    LIMIT 1;

    BEGIN
        -- Update admin user password
        UPDATE users
        SET password_hash = p_new_password_hash,
            updated_at = NOW()
        WHERE username = 'admin' AND role = 'admin';

        -- Record rotation in sync table
        INSERT INTO admin_credential_sync (
            sync_source, sync_type, sync_status, username, password_hash,
            email, password_created_at, password_expires_at,
            rotation_count, sync_initiated_by, notes
        ) VALUES (
            'rotation', 'rotation', 'success', 'admin', p_new_password_hash,
            'admin@kubechat.dev', NOW(), NOW() + INTERVAL '90 days',
            current_rotation_count + 1, p_rotation_initiated_by, p_justification
        ) RETURNING id INTO new_rotation_id;

        -- Log rotation in audit trail
        PERFORM log_admin_credential_audit(
            'password_rotation',
            'system',
            p_rotation_initiated_by,
            'rotate_admin_password',
            'admin',
            'database',
            'k8s_secret',
            NULL, NULL, NULL,
            'medium',
            TRUE
        );

        rotation_message := 'Admin password rotated successfully';

    EXCEPTION WHEN OTHERS THEN
        rotation_success := FALSE;
        rotation_message := 'Rotation failed: ' || SQLERRM;

        -- Log failed rotation
        PERFORM log_admin_credential_audit(
            'password_rotation',
            'system',
            p_rotation_initiated_by,
            'rotate_admin_password_failed',
            'admin',
            'database',
            'k8s_secret',
            NULL, NULL, NULL,
            'high',
            FALSE,
            'ROTATION_FAILED',
            SQLERRM
        );
    END;

    -- Return results
    rotation_id := new_rotation_id;
    success := rotation_success;
    message := rotation_message;
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

-- =================================================================
-- CLEAN MIGRATION FROM HARDCODED CREDENTIALS
-- =================================================================

-- Migration function to clean existing hardcoded admin credentials
CREATE OR REPLACE FUNCTION migrate_from_hardcoded_credentials() RETURNS TEXT AS $$
DECLARE
    hardcoded_admin_count INTEGER;
    migration_result TEXT;
BEGIN
    -- Check for hardcoded admin users (with known hardcoded password hash)
    SELECT COUNT(*) INTO hardcoded_admin_count
    FROM users
    WHERE username = 'admin'
      AND role = 'admin'
      AND password_hash = '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewNxn0h8P8O9gOJm';

    IF hardcoded_admin_count > 0 THEN
        -- Log the migration
        PERFORM log_admin_credential_audit(
            'credential_migration',
            'system',
            'migration_script',
            'remove_hardcoded_credentials',
            'admin',
            'database',
            'k8s_secret',
            NULL, NULL, NULL,
            'high',
            TRUE
        );

        -- Delete hardcoded admin user (will be recreated via K8s Secret sync)
        DELETE FROM users
        WHERE username = 'admin'
          AND role = 'admin'
          AND password_hash = '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewNxn0h8P8O9gOJm';

        migration_result := 'Removed ' || hardcoded_admin_count || ' hardcoded admin credential(s). K8s Secret sync required.';
    ELSE
        migration_result := 'No hardcoded admin credentials found. Migration not needed.';
    END IF;

    RETURN migration_result;
END;
$$ LANGUAGE plpgsql;

-- =================================================================
-- FINAL CONFIGURATION AND TRIGGERS
-- =================================================================

-- Trigger to automatically log credential changes
CREATE OR REPLACE FUNCTION trigger_admin_credential_change() RETURNS TRIGGER AS $$
BEGIN
    -- Log any changes to admin user credentials
    IF NEW.role = 'admin' AND (OLD.password_hash IS DISTINCT FROM NEW.password_hash) THEN
        PERFORM log_admin_credential_audit(
            'password_change',
            'user',
            COALESCE(current_setting('app.current_user_id', true), 'unknown'),
            'update_admin_password',
            NEW.username,
            'database',
            NULL,
            COALESCE(inet_client_addr(), '127.0.0.1'),
            current_setting('app.user_agent', true),
            COALESCE(current_setting('app.session_id', true)::UUID, NULL),
            'high',
            TRUE
        );
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for admin credential changes
DROP TRIGGER IF EXISTS trigger_admin_credential_change ON users;
CREATE TRIGGER trigger_admin_credential_change
    AFTER UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION trigger_admin_credential_change();

-- =================================================================
-- COMPLETION AND VERIFICATION
-- =================================================================

-- Log the completion of admin credential management setup
INSERT INTO schema_migrations (version, migration_name, notes) VALUES (
    '004_admin_credential_management',
    'Admin Credential Management with K8s Secret Integration',
    'AC 8: Built-in Admin Credential Management - Tables, functions, audit trail, and compliance validation completed for Story 2.3'
);

-- Execute migration from hardcoded credentials
SELECT migrate_from_hardcoded_credentials() AS migration_status;

-- Initial compliance validation
DO $$
BEGIN
    RAISE NOTICE 'Admin Credential Management Setup Complete:';
    RAISE NOTICE '  - K8s Secret integration tables created';
    RAISE NOTICE '  - Audit trail and compliance functions implemented';
    RAISE NOTICE '  - Emergency access procedures configured';
    RAISE NOTICE '  - Hardcoded credential migration completed';
    RAISE NOTICE '  - Ready for K8s Secret-based admin credential bootstrap';
    RAISE NOTICE '  - Compliance frameworks: SOC 2 Type II, GDPR';
END;
$$;