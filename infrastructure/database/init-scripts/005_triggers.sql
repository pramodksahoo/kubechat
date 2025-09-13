-- KubeChat Database Triggers for Audit Log Integrity

-- Trigger function for audit log integrity
CREATE OR REPLACE FUNCTION audit_log_integrity_trigger() RETURNS TRIGGER AS $$
DECLARE
    prev_checksum VARCHAR(64);
BEGIN
    -- Get the previous checksum from the last audit log entry
    SELECT checksum INTO prev_checksum 
    FROM audit_logs 
    ORDER BY id DESC 
    LIMIT 1;
    
    -- Calculate checksum for the new record
    NEW.previous_checksum := prev_checksum;
    NEW.checksum := calculate_audit_checksum(
        NEW.user_id,
        NEW.session_id,
        NEW.query_text,
        NEW.generated_command,
        NEW.safety_level,
        NEW.execution_result,
        NEW.execution_status,
        NEW.cluster_context,
        NEW.namespace_context,
        NEW.timestamp,
        NEW.ip_address,
        NEW.user_agent,
        prev_checksum
    );
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for audit log integrity
CREATE TRIGGER audit_log_checksum_trigger
    BEFORE INSERT ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION audit_log_integrity_trigger();