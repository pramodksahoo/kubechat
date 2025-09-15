package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// CommandRepository defines the interface for command execution data operations
type CommandRepository interface {
	// Command execution operations
	CreateExecution(ctx context.Context, execution *models.KubernetesCommandExecution) error
	GetExecution(ctx context.Context, executionID uuid.UUID) (*models.KubernetesCommandExecution, error)
	GetExecutionsByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*models.KubernetesCommandExecution, error)
	GetExecutionsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.KubernetesCommandExecution, error)
	UpdateExecution(ctx context.Context, execution *models.KubernetesCommandExecution) error
	DeleteExecution(ctx context.Context, executionID uuid.UUID) error

	// Status and filtering operations
	GetExecutionsByStatus(ctx context.Context, status string, limit int) ([]*models.KubernetesCommandExecution, error)
	GetPendingExecutions(ctx context.Context, userID uuid.UUID) ([]*models.KubernetesCommandExecution, error)
	GetCompletedExecutions(ctx context.Context, userID uuid.UUID, since time.Time) ([]*models.KubernetesCommandExecution, error)

	// Approval workflow operations
	CreateApproval(ctx context.Context, approval *models.CommandApproval) error
	GetApproval(ctx context.Context, approvalID uuid.UUID) (*models.CommandApproval, error)
	GetApprovalByExecution(ctx context.Context, executionID uuid.UUID) (*models.CommandApproval, error)
	UpdateApproval(ctx context.Context, approval *models.CommandApproval) error
	GetPendingApprovals(ctx context.Context, userID uuid.UUID) ([]*models.CommandApproval, error)
	GetExpiredApprovals(ctx context.Context) ([]*models.CommandApproval, error)

	// Cleanup and maintenance
	CleanupExpiredExecutions(ctx context.Context, olderThan time.Time) (int, error)
	CleanupExpiredApprovals(ctx context.Context) (int, error)

	// Analytics and reporting
	GetExecutionStats(ctx context.Context, userID uuid.UUID, from, to time.Time) (*models.ExecutionStats, error)

	// Health check
	HealthCheck(ctx context.Context) error
}
