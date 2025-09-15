package audit_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/audit"
)

// MockAuditRepository for testing
type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) CreateAuditLog(ctx context.Context, auditLog *models.AuditLog) error {
	args := m.Called(ctx, auditLog)
	return args.Error(0)
}

func (m *MockAuditRepository) GetAuditLogByID(ctx context.Context, id int64) (*models.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) GetAuditLogs(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) GetAuditLogSummary(ctx context.Context, filter models.AuditLogFilter) (*models.AuditLogSummary, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuditLogSummary), args.Error(1)
}

func (m *MockAuditRepository) GetLastChecksum(ctx context.Context) (*string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockAuditRepository) VerifyIntegrity(ctx context.Context, startID, endID int64) ([]models.IntegrityCheckResult, error) {
	args := m.Called(ctx, startID, endID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.IntegrityCheckResult), args.Error(1)
}

func (m *MockAuditRepository) GetAuditLogsByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.AuditLog, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) GetDangerousOperations(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) GetFailedOperations(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) CountAuditLogs(ctx context.Context, filter models.AuditLogFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func TestNewService(t *testing.T) {
	mockRepo := &MockAuditRepository{}

	config := &audit.Config{
		EnableAsyncLogging:      true,
		AsyncBufferSize:         100,
		MaxBatchSize:            50,
		EnableStructuredLogging: true,
		LogLevel:                "info",
		EnableIntegrityCheck:    true,
		RetentionDays:           30,
	}

	service := audit.NewService(mockRepo, config)
	assert.NotNil(t, service)
}

func TestAuditService_LogUserAction(t *testing.T) {
	tests := []struct {
		name        string
		request     models.AuditLogRequest
		setupMock   func(*MockAuditRepository)
		expectError bool
	}{
		{
			name: "successful audit log creation",
			request: models.AuditLogRequest{
				UserID:           &uuid.UUID{},
				SessionID:        &uuid.UUID{},
				QueryText:        "show me all pods",
				GeneratedCommand: "kubectl get pods --all-namespaces",
				SafetyLevel:      models.SafetyLevelSafe,
				ExecutionStatus:  models.ExecutionStatusSuccess,
			},
			setupMock: func(mockRepo *MockAuditRepository) {
				mockRepo.On("CreateAuditLog", mock.Anything, mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name: "invalid request - empty query",
			request: models.AuditLogRequest{
				UserID:           &uuid.UUID{},
				SessionID:        &uuid.UUID{},
				QueryText:        "",
				GeneratedCommand: "kubectl get pods",
				SafetyLevel:      models.SafetyLevelSafe,
				ExecutionStatus:  models.ExecutionStatusSuccess,
			},
			setupMock:   func(mockRepo *MockAuditRepository) {},
			expectError: true,
		},
		{
			name: "invalid request - empty command",
			request: models.AuditLogRequest{
				UserID:           &uuid.UUID{},
				SessionID:        &uuid.UUID{},
				QueryText:        "show pods",
				GeneratedCommand: "",
				SafetyLevel:      models.SafetyLevelSafe,
				ExecutionStatus:  models.ExecutionStatusSuccess,
			},
			setupMock:   func(mockRepo *MockAuditRepository) {},
			expectError: true,
		},
		{
			name: "invalid safety level",
			request: models.AuditLogRequest{
				UserID:           &uuid.UUID{},
				SessionID:        &uuid.UUID{},
				QueryText:        "show pods",
				GeneratedCommand: "kubectl get pods",
				SafetyLevel:      "invalid",
				ExecutionStatus:  models.ExecutionStatusSuccess,
			},
			setupMock:   func(mockRepo *MockAuditRepository) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockAuditRepository{}
			tt.setupMock(mockRepo)

			config := &audit.Config{
				EnableAsyncLogging: false, // Use sync for testing
			}
			service := audit.NewService(mockRepo, config)

			err := service.LogUserAction(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuditService_GetAuditLogs(t *testing.T) {
	mockRepo := &MockAuditRepository{}

	userID := uuid.New()
	sessionID := uuid.New()

	expectedLogs := []*models.AuditLog{
		{
			ID:               1,
			UserID:           &userID,
			SessionID:        &sessionID,
			QueryText:        "show pods",
			GeneratedCommand: "kubectl get pods",
			SafetyLevel:      models.SafetyLevelSafe,
			ExecutionStatus:  models.ExecutionStatusSuccess,
			Timestamp:        time.Now(),
		},
		{
			ID:               2,
			UserID:           &userID,
			SessionID:        &sessionID,
			QueryText:        "delete pod test",
			GeneratedCommand: "kubectl delete pod test",
			SafetyLevel:      models.SafetyLevelDangerous,
			ExecutionStatus:  models.ExecutionStatusSuccess,
			Timestamp:        time.Now(),
		},
	}

	filter := models.AuditLogFilter{
		UserID: &userID,
		Limit:  50,
	}

	mockRepo.On("GetAuditLogs", mock.Anything, filter).Return(expectedLogs, nil)

	service := audit.NewService(mockRepo, nil)
	logs, err := service.GetAuditLogs(context.Background(), filter)

	assert.NoError(t, err)
	assert.Equal(t, expectedLogs, logs)
	assert.Len(t, logs, 2)

	mockRepo.AssertExpectations(t)
}

func TestAuditService_GetAuditLogSummary(t *testing.T) {
	mockRepo := &MockAuditRepository{}

	expectedSummary := &models.AuditLogSummary{
		TotalEntries:   100,
		SafeOperations: 60,
		WarningOps:     25,
		DangerousOps:   15,
		SuccessfulOps:  90,
		FailedOps:      8,
		CancelledOps:   2,
	}

	filter := models.AuditLogFilter{}
	mockRepo.On("GetAuditLogSummary", mock.Anything, filter).Return(expectedSummary, nil)

	service := audit.NewService(mockRepo, nil)
	summary, err := service.GetAuditLogSummary(context.Background(), filter)

	assert.NoError(t, err)
	assert.Equal(t, expectedSummary, summary)
	assert.Equal(t, int64(100), summary.TotalEntries)
	assert.Equal(t, int64(15), summary.DangerousOps)

	mockRepo.AssertExpectations(t)
}

func TestAuditService_VerifyIntegrity(t *testing.T) {
	mockRepo := &MockAuditRepository{}

	expectedResults := []models.IntegrityCheckResult{
		{
			LogID:   1,
			IsValid: true,
		},
		{
			LogID:   2,
			IsValid: true,
		},
		{
			LogID:        3,
			IsValid:      false,
			ErrorMessage: "Checksum verification failed",
		},
	}

	mockRepo.On("VerifyIntegrity", mock.Anything, int64(1), int64(3)).Return(expectedResults, nil)

	service := audit.NewService(mockRepo, nil)
	results, err := service.VerifyIntegrity(context.Background(), 1, 3)

	assert.NoError(t, err)
	assert.Equal(t, expectedResults, results)
	assert.Len(t, results, 3)

	// Verify that integrity check results are correctly classified
	validCount := 0
	invalidCount := 0
	for _, result := range results {
		if result.IsValid {
			validCount++
		} else {
			invalidCount++
		}
	}

	assert.Equal(t, 2, validCount)
	assert.Equal(t, 1, invalidCount)

	mockRepo.AssertExpectations(t)
}

func TestAuditService_GetDangerousOperations(t *testing.T) {
	mockRepo := &MockAuditRepository{}

	userID := uuid.New()

	expectedLogs := []*models.AuditLog{
		{
			ID:               1,
			UserID:           &userID,
			QueryText:        "delete all pods",
			GeneratedCommand: "kubectl delete pods --all",
			SafetyLevel:      models.SafetyLevelDangerous,
			ExecutionStatus:  models.ExecutionStatusSuccess,
			Timestamp:        time.Now(),
		},
		{
			ID:               2,
			UserID:           &userID,
			QueryText:        "force delete namespace",
			GeneratedCommand: "kubectl delete namespace test --force",
			SafetyLevel:      models.SafetyLevelDangerous,
			ExecutionStatus:  models.ExecutionStatusFailed,
			Timestamp:        time.Now(),
		},
	}

	filter := models.AuditLogFilter{
		UserID: &userID,
	}

	mockRepo.On("GetDangerousOperations", mock.Anything, filter).Return(expectedLogs, nil)

	service := audit.NewService(mockRepo, nil)
	logs, err := service.GetDangerousOperations(context.Background(), filter)

	assert.NoError(t, err)
	assert.Equal(t, expectedLogs, logs)
	assert.Len(t, logs, 2)

	// Verify all returned logs are dangerous
	for _, log := range logs {
		assert.Equal(t, models.SafetyLevelDangerous, log.SafetyLevel)
	}

	mockRepo.AssertExpectations(t)
}

func TestAuditService_GetFailedOperations(t *testing.T) {
	mockRepo := &MockAuditRepository{}

	userID := uuid.New()

	expectedLogs := []*models.AuditLog{
		{
			ID:               1,
			UserID:           &userID,
			QueryText:        "get invalid resource",
			GeneratedCommand: "kubectl get invalidresource",
			SafetyLevel:      models.SafetyLevelSafe,
			ExecutionStatus:  models.ExecutionStatusFailed,
			Timestamp:        time.Now(),
		},
	}

	filter := models.AuditLogFilter{
		UserID: &userID,
	}

	mockRepo.On("GetFailedOperations", mock.Anything, filter).Return(expectedLogs, nil)

	service := audit.NewService(mockRepo, nil)
	logs, err := service.GetFailedOperations(context.Background(), filter)

	assert.NoError(t, err)
	assert.Equal(t, expectedLogs, logs)
	assert.Len(t, logs, 1)

	// Verify all returned logs are failed
	for _, log := range logs {
		assert.Equal(t, models.ExecutionStatusFailed, log.ExecutionStatus)
	}

	mockRepo.AssertExpectations(t)
}

func TestAuditService_HealthCheck(t *testing.T) {
	mockRepo := &MockAuditRepository{}

	mockRepo.On("GetLastChecksum", mock.Anything).Return(nil, nil)

	service := audit.NewService(mockRepo, nil)
	err := service.HealthCheck(context.Background())

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_GetMetrics(t *testing.T) {
	mockRepo := &MockAuditRepository{}

	service := audit.NewService(mockRepo, &audit.Config{
		EnableAsyncLogging: false,
	})

	metrics, err := service.GetMetrics(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics.TotalLogsCreated, int64(0))
	assert.GreaterOrEqual(t, metrics.SuccessRate, float64(0))
	assert.LessOrEqual(t, metrics.SuccessRate, float64(100))
}

func TestAuditService_LogKubectlExecution(t *testing.T) {
	mockRepo := &MockAuditRepository{}

	userID := uuid.New()
	sessionID := uuid.New()

	mockRepo.On("CreateAuditLog", mock.Anything, mock.Anything).Return(nil)

	service := audit.NewService(mockRepo, &audit.Config{
		EnableAsyncLogging: false,
	})

	result := map[string]interface{}{
		"pods_found": 5,
		"namespace":  "default",
	}

	err := service.LogKubectlExecution(
		context.Background(),
		&userID,
		&sessionID,
		"show me pods in default namespace",
		"kubectl get pods -n default",
		result,
		models.ExecutionStatusSuccess,
		nil,
	)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogSecurityEvent(t *testing.T) {
	mockRepo := &MockAuditRepository{}

	userID := uuid.New()

	mockRepo.On("CreateAuditLog", mock.Anything, mock.Anything).Return(nil)

	service := audit.NewService(mockRepo, &audit.Config{
		EnableAsyncLogging: false,
	})

	err := service.LogSecurityEvent(
		context.Background(),
		"unauthorized_access_attempt",
		"User attempted to access restricted endpoint",
		&userID,
		"high",
		nil,
	)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_ClassifyCommandSafety(t *testing.T) {
	tests := []struct {
		name           string
		command        string
		expectedSafety string
	}{
		{
			name:           "safe read operation",
			command:        "kubectl get pods",
			expectedSafety: models.SafetyLevelSafe,
		},
		{
			name:           "warning create operation",
			command:        "kubectl create deployment test --image=nginx",
			expectedSafety: models.SafetyLevelWarning,
		},
		{
			name:           "dangerous delete operation",
			command:        "kubectl delete pod test --force",
			expectedSafety: models.SafetyLevelDangerous,
		},
		{
			name:           "empty command",
			command:        "",
			expectedSafety: models.SafetyLevelDangerous,
		},
		{
			name:           "dangerous destroy operation",
			command:        "kubectl destroy cluster",
			expectedSafety: models.SafetyLevelDangerous,
		},
		{
			name:           "warning patch operation",
			command:        "kubectl patch deployment test -p '{\"spec\":{\"replicas\":3}}'",
			expectedSafety: models.SafetyLevelWarning,
		},
	}

	mockRepo := &MockAuditRepository{}
	service := audit.NewService(mockRepo, nil)

	// We need to use reflection or create a public method to test this private method
	// For now, we'll test it indirectly through LogKubectlExecution
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.On("CreateAuditLog", mock.Anything, mock.MatchedBy(func(log *models.AuditLog) bool {
				return log.SafetyLevel == tt.expectedSafety
			})).Return(nil).Once()

			userID := uuid.New()
			err := service.LogKubectlExecution(
				context.Background(),
				&userID,
				&userID,
				"test query",
				tt.command,
				nil,
				models.ExecutionStatusSuccess,
				nil,
			)

			assert.NoError(t, err)
		})
	}

	mockRepo.AssertExpectations(t)
}
