-- Migration 003: Command Execution Tables
-- Story 1.6: Kubernetes Cluster Integration and Safe Command Execution

-- Command execution tracking table
CREATE TABLE command_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    session_id UUID REFERENCES user_sessions(id),
    operation_type VARCHAR(50) NOT NULL,  -- get, list, delete, scale, restart, logs
    resource_type VARCHAR(50) NOT NULL,   -- pods, deployments, services, configmaps, secrets
    namespace VARCHAR(253) NOT NULL,
    resource_name VARCHAR(253),
    safety_level VARCHAR(20) NOT NULL,    -- safe, warning, dangerous
    command_text TEXT NOT NULL,
    execution_status VARCHAR(20) NOT NULL, -- pending, pending_approval, approved, executing, completed, failed, timeout
    result_data JSONB,
    error_message TEXT,
    execution_time_ms INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    executed_at TIMESTAMP,
    completed_at TIMESTAMP,
    
    -- Indexes for performance
    CONSTRAINT valid_operation_type CHECK (operation_type IN ('get', 'list', 'delete', 'scale', 'restart', 'logs')),
    CONSTRAINT valid_resource_type CHECK (resource_type IN ('pods', 'deployments', 'services', 'configmaps', 'secrets')),
    CONSTRAINT valid_safety_level CHECK (safety_level IN ('safe', 'warning', 'dangerous')),
    CONSTRAINT valid_execution_status CHECK (execution_status IN ('pending', 'pending_approval', 'approved', 'executing', 'completed', 'failed', 'timeout'))
);

-- Command approval workflow table (for dangerous operations)
CREATE TABLE command_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    command_execution_id UUID NOT NULL REFERENCES command_executions(id) ON DELETE CASCADE,
    requested_by_user_id UUID NOT NULL REFERENCES users(id),
    approved_by_user_id UUID REFERENCES users(id),
    approval_status VARCHAR(20) NOT NULL, -- pending, approved, rejected, expired
    approval_reason TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    decided_at TIMESTAMP,
    
    CONSTRAINT valid_approval_status CHECK (approval_status IN ('pending', 'approved', 'rejected', 'expired'))
);

-- Rollback information for reversible operations
CREATE TABLE command_rollback_info (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES command_executions(id) ON DELETE CASCADE,
    original_command TEXT NOT NULL,
    rollback_command TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'available', -- available, executed, failed
    created_at TIMESTAMP DEFAULT NOW(),
    executed_at TIMESTAMP,
    
    CONSTRAINT valid_rollback_status CHECK (status IN ('available', 'executed', 'failed'))
);

-- Indexes for command_executions table
CREATE INDEX idx_command_executions_user_id ON command_executions(user_id);
CREATE INDEX idx_command_executions_session_id ON command_executions(session_id);
CREATE INDEX idx_command_executions_status ON command_executions(execution_status);
CREATE INDEX idx_command_executions_safety_level ON command_executions(safety_level);
CREATE INDEX idx_command_executions_created_at ON command_executions(created_at DESC);
CREATE INDEX idx_command_executions_operation_resource ON command_executions(operation_type, resource_type);
CREATE INDEX idx_command_executions_namespace ON command_executions(namespace);

-- Indexes for command_approvals table
CREATE INDEX idx_command_approvals_execution_id ON command_approvals(command_execution_id);
CREATE INDEX idx_command_approvals_requested_by ON command_approvals(requested_by_user_id);
CREATE INDEX idx_command_approvals_approved_by ON command_approvals(approved_by_user_id);
CREATE INDEX idx_command_approvals_status ON command_approvals(approval_status);
CREATE INDEX idx_command_approvals_expires_at ON command_approvals(expires_at);
CREATE INDEX idx_command_approvals_created_at ON command_approvals(created_at DESC);

-- Indexes for command_rollback_info table
CREATE INDEX idx_command_rollback_info_execution_id ON command_rollback_info(execution_id);
CREATE INDEX idx_command_rollback_info_status ON command_rollback_info(status);
CREATE INDEX idx_command_rollback_info_created_at ON command_rollback_info(created_at DESC);

-- Grant permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON command_executions TO kubechat_api;
GRANT SELECT, INSERT, UPDATE, DELETE ON command_approvals TO kubechat_api;
GRANT SELECT, INSERT, UPDATE, DELETE ON command_rollback_info TO kubechat_api;

-- Comments for documentation
COMMENT ON TABLE command_executions IS 'Tracks Kubernetes command executions with safety classification and approval workflow';
COMMENT ON TABLE command_approvals IS 'Manages approval workflow for dangerous Kubernetes operations';
COMMENT ON TABLE command_rollback_info IS 'Stores rollback information for reversible Kubernetes operations';

COMMENT ON COLUMN command_executions.safety_level IS 'Safety classification: safe (immediate execution), warning (logged execution), dangerous (requires approval)';
COMMENT ON COLUMN command_executions.execution_status IS 'Current status of command execution workflow';
COMMENT ON COLUMN command_executions.result_data IS 'JSON result data from successful command execution';
COMMENT ON COLUMN command_executions.execution_time_ms IS 'Total execution time in milliseconds';

COMMENT ON COLUMN command_approvals.expires_at IS 'Approval request expiration timestamp for security';
COMMENT ON COLUMN command_approvals.approval_reason IS 'Reason for approval or rejection decision';

COMMENT ON COLUMN command_rollback_info.original_command IS 'Original command that was executed';
COMMENT ON COLUMN command_rollback_info.rollback_command IS 'Generated rollback command to undo the operation';