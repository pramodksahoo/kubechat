package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an immutable audit log entry with cryptographic integrity
type AuditLog struct {
	ID               int64                  `json:"id" db:"id"`
	UserID           *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	SessionID        *uuid.UUID             `json:"session_id,omitempty" db:"session_id"`
	QueryText        string                 `json:"query_text" db:"query_text"`
	GeneratedCommand string                 `json:"generated_command" db:"generated_command"`
	SafetyLevel      string                 `json:"safety_level" db:"safety_level"`
	ExecutionResult  map[string]interface{} `json:"execution_result,omitempty" db:"execution_result"`
	ExecutionStatus  string                 `json:"execution_status" db:"execution_status"`
	ClusterContext   *string                `json:"cluster_context,omitempty" db:"cluster_context"`
	NamespaceContext *string                `json:"namespace_context,omitempty" db:"namespace_context"`
	Timestamp        time.Time              `json:"timestamp" db:"timestamp"`
	IPAddress        *string                `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent        *string                `json:"user_agent,omitempty" db:"user_agent"`
	Checksum         string                 `json:"checksum" db:"checksum"`
	PreviousChecksum *string                `json:"previous_checksum,omitempty" db:"previous_checksum"`
}

// AuditLogRequest represents the data needed to create an audit log entry
type AuditLogRequest struct {
	UserID           *uuid.UUID             `json:"user_id,omitempty"`
	SessionID        *uuid.UUID             `json:"session_id,omitempty"`
	QueryText        string                 `json:"query_text" validate:"required"`
	GeneratedCommand string                 `json:"generated_command" validate:"required"`
	SafetyLevel      string                 `json:"safety_level" validate:"required,oneof=safe warning dangerous"`
	ExecutionResult  map[string]interface{} `json:"execution_result,omitempty"`
	ExecutionStatus  string                 `json:"execution_status" validate:"required,oneof=success failed cancelled"`
	ClusterContext   *string                `json:"cluster_context,omitempty"`
	NamespaceContext *string                `json:"namespace_context,omitempty"`
	IPAddress        *string                `json:"ip_address,omitempty"`
	UserAgent        *string                `json:"user_agent,omitempty"`
}

// Safety level constants
const (
	SafetyLevelSafe      = "safe"
	SafetyLevelWarning   = "warning"
	SafetyLevelDangerous = "dangerous"
)

// Execution status constants
const (
	ExecutionStatusSuccess   = "success"
	ExecutionStatusFailed    = "failed"
	ExecutionStatusCancelled = "cancelled"
)

// CalculateChecksum calculates SHA-256 checksum for audit log integrity
func (a *AuditLog) CalculateChecksum(previousChecksum *string) string {
	// Convert execution result to JSON string for consistent hashing
	executionResultBytes, _ := json.Marshal(a.ExecutionResult)
	executionResultStr := string(executionResultBytes)

	// Create deterministic string for hashing
	checksumInput := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s",
		uuidPtrToString(a.UserID),
		uuidPtrToString(a.SessionID),
		a.QueryText,
		a.GeneratedCommand,
		a.SafetyLevel,
		executionResultStr,
		a.ExecutionStatus,
		ptrToString(a.ClusterContext),
		ptrToString(a.NamespaceContext),
		a.Timestamp.Format(time.RFC3339Nano),
		ptrToString(a.IPAddress),
		ptrToString(a.UserAgent),
		ptrToString(previousChecksum),
	)

	// Calculate SHA-256 hash
	hash := sha256.Sum256([]byte(checksumInput))
	return hex.EncodeToString(hash[:])
}

// SetChecksum calculates and sets the checksum for the audit log entry
func (a *AuditLog) SetChecksum(previousChecksum *string) {
	a.PreviousChecksum = previousChecksum
	a.Checksum = a.CalculateChecksum(previousChecksum)
}

// VerifyIntegrity verifies the integrity of the audit log entry
func (a *AuditLog) VerifyIntegrity() bool {
	expectedChecksum := a.CalculateChecksum(a.PreviousChecksum)
	return a.Checksum == expectedChecksum
}

// NewAuditLog creates a new audit log entry from request data
func NewAuditLog(req AuditLogRequest) *AuditLog {
	return &AuditLog{
		UserID:           req.UserID,
		SessionID:        req.SessionID,
		QueryText:        req.QueryText,
		GeneratedCommand: req.GeneratedCommand,
		SafetyLevel:      req.SafetyLevel,
		ExecutionResult:  req.ExecutionResult,
		ExecutionStatus:  req.ExecutionStatus,
		ClusterContext:   req.ClusterContext,
		NamespaceContext: req.NamespaceContext,
		Timestamp:        time.Now(),
		IPAddress:        req.IPAddress,
		UserAgent:        req.UserAgent,
	}
}

// IsValidSafetyLevel checks if the safety level is valid
func IsValidSafetyLevel(level string) bool {
	return level == SafetyLevelSafe || level == SafetyLevelWarning || level == SafetyLevelDangerous
}

// IsValidExecutionStatus checks if the execution status is valid
func IsValidExecutionStatus(status string) bool {
	return status == ExecutionStatusSuccess || status == ExecutionStatusFailed || status == ExecutionStatusCancelled
}

// IsDangerous returns true if the audit log represents a dangerous operation
func (a *AuditLog) IsDangerous() bool {
	return a.SafetyLevel == SafetyLevelDangerous
}

// IsSuccessful returns true if the operation was successful
func (a *AuditLog) IsSuccessful() bool {
	return a.ExecutionStatus == ExecutionStatusSuccess
}

// GetUserIDString returns the user ID as a string, or empty string if nil
func (a *AuditLog) GetUserIDString() string {
	if a.UserID == nil {
		return ""
	}
	return a.UserID.String()
}

// GetSessionIDString returns the session ID as a string, or empty string if nil
func (a *AuditLog) GetSessionIDString() string {
	if a.SessionID == nil {
		return ""
	}
	return a.SessionID.String()
}

// Helper function to convert pointer to string
func ptrToString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func uuidPtrToString(ptr *uuid.UUID) string {
	if ptr == nil {
		return ""
	}
	return ptr.String()
}

// AuditLogFilter represents filtering options for audit log queries
type AuditLogFilter struct {
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	SessionID   *uuid.UUID `json:"session_id,omitempty"`
	SafetyLevel *string    `json:"safety_level,omitempty"`
	Status      *string    `json:"status,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
}

// AuditLogSummary represents summary statistics for audit logs
type AuditLogSummary struct {
	TotalEntries   int64 `json:"total_entries"`
	SafeOperations int64 `json:"safe_operations"`
	WarningOps     int64 `json:"warning_operations"`
	DangerousOps   int64 `json:"dangerous_operations"`
	SuccessfulOps  int64 `json:"successful_operations"`
	FailedOps      int64 `json:"failed_operations"`
	CancelledOps   int64 `json:"cancelled_operations"`
}

// IntegrityCheckResult represents the result of an audit log integrity check
type IntegrityCheckResult struct {
	LogID        int64  `json:"log_id"`
	IsValid      bool   `json:"is_valid"`
	ErrorMessage string `json:"error_message,omitempty"`
}
