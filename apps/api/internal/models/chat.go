package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// ChatSession represents a chat conversation session
type ChatSession struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Title         string     `json:"title" db:"title"`
	UserID        *uuid.UUID `json:"userId,omitempty" db:"user_id"`
	ClusterID     *string    `json:"clusterId,omitempty" db:"cluster_id"`
	ClusterName   *string    `json:"clusterName,omitempty" db:"cluster_name"`
	CreatedAt     time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time  `json:"updatedAt" db:"updated_at"`
	MessageCount  int        `json:"messageCount" db:"message_count"`
	LastMessage   *string    `json:"lastMessage,omitempty" db:"last_message"`
	IsActive      bool       `json:"isActive" db:"is_active"`
	SessionType   string     `json:"sessionType" db:"session_type"` // "general", "kubernetes", "command"
}

// ChatMessage represents a message in a chat session
type ChatMessage struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	SessionID   uuid.UUID              `json:"sessionId" db:"session_id"`
	UserID      *uuid.UUID             `json:"userId,omitempty" db:"user_id"`
	Type        string                 `json:"type" db:"type"` // "user", "assistant", "system"
	Content     string                 `json:"content" db:"content"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	ParentID    *uuid.UUID             `json:"parentId,omitempty" db:"parent_id"`
	IsEdited    bool                   `json:"isEdited" db:"is_edited"`
	EditedAt    *time.Time             `json:"editedAt,omitempty" db:"edited_at"`
	Token_Count *int                   `json:"tokenCount,omitempty" db:"token_count"`
}

// CommandPreview represents a command preview before execution
type CommandPreview struct {
	ID               uuid.UUID `json:"id" db:"id"`
	SessionID        uuid.UUID `json:"sessionId" db:"session_id"`
	UserID           *uuid.UUID `json:"userId,omitempty" db:"user_id"`
	NaturalLanguage  string     `json:"naturalLanguage" db:"natural_language"`
	GeneratedCommand string     `json:"generatedCommand" db:"generated_command"`
	Description      string     `json:"description" db:"description"`
	Risks            []string   `json:"risks" db:"risks"`
	Safeguards       []string   `json:"safeguards" db:"safeguards"`
	EstimatedImpact  string     `json:"estimatedImpact" db:"estimated_impact"` // "low", "medium", "high"
	SafetyLevel      string     `json:"safetyLevel" db:"safety_level"`         // "low", "medium", "high"
	RequiresApproval bool       `json:"requiresApproval" db:"requires_approval"`
	ApprovalRequired bool       `json:"approvalRequired" db:"approval_required"`
	Status           string     `json:"status" db:"status"` // "pending", "approved", "rejected", "executed"
	CreatedAt        time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time  `json:"updatedAt" db:"updated_at"`
	ExpiresAt        *time.Time `json:"expiresAt,omitempty" db:"expires_at"`
}

// ChatContext represents the context for chat interactions
type ChatContext struct {
	SessionID         uuid.UUID              `json:"sessionId"`
	UserID            *uuid.UUID             `json:"userId,omitempty"`
	ClusterContext    *KubernetesContext     `json:"clusterContext,omitempty"`
	CommandHistory    []CommandExecution     `json:"commandHistory,omitempty"`
	RecentMessages    []ChatMessage          `json:"recentMessages,omitempty"`
	UserPreferences   map[string]interface{} `json:"userPreferences,omitempty"`
	SafetyConstraints *SafetyConstraints     `json:"safetyConstraints,omitempty"`
}

// KubernetesContext provides cluster-specific context for chat
type KubernetesContext struct {
	ClusterID       string   `json:"clusterId"`
	ClusterName     string   `json:"clusterName"`
	Namespace       string   `json:"namespace"`
	AllowedActions  []string `json:"allowedActions"`
	RestrictedModes []string `json:"restrictedModes"`
	Resources       struct {
		Pods        int `json:"pods"`
		Services    int `json:"services"`
		Deployments int `json:"deployments"`
	} `json:"resources"`
}

// SafetyConstraints defines safety rules for chat interactions
type SafetyConstraints struct {
	MaxTokensPerMessage   int      `json:"maxTokensPerMessage"`
	BlockedCommands       []string `json:"blockedCommands"`
	RequireApprovalFor    []string `json:"requireApprovalFor"`
	AllowedNamespaces     []string `json:"allowedNamespaces"`
	DangerousOperations   []string `json:"dangerousOperations"`
	EnableSafetyChecks    bool     `json:"enableSafetyChecks"`
	AutoRejectUnsafe      bool     `json:"autoRejectUnsafe"`
	MaxConcurrentCommands int      `json:"maxConcurrentCommands"`
}

// ChatMessageRequest represents a request to send a chat message
type ChatMessageRequest struct {
	Content     string                 `json:"content" binding:"required"`
	Type        string                 `json:"type,omitempty"`     // defaults to "user"
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ParentID    *uuid.UUID             `json:"parentId,omitempty"`
	ClusterID   *string                `json:"clusterId,omitempty"`
}

// ChatSessionRequest represents a request to create a chat session
type ChatSessionRequest struct {
	Title       string  `json:"title,omitempty"`
	ClusterID   *string `json:"clusterId,omitempty"`
	SessionType string  `json:"sessionType,omitempty"` // defaults to "general"
}

// CommandPreviewRequest represents a request for command preview
type CommandPreviewRequest struct {
	NaturalLanguage string  `json:"naturalLanguage" binding:"required"`
	SessionID       uuid.UUID `json:"sessionId" binding:"required"`
	ClusterID       *string `json:"clusterId,omitempty"`
	Namespace       *string `json:"namespace,omitempty"`
	Context         map[string]interface{} `json:"context,omitempty"`
}

// ChatResponse represents a generic chat API response
type ChatResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *string     `json:"error,omitempty"`
	Message *string     `json:"message,omitempty"`
}

// ChatSessionResponse represents a chat session API response
type ChatSessionResponse struct {
	Session  *ChatSession `json:"session"`
	Messages []ChatMessage `json:"messages,omitempty"`
	Context  *ChatContext `json:"context,omitempty"`
}

// ChatWebSocketMessage represents a WebSocket message structure
type ChatWebSocketMessage struct {
	Type      string                 `json:"type"` // "message", "command_status", "typing", "error"
	SessionID uuid.UUID              `json:"sessionId"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Validation methods

// Validate validates a chat session request
func (r *ChatSessionRequest) Validate() error {
	if r.SessionType == "" {
		r.SessionType = "general"
	}

	validTypes := []string{"general", "kubernetes", "command"}
	isValid := false
	for _, validType := range validTypes {
		if r.SessionType == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		return ErrInvalidInput{Field: "sessionType", Message: "must be one of: general, kubernetes, command"}
	}

	return nil
}

// Validate validates a chat message request
func (r *ChatMessageRequest) Validate() error {
	if len(r.Content) == 0 {
		return ErrInvalidInput{Field: "content", Message: "message content cannot be empty"}
	}

	if len(r.Content) > 4000 {
		return ErrInvalidInput{Field: "content", Message: "message content exceeds maximum length of 4000 characters"}
	}

	if r.Type == "" {
		r.Type = "user"
	}

	validTypes := []string{"user", "assistant", "system"}
	isValid := false
	for _, validType := range validTypes {
		if r.Type == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		return ErrInvalidInput{Field: "type", Message: "must be one of: user, assistant, system"}
	}

	return nil
}

// Validate validates a command preview request
func (r *CommandPreviewRequest) Validate() error {
	if len(r.NaturalLanguage) == 0 {
		return ErrInvalidInput{Field: "naturalLanguage", Message: "natural language input cannot be empty"}
	}

	if len(r.NaturalLanguage) > 1000 {
		return ErrInvalidInput{Field: "naturalLanguage", Message: "natural language input exceeds maximum length of 1000 characters"}
	}

	if r.SessionID == uuid.Nil {
		return ErrInvalidInput{Field: "sessionId", Message: "session ID is required"}
	}

	return nil
}

// Helper methods

// IsCommandMessage checks if a message looks like a command request
func (m *ChatMessage) IsCommandMessage() bool {
	if m.Type != "user" {
		return false
	}

	commandKeywords := []string{
		"kubectl", "create", "delete", "deploy", "scale", "restart",
		"get pods", "describe", "logs", "exec", "apply", "rollout",
		"list", "show me", "check", "restart", "update", "status",
	}

	content := strings.ToLower(m.Content)
	for _, keyword := range commandKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}

	return false
}

// GetTokenCount estimates token count for a message
func (m *ChatMessage) GetTokenCount() int {
	if m.Token_Count != nil {
		return *m.Token_Count
	}

	// Simple estimation: ~4 characters per token
	return len(m.Content) / 4
}

// IsExpired checks if a command preview has expired
func (cp *CommandPreview) IsExpired() bool {
	if cp.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*cp.ExpiresAt)
}

// CanExecute checks if a command preview can be executed
func (cp *CommandPreview) CanExecute() bool {
	if cp.IsExpired() {
		return false
	}

	if cp.Status != "pending" && cp.Status != "approved" {
		return false
	}

	if cp.RequiresApproval && cp.Status != "approved" {
		return false
	}

	return true
}