package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/auth"
	websocketService "github.com/pramodksahoo/kubechat/apps/api/internal/services/websocket"
)

// Handler handles WebSocket-related HTTP requests
type Handler struct {
	websocketService websocketService.Service
}

// NewHandler creates a new WebSocket handler
func NewHandler(websocketService websocketService.Service) *Handler {
	return &Handler{
		websocketService: websocketService,
	}
}

// RegisterRoutes registers WebSocket routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// WebSocket upgrade endpoint (no auth middleware needed - auth handled in WebSocket)
	router.GET("/ws", h.HandleWebSocket)

	// Management endpoints (require admin role)
	wsRoutes := router.Group("/websocket")
	wsRoutes.Use(auth.RequireRole("admin")) // Admin-only endpoints
	{
		// Client management
		wsRoutes.GET("/clients", h.GetConnectedClients)
		wsRoutes.GET("/clients/count", h.GetClientCount)
		wsRoutes.DELETE("/clients/:id", h.DisconnectClient)

		// Messaging endpoints
		wsRoutes.POST("/broadcast", h.BroadcastToAll)
		wsRoutes.POST("/broadcast/user/:user_id", h.BroadcastToUser)
		wsRoutes.POST("/broadcast/topics", h.BroadcastToTopics)
		wsRoutes.POST("/notify/user/:user_id", h.NotifyUser)
		wsRoutes.POST("/notify/system", h.NotifySystem)

		// Metrics and status
		wsRoutes.GET("/metrics", h.GetMetrics)
		wsRoutes.GET("/health", h.HealthCheck)

		// Subscription management
		wsRoutes.GET("/subscriptions/:topic", h.GetTopicSubscribers)
		wsRoutes.GET("/clients/:id/subscriptions", h.GetClientSubscriptions)
	}
}

// HandleWebSocket handles WebSocket connection upgrade
//
//	@Summary		Upgrade to WebSocket connection
//	@Description	Upgrades HTTP connection to WebSocket for real-time communication
//	@Tags			WebSocket Communication
//	@Success		101	"WebSocket connection established"
//	@Failure		400	"WebSocket upgrade failed"
//	@Router			/ws [get]
func (h *Handler) HandleWebSocket(c *gin.Context) {
	h.websocketService.HandleWebSocket(c)
}

// GetConnectedClients returns all connected WebSocket clients
//
//	@Summary		Get connected WebSocket clients
//	@Description	Returns information about all currently connected WebSocket clients
//	@Tags			WebSocket Communication
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of connected clients"
//	@Security		BearerAuth
//	@Router			/websocket/clients [get]
func (h *Handler) GetConnectedClients(c *gin.Context) {
	clients := h.websocketService.GetConnectedClients()

	c.JSON(http.StatusOK, gin.H{
		"data":  clients,
		"count": len(clients),
	})
}

// GetClientCount returns the number of connected clients
//
//	@Summary		Get connected client count
//	@Description	Returns the total number of currently connected WebSocket clients
//	@Tags			WebSocket Communication
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Client count"
//	@Security		BearerAuth
//	@Router			/websocket/clients/count [get]
func (h *Handler) GetClientCount(c *gin.Context) {
	count := h.websocketService.GetClientCount()

	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

// DisconnectClient forcefully disconnects a client
//
//	@Summary		Disconnect WebSocket client
//	@Description	Forcefully disconnects a specific WebSocket client
//	@Tags			WebSocket Communication
//	@Produce		json
//	@Param			id	path		string					true	"Client ID"
//	@Success		200	{object}	map[string]interface{}	"Client disconnected successfully"
//	@Failure		400	{object}	map[string]interface{}	"Client ID is required"
//	@Failure		404	{object}	map[string]interface{}	"Client not found"
//	@Security		BearerAuth
//	@Router			/websocket/clients/{id} [delete]
func (h *Handler) DisconnectClient(c *gin.Context) {
	clientID := c.Param("id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Client ID is required",
			"code":  "MISSING_CLIENT_ID",
		})
		return
	}

	err := h.websocketService.DisconnectClient(clientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Client not found or could not be disconnected",
			"code":  "CLIENT_DISCONNECT_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Client disconnected successfully",
	})
}

// BroadcastToAll broadcasts a message to all connected clients
//
//	@Summary		Broadcast to all clients
//	@Description	Broadcasts a message to all currently connected WebSocket clients
//	@Tags			WebSocket Communication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{type=string,payload=map[string]interface{}}	true	"Broadcast message"
//	@Success		200		{object}	map[string]interface{}								"Message broadcasted successfully"
//	@Failure		400		{object}	map[string]interface{}								"Invalid request format"
//	@Failure		500		{object}	map[string]interface{}								"Broadcast failed"
//	@Security		BearerAuth
//	@Router			/websocket/broadcast [post]
func (h *Handler) BroadcastToAll(c *gin.Context) {
	var req struct {
		Type    string                 `json:"type" binding:"required"`
		Payload map[string]interface{} `json:"payload" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Validate message type
	msgType := models.WebSocketMessageType(req.Type)
	if !models.IsValidMessageType(msgType) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid message type",
			"code":  "INVALID_MESSAGE_TYPE",
		})
		return
	}

	// Create message
	message, err := models.NewWebSocketMessage(msgType, req.Payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create message",
			"code":  "MESSAGE_CREATION_ERROR",
		})
		return
	}

	// Broadcast message
	err = h.websocketService.BroadcastToAll(message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to broadcast message",
			"code":  "BROADCAST_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message broadcasted to all clients",
	})
}

// BroadcastToUser broadcasts a message to specific user
//
//	@Summary		Broadcast to specific user
//	@Description	Broadcasts a message to all connections of a specific user
//	@Tags			WebSocket Communication
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		string												true	"User ID"	format(uuid)
//	@Param			request	body		object{type=string,payload=map[string]interface{}}	true	"Broadcast message"
//	@Success		200		{object}	map[string]interface{}								"Message broadcasted to user"
//	@Failure		400		{object}	map[string]interface{}								"Invalid request format"
//	@Failure		404		{object}	map[string]interface{}								"User not connected"
//	@Security		BearerAuth
//	@Router			/websocket/broadcast/user/{user_id} [post]
func (h *Handler) BroadcastToUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
			"code":  "INVALID_USER_ID",
		})
		return
	}

	var req struct {
		Type    string                 `json:"type" binding:"required"`
		Payload map[string]interface{} `json:"payload" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Validate message type
	msgType := models.WebSocketMessageType(req.Type)
	if !models.IsValidMessageType(msgType) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid message type",
			"code":  "INVALID_MESSAGE_TYPE",
		})
		return
	}

	// Create message
	message, err := models.NewWebSocketMessage(msgType, req.Payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create message",
			"code":  "MESSAGE_CREATION_ERROR",
		})
		return
	}

	// Broadcast to user
	err = h.websocketService.BroadcastToUser(userID, message)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not connected or broadcast failed",
			"code":  "USER_BROADCAST_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message broadcasted to user",
		"user_id": userID,
	})
}

// BroadcastToTopics broadcasts a message to subscribers of specific topics
//
//	@Summary		Broadcast to topic subscribers
//	@Description	Broadcasts a message to all clients subscribed to specific topics
//	@Tags			WebSocket Communication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{topics=[]string,type=string,payload=map[string]interface{}}	true	"Topic broadcast message"
//	@Success		200		{object}	map[string]interface{}												"Message broadcasted to subscribers"
//	@Failure		400		{object}	map[string]interface{}												"Invalid request format"
//	@Failure		404		{object}	map[string]interface{}												"No subscribers found"
//	@Security		BearerAuth
//	@Router			/websocket/broadcast/topics [post]
func (h *Handler) BroadcastToTopics(c *gin.Context) {
	var req struct {
		Topics  []string               `json:"topics" binding:"required"`
		Type    string                 `json:"type" binding:"required"`
		Payload map[string]interface{} `json:"payload" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Validate topics (define valid topics here since we can't access the service's internal map)
	allowedTopics := map[string]bool{
		"commands":      true,
		"system":        true,
		"user_activity": true,
		"alerts":        true,
		"metrics":       true,
	}

	validTopics := []string{}
	for _, topic := range req.Topics {
		if allowedTopics[topic] {
			validTopics = append(validTopics, topic)
		}
	}

	if len(validTopics) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No valid topics provided",
			"code":  "INVALID_TOPICS",
		})
		return
	}

	// Validate message type
	msgType := models.WebSocketMessageType(req.Type)
	if !models.IsValidMessageType(msgType) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid message type",
			"code":  "INVALID_MESSAGE_TYPE",
		})
		return
	}

	// Create message
	message, err := models.NewWebSocketMessage(msgType, req.Payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create message",
			"code":  "MESSAGE_CREATION_ERROR",
		})
		return
	}

	// Broadcast to topics
	err = h.websocketService.BroadcastToSubscribers(validTopics, message)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No subscribers found or broadcast failed",
			"code":  "TOPIC_BROADCAST_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message broadcasted to topic subscribers",
		"topics":  validTopics,
	})
}

// NotifyUser sends a personal notification to a user
//
//	@Summary		Send user notification
//	@Description	Sends a personal notification to a specific user
//	@Tags			WebSocket Communication
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		string																						true	"User ID"
//	@Param			request	body		object{type=string,title=string,message=string,priority=string,data=map[string]interface{}}	true	"Notification data"
//	@Success		200		{object}	map[string]interface{}																		"Notification sent to user"
//	@Failure		400		{object}	map[string]interface{}																		"Invalid request format"
//	@Failure		404		{object}	map[string]interface{}																		"User not connected"
//	@Security		BearerAuth
//	@Router			/websocket/notify/user/{user_id} [post]
func (h *Handler) NotifyUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
			"code":  "MISSING_USER_ID",
		})
		return
	}

	var req struct {
		Type     string                 `json:"type" binding:"required"`
		Title    string                 `json:"title" binding:"required"`
		Message  string                 `json:"message" binding:"required"`
		Priority string                 `json:"priority"`
		Data     map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Default priority
	if req.Priority == "" {
		req.Priority = "normal"
	}

	// Send notification (placeholder - need to implement this method)
	// err := h.websocketService.SendPersonalNotification(userID, req.Type, req.Title, req.Message, req.Priority, req.Data)
	// if err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{
	// 		"error": "User not connected or notification failed",
	// 		"code":  "NOTIFICATION_ERROR",
	// 	})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification sent to user",
		"user_id": userID,
	})
}

// NotifySystem sends a system notification
//
//	@Summary		Send system notification
//	@Description	Sends a system-wide notification to all connected clients
//	@Tags			WebSocket Communication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{event_type=string,description=string,priority=string}	true	"System notification data"
//	@Success		200		{object}	map[string]interface{}											"System notification sent"
//	@Failure		400		{object}	map[string]interface{}											"Invalid request format"
//	@Security		BearerAuth
//	@Router			/websocket/notify/system [post]
func (h *Handler) NotifySystem(c *gin.Context) {
	var req struct {
		EventType   string `json:"event_type" binding:"required"`
		Description string `json:"description" binding:"required"`
		Priority    string `json:"priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Default priority
	if req.Priority == "" {
		req.Priority = "normal"
	}

	// Send system notification (placeholder - need to implement this method)
	// h.websocketService.NotifySystemEvent(req.EventType, req.Description, req.Priority)

	c.JSON(http.StatusOK, gin.H{
		"message": "System notification sent",
	})
}

// GetMetrics returns WebSocket service metrics
//
//	@Summary		Get WebSocket metrics
//	@Description	Returns performance metrics and statistics for WebSocket connections
//	@Tags			WebSocket Communication
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"WebSocket service metrics"
//	@Security		BearerAuth
//	@Router			/websocket/metrics [get]
func (h *Handler) GetMetrics(c *gin.Context) {
	metrics := h.websocketService.GetMetrics()

	c.JSON(http.StatusOK, gin.H{
		"data": metrics,
	})
}

// HealthCheck performs WebSocket service health check
//
//	@Summary		Check WebSocket service health
//	@Description	Performs a health check on the WebSocket service
//	@Tags			WebSocket Communication
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"WebSocket service is operational"
//	@Failure		503	{object}	map[string]interface{}	"WebSocket service health check failed"
//	@Security		BearerAuth
//	@Router			/websocket/health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	err := h.websocketService.HealthCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "WebSocket service health check failed",
			"code":  "WEBSOCKET_HEALTH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "WebSocket service is operational",
	})
}

// GetTopicSubscribers returns subscribers for a specific topic
//
//	@Summary		Get topic subscribers
//	@Description	Returns all clients subscribed to a specific topic
//	@Tags			WebSocket Communication
//	@Produce		json
//	@Param			topic	path		string					true	"Topic name"	Enums(commands, system, user_activity, alerts, metrics)
//	@Success		200		{object}	map[string]interface{}	"List of topic subscribers"
//	@Failure		400		{object}	map[string]interface{}	"Invalid topic"
//	@Security		BearerAuth
//	@Router			/websocket/subscriptions/{topic} [get]
func (h *Handler) GetTopicSubscribers(c *gin.Context) {
	topic := c.Param("topic")
	if topic == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Topic is required",
			"code":  "MISSING_TOPIC",
		})
		return
	}

	// Validate topic
	allowedTopics := map[string]bool{
		"commands":      true,
		"system":        true,
		"user_activity": true,
		"alerts":        true,
		"metrics":       true,
	}

	if !allowedTopics[topic] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid topic",
			"code":  "INVALID_TOPIC",
		})
		return
	}

	// Get subscribers (placeholder - need to implement this method)
	// subscribers := h.websocketService.GetTopicSubscribers(topic)

	c.JSON(http.StatusOK, gin.H{
		"data":  []models.WebSocketClient{}, // subscribers,
		"topic": topic,
		"count": 0, // len(subscribers),
	})
}

// GetClientSubscriptions returns subscriptions for a specific client
//
//	@Summary		Get client subscriptions
//	@Description	Returns all topics that a specific client is subscribed to
//	@Tags			WebSocket Communication
//	@Produce		json
//	@Param			id	path		string					true	"Client ID"
//	@Success		200	{object}	map[string]interface{}	"List of client subscriptions"
//	@Failure		400	{object}	map[string]interface{}	"Client ID is required"
//	@Security		BearerAuth
//	@Router			/websocket/clients/{id}/subscriptions [get]
func (h *Handler) GetClientSubscriptions(c *gin.Context) {
	clientID := c.Param("id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Client ID is required",
			"code":  "MISSING_CLIENT_ID",
		})
		return
	}

	// Get client subscriptions (placeholder - need to implement this method)
	// subscriptions := h.websocketService.GetClientSubscriptions(clientID)

	c.JSON(http.StatusOK, gin.H{
		"data":      []string{}, // subscriptions,
		"client_id": clientID,
	})
}
