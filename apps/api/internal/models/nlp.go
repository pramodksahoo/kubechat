package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// NLPProvider represents different AI/LLM providers
type NLPProvider string

const (
	ProviderOllama NLPProvider = "ollama"
	ProviderOpenAI NLPProvider = "openai"
	ProviderClaude NLPProvider = "claude"
	ProviderLocal  NLPProvider = "local"
)

// SafetyLevel represents the safety classification of a generated kubectl command
// Note: Uses the same constants as audit.go for consistency
type SafetyLevel string

const (
	// Reuse existing constants from audit.go to avoid duplication
	NLPSafetyLevelSafe      SafetyLevel = SafetyLevel(SafetyLevelSafe)
	NLPSafetyLevelWarning   SafetyLevel = SafetyLevel(SafetyLevelWarning)
	NLPSafetyLevelDangerous SafetyLevel = SafetyLevel(SafetyLevelDangerous)
)

// NLPRequest represents a natural language processing request
type NLPRequest struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	UserID      uuid.UUID   `json:"user_id" db:"user_id"`
	SessionID   uuid.UUID   `json:"session_id" db:"session_id"`
	Query       string      `json:"query" db:"query"`
	Context     string      `json:"context,omitempty" db:"context"`
	ClusterInfo string      `json:"cluster_info,omitempty" db:"cluster_info"`
	Provider    NLPProvider `json:"provider" db:"provider"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
}

// NLPResponse represents the response from natural language processing
type NLPResponse struct {
	ID               uuid.UUID   `json:"id" db:"id"`
	RequestID        uuid.UUID   `json:"request_id" db:"request_id"`
	GeneratedCommand string      `json:"generated_command" db:"generated_command"`
	Explanation      string      `json:"explanation" db:"explanation"`
	SafetyLevel      SafetyLevel `json:"safety_level" db:"safety_level"`
	Confidence       float64     `json:"confidence" db:"confidence"`
	Provider         NLPProvider `json:"provider" db:"provider"`
	ProcessingTimeMs int64       `json:"processing_time_ms" db:"processing_time_ms"`
	TokensUsed       int         `json:"tokens_used,omitempty" db:"tokens_used"`
	Error            string      `json:"error,omitempty" db:"error"`
	CreatedAt        time.Time   `json:"created_at" db:"created_at"`
}

// CommandValidationResult represents kubectl command validation results
type CommandValidationResult struct {
	IsValid     bool        `json:"is_valid"`
	SafetyLevel SafetyLevel `json:"safety_level"`
	Warnings    []string    `json:"warnings,omitempty"`
	Errors      []string    `json:"errors,omitempty"`
	Suggestions []string    `json:"suggestions,omitempty"`
}

// PromptTemplate represents templates for different types of kubectl operations
type PromptTemplate struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Category    string    `json:"category" db:"category"`
	Template    string    `json:"template" db:"template"`
	Description string    `json:"description" db:"description"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// NLPMetrics represents metrics for NLP service monitoring
type NLPMetrics struct {
	Provider         NLPProvider `json:"provider"`
	RequestCount     int64       `json:"request_count"`
	SuccessCount     int64       `json:"success_count"`
	ErrorCount       int64       `json:"error_count"`
	AverageLatencyMs float64     `json:"average_latency_ms"`
	TokensUsed       int64       `json:"tokens_used"`
	LastUsed         time.Time   `json:"last_used"`
}

// Validate validates an NLP request
func (req *NLPRequest) Validate() error {
	if req.Query == "" {
		return ErrInvalidInput{Field: "query", Message: "query is required"}
	}

	if len(req.Query) > 2000 {
		return ErrInvalidInput{Field: "query", Message: "query must be less than 2000 characters"}
	}

	if req.UserID == uuid.Nil {
		return ErrInvalidInput{Field: "user_id", Message: "user_id is required"}
	}

	if req.SessionID == uuid.Nil {
		return ErrInvalidInput{Field: "session_id", Message: "session_id is required"}
	}

	// Validate provider if specified
	if req.Provider != "" {
		validProviders := []NLPProvider{ProviderOllama, ProviderOpenAI, ProviderClaude, ProviderLocal}
		isValid := false
		for _, provider := range validProviders {
			if req.Provider == provider {
				isValid = true
				break
			}
		}
		if !isValid {
			return ErrInvalidInput{Field: "provider", Message: "invalid provider specified"}
		}
	}

	return nil
}

// GetSafetyDescription returns a human-readable description of the safety level
func (level SafetyLevel) GetSafetyDescription() string {
	switch level {
	case NLPSafetyLevelSafe:
		return "Safe operation - read-only or non-destructive commands"
	case NLPSafetyLevelWarning:
		return "Warning - potentially impactful operation, review before executing"
	case NLPSafetyLevelDangerous:
		return "Dangerous - destructive operation that could cause service disruption"
	default:
		return "Unknown safety level"
	}
}

// GetSafetyColor returns a color code for UI representation
func (level SafetyLevel) GetSafetyColor() string {
	switch level {
	case NLPSafetyLevelSafe:
		return "#28a745" // Green
	case NLPSafetyLevelWarning:
		return "#ffc107" // Yellow
	case NLPSafetyLevelDangerous:
		return "#dc3545" // Red
	default:
		return "#6c757d" // Gray
	}
}

// IsHighConfidence checks if the NLP response has high confidence
func (resp *NLPResponse) IsHighConfidence() bool {
	return resp.Confidence >= 0.85
}

// GetProviderDisplayName returns a user-friendly provider name
func (provider NLPProvider) GetProviderDisplayName() string {
	switch provider {
	case ProviderOllama:
		return "Ollama (Local)"
	case ProviderOpenAI:
		return "OpenAI GPT"
	case ProviderClaude:
		return "Anthropic Claude"
	case ProviderLocal:
		return "Local Model"
	default:
		return string(provider)
	}
}

// GetDescription returns a description of the NLP request for logging
func (req *NLPRequest) GetDescription() string {
	if len(req.Query) > 100 {
		return fmt.Sprintf("NLP Query (%.100s...)", req.Query)
	}
	return fmt.Sprintf("NLP Query (%s)", req.Query)
}
