package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// KubernetesCommandExecution represents a Kubernetes command execution with approval workflow
type KubernetesCommandExecution struct {
	ID              uuid.UUID                  `json:"id" db:"id"`
	UserID          uuid.UUID                  `json:"user_id" db:"user_id"`
	SessionID       uuid.UUID                  `json:"session_id" db:"session_id"`
	Command         KubernetesOperation        `json:"command" db:"command"`
	SafetyLevel     string                     `json:"safety_level" db:"safety_level"` // safe, warning, dangerous
	Status          string                     `json:"status" db:"status"`             // pending, approved, executing, completed, failed, timeout
	ApprovalID      *uuid.UUID                 `json:"approval_id,omitempty" db:"approval_id"`
	Result          *KubernetesOperationResult `json:"result,omitempty" db:"result"`
	CreatedAt       time.Time                  `json:"created_at" db:"created_at"`
	ExecutedAt      *time.Time                 `json:"executed_at,omitempty" db:"executed_at"`
	CompletedAt     *time.Time                 `json:"completed_at,omitempty" db:"completed_at"`
	ExecutionTimeMS *int                       `json:"execution_time_ms,omitempty" db:"execution_time_ms"`
	ErrorMessage    string                     `json:"error_message,omitempty" db:"error_message"`
}

// RollbackPlan represents a plan for rolling back a command execution
type RollbackPlan struct {
	ID                 uuid.UUID           `json:"id" db:"id"`
	CommandExecutionID uuid.UUID           `json:"command_execution_id" db:"command_execution_id"`
	UserID             uuid.UUID           `json:"user_id" db:"user_id"`
	SessionID          uuid.UUID           `json:"session_id" db:"session_id"`
	OriginalCommand    KubernetesOperation `json:"original_command" db:"original_command"`
	RollbackSteps      []RollbackStep      `json:"rollback_steps" db:"rollback_steps"`
	Status             string              `json:"status" db:"status"` // planned, executing, completed, failed, cancelled, invalid
	Reason             string              `json:"reason" db:"reason"`
	EstimatedDuration  time.Duration       `json:"estimated_duration" db:"estimated_duration" swaggertype:"integer" format:"int64" example:"30000000000"`
	Validation         *RollbackValidation `json:"validation,omitempty" db:"validation"`
	CreatedAt          time.Time           `json:"created_at" db:"created_at"`
	ExpiresAt          time.Time           `json:"expires_at" db:"expires_at"`
}

// RollbackStep represents a single step in a rollback plan
type RollbackStep struct {
	StepNumber  int                    `json:"step_number" db:"step_number"`
	Operation   string                 `json:"operation" db:"operation"`     // create, delete, patch, scale
	Resource    string                 `json:"resource" db:"resource"`       // pods, services, deployments, etc.
	Name        string                 `json:"name" db:"name"`               // resource name
	Namespace   string                 `json:"namespace" db:"namespace"`     // resource namespace
	Description string                 `json:"description" db:"description"` // human readable description
	Data        map[string]any         `json:"data,omitempty" db:"data"`     // operation data (manifest, patch, etc.)
}

// RollbackExecution represents the execution of a rollback plan
type RollbackExecution struct {
	ID           uuid.UUID            `json:"id" db:"id"`
	PlanID       uuid.UUID            `json:"plan_id" db:"plan_id"`
	UserID       uuid.UUID            `json:"user_id" db:"user_id"`
	Status       string               `json:"status" db:"status"` // executing, completed, failed, cancelled
	StartedAt    time.Time            `json:"started_at" db:"started_at"`
	CompletedAt  *time.Time           `json:"completed_at,omitempty" db:"completed_at"`
	ExecutionLog []RollbackStepResult `json:"execution_log" db:"execution_log"`
	Error        *string              `json:"error,omitempty" db:"error"`
}

// RollbackStepResult represents the result of executing a rollback step
type RollbackStepResult struct {
	StepNumber  int        `json:"step_number" db:"step_number"`
	Operation   string     `json:"operation" db:"operation"`
	Resource    string     `json:"resource" db:"resource"`
	Name        string     `json:"name" db:"name"`
	Namespace   string     `json:"namespace" db:"namespace"`
	Status      string     `json:"status" db:"status"` // executing, completed, failed
	StartedAt   time.Time  `json:"started_at" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	Error       string     `json:"error,omitempty" db:"error"`
	Output      string     `json:"output,omitempty" db:"output"`
}

// RollbackValidation represents validation results for a rollback plan
type RollbackValidation struct {
	ExecutionID uuid.UUID `json:"execution_id" db:"execution_id"`
	IsValid     bool      `json:"is_valid" db:"is_valid"`
	Reasons     []string  `json:"reasons" db:"reasons"`   // reasons why rollback is invalid
	Warnings    []string  `json:"warnings" db:"warnings"` // non-blocking warnings
	CheckedAt   time.Time `json:"checked_at" db:"checked_at"`
}

// RollbackStatus represents the current status of a rollback operation
type RollbackStatus struct {
	RollbackID  uuid.UUID  `json:"rollback_id"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Progress    float64    `json:"progress"` // completion percentage (0-100)
	CurrentStep string     `json:"current_step"`
	TotalSteps  int        `json:"total_steps"`
	Error       *string    `json:"error,omitempty"`
}

// CommandApproval represents the approval workflow for dangerous operations
type CommandApproval struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	CommandExecutionID uuid.UUID  `json:"command_execution_id" db:"command_execution_id"`
	RequestedByUserID  uuid.UUID  `json:"requested_by_user_id" db:"requested_by_user_id"`
	ApprovedByUserID   *uuid.UUID `json:"approved_by_user_id,omitempty" db:"approved_by_user_id"`
	Status             string     `json:"status" db:"status"` // pending, approved, rejected, expired
	Reason             string     `json:"reason,omitempty" db:"reason"`
	ExpiresAt          time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	DecidedAt          *time.Time `json:"decided_at,omitempty" db:"decided_at"`
}

// CommandExecutionRequest represents a request to execute a command
type CommandExecutionRequest struct {
	Command     KubernetesOperation `json:"command"`
	UserID      uuid.UUID           `json:"user_id"`
	SessionID   uuid.UUID           `json:"session_id"`
	SafetyLevel string              `json:"safety_level,omitempty"`
}

// CommandApprovalRequest represents a request to approve a command
type CommandApprovalRequest struct {
	ExecutionID uuid.UUID `json:"execution_id"`
	UserID      uuid.UUID `json:"user_id"`
	Decision    string    `json:"decision"` // approve, reject
	Reason      string    `json:"reason,omitempty"`
}

// CommandRollbackRequest represents a request to rollback a command
type CommandRollbackRequest struct {
	ExecutionID uuid.UUID `json:"execution_id"`
	UserID      uuid.UUID `json:"user_id"`
	Reason      string    `json:"reason,omitempty"`
}

// RollbackInfo represents rollback information for reversible operations
type RollbackInfo struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	ExecutionID     uuid.UUID  `json:"execution_id" db:"execution_id"`
	OriginalCommand string     `json:"original_command" db:"original_command"`
	RollbackCommand string     `json:"rollback_command" db:"rollback_command"`
	Status          string     `json:"status" db:"status"` // available, executed, failed
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	ExecutedAt      *time.Time `json:"executed_at,omitempty" db:"executed_at"`
}

// Validate validates a CommandExecutionRequest
func (req *CommandExecutionRequest) Validate() error {
	if req.UserID == uuid.Nil {
		return ErrInvalidInput{Field: "user_id", Message: "user_id is required"}
	}

	if req.SessionID == uuid.Nil {
		return ErrInvalidInput{Field: "session_id", Message: "session_id is required"}
	}

	// Validate the command itself
	if err := req.Command.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	// Validate safety level if provided
	if req.SafetyLevel != "" {
		validSafetyLevels := map[string]bool{
			"safe":      true,
			"warning":   true,
			"dangerous": true,
		}
		if !validSafetyLevels[req.SafetyLevel] {
			return ErrInvalidInput{Field: "safety_level", Message: "invalid safety level"}
		}
	}

	return nil
}

// Validate validates a CommandApprovalRequest
func (req *CommandApprovalRequest) Validate() error {
	if req.ExecutionID == uuid.Nil {
		return ErrInvalidInput{Field: "execution_id", Message: "execution_id is required"}
	}

	if req.UserID == uuid.Nil {
		return ErrInvalidInput{Field: "user_id", Message: "user_id is required"}
	}

	validDecisions := map[string]bool{
		"approve": true,
		"reject":  true,
	}
	if !validDecisions[req.Decision] {
		return ErrInvalidInput{Field: "decision", Message: "decision must be 'approve' or 'reject'"}
	}

	return nil
}

// Validate validates a CommandRollbackRequest
func (req *CommandRollbackRequest) Validate() error {
	if req.ExecutionID == uuid.Nil {
		return ErrInvalidInput{Field: "execution_id", Message: "execution_id is required"}
	}

	if req.UserID == uuid.Nil {
		return ErrInvalidInput{Field: "user_id", Message: "user_id is required"}
	}

	return nil
}

// IsExpired checks if an approval has expired
func (approval *CommandApproval) IsExpired() bool {
	return time.Now().After(approval.ExpiresAt)
}

// RequiresApproval determines if a command execution requires approval
func (exec *KubernetesCommandExecution) RequiresApproval() bool {
	return exec.SafetyLevel == "dangerous" || exec.SafetyLevel == "warning"
}

// IsReversible determines if a command can be rolled back
func (exec *KubernetesCommandExecution) IsReversible() bool {
	// Scale operations are reversible (can be scaled back)
	if exec.Command.Operation == "scale" {
		return true
	}

	// Restart operations are not reversible (can't "un-restart")
	if exec.Command.Operation == "restart" {
		return false
	}

	// Delete operations are not reversible for safety
	if exec.Command.Operation == "delete" {
		return false
	}

	// Read operations don't need rollback
	if exec.Command.Operation == "get" || exec.Command.Operation == "list" || exec.Command.Operation == "logs" {
		return false
	}

	return false
}

// GetDescription returns a human-readable description of the execution
func (exec *KubernetesCommandExecution) GetDescription() string {
	return exec.Command.GetDescription()
}

// GetApprovalDescription returns a description for approval workflow
func (exec *KubernetesCommandExecution) GetApprovalDescription() string {
	return fmt.Sprintf("Command execution requires approval: %s (Safety Level: %s)",
		exec.GetDescription(), exec.SafetyLevel)
}
