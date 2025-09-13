package communication

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// Service defines the service communication interface
type Service interface {
	// Service Discovery
	RegisterService(service *models.ServiceRegistration) error
	DeregisterService(serviceID string) error
	DiscoverService(serviceName string) ([]*models.ServiceInstance, error)
	GetHealthyServices(serviceName string) ([]*models.ServiceInstance, error)

	// Communication Patterns
	CallService(ctx context.Context, request *models.ServiceRequest) (*models.ServiceResponse, error)
	BroadcastMessage(ctx context.Context, message *models.BroadcastMessage) error
	PublishEvent(ctx context.Context, event *models.ServiceEvent) error
	SubscribeToEvents(eventType string, handler models.EventHandler) error

	// Circuit Breaker
	ExecuteWithCircuitBreaker(ctx context.Context, serviceName string, operation func() (*models.ServiceResponse, error)) (*models.ServiceResponse, error)

	// Load Balancing
	SelectInstance(serviceName string, strategy models.LoadBalancingStrategy) (*models.ServiceInstance, error)

	// Health and Metrics
	GetCommunicationMetrics() *models.CommunicationMetrics
	StartHealthChecking(ctx context.Context) error
	StopHealthChecking()
}

// Config represents the service communication configuration
type Config struct {
	// Service Discovery
	ServiceDiscoveryEnabled bool   `json:"service_discovery_enabled"`
	RegistryType            string `json:"registry_type"` // "memory", "kubernetes", "consul"

	// Circuit Breaker
	CircuitBreakerEnabled bool          `json:"circuit_breaker_enabled"`
	FailureThreshold      int           `json:"failure_threshold"`
	RecoveryTimeout       time.Duration `json:"recovery_timeout"`
	MaxConcurrentRequests int           `json:"max_concurrent_requests"`

	// Retry Policy
	RetryEnabled    bool          `json:"retry_enabled"`
	MaxRetries      int           `json:"max_retries"`
	RetryBackoff    time.Duration `json:"retry_backoff"`
	RetryMultiplier float64       `json:"retry_multiplier"`

	// Timeout Configuration
	RequestTimeout    time.Duration `json:"request_timeout"`
	ConnectionTimeout time.Duration `json:"connection_timeout"`
	KeepAliveTimeout  time.Duration `json:"keep_alive_timeout"`

	// Load Balancing
	DefaultLoadBalancer models.LoadBalancingStrategy `json:"default_load_balancer"`
	HealthCheckInterval time.Duration                `json:"health_check_interval"`
	UnhealthyThreshold  int                          `json:"unhealthy_threshold"`

	// Event System
	EventSystemEnabled    bool `json:"event_system_enabled"`
	EventBufferSize       int  `json:"event_buffer_size"`
	EventProcessorWorkers int  `json:"event_processor_workers"`

	// Monitoring
	EnableMetrics         bool          `json:"enable_metrics"`
	MetricsUpdateInterval time.Duration `json:"metrics_update_interval"`
}

// service implements the Service interface
type service struct {
	config           *Config
	serviceRegistry  map[string][]*models.ServiceInstance
	circuitBreakers  map[string]*models.ServiceCircuitBreaker
	eventSubscribers map[string][]models.EventHandler
	httpClient       *http.Client
	metrics          *models.CommunicationMetrics
	mutex            sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	healthChecker    *time.Ticker
}

// NewService creates a new service communication service
func NewService(config *Config) Service {
	if config == nil {
		config = &Config{
			ServiceDiscoveryEnabled: true,
			RegistryType:            "memory",
			CircuitBreakerEnabled:   true,
			FailureThreshold:        5,
			RecoveryTimeout:         30 * time.Second,
			MaxConcurrentRequests:   100,
			RetryEnabled:            true,
			MaxRetries:              3,
			RetryBackoff:            100 * time.Millisecond,
			RetryMultiplier:         2.0,
			RequestTimeout:          30 * time.Second,
			ConnectionTimeout:       5 * time.Second,
			KeepAliveTimeout:        30 * time.Second,
			DefaultLoadBalancer:     models.LoadBalancingRoundRobin,
			HealthCheckInterval:     30 * time.Second,
			UnhealthyThreshold:      3,
			EventSystemEnabled:      true,
			EventBufferSize:         1000,
			EventProcessorWorkers:   5,
			EnableMetrics:           true,
			MetricsUpdateInterval:   10 * time.Second,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create HTTP client with proper timeouts
	httpClient := &http.Client{
		Timeout: config.RequestTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxConnsPerHost:     10,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	svc := &service{
		config:           config,
		serviceRegistry:  make(map[string][]*models.ServiceInstance),
		circuitBreakers:  make(map[string]*models.ServiceCircuitBreaker),
		eventSubscribers: make(map[string][]models.EventHandler),
		httpClient:       httpClient,
		metrics: &models.CommunicationMetrics{
			TotalRequests:       0,
			SuccessfulRequests:  0,
			FailedRequests:      0,
			CircuitBreakerTrips: 0,
			RetryAttempts:       0,
			ServiceCallLatency:  make(map[string]time.Duration),
			ErrorRates:          make(map[string]float64),
			ActiveConnections:   0,
			StartTime:           time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize default services
	svc.initializeDefaultServices()

	return svc
}

// RegisterService registers a service instance
func (s *service) RegisterService(service *models.ServiceRegistration) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	instance := &models.ServiceInstance{
		ID:              service.ID,
		Name:            service.Name,
		Host:            service.Host,
		Port:            service.Port,
		Protocol:        service.Protocol,
		Health:          models.ServiceHealthHealthy,
		Metadata:        service.Metadata,
		Tags:            service.Tags,
		Version:         service.Version,
		RegisteredAt:    time.Now(),
		LastHealthCheck: time.Now(),
	}

	if s.serviceRegistry[service.Name] == nil {
		s.serviceRegistry[service.Name] = make([]*models.ServiceInstance, 0)
	}

	// Check if service already exists (update if so)
	for i, existing := range s.serviceRegistry[service.Name] {
		if existing.ID == service.ID {
			s.serviceRegistry[service.Name][i] = instance
			return nil
		}
	}

	// Add new service instance
	s.serviceRegistry[service.Name] = append(s.serviceRegistry[service.Name], instance)

	return nil
}

// DeregisterService removes a service instance
func (s *service) DeregisterService(serviceID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for serviceName, instances := range s.serviceRegistry {
		for i, instance := range instances {
			if instance.ID == serviceID {
				// Remove instance from slice
				s.serviceRegistry[serviceName] = append(instances[:i], instances[i+1:]...)
				return nil
			}
		}
	}

	return fmt.Errorf("service with ID %s not found", serviceID)
}

// DiscoverService returns all instances of a service
func (s *service) DiscoverService(serviceName string) ([]*models.ServiceInstance, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	instances, exists := s.serviceRegistry[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Return copy to avoid race conditions
	result := make([]*models.ServiceInstance, len(instances))
	copy(result, instances)

	return result, nil
}

// GetHealthyServices returns only healthy instances of a service
func (s *service) GetHealthyServices(serviceName string) ([]*models.ServiceInstance, error) {
	instances, err := s.DiscoverService(serviceName)
	if err != nil {
		return nil, err
	}

	var healthy []*models.ServiceInstance
	for _, instance := range instances {
		if instance.Health == models.ServiceHealthHealthy {
			healthy = append(healthy, instance)
		}
	}

	if len(healthy) == 0 {
		return nil, fmt.Errorf("no healthy instances found for service %s", serviceName)
	}

	return healthy, nil
}

// CallService makes a service call with retry and circuit breaker patterns
func (s *service) CallService(ctx context.Context, request *models.ServiceRequest) (*models.ServiceResponse, error) {
	startTime := time.Now()
	s.metrics.TotalRequests++

	// Get target service instance
	instance, err := s.SelectInstance(request.ServiceName, s.config.DefaultLoadBalancer)
	if err != nil {
		s.metrics.FailedRequests++
		return nil, fmt.Errorf("failed to select service instance: %w", err)
	}

	// Execute with circuit breaker
	response, err := s.ExecuteWithCircuitBreaker(ctx, request.ServiceName, func() (*models.ServiceResponse, error) {
		return s.executeServiceCall(ctx, instance, request)
	})

	// Update metrics
	duration := time.Since(startTime)
	s.updateServiceMetrics(request.ServiceName, duration, err == nil)

	if err != nil {
		s.metrics.FailedRequests++
		return nil, err
	}

	s.metrics.SuccessfulRequests++
	return response, nil
}

// BroadcastMessage sends a message to all instances of a service
func (s *service) BroadcastMessage(ctx context.Context, message *models.BroadcastMessage) error {
	instances, err := s.GetHealthyServices(message.ServiceName)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errors := make(chan error, len(instances))

	for _, instance := range instances {
		wg.Add(1)
		go func(inst *models.ServiceInstance) {
			defer wg.Done()

			request := &models.ServiceRequest{
				ServiceName: message.ServiceName,
				Method:      "POST",
				Path:        message.Path,
				Body:        message.Payload,
				Headers:     message.Headers,
				Timeout:     message.Timeout,
			}

			_, err := s.executeServiceCall(ctx, inst, request)
			if err != nil {
				errors <- err
			}
		}(instance)
	}

	wg.Wait()
	close(errors)

	// Collect errors
	var broadcastErrors []error
	for err := range errors {
		broadcastErrors = append(broadcastErrors, err)
	}

	if len(broadcastErrors) > 0 {
		return fmt.Errorf("broadcast failed to %d instances: %v", len(broadcastErrors), broadcastErrors)
	}

	return nil
}

// PublishEvent publishes an event to subscribers
func (s *service) PublishEvent(ctx context.Context, event *models.ServiceEvent) error {
	s.mutex.RLock()
	handlers, exists := s.eventSubscribers[event.Type]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("no subscribers for event type %s", event.Type)
	}

	// Publish to all handlers concurrently
	var wg sync.WaitGroup
	for _, handler := range handlers {
		wg.Add(1)
		go func(h models.EventHandler) {
			defer wg.Done()
			if err := h(ctx, event); err != nil {
				fmt.Printf("Event handler error for type %s: %v\n", event.Type, err)
			}
		}(handler)
	}

	wg.Wait()
	return nil
}

// SubscribeToEvents subscribes to events of a specific type
func (s *service) SubscribeToEvents(eventType string, handler models.EventHandler) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.eventSubscribers[eventType] == nil {
		s.eventSubscribers[eventType] = make([]models.EventHandler, 0)
	}

	s.eventSubscribers[eventType] = append(s.eventSubscribers[eventType], handler)
	return nil
}

// ExecuteWithCircuitBreaker executes an operation with circuit breaker protection
func (s *service) ExecuteWithCircuitBreaker(ctx context.Context, serviceName string, operation func() (*models.ServiceResponse, error)) (*models.ServiceResponse, error) {
	if !s.config.CircuitBreakerEnabled {
		return operation()
	}

	breaker := s.getCircuitBreaker(serviceName)

	if !breaker.CanRequest() {
		s.metrics.CircuitBreakerTrips++
		return nil, fmt.Errorf("circuit breaker is open for service %s", serviceName)
	}

	response, err := operation()

	if err != nil {
		breaker.RecordFailure()
	} else {
		breaker.RecordSuccess()
	}

	return response, err
}

// SelectInstance selects a service instance based on load balancing strategy
func (s *service) SelectInstance(serviceName string, strategy models.LoadBalancingStrategy) (*models.ServiceInstance, error) {
	instances, err := s.GetHealthyServices(serviceName)
	if err != nil {
		return nil, err
	}

	switch strategy {
	case models.LoadBalancingRoundRobin:
		return s.selectRoundRobin(serviceName, instances), nil
	case models.LoadBalancingRandom:
		return s.selectRandom(instances), nil
	case models.LoadBalancingLeastConnections:
		return s.selectLeastConnections(instances), nil
	default:
		return instances[0], nil
	}
}

// GetCommunicationMetrics returns current communication metrics
func (s *service) GetCommunicationMetrics() *models.CommunicationMetrics {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Create a copy to avoid race conditions
	metrics := *s.metrics
	metrics.ServiceCallLatency = make(map[string]time.Duration)
	metrics.ErrorRates = make(map[string]float64)

	for k, v := range s.metrics.ServiceCallLatency {
		metrics.ServiceCallLatency[k] = v
	}
	for k, v := range s.metrics.ErrorRates {
		metrics.ErrorRates[k] = v
	}

	metrics.Uptime = time.Since(s.metrics.StartTime)

	return &metrics
}

// StartHealthChecking starts background health checking
func (s *service) StartHealthChecking(ctx context.Context) error {
	s.healthChecker = time.NewTicker(s.config.HealthCheckInterval)

	go s.healthCheckLoop()

	return nil
}

// StopHealthChecking stops background health checking
func (s *service) StopHealthChecking() {
	if s.healthChecker != nil {
		s.healthChecker.Stop()
	}
	s.cancel()
}

// Helper methods

func (s *service) initializeDefaultServices() {
	// Register internal services
	services := []*models.ServiceRegistration{
		{
			ID:       "kubechat-api-1",
			Name:     "kubechat-api",
			Host:     "localhost",
			Port:     8080,
			Protocol: "http",
			Version:  "1.0.0",
			Tags:     []string{"api", "gateway"},
			Metadata: map[string]string{
				"environment": "development",
			},
		},
		{
			ID:       "kubechat-ollama-1",
			Name:     "kubechat-ollama",
			Host:     "kubechat-dev-ollama",
			Port:     11434,
			Protocol: "http",
			Version:  "0.1.48",
			Tags:     []string{"nlp", "ai"},
			Metadata: map[string]string{
				"model": "llama3.2:3b",
			},
		},
		{
			ID:       "kubechat-db-1",
			Name:     "kubechat-database",
			Host:     "kubechat-dev-postgresql",
			Port:     5432,
			Protocol: "tcp",
			Version:  "16",
			Tags:     []string{"database", "postgres"},
			Metadata: map[string]string{
				"database": "kubechat",
			},
		},
	}

	for _, service := range services {
		_ = s.RegisterService(service)
	}
}

func (s *service) executeServiceCall(ctx context.Context, instance *models.ServiceInstance, request *models.ServiceRequest) (*models.ServiceResponse, error) {
	url := fmt.Sprintf("%s://%s:%d%s", instance.Protocol, instance.Host, instance.Port, request.Path)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, request.Method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range request.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request with retry
	return s.executeWithRetry(ctx, httpReq)
}

func (s *service) executeWithRetry(ctx context.Context, req *http.Request) (*models.ServiceResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			s.metrics.RetryAttempts++
			// Calculate backoff delay
			delay := time.Duration(float64(s.config.RetryBackoff) * float64(attempt) * s.config.RetryMultiplier)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		resp, err := s.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		// Check if we should retry based on status code
		if resp.StatusCode >= 500 && attempt < s.config.MaxRetries {
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}

		return &models.ServiceResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       nil, // Would read body if needed
		}, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", s.config.MaxRetries, lastErr)
}

func (s *service) getCircuitBreaker(serviceName string) *models.ServiceCircuitBreaker {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if breaker, exists := s.circuitBreakers[serviceName]; exists {
		return breaker
	}

	breaker := &models.ServiceCircuitBreaker{
		ServiceName:   serviceName,
		State:         models.CircuitBreakerClosed,
		FailureCount:  0,
		LastFailure:   time.Time{},
		NextAttempt:   time.Time{},
		SuccessCount:  0,
		TotalRequests: 0,
	}

	s.circuitBreakers[serviceName] = breaker
	return breaker
}

func (s *service) selectRoundRobin(serviceName string, instances []*models.ServiceInstance) *models.ServiceInstance {
	// Simple round-robin implementation
	// In production, you'd want to track state per service
	index := int(time.Now().UnixNano()) % len(instances)
	return instances[index]
}

func (s *service) selectRandom(instances []*models.ServiceInstance) *models.ServiceInstance {
	index := int(time.Now().UnixNano()) % len(instances)
	return instances[index]
}

func (s *service) selectLeastConnections(instances []*models.ServiceInstance) *models.ServiceInstance {
	// For now, just return the first instance
	// In production, you'd track active connections per instance
	return instances[0]
}

func (s *service) updateServiceMetrics(serviceName string, duration time.Duration, success bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Update latency (simple moving average)
	if currentLatency, exists := s.metrics.ServiceCallLatency[serviceName]; exists {
		s.metrics.ServiceCallLatency[serviceName] = time.Duration(
			(float64(currentLatency) + float64(duration)) / 2,
		)
	} else {
		s.metrics.ServiceCallLatency[serviceName] = duration
	}

	// Update error rates would require more sophisticated tracking
	// For now, just set based on recent success
	if success {
		s.metrics.ErrorRates[serviceName] = 0.0
	} else {
		s.metrics.ErrorRates[serviceName] = 1.0
	}
}

func (s *service) healthCheckLoop() {
	for {
		select {
		case <-s.healthChecker.C:
			s.performHealthChecks()
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *service) performHealthChecks() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, instances := range s.serviceRegistry {
		for _, instance := range instances {
			// Simple health check - in production you'd make actual health requests
			instance.LastHealthCheck = time.Now()
			// For now, assume all services are healthy
			instance.Health = models.ServiceHealthHealthy
		}
	}
}
