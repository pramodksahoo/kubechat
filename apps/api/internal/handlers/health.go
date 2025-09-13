package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Environment string    `json:"environment"`
	Version     string    `json:"version"`
	Uptime      string    `json:"uptime"`
}

var startTime = time.Now()

// HealthCheck returns the health status of the API
//
//	@Summary		API health check
//	@Description	Returns basic health status and uptime information for the API
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	HealthResponse	"API health status"
//	@Router			/health [get]
func HealthCheck(c *gin.Context) {
	uptime := time.Since(startTime)

	response := HealthResponse{
		Status:      "ok",
		Timestamp:   time.Now(),
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
		Version:     getEnvOrDefault("VERSION", "1.0.0"),
		Uptime:      uptime.String(),
	}

	c.JSON(http.StatusOK, response)
}

// GetStatus returns detailed API status information
//
//	@Summary		Get detailed API status
//	@Description	Returns comprehensive status information including service health
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Detailed API status"
//	@Router			/status [get]
func GetStatus(c *gin.Context) {
	uptime := time.Since(startTime)

	status := map[string]interface{}{
		"api": map[string]interface{}{
			"status":      "healthy",
			"version":     getEnvOrDefault("VERSION", "1.0.0"),
			"environment": getEnvOrDefault("ENVIRONMENT", "development"),
			"uptime":      uptime.String(),
		},
		"services": map[string]interface{}{
			"database": map[string]string{
				"status": "pending", // Will be implemented with actual DB connection
			},
			"redis": map[string]string{
				"status": "pending", // Will be implemented with actual Redis connection
			},
			"kubernetes": map[string]string{
				"status": "pending", // Will be implemented with actual K8s connection
			},
			"ollama": map[string]string{
				"status": "pending", // Will be implemented with actual Ollama connection
			},
		},
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, status)
}

func getEnvOrDefault(key, defaultValue string) string {
	// This is a simple implementation - in a real app you'd use the config package
	return defaultValue
}
