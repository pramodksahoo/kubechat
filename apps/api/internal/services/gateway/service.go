package gateway

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// Service defines the API Gateway service interface
type Service interface {
	// Middleware
	RateLimitMiddleware() gin.HandlerFunc
	SecurityHeadersMiddleware() gin.HandlerFunc
	RequestLoggingMiddleware() gin.HandlerFunc
	ResponseTimeMiddleware() gin.HandlerFunc
	CircuitBreakerMiddleware() gin.HandlerFunc
	RequestSizeLimitMiddleware() gin.HandlerFunc
	GzipMiddleware() gin.HandlerFunc

	// Gateway management
	GetMetrics() *models.GatewayMetrics
	GetRateLimitStatus() map[string]*models.RateLimitStatus
	HealthCheck(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// Config represents the API Gateway configuration
type Config struct {
	// Rate limiting
	DefaultRateLimit int           `json:"default_rate_limit"` // requests per minute
	BurstLimit       int           `json:"burst_limit"`        // burst capacity
	RateLimitWindow  time.Duration `json:"rate_limit_window"`  // rate limit window

	// Circuit breaker
	CircuitBreakerThreshold int           `json:"circuit_breaker_threshold"` // failures before opening
	CircuitBreakerTimeout   time.Duration `json:"circuit_breaker_timeout"`   // timeout before half-open
	CircuitBreakerReset     time.Duration `json:"circuit_breaker_reset"`     // reset interval

	// Security
	EnableSecurityHeaders bool     `json:"enable_security_headers"`
	AllowedOrigins        []string `json:"allowed_origins"`
	AllowedMethods        []string `json:"allowed_methods"`
	AllowedHeaders        []string `json:"allowed_headers"`
	MaxRequestSize        int64    `json:"max_request_size"`

	// Logging
	EnableRequestLogging  bool   `json:"enable_request_logging"`
	EnableResponseLogging bool   `json:"enable_response_logging"`
	LogLevel              string `json:"log_level"`

	// Performance
	EnableGzip            bool          `json:"enable_gzip"`
	ReadTimeout           time.Duration `json:"read_timeout"`
	WriteTimeout          time.Duration `json:"write_timeout"`
	MaxConcurrentRequests int           `json:"max_concurrent_requests"`
}

// RateLimiter represents a rate limiter for a client
type RateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// CircuitBreaker represents a circuit breaker state
type CircuitBreaker struct {
	name            string
	maxFailures     int
	timeout         time.Duration
	resetTimeout    time.Duration
	state           models.CircuitBreakerState
	failures        int
	lastFailureTime time.Time
	mutex           sync.RWMutex
}

// service implements the Gateway service interface
type service struct {
	config          *Config
	rateLimiters    map[string]*RateLimiter
	circuitBreakers map[string]*CircuitBreaker
	metrics         *models.GatewayMetrics
	mutex           sync.RWMutex
	cleanupTicker   *time.Ticker
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewService creates a new API Gateway service
func NewService(config *Config) Service {
	if config == nil {
		config = &Config{
			DefaultRateLimit:        60, // 60 requests per minute
			BurstLimit:              10, // allow bursts of 10
			RateLimitWindow:         time.Minute,
			CircuitBreakerThreshold: 5, // 5 failures before opening
			CircuitBreakerTimeout:   30 * time.Second,
			CircuitBreakerReset:     60 * time.Second,
			EnableSecurityHeaders:   true,
			AllowedOrigins:          []string{"*"},
			AllowedMethods:          []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowedHeaders:          []string{"Content-Type", "Authorization", "X-Requested-With"},
			MaxRequestSize:          10 * 1024 * 1024, // 10MB
			EnableRequestLogging:    true,
			EnableResponseLogging:   false,
			LogLevel:                "info",
			EnableGzip:              true,
			ReadTimeout:             30 * time.Second,
			WriteTimeout:            30 * time.Second,
			MaxConcurrentRequests:   1000,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	svc := &service{
		config:          config,
		rateLimiters:    make(map[string]*RateLimiter),
		circuitBreakers: make(map[string]*CircuitBreaker),
		metrics: &models.GatewayMetrics{
			TotalRequests:       0,
			SuccessfulRequests:  0,
			FailedRequests:      0,
			RateLimitedRequests: 0,
			AverageResponseTime: 0,
			ActiveConnections:   0,
			CircuitBreakerTrips: 0,
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Start cleanup ticker for rate limiters
	svc.cleanupTicker = time.NewTicker(5 * time.Minute)
	go svc.cleanup()

	return svc
}

// getRateLimiter gets or creates a rate limiter for a client IP
func (s *service) getRateLimiter(clientIP string) *RateLimiter {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	limiter, exists := s.rateLimiters[clientIP]
	if !exists {
		limiter = &RateLimiter{
			limiter:  rate.NewLimiter(rate.Every(s.config.RateLimitWindow/time.Duration(s.config.DefaultRateLimit)), s.config.BurstLimit),
			lastSeen: time.Now(),
		}
		s.rateLimiters[clientIP] = limiter
	} else {
		limiter.lastSeen = time.Now()
	}

	return limiter
}

// getCircuitBreaker gets or creates a circuit breaker for a service
func (s *service) getCircuitBreaker(serviceName string) *CircuitBreaker {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	breaker, exists := s.circuitBreakers[serviceName]
	if !exists {
		breaker = &CircuitBreaker{
			name:         serviceName,
			maxFailures:  s.config.CircuitBreakerThreshold,
			timeout:      s.config.CircuitBreakerTimeout,
			resetTimeout: s.config.CircuitBreakerReset,
			state:        models.CircuitBreakerClosed,
		}
		s.circuitBreakers[serviceName] = breaker
	}

	return breaker
}

// cleanup removes old rate limiters
func (s *service) cleanup() {
	ticker := s.cleanupTicker
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mutex.Lock()
			now := time.Now()
			for ip, limiter := range s.rateLimiters {
				if now.Sub(limiter.lastSeen) > 10*time.Minute {
					delete(s.rateLimiters, ip)
				}
			}
			s.mutex.Unlock()

		case <-s.ctx.Done():
			return
		}
	}
}

// GetClientIP extracts the real client IP address
func GetClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first
	if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
		// Take the first IP in the chain
		if idx := len(forwarded); idx > 0 {
			if commaIdx := 0; commaIdx != -1 {
				for i, r := range forwarded {
					if r == ',' {
						return forwarded[:i]
					}
				}
				return forwarded
			}
		}
	}

	// Check X-Real-IP header
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return c.ClientIP()
}

// GetMetrics returns current gateway metrics
func (s *service) GetMetrics() *models.GatewayMetrics {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Create a copy to avoid race conditions
	metrics := *s.metrics
	metrics.ActiveRateLimiters = len(s.rateLimiters)
	metrics.ActiveCircuitBreakers = len(s.circuitBreakers)

	return &metrics
}

// GetRateLimitStatus returns rate limit status for all clients
func (s *service) GetRateLimitStatus() map[string]*models.RateLimitStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	status := make(map[string]*models.RateLimitStatus)
	for ip, limiter := range s.rateLimiters {
		status[ip] = &models.RateLimitStatus{
			Limit:     s.config.DefaultRateLimit,
			Remaining: int(limiter.limiter.Tokens()),
			ResetTime: limiter.lastSeen.Add(s.config.RateLimitWindow),
		}
	}

	return status
}

// HealthCheck performs health check
func (s *service) HealthCheck(ctx context.Context) error {
	select {
	case <-s.ctx.Done():
		return fmt.Errorf("gateway service not running")
	default:
	}

	// Check if cleanup routine is running
	if s.cleanupTicker == nil {
		return fmt.Errorf("cleanup routine not running")
	}

	return nil
}

// Shutdown gracefully shuts down the service
func (s *service) Shutdown(ctx context.Context) error {
	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
	}

	s.cancel()
	return nil
}
