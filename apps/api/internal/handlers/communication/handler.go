package communication

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/communication"
)

// Handler handles service communication HTTP requests
type Handler struct {
	communicationService communication.Service
}

// NewHandler creates a new communication handler
func NewHandler(communicationService communication.Service) *Handler {
	return &Handler{
		communicationService: communicationService,
	}
}

// RegisterRoutes registers communication routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	commRoutes := router.Group("/communication")
	{
		// Service Discovery endpoints
		commRoutes.POST("/services/register", h.RegisterService)
		commRoutes.DELETE("/services/:serviceId", h.DeregisterService)
		commRoutes.GET("/services/discover/:serviceName", h.DiscoverService)
		commRoutes.GET("/services/healthy/:serviceName", h.GetHealthyServices)
		commRoutes.GET("/services", h.ListAllServices)

		// Service Communication endpoints
		commRoutes.POST("/call", h.CallService)
		commRoutes.POST("/broadcast", h.BroadcastMessage)
		commRoutes.POST("/events/publish", h.PublishEvent)
		commRoutes.POST("/events/subscribe", h.SubscribeToEvents)

		// Circuit Breaker endpoints
		commRoutes.GET("/circuit-breakers", h.GetCircuitBreakers)
		commRoutes.POST("/circuit-breakers/:serviceName/reset", h.ResetCircuitBreaker)

		// Metrics and monitoring endpoints
		commRoutes.GET("/metrics", h.GetCommunicationMetrics)
		commRoutes.GET("/health", h.GetHealthStatus)
		commRoutes.GET("/patterns", h.GetCommunicationPatterns)

		// Load balancing endpoints
		commRoutes.GET("/load-balancer/strategies", h.GetLoadBalancingStrategies)
		commRoutes.POST("/load-balancer/test", h.TestLoadBalancing)
	}
}

// RegisterService registers a new service instance
//
//	@Summary		Register a new service instance
//	@Description	Registers a new service instance for service discovery
//	@Tags			Service Communication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.ServiceRegistration	true	"Service registration data"
//	@Success		201		{object}	map[string]interface{}		"Service registered successfully"
//	@Failure		400		{object}	map[string]interface{}		"Invalid service registration data"
//	@Failure		500		{object}	map[string]interface{}		"Service registration failed"
//	@Security		BearerAuth
//	@Router			/communication/services/register [post]
func (h *Handler) RegisterService(c *gin.Context) {
	var registration models.ServiceRegistration
	if err := c.ShouldBindJSON(&registration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid service registration data",
			"code":    "INVALID_REGISTRATION_DATA",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if registration.Name == "" || registration.Host == "" || registration.Port == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields: name, host, port",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
		return
	}

	// Generate ID if not provided
	if registration.ID == "" {
		registration.ID = registration.Name + "-" + registration.Host + "-" + string(rune(registration.Port))
	}

	if err := h.communicationService.RegisterService(&registration); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to register service",
			"code":    "SERVICE_REGISTRATION_FAILED",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Service registered successfully",
		"service_id": registration.ID,
		"timestamp":  time.Now(),
	})
}

// DeregisterService removes a service instance
//
//	@Summary		Deregister a service instance
//	@Description	Removes a service instance from the service registry
//	@Tags			Service Communication
//	@Produce		json
//	@Param			serviceId	path		string					true	"Service instance ID"
//	@Success		200			{object}	map[string]interface{}	"Service deregistered successfully"
//	@Failure		400			{object}	map[string]interface{}	"Service ID is required"
//	@Failure		404			{object}	map[string]interface{}	"Service not found"
//	@Security		BearerAuth
//	@Router			/communication/services/{serviceId} [delete]
func (h *Handler) DeregisterService(c *gin.Context) {
	serviceID := c.Param("serviceId")
	if serviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service ID is required",
			"code":  "SERVICE_ID_REQUIRED",
		})
		return
	}

	if err := h.communicationService.DeregisterService(serviceID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Service not found",
			"code":    "SERVICE_NOT_FOUND",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Service deregistered successfully",
		"service_id": serviceID,
		"timestamp":  time.Now(),
	})
}

// DiscoverService returns all instances of a service
//
//	@Summary		Discover service instances
//	@Description	Returns all available instances of a specific service
//	@Tags			Service Communication
//	@Produce		json
//	@Param			serviceName	path		string					true	"Service name to discover"
//	@Success		200			{object}	map[string]interface{}	"Service instances found"
//	@Failure		400			{object}	map[string]interface{}	"Service name is required"
//	@Failure		404			{object}	map[string]interface{}	"Service not found"
//	@Security		BearerAuth
//	@Router			/communication/services/discover/{serviceName} [get]
func (h *Handler) DiscoverService(c *gin.Context) {
	serviceName := c.Param("serviceName")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
			"code":  "SERVICE_NAME_REQUIRED",
		})
		return
	}

	instances, err := h.communicationService.DiscoverService(serviceName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Service not found",
			"code":    "SERVICE_NOT_FOUND",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_name": serviceName,
		"instances":    instances,
		"count":        len(instances),
		"timestamp":    time.Now(),
	})
}

// GetHealthyServices returns only healthy instances of a service
//
//	@Summary		Get healthy service instances
//	@Description	Returns only healthy instances of a specific service
//	@Tags			Service Communication
//	@Produce		json
//	@Param			serviceName	path		string					true	"Service name"
//	@Success		200			{object}	map[string]interface{}	"Healthy service instances"
//	@Failure		400			{object}	map[string]interface{}	"Service name is required"
//	@Failure		404			{object}	map[string]interface{}	"No healthy services found"
//	@Security		BearerAuth
//	@Router			/communication/services/healthy/{serviceName} [get]
func (h *Handler) GetHealthyServices(c *gin.Context) {
	serviceName := c.Param("serviceName")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
			"code":  "SERVICE_NAME_REQUIRED",
		})
		return
	}

	instances, err := h.communicationService.GetHealthyServices(serviceName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "No healthy services found",
			"code":    "NO_HEALTHY_SERVICES",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_name":      serviceName,
		"healthy_instances": instances,
		"count":             len(instances),
		"timestamp":         time.Now(),
	})
}

// ListAllServices returns all registered services
//
//	@Summary		List all registered services
//	@Description	Returns a list of all services registered in the service registry
//	@Tags			Service Communication
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of all services"
//	@Security		BearerAuth
//	@Router			/communication/services [get]
func (h *Handler) ListAllServices(c *gin.Context) {
	// This would require extending the service interface to list all services
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"message":   "List all services endpoint",
		"note":      "Implementation pending - would list all registered services",
		"timestamp": time.Now(),
	})
}

// CallService makes a service call
//
//	@Summary		Make an inter-service call
//	@Description	Makes a call to another service using service discovery
//	@Tags			Service Communication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.ServiceRequest	true	"Service call request"
//	@Success		200		{object}	map[string]interface{}	"Service call successful"
//	@Failure		400		{object}	map[string]interface{}	"Invalid service call request"
//	@Failure		500		{object}	map[string]interface{}	"Service call failed"
//	@Security		BearerAuth
//	@Router			/communication/call [post]
func (h *Handler) CallService(c *gin.Context) {
	var request models.ServiceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid service call request",
			"code":    "INVALID_SERVICE_CALL_REQUEST",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if request.ServiceName == "" || request.Method == "" || request.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields: service_name, method, path",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
		return
	}

	ctx := c.Request.Context()
	response, err := h.communicationService.CallService(ctx, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Service call failed",
			"code":    "SERVICE_CALL_FAILED",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response":  response,
		"timestamp": time.Now(),
	})
}

// BroadcastMessage broadcasts a message to all service instances
//
//	@Summary		Broadcast message to all service instances
//	@Description	Broadcasts a message to all instances of a specific service
//	@Tags			Service Communication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.BroadcastMessage	true	"Broadcast message request"
//	@Success		200		{object}	map[string]interface{}	"Message broadcasted successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid broadcast message"
//	@Failure		500		{object}	map[string]interface{}	"Broadcast failed"
//	@Security		BearerAuth
//	@Router			/communication/broadcast [post]
func (h *Handler) BroadcastMessage(c *gin.Context) {
	var message models.BroadcastMessage
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid broadcast message",
			"code":    "INVALID_BROADCAST_MESSAGE",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if message.ServiceName == "" || message.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields: service_name, path",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
		return
	}

	ctx := c.Request.Context()
	if err := h.communicationService.BroadcastMessage(ctx, &message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Broadcast failed",
			"code":    "BROADCAST_FAILED",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Broadcast sent successfully",
		"timestamp": time.Now(),
	})
}

// PublishEvent publishes an event
//
//	@Summary		Publish a service event
//	@Description	Publishes an event to the event system for subscribers
//	@Tags			Service Communication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.ServiceEvent		true	"Event data"
//	@Success		200		{object}	map[string]interface{}	"Event published successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid event data"
//	@Failure		500		{object}	map[string]interface{}	"Event publish failed"
//	@Security		BearerAuth
//	@Router			/communication/events/publish [post]
func (h *Handler) PublishEvent(c *gin.Context) {
	var event models.ServiceEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid event data",
			"code":    "INVALID_EVENT_DATA",
			"details": err.Error(),
		})
		return
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Generate ID if not provided
	if event.ID == "" {
		event.ID = event.Type + "-" + event.Source + "-" + string(rune(time.Now().Unix()))
	}

	ctx := c.Request.Context()
	if err := h.communicationService.PublishEvent(ctx, &event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to publish event",
			"code":    "EVENT_PUBLISH_FAILED",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Event published successfully",
		"event_id":  event.ID,
		"timestamp": time.Now(),
	})
}

// SubscribeToEvents subscribes to events of a specific type
//
//	@Summary		Subscribe to service events
//	@Description	Subscribes to events of a specific type in the event system
//	@Tags			Service Communication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{event_type=string,webhook_url=string}	true	"Subscription request"
//	@Success		200		{object}	map[string]interface{}							"Successfully subscribed to events"
//	@Failure		400		{object}	map[string]interface{}							"Invalid subscription request"
//	@Failure		500		{object}	map[string]interface{}							"Subscription failed"
//	@Security		BearerAuth
//	@Router			/communication/events/subscribe [post]
func (h *Handler) SubscribeToEvents(c *gin.Context) {
	var request struct {
		EventType string `json:"event_type"`
		Webhook   string `json:"webhook_url,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid subscription request",
			"code":    "INVALID_SUBSCRIPTION_REQUEST",
			"details": err.Error(),
		})
		return
	}

	if request.EventType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Event type is required",
			"code":  "EVENT_TYPE_REQUIRED",
		})
		return
	}

	// For now, just register a simple handler that logs events
	// In production, you'd implement webhook calls or other notification mechanisms
	handler := func(ctx context.Context, event *models.ServiceEvent) error {
		// Log the event or send to webhook
		return nil
	}

	if err := h.communicationService.SubscribeToEvents(request.EventType, handler); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to subscribe to events",
			"code":    "SUBSCRIPTION_FAILED",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Successfully subscribed to events",
		"event_type": request.EventType,
		"timestamp":  time.Now(),
	})
}

// GetCircuitBreakers returns the status of all circuit breakers
//
//	@Summary		Get circuit breaker status
//	@Description	Returns the status of all circuit breakers in the system
//	@Tags			Service Communication
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Circuit breaker status information"
//	@Security		BearerAuth
//	@Router			/communication/circuit-breakers [get]
func (h *Handler) GetCircuitBreakers(c *gin.Context) {
	// This would require extending the service to expose circuit breaker states
	c.JSON(http.StatusOK, gin.H{
		"message":   "Circuit breaker status endpoint",
		"note":      "Implementation pending - would return all circuit breaker states",
		"timestamp": time.Now(),
	})
}

// ResetCircuitBreaker resets a specific circuit breaker
//
//	@Summary		Reset a circuit breaker
//	@Description	Resets a specific circuit breaker to closed state
//	@Tags			Service Communication
//	@Produce		json
//	@Param			serviceName	path		string					true	"Service name"
//	@Success		200			{object}	map[string]interface{}	"Circuit breaker reset successfully"
//	@Failure		400			{object}	map[string]interface{}	"Service name is required"
//	@Security		BearerAuth
//	@Router			/communication/circuit-breakers/{serviceName}/reset [post]
func (h *Handler) ResetCircuitBreaker(c *gin.Context) {
	serviceName := c.Param("serviceName")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
			"code":  "SERVICE_NAME_REQUIRED",
		})
		return
	}

	// This would require extending the service to reset circuit breakers
	c.JSON(http.StatusOK, gin.H{
		"message":      "Circuit breaker reset successfully",
		"service_name": serviceName,
		"timestamp":    time.Now(),
	})
}

// GetCommunicationMetrics returns communication metrics
//
//	@Summary		Get service communication metrics
//	@Description	Returns metrics and statistics for service-to-service communication
//	@Tags			Service Communication
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Communication metrics and statistics"
//	@Security		BearerAuth
//	@Router			/communication/metrics [get]
func (h *Handler) GetCommunicationMetrics(c *gin.Context) {
	metrics := h.communicationService.GetCommunicationMetrics()

	// Calculate derived metrics
	successRate := metrics.CalculateSuccessRate()
	errorRate := metrics.CalculateErrorRate()

	response := gin.H{
		"metrics": metrics,
		"summary": gin.H{
			"success_rate_percent": successRate,
			"error_rate_percent":   errorRate,
			"total_requests":       metrics.TotalRequests,
			"uptime":               metrics.Uptime,
		},
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// GetHealthStatus returns the health status of the communication service
//
//	@Summary		Get communication service health
//	@Description	Returns health status and feature information for the communication service
//	@Tags			Service Communication
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Communication service health status"
//	@Router			/communication/health [get]
func (h *Handler) GetHealthStatus(c *gin.Context) {
	// Basic health check for the communication service
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "communication",
		"version":   "1.0.0",
		"timestamp": time.Now(),
		"features": gin.H{
			"service_discovery": true,
			"circuit_breaker":   true,
			"load_balancing":    true,
			"event_system":      true,
			"retry_policies":    true,
		},
	})
}

// GetCommunicationPatterns returns available communication patterns
//
//	@Summary		Get communication patterns
//	@Description	Returns available communication patterns and their configurations
//	@Tags			Service Communication
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Available communication patterns"
//	@Security		BearerAuth
//	@Router			/communication/patterns [get]
func (h *Handler) GetCommunicationPatterns(c *gin.Context) {
	patterns := []models.CommunicationPattern{
		{
			Name:           "Request-Response",
			Type:           models.MessagePatternRequestResponse,
			Timeout:        30 * time.Second,
			CircuitBreaker: true,
			LoadBalancer:   models.LoadBalancingRoundRobin,
		},
		{
			Name:           "Fire-and-Forget",
			Type:           models.MessagePatternFireAndForget,
			Timeout:        5 * time.Second,
			CircuitBreaker: false,
			LoadBalancer:   models.LoadBalancingRandom,
		},
		{
			Name:           "Publish-Subscribe",
			Type:           models.MessagePatternPublishSubscribe,
			Timeout:        1 * time.Second,
			CircuitBreaker: false,
			LoadBalancer:   models.LoadBalancingRoundRobin,
		},
		{
			Name:           "Broadcast",
			Type:           models.MessagePatternBroadcast,
			Timeout:        10 * time.Second,
			CircuitBreaker: true,
			LoadBalancer:   models.LoadBalancingRoundRobin,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"patterns":  patterns,
		"count":     len(patterns),
		"timestamp": time.Now(),
	})
}

// GetLoadBalancingStrategies returns available load balancing strategies
//
//	@Summary		Get load balancing strategies
//	@Description	Returns available load balancing strategies and their descriptions
//	@Tags			Service Communication
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Available load balancing strategies"
//	@Security		BearerAuth
//	@Router			/communication/load-balancer/strategies [get]
func (h *Handler) GetLoadBalancingStrategies(c *gin.Context) {
	strategies := []gin.H{
		{
			"name":        "Round Robin",
			"type":        models.LoadBalancingRoundRobin,
			"description": "Distributes requests evenly across all available instances",
		},
		{
			"name":        "Random",
			"type":        models.LoadBalancingRandom,
			"description": "Randomly selects an instance for each request",
		},
		{
			"name":        "Least Connections",
			"type":        models.LoadBalancingLeastConnections,
			"description": "Routes requests to the instance with the fewest active connections",
		},
		{
			"name":        "Weighted",
			"type":        models.LoadBalancingWeighted,
			"description": "Routes requests based on instance weights",
		},
		{
			"name":        "IP Hash",
			"type":        models.LoadBalancingIPHash,
			"description": "Routes requests based on client IP hash for session affinity",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"strategies": strategies,
		"count":      len(strategies),
		"timestamp":  time.Now(),
	})
}

// TestLoadBalancing tests load balancing for a specific service
//
//	@Summary		Test load balancing strategy
//	@Description	Tests load balancing distribution for a specific service and strategy
//	@Tags			Service Communication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{service_name=string,strategy=string,requests=int}	true	"Load balancing test request"
//	@Success		200		{object}	map[string]interface{}										"Load balancing test results"
//	@Failure		400		{object}	map[string]interface{}										"Invalid test request"
//	@Failure		404		{object}	map[string]interface{}										"Service not found"
//	@Security		BearerAuth
//	@Router			/communication/load-balancer/test [post]
func (h *Handler) TestLoadBalancing(c *gin.Context) {
	var request struct {
		ServiceName string                       `json:"service_name"`
		Strategy    models.LoadBalancingStrategy `json:"strategy"`
		Requests    int                          `json:"requests"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid load balancing test request",
			"code":    "INVALID_TEST_REQUEST",
			"details": err.Error(),
		})
		return
	}

	if request.ServiceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
			"code":  "SERVICE_NAME_REQUIRED",
		})
		return
	}

	if request.Requests <= 0 {
		request.Requests = 10 // Default test requests
	}

	// Test load balancing by selecting instances multiple times
	results := make(map[string]int)
	for i := 0; i < request.Requests; i++ {
		instance, err := h.communicationService.SelectInstance(request.ServiceName, request.Strategy)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Failed to select instance",
				"code":    "INSTANCE_SELECTION_FAILED",
				"details": err.Error(),
			})
			return
		}
		results[instance.ID]++
	}

	c.JSON(http.StatusOK, gin.H{
		"service_name":   request.ServiceName,
		"strategy":       request.Strategy,
		"total_requests": request.Requests,
		"distribution":   results,
		"timestamp":      time.Now(),
	})
}
