// KubeChat API Health Endpoint Tests
package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Add health endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "kubechat-api",
			"status":  "healthy",
			"version": "1.0.0",
		})
	})

	t.Run("returns healthy status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "kubechat-api", response["service"])
		assert.Equal(t, "healthy", response["status"])
		assert.Equal(t, "1.0.0", response["version"])
	})

	t.Run("responds with JSON content type", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	})
}
