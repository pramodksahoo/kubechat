-- KubeChat Database Schema - Enhanced for Security and Compliance

-- Users table with authentication support
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user', 'viewer')),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_login TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- User sessions with secure token management
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    is_active BOOLEAN DEFAULT TRUE
);

-- Immutable audit logs with cryptographic integrity
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
    -- Immutability protection
    checksum VARCHAR(64) NOT NULL,
    previous_checksum VARCHAR(64)
);

-- Kubernetes cluster configurations with encryption
CREATE TABLE cluster_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    cluster_name VARCHAR(255) NOT NULL,
    cluster_config TEXT NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, cluster_name)
);

-- Schema migrations tracking
CREATE TABLE schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT NOW() NOT NULL,
    dirty BOOLEAN DEFAULT FALSE
);