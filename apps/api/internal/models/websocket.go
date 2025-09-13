package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// WebSocketMessageType represents different types of WebSocket messages
type WebSocketMessageType string

const (
	// Client -> Server message types
	WSMsgTypeAuth        WebSocketMessageType = "auth"
	WSMsgTypeExecute     WebSocketMessageType = "execute"
	WSMsgTypeCancel      WebSocketMessageType = "cancel"
	WSMsgTypeSubscribe   WebSocketMessageType = "subscribe"
	WSMsgTypeUnsubscribe WebSocketMessageType = "unsubscribe"
	WSMsgTypePing        WebSocketMessageType = "ping"
	WSMsgTypeHeartbeat   WebSocketMessageType = "heartbeat"

	// Server -> Client message types
	WSMsgTypeAuthResult   WebSocketMessageType = "auth_result"
	WSMsgTypeExecuting    WebSocketMessageType = "executing"
	WSMsgTypeProgress     WebSocketMessageType = "progress"
	WSMsgTypeOutput       WebSocketMessageType = "output"
	WSMsgTypeResult       WebSocketMessageType = "result"
	WSMsgTypeError        WebSocketMessageType = "error"
	WSMsgTypeStatus       WebSocketMessageType = "status"
	WSMsgTypePong         WebSocketMessageType = "pong"
	WSMsgTypeNotification WebSocketMessageType = "notification"
)

// WebSocketMessage represents a generic WebSocket message
type WebSocketMessage struct {
	ID        string               `json:"id"`
	Type      WebSocketMessageType `json:"type"`
	Payload   json.RawMessage      `json:"payload"`
	Timestamp time.Time            `json:"timestamp"`
	ClientID  string               `json:"client_id,omitempty"`
}

// AuthPayload represents authentication payload
type AuthPayload struct {
	Token     string `json:"token"`
	UserAgent string `json:"user_agent,omitempty"`
}

// AuthResultPayload represents authentication result
type AuthResultPayload struct {
	Success   bool      `json:"success"`
	UserID    uuid.UUID `json:"user_id,omitempty"`
	Username  string    `json:"username,omitempty"`
	Role      string    `json:"role,omitempty"`
	SessionID uuid.UUID `json:"session_id,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// ExecutePayload represents command execution request
type ExecutePayload struct {
	Command      string            `json:"command"`
	Query        string            `json:"query,omitempty"`
	Context      string            `json:"context,omitempty"`
	ClusterInfo  string            `json:"cluster_info,omitempty"`
	Namespace    string            `json:"namespace,omitempty"`
	Provider     string            `json:"provider,omitempty"`
	Options      map[string]string `json:"options,omitempty"`
	StreamOutput bool              `json:"stream_output"`
}

// ExecutingPayload represents command execution start notification
type ExecutingPayload struct {
	CommandID     string    `json:"command_id"`
	Command       string    `json:"command"`
	SafetyLevel   string    `json:"safety_level"`
	EstimatedTime string    `json:"estimated_time,omitempty"`
	StartedAt     time.Time `json:"started_at"`
}

// ProgressPayload represents execution progress
type ProgressPayload struct {
	CommandID   string    `json:"command_id"`
	Percentage  float64   `json:"percentage"`
	Stage       string    `json:"stage"`
	Description string    `json:"description,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// OutputPayload represents command output (streaming)
type OutputPayload struct {
	CommandID string    `json:"command_id"`
	Content   string    `json:"content"`
	Stream    string    `json:"stream"` // "stdout", "stderr"
	Timestamp time.Time `json:"timestamp"`
}

// ResultPayload represents final command execution result
type ResultPayload struct {
	CommandID       string                 `json:"command_id"`
	Success         bool                   `json:"success"`
	ExitCode        int                    `json:"exit_code"`
	Output          string                 `json:"output"`
	Error           string                 `json:"error,omitempty"`
	Duration        string                 `json:"duration"`
	ExecutionResult map[string]interface{} `json:"execution_result,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}

// ErrorPayload represents error message
type ErrorPayload struct {
	CommandID string `json:"command_id,omitempty"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
}

// StatusPayload represents system status update
type StatusPayload struct {
	Type      string                 `json:"type"` // "system", "user", "command"
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// SubscribePayload represents subscription request
type SubscribePayload struct {
	Topics []string `json:"topics"` // "commands", "system", "user_activity"
}

// NotificationPayload represents push notifications
type NotificationPayload struct {
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Message  string                 `json:"message"`
	Priority string                 `json:"priority"` // "low", "normal", "high", "urgent"
	Data     map[string]interface{} `json:"data,omitempty"`
}

// CancelPayload represents command cancellation request
type CancelPayload struct {
	CommandID string `json:"command_id"`
	Reason    string `json:"reason,omitempty"`
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	ID          string    `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Role        string    `json:"role"`
	SessionID   uuid.UUID `json:"session_id"`
	ConnectedAt time.Time `json:"connected_at"`
	LastPing    time.Time `json:"last_ping"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`

	// Subscription topics
	Subscriptions []string `json:"subscriptions"`

	// Active commands
	ActiveCommands []string `json:"active_commands"`
}

// CommandExecution represents an active command execution
type CommandExecution struct {
	ID              string                 `json:"id"`
	ClientID        string                 `json:"client_id"`
	UserID          uuid.UUID              `json:"user_id"`
	SessionID       uuid.UUID              `json:"session_id"`
	Command         string                 `json:"command"`
	Query           string                 `json:"query,omitempty"`
	SafetyLevel     string                 `json:"safety_level"`
	Status          string                 `json:"status"` // "queued", "running", "completed", "failed", "cancelled"
	StartedAt       time.Time              `json:"started_at"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	Duration        string                 `json:"duration,omitempty"`
	ExitCode        *int                   `json:"exit_code,omitempty"`
	Output          string                 `json:"output,omitempty"`
	Error           string                 `json:"error,omitempty"`
	ExecutionResult map[string]interface{} `json:"execution_result,omitempty"`
	StreamOutput    bool                   `json:"stream_output"`
}

// WebSocketMetrics represents WebSocket service metrics
type WebSocketMetrics struct {
	ConnectedClients    int     `json:"connected_clients"`
	TotalConnections    int64   `json:"total_connections"`
	ActiveCommands      int     `json:"active_commands"`
	CompletedCommands   int64   `json:"completed_commands"`
	FailedCommands      int64   `json:"failed_commands"`
	AverageResponseTime float64 `json:"average_response_time_ms"`
	MessagesSent        int64   `json:"messages_sent"`
	MessagesReceived    int64   `json:"messages_received"`
	ErrorCount          int64   `json:"error_count"`
	Uptime              string  `json:"uptime"`
}

// NewWebSocketMessage creates a new WebSocket message
func NewWebSocketMessage(msgType WebSocketMessageType, payload interface{}) (*WebSocketMessage, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &WebSocketMessage{
		ID:        uuid.New().String(),
		Type:      msgType,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}, nil
}

// UnmarshalPayload unmarshals the payload into the specified type
func (msg *WebSocketMessage) UnmarshalPayload(v interface{}) error {
	return json.Unmarshal(msg.Payload, v)
}

// IsValidMessageType checks if the message type is valid
func IsValidMessageType(msgType WebSocketMessageType) bool {
	validTypes := []WebSocketMessageType{
		WSMsgTypeAuth, WSMsgTypeExecute, WSMsgTypeCancel, WSMsgTypeSubscribe,
		WSMsgTypeUnsubscribe, WSMsgTypePing, WSMsgTypeHeartbeat,
		WSMsgTypeAuthResult, WSMsgTypeExecuting, WSMsgTypeProgress,
		WSMsgTypeOutput, WSMsgTypeResult, WSMsgTypeError, WSMsgTypeStatus,
		WSMsgTypePong, WSMsgTypeNotification,
	}

	for _, validType := range validTypes {
		if msgType == validType {
			return true
		}
	}
	return false
}

// IsClientMessage checks if the message type is sent by clients
func (msgType WebSocketMessageType) IsClientMessage() bool {
	clientTypes := []WebSocketMessageType{
		WSMsgTypeAuth, WSMsgTypeExecute, WSMsgTypeCancel,
		WSMsgTypeSubscribe, WSMsgTypeUnsubscribe, WSMsgTypePing, WSMsgTypeHeartbeat,
	}

	for _, clientType := range clientTypes {
		if msgType == clientType {
			return true
		}
	}
	return false
}

// IsServerMessage checks if the message type is sent by server
func (msgType WebSocketMessageType) IsServerMessage() bool {
	serverTypes := []WebSocketMessageType{
		WSMsgTypeAuthResult, WSMsgTypeExecuting, WSMsgTypeProgress,
		WSMsgTypeOutput, WSMsgTypeResult, WSMsgTypeError, WSMsgTypeStatus,
		WSMsgTypePong, WSMsgTypeNotification,
	}

	for _, serverType := range serverTypes {
		if msgType == serverType {
			return true
		}
	}
	return false
}
