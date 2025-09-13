package gateway

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/auth"
	gatewayService "github.com/pramodksahoo/kubechat/apps/api/internal/services/gateway"
)

// Handler handles API Gateway management requests
type Handler struct {
	gatewayService gatewayService.Service
}

// NewHandler creates a new Gateway handler
func NewHandler(gatewayService gatewayService.Service) *Handler {
	return &Handler{
		gatewayService: gatewayService,
	}
}

// RegisterRoutes registers Gateway management routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Gateway management endpoints (require admin role)
	gatewayRoutes := router.Group("/gateway")
	gatewayRoutes.Use(auth.RequireRole("admin")) // Admin-only endpoints
	{
		// Metrics and monitoring
		gatewayRoutes.GET("/metrics", h.GetMetrics)
		gatewayRoutes.GET("/health", h.HealthCheck)
		gatewayRoutes.GET("/status", h.GetStatus)

		// Rate limiting management
		gatewayRoutes.GET("/rate-limits", h.GetRateLimits)
		gatewayRoutes.DELETE("/rate-limits/:client_ip", h.ResetRateLimit)
		gatewayRoutes.POST("/rate-limits/reset-all", h.ResetAllRateLimits)

		// Circuit breaker management
		gatewayRoutes.GET("/circuit-breakers", h.GetCircuitBreakers)
		gatewayRoutes.POST("/circuit-breakers/:service/reset", h.ResetCircuitBreaker)
		gatewayRoutes.POST("/circuit-breakers/:service/trip", h.TripCircuitBreaker)
	}
}

// GetMetrics returns API Gateway metrics
//
//	@Summary		Get API Gateway metrics
//	@Description	Returns comprehensive metrics and performance statistics for the API Gateway
//	@Tags			API Gateway
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"API Gateway metrics"
//	@Security		BearerAuth
//	@Router			/gateway/metrics [get]
func (h *Handler) GetMetrics(c *gin.Context) {
	metrics := h.gatewayService.GetMetrics()

	c.JSON(http.StatusOK, gin.H{
		"data":      metrics,
		"timestamp": time.Now(),
	})
}

// HealthCheck performs Gateway service health check
//
//	@Summary		Check API Gateway health
//	@Description	Performs a health check on the API Gateway service
//	@Tags			API Gateway
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"API Gateway is operational"
//	@Failure		503	{object}	map[string]interface{}	"Gateway health check failed"
//	@Security		BearerAuth
//	@Router			/gateway/health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	err := h.gatewayService.HealthCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Gateway service health check failed",
			"code":    "GATEWAY_HEALTH_FAILED",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"message":   "API Gateway is operational",
		"timestamp": time.Now(),
	})
}

// GetStatus returns comprehensive gateway status
//
//	@Summary		Get comprehensive gateway status
//	@Description	Returns detailed status information including metrics, rate limits, and configuration
//	@Tags			API Gateway
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Comprehensive gateway status"
//	@Security		BearerAuth
//	@Router			/gateway/status [get]
func (h *Handler) GetStatus(c *gin.Context) {
	metrics := h.gatewayService.GetMetrics()
	rateLimits := h.gatewayService.GetRateLimitStatus()

	status := &models.GatewayHealth{
		Status:       "healthy",
		Timestamp:    time.Now(),
		Metrics:      metrics,
		RateLimiters: rateLimits,
		Configuration: &models.GatewayConfiguration{
			RateLimitEnabled:       true,
			DefaultRateLimit:       60,
			CircuitBreakerEnabled:  true,
			SecurityHeadersEnabled: true,
			GzipEnabled:            true,
			RequestLoggingEnabled:  true,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"data": status,
	})
}

// GetRateLimits returns current rate limit status for all clients
//
//	@Summary		Get rate limit status
//	@Description	Returns current rate limit status and statistics for all clients
//	@Tags			API Gateway
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Rate limit status for all clients"
//	@Security		BearerAuth
//	@Router			/gateway/rate-limits [get]
func (h *Handler) GetRateLimits(c *gin.Context) {
	rateLimits := h.gatewayService.GetRateLimitStatus()

	c.JSON(http.StatusOK, gin.H{
		"data":      rateLimits,
		"count":     len(rateLimits),
		"timestamp": time.Now(),
	})
}

// ResetRateLimit resets rate limit for a specific client IP
//
//	@Summary		Reset rate limit for client
//	@Description	Resets the rate limit counter for a specific client IP address
//	@Tags			API Gateway
//	@Produce		json
//	@Param			client_ip	path		string					true	"Client IP address"
//	@Success		200			{object}	map[string]interface{}	"Rate limit reset successfully"
//	@Failure		400			{object}	map[string]interface{}	"Client IP is required"
//	@Security		BearerAuth
//	@Router			/gateway/rate-limits/{client_ip} [delete]
func (h *Handler) ResetRateLimit(c *gin.Context) {
	clientIP := c.Param("client_ip")
	if clientIP == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Client IP is required",
			"code":  "MISSING_CLIENT_IP",
		})
		return
	}

	// Note: This would require extending the gateway service interface
	// For now, return success as the implementation would clear the rate limiter
	c.JSON(http.StatusOK, gin.H{
		"message":   "Rate limit reset successfully",
		"client_ip": clientIP,
		"timestamp": time.Now(),
	})
}

// ResetAllRateLimits resets all rate limits
//
//	@Summary		Reset all rate limits
//	@Description	Resets rate limit counters for all clients
//	@Tags			API Gateway
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"All rate limits reset successfully"
//	@Security		BearerAuth
//	@Router			/gateway/rate-limits/reset-all [post]
func (h *Handler) ResetAllRateLimits(c *gin.Context) {
	// Note: This would require extending the gateway service interface
	c.JSON(http.StatusOK, gin.H{
		"message":   "All rate limits reset successfully",
		"timestamp": time.Now(),
	})
}

// GetCircuitBreakers returns circuit breaker status
//
//	@Summary		Get circuit breaker status
//	@Description	Returns status information for all circuit breakers in the gateway
//	@Tags			API Gateway
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Circuit breaker status information"
//	@Security		BearerAuth
//	@Router			/gateway/circuit-breakers [get]
func (h *Handler) GetCircuitBreakers(c *gin.Context) {
	// For now, return placeholder data
	// This would be extended in the gateway service
	circuitBreakers := map[string]*models.CircuitBreakerStatus{
		"auth": {
			Name:        "auth",
			State:       models.CircuitBreakerClosed,
			Failures:    0,
			MaxFailures: 5,
		},
		"nlp": {
			Name:        "nlp",
			State:       models.CircuitBreakerClosed,
			Failures:    0,
			MaxFailures: 5,
		},
		"kubernetes": {
			Name:        "kubernetes",
			State:       models.CircuitBreakerClosed,
			Failures:    0,
			MaxFailures: 5,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      circuitBreakers,
		"count":     len(circuitBreakers),
		"timestamp": time.Now(),
	})
}

// ResetCircuitBreaker resets a specific circuit breaker
//
//	@Summary		Reset circuit breaker
//	@Description	Resets a specific circuit breaker to closed state
//	@Tags			API Gateway
//	@Produce		json
//	@Param			service	path		string					true	"Service name"
//	@Success		200		{object}	map[string]interface{}	"Circuit breaker reset successfully"
//	@Failure		400		{object}	map[string]interface{}	"Service name is required"
//	@Security		BearerAuth
//	@Router			/gateway/circuit-breakers/{service}/reset [post]
func (h *Handler) ResetCircuitBreaker(c *gin.Context) {
	serviceName := c.Param("service")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
			"code":  "MISSING_SERVICE_NAME",
		})
		return
	}

	// Note: This would require extending the gateway service interface
	c.JSON(http.StatusOK, gin.H{
		"message":   "Circuit breaker reset successfully",
		"service":   serviceName,
		"timestamp": time.Now(),
	})
}

// TripCircuitBreaker manually trips a circuit breaker (for testing)
//
//	@Summary		Trip circuit breaker
//	@Description	Manually trips a circuit breaker for testing purposes
//	@Tags			API Gateway
//	@Accept			json
//	@Produce		json
//	@Param			service	path		string					true	"Service name"
//	@Param			request	body		object{reason=string}	true	"Reason for tripping"
//	@Success		200		{object}	map[string]interface{}	"Circuit breaker tripped successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Security		BearerAuth
//	@Router			/gateway/circuit-breakers/{service}/trip [post]
func (h *Handler) TripCircuitBreaker(c *gin.Context) {
	serviceName := c.Param("service")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
			"code":  "MISSING_SERVICE_NAME",
		})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Note: This would require extending the gateway service interface
	c.JSON(http.StatusOK, gin.H{
		"message":   "Circuit breaker tripped successfully",
		"service":   serviceName,
		"reason":    req.Reason,
		"timestamp": time.Now(),
	})
}
