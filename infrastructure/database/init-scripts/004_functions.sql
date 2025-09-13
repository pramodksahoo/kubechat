-- KubeChat Database Functions for Audit Log Integrity

-- Audit log integrity functions
CREATE OR REPLACE FUNCTION calculate_audit_checksum(
    p_user_id UUID,
    p_session_id UUID,
    p_query_text TEXT,
    p_generated_command TEXT,
    p_safety_level VARCHAR(20),
    p_execution_result JSONB,
    p_execution_status VARCHAR(20),
    p_cluster_context VARCHAR(255),
    p_namespace_context VARCHAR(255),
    p_timestamp TIMESTAMP,
    p_ip_address INET,
    p_user_agent TEXT,
    p_previous_checksum VARCHAR(64)
) RETURNS VARCHAR(64) AS $$
DECLARE
    checksum_input TEXT;
BEGIN
    checksum_input := CONCAT(
        COALESCE(p_user_id::text, ''),
        '|',
        COALESCE(p_session_id::text, ''),
        '|',
        COALESCE(p_query_text, ''),
        '|',
        COALESCE(p_generated_command, ''),
        '|',
        COALESCE(p_safety_level, ''),
        '|',
        COALESCE(p_execution_result::text, ''),
        '|',
        COALESCE(p_execution_status, ''),
        '|',
        COALESCE(p_cluster_context, ''),
        '|',
        COALESCE(p_namespace_context, ''),
        '|',
        COALESCE(p_timestamp::text, ''),
        '|',
        COALESCE(p_ip_address::text, ''),
        '|',
        COALESCE(p_user_agent, ''),
        '|',
        COALESCE(p_previous_checksum, '')
    );
    
    RETURN encode(digest(checksum_input, 'sha256'), 'hex');
END;
$$ LANGUAGE plpgsql IMMUTABLE;