package nlp_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/nlp"
)

// MockOllamaService for testing
type MockOllamaService struct{}

func (m *MockOllamaService) ProcessQuery(ctx context.Context, request *models.NLPRequest) (*models.NLPResponse, error) {
	return &models.NLPResponse{
		ID:               uuid.New(),
		RequestID:        request.ID,
		GeneratedCommand: "kubectl get pods -n default",
		Explanation:      "Lists all pods in the default namespace",
		SafetyLevel:      models.NLPSafetyLevelSafe,
		Confidence:       0.95,
		Provider:         models.ProviderOllama,
		ProcessingTimeMs: 100,
		TokensUsed:       50,
		CreatedAt:        time.Now(),
	}, nil
}

func (m *MockOllamaService) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *MockOllamaService) IsAvailable(ctx context.Context) bool {
	return true
}

// MockOpenAIService for testing
type MockOpenAIService struct{}

func (m *MockOpenAIService) ProcessQuery(ctx context.Context, request *models.NLPRequest) (*models.NLPResponse, error) {
	return &models.NLPResponse{
		ID:               uuid.New(),
		RequestID:        request.ID,
		GeneratedCommand: "kubectl get deployments -n production",
		Explanation:      "Lists all deployments in the production namespace",
		SafetyLevel:      models.NLPSafetyLevelSafe,
		Confidence:       0.98,
		Provider:         models.ProviderOpenAI,
		ProcessingTimeMs: 150,
		TokensUsed:       75,
		CreatedAt:        time.Now(),
	}, nil
}

func (m *MockOpenAIService) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *MockOpenAIService) IsAvailable(ctx context.Context) bool {
	return true
}

func TestNLPService_ProcessQuery(t *testing.T) {
	tests := []struct {
		name           string
		request        *models.NLPRequest
		expectError    bool
		expectedResult *models.NLPResponse
	}{
		{
			name: "successful query processing",
			request: &models.NLPRequest{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				SessionID: uuid.New(),
				Query:     "show me all pods in default namespace",
				Provider:  models.ProviderOllama,
				CreatedAt: time.Now(),
			},
			expectError: false,
		},
		{
			name: "invalid request - empty query",
			request: &models.NLPRequest{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				SessionID: uuid.New(),
				Query:     "",
				Provider:  models.ProviderOllama,
				CreatedAt: time.Now(),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock services
			mockOllama := &MockOllamaService{}
			mockOpenAI := &MockOpenAIService{}

			// Create service
			config := &nlp.Config{
				DefaultProvider:    models.ProviderOllama,
				EnableFallback:     true,
				MaxRetries:         3,
				TimeoutSeconds:     30,
				EnableCaching:      false,
				CacheTTLMinutes:    60,
				EnableRateLimiting: false,
				RateLimit:          30,
			}

			service := nlp.NewService(mockOllama, mockOpenAI, config)

			// Process the query
			result, err := service.ProcessQuery(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.GeneratedCommand)
				assert.NotEmpty(t, result.Explanation)
				assert.True(t, result.Confidence > 0.0)
			}
		})
	}
}

func TestNLPService_ValidateCommand(t *testing.T) {
	tests := []struct {
		name           string
		command        string
		expectError    bool
		expectedSafety models.SafetyLevel
		expectedValid  bool
	}{
		{
			name:           "valid safe command",
			command:        "kubectl get pods -n default",
			expectError:    false,
			expectedSafety: models.NLPSafetyLevelSafe,
			expectedValid:  true,
		},
		{
			name:           "valid warning command",
			command:        "kubectl create deployment test --image=nginx",
			expectError:    false,
			expectedSafety: models.NLPSafetyLevelWarning,
			expectedValid:  true,
		},
		{
			name:           "dangerous command",
			command:        "kubectl delete pod test-pod --force",
			expectError:    false,
			expectedSafety: models.NLPSafetyLevelDangerous,
			expectedValid:  true,
		},
		{
			name:           "empty command",
			command:        "",
			expectError:    false,
			expectedSafety: models.NLPSafetyLevelDangerous,
			expectedValid:  false,
		},
		{
			name:           "invalid command",
			command:        "not a kubectl command",
			expectError:    false,
			expectedSafety: models.NLPSafetyLevelDangerous,
			expectedValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock services
			mockOllama := &MockOllamaService{}
			mockOpenAI := &MockOpenAIService{}

			// Create service
			config := &nlp.Config{
				DefaultProvider:    models.ProviderOllama,
				EnableFallback:     true,
				MaxRetries:         3,
				TimeoutSeconds:     30,
				EnableCaching:      false,
				CacheTTLMinutes:    60,
				EnableRateLimiting: false,
				RateLimit:          30,
			}

			service := nlp.NewService(mockOllama, mockOpenAI, config)

			// Validate the command
			result, err := service.ValidateCommand(context.Background(), tt.command)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedSafety, result.SafetyLevel)
				assert.Equal(t, tt.expectedValid, result.IsValid)
			}
		})
	}
}

func TestNLPService_HealthCheck(t *testing.T) {
	// Create mock services
	mockOllama := &MockOllamaService{}
	mockOpenAI := &MockOpenAIService{}

	// Create service
	config := &nlp.Config{
		DefaultProvider:    models.ProviderOllama,
		EnableFallback:     true,
		MaxRetries:         3,
		TimeoutSeconds:     30,
		EnableCaching:      false,
		CacheTTLMinutes:    60,
		EnableRateLimiting: false,
		RateLimit:          30,
	}

	service := nlp.NewService(mockOllama, mockOpenAI, config)

	// Perform health check
	err := service.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestNLPService_GetSupportedProviders(t *testing.T) {
	// Create mock services
	mockOllama := &MockOllamaService{}
	mockOpenAI := &MockOpenAIService{}

	// Create service
	config := &nlp.Config{
		DefaultProvider:    models.ProviderOllama,
		EnableFallback:     true,
		MaxRetries:         3,
		TimeoutSeconds:     30,
		EnableCaching:      false,
		CacheTTLMinutes:    60,
		EnableRateLimiting: false,
		RateLimit:          30,
	}

	service := nlp.NewService(mockOllama, mockOpenAI, config)

	// Get supported providers
	providers, err := service.GetSupportedProviders(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, providers)
	assert.Contains(t, providers, models.ProviderOllama)
	assert.Contains(t, providers, models.ProviderOpenAI)
}

func TestSafetyLevel_Methods(t *testing.T) {
	// Test GetSafetyDescription
	assert.Equal(t, "Safe operation - read-only or non-destructive commands",
		models.NLPSafetyLevelSafe.GetSafetyDescription())
	assert.Equal(t, "Warning - potentially impactful operation, review before executing",
		models.NLPSafetyLevelWarning.GetSafetyDescription())
	assert.Equal(t, "Dangerous - destructive operation that could cause service disruption",
		models.NLPSafetyLevelDangerous.GetSafetyDescription())

	// Test GetSafetyColor
	assert.Equal(t, "#28a745", models.NLPSafetyLevelSafe.GetSafetyColor())      // Green
	assert.Equal(t, "#ffc107", models.NLPSafetyLevelWarning.GetSafetyColor())   // Yellow
	assert.Equal(t, "#dc3545", models.NLPSafetyLevelDangerous.GetSafetyColor()) // Red
}

func TestNLPRequest_Validate(t *testing.T) {
	tests := []struct {
		name        string
		request     *models.NLPRequest
		expectError bool
	}{
		{
			name: "valid request",
			request: &models.NLPRequest{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				SessionID: uuid.New(),
				Query:     "show me pods",
				Provider:  models.ProviderOllama,
				CreatedAt: time.Now(),
			},
			expectError: false,
		},
		{
			name: "empty query",
			request: &models.NLPRequest{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				SessionID: uuid.New(),
				Query:     "",
				Provider:  models.ProviderOllama,
				CreatedAt: time.Now(),
			},
			expectError: true,
		},
		{
			name: "missing user ID",
			request: &models.NLPRequest{
				ID:        uuid.New(),
				SessionID: uuid.New(),
				Query:     "show me pods",
				Provider:  models.ProviderOllama,
				CreatedAt: time.Now(),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
