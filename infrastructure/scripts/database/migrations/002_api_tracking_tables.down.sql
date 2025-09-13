-- Migration 002 Down: Remove API Configuration and Usage Tracking Tables
-- Rollback for specialized tables for external API management and monitoring
-- Date: 2025-09-12
-- Author: Claude (Dev Agent) - QA Fix for Story 1.4

-- Remove indexes first
DROP INDEX IF EXISTS idx_api_usage_logs_cost;
DROP INDEX IF EXISTS idx_api_usage_logs_success;
DROP INDEX IF EXISTS idx_api_usage_logs_timestamp;
DROP INDEX IF EXISTS idx_api_usage_logs_audit_log_id;
DROP INDEX IF EXISTS idx_api_usage_logs_session_id;
DROP INDEX IF EXISTS idx_api_usage_logs_user_id;
DROP INDEX IF EXISTS idx_api_usage_logs_endpoint;
DROP INDEX IF EXISTS idx_api_usage_logs_provider;

DROP INDEX IF EXISTS idx_api_configurations_updated_at;
DROP INDEX IF EXISTS idx_api_configurations_created_at;
DROP INDEX IF EXISTS idx_api_configurations_provider;

-- Remove trigger
DROP TRIGGER IF EXISTS update_api_configurations_updated_at ON api_configurations;

-- Drop tables (api_usage_logs first due to foreign key dependency)
DROP TABLE IF EXISTS api_usage_logs;
DROP TABLE IF EXISTS api_configurations;