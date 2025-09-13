-- Migration: 005_command_execution_fixed.sql
-- Description: Create command execution tables (fixed version)
-- Story: 1.6 - Kubernetes Cluster Integration and Safe Command Execution

-- Command Executions Table (based on KubernetesCommandExecution model)
CREATE TABLE IF NOT EXISTS command_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id UUID NOT NULL REFERENCES user_sessions(id) ON DELETE CASCADE,
    command JSONB NOT NULL, -- KubernetesOperation as JSON
    safety_level VARCHAR(20) NOT NULL DEFAULT 'safe',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    approval_id UUID, -- Will reference command_approvals(id) after that table is created
    result JSONB, -- KubernetesOperationResult as JSON
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    executed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    execution_time_ms INTEGER,
    error_message TEXT,
    
    -- Constraints based on the model
    CONSTRAINT chk_command_execution_safety_level 
        CHECK (safety_level IN ('safe', 'warning', 'dangerous')),
    CONSTRAINT chk_command_execution_status 
        CHECK (status IN ('pending', 'pending_approval', 'approved', 'executing', 'completed', 'failed', 'timeout', 'cancelled', 'rejected')),
    CONSTRAINT chk_completed_at_when_completed 
        CHECK ((status IN ('pending', 'pending_approval', 'approved', 'executing')) OR (completed_at IS NOT NULL))
);

-- Command Approvals Table (based on CommandApproval model)
CREATE TABLE IF NOT EXISTS command_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    command_execution_id UUID NOT NULL REFERENCES command_executions(id) ON DELETE CASCADE,
    requested_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    approved_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reason TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    decided_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints based on the model
    CONSTRAINT chk_command_approval_status 
        CHECK (status IN ('pending', 'approved', 'rejected', 'expired')),
    CONSTRAINT chk_decided_at_when_decided 
        CHECK ((status = 'pending') OR (decided_at IS NOT NULL))
);

-- Add foreign key constraint for command_executions.approval_id after command_approvals table exists
ALTER TABLE command_executions 
ADD CONSTRAINT fk_command_executions_approval 
FOREIGN KEY (approval_id) REFERENCES command_approvals(id) ON DELETE SET NULL;

-- Rollback Plans Table (based on RollbackPlan model)
CREATE TABLE IF NOT EXISTS rollback_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    command_execution_id UUID NOT NULL REFERENCES command_executions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id UUID NOT NULL REFERENCES user_sessions(id) ON DELETE CASCADE,
    original_command JSONB NOT NULL, -- KubernetesOperation as JSON
    rollback_steps JSONB NOT NULL, -- RollbackStep[] as JSON
    status VARCHAR(20) NOT NULL DEFAULT 'planned',
    reason TEXT NOT NULL,
    estimated_duration INTERVAL NOT NULL,
    validation JSONB, -- RollbackValidation as JSON (optional)
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Constraints based on the model
    CONSTRAINT chk_rollback_plan_status 
        CHECK (status IN ('planned', 'executing', 'completed', 'failed', 'cancelled', 'invalid'))
);

-- Rollback Executions Table (based on RollbackExecution model)
CREATE TABLE IF NOT EXISTS rollback_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES rollback_plans(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'executing',
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    execution_log JSONB NOT NULL DEFAULT '[]'::jsonb, -- RollbackStepResult[] as JSON
    error TEXT,
    
    -- Constraints based on the model
    CONSTRAINT chk_rollback_execution_status 
        CHECK (status IN ('executing', 'completed', 'failed', 'cancelled')),
    CONSTRAINT chk_rollback_completed_at_when_completed 
        CHECK ((status = 'executing') OR (completed_at IS NOT NULL))
);

-- Indexes for Performance (matching the model usage patterns)
CREATE INDEX IF NOT EXISTS idx_command_executions_user_id ON command_executions(user_id);
CREATE INDEX IF NOT EXISTS idx_command_executions_session_id ON command_executions(session_id);  
CREATE INDEX IF NOT EXISTS idx_command_executions_status ON command_executions(status);
CREATE INDEX IF NOT EXISTS idx_command_executions_safety_level ON command_executions(safety_level);
CREATE INDEX IF NOT EXISTS idx_command_executions_created_at ON command_executions(created_at);
CREATE INDEX IF NOT EXISTS idx_command_executions_approval_id ON command_executions(approval_id) WHERE approval_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_command_approvals_command_execution ON command_approvals(command_execution_id);
CREATE INDEX IF NOT EXISTS idx_command_approvals_requested_by ON command_approvals(requested_by_user_id);
CREATE INDEX IF NOT EXISTS idx_command_approvals_approved_by ON command_approvals(approved_by_user_id) WHERE approved_by_user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_command_approvals_status ON command_approvals(status);
CREATE INDEX IF NOT EXISTS idx_command_approvals_expires_at ON command_approvals(expires_at);
CREATE INDEX IF NOT EXISTS idx_command_approvals_created_at ON command_approvals(created_at);

CREATE INDEX IF NOT EXISTS idx_rollback_plans_command_execution ON rollback_plans(command_execution_id);
CREATE INDEX IF NOT EXISTS idx_rollback_plans_user ON rollback_plans(user_id);
CREATE INDEX IF NOT EXISTS idx_rollback_plans_session ON rollback_plans(session_id);
CREATE INDEX IF NOT EXISTS idx_rollback_plans_status ON rollback_plans(status);
CREATE INDEX IF NOT EXISTS idx_rollback_plans_expires_at ON rollback_plans(expires_at);
CREATE INDEX IF NOT EXISTS idx_rollback_plans_created_at ON rollback_plans(created_at);

CREATE INDEX IF NOT EXISTS idx_rollback_executions_plan ON rollback_executions(plan_id);
CREATE INDEX IF NOT EXISTS idx_rollback_executions_user ON rollback_executions(user_id);
CREATE INDEX IF NOT EXISTS idx_rollback_executions_status ON rollback_executions(status);
CREATE INDEX IF NOT EXISTS idx_rollback_executions_started_at ON rollback_executions(started_at);
CREATE INDEX IF NOT EXISTS idx_rollback_executions_completed_at ON rollback_executions(completed_at) WHERE completed_at IS NOT NULL;

-- Comments for Documentation
COMMENT ON TABLE command_executions IS 'Command execution tracking with comprehensive audit trail and safety validation - matches KubernetesCommandExecution model';
COMMENT ON TABLE command_approvals IS 'Approval workflow for dangerous command executions - matches CommandApproval model';
COMMENT ON TABLE rollback_plans IS 'Plans for rolling back command executions with validation and step details - matches RollbackPlan model';
COMMENT ON TABLE rollback_executions IS 'Execution records for rollback plans with detailed logs and status tracking - matches RollbackExecution model';

-- Column comments explaining JSONB structure
COMMENT ON COLUMN command_executions.command IS 'KubernetesOperation JSON: {id, user_id, session_id, operation, resource, namespace, name, context, created_at}';
COMMENT ON COLUMN command_executions.result IS 'KubernetesOperationResult JSON: {operation_id, success, result, error, executed_at, backup_data, previous_state}';
COMMENT ON COLUMN rollback_plans.rollback_steps IS 'RollbackStep[] JSON: [{step_number, operation, resource, name, namespace, description, data}]';
COMMENT ON COLUMN rollback_plans.validation IS 'RollbackValidation JSON: {execution_id, is_valid, reasons, warnings, checked_at}';
COMMENT ON COLUMN rollback_executions.execution_log IS 'RollbackStepResult[] JSON: [{step_number, operation, resource, name, namespace, status, started_at, completed_at, error, output}]';