-- KubeChat Database - Core Setup Script
-- Consolidates: Extensions, Schema, Basic Indexes
-- Version: Story 1.8 - Audit Trail and Advanced Compliance Logging

-- =================================================================
-- EXTENSIONS AND CONFIGURATION
-- =================================================================

-- Enable required PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Set database configuration
ALTER DATABASE kubechat SET timezone = 'UTC';
ALTER DATABASE kubechat SET log_statement = 'mod';
ALTER DATABASE kubechat SET log_min_duration_statement = 1000;

-- =================================================================
-- DROP EXISTING TABLES (Clean Recreation)
-- =================================================================

DROP TABLE IF EXISTS audit_logs_archive CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS cluster_configs CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS schema_migrations CASCADE;

-- =================================================================
-- CORE SCHEMA DEFINITION
-- =================================================================

-- Users table with enhanced authentication support
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user', 'viewer', 'auditor', 'compliance_officer')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_login TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE,
    -- Security enhancements
    failed_login_attempts INTEGER DEFAULT 0,
    account_locked_until TIMESTAMPTZ,
    password_changed_at TIMESTAMPTZ DEFAULT NOW(),
    must_change_password BOOLEAN DEFAULT FALSE,
    -- Audit fields
    created_by UUID,
    updated_by UUID
);

-- User sessions with enhanced security
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    -- Security enhancements
    last_activity_at TIMESTAMPTZ DEFAULT NOW(),
    device_fingerprint VARCHAR(255),
    location_info JSONB,
    -- Session metadata
    login_method VARCHAR(50) DEFAULT 'password', -- password, sso, api_key
    security_level VARCHAR(20) DEFAULT 'standard' -- standard, elevated, admin
);

-- Enhanced audit logs with cryptographic integrity
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    session_id UUID REFERENCES user_sessions(id),
    query_text TEXT NOT NULL,
    generated_command TEXT NOT NULL,
    safety_level VARCHAR(20) NOT NULL CHECK (safety_level IN ('safe', 'warning', 'dangerous')),
    execution_result JSONB,
    execution_status VARCHAR(20) NOT NULL CHECK (execution_status IN ('success', 'failed', 'cancelled', 'pending')),
    cluster_context VARCHAR(255),
    namespace_context VARCHAR(255),
    timestamp TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    ip_address INET,
    user_agent TEXT,
    -- Cryptographic integrity fields
    checksum VARCHAR(64) NOT NULL,
    previous_checksum VARCHAR(64),
    -- Enhanced audit fields for Story 1.8
    command_category VARCHAR(50), -- read, write, delete, admin, security
    risk_score NUMERIC(3,2) DEFAULT 0.0 CHECK (risk_score >= 0.0 AND risk_score <= 1.0),
    compliance_tags TEXT[],
    retention_policy VARCHAR(50) DEFAULT 'standard',
    -- Performance and metadata
    execution_duration_ms INTEGER,
    request_id UUID,
    correlation_id UUID,
    trace_id VARCHAR(255)
);

-- Kubernetes cluster configurations with enhanced security
CREATE TABLE cluster_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    cluster_name VARCHAR(255) NOT NULL,
    cluster_config TEXT NOT NULL, -- Encrypted configuration
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    -- Security and compliance fields
    config_hash VARCHAR(64), -- For integrity verification
    encryption_key_id VARCHAR(255), -- Reference to encryption key
    access_policy JSONB, -- RBAC policies
    compliance_level VARCHAR(20) DEFAULT 'standard', -- standard, high, critical
    last_validated_at TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'pending', -- pending, valid, invalid, expired
    -- Constraints
    UNIQUE(user_id, cluster_name)
);

-- Schema migrations tracking with enhanced metadata
CREATE TABLE schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    dirty BOOLEAN DEFAULT FALSE,
    -- Enhanced tracking
    migration_name VARCHAR(255),
    checksum VARCHAR(64),
    execution_time_ms INTEGER,
    applied_by VARCHAR(255) DEFAULT 'system',
    rollback_script TEXT,
    notes TEXT
);

-- Audit logs archive table for retention policies
CREATE TABLE audit_logs_archive (
    LIKE audit_logs INCLUDING ALL,
    -- Additional archive metadata
    archived_at TIMESTAMPTZ DEFAULT NOW(),
    archive_batch_id UUID,
    original_table_name VARCHAR(255) DEFAULT 'audit_logs',
    archive_reason VARCHAR(100) DEFAULT 'retention_policy'
);

-- =================================================================
-- BASIC INDEXES FOR PERFORMANCE
-- =================================================================

-- Users table indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_active ON users(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_users_last_login ON users(last_login DESC);

-- User sessions indexes
CREATE INDEX idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_active ON user_sessions(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_user_sessions_activity ON user_sessions(last_activity_at DESC);

-- Audit logs core indexes
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX idx_audit_logs_session_id ON audit_logs(session_id);
CREATE INDEX idx_audit_logs_safety_level ON audit_logs(safety_level);
CREATE INDEX idx_audit_logs_execution_status ON audit_logs(execution_status);

-- Cluster configs indexes
CREATE INDEX idx_cluster_configs_user_id ON cluster_configs(user_id);
CREATE INDEX idx_cluster_configs_active ON cluster_configs(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_cluster_configs_compliance ON cluster_configs(compliance_level);

-- Archive table indexes
CREATE INDEX idx_audit_logs_archive_archived_at ON audit_logs_archive(archived_at DESC);
CREATE INDEX idx_audit_logs_archive_batch_id ON audit_logs_archive(archive_batch_id);
CREATE INDEX idx_audit_logs_archive_timestamp ON audit_logs_archive(timestamp DESC);

-- =================================================================
-- BASIC CONSTRAINTS AND RULES
-- =================================================================

-- Add update timestamp triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply update triggers
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cluster_configs_updated_at
    BEFORE UPDATE ON cluster_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Session activity update trigger
CREATE OR REPLACE FUNCTION update_session_activity()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_activity_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_session_activity_trigger
    BEFORE UPDATE ON user_sessions
    FOR EACH ROW EXECUTE FUNCTION update_session_activity();

-- =================================================================
-- BASIC SECURITY POLICIES (ROW LEVEL SECURITY)
-- =================================================================

-- Enable RLS on sensitive tables
ALTER TABLE user_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE cluster_configs ENABLE ROW LEVEL SECURITY;

-- Basic RLS policies (will be enhanced in script 002)
CREATE POLICY user_sessions_own_data ON user_sessions
    FOR ALL TO PUBLIC
    USING (user_id = current_setting('app.current_user_id')::UUID);

CREATE POLICY cluster_configs_own_data ON cluster_configs
    FOR ALL TO PUBLIC
    USING (user_id = current_setting('app.current_user_id')::UUID);

-- =================================================================
-- CORE SETUP COMPLETION
-- =================================================================

-- Log the completion of core setup
INSERT INTO schema_migrations (version, migration_name, notes) VALUES (
    '001_core_setup',
    'Core Database Setup',
    'Extensions, schema, basic indexes, and security policies established'
);

-- Grant necessary permissions
GRANT USAGE ON SCHEMA public TO PUBLIC;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO kubechat;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO kubechat;