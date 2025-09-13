-- Migration: 003_rollback_tables.sql
-- Description: Create tables for command execution rollback functionality
-- Story: 1.6 - Kubernetes Cluster Integration and Safe Command Execution

-- Rollback Plans Table
CREATE TABLE rollback_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    command_execution_id UUID NOT NULL,
    user_id UUID NOT NULL,
    session_id UUID NOT NULL,
    original_command JSONB NOT NULL,
    rollback_steps JSONB NOT NULL, -- Array of rollback steps
    status VARCHAR(20) NOT NULL DEFAULT 'planned', -- planned, executing, completed, failed, cancelled, invalid
    reason TEXT NOT NULL,
    estimated_duration INTERVAL NOT NULL,
    validation JSONB, -- Rollback validation results
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Constraints
    CONSTRAINT fk_rollback_plans_command_execution 
        FOREIGN KEY (command_execution_id) REFERENCES command_executions(id) ON DELETE CASCADE,
    CONSTRAINT fk_rollback_plans_user 
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_rollback_plans_session 
        FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    CONSTRAINT chk_rollback_plan_status 
        CHECK (status IN ('planned', 'executing', 'completed', 'failed', 'cancelled', 'invalid'))
);

-- Rollback Executions Table
CREATE TABLE rollback_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'executing', -- executing, completed, failed, cancelled
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    execution_log JSONB NOT NULL DEFAULT '[]'::jsonb, -- Array of step results
    error TEXT,
    
    -- Constraints
    CONSTRAINT fk_rollback_executions_plan 
        FOREIGN KEY (plan_id) REFERENCES rollback_plans(id) ON DELETE CASCADE,
    CONSTRAINT fk_rollback_executions_user 
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_rollback_execution_status 
        CHECK (status IN ('executing', 'completed', 'failed', 'cancelled')),
    CONSTRAINT chk_completed_at_when_completed 
        CHECK ((status = 'executing') OR (completed_at IS NOT NULL))
);

-- Indexes for Performance
CREATE INDEX idx_rollback_plans_command_execution ON rollback_plans(command_execution_id);
CREATE INDEX idx_rollback_plans_user ON rollback_plans(user_id);
CREATE INDEX idx_rollback_plans_session ON rollback_plans(session_id);
CREATE INDEX idx_rollback_plans_status ON rollback_plans(status);
CREATE INDEX idx_rollback_plans_expires_at ON rollback_plans(expires_at);
CREATE INDEX idx_rollback_plans_created_at ON rollback_plans(created_at);

CREATE INDEX idx_rollback_executions_plan ON rollback_executions(plan_id);
CREATE INDEX idx_rollback_executions_user ON rollback_executions(user_id);
CREATE INDEX idx_rollback_executions_status ON rollback_executions(status);
CREATE INDEX idx_rollback_executions_started_at ON rollback_executions(started_at);
CREATE INDEX idx_rollback_executions_completed_at ON rollback_executions(completed_at) WHERE completed_at IS NOT NULL;

-- Comments for Documentation
COMMENT ON TABLE rollback_plans IS 'Plans for rolling back command executions with validation and step details';
COMMENT ON TABLE rollback_executions IS 'Execution records for rollback plans with detailed logs and status tracking';

COMMENT ON COLUMN rollback_plans.original_command IS 'Original command that is being rolled back (JSONB structure)';
COMMENT ON COLUMN rollback_plans.rollback_steps IS 'Array of rollback steps to be executed (JSONB array)';
COMMENT ON COLUMN rollback_plans.validation IS 'Validation results including warnings and blocking issues (JSONB structure)';
COMMENT ON COLUMN rollback_plans.estimated_duration IS 'Estimated time to complete the rollback operation';

COMMENT ON COLUMN rollback_executions.execution_log IS 'Detailed log of each rollback step execution with results (JSONB array)';
COMMENT ON COLUMN rollback_executions.error IS 'Error message if rollback execution failed';

-- Security Policies (Row Level Security)
ALTER TABLE rollback_plans ENABLE ROW LEVEL SECURITY;
ALTER TABLE rollback_executions ENABLE ROW LEVEL SECURITY;

-- Users can only see their own rollback plans and executions
CREATE POLICY rollback_plans_user_access ON rollback_plans
    FOR ALL USING (user_id = current_setting('app.current_user_id')::UUID);

CREATE POLICY rollback_executions_user_access ON rollback_executions
    FOR ALL USING (user_id = current_setting('app.current_user_id')::UUID);

-- Data Retention and Cleanup
-- Automatically clean up expired rollback plans (handled by application cleanup routines)
-- Consider adding a trigger for automatic cleanup of very old completed executions