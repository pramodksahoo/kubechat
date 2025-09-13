package gateway

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// RateLimitMiddleware implements rate limiting per client IP
func (s *service) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := GetClientIP(c)
		limiter := s.getRateLimiter(clientIP)

		if !limiter.limiter.Allow() {
			s.metrics.RateLimitedRequests++
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"code":        "RATE_LIMIT_EXCEEDED",
				"retry_after": s.config.RateLimitWindow.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func (s *service) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.config.EnableSecurityHeaders {
			// CORS headers
			origin := c.GetHeader("Origin")
			if origin != "" && s.isAllowedOrigin(origin) {
				c.Header("Access-Control-Allow-Origin", origin)
			} else if len(s.config.AllowedOrigins) > 0 && s.config.AllowedOrigins[0] == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			}

			c.Header("Access-Control-Allow-Methods", strings.Join(s.config.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(s.config.AllowedHeaders, ", "))
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")

			// Security headers
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("X-Frame-Options", "DENY")
			c.Header("X-XSS-Protection", "1; mode=block")
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
			c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

			// Handle preflight requests
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		}

		c.Next()
	}
}

// RequestLoggingMiddleware logs incoming requests
func (s *service) RequestLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		if !s.config.EnableRequestLogging {
			return ""
		}

		return fmt.Sprintf("[GATEWAY] %v | %3d | %13v | %15s | %-7s %#v\n%s",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
			param.ErrorMessage,
		)
	})
}

// ResponseTimeMiddleware tracks response times
func (s *service) ResponseTimeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Update active connections
		s.metrics.ActiveConnections++

		c.Next()

		// Calculate response time and update metrics
		duration := time.Since(start)
		s.metrics.ActiveConnections--
		s.metrics.TotalRequests++

		if c.Writer.Status() >= 200 && c.Writer.Status() < 400 {
			s.metrics.SuccessfulRequests++
		} else {
			s.metrics.FailedRequests++
		}

		// Update average response time (simple moving average)
		s.updateAverageResponseTime(duration)

		// Add response time header
		c.Header("X-Response-Time", duration.String())
	}
}

// CircuitBreakerMiddleware implements circuit breaker pattern
func (s *service) CircuitBreakerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := s.getServiceNameFromPath(c.Request.URL.Path)
		breaker := s.getCircuitBreaker(serviceName)

		if !breaker.canRequest() {
			s.metrics.CircuitBreakerTrips++
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service temporarily unavailable",
				"code":    "CIRCUIT_BREAKER_OPEN",
				"service": serviceName,
			})
			c.Abort()
			return
		}

		c.Next()

		// Update circuit breaker based on response
		if c.Writer.Status() >= 500 {
			breaker.recordFailure()
		} else {
			breaker.recordSuccess()
		}
	}
}

// GzipMiddleware adds gzip compression
func (s *service) GzipMiddleware() gin.HandlerFunc {
	if !s.config.EnableGzip {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Set gzip headers
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		// Create gzip writer
		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()

		// Wrap the response writer
		c.Writer = &gzipWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
		}

		c.Next()
	}
}

// RequestSizeLimitMiddleware limits request body size
func (s *service) RequestSizeLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > s.config.MaxRequestSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":    "Request body too large",
				"code":     "REQUEST_TOO_LARGE",
				"max_size": s.config.MaxRequestSize,
			})
			c.Abort()
			return
		}

		// Limit the request body reader
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, s.config.MaxRequestSize)

		c.Next()
	}
}

// Helper methods

func (s *service) isAllowedOrigin(origin string) bool {
	for _, allowed := range s.config.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func (s *service) updateAverageResponseTime(duration time.Duration) {
	// Simple exponential moving average
	if s.metrics.AverageResponseTime == 0 {
		s.metrics.AverageResponseTime = duration
	} else {
		// Weighted average: 90% old, 10% new
		s.metrics.AverageResponseTime = time.Duration(
			float64(s.metrics.AverageResponseTime)*0.9 + float64(duration)*0.1,
		)
	}
}

func (s *service) getServiceNameFromPath(path string) string {
	// Extract service name from path
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "v1" {
		return parts[2] // e.g., "auth", "nlp", "kubernetes", etc.
	}
	return "unknown"
}

// Circuit breaker methods

func (cb *CircuitBreaker) canRequest() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()

	switch cb.state {
	case models.CircuitBreakerClosed:
		return true
	case models.CircuitBreakerOpen:
		if now.Sub(cb.lastFailureTime) > cb.timeout {
			cb.state = models.CircuitBreakerHalfOpen
			return true
		}
		return false
	case models.CircuitBreakerHalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures = 0
	if cb.state == models.CircuitBreakerHalfOpen {
		cb.state = models.CircuitBreakerClosed
	}
}

func (cb *CircuitBreaker) recordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.state = models.CircuitBreakerOpen
	}
}

// gzipWriter wraps gin.ResponseWriter for gzip compression
type gzipWriter struct {
	gin.ResponseWriter
	writer io.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}
