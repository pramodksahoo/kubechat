-- KubeChat Database Integrity Validation Functions

-- Function to verify audit log integrity
CREATE OR REPLACE FUNCTION verify_audit_log_integrity(p_entry_id BIGINT DEFAULT NULL)
RETURNS TABLE(
    entry_id BIGINT,
    is_valid BOOLEAN,
    calculated_checksum VARCHAR(64),
    stored_checksum VARCHAR(64),
    error_message TEXT
) AS $$
DECLARE
    rec RECORD;
    calc_checksum VARCHAR(64);
BEGIN
    FOR rec IN 
        SELECT * FROM audit_logs 
        WHERE (p_entry_id IS NULL OR id = p_entry_id)
        ORDER BY id
    LOOP
        BEGIN
            calc_checksum := calculate_audit_checksum(
                rec.user_id,
                rec.session_id,
                rec.query_text,
                rec.generated_command,
                rec.safety_level,
                rec.execution_result,
                rec.execution_status,
                rec.cluster_context,
                rec.namespace_context,
                rec.timestamp,
                rec.ip_address,
                rec.user_agent,
                rec.previous_checksum
            );
            
            entry_id := rec.id;
            calculated_checksum := calc_checksum;
            stored_checksum := rec.checksum;
            is_valid := (calc_checksum = rec.checksum);
            error_message := CASE 
                WHEN NOT is_valid THEN 'Checksum mismatch detected'
                ELSE NULL
            END;
            
            RETURN NEXT;
        EXCEPTION WHEN OTHERS THEN
            entry_id := rec.id;
            calculated_checksum := NULL;
            stored_checksum := rec.checksum;
            is_valid := FALSE;
            error_message := SQLERRM;
            RETURN NEXT;
        END;
    END LOOP;
END;
$$ LANGUAGE plpgsql;