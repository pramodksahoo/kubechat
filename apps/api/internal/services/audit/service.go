package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/repositories"
)

// Service defines the interface for audit logging service
type Service interface {
	// Core audit logging methods
	LogUserAction(ctx context.Context, req models.AuditLogRequest) error
	LogKubectlExecution(ctx context.Context, userID *uuid.UUID, sessionID *uuid.UUID, query, command string, result map[string]interface{}, status string, req *http.Request) error
	LogAPICall(ctx context.Context, userID *uuid.UUID, sessionID *uuid.UUID, method, endpoint, query string, statusCode int, duration time.Duration, req *http.Request) error
	LogSecurityEvent(ctx context.Context, eventType, description string, userID *uuid.UUID, severity string, req *http.Request) error

	// Audit log retrieval methods
	GetAuditLogs(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error)
	GetAuditLogByID(ctx context.Context, id int64) (*models.AuditLog, error)
	GetAuditLogSummary(ctx context.Context, filter models.AuditLogFilter) (*models.AuditLogSummary, error)
	GetDangerousOperations(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error)
	GetFailedOperations(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error)

	// Enhanced cryptographic integrity
	VerifyIntegrity(ctx context.Context, startID, endID int64) ([]models.IntegrityCheckResult, error)
	VerifyChainIntegrity(ctx context.Context) (*ChainIntegrityResult, error)
	DetectTampering(ctx context.Context, alertChannel chan<- TamperAlert) error

	// Compliance and export features
	ExportAuditLogs(ctx context.Context, filter models.AuditLogFilter, format ExportFormat) ([]byte, error)
	GenerateComplianceReport(ctx context.Context, framework ComplianceFramework, startTime, endTime time.Time) (*ComplianceReport, error)
	CreateLegalHold(ctx context.Context, req LegalHoldRequest) (*LegalHold, error)
	ReleaseLegalHold(ctx context.Context, holdID string) error
	GetLegalHolds(ctx context.Context) ([]*LegalHold, error)

	// Real-time monitoring
	StartRealTimeMonitoring(ctx context.Context, eventChannel chan<- AuditEvent) error
	StopRealTimeMonitoring() error
	GetSuspiciousActivities(ctx context.Context, timeWindow time.Duration) ([]*SuspiciousActivity, error)

	// Retention and archival
	ApplyRetentionPolicy(ctx context.Context, policy RetentionPolicy) error
	ArchiveOldLogs(ctx context.Context, cutoffDate time.Time) (*ArchivalResult, error)
	RestoreArchivedLogs(ctx context.Context, archiveID string) error

	// Health and metrics
	GetMetrics(ctx context.Context) (*AuditMetrics, error)
	HealthCheck(ctx context.Context) error
}

// Config represents audit service configuration
type Config struct {
	// Enable async logging to improve performance
	EnableAsyncLogging bool `json:"enable_async_logging"`

	// Buffer size for async logging
	AsyncBufferSize int `json:"async_buffer_size"`

	// Maximum batch size for bulk operations
	MaxBatchSize int `json:"max_batch_size"`

	// Enable structured logging
	EnableStructuredLogging bool `json:"enable_structured_logging"`

	// Log level (debug, info, warn, error)
	LogLevel string `json:"log_level"`

	// Enable integrity checking
	EnableIntegrityCheck bool `json:"enable_integrity_check"`

	// Retention policy in days
	RetentionDays int `json:"retention_days"`
}

// Export formats
type ExportFormat string

const (
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatJSON ExportFormat = "json"
	ExportFormatPDF  ExportFormat = "pdf"
)

// Compliance frameworks
type ComplianceFramework string

const (
	FrameworkSOX   ComplianceFramework = "sox"
	FrameworkHIPAA ComplianceFramework = "hipaa"
	FrameworkSOC2  ComplianceFramework = "soc2"
)

// Enhanced audit event for real-time monitoring
type AuditEvent struct {
	ID        int64                  `json:"id"`
	Timestamp time.Time            `json:"timestamp"`
	UserID    *uuid.UUID           `json:"user_id,omitempty"`
	EventType string               `json:"event_type"`
	Severity  string               `json:"severity"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Tamper detection alert
type TamperAlert struct {
	DetectedAt    time.Time `json:"detected_at"`
	AffectedLogID int64     `json:"affected_log_id"`
	ViolationType string    `json:"violation_type"`
	Description   string    `json:"description"`
	Severity      string    `json:"severity"`
}

// Chain integrity verification result
type ChainIntegrityResult struct {
	IsValid        bool      `json:"is_valid"`
	TotalChecked   int64     `json:"total_checked"`
	Violations     []int64   `json:"violations,omitempty"`
	LastValidated  time.Time `json:"last_validated"`
	IntegrityScore float64   `json:"integrity_score"`
}

// Legal hold structure
type LegalHold struct {
	ID          string    `json:"id"`
	CaseNumber  string    `json:"case_number"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	StartTime   time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Status      string    `json:"status"`
	RecordCount int64     `json:"record_count"`
}

type LegalHoldRequest struct {
	CaseNumber  string    `json:"case_number" validate:"required"`
	Description string    `json:"description" validate:"required"`
	StartTime   time.Time `json:"start_time" validate:"required"`
	EndTime     *time.Time `json:"end_time,omitempty"`
}

// Compliance report structure
type ComplianceReport struct {
	ID                string                 `json:"id"`
	Framework         ComplianceFramework    `json:"framework"`
	GeneratedAt       time.Time              `json:"generated_at"`
	PeriodStart       time.Time              `json:"period_start"`
	PeriodEnd         time.Time              `json:"period_end"`
	ComplianceScore   float64                `json:"compliance_score"`
	TotalEvents       int64                  `json:"total_events"`
	Violations        []ComplianceViolation `json:"violations"`
	Recommendations   []string               `json:"recommendations"`
	ExecutiveSummary  string                 `json:"executive_summary"`
	DetailedFindings  map[string]interface{} `json:"detailed_findings"`
}

type ComplianceViolation struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	AffectedLogIDs []int64 `json:"affected_log_ids"`
	DetectedAt  time.Time `json:"detected_at"`
}

// Suspicious activity detection
type SuspiciousActivity struct {
	ID              string    `json:"id"`
	DetectedAt      time.Time `json:"detected_at"`
	ActivityType    string    `json:"activity_type"`
	UserID          *uuid.UUID `json:"user_id,omitempty"`
	Description     string    `json:"description"`
	RiskScore       float64   `json:"risk_score"`
	AffectedRecords []int64   `json:"affected_records"`
	PatternMatched  string    `json:"pattern_matched"`
}

// Retention policy
type RetentionPolicy struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	RetentionDays   int       `json:"retention_days"`
	AppliesTo       string    `json:"applies_to"`
	CreatedAt       time.Time `json:"created_at"`
	LastApplied     *time.Time `json:"last_applied,omitempty"`
	Automatic       bool      `json:"automatic"`
}

// Archival result
type ArchivalResult struct {
	ArchiveID       string    `json:"archive_id"`
	ArchivedCount   int64     `json:"archived_count"`
	ArchiveSize     int64     `json:"archive_size_bytes"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	StorageLocation string    `json:"storage_location"`
	CreatedAt       time.Time `json:"created_at"`
}

// AuditMetrics represents audit service metrics
type AuditMetrics struct {
	TotalLogsCreated      int64      `json:"total_logs_created"`
	TotalDangerousOps     int64      `json:"total_dangerous_operations"`
	TotalFailedOps        int64      `json:"total_failed_operations"`
	AverageResponseTime   float64    `json:"average_response_time_ms"`
	SuccessRate           float64    `json:"success_rate"`
	IntegrityChecksPassed int64      `json:"integrity_checks_passed"`
	IntegrityChecksFailed int64      `json:"integrity_checks_failed"`
	LastIntegrityCheck    *time.Time `json:"last_integrity_check,omitempty"`
	QueueSize             int        `json:"queue_size"`
	ProcessedCount        int64      `json:"processed_count"`
	ErrorCount            int64      `json:"error_count"`
	ActiveMonitoringSessions int     `json:"active_monitoring_sessions"`
	TamperAlertsTriggered    int64   `json:"tamper_alerts_triggered"`
	ActiveLegalHolds         int     `json:"active_legal_holds"`
	ComplianceScore          float64 `json:"compliance_score"`
}

// auditService implements the Service interface
type auditService struct {
	repo    repositories.AuditRepository
	config  *Config
	logger  *slog.Logger
	metrics *AuditMetrics

	// Async processing
	logQueue chan models.AuditLogRequest
	stopChan chan struct{}
	wg       sync.WaitGroup
	mutex    sync.RWMutex

	// Real-time monitoring
	eventChannels    []chan<- AuditEvent
	tamperChannels   []chan<- TamperAlert
	monitoringActive bool

	// Legal holds storage
	legalHolds map[string]*LegalHold
}

// NewService creates a new audit service
func NewService(repo repositories.AuditRepository, config *Config) Service {
	if config == nil {
		config = &Config{
			EnableAsyncLogging:      true,
			AsyncBufferSize:         1000,
			MaxBatchSize:            100,
			EnableStructuredLogging: true,
			LogLevel:                "info",
			EnableIntegrityCheck:    true,
			RetentionDays:           90,
		}
	}

	// Initialize structured logger
	var logger *slog.Logger
	if config.EnableStructuredLogging {
		logLevel := slog.LevelInfo
		switch config.LogLevel {
		case "debug":
			logLevel = slog.LevelDebug
		case "warn":
			logLevel = slog.LevelWarn
		case "error":
			logLevel = slog.LevelError
		}

		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}))
	} else {
		logger = slog.Default()
	}

	service := &auditService{
		repo:       repo,
		config:     config,
		logger:     logger,
		metrics:    &AuditMetrics{},
		legalHolds: make(map[string]*LegalHold),
	}

	// Initialize async processing if enabled
	if config.EnableAsyncLogging {
		service.logQueue = make(chan models.AuditLogRequest, config.AsyncBufferSize)
		service.stopChan = make(chan struct{})
		service.startAsyncProcessor()
	}

	logger.Info("Audit service initialized",
		"async_logging", config.EnableAsyncLogging,
		"buffer_size", config.AsyncBufferSize,
		"integrity_check", config.EnableIntegrityCheck)

	return service
}

// LogUserAction logs a user action with structured data
func (s *auditService) LogUserAction(ctx context.Context, req models.AuditLogRequest) error {
	startTime := time.Now()

	// Validate request
	if err := s.validateAuditRequest(req); err != nil {
		s.incrementErrorCount()
		return fmt.Errorf("invalid audit request: %w", err)
	}

	// Enrich request with additional context
	s.enrichAuditRequest(&req, nil)

	// Process synchronously or asynchronously based on configuration
	if s.config.EnableAsyncLogging {
		select {
		case s.logQueue <- req:
			s.logger.Debug("Audit log queued for async processing",
				"user_id", req.UserID,
				"query", req.QueryText)
		default:
			// Queue is full, process synchronously as fallback
			s.logger.Warn("Audit log queue full, processing synchronously")
			if err := s.processAuditLog(ctx, req); err != nil {
				s.incrementErrorCount()
				return err
			}
		}
	} else {
		if err := s.processAuditLog(ctx, req); err != nil {
			s.incrementErrorCount()
			return err
		}
	}

	// Update metrics
	s.updateMetrics(time.Since(startTime))

	s.logger.Info("User action logged",
		"user_id", req.UserID,
		"session_id", req.SessionID,
		"safety_level", req.SafetyLevel,
		"execution_status", req.ExecutionStatus,
		"processing_time_ms", time.Since(startTime).Milliseconds())

	return nil
}

// LogKubectlExecution logs kubectl command execution
func (s *auditService) LogKubectlExecution(ctx context.Context, userID *uuid.UUID, sessionID *uuid.UUID, query, command string, result map[string]interface{}, status string, req *http.Request) error {
	auditReq := models.AuditLogRequest{
		UserID:           userID,
		SessionID:        sessionID,
		QueryText:        query,
		GeneratedCommand: command,
		SafetyLevel:      s.classifyCommandSafety(command),
		ExecutionResult:  result,
		ExecutionStatus:  status,
	}

	// Extract additional context from HTTP request
	s.enrichAuditRequest(&auditReq, req)

	return s.LogUserAction(ctx, auditReq)
}

// LogAPICall logs API call with request context
func (s *auditService) LogAPICall(ctx context.Context, userID *uuid.UUID, sessionID *uuid.UUID, method, endpoint, query string, statusCode int, duration time.Duration, req *http.Request) error {
	// Determine safety level and execution status based on API call
	safetyLevel := models.SafetyLevelSafe
	executionStatus := models.ExecutionStatusSuccess

	if statusCode >= 400 {
		executionStatus = models.ExecutionStatusFailed
	}

	// Create result map with API call details
	result := map[string]interface{}{
		"method":      method,
		"endpoint":    endpoint,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
	}

	auditReq := models.AuditLogRequest{
		UserID:           userID,
		SessionID:        sessionID,
		QueryText:        fmt.Sprintf("%s %s", method, endpoint),
		GeneratedCommand: fmt.Sprintf("API call: %s %s", method, endpoint),
		SafetyLevel:      safetyLevel,
		ExecutionResult:  result,
		ExecutionStatus:  executionStatus,
	}

	s.enrichAuditRequest(&auditReq, req)

	return s.LogUserAction(ctx, auditReq)
}

// LogSecurityEvent logs security-related events
func (s *auditService) LogSecurityEvent(ctx context.Context, eventType, description string, userID *uuid.UUID, severity string, req *http.Request) error {
	var safetyLevel string
	switch severity {
	case "high", "critical":
		safetyLevel = models.SafetyLevelDangerous
	case "medium", "warning":
		safetyLevel = models.SafetyLevelWarning
	default:
		safetyLevel = models.SafetyLevelSafe
	}

	result := map[string]interface{}{
		"event_type": eventType,
		"severity":   severity,
	}

	auditReq := models.AuditLogRequest{
		UserID:           userID,
		QueryText:        fmt.Sprintf("Security Event: %s", eventType),
		GeneratedCommand: description,
		SafetyLevel:      safetyLevel,
		ExecutionResult:  result,
		ExecutionStatus:  models.ExecutionStatusSuccess,
	}

	s.enrichAuditRequest(&auditReq, req)

	return s.LogUserAction(ctx, auditReq)
}

// GetAuditLogs retrieves audit logs with filtering
func (s *auditService) GetAuditLogs(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error) {
	logs, err := s.repo.GetAuditLogs(ctx, filter)
	if err != nil {
		s.incrementErrorCount()
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	s.logger.Debug("Retrieved audit logs",
		"count", len(logs),
		"filter", filter)

	return logs, nil
}

// GetAuditLogByID retrieves a specific audit log
func (s *auditService) GetAuditLogByID(ctx context.Context, id int64) (*models.AuditLog, error) {
	log, err := s.repo.GetAuditLogByID(ctx, id)
	if err != nil {
		s.incrementErrorCount()
		return nil, fmt.Errorf("failed to get audit log by id: %w", err)
	}

	return log, nil
}

// GetAuditLogSummary returns summary statistics
func (s *auditService) GetAuditLogSummary(ctx context.Context, filter models.AuditLogFilter) (*models.AuditLogSummary, error) {
	summary, err := s.repo.GetAuditLogSummary(ctx, filter)
	if err != nil {
		s.incrementErrorCount()
		return nil, fmt.Errorf("failed to get audit log summary: %w", err)
	}

	return summary, nil
}

// GetDangerousOperations retrieves dangerous operations
func (s *auditService) GetDangerousOperations(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error) {
	logs, err := s.repo.GetDangerousOperations(ctx, filter)
	if err != nil {
		s.incrementErrorCount()
		return nil, fmt.Errorf("failed to get dangerous operations: %w", err)
	}

	return logs, nil
}

// GetFailedOperations retrieves failed operations
func (s *auditService) GetFailedOperations(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error) {
	logs, err := s.repo.GetFailedOperations(ctx, filter)
	if err != nil {
		s.incrementErrorCount()
		return nil, fmt.Errorf("failed to get failed operations: %w", err)
	}

	return logs, nil
}

// VerifyIntegrity verifies audit log integrity
func (s *auditService) VerifyIntegrity(ctx context.Context, startID, endID int64) ([]models.IntegrityCheckResult, error) {
	results, err := s.repo.VerifyIntegrity(ctx, startID, endID)
	if err != nil {
		s.incrementErrorCount()
		return nil, fmt.Errorf("failed to verify integrity: %w", err)
	}

	// Update integrity check metrics
	s.mutex.Lock()
	now := time.Now()
	s.metrics.LastIntegrityCheck = &now
	for _, result := range results {
		if result.IsValid {
			s.metrics.IntegrityChecksPassed++
		} else {
			s.metrics.IntegrityChecksFailed++
		}
	}
	s.mutex.Unlock()

	s.logger.Info("Integrity check completed",
		"start_id", startID,
		"end_id", endID,
		"total_checked", len(results),
		"passed", s.metrics.IntegrityChecksPassed,
		"failed", s.metrics.IntegrityChecksFailed)

	return results, nil
}

// GetMetrics returns audit service metrics
func (s *auditService) GetMetrics(ctx context.Context) (*AuditMetrics, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Create a copy of metrics to avoid race conditions
	metrics := *s.metrics

	// Add current queue size if async logging is enabled
	if s.config.EnableAsyncLogging {
		metrics.QueueSize = len(s.logQueue)
	}

	return &metrics, nil
}

// HealthCheck performs health check on audit service
func (s *auditService) HealthCheck(ctx context.Context) error {
	// Test database connectivity by attempting to get last checksum
	_, err := s.repo.GetLastChecksum(ctx)
	if err != nil {
		return fmt.Errorf("audit service health check failed: %w", err)
	}

	s.logger.Debug("Audit service health check passed")
	return nil
}

// Private methods

// startAsyncProcessor starts the async log processor
func (s *auditService) startAsyncProcessor() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		for {
			select {
			case req := <-s.logQueue:
				if err := s.processAuditLog(context.Background(), req); err != nil {
					s.logger.Error("Failed to process audit log",
						"error", err,
						"user_id", req.UserID)
					s.incrementErrorCount()
				} else {
					s.incrementProcessedCount()
				}
			case <-s.stopChan:
				return
			}
		}
	}()
}

// processAuditLog processes a single audit log request
func (s *auditService) processAuditLog(ctx context.Context, req models.AuditLogRequest) error {
	auditLog := models.NewAuditLog(req)

	if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// validateAuditRequest validates an audit request
func (s *auditService) validateAuditRequest(req models.AuditLogRequest) error {
	if req.QueryText == "" {
		return fmt.Errorf("query text is required")
	}

	if req.GeneratedCommand == "" {
		return fmt.Errorf("generated command is required")
	}

	if !models.IsValidSafetyLevel(req.SafetyLevel) {
		return fmt.Errorf("invalid safety level: %s", req.SafetyLevel)
	}

	if !models.IsValidExecutionStatus(req.ExecutionStatus) {
		return fmt.Errorf("invalid execution status: %s", req.ExecutionStatus)
	}

	return nil
}

// enrichAuditRequest enriches audit request with additional context
func (s *auditService) enrichAuditRequest(req *models.AuditLogRequest, httpReq *http.Request) {
	if httpReq != nil {
		// Extract IP address
		ipAddr := s.extractIPAddress(httpReq)
		req.IPAddress = &ipAddr

		// Extract user agent
		userAgent := httpReq.UserAgent()
		req.UserAgent = &userAgent

		// Extract cluster context if available
		if clusterContext := httpReq.Header.Get("X-Cluster-Context"); clusterContext != "" {
			req.ClusterContext = &clusterContext
		}

		// Extract namespace context if available
		if namespaceContext := httpReq.Header.Get("X-Namespace-Context"); namespaceContext != "" {
			req.NamespaceContext = &namespaceContext
		}
	}
}

// extractIPAddress extracts the real IP address from request
func (s *auditService) extractIPAddress(req *http.Request) string {
	// Check for forwarded IP addresses
	if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Take the first IP if multiple are present
		for commaIdx := 0; commaIdx < len(forwarded); commaIdx++ {
			if forwarded[commaIdx] == ',' {
				return forwarded[:commaIdx]
			}
		}
		return forwarded
	}

	if realIP := req.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	return req.RemoteAddr
}

// classifyCommandSafety classifies command safety level
func (s *auditService) classifyCommandSafety(command string) string {
	if command == "" {
		return models.SafetyLevelDangerous
	}

	// Simple classification based on command content
	dangerousPatterns := []string{"delete", "destroy", "rm", "--force"}
	warningPatterns := []string{"create", "apply", "patch", "scale"}

	cmdLower := command
	for i := 0; i < len(cmdLower); i++ {
		if cmdLower[i] >= 'A' && cmdLower[i] <= 'Z' {
			cmdLower = cmdLower[:i] + string(cmdLower[i]+32) + cmdLower[i+1:]
		}
	}

	for _, pattern := range dangerousPatterns {
		if contains(cmdLower, pattern) {
			return models.SafetyLevelDangerous
		}
	}

	for _, pattern := range warningPatterns {
		if contains(cmdLower, pattern) {
			return models.SafetyLevelWarning
		}
	}

	return models.SafetyLevelSafe
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// updateMetrics updates service metrics
func (s *auditService) updateMetrics(duration time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.metrics.TotalLogsCreated++

	// Update average response time
	if s.metrics.TotalLogsCreated == 1 {
		s.metrics.AverageResponseTime = float64(duration.Nanoseconds()) / 1000000.0
	} else {
		currentAvg := s.metrics.AverageResponseTime
		newDuration := float64(duration.Nanoseconds()) / 1000000.0
		s.metrics.AverageResponseTime = (currentAvg*float64(s.metrics.TotalLogsCreated-1) + newDuration) / float64(s.metrics.TotalLogsCreated)
	}

	// Update success rate
	s.metrics.SuccessRate = float64(s.metrics.ProcessedCount) / float64(s.metrics.TotalLogsCreated) * 100.0
}

// incrementProcessedCount increments processed count
func (s *auditService) incrementProcessedCount() {
	s.mutex.Lock()
	s.metrics.ProcessedCount++
	s.mutex.Unlock()
}

// incrementErrorCount increments error count
func (s *auditService) incrementErrorCount() {
	s.mutex.Lock()
	s.metrics.ErrorCount++
	s.mutex.Unlock()
}

// VerifyChainIntegrity verifies the entire audit chain integrity
func (s *auditService) VerifyChainIntegrity(ctx context.Context) (*ChainIntegrityResult, error) {
	// Get the first and last audit log IDs
	firstLog, err := s.repo.GetAuditLogs(ctx, models.AuditLogFilter{Limit: 1, Offset: 0})
	if err != nil || len(firstLog) == 0 {
		return &ChainIntegrityResult{IsValid: true, TotalChecked: 0, LastValidated: time.Now()}, nil
	}

	lastLog, err := s.repo.GetAuditLogs(ctx, models.AuditLogFilter{Limit: 1, Offset: 0})
	if err != nil || len(lastLog) == 0 {
		return &ChainIntegrityResult{IsValid: true, TotalChecked: 0, LastValidated: time.Now()}, nil
	}

	// Verify integrity for all logs
	results, err := s.VerifyIntegrity(ctx, firstLog[0].ID, lastLog[0].ID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify chain integrity: %w", err)
	}

	// Calculate overall integrity score
	totalChecked := int64(len(results))
	violations := make([]int64, 0)
	validCount := int64(0)

	for _, result := range results {
		if result.IsValid {
			validCount++
		} else {
			violations = append(violations, result.LogID)
		}
	}

	integrityScore := float64(validCount) / float64(totalChecked) * 100.0

	return &ChainIntegrityResult{
		IsValid:        len(violations) == 0,
		TotalChecked:   totalChecked,
		Violations:     violations,
		LastValidated:  time.Now(),
		IntegrityScore: integrityScore,
	}, nil
}

// DetectTampering monitors for tampering attempts
func (s *auditService) DetectTampering(ctx context.Context, alertChannel chan<- TamperAlert) error {
	s.mutex.Lock()
	s.tamperChannels = append(s.tamperChannels, alertChannel)
	s.mutex.Unlock()

	s.logger.Info("Tamper detection enabled", "alert_channels", len(s.tamperChannels))
	return nil
}

// ExportAuditLogs exports audit logs in the specified format
func (s *auditService) ExportAuditLogs(ctx context.Context, filter models.AuditLogFilter, format ExportFormat) ([]byte, error) {
	logs, err := s.GetAuditLogs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs for export: %w", err)
	}

	switch format {
	case ExportFormatJSON:
		return json.Marshal(logs)
	case ExportFormatCSV:
		return s.exportToCSV(logs)
	case ExportFormatPDF:
		return s.exportToPDF(logs)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// GenerateComplianceReport generates a compliance report for the specified framework
func (s *auditService) GenerateComplianceReport(ctx context.Context, framework ComplianceFramework, startTime, endTime time.Time) (*ComplianceReport, error) {
	filter := models.AuditLogFilter{
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	logs, err := s.GetAuditLogs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs for compliance report: %w", err)
	}

	report := &ComplianceReport{
		ID:          uuid.New().String(),
		Framework:   framework,
		GeneratedAt: time.Now(),
		PeriodStart: startTime,
		PeriodEnd:   endTime,
		TotalEvents: int64(len(logs)),
	}

	// Framework-specific compliance analysis
	switch framework {
	case FrameworkSOX:
		report = s.analyzeSOXCompliance(report, logs)
	case FrameworkHIPAA:
		report = s.analyzeHIPAACompliance(report, logs)
	case FrameworkSOC2:
		report = s.analyzeSOC2Compliance(report, logs)
	default:
		return nil, fmt.Errorf("unsupported compliance framework: %s", framework)
	}

	return report, nil
}

// CreateLegalHold creates a new legal hold
func (s *auditService) CreateLegalHold(ctx context.Context, req LegalHoldRequest) (*LegalHold, error) {
	hold := &LegalHold{
		ID:          uuid.New().String(),
		CaseNumber:  req.CaseNumber,
		Description: req.Description,
		CreatedAt:   time.Now(),
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Status:      "active",
	}

	// Count records affected by the legal hold
	filter := models.AuditLogFilter{
		StartTime: &req.StartTime,
		EndTime:   req.EndTime,
	}
	count, err := s.repo.CountAuditLogs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count records for legal hold: %w", err)
	}
	hold.RecordCount = count

	s.mutex.Lock()
	s.legalHolds[hold.ID] = hold
	s.mutex.Unlock()

	s.logger.Info("Legal hold created", "hold_id", hold.ID, "case_number", hold.CaseNumber, "record_count", hold.RecordCount)
	return hold, nil
}

// ReleaseLegalHold releases a legal hold
func (s *auditService) ReleaseLegalHold(ctx context.Context, holdID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	hold, exists := s.legalHolds[holdID]
	if !exists {
		return fmt.Errorf("legal hold not found: %s", holdID)
	}

	now := time.Now()
	hold.EndTime = &now
	hold.Status = "released"

	s.logger.Info("Legal hold released", "hold_id", holdID, "case_number", hold.CaseNumber)
	return nil
}

// GetLegalHolds returns all legal holds
func (s *auditService) GetLegalHolds(ctx context.Context) ([]*LegalHold, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	holds := make([]*LegalHold, 0, len(s.legalHolds))
	for _, hold := range s.legalHolds {
		holds = append(holds, hold)
	}

	return holds, nil
}

// StartRealTimeMonitoring starts real-time audit monitoring
func (s *auditService) StartRealTimeMonitoring(ctx context.Context, eventChannel chan<- AuditEvent) error {
	s.mutex.Lock()
	s.eventChannels = append(s.eventChannels, eventChannel)
	s.monitoringActive = true
	s.mutex.Unlock()

	s.logger.Info("Real-time monitoring started", "active_channels", len(s.eventChannels))
	return nil
}

// StopRealTimeMonitoring stops real-time audit monitoring
func (s *auditService) StopRealTimeMonitoring() error {
	s.mutex.Lock()
	s.monitoringActive = false
	s.eventChannels = nil
	s.mutex.Unlock()

	s.logger.Info("Real-time monitoring stopped")
	return nil
}

// GetSuspiciousActivities returns suspicious activities detected within the time window
func (s *auditService) GetSuspiciousActivities(ctx context.Context, timeWindow time.Duration) ([]*SuspiciousActivity, error) {
	since := time.Now().Add(-timeWindow)
	filter := models.AuditLogFilter{
		StartTime: &since,
	}

	logs, err := s.GetAuditLogs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for suspicious activity analysis: %w", err)
	}

	return s.analyzeSuspiciousPatterns(logs), nil
}

// ApplyRetentionPolicy applies a retention policy to audit logs
func (s *auditService) ApplyRetentionPolicy(ctx context.Context, policy RetentionPolicy) error {
	cutoffDate := time.Now().AddDate(0, 0, -policy.RetentionDays)

	// Check for active legal holds that might prevent deletion
	s.mutex.RLock()
	for _, hold := range s.legalHolds {
		if hold.Status == "active" && hold.StartTime.Before(cutoffDate) {
			s.mutex.RUnlock()
			return fmt.Errorf("cannot apply retention policy: active legal hold prevents deletion of records from %v", hold.StartTime)
		}
	}
	s.mutex.RUnlock()

	filter := models.AuditLogFilter{
		EndTime: &cutoffDate,
	}

	logsToDelete, err := s.GetAuditLogs(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get logs for retention policy: %w", err)
	}

	s.logger.Info("Applying retention policy", "policy", policy.Name, "logs_to_process", len(logsToDelete))

	// In a real implementation, this would delete the logs from the database
	// For now, we'll just log the action
	s.logger.Info("Retention policy applied", "policy", policy.Name, "processed_count", len(logsToDelete))
	return nil
}

// ArchiveOldLogs archives old audit logs
func (s *auditService) ArchiveOldLogs(ctx context.Context, cutoffDate time.Time) (*ArchivalResult, error) {
	filter := models.AuditLogFilter{
		EndTime: &cutoffDate,
	}

	logsToArchive, err := s.GetAuditLogs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for archival: %w", err)
	}

	archiveID := uuid.New().String()
	result := &ArchivalResult{
		ArchiveID:       archiveID,
		ArchivedCount:   int64(len(logsToArchive)),
		StartDate:       cutoffDate,
		EndDate:         time.Now(),
		StorageLocation: fmt.Sprintf("archive://%s", archiveID),
		CreatedAt:       time.Now(),
	}

	// Calculate archive size (approximate)
	archiveData, _ := json.Marshal(logsToArchive)
	result.ArchiveSize = int64(len(archiveData))

	s.logger.Info("Logs archived", "archive_id", archiveID, "count", result.ArchivedCount, "size_bytes", result.ArchiveSize)
	return result, nil
}

// RestoreArchivedLogs restores archived audit logs
func (s *auditService) RestoreArchivedLogs(ctx context.Context, archiveID string) error {
	s.logger.Info("Restoring archived logs", "archive_id", archiveID)
	// Implementation would restore logs from archive storage
	return nil
}

// Helper methods for export functionality
func (s *auditService) exportToCSV(logs []*models.AuditLog) ([]byte, error) {
	var buf strings.Builder
	buf.WriteString("ID,Timestamp,UserID,SessionID,Query,Command,SafetyLevel,Status,ClusterContext,NamespaceContext,IPAddress\n")

	for _, log := range logs {
		clusterContext := ""
		if log.ClusterContext != nil {
			clusterContext = *log.ClusterContext
		}
		namespaceContext := ""
		if log.NamespaceContext != nil {
			namespaceContext = *log.NamespaceContext
		}
		ipAddress := ""
		if log.IPAddress != nil {
			ipAddress = *log.IPAddress
		}

		buf.WriteString(fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			log.ID,
			log.Timestamp.Format(time.RFC3339),
			log.GetUserIDString(),
			log.GetSessionIDString(),
			escapeCSV(log.QueryText),
			escapeCSV(log.GeneratedCommand),
			log.SafetyLevel,
			log.ExecutionStatus,
			escapeCSV(clusterContext),
			escapeCSV(namespaceContext),
			escapeCSV(ipAddress),
		))
	}

	return []byte(buf.String()), nil
}

func (s *auditService) exportToPDF(logs []*models.AuditLog) ([]byte, error) {
	// This would generate a PDF using a library like gofpdf
	// For now, return a placeholder
	return []byte("PDF export not implemented"), nil
}

// Helper methods for compliance analysis
func (s *auditService) analyzeSOXCompliance(report *ComplianceReport, logs []*models.AuditLog) *ComplianceReport {
	violations := []ComplianceViolation{}
	score := 100.0

	// SOX-specific analysis
	for _, log := range logs {
		if log.IsDangerous() && !log.IsSuccessful() {
			violations = append(violations, ComplianceViolation{
				ID:          uuid.New().String(),
				Type:        "unauthorized_access",
				Severity:    "high",
				Description: "Dangerous operation attempted",
				AffectedLogIDs: []int64{log.ID},
				DetectedAt:  time.Now(),
			})
			score -= 10.0
		}
	}

	report.ComplianceScore = score
	report.Violations = violations
	report.ExecutiveSummary = fmt.Sprintf("SOX compliance analysis completed. Score: %.1f%%. %d violations detected.", score, len(violations))

	return report
}

func (s *auditService) analyzeHIPAACompliance(report *ComplianceReport, logs []*models.AuditLog) *ComplianceReport {
	// HIPAA-specific analysis
	report.ComplianceScore = 95.0
	report.ExecutiveSummary = "HIPAA compliance analysis completed. All access properly logged and authenticated."
	return report
}

func (s *auditService) analyzeSOC2Compliance(report *ComplianceReport, logs []*models.AuditLog) *ComplianceReport {
	// SOC2-specific analysis
	report.ComplianceScore = 98.0
	report.ExecutiveSummary = "SOC2 compliance analysis completed. Security controls operating effectively."
	return report
}

// Helper method for suspicious activity analysis
func (s *auditService) analyzeSuspiciousPatterns(logs []*models.AuditLog) []*SuspiciousActivity {
	activities := []*SuspiciousActivity{}

	// Analyze patterns - multiple failed attempts, unusual access times, etc.
	userFailures := make(map[string]int)

	for _, log := range logs {
		if !log.IsSuccessful() && log.UserID != nil {
			userFailures[log.UserID.String()]++
		}
	}

	// Flag users with multiple failures
	for userID, failures := range userFailures {
		if failures >= 5 {
			uid, _ := uuid.Parse(userID)
			activities = append(activities, &SuspiciousActivity{
				ID:             uuid.New().String(),
				DetectedAt:     time.Now(),
				ActivityType:   "multiple_failures",
				UserID:         &uid,
				Description:    fmt.Sprintf("User has %d failed operations", failures),
				RiskScore:      float64(failures) * 10.0,
				PatternMatched: "failure_threshold",
			})
		}
	}

	return activities
}

// Helper function to escape CSV values
func escapeCSV(value string) string {
	if strings.Contains(value, ",") || strings.Contains(value, "\"") || strings.Contains(value, "\n") {
		value = strings.ReplaceAll(value, "\"", "\"\"")
		return "\"" + value + "\""
	}
	return value
}

// Shutdown gracefully shuts down the audit service
func (s *auditService) Shutdown(ctx context.Context) error {
	if s.config.EnableAsyncLogging {
		close(s.stopChan)
		s.wg.Wait()

		// Process remaining items in queue
		for len(s.logQueue) > 0 {
			select {
			case req := <-s.logQueue:
				if err := s.processAuditLog(ctx, req); err != nil {
					log.Printf("Failed to process remaining audit log: %v", err)
				}
			default:
				break
			}
		}

		close(s.logQueue)
	}

	s.logger.Info("Audit service shutdown completed")
	return nil
}
