package security

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityEventType represents different types of security events
type SecurityEventType string

const (
	AuthenticationAttempt SecurityEventType = "auth_attempt"
	AuthenticationFailure SecurityEventType = "auth_failure"
	AuthenticationSuccess SecurityEventType = "auth_success"
	TokenRefresh          SecurityEventType = "token_refresh"
	TokenRevocation       SecurityEventType = "token_revocation"
	PasswordChange        SecurityEventType = "password_change"
	ConfigurationChange   SecurityEventType = "config_change"
	SecurityViolation     SecurityEventType = "security_violation"
	PermissionDenied      SecurityEventType = "permission_denied"
	RateLimitExceeded     SecurityEventType = "rate_limit_exceeded"
	SuspiciousActivity    SecurityEventType = "suspicious_activity"
)

// SecurityEvent represents a security-related event for logging
type SecurityEvent struct {
	EventID     string            `json:"event_id"`
	EventType   SecurityEventType `json:"event_type"`
	Timestamp   time.Time         `json:"timestamp"`
	UserID      string            `json:"user_id,omitempty"`
	Username    string            `json:"username,omitempty"`
	ClientIP    string            `json:"client_ip"`
	UserAgent   string            `json:"user_agent,omitempty"`
	Resource    string            `json:"resource,omitempty"`
	Action      string            `json:"action,omitempty"`
	Success     bool              `json:"success"`
	ErrorCode   string            `json:"error_code,omitempty"`
	Message     string            `json:"message"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Severity    string            `json:"severity"`
	SessionID   string            `json:"session_id,omitempty"`
	RequestID   string            `json:"request_id,omitempty"`
}

// SecurityLogger handles security event logging with sanitization
type SecurityLogger struct {
	service *service
}

// NewSecurityLogger creates a new security logger instance
func NewSecurityLogger(svc *service) *SecurityLogger {
	return &SecurityLogger{
		service: svc,
	}
}

// LogSecurityEvent logs a security event with proper sanitization
func (sl *SecurityLogger) LogSecurityEvent(ctx context.Context, event SecurityEvent) {
	// Sanitize the event data
	sanitizedEvent := sl.sanitizeEvent(event)

	// Add context information
	if ginCtx, ok := ctx.(*gin.Context); ok {
		sanitizedEvent.ClientIP = GetClientIP(ginCtx)
		sanitizedEvent.UserAgent = SanitizeUserAgent(ginCtx.GetHeader("User-Agent"))
		sanitizedEvent.RequestID = GetRequestID(ginCtx)
	}

	// Set severity if not provided
	if sanitizedEvent.Severity == "" {
		sanitizedEvent.Severity = sl.determineSeverity(sanitizedEvent.EventType, sanitizedEvent.Success)
	}

	// Log the event
	eventJSON, _ := json.Marshal(sanitizedEvent)
	log.Printf("[SECURITY] %s", string(eventJSON))

	// Store in audit service if available
	// Note: Full audit integration will be completed in phase 2
	// For now, we log to standard logging system
}

// LogAuthenticationAttempt logs authentication attempts
func (sl *SecurityLogger) LogAuthenticationAttempt(ctx context.Context, username string, success bool, errorCode string) {
	event := SecurityEvent{
		EventID:   GenerateEventID(),
		EventType: AuthenticationAttempt,
		Timestamp: time.Now(),
		Username:  SanitizeUsername(username),
		Success:   success,
		ErrorCode: errorCode,
		Action:    "login",
		Resource:  "/api/v1/auth/login",
	}

	if success {
		event.EventType = AuthenticationSuccess
		event.Message = "User authentication successful"
	} else {
		event.EventType = AuthenticationFailure
		event.Message = "User authentication failed"
	}

	sl.LogSecurityEvent(ctx, event)
}

// LogConfigurationChange logs configuration changes
func (sl *SecurityLogger) LogConfigurationChange(ctx context.Context, userID, configKey, action string, metadata map[string]string) {
	event := SecurityEvent{
		EventID:   GenerateEventID(),
		EventType: ConfigurationChange,
		Timestamp: time.Now(),
		UserID:    userID,
		Action:    action,
		Resource:  fmt.Sprintf("config:%s", configKey),
		Success:   true,
		Message:   fmt.Sprintf("Configuration changed: %s", action),
		Metadata:  sl.sanitizeMetadata(metadata),
	}

	sl.LogSecurityEvent(ctx, event)
}

// LogSecurityViolation logs security violations
func (sl *SecurityLogger) LogSecurityViolation(ctx context.Context, violationType, description string, metadata map[string]string) {
	event := SecurityEvent{
		EventID:   GenerateEventID(),
		EventType: SecurityViolation,
		Timestamp: time.Now(),
		Success:   false,
		Message:   fmt.Sprintf("Security violation: %s - %s", violationType, description),
		Metadata:  sl.sanitizeMetadata(metadata),
		Severity:  "high",
	}

	sl.LogSecurityEvent(ctx, event)
}

// LogPermissionDenied logs permission denied events
func (sl *SecurityLogger) LogPermissionDenied(ctx context.Context, userID, resource, action string) {
	event := SecurityEvent{
		EventID:   GenerateEventID(),
		EventType: PermissionDenied,
		Timestamp: time.Now(),
		UserID:    userID,
		Resource:  resource,
		Action:    action,
		Success:   false,
		Message:   "Access denied due to insufficient permissions",
	}

	sl.LogSecurityEvent(ctx, event)
}

// sanitizeEvent removes sensitive information from security events
func (sl *SecurityLogger) sanitizeEvent(event SecurityEvent) SecurityEvent {
	// Remove sensitive information from metadata
	if event.Metadata != nil {
		event.Metadata = sl.sanitizeMetadata(event.Metadata)
	}

	// Sanitize message to remove potential sensitive data
	event.Message = SanitizeLogMessage(event.Message)

	// Ensure no sensitive data in error codes
	event.ErrorCode = SanitizeErrorCode(event.ErrorCode)

	return event
}

// sanitizeMetadata removes sensitive information from metadata
func (sl *SecurityLogger) sanitizeMetadata(metadata map[string]string) map[string]string {
	if metadata == nil {
		return nil
	}

	sanitized := make(map[string]string)
	sensitiveKeys := []string{"password", "token", "secret", "key", "credential", "auth"}

	for k, v := range metadata {
		key := strings.ToLower(k)
		isSensitive := false

		for _, sensitiveKey := range sensitiveKeys {
			if strings.Contains(key, sensitiveKey) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			sanitized[k] = "[REDACTED]"
		} else {
			sanitized[k] = SanitizeLogMessage(v)
		}
	}

	return sanitized
}

// determineSeverity determines the severity level based on event type and success
func (sl *SecurityLogger) determineSeverity(eventType SecurityEventType, success bool) string {
	switch eventType {
	case SecurityViolation, SuspiciousActivity:
		return "high"
	case AuthenticationFailure, PermissionDenied, RateLimitExceeded:
		return "medium"
	case AuthenticationSuccess, TokenRefresh:
		return "low"
	case ConfigurationChange:
		return "medium"
	default:
		if success {
			return "low"
		}
		return "medium"
	}
}

// Utility functions for sanitization

// SanitizeUsername removes sensitive information from usernames for logging
func SanitizeUsername(username string) string {
	if username == "" {
		return ""
	}
	// Redact email domains for privacy
	if strings.Contains(username, "@") {
		parts := strings.Split(username, "@")
		if len(parts) == 2 {
			return parts[0] + "@[DOMAIN]"
		}
	}
	return username
}

// SanitizeUserAgent sanitizes user agent strings for logging
func SanitizeUserAgent(userAgent string) string {
	// Truncate very long user agents and remove potential injection attempts
	if len(userAgent) > 200 {
		userAgent = userAgent[:200] + "..."
	}
	// Remove potential script injections
	userAgent = strings.ReplaceAll(userAgent, "<", "&lt;")
	userAgent = strings.ReplaceAll(userAgent, ">", "&gt;")
	return userAgent
}

// SanitizeLogMessage removes sensitive information from log messages
func SanitizeLogMessage(message string) string {
	if message == "" {
		return ""
	}

	// Remove potential passwords, tokens, etc.
	sensitivePatterns := []string{
		"password=",
		"token=",
		"secret=",
		"key=",
		"credential=",
		"Authorization:",
		"Bearer ",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(strings.ToLower(message), strings.ToLower(pattern)) {
			// Replace the sensitive value with [REDACTED]
			message = "[CONTAINS_SENSITIVE_DATA]"
			break
		}
	}

	return message
}

// SanitizeErrorCode ensures error codes don't contain sensitive information
func SanitizeErrorCode(errorCode string) string {
	if errorCode == "" {
		return ""
	}
	// Ensure error codes are alphanumeric and underscores only
	var sanitized strings.Builder
	for _, r := range errorCode {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			sanitized.WriteRune(r)
		}
	}
	return sanitized.String()
}

// GetClientIP safely extracts client IP from gin context
func GetClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first (for proxies)
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the chain
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	xri := c.GetHeader("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return c.ClientIP()
}

// GetRequestID gets or generates a request ID for tracking
func GetRequestID(c *gin.Context) string {
	// Check if request ID is already set
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}

	// Generate a new request ID
	requestID := GenerateEventID()
	c.Header("X-Request-ID", requestID)
	return requestID
}

// GenerateEventID generates a unique event ID for tracking
func GenerateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}