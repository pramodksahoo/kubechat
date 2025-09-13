package audit

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/repositories"
)

// Service defines the interface for audit logging service
type Service interface {
	// LogUserAction logs a user action with structured data
	LogUserAction(ctx context.Context, req models.AuditLogRequest) error

	// LogKubectlExecution logs kubectl command execution
	LogKubectlExecution(ctx context.Context, userID *uuid.UUID, sessionID *uuid.UUID, query, command string, result map[string]interface{}, status string, req *http.Request) error

	// LogAPICall logs API call with request context
	LogAPICall(ctx context.Context, userID *uuid.UUID, sessionID *uuid.UUID, method, endpoint, query string, statusCode int, duration time.Duration, req *http.Request) error

	// LogSecurityEvent logs security-related events
	LogSecurityEvent(ctx context.Context, eventType, description string, userID *uuid.UUID, severity string, req *http.Request) error

	// GetAuditLogs retrieves audit logs with filtering
	GetAuditLogs(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error)

	// GetAuditLogByID retrieves a specific audit log
	GetAuditLogByID(ctx context.Context, id int64) (*models.AuditLog, error)

	// GetAuditLogSummary returns summary statistics
	GetAuditLogSummary(ctx context.Context, filter models.AuditLogFilter) (*models.AuditLogSummary, error)

	// GetDangerousOperations retrieves dangerous operations
	GetDangerousOperations(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error)

	// GetFailedOperations retrieves failed operations
	GetFailedOperations(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error)

	// VerifyIntegrity verifies audit log integrity
	VerifyIntegrity(ctx context.Context, startID, endID int64) ([]models.IntegrityCheckResult, error)

	// GetMetrics returns audit service metrics
	GetMetrics(ctx context.Context) (*AuditMetrics, error)

	// HealthCheck performs health check on audit service
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
		repo:    repo,
		config:  config,
		logger:  logger,
		metrics: &AuditMetrics{},
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
