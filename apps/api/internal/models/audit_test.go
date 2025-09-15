package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuditLog_CalculateChecksum(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	cluster := "test-cluster"
	namespace := "test-namespace"
	ipAddress := "192.168.1.1"
	userAgent := "test-agent"
	previousChecksum := "previous-hash"

	auditLog := &AuditLog{
		ID:               1,
		UserID:           &userID,
		SessionID:        &sessionID,
		QueryText:        "list pods",
		GeneratedCommand: "kubectl get pods",
		SafetyLevel:      SafetyLevelSafe,
		ExecutionResult:  map[string]interface{}{"status": "success"},
		ExecutionStatus:  ExecutionStatusSuccess,
		ClusterContext:   &cluster,
		NamespaceContext: &namespace,
		Timestamp:        time.Now(),
		IPAddress:        &ipAddress,
		UserAgent:        &userAgent,
	}

	checksum := auditLog.CalculateChecksum(&previousChecksum)
	assert.NotEmpty(t, checksum)
	assert.Len(t, checksum, 64) // SHA-256 hash length in hex
}

func TestAuditLog_SetChecksum(t *testing.T) {
	auditLog := &AuditLog{
		QueryText:        "test query",
		GeneratedCommand: "test command",
		SafetyLevel:      SafetyLevelSafe,
		ExecutionStatus:  ExecutionStatusSuccess,
		Timestamp:        time.Now(),
	}

	previousChecksum := "previous-hash"
	auditLog.SetChecksum(&previousChecksum)

	assert.NotEmpty(t, auditLog.Checksum)
	assert.Equal(t, &previousChecksum, auditLog.PreviousChecksum)
}

func TestAuditLog_VerifyIntegrity(t *testing.T) {
	auditLog := &AuditLog{
		QueryText:        "test query",
		GeneratedCommand: "test command",
		SafetyLevel:      SafetyLevelSafe,
		ExecutionStatus:  ExecutionStatusSuccess,
		Timestamp:        time.Now(),
	}

	// Set checksum
	auditLog.SetChecksum(nil)

	// Should verify successfully
	assert.True(t, auditLog.VerifyIntegrity())

	// Tamper with data
	auditLog.QueryText = "tampered query"

	// Should fail verification
	assert.False(t, auditLog.VerifyIntegrity())
}

func TestNewAuditLog(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	cluster := "test-cluster"
	namespace := "test-namespace"
	ipAddress := "192.168.1.1"
	userAgent := "test-agent"

	req := AuditLogRequest{
		UserID:           &userID,
		SessionID:        &sessionID,
		QueryText:        "list pods",
		GeneratedCommand: "kubectl get pods",
		SafetyLevel:      SafetyLevelSafe,
		ExecutionResult:  map[string]interface{}{"pods": 5},
		ExecutionStatus:  ExecutionStatusSuccess,
		ClusterContext:   &cluster,
		NamespaceContext: &namespace,
		IPAddress:        &ipAddress,
		UserAgent:        &userAgent,
	}

	auditLog := NewAuditLog(req)

	assert.Equal(t, req.UserID, auditLog.UserID)
	assert.Equal(t, req.SessionID, auditLog.SessionID)
	assert.Equal(t, req.QueryText, auditLog.QueryText)
	assert.Equal(t, req.GeneratedCommand, auditLog.GeneratedCommand)
	assert.Equal(t, req.SafetyLevel, auditLog.SafetyLevel)
	assert.Equal(t, req.ExecutionResult, auditLog.ExecutionResult)
	assert.Equal(t, req.ExecutionStatus, auditLog.ExecutionStatus)
	assert.Equal(t, req.ClusterContext, auditLog.ClusterContext)
	assert.Equal(t, req.NamespaceContext, auditLog.NamespaceContext)
	assert.Equal(t, req.IPAddress, auditLog.IPAddress)
	assert.Equal(t, req.UserAgent, auditLog.UserAgent)
	assert.False(t, auditLog.Timestamp.IsZero())
}

func TestIsValidSafetyLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected bool
	}{
		{"safe level", SafetyLevelSafe, true},
		{"warning level", SafetyLevelWarning, true},
		{"dangerous level", SafetyLevelDangerous, true},
		{"empty level", "", false},
		{"invalid level", "invalid", false},
		{"case sensitive", "Safe", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidSafetyLevel(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidExecutionStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"success status", ExecutionStatusSuccess, true},
		{"failed status", ExecutionStatusFailed, true},
		{"cancelled status", ExecutionStatusCancelled, true},
		{"empty status", "", false},
		{"invalid status", "invalid", false},
		{"case sensitive", "Success", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidExecutionStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuditLog_IsDangerous(t *testing.T) {
	tests := []struct {
		name        string
		safetyLevel string
		expected    bool
	}{
		{"dangerous level", SafetyLevelDangerous, true},
		{"warning level", SafetyLevelWarning, false},
		{"safe level", SafetyLevelSafe, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditLog := &AuditLog{SafetyLevel: tt.safetyLevel}
			result := auditLog.IsDangerous()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuditLog_IsSuccessful(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"success status", ExecutionStatusSuccess, true},
		{"failed status", ExecutionStatusFailed, false},
		{"cancelled status", ExecutionStatusCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditLog := &AuditLog{ExecutionStatus: tt.status}
			result := auditLog.IsSuccessful()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuditLog_GetUserIDString(t *testing.T) {
	// Test with nil UserID
	auditLog := &AuditLog{}
	assert.Equal(t, "", auditLog.GetUserIDString())

	// Test with valid UserID
	userID := uuid.New()
	auditLog.UserID = &userID
	assert.Equal(t, userID.String(), auditLog.GetUserIDString())
}

func TestAuditLog_GetSessionIDString(t *testing.T) {
	// Test with nil SessionID
	auditLog := &AuditLog{}
	assert.Equal(t, "", auditLog.GetSessionIDString())

	// Test with valid SessionID
	sessionID := uuid.New()
	auditLog.SessionID = &sessionID
	assert.Equal(t, sessionID.String(), auditLog.GetSessionIDString())
}

func TestCalculateChecksum_Deterministic(t *testing.T) {
	// Create two identical audit logs
	timestamp := time.Now()

	auditLog1 := &AuditLog{
		QueryText:        "test query",
		GeneratedCommand: "test command",
		SafetyLevel:      SafetyLevelSafe,
		ExecutionStatus:  ExecutionStatusSuccess,
		Timestamp:        timestamp,
	}

	auditLog2 := &AuditLog{
		QueryText:        "test query",
		GeneratedCommand: "test command",
		SafetyLevel:      SafetyLevelSafe,
		ExecutionStatus:  ExecutionStatusSuccess,
		Timestamp:        timestamp,
	}

	// Should produce the same checksum
	checksum1 := auditLog1.CalculateChecksum(nil)
	checksum2 := auditLog2.CalculateChecksum(nil)
	assert.Equal(t, checksum1, checksum2)

	// Change one field and checksum should be different
	auditLog2.QueryText = "different query"
	checksum3 := auditLog2.CalculateChecksum(nil)
	assert.NotEqual(t, checksum1, checksum3)
}
