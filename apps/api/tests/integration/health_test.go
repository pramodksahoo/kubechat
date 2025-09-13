package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	BaseURL = "http://localhost:30080"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

// TestHealthEndpoint tests the health endpoint against the running Kubernetes service
func TestHealthEndpoint(t *testing.T) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(BaseURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var health HealthResponse
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err)

	assert.Equal(t, "kubechat-api", health.Service)
	assert.Equal(t, "healthy", health.Status)
	assert.NotEmpty(t, health.Version)
}

// TestHealthEndpointConcurrency tests the health endpoint with concurrent requests
func TestHealthEndpointConcurrency(t *testing.T) {
	const numRequests = 10

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := client.Get(BaseURL + "/health")
			if err != nil {
				results <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results <- err
				return
			}

			results <- nil
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}

// TestAPIConnectivity tests basic connectivity to the Kubernetes-deployed API
func TestAPIConnectivity(t *testing.T) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Test health endpoint specifically (avoiding potential rate limiting on other endpoints)
	t.Run("health_endpoint", func(t *testing.T) {
		resp, err := client.Get(BaseURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should get either OK (200) or rate limited (429) - both indicate API is running
		assert.Contains(t, []int{http.StatusOK, http.StatusTooManyRequests}, resp.StatusCode)
	})
}
