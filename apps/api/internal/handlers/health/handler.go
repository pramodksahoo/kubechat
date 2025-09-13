package health

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/health"
)

// Handler handles health check HTTP requests
type Handler struct {
	healthService health.Service
}

// NewHandler creates a new health handler
func NewHandler(healthService health.Service) *Handler {
	return &Handler{
		healthService: healthService,
	}
}

// RegisterRoutes registers health check routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	healthRoutes := router.Group("/health")
	{
		// Public endpoints (no authentication required)
		healthRoutes.GET("/", h.GetOverallHealth)
		healthRoutes.GET("/status", h.GetOverallHealth)
		healthRoutes.GET("/live", h.GetLivenessProbe)
		healthRoutes.GET("/ready", h.GetReadinessProbe)

		// Detailed health endpoints
		healthRoutes.GET("/detailed", h.GetDetailedHealth)
		healthRoutes.GET("/components", h.GetComponentsHealth)
		healthRoutes.GET("/components/:component", h.GetComponentHealth)

		// Metrics endpoints
		healthRoutes.GET("/metrics", h.GetHealthMetrics)

		// Database-specific health checks
		healthRoutes.GET("/database", h.GetDatabaseHealth)
		healthRoutes.GET("/redis", h.GetRedisHealth)
		healthRoutes.GET("/external", h.GetExternalServicesHealth)
	}
}

// GetOverallHealth returns the overall health status
//
//	@Summary		Get overall system health status
//	@Description	Returns comprehensive health status of all system components
//	@Tags			Health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	models.HealthStatus		"System healthy"
//	@Success		503	{object}	models.HealthStatus		"System degraded or unhealthy"
//	@Failure		500	{object}	map[string]interface{}	"Health check failed"
//	@Router			/health/ [get]
func (h *Handler) GetOverallHealth(c *gin.Context) {
	ctx := c.Request.Context()

	status, err := h.healthService.CheckOverallHealth(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Health check failed",
			"code":    "HEALTH_CHECK_FAILED",
			"details": err.Error(),
		})
		return
	}

	// Determine HTTP status code based on health status
	httpStatus := h.getHTTPStatusFromHealth(status.Status)

	c.JSON(httpStatus, status)
}

// GetLivenessProbe returns a simple liveness probe
//
//	@Summary		Kubernetes liveness probe endpoint
//	@Description	Simple liveness check for Kubernetes pod health monitoring
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Service is alive"
//	@Router			/health/live [get]
func (h *Handler) GetLivenessProbe(c *gin.Context) {
	// Liveness probe should always return 200 unless the service is completely down
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now(),
		"service":   "kubechat-api",
	})
}

// GetReadinessProbe returns a readiness probe
func (h *Handler) GetReadinessProbe(c *gin.Context) {
	ctx := c.Request.Context()

	// Check if essential components are ready
	status, err := h.healthService.CheckOverallHealth(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not_ready",
			"error":     err.Error(),
			"timestamp": time.Now(),
		})
		return
	}

	// Service is ready if it's healthy or degraded (but not unhealthy)
	if status.Status == models.HealthStatusUnhealthy {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not_ready",
			"message":   status.Message,
			"timestamp": time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"message":   status.Message,
		"timestamp": time.Now(),
	})
}

// GetDetailedHealth returns detailed health information
func (h *Handler) GetDetailedHealth(c *gin.Context) {
	ctx := c.Request.Context()

	status, err := h.healthService.CheckOverallHealth(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Detailed health check failed",
			"code":    "DETAILED_HEALTH_CHECK_FAILED",
			"details": err.Error(),
		})
		return
	}

	// Add metrics to the response
	metrics := h.healthService.GetHealthMetrics()

	response := gin.H{
		"health":  status,
		"metrics": metrics,
		"system": gin.H{
			"uptime":    time.Since(metrics.StartTime),
			"timestamp": time.Now(),
			"version":   "1.0.0", // Should come from build info
		},
	}

	httpStatus := h.getHTTPStatusFromHealth(status.Status)
	c.JSON(httpStatus, response)
}

// GetComponentsHealth returns health status for all components
func (h *Handler) GetComponentsHealth(c *gin.Context) {
	ctx := c.Request.Context()

	status, err := h.healthService.CheckOverallHealth(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Components health check failed",
			"code":    "COMPONENTS_HEALTH_CHECK_FAILED",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"components": status.Components,
		"summary": gin.H{
			"total":     len(status.Components),
			"healthy":   h.countComponentsByStatus(status.Components, models.HealthStatusHealthy),
			"degraded":  h.countComponentsByStatus(status.Components, models.HealthStatusDegraded),
			"unhealthy": h.countComponentsByStatus(status.Components, models.HealthStatusUnhealthy),
		},
		"timestamp": time.Now(),
	})
}

// GetComponentHealth returns health status for a specific component
func (h *Handler) GetComponentHealth(c *gin.Context) {
	ctx := c.Request.Context()
	componentName := c.Param("component")

	if componentName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Component name is required",
			"code":  "COMPONENT_NAME_REQUIRED",
		})
		return
	}

	// Get overall health to find the specific component
	status, err := h.healthService.CheckOverallHealth(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Component health check failed",
			"code":    "COMPONENT_HEALTH_CHECK_FAILED",
			"details": err.Error(),
		})
		return
	}

	component, exists := status.Components[componentName]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":     "Component not found",
			"code":      "COMPONENT_NOT_FOUND",
			"component": componentName,
		})
		return
	}

	httpStatus := h.getHTTPStatusFromHealth(component.Status)
	c.JSON(httpStatus, component)
}

// GetHealthMetrics returns health metrics
func (h *Handler) GetHealthMetrics(c *gin.Context) {
	metrics := h.healthService.GetHealthMetrics()

	// Calculate success rates
	metrics.CalculateSuccessRate()
	for _, componentMetrics := range metrics.ComponentMetrics {
		componentMetrics.CalculateSuccessRate()
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics":   metrics,
		"timestamp": time.Now(),
	})
}

// GetDatabaseHealth returns database-specific health information
func (h *Handler) GetDatabaseHealth(c *gin.Context) {
	ctx := c.Request.Context()

	health, err := h.healthService.CheckDatabaseHealth(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database health check failed",
			"code":    "DATABASE_HEALTH_CHECK_FAILED",
			"details": err.Error(),
		})
		return
	}

	httpStatus := h.getHTTPStatusFromHealth(health.Status)
	c.JSON(httpStatus, health)
}

// GetRedisHealth returns Redis-specific health information
func (h *Handler) GetRedisHealth(c *gin.Context) {
	ctx := c.Request.Context()

	health, err := h.healthService.CheckRedisHealth(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Redis health check failed",
			"code":    "REDIS_HEALTH_CHECK_FAILED",
			"details": err.Error(),
		})
		return
	}

	httpStatus := h.getHTTPStatusFromHealth(health.Status)
	c.JSON(httpStatus, health)
}

// GetExternalServicesHealth returns external services health information
func (h *Handler) GetExternalServicesHealth(c *gin.Context) {
	ctx := c.Request.Context()

	health, err := h.healthService.CheckExternalServicesHealth(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "External services health check failed",
			"code":    "EXTERNAL_SERVICES_HEALTH_CHECK_FAILED",
			"details": err.Error(),
		})
		return
	}

	httpStatus := h.getHTTPStatusFromHealth(health.Status)
	c.JSON(httpStatus, health)
}

// Helper methods

// getHTTPStatusFromHealth converts health status to appropriate HTTP status code
func (h *Handler) getHTTPStatusFromHealth(status models.HealthStatusEnum) int {
	switch status {
	case models.HealthStatusHealthy:
		return http.StatusOK
	case models.HealthStatusDegraded:
		return http.StatusOK // Still operational, but with warnings
	case models.HealthStatusUnhealthy:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// countComponentsByStatus counts components by their health status
func (h *Handler) countComponentsByStatus(components map[string]*models.ComponentHealth, status models.HealthStatusEnum) int {
	count := 0
	for _, component := range components {
		if component.Status == status {
			count++
		}
	}
	return count
}

// parseTimeParam parses time parameter from query string
func (h *Handler) parseTimeParam(c *gin.Context, param string, defaultValue time.Duration) time.Duration {
	if value := c.Query(param); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultValue
}

// parseBoolParam parses boolean parameter from query string
func (h *Handler) parseBoolParam(c *gin.Context, param string, defaultValue bool) bool {
	if value := c.Query(param); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// parseIntParam parses integer parameter from query string
func (h *Handler) parseIntParam(c *gin.Context, param string, defaultValue int) int {
	if value := c.Query(param); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
