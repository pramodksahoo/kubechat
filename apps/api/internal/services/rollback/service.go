package rollback

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/repositories"
	kubernetesService "github.com/pramodksahoo/kubechat/apps/api/internal/services/kubernetes"
)

// Service defines the rollback service interface
type Service interface {
	// Rollback operations
	CreateRollbackPlan(ctx context.Context, executionID uuid.UUID) (*models.RollbackPlan, error)
	ExecuteRollback(ctx context.Context, planID uuid.UUID, userID uuid.UUID) (*models.RollbackExecution, error)
	ValidateRollback(ctx context.Context, executionID uuid.UUID) (*models.RollbackValidation, error)

	// Rollback management
	GetRollbackPlan(ctx context.Context, planID uuid.UUID) (*models.RollbackPlan, error)
	GetRollbackHistory(ctx context.Context, executionID uuid.UUID) ([]*models.RollbackExecution, error)
	CancelRollback(ctx context.Context, rollbackID uuid.UUID, reason string) error

	// Status and monitoring
	GetRollbackStatus(ctx context.Context, rollbackID uuid.UUID) (*models.RollbackStatus, error)
	ListPendingRollbacks(ctx context.Context, userID uuid.UUID) ([]*models.RollbackPlan, error)

	// Health check
	HealthCheck(ctx context.Context) error
}

// service implements the rollback Service interface
type service struct {
	commandRepo       repositories.CommandRepository
	rollbackRepo      repositories.RollbackRepository
	kubernetesService kubernetesService.Service
}

// NewService creates a new rollback service
func NewService(
	commandRepo repositories.CommandRepository,
	rollbackRepo repositories.RollbackRepository,
	kubernetesService kubernetesService.Service,
) Service {
	return &service{
		commandRepo:       commandRepo,
		rollbackRepo:      rollbackRepo,
		kubernetesService: kubernetesService,
	}
}

// CreateRollbackPlan analyzes a command execution and creates a rollback plan
func (s *service) CreateRollbackPlan(ctx context.Context, executionID uuid.UUID) (*models.RollbackPlan, error) {
	// Get the original command execution
	execution, err := s.commandRepo.GetExecution(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get command execution: %w", err)
	}

	if execution.Status != "completed" {
		return nil, fmt.Errorf("cannot create rollback plan for incomplete execution")
	}

	// Check if operation is rollbackable
	if !s.isRollbackable(execution.Command.Operation) {
		return nil, fmt.Errorf("operation '%s' is not rollbackable", execution.Command.Operation)
	}

	// Generate rollback steps
	rollbackSteps, err := s.generateRollbackSteps(ctx, execution)
	if err != nil {
		return nil, fmt.Errorf("failed to generate rollback steps: %w", err)
	}

	// Create rollback plan
	plan := &models.RollbackPlan{
		ID:                 uuid.New(),
		CommandExecutionID: executionID,
		UserID:             execution.UserID,
		SessionID:          execution.SessionID,
		OriginalCommand:    execution.Command,
		RollbackSteps:      rollbackSteps,
		Status:             "planned",
		Reason:             "User requested rollback",
		EstimatedDuration:  s.estimateRollbackDuration(rollbackSteps),
		CreatedAt:          time.Now(),
		ExpiresAt:          time.Now().Add(24 * time.Hour), // Plans expire after 24 hours
	}

	// Validate the rollback plan
	validation, err := s.validateRollbackPlan(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("rollback plan validation failed: %w", err)
	}

	plan.Validation = validation
	if !validation.IsValid {
		plan.Status = "invalid"
	}

	// Save rollback plan
	if err := s.rollbackRepo.CreatePlan(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to save rollback plan: %w", err)
	}

	return plan, nil
}

// ExecuteRollback executes a rollback plan
func (s *service) ExecuteRollback(ctx context.Context, planID uuid.UUID, userID uuid.UUID) (*models.RollbackExecution, error) {
	// Get rollback plan
	plan, err := s.rollbackRepo.GetPlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rollback plan: %w", err)
	}

	// Validate plan status and ownership
	if plan.Status != "planned" {
		return nil, fmt.Errorf("rollback plan status is '%s', expected 'planned'", plan.Status)
	}
	if plan.UserID != userID {
		return nil, fmt.Errorf("unauthorized: user does not own this rollback plan")
	}
	if time.Now().After(plan.ExpiresAt) {
		return nil, fmt.Errorf("rollback plan has expired")
	}

	// Create rollback execution record
	execution := &models.RollbackExecution{
		ID:           uuid.New(),
		PlanID:       planID,
		UserID:       userID,
		Status:       "executing",
		StartedAt:    time.Now(),
		ExecutionLog: []models.RollbackStepResult{},
	}

	// Save initial execution record
	if err := s.rollbackRepo.CreateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create rollback execution: %w", err)
	}

	// Update plan status
	plan.Status = "executing"
	if err := s.rollbackRepo.UpdatePlan(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to update plan status: %w", err)
	}

	// Execute rollback steps
	go s.executeRollbackSteps(context.Background(), execution, plan)

	return execution, nil
}

// ValidateRollback validates if a command execution can be rolled back
func (s *service) ValidateRollback(ctx context.Context, executionID uuid.UUID) (*models.RollbackValidation, error) {
	execution, err := s.commandRepo.GetExecution(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get command execution: %w", err)
	}

	validation := &models.RollbackValidation{
		ExecutionID: executionID,
		IsValid:     true,
		Reasons:     []string{},
		Warnings:    []string{},
		CheckedAt:   time.Now(),
	}

	// Check if operation is rollbackable
	if !s.isRollbackable(execution.Command.Operation) {
		validation.IsValid = false
		validation.Reasons = append(validation.Reasons, fmt.Sprintf("Operation '%s' is not rollbackable", execution.Command.Operation))
	}

	// Check execution status
	if execution.Status != "completed" {
		validation.IsValid = false
		validation.Reasons = append(validation.Reasons, "Command execution is not completed")
	}

	// Check if resources still exist for rollback
	if execution.Command.Operation == "delete" && validation.IsValid {
		exists, err := s.checkResourceExists(ctx, execution)
		if err != nil {
			validation.Warnings = append(validation.Warnings, fmt.Sprintf("Could not verify resource existence: %v", err))
		} else if exists {
			validation.Warnings = append(validation.Warnings, "Resource has been recreated since deletion")
		}
	}

	// Check for existing rollback plans
	existingPlans, err := s.rollbackRepo.GetPlansByExecution(ctx, executionID)
	if err != nil {
		validation.Warnings = append(validation.Warnings, fmt.Sprintf("Could not check existing rollback plans: %v", err))
	} else if len(existingPlans) > 0 {
		validation.Warnings = append(validation.Warnings, "Rollback plans already exist for this execution")
	}

	return validation, nil
}

// GetRollbackPlan retrieves a rollback plan by ID
func (s *service) GetRollbackPlan(ctx context.Context, planID uuid.UUID) (*models.RollbackPlan, error) {
	return s.rollbackRepo.GetPlan(ctx, planID)
}

// GetRollbackHistory retrieves rollback history for a command execution
func (s *service) GetRollbackHistory(ctx context.Context, executionID uuid.UUID) ([]*models.RollbackExecution, error) {
	return s.rollbackRepo.GetExecutionsByCommandExecution(ctx, executionID)
}

// CancelRollback cancels an ongoing rollback operation
func (s *service) CancelRollback(ctx context.Context, rollbackID uuid.UUID, reason string) error {
	execution, err := s.rollbackRepo.GetExecution(ctx, rollbackID)
	if err != nil {
		return fmt.Errorf("failed to get rollback execution: %w", err)
	}

	if execution.Status != "executing" {
		return fmt.Errorf("rollback is not currently executing (status: %s)", execution.Status)
	}

	// Update execution status
	execution.Status = "cancelled"
	execution.CompletedAt = &time.Time{}
	*execution.CompletedAt = time.Now()
	execution.Error = &reason

	return s.rollbackRepo.UpdateExecution(ctx, execution)
}

// GetRollbackStatus gets the current status of a rollback operation
func (s *service) GetRollbackStatus(ctx context.Context, rollbackID uuid.UUID) (*models.RollbackStatus, error) {
	execution, err := s.rollbackRepo.GetExecution(ctx, rollbackID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rollback execution: %w", err)
	}

	status := &models.RollbackStatus{
		RollbackID:  rollbackID,
		Status:      execution.Status,
		StartedAt:   execution.StartedAt,
		CompletedAt: execution.CompletedAt,
		Progress:    s.calculateProgress(execution),
		CurrentStep: s.getCurrentStep(execution),
		TotalSteps:  len(execution.ExecutionLog),
		Error:       execution.Error,
	}

	return status, nil
}

// ListPendingRollbacks lists rollback plans that are pending execution
func (s *service) ListPendingRollbacks(ctx context.Context, userID uuid.UUID) ([]*models.RollbackPlan, error) {
	return s.rollbackRepo.GetPendingPlans(ctx, userID)
}

// HealthCheck performs health check on the rollback service
func (s *service) HealthCheck(ctx context.Context) error {
	// Check repository connectivity
	if err := s.rollbackRepo.HealthCheck(ctx); err != nil {
		return fmt.Errorf("rollback repository health check failed: %w", err)
	}

	// Check kubernetes service connectivity
	if err := s.kubernetesService.HealthCheck(ctx); err != nil {
		return fmt.Errorf("kubernetes service health check failed: %w", err)
	}

	return nil
}

// isRollbackable determines if an operation can be rolled back
func (s *service) isRollbackable(operation string) bool {
	rollbackableOps := map[string]bool{
		"create":   true,
		"apply":    true,
		"delete":   true,
		"scale":    true,
		"patch":    true,
		"label":    true,
		"annotate": true,
	}

	return rollbackableOps[operation]
}

// generateRollbackSteps creates the sequence of steps needed to rollback an operation
func (s *service) generateRollbackSteps(ctx context.Context, execution *models.KubernetesCommandExecution) ([]models.RollbackStep, error) {
	var steps []models.RollbackStep

	switch execution.Command.Operation {
	case "create", "apply":
		// For create/apply operations, rollback is delete
		steps = append(steps, models.RollbackStep{
			StepNumber:  1,
			Operation:   "delete",
			Resource:    execution.Command.Resource,
			Name:        execution.Command.Name,
			Namespace:   execution.Command.Namespace,
			Description: fmt.Sprintf("Delete %s/%s that was created", execution.Command.Resource, execution.Command.Name),
		})

	case "delete":
		// For delete operations, need to recreate from backup or original spec
		if execution.Result != nil && execution.Result.BackupData != nil {
			steps = append(steps, models.RollbackStep{
				StepNumber:  1,
				Operation:   "create",
				Resource:    execution.Command.Resource,
				Name:        execution.Command.Name,
				Namespace:   execution.Command.Namespace,
				Description: fmt.Sprintf("Recreate %s/%s from backup", execution.Command.Resource, execution.Command.Name),
				Data:        execution.Result.BackupData,
			})
		}

	case "scale":
		// For scale operations, restore previous replica count
		if execution.Result != nil && execution.Result.PreviousState != nil {
			if prevReplicas, exists := execution.Result.PreviousState["replicas"]; exists {
				steps = append(steps, models.RollbackStep{
					StepNumber:  1,
					Operation:   "scale",
					Resource:    execution.Command.Resource,
					Name:        execution.Command.Name,
					Namespace:   execution.Command.Namespace,
					Description: fmt.Sprintf("Restore %s/%s to previous replica count", execution.Command.Resource, execution.Command.Name),
					Data:        map[string]interface{}{"replicas": prevReplicas},
				})
			}
		}

	case "patch":
		// For patch operations, apply reverse patch
		if execution.Result != nil && execution.Result.PreviousState != nil {
			steps = append(steps, models.RollbackStep{
				StepNumber:  1,
				Operation:   "patch",
				Resource:    execution.Command.Resource,
				Name:        execution.Command.Name,
				Namespace:   execution.Command.Namespace,
				Description: fmt.Sprintf("Apply reverse patch to %s/%s", execution.Command.Resource, execution.Command.Name),
				Data:        execution.Result.PreviousState,
			})
		}
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("no rollback steps could be generated for operation: %s", execution.Command.Operation)
	}

	return steps, nil
}

// validateRollbackPlan validates that a rollback plan can be executed
func (s *service) validateRollbackPlan(ctx context.Context, plan *models.RollbackPlan) (*models.RollbackValidation, error) {
	validation := &models.RollbackValidation{
		ExecutionID: plan.CommandExecutionID,
		IsValid:     true,
		Reasons:     []string{},
		Warnings:    []string{},
		CheckedAt:   time.Now(),
	}

	// Validate each rollback step
	for _, step := range plan.RollbackSteps {
		if err := s.validateRollbackStep(ctx, &step); err != nil {
			validation.IsValid = false
			validation.Reasons = append(validation.Reasons, fmt.Sprintf("Step %d validation failed: %v", step.StepNumber, err))
		}
	}

	return validation, nil
}

// validateRollbackStep validates an individual rollback step
func (s *service) validateRollbackStep(ctx context.Context, step *models.RollbackStep) error {
	// Basic validation
	if step.Resource == "" {
		return fmt.Errorf("resource type is required")
	}
	if step.Name == "" {
		return fmt.Errorf("resource name is required")
	}
	if step.Operation == "" {
		return fmt.Errorf("operation is required")
	}

	// Operation-specific validation
	switch step.Operation {
	case "delete":
		// Check if resource exists before attempting delete
		exists, err := s.kubernetesService.ResourceExists(ctx, step.Resource, step.Name, step.Namespace)
		if err != nil {
			return fmt.Errorf("failed to check resource existence: %w", err)
		}
		if !exists {
			return fmt.Errorf("resource %s/%s does not exist", step.Resource, step.Name)
		}

	case "create":
		// Check if resource data is available for creation
		if step.Data == nil {
			return fmt.Errorf("resource data is required for create operation")
		}
		// Check if resource already exists
		exists, err := s.kubernetesService.ResourceExists(ctx, step.Resource, step.Name, step.Namespace)
		if err != nil {
			return fmt.Errorf("failed to check resource existence: %w", err)
		}
		if exists {
			return fmt.Errorf("resource %s/%s already exists", step.Resource, step.Name)
		}
	}

	return nil
}

// estimateRollbackDuration estimates how long the rollback will take
func (s *service) estimateRollbackDuration(steps []models.RollbackStep) time.Duration {
	baseDuration := 30 * time.Second // Base time per step
	return time.Duration(len(steps)) * baseDuration
}

// executeRollbackSteps executes rollback steps in sequence
func (s *service) executeRollbackSteps(ctx context.Context, execution *models.RollbackExecution, plan *models.RollbackPlan) {
	for _, step := range plan.RollbackSteps {
		stepResult := s.executeRollbackStep(ctx, &step)
		execution.ExecutionLog = append(execution.ExecutionLog, stepResult)

		// Update execution with current progress
		s.rollbackRepo.UpdateExecution(ctx, execution)

		// If step failed, stop execution
		if stepResult.Status == "failed" {
			execution.Status = "failed"
			execution.Error = &stepResult.Error
			break
		}
	}

	// Mark execution as completed if all steps succeeded
	if execution.Status == "executing" {
		execution.Status = "completed"
	}

	// Set completion time
	now := time.Now()
	execution.CompletedAt = &now

	// Update final execution state
	s.rollbackRepo.UpdateExecution(ctx, execution)

	// Update plan status
	plan.Status = execution.Status
	s.rollbackRepo.UpdatePlan(ctx, plan)
}

// executeRollbackStep executes a single rollback step
func (s *service) executeRollbackStep(ctx context.Context, step *models.RollbackStep) models.RollbackStepResult {
	result := models.RollbackStepResult{
		StepNumber: step.StepNumber,
		Operation:  step.Operation,
		Resource:   step.Resource,
		Name:       step.Name,
		Namespace:  step.Namespace,
		StartedAt:  time.Now(),
		Status:     "executing",
	}

	// Execute the operation
	var err error
	switch step.Operation {
	case "delete":
		err = s.kubernetesService.DeleteResource(ctx, step.Resource, step.Name, step.Namespace)
	case "create":
		err = s.kubernetesService.CreateResource(ctx, step.Resource, step.Data)
	case "patch":
		err = s.kubernetesService.PatchResource(ctx, step.Resource, step.Name, step.Namespace, step.Data)
	case "scale":
		if step.Data != nil {
			if replicasVal, exists := step.Data["replicas"]; exists {
				if replicas, ok := replicasVal.(int); ok {
					err = s.kubernetesService.ScaleResource(ctx, step.Resource, step.Name, step.Namespace, replicas)
				} else if replicasFloat, ok := replicasVal.(float64); ok {
					err = s.kubernetesService.ScaleResource(ctx, step.Resource, step.Name, step.Namespace, int(replicasFloat))
				} else {
					err = fmt.Errorf("invalid replicas data type for scale operation")
				}
			} else {
				err = fmt.Errorf("missing replicas data for scale operation")
			}
		} else {
			err = fmt.Errorf("no data provided for scale operation")
		}
	default:
		err = fmt.Errorf("unsupported rollback operation: %s", step.Operation)
	}

	// Update result based on execution outcome
	now := time.Now()
	result.CompletedAt = &now

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
	} else {
		result.Status = "completed"
	}

	return result
}

// checkResourceExists checks if a resource still exists (for delete rollbacks)
func (s *service) checkResourceExists(ctx context.Context, execution *models.KubernetesCommandExecution) (bool, error) {
	return s.kubernetesService.ResourceExists(ctx, execution.Command.Resource, execution.Command.Name, execution.Command.Namespace)
}

// calculateProgress calculates the completion percentage of a rollback
func (s *service) calculateProgress(execution *models.RollbackExecution) float64 {
	if len(execution.ExecutionLog) == 0 {
		return 0.0
	}

	completedSteps := 0
	for _, step := range execution.ExecutionLog {
		if step.Status == "completed" || step.Status == "failed" {
			completedSteps++
		}
	}

	return float64(completedSteps) / float64(len(execution.ExecutionLog)) * 100.0
}

// getCurrentStep gets the currently executing step
func (s *service) getCurrentStep(execution *models.RollbackExecution) string {
	for _, step := range execution.ExecutionLog {
		if step.Status == "executing" {
			return fmt.Sprintf("Step %d: %s %s/%s", step.StepNumber, step.Operation, step.Resource, step.Name)
		}
	}

	if len(execution.ExecutionLog) > 0 {
		lastStep := execution.ExecutionLog[len(execution.ExecutionLog)-1]
		return fmt.Sprintf("Step %d: %s %s/%s", lastStep.StepNumber, lastStep.Operation, lastStep.Resource, lastStep.Name)
	}

	return "No steps executed"
}
