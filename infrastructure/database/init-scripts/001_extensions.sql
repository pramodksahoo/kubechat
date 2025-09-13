-- KubeChat Database Extensions
-- Enable required PostgreSQL extensions

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Set timezone
ALTER DATABASE kubechat SET timezone = 'UTC';