package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/nlp"
)

// Service defines the chat service interface
type Service interface {
	// Session management
	CreateSession(ctx context.Context, req *models.ChatSessionRequest, userID *uuid.UUID) (*models.ChatSession, error)
	GetSessions(ctx context.Context, userID *uuid.UUID) ([]*models.ChatSession, error)
	GetSession(ctx context.Context, sessionID uuid.UUID, userID *uuid.UUID) (*models.ChatSession, error)
	UpdateSession(ctx context.Context, sessionID uuid.UUID, updates map[string]interface{}, userID *uuid.UUID) (*models.ChatSession, error)
	DeleteSession(ctx context.Context, sessionID uuid.UUID, userID *uuid.UUID) error

	// Message management
	GetMessages(ctx context.Context, sessionID uuid.UUID, limit, offset int, userID *uuid.UUID) ([]*models.ChatMessage, error)
	SendMessage(ctx context.Context, sessionID uuid.UUID, req *models.ChatMessageRequest, userID *uuid.UUID) (*models.ChatMessage, error)
	ProcessMessage(ctx context.Context, message *models.ChatMessage) (*models.ChatMessage, error)

	// Command management
	GenerateCommandPreview(ctx context.Context, req *models.CommandPreviewRequest, userID *uuid.UUID) (*models.CommandPreview, error)
	GetCommandPreviews(ctx context.Context, sessionID uuid.UUID, userID *uuid.UUID) ([]*models.CommandPreview, error)

	// Context management
	GetChatContext(ctx context.Context, sessionID uuid.UUID, userID *uuid.UUID) (*models.ChatContext, error)
	UpdateChatContext(ctx context.Context, sessionID uuid.UUID, context *models.ChatContext) error

	// Health and metrics
	HealthCheck(ctx context.Context) error
	GetMetrics(ctx context.Context) (map[string]interface{}, error)
}

// service implements the chat service
type service struct {
	nlpService nlp.Service
	config     *Config
}

// Config represents chat service configuration
type Config struct {
	MaxSessionsPerUser    int           `json:"maxSessionsPerUser"`
	MaxMessagesPerSession int           `json:"maxMessagesPerSession"`
	MaxTokensPerMessage   int           `json:"maxTokensPerMessage"`
	SessionTimeout        time.Duration `json:"sessionTimeout"`
	MessageRetention      time.Duration `json:"messageRetention"`
	EnableCaching         bool          `json:"enableCaching"`
	EnableMetrics         bool          `json:"enableMetrics"`
	DefaultModel          string        `json:"defaultModel"`
	SafetyThreshold       float64       `json:"safetyThreshold"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		MaxSessionsPerUser:    50,
		MaxMessagesPerSession: 1000,
		MaxTokensPerMessage:   4000,
		SessionTimeout:        24 * time.Hour,
		MessageRetention:      30 * 24 * time.Hour, // 30 days
		EnableCaching:         true,
		EnableMetrics:         true,
		DefaultModel:          "ollama",
		SafetyThreshold:       0.7,
	}
}

// NewService creates a new chat service
func NewService(nlpSvc nlp.Service, cfg *Config) Service {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return &service{
		nlpService: nlpSvc,
		config:     cfg,
	}
}

// CreateSession creates a new chat session
func (s *service) CreateSession(ctx context.Context, req *models.ChatSessionRequest, userID *uuid.UUID) (*models.ChatSession, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Generate session title if not provided
	title := req.Title
	if title == "" {
		title = s.generateSessionTitle(req.SessionType)
	}

	session := &models.ChatSession{
		ID:          uuid.New(),
		Title:       title,
		UserID:      userID,
		ClusterID:   req.ClusterID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		MessageCount: 0,
		IsActive:    true,
		SessionType: req.SessionType,
	}

	// For now, store in memory - in production this would use a repository
	// TODO: Implement proper database storage

	return session, nil
}

// GetSessions retrieves all sessions for a user
func (s *service) GetSessions(ctx context.Context, userID *uuid.UUID) ([]*models.ChatSession, error) {
	// For now, return empty slice - in production this would query database
	// TODO: Implement proper database query
	return []*models.ChatSession{}, nil
}

// GetSession retrieves a specific session
func (s *service) GetSession(ctx context.Context, sessionID uuid.UUID, userID *uuid.UUID) (*models.ChatSession, error) {
	// For now, return a mock session - in production this would query database
	// TODO: Implement proper database query
	return &models.ChatSession{
		ID:          sessionID,
		Title:       "Sample Chat Session",
		UserID:      userID,
		CreatedAt:   time.Now().Add(-1 * time.Hour),
		UpdatedAt:   time.Now(),
		MessageCount: 0,
		IsActive:    true,
		SessionType: "general",
	}, nil
}

// UpdateSession updates a session
func (s *service) UpdateSession(ctx context.Context, sessionID uuid.UUID, updates map[string]interface{}, userID *uuid.UUID) (*models.ChatSession, error) {
	// For now, return the updated session - in production this would update database
	// TODO: Implement proper database update
	session, err := s.GetSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	session.UpdatedAt = time.Now()
	return session, nil
}

// DeleteSession deletes a session
func (s *service) DeleteSession(ctx context.Context, sessionID uuid.UUID, userID *uuid.UUID) error {
	// For now, do nothing - in production this would delete from database
	// TODO: Implement proper database deletion
	return nil
}

// GetMessages retrieves messages for a session
func (s *service) GetMessages(ctx context.Context, sessionID uuid.UUID, limit, offset int, userID *uuid.UUID) ([]*models.ChatMessage, error) {
	// For now, return empty slice - in production this would query database
	// TODO: Implement proper database query
	return []*models.ChatMessage{}, nil
}

// SendMessage sends a message and processes it
func (s *service) SendMessage(ctx context.Context, sessionID uuid.UUID, req *models.ChatMessageRequest, userID *uuid.UUID) (*models.ChatMessage, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create user message
	userMessage := &models.ChatMessage{
		ID:        uuid.New(),
		SessionID: sessionID,
		UserID:    userID,
		Type:      req.Type,
		Content:   req.Content,
		Timestamp: time.Now(),
		Metadata:  req.Metadata,
		ParentID:  req.ParentID,
		IsEdited:  false,
	}

	// For now, just return the user message - in production this would:
	// 1. Save user message to database
	// 2. Process with NLP service
	// 3. Generate assistant response
	// 4. Save assistant message to database
	// TODO: Implement proper message processing and storage

	return userMessage, nil
}

// ProcessMessage processes a message through NLP
func (s *service) ProcessMessage(ctx context.Context, message *models.ChatMessage) (*models.ChatMessage, error) {
	if message.Type != "user" {
		return message, nil // Only process user messages
	}

	// Use NLP service to generate response
	// Handle nil userID
	userID := uuid.Nil
	if message.UserID != nil {
		userID = *message.UserID
	}

	nlpRequest := &models.NLPRequest{
		Query:     message.Content,
		Context:   fmt.Sprintf("sessionId:%s,messageId:%s,timestamp:%s", message.SessionID.String(), message.ID.String(), message.Timestamp.Format(time.RFC3339)),
		Provider:  models.ProviderOllama, // Use enum value
		SessionID: message.SessionID,
		UserID:    userID,
	}

	response, err := s.nlpService.ProcessQuery(ctx, nlpRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to process message with NLP: %w", err)
	}

	// Create assistant response message
	assistantMessage := &models.ChatMessage{
		ID:        uuid.New(),
		SessionID: message.SessionID,
		UserID:    message.UserID,
		Type:      "assistant",
		Content:   response.Explanation, // Use Explanation field instead of Response
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"model":           response.Provider,
			"processingTime":  response.ProcessingTimeMs, // Use correct field name
			"confidence":      response.Confidence,
			"safetyLevel":     response.SafetyLevel,
			"command":         response.GeneratedCommand,
		},
		ParentID: &message.ID,
		IsEdited: false,
	}

	if response.TokensUsed > 0 {
		tokenCount := int(response.TokensUsed)
		assistantMessage.Token_Count = &tokenCount
	}

	return assistantMessage, nil
}

// GenerateCommandPreview generates a command preview from natural language
func (s *service) GenerateCommandPreview(ctx context.Context, req *models.CommandPreviewRequest, userID *uuid.UUID) (*models.CommandPreview, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Use NLP service to generate command
	contextStr := fmt.Sprintf("task:command_generation,sessionId:%s", req.SessionID.String())
	if req.Context != nil {
		for k, v := range req.Context {
			contextStr += fmt.Sprintf(",%s:%v", k, v)
		}
	}

	// Handle nil userID
	requestUserID := uuid.Nil
	if userID != nil {
		requestUserID = *userID
	}

	nlpRequest := &models.NLPRequest{
		Query:     req.NaturalLanguage,
		Context:   contextStr,
		Provider:  models.ProviderOllama,
		SessionID: req.SessionID,
		UserID:    requestUserID,
	}

	response, err := s.nlpService.ProcessQuery(ctx, nlpRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to generate command preview: %w", err)
	}

	// Parse command from response
	generatedCommand := response.GeneratedCommand
	if generatedCommand == "" {
		// Fallback to explanation if no command generated
		generatedCommand = response.Explanation
	}

	// Analyze safety and risks
	risks := s.analyzeCommandRisks(generatedCommand)
	safeguards := s.generateSafeguards(generatedCommand)
	safetyLevel := s.calculateSafetyLevel(generatedCommand, risks)
	requiresApproval := safetyLevel == "high" || len(risks) > 2

	preview := &models.CommandPreview{
		ID:               uuid.New(),
		SessionID:        req.SessionID,
		UserID:           userID,
		NaturalLanguage:  req.NaturalLanguage,
		GeneratedCommand: generatedCommand,
		Description:      s.generateCommandDescription(generatedCommand),
		Risks:            risks,
		Safeguards:       safeguards,
		EstimatedImpact:  s.getEstimatedImpact(safetyLevel),
		SafetyLevel:      safetyLevel,
		RequiresApproval: requiresApproval,
		ApprovalRequired: requiresApproval,
		Status:           "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Set expiration time (15 minutes for now)
	expiresAt := time.Now().Add(15 * time.Minute)
	preview.ExpiresAt = &expiresAt

	return preview, nil
}

// GetCommandPreviews retrieves command previews for a session
func (s *service) GetCommandPreviews(ctx context.Context, sessionID uuid.UUID, userID *uuid.UUID) ([]*models.CommandPreview, error) {
	// For now, return empty slice - in production this would query database
	// TODO: Implement proper database query
	return []*models.CommandPreview{}, nil
}

// GetChatContext retrieves the chat context for a session
func (s *service) GetChatContext(ctx context.Context, sessionID uuid.UUID, userID *uuid.UUID) (*models.ChatContext, error) {
	// For now, return basic context - in production this would build comprehensive context
	// TODO: Implement proper context building
	return &models.ChatContext{
		SessionID:      sessionID,
		UserID:         userID,
		RecentMessages: []models.ChatMessage{}, // Use slice, not pointer slice
		SafetyConstraints: &models.SafetyConstraints{
			MaxTokensPerMessage:   s.config.MaxTokensPerMessage,
			EnableSafetyChecks:    true,
			AutoRejectUnsafe:      true,
			MaxConcurrentCommands: 3,
		},
	}, nil
}

// UpdateChatContext updates the chat context
func (s *service) UpdateChatContext(ctx context.Context, sessionID uuid.UUID, context *models.ChatContext) error {
	// For now, do nothing - in production this would update context storage
	// TODO: Implement proper context storage
	return nil
}

// HealthCheck performs a health check
func (s *service) HealthCheck(ctx context.Context) error {
	// Check NLP service health
	if s.nlpService != nil {
		if err := s.nlpService.HealthCheck(ctx); err != nil {
			return fmt.Errorf("NLP service unhealthy: %w", err)
		}
	}

	return nil
}

// GetMetrics returns service metrics
func (s *service) GetMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := map[string]interface{}{
		"service":           "chat",
		"status":            "healthy",
		"uptime":            time.Since(time.Now()).String(),
		"totalSessions":     0, // TODO: Implement real metrics
		"totalMessages":     0,
		"totalCommands":     0,
		"averageResponseTime": "0ms",
	}

	return metrics, nil
}

// Helper methods

func (s *service) generateSessionTitle(sessionType string) string {
	switch sessionType {
	case "kubernetes":
		return "Kubernetes Chat - " + time.Now().Format("Jan 02, 15:04")
	case "command":
		return "Command Session - " + time.Now().Format("Jan 02, 15:04")
	default:
		return "Chat Session - " + time.Now().Format("Jan 02, 15:04")
	}
}

func (s *service) analyzeCommandRisks(command string) []string {
	risks := []string{}

	// Simple risk analysis - in production this would be more sophisticated
	dangerousKeywords := []string{"delete", "rm", "destroy", "drop", "truncate"}
	for _, keyword := range dangerousKeywords {
		if contains(command, keyword) {
			risks = append(risks, fmt.Sprintf("Contains potentially destructive operation: %s", keyword))
		}
	}

	if len(risks) == 0 {
		risks = append(risks, "No significant risks identified")
	}

	return risks
}

func (s *service) generateSafeguards(command string) []string {
	safeguards := []string{
		"Command will be executed with limited permissions",
		"Execution will be logged for audit purposes",
		"Command can be cancelled before completion",
	}

	if contains(command, "delete") || contains(command, "rm") {
		safeguards = append(safeguards, "Backup will be created before execution")
	}

	return safeguards
}

func (s *service) calculateSafetyLevel(command string, risks []string) string {
	// Simple safety calculation - in production this would use ML models
	if len(risks) > 3 || contains(command, "delete") {
		return "high"
	}
	if len(risks) > 1 {
		return "medium"
	}
	return "low"
}

func (s *service) getEstimatedImpact(safetyLevel string) string {
	switch safetyLevel {
	case "high":
		return "high"
	case "medium":
		return "medium"
	default:
		return "low"
	}
}

func (s *service) generateCommandDescription(command string) string {
	// Simple description generation - in production this would be more sophisticated
	if command == "" {
		return "No command generated"
	}
	return fmt.Sprintf("Execute command: %s", command)
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr ||
		(len(str) > len(substr) &&
			(str[:len(substr)] == substr ||
			str[len(str)-len(substr):] == substr ||
			findInString(str, substr))))
}

func findInString(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}