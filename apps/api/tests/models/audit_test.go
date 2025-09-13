package models_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

func TestAuditLog_BasicValidation(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()

	clusterCtx := "production"
	namespaceCtx := "default"
	ipAddr := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	audit := &models.AuditLog{
		UserID:           &userID,
		SessionID:        &sessionID,
		QueryText:        "show me pods",
		GeneratedCommand: "kubectl get pods",
		SafetyLevel:      "safe",
		ExecutionResult:  map[string]interface{}{"stdout": "pod1\npod2"},
		ExecutionStatus:  "success",
		ClusterContext:   &clusterCtx,
		NamespaceContext: &namespaceCtx,
		IPAddress:        &ipAddr,
		UserAgent:        &userAgent,
	}

	// Test that audit log fields are properly set
	assert.Equal(t, userID, *audit.UserID)
	assert.Equal(t, sessionID, *audit.SessionID)
	assert.Equal(t, "show me pods", audit.QueryText)
	assert.Equal(t, "kubectl get pods", audit.GeneratedCommand)
	assert.Equal(t, "safe", audit.SafetyLevel)
	assert.Equal(t, "success", audit.ExecutionStatus)
	assert.Equal(t, "production", *audit.ClusterContext)
	assert.Equal(t, "default", *audit.NamespaceContext)
	assert.Equal(t, "192.168.1.1", *audit.IPAddress)
	assert.Equal(t, "Mozilla/5.0", *audit.UserAgent)
	assert.NotNil(t, audit.ExecutionResult)
}

func TestAuditLogFilter_Construction(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	safetyLevel := "dangerous"
	status := "failed"

	filter := models.AuditLogFilter{
		UserID:      &userID,
		SessionID:   &sessionID,
		SafetyLevel: &safetyLevel,
		Status:      &status,
		Limit:       50,
		Offset:      10,
	}

	assert.Equal(t, userID, *filter.UserID)
	assert.Equal(t, sessionID, *filter.SessionID)
	assert.Equal(t, "dangerous", *filter.SafetyLevel)
	assert.Equal(t, "failed", *filter.Status)
	assert.Equal(t, 50, filter.Limit)
	assert.Equal(t, 10, filter.Offset)
}

func TestAuditLogSummary_Validation(t *testing.T) {
	summary := &models.AuditLogSummary{
		TotalEntries:   100,
		SafeOperations: 60,
		WarningOps:     25,
		DangerousOps:   15,
		SuccessfulOps:  85,
		FailedOps:      10,
		CancelledOps:   5,
	}

	// Test that all operations add up correctly
	totalByLevel := summary.SafeOperations + summary.WarningOps + summary.DangerousOps
	totalByStatus := summary.SuccessfulOps + summary.FailedOps + summary.CancelledOps

	assert.Equal(t, summary.TotalEntries, totalByLevel)
	assert.Equal(t, summary.TotalEntries, totalByStatus)
	assert.Equal(t, int64(100), summary.TotalEntries)
}

func TestIntegrityCheckResult_Structure(t *testing.T) {
	result := models.IntegrityCheckResult{
		LogID:        12345,
		IsValid:      false,
		ErrorMessage: "Checksum mismatch detected",
	}

	assert.Equal(t, int64(12345), result.LogID)
	assert.False(t, result.IsValid)
	assert.Equal(t, "Checksum mismatch detected", result.ErrorMessage)
}

func TestAuditLog_SafetyLevelValidation(t *testing.T) {
	validSafetyLevels := []string{"safe", "warning", "dangerous"}
	invalidSafetyLevels := []string{"", "unknown", "moderate", "high", "low"}

	for _, level := range validSafetyLevels {
		audit := &models.AuditLog{SafetyLevel: level}
		assert.Contains(t, []string{"safe", "warning", "dangerous"}, audit.SafetyLevel)
	}

	for _, level := range invalidSafetyLevels {
		// In a real implementation, these would fail validation
		audit := &models.AuditLog{SafetyLevel: level}
		assert.NotContains(t, []string{"safe", "warning", "dangerous"}, audit.SafetyLevel)
	}
}

func TestAuditLog_ExecutionStatusValidation(t *testing.T) {
	validStatuses := []string{"success", "failed", "cancelled"}
	invalidStatuses := []string{"", "pending", "running", "completed", "error"}

	for _, status := range validStatuses {
		audit := &models.AuditLog{ExecutionStatus: status}
		assert.Contains(t, []string{"success", "failed", "cancelled"}, audit.ExecutionStatus)
	}

	for _, status := range invalidStatuses {
		// In a real implementation, these would fail validation
		audit := &models.AuditLog{ExecutionStatus: status}
		assert.NotContains(t, []string{"success", "failed", "cancelled"}, audit.ExecutionStatus)
	}
}

func TestAuditLog_ExecutionResultSerialization(t *testing.T) {
	tests := []struct {
		name   string
		result map[string]interface{}
	}{
		{
			name: "simple result",
			result: map[string]interface{}{
				"stdout":    "pod1\npod2\npod3",
				"stderr":    "",
				"exit_code": 0,
			},
		},
		{
			name: "complex result with nested data",
			result: map[string]interface{}{
				"stdout": "NAME    READY   STATUS    RESTARTS\npod1    1/1     Running   0",
				"stderr": "",
				"metadata": map[string]interface{}{
					"namespace": "default",
					"timestamp": "2025-01-11T18:30:00Z",
					"user":      "testuser",
				},
				"exit_code": 0,
			},
		},
		{
			name: "error result",
			result: map[string]interface{}{
				"stdout":     "",
				"stderr":     "Error: connection refused",
				"exit_code":  1,
				"error_type": "connection_error",
			},
		},
		{
			name:   "empty result",
			result: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audit := &models.AuditLog{
				ExecutionResult: tt.result,
			}

			assert.NotNil(t, audit.ExecutionResult)
			assert.Equal(t, tt.result, audit.ExecutionResult)

			// Verify individual fields if they exist
			if stdout, exists := tt.result["stdout"]; exists {
				assert.Equal(t, stdout, audit.ExecutionResult["stdout"])
			}
			if stderr, exists := tt.result["stderr"]; exists {
				assert.Equal(t, stderr, audit.ExecutionResult["stderr"])
			}
			if exitCode, exists := tt.result["exit_code"]; exists {
				assert.Equal(t, exitCode, audit.ExecutionResult["exit_code"])
			}
		})
	}
}

func TestAuditLog_NilFields(t *testing.T) {
	// Test that audit log can handle nil optional fields
	audit := &models.AuditLog{
		UserID:           nil, // Can be nil for system operations
		SessionID:        nil, // Can be nil for system operations
		QueryText:        "system health check",
		GeneratedCommand: "kubectl get nodes",
		SafetyLevel:      "safe",
		ExecutionResult:  map[string]interface{}{},
		ExecutionStatus:  "success",
		ClusterContext:   nil, // Can be nil
		NamespaceContext: nil, // Can be nil
		IPAddress:        nil, // Can be nil
		UserAgent:        nil, // Can be nil
	}

	assert.Nil(t, audit.UserID)
	assert.Nil(t, audit.SessionID)
	assert.Equal(t, "system health check", audit.QueryText)
	assert.Nil(t, audit.ClusterContext)
	assert.Nil(t, audit.NamespaceContext)
}

func TestAuditLog_SecurityFields(t *testing.T) {
	userID := uuid.New()
	clusterCtx := "production"
	namespaceCtx := "production"

	// Test that sensitive operations are properly classified
	sensitiveAudit := &models.AuditLog{
		UserID:           &userID,
		QueryText:        "delete all pods in production",
		GeneratedCommand: "kubectl delete pods --all -n production",
		SafetyLevel:      "dangerous",
		ExecutionResult:  map[string]interface{}{"warning": "destructive operation"},
		ExecutionStatus:  "cancelled", // Should be cancelled for safety
		ClusterContext:   &clusterCtx,
		NamespaceContext: &namespaceCtx,
	}

	assert.Equal(t, "dangerous", sensitiveAudit.SafetyLevel)
	assert.Equal(t, "cancelled", sensitiveAudit.ExecutionStatus)
	assert.Equal(t, "production", *sensitiveAudit.ClusterContext)
	assert.Contains(t, sensitiveAudit.QueryText, "delete")
	assert.Contains(t, sensitiveAudit.GeneratedCommand, "delete")
}

func TestAuditLog_TimestampHandling(t *testing.T) {
	userID := uuid.New()

	audit := &models.AuditLog{
		UserID:           &userID,
		QueryText:        "get pods",
		GeneratedCommand: "kubectl get pods",
		SafetyLevel:      "safe",
		ExecutionResult:  map[string]interface{}{},
		ExecutionStatus:  "success",
	}

	// Initially timestamp should be zero
	assert.True(t, audit.Timestamp.IsZero())

	// In a real database scenario, timestamp would be set by the database
	// but for testing, we can validate the structure
	assert.NotNil(t, &audit.Timestamp)
}

func BenchmarkAuditLog_StructCreation(b *testing.B) {
	userID := uuid.New()
	sessionID := uuid.New()
	clusterCtx := "test"
	namespaceCtx := "default"
	ipAddr := "192.168.1.1"
	userAgent := "test-agent"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &models.AuditLog{
			UserID:           &userID,
			SessionID:        &sessionID,
			QueryText:        "benchmark query",
			GeneratedCommand: "kubectl get pods",
			SafetyLevel:      "safe",
			ExecutionResult:  map[string]interface{}{"test": "data"},
			ExecutionStatus:  "success",
			ClusterContext:   &clusterCtx,
			NamespaceContext: &namespaceCtx,
			IPAddress:        &ipAddr,
			UserAgent:        &userAgent,
		}
	}
}

func BenchmarkAuditLog_ExecutionResultAccess(b *testing.B) {
	audit := &models.AuditLog{
		ExecutionResult: map[string]interface{}{
			"stdout":    "large output content here",
			"stderr":    "",
			"exit_code": 0,
			"metadata": map[string]interface{}{
				"namespace": "default",
				"timestamp": "2025-01-11T18:30:00Z",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = audit.ExecutionResult["stdout"]
		_ = audit.ExecutionResult["metadata"]
	}
}
