package websocket

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// Available subscription topics
const (
	TopicCommands     = "commands"
	TopicSystem       = "system"
	TopicUserActivity = "user_activity"
	TopicAlerts       = "alerts"
	TopicMetrics      = "metrics"
)

// ValidTopics defines all valid subscription topics
var ValidTopics = map[string]bool{
	TopicCommands:     true,
	TopicSystem:       true,
	TopicUserActivity: true,
	TopicAlerts:       true,
	TopicMetrics:      true,
}

// handleSubscribe processes subscription requests
func (c *client) handleSubscribe(message *models.WebSocketMessage) error {
	if c.requireAuth() {
		c.sendError("AUTH_REQUIRED", "Authentication required", "")
		return nil
	}

	var payload models.SubscribePayload
	if err := message.UnmarshalPayload(&payload); err != nil {
		return fmt.Errorf("invalid subscribe payload: %w", err)
	}

	// Validate topics
	validTopics := []string{}
	for _, topic := range payload.Topics {
		if ValidTopics[topic] {
			validTopics = append(validTopics, topic)
		} else {
			c.logger.Warn("Invalid subscription topic",
				"client_id", c.client.ID,
				"topic", topic)
		}
	}

	if len(validTopics) == 0 {
		c.sendError("INVALID_TOPICS", "No valid topics provided", "")
		return nil
	}

	// Add subscriptions
	c.hub.mutex.Lock()
	for _, topic := range validTopics {
		// Add client to topic subscribers
		c.hub.topicSubscribers[topic] = append(c.hub.topicSubscribers[topic], c)

		// Add topic to client subscriptions
		if !containsString(c.client.Subscriptions, topic) {
			c.client.Subscriptions = append(c.client.Subscriptions, topic)
		}
	}
	c.hub.mutex.Unlock()

	c.logger.Info("Client subscribed to topics",
		"client_id", c.client.ID,
		"user_id", c.client.UserID,
		"topics", validTopics)

	// Send confirmation
	statusPayload := models.StatusPayload{
		Type:    "subscription",
		Status:  "subscribed",
		Message: fmt.Sprintf("Subscribed to topics: %v", validTopics),
		Metadata: map[string]interface{}{
			"topics": validTopics,
		},
		Timestamp: time.Now(),
	}

	statusMessage, _ := models.NewWebSocketMessage(models.WSMsgTypeStatus, statusPayload)
	select {
	case c.send <- statusMessage:
	default:
		c.logger.Warn("Failed to send subscription confirmation", "client_id", c.client.ID)
	}

	return nil
}

// handleUnsubscribe processes unsubscription requests
func (c *client) handleUnsubscribe(message *models.WebSocketMessage) error {
	if c.requireAuth() {
		c.sendError("AUTH_REQUIRED", "Authentication required", "")
		return nil
	}

	var payload models.SubscribePayload // Reuse the same payload structure
	if err := message.UnmarshalPayload(&payload); err != nil {
		return fmt.Errorf("invalid unsubscribe payload: %w", err)
	}

	// Remove subscriptions
	c.hub.mutex.Lock()
	removedTopics := []string{}

	for _, topic := range payload.Topics {
		if ValidTopics[topic] {
			// Remove client from topic subscribers
			if subscribers, exists := c.hub.topicSubscribers[topic]; exists {
				for i, subscriber := range subscribers {
					if subscriber.client.ID == c.client.ID {
						c.hub.topicSubscribers[topic] = append(subscribers[:i], subscribers[i+1:]...)
						break
					}
				}

				// Remove topic if no more subscribers
				if len(c.hub.topicSubscribers[topic]) == 0 {
					delete(c.hub.topicSubscribers, topic)
				}
			}

			// Remove topic from client subscriptions
			for i, subscription := range c.client.Subscriptions {
				if subscription == topic {
					c.client.Subscriptions = append(c.client.Subscriptions[:i], c.client.Subscriptions[i+1:]...)
					break
				}
			}

			removedTopics = append(removedTopics, topic)
		}
	}
	c.hub.mutex.Unlock()

	c.logger.Info("Client unsubscribed from topics",
		"client_id", c.client.ID,
		"user_id", c.client.UserID,
		"topics", removedTopics)

	// Send confirmation
	statusPayload := models.StatusPayload{
		Type:    "subscription",
		Status:  "unsubscribed",
		Message: fmt.Sprintf("Unsubscribed from topics: %v", removedTopics),
		Metadata: map[string]interface{}{
			"topics": removedTopics,
		},
		Timestamp: time.Now(),
	}

	statusMessage, _ := models.NewWebSocketMessage(models.WSMsgTypeStatus, statusPayload)
	select {
	case c.send <- statusMessage:
	default:
		c.logger.Warn("Failed to send unsubscription confirmation", "client_id", c.client.ID)
	}

	return nil
}

// NotifyCommandExecution notifies subscribers about command execution events
func (h *hub) NotifyCommandExecution(userID string, command, status string) {
	if len(h.topicSubscribers[TopicCommands]) == 0 {
		return
	}

	payload := models.NotificationPayload{
		Type:     "command_execution",
		Title:    "Command Executed",
		Message:  fmt.Sprintf("User %s executed: %s (Status: %s)", userID, command, status),
		Priority: "normal",
		Data: map[string]interface{}{
			"user_id": userID,
			"command": command,
			"status":  status,
		},
	}

	message, _ := models.NewWebSocketMessage(models.WSMsgTypeNotification, payload)
	h.broadcastToTopic(TopicCommands, message)
}

// NotifySystemEvent notifies subscribers about system events
func (h *hub) NotifySystemEvent(eventType, description string, priority string) {
	if len(h.topicSubscribers[TopicSystem]) == 0 {
		return
	}

	payload := models.NotificationPayload{
		Type:     eventType,
		Title:    "System Event",
		Message:  description,
		Priority: priority,
		Data: map[string]interface{}{
			"event_type":  eventType,
			"description": description,
			"timestamp":   time.Now(),
		},
	}

	message, _ := models.NewWebSocketMessage(models.WSMsgTypeNotification, payload)
	h.broadcastToTopic(TopicSystem, message)
}

// NotifyUserActivity notifies subscribers about user activity
func (h *hub) NotifyUserActivity(userID, username, activity string) {
	if len(h.topicSubscribers[TopicUserActivity]) == 0 {
		return
	}

	payload := models.NotificationPayload{
		Type:     "user_activity",
		Title:    "User Activity",
		Message:  fmt.Sprintf("%s: %s", username, activity),
		Priority: "low",
		Data: map[string]interface{}{
			"user_id":   userID,
			"username":  username,
			"activity":  activity,
			"timestamp": time.Now(),
		},
	}

	message, _ := models.NewWebSocketMessage(models.WSMsgTypeNotification, payload)
	h.broadcastToTopic(TopicUserActivity, message)
}

// NotifyAlert sends alert notifications
func (h *hub) NotifyAlert(alertType, title, message, priority string, data map[string]interface{}) {
	if len(h.topicSubscribers[TopicAlerts]) == 0 {
		return
	}

	if data == nil {
		data = make(map[string]interface{})
	}
	data["alert_type"] = alertType
	data["timestamp"] = time.Now()

	payload := models.NotificationPayload{
		Type:     alertType,
		Title:    title,
		Message:  message,
		Priority: priority,
		Data:     data,
	}

	notificationMessage, _ := models.NewWebSocketMessage(models.WSMsgTypeNotification, payload)
	h.broadcastToTopic(TopicAlerts, notificationMessage)
}

// NotifyMetrics sends metrics updates
func (h *hub) NotifyMetrics(metrics *models.WebSocketMetrics) {
	if len(h.topicSubscribers[TopicMetrics]) == 0 {
		return
	}

	payload := models.NotificationPayload{
		Type:     "metrics_update",
		Title:    "Metrics Update",
		Message:  "WebSocket service metrics updated",
		Priority: "low",
		Data: map[string]interface{}{
			"metrics":   metrics,
			"timestamp": time.Now(),
		},
	}

	message, _ := models.NewWebSocketMessage(models.WSMsgTypeNotification, payload)
	h.broadcastToTopic(TopicMetrics, message)
}

// broadcastToTopic broadcasts a message to all subscribers of a specific topic
func (h *hub) broadcastToTopic(topic string, message *models.WebSocketMessage) {
	h.mutex.RLock()
	subscribers := make([]*client, len(h.topicSubscribers[topic]))
	copy(subscribers, h.topicSubscribers[topic])
	h.mutex.RUnlock()

	for _, client := range subscribers {
		select {
		case client.send <- message:
			h.metrics.MessagesSent++
		default:
			h.logger.Warn("Client send channel full during topic broadcast",
				"client_id", client.client.ID,
				"topic", topic)
		}
	}
}

// SendPersonalNotification sends a notification to a specific user
func (h *hub) SendPersonalNotification(userID string, notificationType, title, message, priority string, data map[string]interface{}) error {
	h.mutex.RLock()
	userUUID, err := parseUUID(userID)
	if err != nil {
		h.mutex.RUnlock()
		return fmt.Errorf("invalid user ID: %w", err)
	}

	clients := h.userClients[userUUID]
	h.mutex.RUnlock()

	if len(clients) == 0 {
		return fmt.Errorf("user %s not connected", userID)
	}

	if data == nil {
		data = make(map[string]interface{})
	}
	data["user_id"] = userID
	data["timestamp"] = time.Now()

	payload := models.NotificationPayload{
		Type:     notificationType,
		Title:    title,
		Message:  message,
		Priority: priority,
		Data:     data,
	}

	notificationMessage, _ := models.NewWebSocketMessage(models.WSMsgTypeNotification, payload)

	for _, client := range clients {
		select {
		case client.send <- notificationMessage:
			h.metrics.MessagesSent++
		default:
			h.logger.Warn("Client send channel full during personal notification",
				"client_id", client.client.ID,
				"user_id", userID)
		}
	}

	return nil
}

// GetSubscribers returns all subscribers for a specific topic
func (h *hub) GetSubscribers(topic string) []*models.WebSocketClient {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	subscribers := make([]*models.WebSocketClient, 0)
	if clients, exists := h.topicSubscribers[topic]; exists {
		for _, client := range clients {
			subscribers = append(subscribers, client.client)
		}
	}

	return subscribers
}

// GetClientSubscriptions returns all subscriptions for a specific client
func (h *hub) GetClientSubscriptions(clientID string) []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if client, exists := h.clients[clientID]; exists {
		subscriptions := make([]string, len(client.client.Subscriptions))
		copy(subscriptions, client.client.Subscriptions)
		return subscriptions
	}

	return []string{}
}

// Helper functions

// containsString checks if a slice contains a string
func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// parseUUID parses a UUID string (placeholder - should use proper UUID parsing)
func parseUUID(uuidStr string) (uuid.UUID, error) {
	return uuid.Parse(uuidStr)
}
