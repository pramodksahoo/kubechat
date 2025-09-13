package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/cache"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/errors"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/health"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/kubernetes"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/results"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/safety"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/timeout"
)

// Service defines the command execution service interface
type Service interface {
	// Command execution
	ExecuteCommand(ctx context.Context, req *models.CommandExecutionRequest) (*models.KubernetesCommandExecution, *models.CommandApproval, error)
	GetExecution(ctx context.Context, executionID uuid.UUID) (*models.KubernetesCommandExecution, error)
	ListExecutions(ctx context.Context, userID uuid.UUID, status string, limit, offset int) ([]*models.KubernetesCommandExecution, int, error)

	// Approval workflow
	ApproveCommand(ctx context.Context, executionID, approverID uuid.UUID, decision, reason string) (*models.CommandApproval, error)
	GetPendingApprovals(ctx context.Context, userID uuid.UUID) ([]*models.CommandApproval, error)

	// Analytics and stats
	GetExecutionStats(ctx context.Context, userID uuid.UUID, from, to time.Time) (*models.ExecutionStats, error)

	// Health check
	HealthCheck(ctx context.Context) error

	// RBAC and permission validation
	ValidatePermissions(ctx context.Context, userID uuid.UUID, operation *models.KubernetesOperation) error
	ValidateSecurityContext(ctx context.Context, userID uuid.UUID, operation *models.KubernetesOperation) error

	// Multi-cluster context support
	ValidateClusterContext(ctx context.Context, userID uuid.UUID, clusterName string) error
	GetAllowedNamespaces(ctx context.Context, userID uuid.UUID, clusterName string) ([]string, error)

	// Result processing and caching
	GetProcessedResult(ctx context.Context, executionID uuid.UUID) (*results.ProcessedResult, error)
	GetCachedResult(ctx context.Context, executionID uuid.UUID) (*models.KubernetesOperationResult, error)
	InvalidateCache(ctx context.Context, resourceType, namespace, name string) error

	// Performance analysis
	GetPerformanceAnalysis(ctx context.Context, userID uuid.UUID, timeframe time.Duration) (*results.PerformanceAnalysis, error)

	// Timeout and error handling
	CancelExecution(ctx context.Context, executionID uuid.UUID, reason string) error
	GetExecutionTimeout(ctx context.Context, executionID uuid.UUID) (*timeout.TimeoutInfo, error)
	GetErrorStatistics(ctx context.Context, timeframe time.Duration) (*errors.ErrorStatistics, error)

	// Cluster health monitoring
	GetClusterHealth(ctx context.Context) (*health.ClusterHealth, error)
	IsClusterHealthy(ctx context.Context) (bool, error)
	GetHealthSummary(ctx context.Context) (*health.HealthSummary, error)
}

// Repository defines the command execution repository interface
type Repository interface {
	Create(ctx context.Context, execution *models.KubernetesCommandExecution) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.KubernetesCommandExecution, error)
	Update(ctx context.Context, execution *models.KubernetesCommandExecution) error
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*models.KubernetesCommandExecution, error)
}

// service implements the command execution service
type service struct {
	repo            Repository
	k8sService      kubernetes.Service
	safetyService   safety.Service
	cacheService    cache.Service
	resultProcessor results.Processor
	timeoutManager  timeout.Manager
	errorHandler    errors.Handler
	healthMonitor   health.Monitor
	timeout         time.Duration
}

// NewService creates a new command execution service
func NewService(repo Repository, k8sService kubernetes.Service, safetyService safety.Service, cacheService cache.Service) Service {
	return &service{
		repo:            repo,
		k8sService:      k8sService,
		safetyService:   safetyService,
		cacheService:    cacheService,
		resultProcessor: results.NewProcessor(),
		timeoutManager:  timeout.NewManager(nil), // Use default config
		errorHandler:    errors.NewHandler(),
		healthMonitor:   health.NewSimplifiedMonitor(k8sService),
		timeout:         30 * time.Second,
	}
}

// ExecuteCommand executes a Kubernetes command with comprehensive safety validation
func (s *service) ExecuteCommand(ctx context.Context, req *models.CommandExecutionRequest) (*models.KubernetesCommandExecution, *models.CommandApproval, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid request: %w", err)
	}

	// RBAC permission validation
	if err := s.ValidatePermissions(ctx, req.UserID, &req.Command); err != nil {
		return nil, nil, fmt.Errorf("permission denied: %w", err)
	}

	// Security context validation
	if err := s.ValidateSecurityContext(ctx, req.UserID, &req.Command); err != nil {
		return nil, nil, fmt.Errorf("security validation failed: %w", err)
	}

	// Determine safety level if not provided
	safetyLevel := req.SafetyLevel
	if safetyLevel == "" {
		safetyLevel = req.Command.GetSafetyLevel()
	}

	// Create execution record
	execution := &models.KubernetesCommandExecution{
		ID:          uuid.New(),
		UserID:      req.UserID,
		SessionID:   req.SessionID,
		Command:     req.Command,
		SafetyLevel: safetyLevel,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	// Store execution record
	if err := s.repo.Create(ctx, execution); err != nil {
		return nil, nil, fmt.Errorf("failed to create execution record: %w", err)
	}

	// For dangerous operations, require approval workflow
	if execution.RequiresApproval() {
		execution.Status = "pending_approval"
		if err := s.repo.Update(ctx, execution); err != nil {
			return nil, nil, fmt.Errorf("failed to update execution status: %w", err)
		}

		// Create approval object for pending execution
		approval := &models.CommandApproval{
			ID:                 uuid.New(),
			CommandExecutionID: execution.ID,
			RequestedByUserID:  execution.UserID,
			Status:             "pending",
			ExpiresAt:          time.Now().Add(24 * time.Hour), // Expires after 24 hours
			CreatedAt:          time.Now(),
		}

		return execution, approval, nil
	}

	// For safe operations, execute immediately
	if err := s.executeKubernetesCommand(ctx, execution); err != nil {
		execution.Status = "failed"
		execution.ErrorMessage = err.Error()
		execution.CompletedAt = &time.Time{}
		*execution.CompletedAt = time.Now()

		if updateErr := s.repo.Update(ctx, execution); updateErr != nil {
			return nil, nil, fmt.Errorf("failed to update execution after error: %w", updateErr)
		}

		return execution, nil, fmt.Errorf("command execution failed: %w", err)
	}

	return execution, nil, nil
}

// executeKubernetesCommand executes the actual Kubernetes command
func (s *service) executeKubernetesCommand(ctx context.Context, execution *models.KubernetesCommandExecution) error {
	// Set execution start time
	executedAt := time.Now()
	execution.ExecutedAt = &executedAt
	execution.Status = "executing"

	if err := s.repo.Update(ctx, execution); err != nil {
		return fmt.Errorf("failed to update execution status: %w", err)
	}

	// Get appropriate timeout for the operation type
	operationTimeout := s.getTimeoutForOperation(execution.Command.Operation)

	// Set timeout in timeout manager
	if err := s.timeoutManager.SetCommandTimeout(ctx, execution.ID, operationTimeout); err != nil {
		fmt.Printf("Failed to set timeout for execution %s: %v\n", execution.ID, err)
	}

	// Create timeout context using timeout manager
	execCtx, cancel := s.timeoutManager.CreateTimedContext(ctx, operationTimeout)
	defer cancel()

	// Execute the command based on operation type
	var result interface{}
	var err error

	switch execution.Command.Operation {
	case "get":
		result, err = s.executeGetCommand(execCtx, &execution.Command)
	case "list":
		result, err = s.executeListCommand(execCtx, &execution.Command)
	case "delete":
		result, err = s.executeDeleteCommand(execCtx, &execution.Command)
	case "scale":
		result, err = s.executeScaleCommand(execCtx, &execution.Command)
	case "restart":
		result, err = s.executeRestartCommand(execCtx, &execution.Command)
	case "logs":
		result, err = s.executeLogsCommand(execCtx, &execution.Command)
	default:
		err = fmt.Errorf("unsupported operation: %s", execution.Command.Operation)
	}

	// Calculate execution time
	executionTime := time.Since(executedAt)
	executionTimeMS := int(executionTime.Milliseconds())
	execution.ExecutionTimeMS = &executionTimeMS

	// Update execution with results
	completedAt := time.Now()
	execution.CompletedAt = &completedAt

	if err != nil {
		execution.Status = "failed"
		execution.ErrorMessage = err.Error()

		// Process error with error handler
		if s.errorHandler != nil {
			processedError, processErr := s.errorHandler.ProcessError(ctx, err, execution)
			if processErr == nil {
				// Track the error for analytics
				if trackErr := s.errorHandler.TrackError(ctx, processedError); trackErr != nil {
					fmt.Printf("Failed to track error: %v\n", trackErr)
				}

				// Enrich error message with recovery recommendations
				if processedError.Recovery != nil && len(processedError.Recovery.Solutions) > 0 {
					execution.ErrorMessage = fmt.Sprintf("%s\n\nSuggested solutions:\n", execution.ErrorMessage)
					for i, solution := range processedError.Recovery.Solutions {
						if i < 3 { // Limit to top 3 suggestions
							execution.ErrorMessage += fmt.Sprintf("- %s: %s\n", solution.Title, solution.Description)
						}
					}
				}
			}
		}
	} else {
		execution.Status = "completed"
		execution.Result = &models.KubernetesOperationResult{
			OperationID: execution.Command.ID,
			Success:     true,
			Result:      result,
			ExecutedAt:  executedAt,
		}

		// Store result in cache for future retrieval
		if s.cacheService != nil {
			if cacheErr := s.cacheService.StoreCommandResult(ctx, execution.ID, execution.Result); cacheErr != nil {
				// Log error but don't fail the execution
				fmt.Printf("Failed to cache command result: %v\n", cacheErr)
			}
		}
	}

	// Save final execution state
	if updateErr := s.repo.Update(ctx, execution); updateErr != nil {
		return fmt.Errorf("failed to update execution results: %w", updateErr)
	}

	return err
}

// executeGetCommand executes a get command with caching support
func (s *service) executeGetCommand(ctx context.Context, command *models.KubernetesOperation) (interface{}, error) {
	// Create cache key for this operation
	var cacheKey string
	if command.Name != "" {
		cacheKey = fmt.Sprintf("resource:%s:%s:%s", command.Resource, command.Namespace, command.Name)
	} else {
		cacheKey = fmt.Sprintf("list:%s:%s", command.Resource, command.Namespace)
	}

	// Try to get from cache first for read operations
	if s.cacheService != nil {
		var cachedResult interface{}
		if err := s.cacheService.GetResourceData(ctx, cacheKey, &cachedResult); err == nil && cachedResult != nil {
			return cachedResult, nil
		}
	}
	// Execute the actual Kubernetes operation
	var result interface{}
	var err error

	switch command.Resource {
	case "pods":
		if command.Name != "" {
			result, err = s.k8sService.GetPod(ctx, command.Namespace, command.Name)
		} else {
			// If no name specified, list instead
			result, err = s.k8sService.ListPods(ctx, command.Namespace)
		}
	case "deployments":
		if command.Name != "" {
			result, err = s.k8sService.GetDeployment(ctx, command.Namespace, command.Name)
		} else {
			result, err = s.k8sService.ListDeployments(ctx, command.Namespace)
		}
	case "services":
		if command.Name != "" {
			result, err = s.k8sService.GetService(ctx, command.Namespace, command.Name)
		} else {
			result, err = s.k8sService.ListServices(ctx, command.Namespace)
		}
	case "configmaps":
		if command.Name != "" {
			result, err = s.k8sService.GetConfigMap(ctx, command.Namespace, command.Name)
		} else {
			result, err = s.k8sService.ListConfigMaps(ctx, command.Namespace)
		}
	case "secrets":
		// Secrets only support list for security
		result, err = s.k8sService.ListSecrets(ctx, command.Namespace)
	default:
		return nil, fmt.Errorf("unsupported resource for get operation: %s", command.Resource)
	}

	// If successful, cache the result
	if err == nil && result != nil && s.cacheService != nil {
		// Use appropriate TTL based on resource type
		ttl := 5 * time.Minute // Default cache TTL
		if command.Resource == "secrets" {
			ttl = 30 * time.Second // Shorter TTL for sensitive data
		}

		if cacheErr := s.cacheService.StoreResourceData(ctx, cacheKey, result, ttl); cacheErr != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to cache resource data: %v\n", cacheErr)
		}
	}

	return result, err
}

// executeListCommand executes a list command
func (s *service) executeListCommand(ctx context.Context, command *models.KubernetesOperation) (interface{}, error) {
	switch command.Resource {
	case "pods":
		return s.k8sService.ListPods(ctx, command.Namespace)
	case "deployments":
		return s.k8sService.ListDeployments(ctx, command.Namespace)
	case "services":
		return s.k8sService.ListServices(ctx, command.Namespace)
	case "configmaps":
		return s.k8sService.ListConfigMaps(ctx, command.Namespace)
	case "secrets":
		return s.k8sService.ListSecrets(ctx, command.Namespace)
	default:
		return nil, fmt.Errorf("unsupported resource for list operation: %s", command.Resource)
	}
}

// executeDeleteCommand executes a delete command
func (s *service) executeDeleteCommand(ctx context.Context, command *models.KubernetesOperation) (interface{}, error) {
	if command.Name == "" {
		return nil, fmt.Errorf("resource name is required for delete operation")
	}

	var err error
	var result map[string]string

	switch command.Resource {
	case "pods":
		err = s.k8sService.DeletePod(ctx, command.Namespace, command.Name)
		if err != nil {
			return nil, err
		}
		result = map[string]string{"message": fmt.Sprintf("Pod %s deleted successfully", command.Name)}
	default:
		return nil, fmt.Errorf("delete operation not supported for resource: %s", command.Resource)
	}

	// Invalidate related cache entries after successful delete
	if err == nil && s.cacheService != nil {
		if cacheErr := s.cacheService.InvalidateResourceCache(ctx, command.Resource, command.Namespace, command.Name); cacheErr != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to invalidate cache after delete: %v\n", cacheErr)
		}
	}

	return result, nil
}

// executeScaleCommand executes a scale command
func (s *service) executeScaleCommand(ctx context.Context, command *models.KubernetesOperation) (interface{}, error) {
	if command.Name == "" {
		return nil, fmt.Errorf("resource name is required for scale operation")
	}

	// For now, assume scaling to 1 replica if not specified
	// In a real implementation, this would come from command context
	replicas := int32(1)

	var err error
	var result map[string]interface{}

	switch command.Resource {
	case "deployments":
		err = s.k8sService.ScaleDeployment(ctx, command.Namespace, command.Name, replicas)
		if err != nil {
			return nil, err
		}
		result = map[string]interface{}{
			"message":  fmt.Sprintf("Deployment %s scaled successfully", command.Name),
			"replicas": replicas,
		}
	default:
		return nil, fmt.Errorf("scale operation not supported for resource: %s", command.Resource)
	}

	// Invalidate related cache entries after successful scale
	if err == nil && s.cacheService != nil {
		if cacheErr := s.cacheService.InvalidateResourceCache(ctx, command.Resource, command.Namespace, command.Name); cacheErr != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to invalidate cache after scale: %v\n", cacheErr)
		}
	}

	return result, err
}

// executeRestartCommand executes a restart command
func (s *service) executeRestartCommand(ctx context.Context, command *models.KubernetesOperation) (interface{}, error) {
	if command.Name == "" {
		return nil, fmt.Errorf("resource name is required for restart operation")
	}

	switch command.Resource {
	case "deployments":
		err := s.k8sService.RestartDeployment(ctx, command.Namespace, command.Name)
		if err != nil {
			return nil, err
		}
		return map[string]string{"message": fmt.Sprintf("Deployment %s restarted successfully", command.Name)}, nil
	default:
		return nil, fmt.Errorf("restart operation not supported for resource: %s", command.Resource)
	}
}

// executeLogsCommand executes a logs command
func (s *service) executeLogsCommand(ctx context.Context, command *models.KubernetesOperation) (interface{}, error) {
	if command.Name == "" {
		return nil, fmt.Errorf("resource name is required for logs operation")
	}

	switch command.Resource {
	case "pods":
		// Default log options
		logOptions := &models.LogOptions{
			TailLines:  100,
			Timestamps: true,
		}

		logs, err := s.k8sService.GetPodLogs(ctx, command.Namespace, command.Name, logOptions)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"logs":  logs,
			"pod":   command.Name,
			"lines": len(logs),
		}, nil
	default:
		return nil, fmt.Errorf("logs operation not supported for resource: %s", command.Resource)
	}
}

// GetExecution retrieves a command execution by ID
func (s *service) GetExecution(ctx context.Context, executionID uuid.UUID) (*models.KubernetesCommandExecution, error) {
	return s.repo.GetByID(ctx, executionID)
}

// ValidatePermissions validates RBAC permissions for the operation
func (s *service) ValidatePermissions(ctx context.Context, userID uuid.UUID, operation *models.KubernetesOperation) error {
	// Basic validation using the existing Kubernetes service
	if err := s.k8sService.ValidateOperation(ctx, operation); err != nil {
		return fmt.Errorf("operation validation failed: %w", err)
	}

	// TODO: Implement proper RBAC validation with user roles
	// For now, rely on the k8s service namespace validation
	return nil
}

// ValidateSecurityContext validates the security context for the operation
func (s *service) ValidateSecurityContext(ctx context.Context, userID uuid.UUID, operation *models.KubernetesOperation) error {
	// Validate against safety service if available
	if s.safetyService != nil {
		// Create a safety validation request
		// This would integrate with the existing safety classification from Story 1.5
		safetyResult := operation.GetSafetyLevel()

		// Block dangerous operations without explicit approval
		if safetyResult == "dangerous" {
			return fmt.Errorf("dangerous operation requires approval workflow")
		}
	}

	return nil
}

// ValidateClusterContext validates access to a specific cluster context
func (s *service) ValidateClusterContext(ctx context.Context, userID uuid.UUID, clusterName string) error {
	// TODO: Implement cluster context validation
	// For now, assume single cluster access
	return nil
}

// GetAllowedNamespaces returns namespaces the user can access in a cluster
func (s *service) GetAllowedNamespaces(ctx context.Context, userID uuid.UUID, clusterName string) ([]string, error) {
	// Get namespaces from Kubernetes service
	namespaces, err := s.k8sService.ListNamespaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	// Extract namespace names
	var allowedNamespaces []string
	for _, ns := range namespaces {
		allowedNamespaces = append(allowedNamespaces, ns.Name)
	}

	return allowedNamespaces, nil
}

// GetProcessedResult retrieves and processes a command execution result
func (s *service) GetProcessedResult(ctx context.Context, executionID uuid.UUID) (*results.ProcessedResult, error) {
	// Get the execution record
	execution, err := s.repo.GetByID(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	if execution.Result == nil {
		return nil, fmt.Errorf("execution has no result")
	}

	// Process the result using the result processor
	processed, err := s.resultProcessor.ProcessCommandResult(execution.Result, execution)
	if err != nil {
		return nil, fmt.Errorf("failed to process result: %w", err)
	}

	return processed, nil
}

// GetCachedResult retrieves cached command result
func (s *service) GetCachedResult(ctx context.Context, executionID uuid.UUID) (*models.KubernetesOperationResult, error) {
	if s.cacheService == nil {
		return nil, fmt.Errorf("cache service not available")
	}

	// Try to get result from cache
	result, err := s.cacheService.GetCommandResult(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached result: %w", err)
	}

	return result, nil
}

// InvalidateCache invalidates cache entries for specific resources
func (s *service) InvalidateCache(ctx context.Context, resourceType, namespace, name string) error {
	if s.cacheService == nil {
		return nil // No-op if cache service not available
	}

	if err := s.cacheService.InvalidateResourceCache(ctx, resourceType, namespace, name); err != nil {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}

	return nil
}

// GetPerformanceAnalysis analyzes command execution performance for a user
func (s *service) GetPerformanceAnalysis(ctx context.Context, userID uuid.UUID, timeframe time.Duration) (*results.PerformanceAnalysis, error) {
	// Get recent executions for the user
	executions, err := s.repo.ListByUser(ctx, userID, 1000) // Get up to 1000 recent executions
	if err != nil {
		return nil, fmt.Errorf("failed to get executions for analysis: %w", err)
	}

	// Filter by timeframe
	cutoffTime := time.Now().Add(-timeframe)
	var filteredExecutions []*models.KubernetesCommandExecution
	for _, exec := range executions {
		if exec.CreatedAt.After(cutoffTime) {
			filteredExecutions = append(filteredExecutions, exec)
		}
	}

	// Analyze performance using the result processor
	analysis, err := s.resultProcessor.AnalyzeCommandPerformance(filteredExecutions)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze performance: %w", err)
	}

	return analysis, nil
}

// CancelExecution cancels a running command execution
func (s *service) CancelExecution(ctx context.Context, executionID uuid.UUID, reason string) error {
	// Cancel the timeout
	if err := s.timeoutManager.CancelCommand(ctx, executionID, reason); err != nil {
		// Continue even if timeout cancellation fails
		fmt.Printf("Failed to cancel timeout for execution %s: %v\n", executionID, err)
	}

	// Update execution status in database
	execution, err := s.repo.GetByID(ctx, executionID)
	if err != nil {
		return fmt.Errorf("failed to get execution for cancellation: %w", err)
	}

	if execution.Status != "executing" && execution.Status != "pending_approval" {
		return fmt.Errorf("execution %s cannot be cancelled (status: %s)", executionID, execution.Status)
	}

	execution.Status = "cancelled"
	execution.ErrorMessage = fmt.Sprintf("Cancelled: %s", reason)
	completedAt := time.Now()
	execution.CompletedAt = &completedAt

	if err := s.repo.Update(ctx, execution); err != nil {
		return fmt.Errorf("failed to update cancelled execution: %w", err)
	}

	return nil
}

// GetExecutionTimeout retrieves timeout information for an execution
func (s *service) GetExecutionTimeout(ctx context.Context, executionID uuid.UUID) (*timeout.TimeoutInfo, error) {
	// Get timeout info from timeout manager
	timeouts := s.timeoutManager.GetActiveTimeouts()

	for _, timeoutInfo := range timeouts {
		if timeoutInfo.ExecutionID == executionID {
			return timeoutInfo, nil
		}
	}

	return nil, fmt.Errorf("timeout information not found for execution: %s", executionID)
}

// GetErrorStatistics retrieves error statistics from the error handler
func (s *service) GetErrorStatistics(ctx context.Context, timeframe time.Duration) (*errors.ErrorStatistics, error) {
	return s.errorHandler.GetErrorStatistics(ctx, timeframe)
}

// GetClusterHealth returns current cluster health status
func (s *service) GetClusterHealth(ctx context.Context) (*health.ClusterHealth, error) {
	return s.healthMonitor.GetClusterHealth(ctx)
}

// IsClusterHealthy returns whether the cluster is healthy
func (s *service) IsClusterHealthy(ctx context.Context) (bool, error) {
	return s.healthMonitor.IsClusterHealthy(ctx)
}

// GetHealthSummary returns a health summary
func (s *service) GetHealthSummary(ctx context.Context) (*health.HealthSummary, error) {
	return s.healthMonitor.GetHealthSummary(ctx)
}

// getTimeoutForOperation returns appropriate timeout duration for operation type
func (s *service) getTimeoutForOperation(operation string) time.Duration {
	// Default timeout values for different operations
	switch operation {
	case "get", "describe":
		return 30 * time.Second
	case "list":
		return 60 * time.Second
	case "delete":
		return 2 * time.Minute
	case "scale":
		return 3 * time.Minute
	case "restart":
		return 5 * time.Minute
	case "logs":
		return 2 * time.Minute
	default:
		return s.timeout // Use service default
	}
}

// ApproveCommand approves a pending command execution
func (s *service) ApproveCommand(ctx context.Context, executionID, approverID uuid.UUID, decision, reason string) (*models.CommandApproval, error) {
	// Get the command execution
	execution, err := s.repo.GetByID(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("execution not found: %w", err)
	}

	if execution.Status != "pending_approval" {
		return nil, fmt.Errorf("execution is not pending approval (status: %s)", execution.Status)
	}

	// Process the approval
	approval := &models.CommandApproval{
		ID:                 uuid.New(),
		CommandExecutionID: executionID,
		RequestedByUserID:  execution.UserID,
		ApprovedByUserID:   &approverID,
		Status:             decision, // "approved" or "rejected"
		Reason:             reason,
		CreatedAt:          time.Now(),
		DecidedAt:          &[]time.Time{time.Now()}[0],
	}

	// Update execution status based on decision
	if decision == "approve" {
		execution.Status = "approved"
		execution.ApprovalID = &approval.ID

		// Execute the command if approved
		// This would trigger the actual execution in the background
		go func() {
			s.executeApprovedCommand(context.Background(), execution)
		}()
	} else {
		execution.Status = "rejected"
		execution.ErrorMessage = fmt.Sprintf("Command rejected: %s", reason)
		completedAt := time.Now()
		execution.CompletedAt = &completedAt
	}

	// Update execution in repository
	if err := s.repo.Update(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to update execution: %w", err)
	}

	// Note: In a real implementation, you'd save the approval to a repository
	// For now, we return the approval object directly
	return approval, nil
}

// executeApprovedCommand executes a command that has been approved
func (s *service) executeApprovedCommand(ctx context.Context, execution *models.KubernetesCommandExecution) {
	// Execute the command
	result, err := s.k8sService.ExecuteOperation(ctx, &execution.Command)
	if err != nil {
		execution.Status = "failed"
		execution.ErrorMessage = err.Error()
	} else {
		execution.Status = "completed"
		execution.Result = result
	}

	// Update completion time
	completedAt := time.Now()
	execution.CompletedAt = &completedAt

	// Calculate execution time
	if execution.ExecutedAt != nil {
		execTime := int(completedAt.Sub(*execution.ExecutedAt).Milliseconds())
		execution.ExecutionTimeMS = &execTime
	}

	// Save updated execution
	s.repo.Update(ctx, execution)
}

// GetPendingApprovals returns pending approvals for a user
func (s *service) GetPendingApprovals(ctx context.Context, userID uuid.UUID) ([]*models.CommandApproval, error) {
	// Get pending executions that require approval
	executions, err := s.repo.ListByUser(ctx, userID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get executions: %w", err)
	}

	var approvals []*models.CommandApproval
	for _, exec := range executions {
		if exec.Status == "pending_approval" {
			// Create a mock approval object for pending executions
			approval := &models.CommandApproval{
				ID:                 uuid.New(),
				CommandExecutionID: exec.ID,
				RequestedByUserID:  exec.UserID,
				Status:             "pending",
				ExpiresAt:          exec.CreatedAt.Add(24 * time.Hour), // Expires after 24 hours
				CreatedAt:          exec.CreatedAt,
			}
			approvals = append(approvals, approval)
		}
	}

	return approvals, nil
}

// GetExecutionStats returns execution statistics for a user
func (s *service) GetExecutionStats(ctx context.Context, userID uuid.UUID, from, to time.Time) (*models.ExecutionStats, error) {
	// Get executions for the user
	executions, err := s.repo.ListByUser(ctx, userID, 1000) // Get up to 1000 executions
	if err != nil {
		return nil, fmt.Errorf("failed to get executions: %w", err)
	}

	// Filter by date range
	var filteredExecs []*models.KubernetesCommandExecution
	for _, exec := range executions {
		if exec.CreatedAt.After(from) && exec.CreatedAt.Before(to) {
			filteredExecs = append(filteredExecs, exec)
		}
	}

	// Calculate statistics
	stats := &models.ExecutionStats{
		UserID: userID,
		From:   from,
		To:     to,
	}

	stats.TotalExecutions = len(filteredExecs)

	var totalTimeMS int64
	resourceCount := make(map[string]int)

	for _, exec := range filteredExecs {
		// Count by status
		switch exec.Status {
		case "completed":
			stats.SuccessfulOnes++
		case "failed", "timeout", "cancelled":
			stats.FailedOnes++
		}

		// Sum execution times
		if exec.ExecutionTimeMS != nil {
			totalTimeMS += int64(*exec.ExecutionTimeMS)
		}

		// Count resources
		resourceCount[exec.Command.Resource]++
	}

	// Calculate average time
	if stats.SuccessfulOnes > 0 {
		stats.AverageTime = float64(totalTimeMS) / float64(stats.SuccessfulOnes)
	}

	// Find most used resource
	maxCount := 0
	for resource, count := range resourceCount {
		if count > maxCount {
			maxCount = count
			stats.MostUsedResource = resource
		}
	}

	return stats, nil
}

// HealthCheck performs health check on the command service
func (s *service) HealthCheck(ctx context.Context) error {
	// Check Kubernetes service
	if err := s.k8sService.HealthCheck(ctx); err != nil {
		return fmt.Errorf("kubernetes service health check failed: %w", err)
	}

	// Check other services if available
	if s.cacheService != nil {
		if err := s.cacheService.HealthCheck(ctx); err != nil {
			return fmt.Errorf("cache service health check failed: %w", err)
		}
	}

	if s.timeoutManager != nil {
		if err := s.timeoutManager.HealthCheck(ctx); err != nil {
			return fmt.Errorf("timeout manager health check failed: %w", err)
		}
	}

	// Error handler and health monitor health checks would be implemented here
	// if s.errorHandler != nil {
	// 	if err := s.errorHandler.HealthCheck(ctx); err != nil {
	// 		return fmt.Errorf("error handler health check failed: %w", err)
	// 	}
	// }

	// if s.healthMonitor != nil {
	// 	if err := s.healthMonitor.HealthCheck(ctx); err != nil {
	// 		return fmt.Errorf("health monitor health check failed: %w", err)
	// 	}
	// }

	return nil
}

// ListExecutions retrieves command executions with filtering and pagination
func (s *service) ListExecutions(ctx context.Context, userID uuid.UUID, status string, limit, offset int) ([]*models.KubernetesCommandExecution, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 50 // Default limit
	}

	// Get all executions for the user first (in a real implementation, this would be paginated at the DB level)
	allExecutions, err := s.repo.ListByUser(ctx, userID, 1000)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get executions: %w", err)
	}

	// Filter by status if specified
	var filteredExecutions []*models.KubernetesCommandExecution
	for _, exec := range allExecutions {
		if status == "" || exec.Status == status {
			filteredExecutions = append(filteredExecutions, exec)
		}
	}

	total := len(filteredExecutions)

	// Apply pagination
	start := offset
	end := offset + limit
	if start > total {
		return []*models.KubernetesCommandExecution{}, total, nil
	}
	if end > total {
		end = total
	}

	pagedExecutions := filteredExecutions[start:end]
	return pagedExecutions, total, nil
}
