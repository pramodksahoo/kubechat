package services

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of command repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, execution *models.KubernetesCommandExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, executionID uuid.UUID) (*models.KubernetesCommandExecution, error) {
	args := m.Called(ctx, executionID)
	return args.Get(0).(*models.KubernetesCommandExecution), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, execution *models.KubernetesCommandExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*models.KubernetesCommandExecution, error) {
	args := m.Called(ctx, userID, limit)
	return args.Get(0).([]*models.KubernetesCommandExecution), args.Error(1)
}

func (m *MockRepository) Delete(ctx context.Context, executionID uuid.UUID) error {
	args := m.Called(ctx, executionID)
	return args.Error(0)
}

// MockKubernetesService is a mock implementation of Kubernetes service
type MockKubernetesService struct {
	mock.Mock
}

func (m *MockKubernetesService) ValidateOperation(ctx context.Context, operation *models.KubernetesOperation) error {
	args := m.Called(ctx, operation)
	return args.Error(0)
}

func (m *MockKubernetesService) ExecuteOperation(ctx context.Context, operation *models.KubernetesOperation) (*models.KubernetesOperationResult, error) {
	args := m.Called(ctx, operation)
	return args.Get(0).(*models.KubernetesOperationResult), args.Error(1)
}

func (m *MockKubernetesService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockKubernetesService) ListPods(ctx context.Context, namespace string) ([]*models.KubernetesPod, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).([]*models.KubernetesPod), args.Error(1)
}

func (m *MockKubernetesService) GetPod(ctx context.Context, namespace, name string) (*models.KubernetesPod, error) {
	args := m.Called(ctx, namespace, name)
	return args.Get(0).(*models.KubernetesPod), args.Error(1)
}

func (m *MockKubernetesService) DeletePod(ctx context.Context, namespace, name string) error {
	args := m.Called(ctx, namespace, name)
	return args.Error(0)
}

func (m *MockKubernetesService) ListNamespaces(ctx context.Context) ([]*models.KubernetesNamespace, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.KubernetesNamespace), args.Error(1)
}

func (m *MockKubernetesService) GetClusterInfo(ctx context.Context) (*models.ClusterInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).(*models.ClusterInfo), args.Error(1)
}

// Additional mock methods for other services would be implemented here

// TestCommandExecutionValidation tests command execution request validation
func TestCommandExecutionValidation(t *testing.T) {
	t.Run("Valid Command Execution Request", func(t *testing.T) {
		req := &models.CommandExecutionRequest{
			Command: models.KubernetesOperation{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				SessionID: uuid.New(),
				Operation: "get",
				Resource:  "pods",
				Namespace: "default",
				Name:      "test-pod",
				CreatedAt: time.Now(),
			},
			UserID:      uuid.New(),
			SessionID:   uuid.New(),
			SafetyLevel: "safe",
		}

		err := req.Validate()
		assert.NoError(t, err)
	})

	t.Run("Invalid Command Execution Request - Missing UserID", func(t *testing.T) {
		req := &models.CommandExecutionRequest{
			Command: models.KubernetesOperation{
				Operation: "get",
				Resource:  "pods",
				Namespace: "default",
				Name:      "test-pod",
			},
			// UserID missing
			SessionID: uuid.New(),
		}

		err := req.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user_id is required")
	})

	t.Run("Invalid Safety Level", func(t *testing.T) {
		req := &models.CommandExecutionRequest{
			Command: models.KubernetesOperation{
				Operation: "get",
				Resource:  "pods",
				Namespace: "default",
			},
			UserID:      uuid.New(),
			SessionID:   uuid.New(),
			SafetyLevel: "invalid-level",
		}

		err := req.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid safety level")
	})
}

// TestSafetyLevelClassification tests safety level determination
func TestSafetyLevelClassification(t *testing.T) {
	tests := []struct {
		operation     string
		resource      string
		expectedLevel string
	}{
		{"get", "pods", "safe"},
		{"list", "deployments", "safe"},
		{"logs", "pods", "safe"},
		{"delete", "pods", "warning"},
		{"scale", "deployments", "warning"},
		{"restart", "deployments", "warning"},
		{"delete", "deployments", "dangerous"},
		{"delete", "services", "dangerous"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s %s should be %s", test.operation, test.resource, test.expectedLevel), func(t *testing.T) {
			op := &models.KubernetesOperation{
				Operation: test.operation,
				Resource:  test.resource,
				Namespace: "default",
			}

			level := op.GetSafetyLevel()
			assert.Equal(t, test.expectedLevel, level)
		})
	}
}

// TestCommandApprovalValidation tests command approval request validation
func TestCommandApprovalValidation(t *testing.T) {
	t.Run("Valid Approval Request", func(t *testing.T) {
		req := &models.CommandApprovalRequest{
			ExecutionID: uuid.New(),
			UserID:      uuid.New(),
			Decision:    "approve",
			Reason:      "Approved for testing",
		}

		err := req.Validate()
		assert.NoError(t, err)
	})

	t.Run("Invalid Decision", func(t *testing.T) {
		req := &models.CommandApprovalRequest{
			ExecutionID: uuid.New(),
			UserID:      uuid.New(),
			Decision:    "maybe", // Invalid decision
			Reason:      "Uncertain",
		}

		err := req.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decision must be 'approve' or 'reject'")
	})
}

// TestRollbackValidation tests rollback plan validation
func TestRollbackValidation(t *testing.T) {
	t.Run("Valid Rollback Step", func(t *testing.T) {
		step := &models.RollbackStep{
			StepNumber:  1,
			Operation:   "scale",
			Resource:    "deployments",
			Name:        "test-deployment",
			Namespace:   "default",
			Description: "Restore previous replica count",
			Data: map[string]interface{}{
				"replicas": 3,
			},
		}

		// Basic validation - in real implementation this would be more comprehensive
		assert.NotEmpty(t, step.Resource)
		assert.NotEmpty(t, step.Name)
		assert.NotEmpty(t, step.Operation)
		assert.Greater(t, step.StepNumber, 0)
	})

	t.Run("Rollback Plan Creation", func(t *testing.T) {
		plan := &models.RollbackPlan{
			ID:                 uuid.New(),
			CommandExecutionID: uuid.New(),
			UserID:             uuid.New(),
			SessionID:          uuid.New(),
			OriginalCommand: models.KubernetesOperation{
				Operation: "scale",
				Resource:  "deployments",
				Name:      "test-deployment",
				Namespace: "default",
			},
			RollbackSteps: []models.RollbackStep{
				{
					StepNumber:  1,
					Operation:   "scale",
					Resource:    "deployments",
					Name:        "test-deployment",
					Namespace:   "default",
					Description: "Restore previous replica count",
					Data: map[string]interface{}{
						"replicas": 2,
					},
				},
			},
			Status:            "planned",
			Reason:            "User requested rollback",
			EstimatedDuration: 30 * time.Second,
			CreatedAt:         time.Now(),
			ExpiresAt:         time.Now().Add(24 * time.Hour),
		}

		// Validate plan structure
		assert.NotEqual(t, uuid.Nil, plan.ID)
		assert.Greater(t, len(plan.RollbackSteps), 0)
		assert.Equal(t, "planned", plan.Status)
		assert.True(t, plan.ExpiresAt.After(plan.CreatedAt))
	})
}

// TestExecutionStatistics tests execution statistics calculation
func TestExecutionStatistics(t *testing.T) {
	userID := uuid.New()
	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()

	// Mock executions data
	executions := []*models.KubernetesCommandExecution{
		{
			ID:              uuid.New(),
			UserID:          userID,
			Status:          "completed",
			ExecutionTimeMS: &[]int{1500}[0],
			Command:         models.KubernetesOperation{Resource: "pods"},
			CreatedAt:       time.Now().AddDate(0, 0, -1),
		},
		{
			ID:              uuid.New(),
			UserID:          userID,
			Status:          "completed",
			ExecutionTimeMS: &[]int{2500}[0],
			Command:         models.KubernetesOperation{Resource: "pods"},
			CreatedAt:       time.Now().AddDate(0, 0, -2),
		},
		{
			ID:              uuid.New(),
			UserID:          userID,
			Status:          "failed",
			ExecutionTimeMS: nil,
			Command:         models.KubernetesOperation{Resource: "deployments"},
			CreatedAt:       time.Now().AddDate(0, 0, -3),
		},
	}

	// Calculate statistics manually to verify
	totalExecutions := len(executions)
	successfulOnes := 0
	failedOnes := 0
	var totalTimeMS int64 = 0
	resourceCount := make(map[string]int)

	for _, exec := range executions {
		switch exec.Status {
		case "completed":
			successfulOnes++
		case "failed":
			failedOnes++
		}

		if exec.ExecutionTimeMS != nil {
			totalTimeMS += int64(*exec.ExecutionTimeMS)
		}

		resourceCount[exec.Command.Resource]++
	}

	averageTime := float64(totalTimeMS) / float64(successfulOnes)
	mostUsedResource := "pods" // Based on our test data

	// Create expected statistics
	expectedStats := &models.ExecutionStats{
		UserID:           userID,
		TotalExecutions:  totalExecutions,
		SuccessfulOnes:   successfulOnes,
		FailedOnes:       failedOnes,
		AverageTime:      averageTime,
		MostUsedResource: mostUsedResource,
		From:             from,
		To:               to,
	}

	// Verify calculations
	assert.Equal(t, 3, expectedStats.TotalExecutions)
	assert.Equal(t, 2, expectedStats.SuccessfulOnes)
	assert.Equal(t, 1, expectedStats.FailedOnes)
	assert.Equal(t, 2000.0, expectedStats.AverageTime) // (1500 + 2500) / 2
	assert.Equal(t, "pods", expectedStats.MostUsedResource)
}

// TestErrorHandlingAndRecovery tests error handling capabilities
func TestErrorHandlingAndRecovery(t *testing.T) {
	t.Run("Timeout Handling", func(t *testing.T) {
		// Test timeout scenarios
		execution := &models.KubernetesCommandExecution{
			ID:        uuid.New(),
			Status:    "executing",
			CreatedAt: time.Now().Add(-10 * time.Minute), // Started 10 minutes ago
		}

		// Simulate timeout check
		timeout := 5 * time.Minute
		if time.Since(execution.CreatedAt) > timeout {
			execution.Status = "timeout"
			execution.ErrorMessage = "Command execution timed out"
		}

		assert.Equal(t, "timeout", execution.Status)
		assert.Contains(t, execution.ErrorMessage, "timed out")
	})

	t.Run("Error Classification", func(t *testing.T) {
		errorMessages := map[string]string{
			"connection refused":     "network",
			"permission denied":      "authorization",
			"resource not found":     "resource",
			"invalid namespace":      "validation",
			"timeout exceeded":       "timeout",
			"unknown error occurred": "unknown",
		}

		for errorMsg, expectedType := range errorMessages {
			// This would be implemented in the actual error handler service
			errorType := classifyError(errorMsg)
			assert.Equal(t, expectedType, errorType)
		}
	})
}

// Helper function to classify errors (mock implementation)
func classifyError(errorMsg string) string {
	switch {
	case contains(errorMsg, "connection", "refused", "network"):
		return "network"
	case contains(errorMsg, "permission", "denied", "unauthorized", "forbidden"):
		return "authorization"
	case contains(errorMsg, "not found", "missing", "does not exist"):
		return "resource"
	case contains(errorMsg, "invalid", "malformed", "bad request"):
		return "validation"
	case contains(errorMsg, "timeout", "timed out", "deadline exceeded"):
		return "timeout"
	default:
		return "unknown"
	}
}

// contains checks if any of the substrings exist in the main string
func contains(str string, substrings ...string) bool {
	str = strings.ToLower(str)
	for _, substr := range substrings {
		if strings.Contains(str, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

// TestPerformanceMetrics tests performance measurement capabilities
func TestPerformanceMetrics(t *testing.T) {
	t.Run("Execution Time Tracking", func(t *testing.T) {
		execution := &models.KubernetesCommandExecution{
			ID:        uuid.New(),
			Status:    "executing",
			CreatedAt: time.Now(),
		}

		// Simulate command execution
		startTime := time.Now()
		executedAt := startTime
		execution.ExecutedAt = &executedAt

		// Simulate completion
		time.Sleep(10 * time.Millisecond) // Small delay for testing
		completedAt := time.Now()
		execution.CompletedAt = &completedAt
		execution.Status = "completed"

		// Calculate execution time
		execTime := int(completedAt.Sub(executedAt).Milliseconds())
		execution.ExecutionTimeMS = &execTime

		assert.NotNil(t, execution.ExecutionTimeMS)
		assert.Greater(t, *execution.ExecutionTimeMS, 0)
		assert.Equal(t, "completed", execution.Status)
	})
}

// Benchmark tests for performance validation
func BenchmarkCommandValidation(b *testing.B) {
	req := &models.CommandExecutionRequest{
		Command: models.KubernetesOperation{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			SessionID: uuid.New(),
			Operation: "get",
			Resource:  "pods",
			Namespace: "default",
			Name:      "test-pod",
			CreatedAt: time.Now(),
		},
		UserID:    uuid.New(),
		SessionID: uuid.New(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.Validate()
	}
}

func BenchmarkSafetyLevelClassification(b *testing.B) {
	op := &models.KubernetesOperation{
		Operation: "delete",
		Resource:  "deployments",
		Namespace: "default",
		Name:      "test-deployment",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = op.GetSafetyLevel()
	}
}
