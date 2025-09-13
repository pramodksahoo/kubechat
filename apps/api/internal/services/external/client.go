package external

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/audit"
)

// APIClient represents a generic external API client
type APIClient interface {
	// Request makes an HTTP request to the external API
	Request(ctx context.Context, req *APIRequest) (*APIResponse, error)

	// HealthCheck validates API connectivity
	HealthCheck(ctx context.Context) error

	// GetProviderInfo returns information about the API provider
	GetProviderInfo() *ProviderInfo

	// Close gracefully shuts down the client
	Close() error
}

// APIRequest represents a generic request to external APIs
type APIRequest struct {
	Method     string            `json:"method"`
	Endpoint   string            `json:"endpoint"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Timeout    time.Duration     `json:"timeout"`
	RetryCount int               `json:"retry_count"`
	UserID     string            `json:"user_id,omitempty"`
	SessionID  string            `json:"session_id,omitempty"`
	RequestID  string            `json:"request_id,omitempty"`
}

// APIResponse represents a generic response from external APIs
type APIResponse struct {
	StatusCode   int               `json:"status_code"`
	Headers      map[string]string `json:"headers"`
	Body         []byte            `json:"body"`
	ResponseTime time.Duration     `json:"response_time"`
	RequestSize  int64             `json:"request_size"`
	ResponseSize int64             `json:"response_size"`
	TokensUsed   int               `json:"tokens_used,omitempty"`
	Cost         float64           `json:"cost,omitempty"`
	Error        string            `json:"error,omitempty"`
	Retries      int               `json:"retries"`
	CircuitState string            `json:"circuit_state,omitempty"`
	CacheHit     bool              `json:"cache_hit,omitempty"`
}

// ProviderInfo contains information about an API provider
type ProviderInfo struct {
	Name            string             `json:"name"`
	Type            models.NLPProvider `json:"type"`
	BaseURL         string             `json:"base_url"`
	Version         string             `json:"version"`
	RateLimitPerMin int                `json:"rate_limit_per_min"`
	MaxRetries      int                `json:"max_retries"`
	Timeout         time.Duration      `json:"timeout"`
	Features        []string           `json:"features"`
	Status          string             `json:"status"`
	LastHealthCheck time.Time          `json:"last_health_check"`
}

// ClientConfig represents configuration for API clients
type ClientConfig struct {
	// Authentication
	APIKey        string `json:"api_key"`
	SecretKeyName string `json:"secret_key_name"`
	Organization  string `json:"organization,omitempty"`

	// Connection settings
	BaseURL    string        `json:"base_url"`
	Timeout    time.Duration `json:"timeout"`
	MaxRetries int           `json:"max_retries"`

	// Rate limiting
	RateLimitPerMin int `json:"rate_limit_per_min"`
	RateLimitBurst  int `json:"rate_limit_burst"`

	// Circuit breaker
	FailureThreshold int           `json:"failure_threshold"`
	RecoveryTimeout  time.Duration `json:"recovery_timeout"`
	HalfOpenMaxCalls int           `json:"half_open_max_calls"`

	// Monitoring
	EnableMetrics  bool `json:"enable_metrics"`
	EnableAuditLog bool `json:"enable_audit_log"`

	// Advanced settings
	UserAgent     string            `json:"user_agent"`
	CustomHeaders map[string]string `json:"custom_headers"`
	ProxyURL      string            `json:"proxy_url,omitempty"`
}

// ClientManager manages multiple API clients
type ClientManager struct {
	clients     map[string]APIClient
	configs     map[string]*ClientConfig
	auditSvc    audit.Service
	mu          sync.RWMutex
	healthCheck *time.Ticker
	metrics     *ClientMetrics
}

// ClientMetrics tracks metrics for API clients
type ClientMetrics struct {
	RequestCount  map[string]int64     `json:"request_count"`
	SuccessCount  map[string]int64     `json:"success_count"`
	ErrorCount    map[string]int64     `json:"error_count"`
	ResponseTimes map[string][]float64 `json:"response_times"`
	TokensUsed    map[string]int64     `json:"tokens_used"`
	CostsAccrued  map[string]float64   `json:"costs_accrued"`
	CircuitStates map[string]string    `json:"circuit_states"`
	LastUpdated   time.Time            `json:"last_updated"`
	mu            sync.RWMutex
}

// NewClientManager creates a new client manager
func NewClientManager(auditSvc audit.Service) *ClientManager {
	cm := &ClientManager{
		clients:  make(map[string]APIClient),
		configs:  make(map[string]*ClientConfig),
		auditSvc: auditSvc,
		metrics: &ClientMetrics{
			RequestCount:  make(map[string]int64),
			SuccessCount:  make(map[string]int64),
			ErrorCount:    make(map[string]int64),
			ResponseTimes: make(map[string][]float64),
			TokensUsed:    make(map[string]int64),
			CostsAccrued:  make(map[string]float64),
			CircuitStates: make(map[string]string),
			LastUpdated:   time.Now(),
		},
	}

	// Start periodic health checks
	cm.healthCheck = time.NewTicker(5 * time.Minute)
	go cm.runHealthChecks()

	log.Println("External API Client Manager initialized")
	return cm
}

// RegisterClient registers a new API client
func (cm *ClientManager) RegisterClient(name string, client APIClient, config *ClientConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.clients[name]; exists {
		return fmt.Errorf("client %s already registered", name)
	}

	cm.clients[name] = client
	cm.configs[name] = config

	// Initialize metrics for this client
	cm.metrics.mu.Lock()
	cm.metrics.RequestCount[name] = 0
	cm.metrics.SuccessCount[name] = 0
	cm.metrics.ErrorCount[name] = 0
	cm.metrics.ResponseTimes[name] = make([]float64, 0)
	cm.metrics.TokensUsed[name] = 0
	cm.metrics.CostsAccrued[name] = 0.0
	cm.metrics.CircuitStates[name] = "closed"
	cm.metrics.mu.Unlock()

	log.Printf("Registered external API client: %s", name)
	return nil
}

// GetClient retrieves a registered API client
func (cm *ClientManager) GetClient(name string) (APIClient, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.clients[name]
	if !exists {
		return nil, fmt.Errorf("client %s not found", name)
	}

	return client, nil
}

// ListClients returns a list of registered client names
func (cm *ClientManager) ListClients() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	names := make([]string, 0, len(cm.clients))
	for name := range cm.clients {
		names = append(names, name)
	}

	return names
}

// UpdateMetrics updates metrics for a client after a request
func (cm *ClientManager) UpdateMetrics(clientName string, resp *APIResponse, success bool) {
	cm.metrics.mu.Lock()
	defer cm.metrics.mu.Unlock()

	cm.metrics.RequestCount[clientName]++
	if success {
		cm.metrics.SuccessCount[clientName]++
	} else {
		cm.metrics.ErrorCount[clientName]++
	}

	if resp != nil {
		responseTimeMs := float64(resp.ResponseTime.Milliseconds())
		cm.metrics.ResponseTimes[clientName] = append(cm.metrics.ResponseTimes[clientName], responseTimeMs)

		// Keep only the last 100 response times to prevent memory growth
		if len(cm.metrics.ResponseTimes[clientName]) > 100 {
			cm.metrics.ResponseTimes[clientName] = cm.metrics.ResponseTimes[clientName][1:]
		}

		cm.metrics.TokensUsed[clientName] += int64(resp.TokensUsed)
		cm.metrics.CostsAccrued[clientName] += resp.Cost

		if resp.CircuitState != "" {
			cm.metrics.CircuitStates[clientName] = resp.CircuitState
		}
	}

	cm.metrics.LastUpdated = time.Now()
}

// GetMetrics returns current metrics for all clients
func (cm *ClientManager) GetMetrics() *ClientMetrics {
	cm.metrics.mu.RLock()
	defer cm.metrics.mu.RUnlock()

	// Create a deep copy to avoid race conditions
	metrics := &ClientMetrics{
		RequestCount:  make(map[string]int64),
		SuccessCount:  make(map[string]int64),
		ErrorCount:    make(map[string]int64),
		ResponseTimes: make(map[string][]float64),
		TokensUsed:    make(map[string]int64),
		CostsAccrued:  make(map[string]float64),
		CircuitStates: make(map[string]string),
		LastUpdated:   cm.metrics.LastUpdated,
	}

	for k, v := range cm.metrics.RequestCount {
		metrics.RequestCount[k] = v
	}
	for k, v := range cm.metrics.SuccessCount {
		metrics.SuccessCount[k] = v
	}
	for k, v := range cm.metrics.ErrorCount {
		metrics.ErrorCount[k] = v
	}
	for k, v := range cm.metrics.TokensUsed {
		metrics.TokensUsed[k] = v
	}
	for k, v := range cm.metrics.CostsAccrued {
		metrics.CostsAccrued[k] = v
	}
	for k, v := range cm.metrics.CircuitStates {
		metrics.CircuitStates[k] = v
	}
	for k, v := range cm.metrics.ResponseTimes {
		metrics.ResponseTimes[k] = make([]float64, len(v))
		copy(metrics.ResponseTimes[k], v)
	}

	return metrics
}

// runHealthChecks performs periodic health checks on all clients
func (cm *ClientManager) runHealthChecks() {
	for range cm.healthCheck.C {
		cm.performHealthChecks()
	}
}

// performHealthChecks checks health of all registered clients
func (cm *ClientManager) performHealthChecks() {
	cm.mu.RLock()
	clients := make(map[string]APIClient)
	for name, client := range cm.clients {
		clients[name] = client
	}
	cm.mu.RUnlock()

	for name, client := range clients {
		go func(clientName string, c APIClient) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := c.HealthCheck(ctx); err != nil {
				log.Printf("Health check failed for client %s: %v", clientName, err)
				cm.metrics.mu.Lock()
				cm.metrics.CircuitStates[clientName] = "open"
				cm.metrics.mu.Unlock()
			} else {
				cm.metrics.mu.Lock()
				if cm.metrics.CircuitStates[clientName] == "open" {
					cm.metrics.CircuitStates[clientName] = "half-open"
				} else {
					cm.metrics.CircuitStates[clientName] = "closed"
				}
				cm.metrics.mu.Unlock()
			}
		}(name, client)
	}
}

// GetClientHealth returns health status for a specific client
func (cm *ClientManager) GetClientHealth(clientName string) (bool, error) {
	client, err := cm.GetClient(clientName)
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.HealthCheck(ctx); err != nil {
		return false, err
	}

	return true, nil
}

// Close gracefully shuts down the client manager
func (cm *ClientManager) Close() error {
	if cm.healthCheck != nil {
		cm.healthCheck.Stop()
	}

	cm.mu.RLock()
	clients := make([]APIClient, 0, len(cm.clients))
	for _, client := range cm.clients {
		clients = append(clients, client)
	}
	cm.mu.RUnlock()

	for _, client := range clients {
		if err := client.Close(); err != nil {
			log.Printf("Error closing client: %v", err)
		}
	}

	log.Println("External API Client Manager shut down")
	return nil
}

// ValidateConfig validates client configuration
func ValidateConfig(config *ClientConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	if config.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}

	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	if config.MaxRetries < 0 {
		config.MaxRetries = 3
	}

	if config.RateLimitPerMin <= 0 {
		config.RateLimitPerMin = 60
	}

	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 5
	}

	if config.RecoveryTimeout <= 0 {
		config.RecoveryTimeout = 60 * time.Second
	}

	if config.HalfOpenMaxCalls <= 0 {
		config.HalfOpenMaxCalls = 3
	}

	return nil
}
