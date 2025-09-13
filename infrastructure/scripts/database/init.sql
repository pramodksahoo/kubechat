-- KubeChat Database Initialization Script
-- This script initializes the database with required extensions and permissions
-- Date: 2025-01-11
-- Author: James (Full Stack Developer Agent)

-- Enable required PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create database user if not exists (for production setups)
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'kubechat') THEN
        CREATE ROLE kubechat WITH LOGIN PASSWORD 'kubechat';
    END IF;
END
$$;

-- Grant necessary permissions
GRANT CONNECT ON DATABASE kubechat TO kubechat;
GRANT USAGE ON SCHEMA public TO kubechat;
GRANT CREATE ON SCHEMA public TO kubechat;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO kubechat;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO kubechat;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO kubechat;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO kubechat;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO kubechat;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO kubechat;

-- Database configuration for optimal performance
ALTER DATABASE kubechat SET timezone = 'UTC';

-- Log initialization
SELECT 'KubeChat database initialized successfully' AS initialization_status;