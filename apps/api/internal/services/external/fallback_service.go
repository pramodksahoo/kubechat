package external

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// FallbackService provides fallback mechanisms when external APIs are unavailable
type FallbackService interface {
	// ExecuteWithFallback executes a primary function with fallback options
	ExecuteWithFallback(ctx context.Context, req *FallbackRequest) (*FallbackResponse, error)

	// RegisterFallbackChain registers a fallback chain for a specific service
	RegisterFallbackChain(serviceName string, chain *FallbackChain) error

	// GetFallbackChain returns the fallback chain for a service
	GetFallbackChain(serviceName string) *FallbackChain

	// SetMockMode enables/disables mock responses for development/testing
	SetMockMode(enabled bool)

	// GetFallbackMetrics returns metrics about fallback usage
	GetFallbackMetrics() *FallbackMetrics
}

// FallbackRequest contains the request for fallback execution
type FallbackRequest struct {
	ServiceName string                      `json:"service_name"`
	Operation   string                      `json:"operation"`
	PrimaryFunc func() (interface{}, error) `json:"-"` // Primary function to execute
	Context     map[string]interface{}      `json:"context,omitempty"`
	Timeout     time.Duration               `json:"timeout,omitempty"`
	UserID      string                      `json:"user_id,omitempty"`
	SessionID   string                      `json:"session_id,omitempty"`
}

// FallbackResponse contains the response from fallback execution
type FallbackResponse struct {
	Data          interface{}            `json:"data"`
	Source        FallbackSource         `json:"source"`
	ExecutionTime time.Duration          `json:"execution_time"`
	FallbackUsed  bool                   `json:"fallback_used"`
	FallbackLevel int                    `json:"fallback_level"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// FallbackSource indicates the source of the response
type FallbackSource int

const (
	SourcePrimary FallbackSource = iota
	SourceSecondary
	SourceCache
	SourceMock
	SourceDefault
	SourceDegraded
)

func (s FallbackSource) String() string {
	switch s {
	case SourcePrimary:
		return "primary"
	case SourceSecondary:
		return "secondary"
	case SourceCache:
		return "cache"
	case SourceMock:
		return "mock"
	case SourceDefault:
		return "default"
	case SourceDegraded:
		return "degraded"
	default:
		return "unknown"
	}
}

// FallbackChain defines the fallback strategy for a service
type FallbackChain struct {
	ServiceName     string           `json:"service_name"`
	PrimaryProvider string           `json:"primary_provider"`
	Fallbacks       []FallbackOption `json:"fallbacks"`
	DefaultResponse interface{}      `json:"default_response,omitempty"`
	CacheEnabled    bool             `json:"cache_enabled"`
	CacheTTL        time.Duration    `json:"cache_ttl"`
	MockEnabled     bool             `json:"mock_enabled"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// FallbackOption represents a single fallback option in the chain
type FallbackOption struct {
	Type        FallbackType                `json:"type"`
	Provider    string                      `json:"provider,omitempty"`
	Endpoint    string                      `json:"endpoint,omitempty"`
	Config      map[string]interface{}      `json:"config,omitempty"`
	Function    func() (interface{}, error) `json:"-"` // Custom fallback function
	Timeout     time.Duration               `json:"timeout"`
	Priority    int                         `json:"priority"` // Lower number = higher priority
	Enabled     bool                        `json:"enabled"`
	Description string                      `json:"description,omitempty"`
}

// FallbackType represents the type of fallback
type FallbackType int

const (
	FallbackTypeSecondaryAPI FallbackType = iota
	FallbackTypeCache
	FallbackTypeMock
	FallbackTypeDefault
	FallbackTypeCustom
	FallbackTypeDegraded
)

func (t FallbackType) String() string {
	switch t {
	case FallbackTypeSecondaryAPI:
		return "secondary_api"
	case FallbackTypeCache:
		return "cache"
	case FallbackTypeMock:
		return "mock"
	case FallbackTypeDefault:
		return "default"
	case FallbackTypeCustom:
		return "custom"
	case FallbackTypeDegraded:
		return "degraded"
	default:
		return "unknown"
	}
}

// FallbackMetrics contains metrics about fallback usage
type FallbackMetrics struct {
	TotalRequests       int64                      `json:"total_requests"`
	PrimarySuccesses    int64                      `json:"primary_successes"`
	FallbacksTriggered  int64                      `json:"fallbacks_triggered"`
	FallbacksBySource   map[string]int64           `json:"fallbacks_by_source"`
	FallbacksByService  map[string]int64           `json:"fallbacks_by_service"`
	AverageResponseTime time.Duration              `json:"average_response_time"`
	ServiceMetrics      map[string]*ServiceMetrics `json:"service_metrics"`
	LastUpdated         time.Time                  `json:"last_updated"`
}

// ServiceMetrics contains metrics for a specific service
type ServiceMetrics struct {
	ServiceName      string        `json:"service_name"`
	TotalRequests    int64         `json:"total_requests"`
	PrimarySuccesses int64         `json:"primary_successes"`
	FallbacksUsed    int64         `json:"fallbacks_used"`
	AverageResponse  time.Duration `json:"average_response_time"`
	LastFallbackUsed time.Time     `json:"last_fallback_used,omitempty"`
	SuccessRate      float64       `json:"success_rate"`
}

// fallbackServiceImpl implements FallbackService
type fallbackServiceImpl struct {
	chains   map[string]*FallbackChain
	metrics  *FallbackMetrics
	mockMode bool
	mu       sync.RWMutex
}

// NewFallbackService creates a new fallback service
func NewFallbackService() FallbackService {
	return &fallbackServiceImpl{
		chains: make(map[string]*FallbackChain),
		metrics: &FallbackMetrics{
			FallbacksBySource:  make(map[string]int64),
			FallbacksByService: make(map[string]int64),
			ServiceMetrics:     make(map[string]*ServiceMetrics),
			LastUpdated:        time.Now(),
		},
	}
}

// ExecuteWithFallback executes a primary function with fallback options
func (s *fallbackServiceImpl) ExecuteWithFallback(ctx context.Context, req *FallbackRequest) (*FallbackResponse, error) {
	startTime := time.Now()
	s.updateRequestMetrics(req.ServiceName)

	// Set default timeout if not provided
	if req.Timeout == 0 {
		req.Timeout = 30 * time.Second
	}

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	// Try primary function first
	if req.PrimaryFunc != nil {
		select {
		case <-ctxWithTimeout.Done():
			return s.executeFallbacks(ctx, req, startTime, fmt.Errorf("primary request timeout"))
		default:
			primaryResult, primaryErr := s.executePrimary(req.PrimaryFunc)
			if primaryErr == nil {
				s.updateSuccessMetrics(req.ServiceName, time.Since(startTime))
				return &FallbackResponse{
					Data:          primaryResult,
					Source:        SourcePrimary,
					ExecutionTime: time.Since(startTime),
					FallbackUsed:  false,
					FallbackLevel: 0,
					Metadata: map[string]interface{}{
						"service":   req.ServiceName,
						"operation": req.Operation,
					},
				}, nil
			}

			log.Printf("Primary function failed for %s: %v", req.ServiceName, primaryErr)
			return s.executeFallbacks(ctx, req, startTime, primaryErr)
		}
	}

	// No primary function provided, go directly to fallbacks
	return s.executeFallbacks(ctx, req, startTime, fmt.Errorf("no primary function provided"))
}

// executePrimary executes the primary function with error handling
func (s *fallbackServiceImpl) executePrimary(primaryFunc func() (interface{}, error)) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Primary function panicked: %v", r)
		}
	}()

	return primaryFunc()
}

// executeFallbacks executes the fallback chain
func (s *fallbackServiceImpl) executeFallbacks(ctx context.Context, req *FallbackRequest, startTime time.Time, primaryErr error) (*FallbackResponse, error) {
	chain := s.GetFallbackChain(req.ServiceName)
	if chain == nil {
		return s.createDefaultResponse(req, startTime, primaryErr)
	}

	s.updateFallbackMetrics(req.ServiceName)

	// Check if mock mode is enabled
	if s.mockMode || chain.MockEnabled {
		if mockResponse := s.createMockResponse(req); mockResponse != nil {
			return &FallbackResponse{
				Data:          mockResponse,
				Source:        SourceMock,
				ExecutionTime: time.Since(startTime),
				FallbackUsed:  true,
				FallbackLevel: 0,
				Metadata: map[string]interface{}{
					"service":       req.ServiceName,
					"operation":     req.Operation,
					"primary_error": primaryErr.Error(),
					"mock_mode":     true,
				},
			}, nil
		}
	}

	// Try fallback options in order of priority
	for level, fallback := range chain.Fallbacks {
		if !fallback.Enabled {
			continue
		}

		result, err := s.executeFallbackOption(ctx, &fallback, req)
		if err == nil {
			source := s.getFallbackSource(fallback.Type)
			s.updateFallbackSourceMetrics(source.String())

			return &FallbackResponse{
				Data:          result,
				Source:        source,
				ExecutionTime: time.Since(startTime),
				FallbackUsed:  true,
				FallbackLevel: level + 1,
				Metadata: map[string]interface{}{
					"service":           req.ServiceName,
					"operation":         req.Operation,
					"primary_error":     primaryErr.Error(),
					"fallback_type":     fallback.Type.String(),
					"fallback_provider": fallback.Provider,
				},
			}, nil
		}

		log.Printf("Fallback level %d failed for %s: %v", level+1, req.ServiceName, err)
	}

	// All fallbacks failed, return default response
	return s.createDefaultResponse(req, startTime, primaryErr)
}

// executeFallbackOption executes a specific fallback option
func (s *fallbackServiceImpl) executeFallbackOption(ctx context.Context, option *FallbackOption, req *FallbackRequest) (interface{}, error) {
	if option.Function != nil {
		return option.Function()
	}

	switch option.Type {
	case FallbackTypeCache:
		return s.executeCache(option, req)
	case FallbackTypeMock:
		return s.createMockResponse(req), nil
	case FallbackTypeDefault:
		chain := s.GetFallbackChain(req.ServiceName)
		if chain != nil && chain.DefaultResponse != nil {
			return chain.DefaultResponse, nil
		}
		return nil, fmt.Errorf("no default response configured")
	case FallbackTypeDegraded:
		return s.createDegradedResponse(req), nil
	default:
		return nil, fmt.Errorf("unsupported fallback type: %s", option.Type.String())
	}
}

// executeCache attempts to retrieve cached response
func (s *fallbackServiceImpl) executeCache(option *FallbackOption, req *FallbackRequest) (interface{}, error) {
	// Mock cache implementation - in production, this would integrate with Redis or similar
	cacheKey := fmt.Sprintf("fallback:%s:%s", req.ServiceName, req.Operation)

	// Simulate cache miss for now
	return nil, fmt.Errorf("cache miss for key: %s", cacheKey)
}

// createMockResponse creates a mock response based on the service and operation
func (s *fallbackServiceImpl) createMockResponse(req *FallbackRequest) interface{} {
	switch req.ServiceName {
	case "openai":
		return map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "This is a mock response from the OpenAI fallback system. The primary API is currently unavailable.",
					},
					"finish_reason": "mock_fallback",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
			"model": "mock-gpt-3.5-turbo",
			"mock":  true,
		}

	case "ollama":
		return map[string]interface{}{
			"model":    "mock-llama3.2",
			"response": "This is a mock response from the Ollama fallback system. The primary service is currently unavailable.",
			"done":     true,
			"mock":     true,
		}

	default:
		return map[string]interface{}{
			"message":   fmt.Sprintf("Mock response for %s service", req.ServiceName),
			"status":    "fallback_active",
			"mock":      true,
			"service":   req.ServiceName,
			"operation": req.Operation,
		}
	}
}

// createDegradedResponse creates a degraded response with limited functionality
func (s *fallbackServiceImpl) createDegradedResponse(req *FallbackRequest) interface{} {
	return map[string]interface{}{
		"status":                "degraded",
		"message":               fmt.Sprintf("Service %s is running in degraded mode", req.ServiceName),
		"service":               req.ServiceName,
		"operation":             req.Operation,
		"limited_functionality": true,
		"recommendation":        "Please try again later for full functionality",
	}
}

// createDefaultResponse creates a default error response
func (s *fallbackServiceImpl) createDefaultResponse(req *FallbackRequest, startTime time.Time, primaryErr error) (*FallbackResponse, error) {
	defaultData := map[string]interface{}{
		"error":       "Service temporarily unavailable",
		"message":     "All fallback mechanisms have been exhausted",
		"service":     req.ServiceName,
		"operation":   req.Operation,
		"retry_after": 60, // seconds
	}

	return &FallbackResponse{
		Data:          defaultData,
		Source:        SourceDefault,
		ExecutionTime: time.Since(startTime),
		FallbackUsed:  true,
		FallbackLevel: -1, // Indicates default response
		ErrorMessage:  primaryErr.Error(),
		Metadata: map[string]interface{}{
			"all_fallbacks_exhausted": true,
			"primary_error":           primaryErr.Error(),
		},
	}, nil
}

// getFallbackSource maps fallback type to source
func (s *fallbackServiceImpl) getFallbackSource(fallbackType FallbackType) FallbackSource {
	switch fallbackType {
	case FallbackTypeSecondaryAPI:
		return SourceSecondary
	case FallbackTypeCache:
		return SourceCache
	case FallbackTypeMock:
		return SourceMock
	case FallbackTypeDefault:
		return SourceDefault
	case FallbackTypeDegraded:
		return SourceDegraded
	default:
		return SourceDefault
	}
}

// RegisterFallbackChain registers a fallback chain for a specific service
func (s *fallbackServiceImpl) RegisterFallbackChain(serviceName string, chain *FallbackChain) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if chain == nil {
		return fmt.Errorf("fallback chain cannot be nil")
	}

	chain.ServiceName = serviceName
	chain.UpdatedAt = time.Now()
	if chain.CreatedAt.IsZero() {
		chain.CreatedAt = time.Now()
	}

	s.chains[serviceName] = chain

	// Initialize metrics for this service
	if _, exists := s.metrics.ServiceMetrics[serviceName]; !exists {
		s.metrics.ServiceMetrics[serviceName] = &ServiceMetrics{
			ServiceName: serviceName,
		}
	}

	log.Printf("Registered fallback chain for service: %s with %d fallback options", serviceName, len(chain.Fallbacks))
	return nil
}

// GetFallbackChain returns the fallback chain for a service
func (s *fallbackServiceImpl) GetFallbackChain(serviceName string) *FallbackChain {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.chains[serviceName]
}

// SetMockMode enables/disables mock responses for development/testing
func (s *fallbackServiceImpl) SetMockMode(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mockMode = enabled
	log.Printf("Fallback mock mode set to: %v", enabled)
}

// GetFallbackMetrics returns metrics about fallback usage
func (s *fallbackServiceImpl) GetFallbackMetrics() *FallbackMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a deep copy to avoid race conditions
	metrics := &FallbackMetrics{
		TotalRequests:       s.metrics.TotalRequests,
		PrimarySuccesses:    s.metrics.PrimarySuccesses,
		FallbacksTriggered:  s.metrics.FallbacksTriggered,
		FallbacksBySource:   make(map[string]int64),
		FallbacksByService:  make(map[string]int64),
		AverageResponseTime: s.metrics.AverageResponseTime,
		ServiceMetrics:      make(map[string]*ServiceMetrics),
		LastUpdated:         time.Now(),
	}

	for k, v := range s.metrics.FallbacksBySource {
		metrics.FallbacksBySource[k] = v
	}
	for k, v := range s.metrics.FallbacksByService {
		metrics.FallbacksByService[k] = v
	}
	for k, v := range s.metrics.ServiceMetrics {
		metrics.ServiceMetrics[k] = &ServiceMetrics{
			ServiceName:      v.ServiceName,
			TotalRequests:    v.TotalRequests,
			PrimarySuccesses: v.PrimarySuccesses,
			FallbacksUsed:    v.FallbacksUsed,
			AverageResponse:  v.AverageResponse,
			LastFallbackUsed: v.LastFallbackUsed,
			SuccessRate:      v.SuccessRate,
		}
	}

	return metrics
}

// Metrics update methods

func (s *fallbackServiceImpl) updateRequestMetrics(serviceName string) {
	s.metrics.TotalRequests++
	if serviceMetrics, exists := s.metrics.ServiceMetrics[serviceName]; exists {
		serviceMetrics.TotalRequests++
	}
}

func (s *fallbackServiceImpl) updateSuccessMetrics(serviceName string, responseTime time.Duration) {
	s.metrics.PrimarySuccesses++
	if serviceMetrics, exists := s.metrics.ServiceMetrics[serviceName]; exists {
		serviceMetrics.PrimarySuccesses++
		serviceMetrics.SuccessRate = float64(serviceMetrics.PrimarySuccesses) / float64(serviceMetrics.TotalRequests)
	}
}

func (s *fallbackServiceImpl) updateFallbackMetrics(serviceName string) {
	s.metrics.FallbacksTriggered++
	s.metrics.FallbacksByService[serviceName]++

	if serviceMetrics, exists := s.metrics.ServiceMetrics[serviceName]; exists {
		serviceMetrics.FallbacksUsed++
		serviceMetrics.LastFallbackUsed = time.Now()
		serviceMetrics.SuccessRate = float64(serviceMetrics.PrimarySuccesses) / float64(serviceMetrics.TotalRequests)
	}
}

func (s *fallbackServiceImpl) updateFallbackSourceMetrics(source string) {
	s.metrics.FallbacksBySource[source]++
	s.metrics.LastUpdated = time.Now()
}
