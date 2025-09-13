package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// run starts the hub and handles client registration/unregistration and message broadcasting
func (h *hub) run() {
	defer h.wg.Done()

	h.logger.Info("WebSocket hub started")

	// Start periodic tasks
	heartbeatTicker := time.NewTicker(h.config.HeartbeatInterval)
	defer heartbeatTicker.Stop()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-heartbeatTicker.C:
			h.performHeartbeat()

		case <-h.ctx.Done():
			h.logger.Info("WebSocket hub shutting down")
			h.closeAllClients()
			return
		}
	}
}

// registerClient registers a new client connection
func (h *hub) registerClient(client *client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Add to clients map
	h.clients[client.client.ID] = client

	// Update metrics
	h.metrics.ConnectedClients++
	h.metrics.TotalConnections++

	h.logger.Info("Client registered",
		"client_id", client.client.ID,
		"total_clients", len(h.clients))

	// Send connection acknowledgment
	ackMessage, _ := models.NewWebSocketMessage(models.WSMsgTypeStatus, models.StatusPayload{
		Type:      "connection",
		Status:    "connected",
		Message:   "WebSocket connection established",
		Timestamp: time.Now(),
	})

	select {
	case client.send <- ackMessage:
	default:
		h.logger.Warn("Failed to send connection ack", "client_id", client.client.ID)
	}
}

// unregisterClient unregisters a client connection
func (h *hub) unregisterClient(client *client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, exists := h.clients[client.client.ID]; !exists {
		return
	}

	// Remove from clients map
	delete(h.clients, client.client.ID)

	// Remove from user clients map
	if client.client.UserID != uuid.Nil {
		userClients := h.userClients[client.client.UserID]
		for i, c := range userClients {
			if c.client.ID == client.client.ID {
				h.userClients[client.client.UserID] = append(userClients[:i], userClients[i+1:]...)
				break
			}
		}
		// Remove user entry if no more clients
		if len(h.userClients[client.client.UserID]) == 0 {
			delete(h.userClients, client.client.UserID)
		}
	}

	// Remove from topic subscriptions
	for topic, subscribers := range h.topicSubscribers {
		for i, c := range subscribers {
			if c.client.ID == client.client.ID {
				h.topicSubscribers[topic] = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}
		// Remove topic if no more subscribers
		if len(h.topicSubscribers[topic]) == 0 {
			delete(h.topicSubscribers, topic)
		}
	}

	// Cancel any active commands
	client.cancelActiveCommands()

	// Close send channel
	close(client.send)

	// Update metrics
	h.metrics.ConnectedClients--

	h.logger.Info("Client unregistered",
		"client_id", client.client.ID,
		"user_id", client.client.UserID,
		"total_clients", len(h.clients))

	// Log audit event for disconnection
	if h.auditService != nil && client.client.UserID != uuid.Nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			h.auditService.LogSecurityEvent(
				ctx,
				"websocket_disconnect",
				fmt.Sprintf("WebSocket client disconnected: %s", client.client.ID),
				&client.client.UserID,
				"info",
				nil,
			)
		}()
	}
}

// broadcastMessage broadcasts a message to all connected clients
func (h *hub) broadcastMessage(message *models.WebSocketMessage) {
	h.mutex.RLock()
	clients := make([]*client, 0, len(h.clients))
	for _, client := range h.clients {
		clients = append(clients, client)
	}
	h.mutex.RUnlock()

	for _, c := range clients {
		select {
		case c.send <- message:
			h.metrics.MessagesSent++
		default:
			h.logger.Warn("Client send channel full during broadcast", "client_id", c.client.ID)
			// Client is not responding, close it
			go func(clientToClose *client) {
				h.unregister <- clientToClose
			}(c)
		}
	}
}

// performHeartbeat sends ping to all clients and removes stale connections
func (h *hub) performHeartbeat() {
	h.mutex.RLock()
	clients := make([]*client, 0, len(h.clients))
	for _, c := range h.clients {
		clients = append(clients, c)
	}
	h.mutex.RUnlock()

	now := time.Now()
	staleClients := []*client{}

	for _, c := range clients {
		// Check if client is stale
		if now.Sub(c.client.LastPing) > h.config.ClientTimeout {
			staleClients = append(staleClients, c)
			continue
		}

		// Send ping
		pingMessage, _ := models.NewWebSocketMessage(models.WSMsgTypePing, map[string]interface{}{
			"timestamp": now,
		})

		select {
		case c.send <- pingMessage:
		default:
			staleClients = append(staleClients, c)
		}
	}

	// Remove stale clients
	for _, staleClient := range staleClients {
		h.logger.Info("Removing stale client", "client_id", staleClient.client.ID)
		h.unregister <- staleClient
	}
}

// closeAllClients closes all client connections
func (h *hub) closeAllClients() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for _, c := range h.clients {
		c.cancel()
		close(c.send)
	}

	h.clients = make(map[string]*client)
	h.userClients = make(map[uuid.UUID][]*client)
	h.topicSubscribers = make(map[string][]*client)
}

// readMessages handles incoming WebSocket messages from a client
func (c *client) readMessages() {
	defer func() {
		c.cancel()
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// Read message from WebSocket
			_, messageBytes, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Error("WebSocket error", "client_id", c.client.ID, "error", err)
				}
				return
			}

			// Parse message
			var message models.WebSocketMessage
			if err := json.Unmarshal(messageBytes, &message); err != nil {
				c.sendError("INVALID_MESSAGE", "Failed to parse message", "")
				continue
			}

			// Update metrics
			c.hub.metrics.MessagesReceived++

			// Validate message type
			if !models.IsValidMessageType(message.Type) {
				c.sendError("INVALID_MESSAGE_TYPE", "Invalid message type", "")
				continue
			}

			// Handle message based on type
			if err := c.handleMessage(&message); err != nil {
				c.logger.Error("Error handling message",
					"client_id", c.client.ID,
					"message_type", message.Type,
					"error", err)
				c.sendError("MESSAGE_HANDLING_ERROR", err.Error(), "")
			}
		}
	}
}

// writeMessages handles outgoing WebSocket messages to a client
func (c *client) writeMessages() {
	ticker := time.NewTicker(c.hub.config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.config.WriteTimeout))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Serialize message
			messageBytes, err := json.Marshal(message)
			if err != nil {
				c.logger.Error("Failed to serialize message", "error", err)
				continue
			}

			// Send message
			if err := c.conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
				c.logger.Error("Failed to write message", "error", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.config.WriteTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}

// handleMessage processes incoming messages based on their type
func (c *client) handleMessage(message *models.WebSocketMessage) error {
	switch message.Type {
	case models.WSMsgTypeAuth:
		return c.handleAuth(message)
	case models.WSMsgTypeExecute:
		return c.handleExecute(message)
	case models.WSMsgTypeCancel:
		return c.handleCancel(message)
	case models.WSMsgTypeSubscribe:
		return c.handleSubscribe(message)
	case models.WSMsgTypeUnsubscribe:
		return c.handleUnsubscribe(message)
	case models.WSMsgTypePong:
		return c.handlePong(message)
	case models.WSMsgTypeHeartbeat:
		return c.handleHeartbeat(message)
	default:
		return fmt.Errorf("unhandled message type: %s", message.Type)
	}
}

// handleAuth processes authentication messages
func (c *client) handleAuth(message *models.WebSocketMessage) error {
	var payload models.AuthPayload
	if err := message.UnmarshalPayload(&payload); err != nil {
		return fmt.Errorf("invalid auth payload: %w", err)
	}

	// Validate JWT token
	claims, err := c.hub.authService.ValidateJWT(payload.Token)
	if err != nil {
		c.sendAuthResult(false, uuid.Nil, "", "", uuid.Nil, "Invalid token")
		return nil
	}

	// Extract information from claims struct
	userID := claims.UserID
	username := claims.Username
	role := claims.Role
	sessionID := claims.SessionID

	// Update client information
	c.client.UserID = userID
	c.client.Username = username
	c.client.Role = role
	c.client.SessionID = sessionID

	// Add to user clients map
	c.hub.mutex.Lock()
	c.hub.userClients[userID] = append(c.hub.userClients[userID], c)
	c.hub.mutex.Unlock()

	// Send success response
	c.sendAuthResult(true, userID, c.client.Username, c.client.Role, sessionID, "")

	// Log audit event
	if c.hub.auditService != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			c.hub.auditService.LogSecurityEvent(
				ctx,
				"websocket_auth",
				fmt.Sprintf("WebSocket client authenticated: %s", c.client.Username),
				&userID,
				"info",
				nil,
			)
		}()
	}

	c.logger.Info("Client authenticated",
		"client_id", c.client.ID,
		"user_id", userID,
		"username", c.client.Username)

	return nil
}

// sendAuthResult sends authentication result to client
func (c *client) sendAuthResult(success bool, userID uuid.UUID, username, role string, sessionID uuid.UUID, errorMsg string) {
	payload := models.AuthResultPayload{
		Success:   success,
		UserID:    userID,
		Username:  username,
		Role:      role,
		SessionID: sessionID,
		Error:     errorMsg,
	}

	message, _ := models.NewWebSocketMessage(models.WSMsgTypeAuthResult, payload)
	select {
	case c.send <- message:
	default:
		c.logger.Warn("Failed to send auth result", "client_id", c.client.ID)
	}
}

// sendError sends an error message to the client
func (c *client) sendError(code, message, details string) {
	payload := models.ErrorPayload{
		Code:    code,
		Message: message,
		Details: details,
	}

	errorMessage, _ := models.NewWebSocketMessage(models.WSMsgTypeError, payload)
	select {
	case c.send <- errorMessage:
		c.hub.metrics.ErrorCount++
	default:
		c.logger.Warn("Failed to send error message", "client_id", c.client.ID)
	}
}

// requireAuth checks if client is authenticated
func (c *client) requireAuth() bool {
	return c.hub.config.RequireAuth && c.client.UserID == uuid.Nil
}

// cancelActiveCommands cancels all active commands for this client
func (c *client) cancelActiveCommands() {
	c.commandsMutex.Lock()
	defer c.commandsMutex.Unlock()

	for _, cmd := range c.activeCommands {
		cmd.Status = "cancelled"
		if cmd.CompletedAt == nil {
			now := time.Now()
			cmd.CompletedAt = &now
		}
	}
}

// handlePong processes pong messages
func (c *client) handlePong(message *models.WebSocketMessage) error {
	c.client.LastPing = time.Now()
	return nil
}

// handleHeartbeat processes heartbeat messages
func (c *client) handleHeartbeat(message *models.WebSocketMessage) error {
	c.client.LastPing = time.Now()

	// Send heartbeat response
	response, _ := models.NewWebSocketMessage(models.WSMsgTypePong, map[string]interface{}{
		"timestamp": time.Now(),
		"client_id": c.client.ID,
	})

	select {
	case c.send <- response:
	default:
		return fmt.Errorf("failed to send heartbeat response")
	}

	return nil
}
