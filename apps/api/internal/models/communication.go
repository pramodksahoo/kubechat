package models

import (
	"context"
	"net/http"
	"time"
)

// Service Communication Models

// ServiceRegistration represents a service registration request
type ServiceRegistration struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Protocol string            `json:"protocol"` // "http", "https", "tcp", "grpc"
	Version  string            `json:"version"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
}

// ServiceInstance represents a registered service instance
type ServiceInstance struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Host            string              `json:"host"`
	Port            int                 `json:"port"`
	Protocol        string              `json:"protocol"`
	Health          ServiceHealthStatus `json:"health"`
	Metadata        map[string]string   `json:"metadata"`
	Tags            []string            `json:"tags"`
	Version         string              `json:"version"`
	RegisteredAt    time.Time           `json:"registered_at"`
	LastHealthCheck time.Time           `json:"last_health_check"`
	Weight          int                 `json:"weight"` // For weighted load balancing
}

// ServiceHealthStatus represents the health status of a service
type ServiceHealthStatus string

const (
	ServiceHealthHealthy   ServiceHealthStatus = "healthy"
	ServiceHealthUnhealthy ServiceHealthStatus = "unhealthy"
	ServiceHealthUnknown   ServiceHealthStatus = "unknown"
)

// ServiceRequest represents a request to call another service
type ServiceRequest struct {
	ServiceName string            `json:"service_name"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Body        interface{}       `json:"body,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty" swaggertype:"string" format:"duration" example:"30s"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ServiceResponse represents a response from a service call
type ServiceResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    http.Header       `json:"headers"`
	Body       interface{}       `json:"body,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Latency    time.Duration     `json:"latency" swaggertype:"string" format:"duration" example:"50ms"`
}

// BroadcastMessage represents a message to broadcast to all service instances
type BroadcastMessage struct {
	ServiceName string            `json:"service_name"`
	Path        string            `json:"path"`
	Payload     interface{}       `json:"payload"`
	Headers     map[string]string `json:"headers,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty" swaggertype:"string" format:"duration" example:"30s"`
}

// ServiceEvent represents an event published between services
type ServiceEvent struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Source        string                 `json:"source"`
	Data          interface{}            `json:"data"`
	Timestamp     time.Time              `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
}

// EventHandler defines the function signature for event handlers
type EventHandler func(ctx context.Context, event *ServiceEvent) error

// Load Balancing Strategies

// LoadBalancingStrategy represents different load balancing strategies
type LoadBalancingStrategy string

const (
	LoadBalancingRoundRobin       LoadBalancingStrategy = "round_robin"
	LoadBalancingRandom           LoadBalancingStrategy = "random"
	LoadBalancingLeastConnections LoadBalancingStrategy = "least_connections"
	LoadBalancingWeighted         LoadBalancingStrategy = "weighted"
	LoadBalancingIPHash           LoadBalancingStrategy = "ip_hash"
)

// Circuit Breaker Models

// ServiceCircuitBreaker represents the state of a circuit breaker for services
type ServiceCircuitBreaker struct {
	ServiceName   string              `json:"service_name"`
	State         CircuitBreakerState `json:"state"`
	FailureCount  int                 `json:"failure_count"`
	SuccessCount  int                 `json:"success_count"`
	TotalRequests int64               `json:"total_requests"`
	LastFailure   time.Time           `json:"last_failure"`
	NextAttempt   time.Time           `json:"next_attempt"`
}

// CanRequest checks if the circuit breaker allows requests
func (cb *ServiceCircuitBreaker) CanRequest() bool {
	switch cb.State {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		return time.Now().After(cb.NextAttempt)
	case CircuitBreakerHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful request
func (cb *ServiceCircuitBreaker) RecordSuccess() {
	cb.SuccessCount++
	cb.TotalRequests++

	if cb.State == CircuitBreakerHalfOpen {
		cb.State = CircuitBreakerClosed
		cb.FailureCount = 0
	}
}

// RecordFailure records a failed request
func (cb *ServiceCircuitBreaker) RecordFailure() {
	cb.FailureCount++
	cb.TotalRequests++
	cb.LastFailure = time.Now()

	// Open circuit if failure threshold is exceeded
	if cb.FailureCount >= 5 { // Configurable threshold
		cb.State = CircuitBreakerOpen
		cb.NextAttempt = time.Now().Add(30 * time.Second) // Configurable timeout
	}
}

// Communication Metrics

// CommunicationMetrics represents metrics for service communication
type CommunicationMetrics struct {
	TotalRequests       int64                          `json:"total_requests"`
	SuccessfulRequests  int64                          `json:"successful_requests"`
	FailedRequests      int64                          `json:"failed_requests"`
	CircuitBreakerTrips int64                          `json:"circuit_breaker_trips"`
	RetryAttempts       int64                          `json:"retry_attempts"`
	AverageLatency      time.Duration                  `json:"average_latency" swaggertype:"string" format:"duration" example:"25ms"`
	ServiceCallLatency  map[string]time.Duration       `json:"service_call_latency" swaggertype:"object" example:"{\"user-service\": \"15ms\", \"auth-service\": \"10ms\"}"`
	ErrorRates          map[string]float64             `json:"error_rates"`
	ActiveConnections   int64                          `json:"active_connections"`
	ServiceHealth       map[string]ServiceHealthStatus `json:"service_health"`
	LoadBalancingStats  map[string]*LoadBalancingStats `json:"load_balancing_stats"`
	StartTime           time.Time                      `json:"start_time"`
	Uptime              time.Duration                  `json:"uptime" swaggertype:"string" format:"duration" example:"24h30m"`
}

// LoadBalancingStats represents statistics for load balancing
type LoadBalancingStats struct {
	Strategy      LoadBalancingStrategy `json:"strategy"`
	TotalRequests int64                 `json:"total_requests"`
	InstanceStats map[string]int64      `json:"instance_stats"` // instance_id -> request_count
}

// Service Discovery Models

// ServiceDiscoveryConfig represents service discovery configuration
type ServiceDiscoveryConfig struct {
	Enabled         bool          `json:"enabled"`
	Provider        string        `json:"provider"` // "kubernetes", "consul", "memory"
	RefreshInterval time.Duration `json:"refresh_interval" swaggertype:"string" format:"duration" example:"5m"`
	HealthCheck     bool          `json:"health_check"`
	Tags            []string      `json:"tags"`
}

// Communication Patterns

// MessagePattern represents different messaging patterns
type MessagePattern string

const (
	MessagePatternRequestResponse  MessagePattern = "request_response"
	MessagePatternFireAndForget    MessagePattern = "fire_and_forget"
	MessagePatternPublishSubscribe MessagePattern = "publish_subscribe"
	MessagePatternBroadcast        MessagePattern = "broadcast"
)

// CommunicationPattern represents a communication pattern configuration
type CommunicationPattern struct {
	Name           string                `json:"name"`
	Type           MessagePattern        `json:"type"`
	Timeout        time.Duration         `json:"timeout" swaggertype:"string" format:"duration" example:"10s"`
	RetryPolicy    *RetryPolicyConfig    `json:"retry_policy,omitempty"`
	CircuitBreaker bool                  `json:"circuit_breaker"`
	LoadBalancer   LoadBalancingStrategy `json:"load_balancer"`
}

// Retry Policy Models

// RetryPolicyConfig represents retry policy configuration
type RetryPolicyConfig struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay" swaggertype:"string" format:"duration" example:"1s"`
	MaxDelay        time.Duration `json:"max_delay" swaggertype:"string" format:"duration" example:"32s"`
	Multiplier      float64       `json:"multiplier"`
	RandomJitter    bool          `json:"random_jitter"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// Rate Limiting Models

// RateLimitConfig represents rate limiting configuration for service calls
type RateLimitConfig struct {
	Enabled           bool          `json:"enabled"`
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	TimeWindow        time.Duration `json:"time_window" swaggertype:"string" format:"duration" example:"1m"`
}

// Monitoring and Observability

// ServiceTrace represents tracing information for service calls
type ServiceTrace struct {
	TraceID       string                 `json:"trace_id"`
	SpanID        string                 `json:"span_id"`
	ParentSpanID  string                 `json:"parent_span_id,omitempty"`
	ServiceName   string                 `json:"service_name"`
	OperationName string                 `json:"operation_name"`
	StartTime     time.Time              `json:"start_time"`
	Duration      time.Duration          `json:"duration" swaggertype:"string" format:"duration" example:"15m"`
	Tags          map[string]interface{} `json:"tags"`
	Status        string                 `json:"status"` // "success", "error"
}

// ServiceLog represents structured logging for service communication
type ServiceLog struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	ServiceName string                 `json:"service_name"`
	Operation   string                 `json:"operation"`
	TraceID     string                 `json:"trace_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	Error       string                 `json:"error,omitempty"`
}

// Advanced Communication Patterns

// SagaPattern represents a saga pattern for distributed transactions
type SagaPattern struct {
	ID          string                 `json:"id"`
	Steps       []SagaStep             `json:"steps"`
	CurrentStep int                    `json:"current_step"`
	Status      SagaStatus             `json:"status"`
	StartTime   time.Time              `json:"start_time"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SagaStep represents a step in a saga
type SagaStep struct {
	ID               string             `json:"id"`
	ServiceName      string             `json:"service_name"`
	Action           string             `json:"action"`
	CompensateAction string             `json:"compensate_action"`
	Status           SagaStepStatus     `json:"status"`
	Timeout          time.Duration      `json:"timeout" swaggertype:"string" format:"duration" example:"30s"`
	RetryPolicy      *RetryPolicyConfig `json:"retry_policy,omitempty"`
}

// SagaStatus represents the status of a saga
type SagaStatus string

const (
	SagaStatusPending      SagaStatus = "pending"
	SagaStatusRunning      SagaStatus = "running"
	SagaStatusCompleted    SagaStatus = "completed"
	SagaStatusFailed       SagaStatus = "failed"
	SagaStatusCompensating SagaStatus = "compensating"
	SagaStatusCompensated  SagaStatus = "compensated"
)

// SagaStepStatus represents the status of a saga step
type SagaStepStatus string

const (
	SagaStepStatusPending      SagaStepStatus = "pending"
	SagaStepStatusRunning      SagaStepStatus = "running"
	SagaStepStatusCompleted    SagaStepStatus = "completed"
	SagaStepStatusFailed       SagaStepStatus = "failed"
	SagaStepStatusCompensating SagaStepStatus = "compensating"
	SagaStepStatusCompensated  SagaStepStatus = "compensated"
)

// CQRS Pattern Models

// Command represents a command in CQRS pattern
type Command struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	AggregateID string                 `json:"aggregate_id"`
	Data        interface{}            `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
	UserID      string                 `json:"user_id,omitempty"`
}

// Query represents a query in CQRS pattern
type Query struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Params    map[string]interface{} `json:"params"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id,omitempty"`
}

// Event Sourcing Models

// DomainEvent represents a domain event for event sourcing
type DomainEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	AggregateID string                 `json:"aggregate_id"`
	Version     int64                  `json:"version"`
	Data        interface{}            `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Methods for calculating derived values

// CalculateSuccessRate calculates the success rate for communication metrics
func (cm *CommunicationMetrics) CalculateSuccessRate() float64 {
	if cm.TotalRequests == 0 {
		return 0.0
	}
	return float64(cm.SuccessfulRequests) / float64(cm.TotalRequests) * 100.0
}

// CalculateErrorRate calculates the error rate for communication metrics
func (cm *CommunicationMetrics) CalculateErrorRate() float64 {
	if cm.TotalRequests == 0 {
		return 0.0
	}
	return float64(cm.FailedRequests) / float64(cm.TotalRequests) * 100.0
}

// IsHealthy returns true if the service health status is healthy
func (s ServiceHealthStatus) IsHealthy() bool {
	return s == ServiceHealthHealthy
}

// String methods for enums

func (s ServiceHealthStatus) String() string {
	return string(s)
}

func (lbs LoadBalancingStrategy) String() string {
	return string(lbs)
}

func (mp MessagePattern) String() string {
	return string(mp)
}

func (ss SagaStatus) String() string {
	return string(ss)
}

func (sss SagaStepStatus) String() string {
	return string(sss)
}
