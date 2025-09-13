-- Migration 001: Initial Database Schema
-- Creates the foundational database schema for KubeChat with enhanced security
-- Date: 2025-01-11
-- Author: James (Full Stack Developer Agent)

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users and Authentication Table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Constraints for data integrity
    CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,4}$'),
    CONSTRAINT valid_username CHECK (LENGTH(username) >= 3 AND LENGTH(username) <= 255),
    CONSTRAINT valid_role CHECK (role IN ('admin', 'user', 'viewer'))
);

-- User Sessions Table with Security Tracking
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    
    -- Security constraints
    CONSTRAINT valid_session_token CHECK (LENGTH(session_token) >= 32),
    CONSTRAINT valid_expiry CHECK (expires_at > created_at)
);

-- Audit Trail Table (IMMUTABLE with Cryptographic Integrity)
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    session_id UUID REFERENCES user_sessions(id),
    query_text TEXT NOT NULL,
    generated_command TEXT NOT NULL,
    safety_level VARCHAR(20) NOT NULL CHECK (safety_level IN ('safe', 'warning', 'dangerous')),
    execution_result JSONB,
    execution_status VARCHAR(20) NOT NULL CHECK (execution_status IN ('success', 'failed', 'cancelled')),
    cluster_context VARCHAR(255),
    namespace_context VARCHAR(255),
    timestamp TIMESTAMP DEFAULT NOW() NOT NULL,
    ip_address INET,
    user_agent TEXT,
    
    -- Immutability protection with cryptographic integrity
    checksum VARCHAR(64) NOT NULL, -- SHA-256 of record
    previous_checksum VARCHAR(64) -- Chain to previous record for tamper detection
);

-- Kubernetes Cluster Configurations Table
CREATE TABLE cluster_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    cluster_name VARCHAR(255) NOT NULL,
    cluster_config JSONB NOT NULL, -- Encrypted kubeconfig data
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Business constraints
    CONSTRAINT valid_cluster_name CHECK (LENGTH(cluster_name) >= 1 AND LENGTH(cluster_name) <= 255)
);

-- Performance Indexes
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_created_at ON users(created_at);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_created_at ON user_sessions(created_at);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_session_id ON audit_logs(session_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_logs_safety_level ON audit_logs(safety_level);
CREATE INDEX idx_audit_logs_execution_status ON audit_logs(execution_status);
CREATE INDEX idx_audit_logs_cluster_context ON audit_logs(cluster_context);
CREATE INDEX idx_audit_logs_checksum ON audit_logs(checksum);

CREATE INDEX idx_cluster_configs_user_id ON cluster_configs(user_id);
CREATE INDEX idx_cluster_configs_name ON cluster_configs(cluster_name);
CREATE INDEX idx_cluster_configs_active ON cluster_configs(is_active);
CREATE INDEX idx_cluster_configs_created_at ON cluster_configs(created_at);

-- Audit Trail Functions and Triggers

-- Function to calculate SHA-256 checksum for audit log integrity
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
    p_timestamp TIMESTAMP,
    p_ip_address INET,
    p_user_agent TEXT,
    p_previous_checksum VARCHAR(64)
)
RETURNS VARCHAR(64) AS $$
DECLARE
    checksum_input TEXT;
BEGIN
    -- Create deterministic string for SHA-256 hashing
    checksum_input := COALESCE(p_user_id::TEXT, '') || '|' ||
                      COALESCE(p_session_id::TEXT, '') || '|' ||
                      COALESCE(p_query_text, '') || '|' ||
                      COALESCE(p_generated_command, '') || '|' ||
                      COALESCE(p_safety_level, '') || '|' ||
                      COALESCE(p_execution_result::TEXT, '') || '|' ||
                      COALESCE(p_execution_status, '') || '|' ||
                      COALESCE(p_cluster_context, '') || '|' ||
                      COALESCE(p_namespace_context, '') || '|' ||
                      COALESCE(p_timestamp::TEXT, '') || '|' ||
                      COALESCE(p_ip_address::TEXT, '') || '|' ||
                      COALESCE(p_user_agent, '') || '|' ||
                      COALESCE(p_previous_checksum, '');
    
    RETURN encode(digest(checksum_input, 'sha256'), 'hex');
END;
$$ language 'plpgsql';

-- Trigger function to automatically calculate checksums and implement checksum chaining
CREATE OR REPLACE FUNCTION set_audit_log_checksum()
RETURNS TRIGGER AS $$
DECLARE
    prev_checksum VARCHAR(64);
BEGIN
    -- Get the previous checksum (last entry's checksum for chain validation)
    SELECT checksum INTO prev_checksum
    FROM audit_logs
    ORDER BY id DESC
    LIMIT 1;
    
    -- Calculate checksum for new entry
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
        prev_checksum
    );
    
    NEW.previous_checksum := prev_checksum;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to enforce immutability of audit logs (prevent updates/deletes)
CREATE OR REPLACE FUNCTION prevent_audit_log_changes()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Audit log entries are immutable and cannot be modified or deleted for compliance';
END;
$$ language 'plpgsql';

-- Function for updated_at timestamp management
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers
CREATE TRIGGER set_audit_log_checksum_trigger BEFORE INSERT ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION set_audit_log_checksum();

CREATE TRIGGER prevent_audit_log_updates BEFORE UPDATE ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_log_changes();

CREATE TRIGGER prevent_audit_log_deletes BEFORE DELETE ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_log_changes();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cluster_configs_updated_at BEFORE UPDATE ON cluster_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to verify audit log integrity chain
CREATE OR REPLACE FUNCTION verify_audit_log_integrity(log_id BIGINT DEFAULT NULL)
RETURNS TABLE(log_id BIGINT, is_valid BOOLEAN, error_message TEXT) AS $$
DECLARE
    audit_record RECORD;
    calculated_checksum VARCHAR(64);
    prev_checksum VARCHAR(64);
BEGIN
    -- If log_id is specified, check only that record, otherwise check all
    FOR audit_record IN 
        SELECT * FROM audit_logs 
        WHERE (verify_audit_log_integrity.log_id IS NULL OR audit_logs.id = verify_audit_log_integrity.log_id)
        ORDER BY audit_logs.id
    LOOP
        -- Get previous checksum
        SELECT audit_logs.checksum INTO prev_checksum
        FROM audit_logs
        WHERE audit_logs.id < audit_record.id
        ORDER BY audit_logs.id DESC
        LIMIT 1;
        
        -- Calculate expected checksum
        calculated_checksum := calculate_audit_checksum(
            audit_record.user_id,
            audit_record.session_id,
            audit_record.query_text,
            audit_record.generated_command,
            audit_record.safety_level,
            audit_record.execution_result,
            audit_record.execution_status,
            audit_record.cluster_context,
            audit_record.namespace_context,
            audit_record.timestamp,
            audit_record.ip_address,
            audit_record.user_agent,
            prev_checksum
        );
        
        -- Check integrity
        IF audit_record.checksum = calculated_checksum AND 
           (audit_record.previous_checksum = prev_checksum OR 
            (audit_record.previous_checksum IS NULL AND prev_checksum IS NULL)) THEN
            RETURN QUERY SELECT audit_record.id, true, NULL::TEXT;
        ELSE
            RETURN QUERY SELECT audit_record.id, false, 
                'Checksum mismatch: expected ' || calculated_checksum || 
                ', got ' || audit_record.checksum;
        END IF;
    END LOOP;
END;
$$ language 'plpgsql';

-- Table comments for documentation
COMMENT ON TABLE users IS 'User authentication and profile information with role-based access';
COMMENT ON TABLE user_sessions IS 'User session management with security tracking and IP validation';
COMMENT ON TABLE audit_logs IS 'Immutable audit trail with cryptographic integrity for compliance (SOX, HIPAA, SOC 2)';
COMMENT ON TABLE cluster_configs IS 'Kubernetes cluster configurations with encrypted kubeconfig storage';