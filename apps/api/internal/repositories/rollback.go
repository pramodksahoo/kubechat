package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// RollbackRepository defines the interface for rollback data operations
type RollbackRepository interface {
	// Plan operations
	CreatePlan(ctx context.Context, plan *models.RollbackPlan) error
	GetPlan(ctx context.Context, planID uuid.UUID) (*models.RollbackPlan, error)
	GetPlansByExecution(ctx context.Context, executionID uuid.UUID) ([]*models.RollbackPlan, error)
	GetPendingPlans(ctx context.Context, userID uuid.UUID) ([]*models.RollbackPlan, error)
	UpdatePlan(ctx context.Context, plan *models.RollbackPlan) error
	DeletePlan(ctx context.Context, planID uuid.UUID) error

	// Execution operations
	CreateExecution(ctx context.Context, execution *models.RollbackExecution) error
	GetExecution(ctx context.Context, executionID uuid.UUID) (*models.RollbackExecution, error)
	GetExecutionsByPlan(ctx context.Context, planID uuid.UUID) ([]*models.RollbackExecution, error)
	GetExecutionsByCommandExecution(ctx context.Context, executionID uuid.UUID) ([]*models.RollbackExecution, error)
	UpdateExecution(ctx context.Context, execution *models.RollbackExecution) error
	DeleteExecution(ctx context.Context, executionID uuid.UUID) error

	// Cleanup and maintenance
	CleanupExpiredPlans(ctx context.Context) (int, error)
	GetOldExecutions(ctx context.Context, olderThan time.Time) ([]*models.RollbackExecution, error)

	// Health check
	HealthCheck(ctx context.Context) error
}
