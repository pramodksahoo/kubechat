package models

import (
	"fmt"
	"time"
)

// GatewayMetrics represents API Gateway performance metrics
type GatewayMetrics struct {
	TotalRequests         int64         `json:"total_requests"`
	SuccessfulRequests    int64         `json:"successful_requests"`
	FailedRequests        int64         `json:"failed_requests"`
	RateLimitedRequests   int64         `json:"rate_limited_requests"`
	AverageResponseTime   time.Duration `json:"average_response_time" swaggertype:"string" format:"duration" example:"15ms"`
	ActiveConnections     int64         `json:"active_connections"`
	CircuitBreakerTrips   int64         `json:"circuit_breaker_trips"`
	ActiveRateLimiters    int           `json:"active_rate_limiters"`
	ActiveCircuitBreakers int           `json:"active_circuit_breakers"`
	Uptime                string        `json:"uptime"`
}

// RateLimitStatus represents rate limiting status for a client
type RateLimitStatus struct {
	Limit     int       `json:"limit"`      // requests per window
	Remaining int       `json:"remaining"`  // remaining requests
	ResetTime time.Time `json:"reset_time"` // when the limit resets
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerClosed   CircuitBreakerState = "closed"
	CircuitBreakerOpen     CircuitBreakerState = "open"
	CircuitBreakerHalfOpen CircuitBreakerState = "half_open"
)

// CircuitBreakerStatus represents circuit breaker status
type CircuitBreakerStatus struct {
	Name        string              `json:"name"`
	State       CircuitBreakerState `json:"state"`
	Failures    int                 `json:"failures"`
	MaxFailures int                 `json:"max_failures"`
	LastFailure *time.Time          `json:"last_failure,omitempty"`
	NextAttempt *time.Time          `json:"next_attempt,omitempty"`
}

// GatewayHealth represents the health status of the API Gateway
type GatewayHealth struct {
	Status          string                           `json:"status"`
	Timestamp       time.Time                        `json:"timestamp"`
	Metrics         *GatewayMetrics                  `json:"metrics"`
	RateLimiters    map[string]*RateLimitStatus      `json:"rate_limiters,omitempty"`
	CircuitBreakers map[string]*CircuitBreakerStatus `json:"circuit_breakers,omitempty"`
	Configuration   *GatewayConfiguration            `json:"configuration"`
}

// GatewayConfiguration represents gateway configuration for health checks
type GatewayConfiguration struct {
	RateLimitEnabled       bool `json:"rate_limit_enabled"`
	DefaultRateLimit       int  `json:"default_rate_limit"`
	CircuitBreakerEnabled  bool `json:"circuit_breaker_enabled"`
	SecurityHeadersEnabled bool `json:"security_headers_enabled"`
	GzipEnabled            bool `json:"gzip_enabled"`
	RequestLoggingEnabled  bool `json:"request_logging_enabled"`
}

// ServiceEndpoint represents an API endpoint configuration
type ServiceEndpoint struct {
	Path          string         `json:"path"`
	Method        string         `json:"method"`
	Service       string         `json:"service"`
	RateLimit     *int           `json:"rate_limit,omitempty"`                                                   // Override default rate limit
	Timeout       *time.Duration `json:"timeout,omitempty" swaggertype:"string" format:"duration" example:"30s"` // Request timeout
	RetryCount    *int           `json:"retry_count,omitempty"`                                                  // Number of retries
	AuthRequired  bool           `json:"auth_required"`                                                          // Whether authentication is required
	AdminRequired bool           `json:"admin_required"`                                                         // Whether admin role is required
}

// RequestContext represents enhanced request context for gateway
type RequestContext struct {
	RequestID      string            `json:"request_id"`
	ClientIP       string            `json:"client_ip"`
	UserAgent      string            `json:"user_agent"`
	Method         string            `json:"method"`
	Path           string            `json:"path"`
	Headers        map[string]string `json:"headers"`
	StartTime      time.Time         `json:"start_time"`
	EndTime        *time.Time        `json:"end_time,omitempty"`
	ResponseTime   *time.Duration    `json:"response_time,omitempty" swaggertype:"string" format:"duration" example:"20ms"`
	StatusCode     *int              `json:"status_code,omitempty"`
	ResponseSize   *int64            `json:"response_size,omitempty"`
	UserID         *string           `json:"user_id,omitempty"`
	SessionID      *string           `json:"session_id,omitempty"`
	ServiceName    string            `json:"service_name"`
	RateLimited    bool              `json:"rate_limited"`
	CircuitBreaker string            `json:"circuit_breaker,omitempty"`
}

// GatewayError represents API Gateway specific errors
type GatewayError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   string                 `json:"details,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Error implements the error interface
func (e *GatewayError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
}

// Common gateway error codes
const (
	GatewayErrorRateLimit       = "RATE_LIMIT_EXCEEDED"
	GatewayErrorCircuitBreaker  = "CIRCUIT_BREAKER_OPEN"
	GatewayErrorRequestTooLarge = "REQUEST_TOO_LARGE"
	GatewayErrorTimeout         = "REQUEST_TIMEOUT"
	GatewayErrorUnauthorized    = "UNAUTHORIZED"
	GatewayErrorForbidden       = "FORBIDDEN"
	GatewayErrorServiceDown     = "SERVICE_UNAVAILABLE"
	GatewayErrorInternalError   = "INTERNAL_GATEWAY_ERROR"
)
