-- KubeChat Development Seed Data
-- This script populates the database with development data

SET search_path TO kubechat, public;

-- Insert default admin user
INSERT INTO users (id, username, email, password_hash, first_name, last_name, roles, is_active)
VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    'admin',
    'admin@kubechat.dev',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: password
    'Admin',
    'User',
    ARRAY['admin', 'user'],
    TRUE
) ON CONFLICT (username) DO NOTHING;

-- Insert developer user
INSERT INTO users (id, username, email, password_hash, first_name, last_name, roles, is_active)
VALUES (
    '550e8400-e29b-41d4-a716-446655440001',
    'developer',
    'dev@kubechat.dev',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: password
    'Developer',
    'User',
    ARRAY['user'],
    TRUE
) ON CONFLICT (username) DO NOTHING;

-- Insert read-only user
INSERT INTO users (id, username, email, password_hash, first_name, last_name, roles, is_active)
VALUES (
    '550e8400-e29b-41d4-a716-446655440002',
    'viewer',
    'viewer@kubechat.dev',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: password
    'Read Only',
    'User',
    ARRAY['viewer'],
    TRUE
) ON CONFLICT (username) DO NOTHING;

-- Insert default local cluster
INSERT INTO clusters (id, name, display_name, description, kubeconfig, created_by)
VALUES (
    '660e8400-e29b-41d4-a716-446655440000',
    'local-dev',
    'Local Development Cluster',
    'Rancher Desktop development cluster',
    'placeholder-kubeconfig-will-be-updated-by-app',
    '550e8400-e29b-41d4-a716-446655440000'
) ON CONFLICT (name) DO NOTHING;

-- Insert development query session
INSERT INTO query_sessions (id, user_id, cluster_id, name, description)
VALUES (
    '770e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440000',
    '660e8400-e29b-41d4-a716-446655440000',
    'Development Session',
    'Default development session for testing'
) ON CONFLICT DO NOTHING;

-- Insert sample queries for development
INSERT INTO queries (id, session_id, user_id, natural_language, generated_command, safety_level, execution_status)
VALUES 
(
    '880e8400-e29b-41d4-a716-446655440000',
    '770e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440000',
    'Show me all pods in the default namespace',
    'kubectl get pods -n default',
    'safe',
    'completed'
),
(
    '880e8400-e29b-41d4-a716-446655440001',
    '770e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440000',
    'List all services',
    'kubectl get services --all-namespaces',
    'safe',
    'completed'
),
(
    '880e8400-e29b-41d4-a716-446655440002',
    '770e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440000',
    'Show cluster information',
    'kubectl cluster-info',
    'safe',
    'completed'
);

-- Insert default AI models
INSERT INTO ai_models (id, name, provider, model_id, description, configuration)
VALUES 
(
    '990e8400-e29b-41d4-a716-446655440000',
    'Ollama Phi-3.5-mini',
    'ollama',
    'phi-3.5-mini',
    'Local Phi-3.5-mini model for Kubernetes command generation',
    '{"endpoint": "http://ollama:11434", "temperature": 0.1, "max_tokens": 500}'
),
(
    '990e8400-e29b-41d4-a716-446655440001',
    'OpenAI GPT-4',
    'openai',
    'gpt-4',
    'OpenAI GPT-4 model for complex queries (fallback)',
    '{"temperature": 0.1, "max_tokens": 500}'
);

-- Insert user preferences for development users
INSERT INTO user_preferences (user_id, theme, language, timezone, preferences)
VALUES 
(
    '550e8400-e29b-41d4-a716-446655440000',
    'dark',
    'en',
    'UTC',
    '{"preferred_ai_model": "990e8400-e29b-41d4-a716-446655440000", "auto_execute": false, "show_command_preview": true}'
),
(
    '550e8400-e29b-41d4-a716-446655440001',
    'light',
    'en',
    'UTC',
    '{"preferred_ai_model": "990e8400-e29b-41d4-a716-446655440000", "auto_execute": false, "show_command_preview": true}'
),
(
    '550e8400-e29b-41d4-a716-446655440002',
    'light',
    'en',
    'UTC',
    '{"preferred_ai_model": "990e8400-e29b-41d4-a716-446655440000", "auto_execute": false, "show_command_preview": true}'
);

-- Insert some audit log entries for development
INSERT INTO audit_log_entries (user_id, action, resource_type, resource_id, cluster_id, details, ip_address, user_agent)
VALUES 
(
    '550e8400-e29b-41d4-a716-446655440000',
    'user.login',
    'user',
    '550e8400-e29b-41d4-a716-446655440000',
    NULL,
    '{"login_method": "password", "success": true}',
    '127.0.0.1',
    'KubeChat-Web/1.0.0'
),
(
    '550e8400-e29b-41d4-a716-446655440000',
    'query.execute',
    'query',
    '880e8400-e29b-41d4-a716-446655440000',
    '660e8400-e29b-41d4-a716-446655440000',
    '{"command": "kubectl get pods -n default", "execution_time_ms": 245}',
    '127.0.0.1',
    'KubeChat-Web/1.0.0'
);

-- Update statistics
ANALYZE;