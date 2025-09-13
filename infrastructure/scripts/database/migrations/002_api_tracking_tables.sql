-- Migration 002: API Configuration and Usage Tracking Tables  
-- Adds specialized tables for external API management and monitoring
-- Date: 2025-09-12
-- Author: Claude (Dev Agent) - QA Fix for Story 1.4

-- API Configuration Table for centralized API provider configuration
CREATE TABLE api_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_name VARCHAR(100) NOT NULL,
    endpoint_url VARCHAR(500) NOT NULL,
    api_key_secret_name VARCHAR(200) NOT NULL,
    rate_limit_per_minute INTEGER NOT NULL DEFAULT 60,
    timeout_seconds INTEGER NOT NULL DEFAULT 30,
    retry_attempts INTEGER NOT NULL DEFAULT 3,
    circuit_breaker_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_provider_name CHECK (LENGTH(provider_name) >= 1 AND LENGTH(provider_name) <= 100),
    CONSTRAINT valid_endpoint_url CHECK (endpoint_url ~* '^https?://.*'),
    CONSTRAINT valid_rate_limit CHECK (rate_limit_per_minute > 0),
    CONSTRAINT valid_timeout CHECK (timeout_seconds > 0),
    CONSTRAINT valid_retry_attempts CHECK (retry_attempts >= 0)
);

-- API Usage Logs Table for detailed API usage analytics
CREATE TABLE api_usage_logs (
    id BIGSERIAL PRIMARY KEY,
    provider_name VARCHAR(100) NOT NULL,
    endpoint VARCHAR(200) NOT NULL,
    user_id UUID REFERENCES users(id),
    request_size_bytes INTEGER,
    response_size_bytes INTEGER,
    processing_time_ms INTEGER,
    cost_cents INTEGER DEFAULT 0,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    timestamp TIMESTAMP DEFAULT NOW(),
    
    -- Additional tracking fields for correlation with audit logs
    session_id UUID REFERENCES user_sessions(id),
    audit_log_id BIGINT REFERENCES audit_logs(id),
    
    -- Constraints
    CONSTRAINT valid_provider_name_usage CHECK (LENGTH(provider_name) >= 1),
    CONSTRAINT valid_endpoint_usage CHECK (LENGTH(endpoint) >= 1),
    CONSTRAINT valid_request_size CHECK (request_size_bytes >= 0),
    CONSTRAINT valid_response_size CHECK (response_size_bytes >= 0),
    CONSTRAINT valid_processing_time CHECK (processing_time_ms >= 0),
    CONSTRAINT valid_cost CHECK (cost_cents >= 0)
);

-- Performance Indexes for api_configurations
CREATE INDEX idx_api_configurations_provider ON api_configurations(provider_name);
CREATE INDEX idx_api_configurations_created_at ON api_configurations(created_at);
CREATE INDEX idx_api_configurations_updated_at ON api_configurations(updated_at);

-- Performance Indexes for api_usage_logs  
CREATE INDEX idx_api_usage_logs_provider ON api_usage_logs(provider_name);
CREATE INDEX idx_api_usage_logs_endpoint ON api_usage_logs(endpoint);
CREATE INDEX idx_api_usage_logs_user_id ON api_usage_logs(user_id);
CREATE INDEX idx_api_usage_logs_session_id ON api_usage_logs(session_id);
CREATE INDEX idx_api_usage_logs_audit_log_id ON api_usage_logs(audit_log_id);
CREATE INDEX idx_api_usage_logs_timestamp ON api_usage_logs(timestamp);
CREATE INDEX idx_api_usage_logs_success ON api_usage_logs(success);
CREATE INDEX idx_api_usage_logs_cost ON api_usage_logs(cost_cents);

-- Apply updated_at trigger to api_configurations
CREATE TRIGGER update_api_configurations_updated_at BEFORE UPDATE ON api_configurations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Table comments for documentation
COMMENT ON TABLE api_configurations IS 'Centralized configuration for external API providers with circuit breaker and rate limiting settings';
COMMENT ON TABLE api_usage_logs IS 'Detailed analytics and monitoring for external API usage with cost tracking and performance metrics';

-- Insert default OpenAI configuration
INSERT INTO api_configurations (
    provider_name, 
    endpoint_url, 
    api_key_secret_name, 
    rate_limit_per_minute, 
    timeout_seconds, 
    retry_attempts, 
    circuit_breaker_enabled
) VALUES (
    'openai',
    'https://api.openai.com/v1/chat/completions',
    'openai-api-key',
    3, -- Conservative rate limit for free tier
    30,
    3,
    true
);