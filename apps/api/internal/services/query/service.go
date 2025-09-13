package query

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/nlp"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/safety"
)

// ProcessingResult represents the result of query processing
type ProcessingResult struct {
	ID               uuid.UUID `json:"id"`
	Command          string    `json:"command"`
	Explanation      string    `json:"explanation"`
	SafetyLevel      string    `json:"safety_level"`
	Confidence       float64   `json:"confidence"`
	ProcessingTime   int64     `json:"processing_time_ms"`
	Suggestions      []string  `json:"suggestions,omitempty"`
	Warnings         []string  `json:"warnings,omitempty"`
	SafetyReasons    []string  `json:"safety_reasons,omitempty"`
	IsBlocked        bool      `json:"is_blocked"`
	RequiresApproval bool      `json:"requires_approval"`
}

// QueryRequest represents an enhanced query processing request
type QueryRequest struct {
	ID          uuid.UUID         `json:"id"`
	UserID      uuid.UUID         `json:"user_id"`
	SessionID   uuid.UUID         `json:"session_id"`
	Query       string            `json:"query"`
	Namespace   string            `json:"namespace,omitempty"`
	Context     map[string]string `json:"context,omitempty"`
	UserRole    string            `json:"user_role"`
	Environment string            `json:"environment"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
}

// Service defines the query processing service interface
type Service interface {
	// ProcessQuery processes a natural language query with full safety integration
	ProcessQuery(ctx context.Context, req QueryRequest) (*ProcessingResult, error)

	// ValidateQueryFormat validates query format and basic requirements
	ValidateQueryFormat(ctx context.Context, query string) error

	// FormatResponse formats the processing result for presentation
	FormatResponse(ctx context.Context, result *ProcessingResult) (map[string]interface{}, error)

	// HandleMalformedQuery handles queries that couldn't be parsed properly
	HandleMalformedQuery(ctx context.Context, query string) (*ProcessingResult, error)

	// ManageTimeout handles timeout management for AI processing
	ManageTimeout(ctx context.Context, req QueryRequest) (context.Context, context.CancelFunc)

	// HealthCheck validates query service health
	HealthCheck(ctx context.Context) error
}

// Config represents query processing configuration
type Config struct {
	DefaultTimeout         time.Duration `json:"default_timeout"`
	MaxQueryLength         int           `json:"max_query_length"`
	MinQueryLength         int           `json:"min_query_length"`
	EnableFormatting       bool          `json:"enable_formatting"`
	EnableSuggestions      bool          `json:"enable_suggestions"`
	DefaultNamespace       string        `json:"default_namespace"`
	EnableContextExpansion bool          `json:"enable_context_expansion"`
}

// serviceImpl implements the Query Processing Service interface
type serviceImpl struct {
	config        *Config
	nlpService    nlp.Service
	safetyService safety.Service
}

// NewService creates a new query processing service
func NewService(nlpService nlp.Service, safetyService safety.Service, config *Config) Service {
	if config == nil {
		config = &Config{
			DefaultTimeout:         30 * time.Second,
			MaxQueryLength:         2000,
			MinQueryLength:         3,
			EnableFormatting:       true,
			EnableSuggestions:      true,
			DefaultNamespace:       "default",
			EnableContextExpansion: true,
		}
	}

	service := &serviceImpl{
		config:        config,
		nlpService:    nlpService,
		safetyService: safetyService,
	}

	log.Printf("Query processing service initialized with timeout: %v", config.DefaultTimeout)
	return service
}

// ProcessQuery processes a natural language query with full safety integration
func (s *serviceImpl) ProcessQuery(ctx context.Context, req QueryRequest) (*ProcessingResult, error) {
	startTime := time.Now()

	// Validate query format
	if err := s.ValidateQueryFormat(ctx, req.Query); err != nil {
		return nil, fmt.Errorf("query validation failed: %w", err)
	}

	// Set up timeout management
	processCtx, cancel := s.ManageTimeout(ctx, req)
	defer cancel()

	// Expand context if enabled
	expandedContext := s.expandContext(req)

	// Process with NLP service
	nlpRequest := nlp.NLPRequest{
		Query:       req.Query,
		Context:     expandedContext,
		Namespace:   req.Namespace,
		UserRole:    req.UserRole,
		Environment: req.Environment,
	}

	nlpResult, err := s.nlpService.GenerateCommand(processCtx, nlpRequest)
	if err != nil {
		return s.handleProcessingError(ctx, req, err)
	}

	// Perform safety analysis
	safetyRequest := safety.ContextualSafetyRequest{
		Command:     nlpResult.Command,
		UserRole:    req.UserRole,
		Environment: req.Environment,
		Namespace:   req.Namespace,
		Context:     expandedContext,
	}

	safetyResult, err := s.safetyService.ClassifyWithContext(processCtx, safetyRequest)
	if err != nil {
		log.Printf("Safety analysis failed: %v", err)
		// Continue with basic classification
		basicSafety, _ := s.safetyService.ClassifyCommand(processCtx, nlpResult.Command)
		safetyResult = basicSafety
	}

	// Combine results
	processingTime := time.Since(startTime).Milliseconds()
	result := &ProcessingResult{
		ID:               req.ID,
		Command:          nlpResult.Command,
		Explanation:      nlpResult.Explanation,
		SafetyLevel:      string(safetyResult.Level),
		Confidence:       nlpResult.Confidence,
		ProcessingTime:   processingTime,
		Suggestions:      s.mergeSuggestions(nlpResult.Suggestions, safetyResult.Suggestions),
		Warnings:         s.mergeWarnings(nlpResult.Warnings, safetyResult.Warnings),
		SafetyReasons:    safetyResult.Reasons,
		IsBlocked:        safetyResult.Blocked,
		RequiresApproval: safetyResult.RequiresApproval,
	}

	return result, nil
}

// ValidateQueryFormat validates query format and basic requirements
func (s *serviceImpl) ValidateQueryFormat(ctx context.Context, query string) error {
	if query == "" {
		return fmt.Errorf("query cannot be empty")
	}

	if len(query) < s.config.MinQueryLength {
		return fmt.Errorf("query too short: minimum %d characters required", s.config.MinQueryLength)
	}

	if len(query) > s.config.MaxQueryLength {
		return fmt.Errorf("query too long: maximum %d characters allowed", s.config.MaxQueryLength)
	}

	// Check for potential malicious content
	if valid, issues, err := s.safetyService.ValidatePrompt(ctx, query); !valid {
		if err != nil {
			return fmt.Errorf("prompt validation error: %w", err)
		}
		return fmt.Errorf("query validation failed: %s", strings.Join(issues, ", "))
	}

	return nil
}

// FormatResponse formats the processing result for presentation
func (s *serviceImpl) FormatResponse(ctx context.Context, result *ProcessingResult) (map[string]interface{}, error) {
	if !s.config.EnableFormatting {
		return map[string]interface{}{
			"command":     result.Command,
			"explanation": result.Explanation,
		}, nil
	}

	response := map[string]interface{}{
		"id":                result.ID,
		"command":           result.Command,
		"explanation":       result.Explanation,
		"safety_level":      result.SafetyLevel,
		"confidence":        result.Confidence,
		"processing_time":   result.ProcessingTime,
		"is_blocked":        result.IsBlocked,
		"requires_approval": result.RequiresApproval,
	}

	// Add optional fields if they have content
	if len(result.Suggestions) > 0 && s.config.EnableSuggestions {
		response["suggestions"] = result.Suggestions
	}

	if len(result.Warnings) > 0 {
		response["warnings"] = result.Warnings
	}

	if len(result.SafetyReasons) > 0 {
		response["safety_reasons"] = result.SafetyReasons
	}

	// Add formatting hints for CLI presentation
	response["formatting"] = map[string]interface{}{
		"command_syntax_highlight": true,
		"safety_color":             s.getSafetyColor(result.SafetyLevel),
		"confidence_indicator":     s.getConfidenceIndicator(result.Confidence),
	}

	return response, nil
}

// HandleMalformedQuery handles queries that couldn't be parsed properly
func (s *serviceImpl) HandleMalformedQuery(ctx context.Context, query string) (*ProcessingResult, error) {
	result := &ProcessingResult{
		ID:               uuid.New(),
		Command:          "",
		Explanation:      "Query could not be processed due to formatting issues",
		SafetyLevel:      "dangerous",
		Confidence:       0.0,
		ProcessingTime:   0,
		Suggestions:      s.generateMalformedSuggestions(query),
		Warnings:         []string{"Query could not be parsed correctly"},
		SafetyReasons:    []string{"Malformed query detected"},
		IsBlocked:        true,
		RequiresApproval: false,
	}

	return result, nil
}

// ManageTimeout handles timeout management for AI processing
func (s *serviceImpl) ManageTimeout(ctx context.Context, req QueryRequest) (context.Context, context.CancelFunc) {
	timeout := s.config.DefaultTimeout
	if req.Timeout > 0 {
		timeout = req.Timeout
	}

	return context.WithTimeout(ctx, timeout)
}

// HealthCheck validates query service health
func (s *serviceImpl) HealthCheck(ctx context.Context) error {
	// Test basic query processing with a safe command
	testReq := QueryRequest{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		SessionID:   uuid.New(),
		Query:       "show me the pods",
		Namespace:   "default",
		Context:     map[string]string{"test": "health_check"},
		UserRole:    "admin",
		Environment: "development",
	}

	result, err := s.ProcessQuery(ctx, testReq)
	if err != nil {
		return fmt.Errorf("query processing health check failed: %w", err)
	}

	if result.Command == "" {
		return fmt.Errorf("health check produced empty command")
	}

	// Test format validation
	if err := s.ValidateQueryFormat(ctx, "test query"); err != nil {
		return fmt.Errorf("query format validation health check failed: %w", err)
	}

	log.Println("Query processing service health check passed")
	return nil
}

// Helper methods

func (s *serviceImpl) expandContext(req QueryRequest) map[string]string {
	if !s.config.EnableContextExpansion {
		return req.Context
	}

	expanded := make(map[string]string)

	// Copy existing context
	for k, v := range req.Context {
		expanded[k] = v
	}

	// Add default namespace if not specified
	if req.Namespace == "" && s.config.DefaultNamespace != "" {
		expanded["default_namespace"] = s.config.DefaultNamespace
	} else if req.Namespace != "" {
		expanded["target_namespace"] = req.Namespace
	}

	// Add user context
	if req.UserRole != "" {
		expanded["user_role"] = req.UserRole
	}

	if req.Environment != "" {
		expanded["environment"] = req.Environment
	}

	// Add processing hints
	expanded["processing_timestamp"] = time.Now().Format(time.RFC3339)
	expanded["enable_safety"] = "true"

	return expanded
}

func (s *serviceImpl) handleProcessingError(ctx context.Context, req QueryRequest, err error) (*ProcessingResult, error) {
	result := &ProcessingResult{
		ID:               req.ID,
		Command:          "",
		Explanation:      fmt.Sprintf("Processing failed: %s", err.Error()),
		SafetyLevel:      "dangerous",
		Confidence:       0.0,
		ProcessingTime:   0,
		Suggestions:      []string{"Try rephrasing your query", "Check for typos or unclear wording"},
		Warnings:         []string{"Query processing encountered an error"},
		SafetyReasons:    []string{"Processing failure detected"},
		IsBlocked:        true,
		RequiresApproval: false,
	}

	return result, fmt.Errorf("query processing failed: %w", err)
}

func (s *serviceImpl) mergeSuggestions(nlpSuggestions, safetySuggestions []string) []string {
	merged := make([]string, 0, len(nlpSuggestions)+len(safetySuggestions))
	merged = append(merged, nlpSuggestions...)
	merged = append(merged, safetySuggestions...)
	return s.deduplicate(merged)
}

func (s *serviceImpl) mergeWarnings(nlpWarnings, safetyWarnings []string) []string {
	merged := make([]string, 0, len(nlpWarnings)+len(safetyWarnings))
	merged = append(merged, nlpWarnings...)
	merged = append(merged, safetyWarnings...)
	return s.deduplicate(merged)
}

func (s *serviceImpl) deduplicate(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func (s *serviceImpl) generateMalformedSuggestions(query string) []string {
	suggestions := []string{
		"Try using simpler language",
		"Specify the Kubernetes resource you want to work with",
		"Example: 'show me pods in default namespace'",
	}

	// Add specific suggestions based on query content
	lowerQuery := strings.ToLower(query)

	if strings.Contains(lowerQuery, "delete") || strings.Contains(lowerQuery, "remove") {
		suggestions = append(suggestions, "For deletions, be specific about what you want to delete")
	}

	if strings.Contains(lowerQuery, "create") || strings.Contains(lowerQuery, "make") {
		suggestions = append(suggestions, "For creating resources, specify the resource type and configuration")
	}

	return suggestions
}

func (s *serviceImpl) getSafetyColor(level string) string {
	switch strings.ToLower(level) {
	case "safe":
		return "#28a745" // Green
	case "warning":
		return "#ffc107" // Yellow
	case "dangerous":
		return "#dc3545" // Red
	default:
		return "#6c757d" // Gray
	}
}

func (s *serviceImpl) getConfidenceIndicator(confidence float64) string {
	if confidence >= 0.9 {
		return "high"
	} else if confidence >= 0.7 {
		return "medium"
	} else {
		return "low"
	}
}

// Helper function for safety level string conversion (using safety package type)
func safetyLevelToString(level safety.SafetyLevel) string {
	return string(level)
}
