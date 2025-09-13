-- KubeChat Database Initial Data

-- Insert initial migration record
INSERT INTO schema_migrations (version) VALUES ('001_initial_schema');

-- Create default admin user (password: admin123)
INSERT INTO users (id, username, email, password_hash, role) VALUES (
    gen_random_uuid(),
    'admin',
    'admin@kubechat.dev',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewNxn0h8P8O9gOJm',
    'admin'
);