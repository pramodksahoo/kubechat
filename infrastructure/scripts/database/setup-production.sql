-- KubeChat Production Database Setup
-- This script sets up the production database with proper security and optimization
-- Date: 2025-01-11
-- Author: James (Full Stack Developer Agent)

-- Create database if it doesn't exist (run as superuser)
-- CREATE DATABASE kubechat WITH ENCODING 'UTF8' LC_COLLATE='en_US.UTF-8' LC_CTYPE='en_US.UTF-8';

-- Connect to kubechat database
\c kubechat;

-- Enable required PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create application user with limited privileges
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'kubechat_app') THEN
        CREATE ROLE kubechat_app WITH LOGIN PASSWORD 'CHANGE_THIS_PASSWORD_IN_PRODUCTION';
    END IF;
END
$$;

-- Create backup user with read-only access
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'kubechat_backup') THEN
        CREATE ROLE kubechat_backup WITH LOGIN PASSWORD 'CHANGE_THIS_BACKUP_PASSWORD';
    END IF;
END
$$;

-- Create monitoring user with limited access
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'kubechat_monitor') THEN
        CREATE ROLE kubechat_monitor WITH LOGIN PASSWORD 'CHANGE_THIS_MONITOR_PASSWORD';
    END IF;
END
$$;

-- Grant necessary permissions to application user
GRANT CONNECT ON DATABASE kubechat TO kubechat_app;
GRANT USAGE ON SCHEMA public TO kubechat_app;
GRANT CREATE ON SCHEMA public TO kubechat_app;

-- Grant backup user read-only access
GRANT CONNECT ON DATABASE kubechat TO kubechat_backup;
GRANT USAGE ON SCHEMA public TO kubechat_backup;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO kubechat_backup;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO kubechat_backup;

-- Grant monitoring user limited access
GRANT CONNECT ON DATABASE kubechat TO kubechat_monitor;
GRANT USAGE ON SCHEMA public TO kubechat_monitor;

-- Configure database for production performance
ALTER DATABASE kubechat SET timezone = 'UTC';
ALTER DATABASE kubechat SET log_statement = 'mod';
ALTER DATABASE kubechat SET log_min_duration_statement = 1000;

-- Set up connection limits
ALTER ROLE kubechat_app CONNECTION LIMIT 100;
ALTER ROLE kubechat_backup CONNECTION LIMIT 5;
ALTER ROLE kubechat_monitor CONNECTION LIMIT 10;

-- Create tablespace for audit logs (optional - for better performance)
-- This would typically be on a separate disk for performance
-- CREATE TABLESPACE audit_logs_space LOCATION '/var/lib/postgresql/audit_logs';

-- Production security settings
-- Revoke default permissions from public schema
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON DATABASE kubechat FROM PUBLIC;

-- Log production setup completion
SELECT 'Production database setup completed successfully' AS setup_status;

-- Display important security reminders
SELECT '
IMPORTANT SECURITY REMINDERS:
1. Change all default passwords immediately
2. Configure PostgreSQL authentication (pg_hba.conf)
3. Enable SSL/TLS connections
4. Set up regular backups
5. Configure log rotation
6. Monitor audit log integrity
7. Set up alerting for security events
' AS security_reminders;