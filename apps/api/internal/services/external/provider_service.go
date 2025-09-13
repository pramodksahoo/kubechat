package external

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ProviderInterface defines the pluggable provider interface (Task 7.1)
type ProviderInterface interface {
	// Core provider methods
	GetName() string
	GetType() string
	GetVersion() string
	IsHealthy(ctx context.Context) bool
	GetCapabilities() []string
	GetSupportedModels() []string

	// Request handling
	ProcessRequest(ctx context.Context, request *ProviderRequest) (*ProviderResponse, error)
	ValidateRequest(request *ProviderRequest) error
	EstimateCost(request *ProviderRequest) (*CostEstimate, error)

	// Configuration and lifecycle
	Initialize(config *ProviderConfig) error
	Shutdown(ctx context.Context) error
	GetConfiguration() *ProviderConfig
	UpdateConfiguration(config *ProviderConfig) error

	// Metrics and monitoring
	GetMetrics(ctx context.Context) (*MultiProviderMetrics, error)
	GetStatus(ctx context.Context) *ProviderStatus
}

// ProviderRegistry manages multiple AI service providers (Task 7.2)
type ProviderRegistry interface {
	RegisterProvider(provider ProviderInterface) error
	UnregisterProvider(providerName string) error
	GetProvider(providerName string) (ProviderInterface, error)
	GetAllProviders() []ProviderInterface
	GetProvidersByType(providerType string) []ProviderInterface
	GetProvidersByCapability(capability string) []ProviderInterface
	DiscoverProviders() error
	RefreshProviders() error
}

// ProviderLoadBalancer handles load balancing across providers (Task 7.3)
type ProviderLoadBalancer interface {
	SelectProvider(ctx context.Context, request *ProviderRequest) (ProviderInterface, error)
	UpdateProviderWeights(weights map[string]float64) error
	GetProviderWeights() map[string]float64
	SetLoadBalancingStrategy(strategy LoadBalancingStrategy) error
	GetLoadBalancingStats() *LoadBalancingStats
}

// ProviderFailover handles failover and redundancy (Task 7.4)
type ProviderFailover interface {
	ExecuteWithFailover(ctx context.Context, request *ProviderRequest, primaryProvider string) (*ProviderResponse, error)
	ConfigureFailoverRules(rules []*FailoverRule) error
	GetFailoverStatus() *FailoverStatus
	TriggerManualFailover(fromProvider, toProvider string) error
	GetFailoverHistory() []*FailoverEvent
}

// ProviderConfigManager handles provider-specific configurations (Task 7.5)
type ProviderConfigManager interface {
	SetProviderConfig(providerName string, config *ProviderConfig) error
	GetProviderConfig(providerName string) (*ProviderConfig, error)
	ValidateProviderConfig(config *ProviderConfig) error
	UpdateProviderConfig(providerName string, updates map[string]interface{}) error
	GetAllProviderConfigs() map[string]*ProviderConfig
	BackupConfigurations() error
	RestoreConfigurations(backup string) error
}

// Data structures for provider support

type ProviderRequest struct {
	ID           string                 `json:"id"`
	ProviderName string                 `json:"provider_name,omitempty"`
	RequestType  string                 `json:"request_type"`
	Model        string                 `json:"model"`
	Input        string                 `json:"input"`
	Parameters   map[string]interface{} `json:"parameters"`
	Context      map[string]string      `json:"context"`
	Priority     int                    `json:"priority"`
	Timeout      time.Duration          `json:"timeout"`
	RetryPolicy  *RetryPolicy           `json:"retry_policy"`
	Metadata     map[string]string      `json:"metadata"`
	Timestamp    time.Time              `json:"timestamp"`
}

type ProviderResponse struct {
	ID             string                 `json:"id"`
	ProviderName   string                 `json:"provider_name"`
	Model          string                 `json:"model"`
	Output         string                 `json:"output"`
	Confidence     float64                `json:"confidence"`
	TokensUsed     int64                  `json:"tokens_used"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Cost           float64                `json:"cost"`
	Metadata       map[string]interface{} `json:"metadata"`
	Error          string                 `json:"error,omitempty"`
	Success        bool                   `json:"success"`
	Timestamp      time.Time              `json:"timestamp"`
}

type ProviderConfig struct {
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`
	Version          string                 `json:"version"`
	Enabled          bool                   `json:"enabled"`
	BaseURL          string                 `json:"base_url"`
	APIKey           string                 `json:"api_key,omitempty"`
	SecretKeyName    string                 `json:"secret_key_name,omitempty"`
	Models           []string               `json:"models"`
	Capabilities     []string               `json:"capabilities"`
	MaxConcurrency   int                    `json:"max_concurrency"`
	RateLimits       *RateLimits            `json:"rate_limits"`
	Timeouts         *TimeoutConfig         `json:"timeouts"`
	HealthCheck      *HealthCheckConfig     `json:"health_check"`
	CustomParameters map[string]interface{} `json:"custom_parameters"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

type MultiProviderMetrics struct {
	ProviderName        string        `json:"provider_name"`
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	TokensProcessed     int64         `json:"tokens_processed"`
	TotalCost           float64       `json:"total_cost"`
	SuccessRate         float64       `json:"success_rate"`
	CurrentLoad         int           `json:"current_load"`
	MaxLoad             int           `json:"max_load"`
	HealthStatus        string        `json:"health_status"`
	LastHealthCheck     time.Time     `json:"last_health_check"`
	Timestamp           time.Time     `json:"timestamp"`
}

type ProviderStatus struct {
	ProviderName    string                 `json:"provider_name"`
	Type            string                 `json:"type"`
	Version         string                 `json:"version"`
	Enabled         bool                   `json:"enabled"`
	Healthy         bool                   `json:"healthy"`
	Available       bool                   `json:"available"`
	Load            float64                `json:"load_percentage"`
	ResponseTime    time.Duration          `json:"avg_response_time"`
	ErrorRate       float64                `json:"error_rate"`
	Capabilities    []string               `json:"capabilities"`
	SupportedModels []string               `json:"supported_models"`
	LastSeen        time.Time              `json:"last_seen"`
	Details         map[string]interface{} `json:"details"`
}

type CostEstimate struct {
	ProviderName    string  `json:"provider_name"`
	Model           string  `json:"model"`
	EstimatedTokens int64   `json:"estimated_tokens"`
	TokenCost       float64 `json:"token_cost"`
	RequestCost     float64 `json:"request_cost"`
	TotalCost       float64 `json:"total_cost"`
	Currency        string  `json:"currency"`
	Confidence      float64 `json:"confidence"`
}

type RetryPolicy struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors"`
}

type RateLimits struct {
	RequestsPerSecond  int `json:"requests_per_second"`
	RequestsPerMinute  int `json:"requests_per_minute"`
	RequestsPerHour    int `json:"requests_per_hour"`
	TokensPerMinute    int `json:"tokens_per_minute"`
	ConcurrentRequests int `json:"concurrent_requests"`
}

type TimeoutConfig struct {
	RequestTimeout    time.Duration `json:"request_timeout"`
	ConnectionTimeout time.Duration `json:"connection_timeout"`
	ReadTimeout       time.Duration `json:"read_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout"`
}

type HealthCheckConfig struct {
	Enabled          bool          `json:"enabled"`
	Interval         time.Duration `json:"interval"`
	Timeout          time.Duration `json:"timeout"`
	FailureThreshold int           `json:"failure_threshold"`
	SuccessThreshold int           `json:"success_threshold"`
	Endpoint         string        `json:"endpoint"`
}

type LoadBalancingStrategy string

const (
	LoadBalancingRoundRobin   LoadBalancingStrategy = "round_robin"
	LoadBalancingWeighted     LoadBalancingStrategy = "weighted"
	LoadBalancingLeastLoaded  LoadBalancingStrategy = "least_loaded"
	LoadBalancingResponseTime LoadBalancingStrategy = "response_time"
	LoadBalancingCost         LoadBalancingStrategy = "cost"
	LoadBalancingAvailability LoadBalancingStrategy = "availability"
)

type LoadBalancingStats struct {
	Strategy           LoadBalancingStrategy       `json:"strategy"`
	TotalRequests      int64                       `json:"total_requests"`
	ProviderRequests   map[string]int64            `json:"provider_requests"`
	ProviderWeights    map[string]float64          `json:"provider_weights"`
	SelectionHistory   []*SelectionEvent           `json:"selection_history"`
	PerformanceMetrics map[string]*ProviderMetrics `json:"performance_metrics"`
	Timestamp          time.Time                   `json:"timestamp"`
}

type SelectionEvent struct {
	ProviderName    string                 `json:"provider_name"`
	RequestID       string                 `json:"request_id"`
	SelectionReason string                 `json:"selection_reason"`
	LoadFactor      float64                `json:"load_factor"`
	ResponseTime    time.Duration          `json:"response_time"`
	Success         bool                   `json:"success"`
	Metadata        map[string]interface{} `json:"metadata"`
	Timestamp       time.Time              `json:"timestamp"`
}

type FailoverRule struct {
	ID                string              `json:"id"`
	Name              string              `json:"name"`
	PrimaryProvider   string              `json:"primary_provider"`
	BackupProviders   []string            `json:"backup_providers"`
	TriggerConditions []*TriggerCondition `json:"trigger_conditions"`
	FailoverStrategy  string              `json:"failover_strategy"`
	Enabled           bool                `json:"enabled"`
	Priority          int                 `json:"priority"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
}

type TriggerCondition struct {
	Type      string        `json:"type"` // "error_rate", "response_time", "availability"
	Threshold interface{}   `json:"threshold"`
	Duration  time.Duration `json:"duration"`
	Operator  string        `json:"operator"` // ">=", ">", "<=", "<", "=="
}

type FailoverStatus struct {
	ActiveFailovers  map[string]*ActiveFailover `json:"active_failovers"`
	FailoverHistory  []*FailoverEvent           `json:"failover_history"`
	ProviderHealth   map[string]bool            `json:"provider_health"`
	MonitoringActive bool                       `json:"monitoring_active"`
	Timestamp        time.Time                  `json:"timestamp"`
}

type ActiveFailover struct {
	ID               string        `json:"id"`
	PrimaryProvider  string        `json:"primary_provider"`
	CurrentProvider  string        `json:"current_provider"`
	FailoverReason   string        `json:"failover_reason"`
	StartTime        time.Time     `json:"start_time"`
	ExpectedDuration time.Duration `json:"expected_duration"`
	Status           string        `json:"status"`
}

type FailoverEvent struct {
	ID               string        `json:"id"`
	FromProvider     string        `json:"from_provider"`
	ToProvider       string        `json:"to_provider"`
	Reason           string        `json:"reason"`
	TriggerCondition string        `json:"trigger_condition"`
	Duration         time.Duration `json:"duration"`
	Success          bool          `json:"success"`
	Impact           string        `json:"impact"`
	Timestamp        time.Time     `json:"timestamp"`
}

// Implementation of provider registry
type providerRegistryImpl struct {
	mu        sync.RWMutex
	providers map[string]ProviderInterface
	discovery *ProviderDiscovery
}

type ProviderDiscovery struct {
	Enabled           bool          `json:"enabled"`
	DiscoveryInterval time.Duration `json:"discovery_interval"`
	DiscoveryMethods  []string      `json:"discovery_methods"`
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() ProviderRegistry {
	return &providerRegistryImpl{
		providers: make(map[string]ProviderInterface),
		discovery: &ProviderDiscovery{
			Enabled:           true,
			DiscoveryInterval: 5 * time.Minute,
			DiscoveryMethods:  []string{"config", "environment", "service_discovery"},
		},
	}
}

// RegisterProvider registers a new provider in the registry
func (r *providerRegistryImpl) RegisterProvider(provider ProviderInterface) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	r.providers[name] = provider
	return nil
}

// UnregisterProvider removes a provider from the registry
func (r *providerRegistryImpl) UnregisterProvider(providerName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[providerName]; !exists {
		return fmt.Errorf("provider %s not found", providerName)
	}

	delete(r.providers, providerName)
	return nil
}

// GetProvider retrieves a provider by name
func (r *providerRegistryImpl) GetProvider(providerName string) (ProviderInterface, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	return provider, nil
}

// GetAllProviders returns all registered providers
func (r *providerRegistryImpl) GetAllProviders() []ProviderInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]ProviderInterface, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}

	return providers
}

// GetProvidersByType returns providers filtered by type
func (r *providerRegistryImpl) GetProvidersByType(providerType string) []ProviderInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filteredProviders []ProviderInterface
	for _, provider := range r.providers {
		if provider.GetType() == providerType {
			filteredProviders = append(filteredProviders, provider)
		}
	}

	return filteredProviders
}

// GetProvidersByCapability returns providers that support a specific capability
func (r *providerRegistryImpl) GetProvidersByCapability(capability string) []ProviderInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filteredProviders []ProviderInterface
	for _, provider := range r.providers {
		for _, cap := range provider.GetCapabilities() {
			if cap == capability {
				filteredProviders = append(filteredProviders, provider)
				break
			}
		}
	}

	return filteredProviders
}

// DiscoverProviders discovers and registers new providers
func (r *providerRegistryImpl) DiscoverProviders() error {
	// Mock implementation - in real scenario would discover from:
	// - Configuration files
	// - Environment variables
	// - Service discovery systems
	// - Plugin directories

	discoveredProviders := []struct {
		name         string
		providerType string
		config       *ProviderConfig
	}{
		{
			name:         "anthropic_claude",
			providerType: "ai",
			config: &ProviderConfig{
				Name:         "anthropic_claude",
				Type:         "ai",
				Version:      "1.0.0",
				Enabled:      false, // Disabled by default - requires API key
				BaseURL:      "https://api.anthropic.com",
				Models:       []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"},
				Capabilities: []string{"text_generation", "conversation", "reasoning"},
				Timeouts: &TimeoutConfig{
					RequestTimeout: 30 * time.Second,
				},
				CreatedAt: time.Now(),
			},
		},
		{
			name:         "google_gemini",
			providerType: "ai",
			config: &ProviderConfig{
				Name:         "google_gemini",
				Type:         "ai",
				Version:      "1.0.0",
				Enabled:      false, // Disabled by default - requires API key
				BaseURL:      "https://generativelanguage.googleapis.com",
				Models:       []string{"gemini-pro", "gemini-pro-vision"},
				Capabilities: []string{"text_generation", "vision", "multimodal"},
				Timeouts: &TimeoutConfig{
					RequestTimeout: 30 * time.Second,
				},
				CreatedAt: time.Now(),
			},
		},
		{
			name:         "huggingface_inference",
			providerType: "ai",
			config: &ProviderConfig{
				Name:         "huggingface_inference",
				Type:         "ai",
				Version:      "1.0.0",
				Enabled:      false, // Disabled by default - requires API key
				BaseURL:      "https://api-inference.huggingface.co",
				Models:       []string{"gpt2", "bert-base-uncased", "distilbert-base-uncased"},
				Capabilities: []string{"text_generation", "text_classification", "embeddings"},
				Timeouts: &TimeoutConfig{
					RequestTimeout: 30 * time.Second,
				},
				CreatedAt: time.Now(),
			},
		},
	}

	for _, discovered := range discoveredProviders {
		// Create a mock provider implementation
		mockProvider := &MockProvider{
			config: discovered.config,
		}

		// Register if not already present
		if _, err := r.GetProvider(discovered.name); err != nil {
			if err := r.RegisterProvider(mockProvider); err != nil {
				fmt.Printf("Warning: Failed to register discovered provider %s: %v\n", discovered.name, err)
			}
		}
	}

	return nil
}

// RefreshProviders refreshes all provider configurations
func (r *providerRegistryImpl) RefreshProviders() error {
	r.mu.RLock()
	providers := make([]ProviderInterface, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}
	r.mu.RUnlock()

	// Refresh each provider's status and configuration
	for _, provider := range providers {
		// In a real implementation, this would refresh the provider's configuration
		// from external sources and update its status
		_ = provider.IsHealthy(context.Background())
	}

	return nil
}

// MockProvider is a basic implementation of ProviderInterface for discovered providers
type MockProvider struct {
	config  *ProviderConfig
	metrics *MultiProviderMetrics
}

func (p *MockProvider) GetName() string                    { return p.config.Name }
func (p *MockProvider) GetType() string                    { return p.config.Type }
func (p *MockProvider) GetVersion() string                 { return p.config.Version }
func (p *MockProvider) GetCapabilities() []string          { return p.config.Capabilities }
func (p *MockProvider) GetSupportedModels() []string       { return p.config.Models }
func (p *MockProvider) IsHealthy(ctx context.Context) bool { return p.config.Enabled }

func (p *MockProvider) ProcessRequest(ctx context.Context, request *ProviderRequest) (*ProviderResponse, error) {
	if !p.config.Enabled {
		return nil, fmt.Errorf("provider %s is not enabled", p.config.Name)
	}

	return &ProviderResponse{
		ID:             request.ID,
		ProviderName:   p.config.Name,
		Model:          request.Model,
		Output:         fmt.Sprintf("Mock response from %s", p.config.Name),
		Confidence:     0.95,
		TokensUsed:     100,
		ProcessingTime: 150 * time.Millisecond,
		Cost:           0.01,
		Success:        true,
		Timestamp:      time.Now(),
	}, nil
}

func (p *MockProvider) ValidateRequest(request *ProviderRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if request.Input == "" {
		return fmt.Errorf("request input cannot be empty")
	}
	return nil
}

func (p *MockProvider) EstimateCost(request *ProviderRequest) (*CostEstimate, error) {
	return &CostEstimate{
		ProviderName:    p.config.Name,
		Model:           request.Model,
		EstimatedTokens: 100,
		TokenCost:       0.008,
		RequestCost:     0.002,
		TotalCost:       0.01,
		Currency:        "USD",
		Confidence:      0.8,
	}, nil
}

func (p *MockProvider) Initialize(config *ProviderConfig) error {
	p.config = config
	return nil
}

func (p *MockProvider) Shutdown(ctx context.Context) error {
	return nil
}

func (p *MockProvider) GetConfiguration() *ProviderConfig {
	return p.config
}

func (p *MockProvider) UpdateConfiguration(config *ProviderConfig) error {
	p.config = config
	p.config.UpdatedAt = time.Now()
	return nil
}

func (p *MockProvider) GetMetrics(ctx context.Context) (*MultiProviderMetrics, error) {
	if p.metrics == nil {
		p.metrics = &MultiProviderMetrics{
			ProviderName:        p.config.Name,
			TotalRequests:       0,
			SuccessfulRequests:  0,
			FailedRequests:      0,
			AverageResponseTime: 150 * time.Millisecond,
			TokensProcessed:     0,
			TotalCost:           0,
			SuccessRate:         0,
			CurrentLoad:         0,
			MaxLoad:             100,
			HealthStatus:        "healthy",
			LastHealthCheck:     time.Now(),
			Timestamp:           time.Now(),
		}
	}
	return p.metrics, nil
}

func (p *MockProvider) GetStatus(ctx context.Context) *ProviderStatus {
	return &ProviderStatus{
		ProviderName:    p.config.Name,
		Type:            p.config.Type,
		Version:         p.config.Version,
		Enabled:         p.config.Enabled,
		Healthy:         p.config.Enabled,
		Available:       p.config.Enabled,
		Load:            0.0,
		ResponseTime:    150 * time.Millisecond,
		ErrorRate:       0.0,
		Capabilities:    p.config.Capabilities,
		SupportedModels: p.config.Models,
		LastSeen:        time.Now(),
		Details: map[string]interface{}{
			"base_url": p.config.BaseURL,
			"version":  p.config.Version,
		},
	}
}
