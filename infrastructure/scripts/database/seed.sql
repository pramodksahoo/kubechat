-- KubeChat Database Seed Data
-- Development environment seed data with sample users and configurations
-- Date: 2025-01-11
-- Author: James (Full Stack Developer Agent)

-- Insert default admin user (password: admin123)
-- Password hash for 'admin123' with bcrypt cost 14
INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at) VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin',
    'admin@kubechat.dev',
    '$2a$14$8K9kGjFTN5QVH7XWyYyFSuGfEKmYmON5gLZlQqOdx9j5mVzq8H2Ya',
    'admin',
    NOW(),
    NOW()
) ON CONFLICT (username) DO NOTHING;

-- Insert default regular user (password: user123)
-- Password hash for 'user123' with bcrypt cost 14
INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at) VALUES (
    '00000000-0000-0000-0000-000000000002',
    'developer',
    'developer@kubechat.dev',
    '$2a$14$kL7vJzFOKmW9X4QyYzF5muGfEKmYmON5gLZlQqOdx9j5mVzq8H3Kb',
    'user',
    NOW(),
    NOW()
) ON CONFLICT (username) DO NOTHING;

-- Insert viewer user (password: viewer123)
-- Password hash for 'viewer123' with bcrypt cost 14
INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at) VALUES (
    '00000000-0000-0000-0000-000000000003',
    'viewer',
    'viewer@kubechat.dev',
    '$2a$14$nP2qJzGHNmY8V3RxYzG6suGfEKmYmON5gLZlQqOdx9j5mVzq8H4Lc',
    'viewer',
    NOW(),
    NOW()
) ON CONFLICT (username) DO NOTHING;

-- Insert sample cluster configurations (encrypted placeholders)
-- Note: In production, these would be properly encrypted kubeconfig data
INSERT INTO cluster_configs (id, user_id, cluster_name, cluster_config, is_active, created_at, updated_at) VALUES (
    '10000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'local-development',
    '{"apiVersion": "v1", "kind": "Config", "clusters": [{"name": "local-dev", "cluster": {"server": "https://127.0.0.1:6443", "insecure-skip-tls-verify": true}}], "contexts": [{"name": "local-dev", "context": {"cluster": "local-dev", "user": "local-dev"}}], "current-context": "local-dev", "users": [{"name": "local-dev", "user": {"token": "dev-token"}}]}',
    true,
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

INSERT INTO cluster_configs (id, user_id, cluster_name, cluster_config, is_active, created_at, updated_at) VALUES (
    '10000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000002',
    'staging-cluster',
    '{"apiVersion": "v1", "kind": "Config", "clusters": [{"name": "staging", "cluster": {"server": "https://staging.k8s.example.com", "certificate-authority-data": "LS0tLS1CRUdJTi..."}}], "contexts": [{"name": "staging", "context": {"cluster": "staging", "user": "staging-user"}}], "current-context": "staging", "users": [{"name": "staging-user", "user": {"client-certificate-data": "LS0tLS1CRUdJTi...", "client-key-data": "LS0tLS1CRUdJTi..."}}]}',
    false,
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Insert sample user sessions for development testing
INSERT INTO user_sessions (id, user_id, session_token, expires_at, created_at, ip_address, user_agent) VALUES (
    '20000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'dev-admin-session-token-12345678901234567890',
    NOW() + INTERVAL '24 hours',
    NOW(),
    '127.0.0.1',
    'KubeChat-Dev-Client/1.0'
) ON CONFLICT (session_token) DO NOTHING;

INSERT INTO user_sessions (id, user_id, session_token, expires_at, created_at, ip_address, user_agent) VALUES (
    '20000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000002',
    'dev-user-session-token-12345678901234567890',
    NOW() + INTERVAL '24 hours',
    NOW(),
    '127.0.0.1',
    'KubeChat-Dev-Client/1.0'
) ON CONFLICT (session_token) DO NOTHING;

-- Insert sample audit log entries to demonstrate functionality
-- Note: These will automatically get checksums via database triggers
INSERT INTO audit_logs (
    user_id, session_id, query_text, generated_command, safety_level,
    execution_result, execution_status, cluster_context, namespace_context,
    timestamp, ip_address, user_agent
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000001',
    'Show me all pods in the default namespace',
    'kubectl get pods -n default',
    'safe',
    '{"pods": [{"name": "test-pod", "status": "Running"}], "count": 1}',
    'success',
    'local-development',
    'default',
    NOW() - INTERVAL '1 hour',
    '127.0.0.1',
    'KubeChat-Dev-Client/1.0'
);

INSERT INTO audit_logs (
    user_id, session_id, query_text, generated_command, safety_level,
    execution_result, execution_status, cluster_context, namespace_context,
    timestamp, ip_address, user_agent
) VALUES (
    '00000000-0000-0000-0000-000000000002',
    '20000000-0000-0000-0000-000000000002',
    'Delete all deployments in the kube-system namespace',
    'kubectl delete deployments --all -n kube-system',
    'dangerous',
    '{"error": "Operation cancelled due to safety concerns"}',
    'cancelled',
    'local-development',
    'kube-system',
    NOW() - INTERVAL '30 minutes',
    '127.0.0.1',
    'KubeChat-Dev-Client/1.0'
);

INSERT INTO audit_logs (
    user_id, session_id, query_text, generated_command, safety_level,
    execution_result, execution_status, cluster_context, namespace_context,
    timestamp, ip_address, user_agent
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000001',
    'Scale the nginx deployment to 3 replicas',
    'kubectl scale deployment nginx --replicas=3',
    'warning',
    '{"deployment": "nginx", "replicas": 3, "status": "scaled"}',
    'success',
    'local-development',
    'default',
    NOW() - INTERVAL '15 minutes',
    '127.0.0.1',
    'KubeChat-Dev-Client/1.0'
);

-- Log successful seeding
SELECT 'Development seed data inserted successfully' AS seeding_status;