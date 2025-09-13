package nlp

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/audit"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/safety"
)

// AuditIntegration handles audit logging for NLP operations
type AuditIntegration struct {
	auditService audit.Service
	enabled      bool
}

// QueryAuditEntry represents audit data for NLP queries
type QueryAuditEntry struct {
	ID               uuid.UUID         `json:"id"`
	UserID           uuid.UUID         `json:"user_id"`
	SessionID        uuid.UUID         `json:"session_id"`
	NaturalLanguage  string            `json:"natural_language"`
	GeneratedCommand string            `json:"generated_command"`
	SafetyLevel      string            `json:"safety_level"`
	SafetyReasons    []string          `json:"safety_reasons"`
	Context          map[string]string `json:"context"`
	ProcessingTime   time.Duration     `json:"processing_time"`
	Success          bool              `json:"success"`
	Error            string            `json:"error,omitempty"`
	Timestamp        time.Time         `json:"timestamp"`
	Provider         string            `json:"provider"`
	Confidence       float64           `json:"confidence"`
	IsBlocked        bool              `json:"is_blocked"`
	RequiredApproval bool              `json:"required_approval"`
}

// NewAuditIntegration creates a new audit integration instance
func NewAuditIntegration(auditService audit.Service) *AuditIntegration {
	return &AuditIntegration{
		auditService: auditService,
		enabled:      auditService != nil,
	}
}

// LogQuery logs a natural language query with safety analysis
func (ai *AuditIntegration) LogQuery(ctx context.Context, entry QueryAuditEntry) error {
	if !ai.enabled {
		log.Printf("Audit logging disabled - query: %s", entry.NaturalLanguage)
		return nil
	}

	// Sanitize sensitive data before logging
	sanitizedEntry := ai.sanitizeAuditEntry(entry)

	// Create structured description with key information
	description := fmt.Sprintf("NLP Query: '%s' -> Command: '%s' | Safety: %s | Provider: %s | Confidence: %.2f | Processing: %dms",
		sanitizedEntry.NaturalLanguage,
		sanitizedEntry.GeneratedCommand,
		sanitizedEntry.SafetyLevel,
		sanitizedEntry.Provider,
		sanitizedEntry.Confidence,
		sanitizedEntry.ProcessingTime.Milliseconds())

	// Determine severity based on safety level and outcome
	severity := "info"
	if sanitizedEntry.IsBlocked || sanitizedEntry.SafetyLevel == "dangerous" {
		severity = "high"
	} else if sanitizedEntry.RequiredApproval || sanitizedEntry.SafetyLevel == "warning" {
		severity = "medium"
	} else if !sanitizedEntry.Success {
		severity = "medium"
	}

	// Log to audit service with structured security event
	if err := ai.auditService.LogSecurityEvent(ctx, "NLP_QUERY_PROCESSED", description, &sanitizedEntry.UserID, severity, nil); err != nil {
		log.Printf("Failed to log NLP query audit: %v", err)
		return fmt.Errorf("audit logging failed: %w", err)
	}

	// Additional structured logging for development and debugging
	log.Printf("NLP Query Audit: user=%s session=%s safety=%s success=%v blocked=%v approval=%v",
		sanitizedEntry.UserID, sanitizedEntry.SessionID, sanitizedEntry.SafetyLevel,
		sanitizedEntry.Success, sanitizedEntry.IsBlocked, sanitizedEntry.RequiredApproval)

	log.Printf("NLP query audited: user=%s, safety=%s, success=%v",
		sanitizedEntry.UserID, sanitizedEntry.SafetyLevel, sanitizedEntry.Success)

	return nil
}

// LogSafetyDecision logs safety classification decisions
func (ai *AuditIntegration) LogSafetyDecision(ctx context.Context, userID, sessionID uuid.UUID,
	command string, classification *safety.SafetyClassification) error {

	if !ai.enabled {
		return nil
	}

	// Create structured description for safety decision
	description := fmt.Sprintf("Safety Classification: Command='%s' | Level=%s | Score=%.1f | Blocked=%v | Approval=%v | Reasons=%v",
		command, string(classification.Level), classification.Score,
		classification.Blocked, classification.RequiresApproval, classification.Reasons)

	// Determine severity based on classification
	severity := "info"
	if classification.Blocked || classification.Level == safety.SafetyLevelDangerous {
		severity = "high"
	} else if classification.RequiresApproval || classification.Level == safety.SafetyLevelWarning {
		severity = "medium"
	}

	if err := ai.auditService.LogSecurityEvent(ctx, "SAFETY_CLASSIFICATION", description, &userID, severity, nil); err != nil {
		log.Printf("Failed to log safety decision audit: %v", err)
		return fmt.Errorf("safety audit logging failed: %w", err)
	}

	return nil
}

// LogSafetyOverride logs when safety restrictions are overridden
func (ai *AuditIntegration) LogSafetyOverride(ctx context.Context, userID, sessionID uuid.UUID,
	command, reason, approver string) error {

	if !ai.enabled {
		return nil
	}

	// Create structured description for safety override
	description := fmt.Sprintf("Safety Override: Command='%s' | User=%s | Session=%s | Approver='%s' | Reason='%s'",
		command, userID, sessionID, approver, reason)

	// Safety overrides are always high severity security events
	if err := ai.auditService.LogSecurityEvent(ctx, "SAFETY_OVERRIDE", description, &userID, "high", nil); err != nil {
		log.Printf("Failed to log safety override audit: %v", err)
		return fmt.Errorf("safety override audit logging failed: %w", err)
	}

	log.Printf("SAFETY OVERRIDE logged: user=%s, command=%s, approver=%s", userID, command, approver)
	return nil
}

// sanitizeAuditEntry removes sensitive information from audit entries
func (ai *AuditIntegration) sanitizeAuditEntry(entry QueryAuditEntry) QueryAuditEntry {
	// Create a copy to avoid modifying the original
	sanitized := entry

	// Sanitize context map
	if sanitized.Context != nil {
		sanitizedContext := make(map[string]string)
		for k, v := range sanitized.Context {
			// Remove potential secrets
			if ai.isSensitiveKey(k) {
				sanitizedContext[k] = "[REDACTED]"
			} else {
				sanitizedContext[k] = v
			}
		}
		sanitized.Context = sanitizedContext
	}

	// Truncate very long queries to prevent log spam
	if len(sanitized.NaturalLanguage) > 500 {
		sanitized.NaturalLanguage = sanitized.NaturalLanguage[:497] + "..."
	}

	return sanitized
}

// isSensitiveKey checks if a context key might contain sensitive information
func (ai *AuditIntegration) isSensitiveKey(key string) bool {
	sensitiveKeys := []string{
		"password", "token", "secret", "key", "auth",
		"credential", "private", "sensitive",
	}

	keyLower := key
	for _, sensitive := range sensitiveKeys {
		if keyLower == sensitive {
			return true
		}
	}

	return false
}
