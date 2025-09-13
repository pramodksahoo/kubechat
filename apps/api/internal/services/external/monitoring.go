package external

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/audit"
)

// MonitoringService handles monitoring and logging for external API calls
type MonitoringService interface {
	// LogRequest logs an outgoing API request
	LogRequest(ctx context.Context, req *APIRequestLog) error

	// LogResponse logs an API response
	LogResponse(ctx context.Context, resp *APIResponseLog) error

	// LogError logs an API error
	LogError(ctx context.Context, err *APIErrorLog) error

	// GetMetrics returns current metrics
	GetMetrics(ctx context.Context) (*APIMetrics, error)

	// GetHealthStatus returns health status of all monitored APIs
	GetHealthStatus(ctx context.Context) (*HealthStatus, error)

	// CreateAlert creates an alert for API issues
	CreateAlert(ctx context.Context, alert *Alert) error

	// GetRequestLogs returns request logs with filtering
	GetRequestLogs(ctx context.Context, filter *LogFilter) ([]*APIRequestLog, error)

	// GetPerformanceStats returns performance statistics
	GetPerformanceStats(ctx context.Context, provider string, timeRange time.Duration) (*PerformanceStats, error)
}

// APIRequestLog represents a logged API request
type APIRequestLog struct {
	ID          uuid.UUID              `json:"id"`
	RequestID   string                 `json:"request_id"`
	Provider    string                 `json:"provider"`
	Endpoint    string                 `json:"endpoint"`
	Method      string                 `json:"method"`
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty"`
	RequestSize int64                  `json:"request_size"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// APIResponseLog represents a logged API response
type APIResponseLog struct {
	ID           uuid.UUID         `json:"id"`
	RequestID    string            `json:"request_id"`
	Provider     string            `json:"provider"`
	StatusCode   int               `json:"status_code"`
	ResponseSize int64             `json:"response_size"`
	Duration     time.Duration     `json:"duration"`
	TokensUsed   int               `json:"tokens_used"`
	Cost         float64           `json:"cost"`
	Success      bool              `json:"success"`
	Headers      map[string]string `json:"headers,omitempty"`
	Timestamp    time.Time         `json:"timestamp"`
	CacheHit     bool              `json:"cache_hit"`
	RetryCount   int               `json:"retry_count"`
}

// APIErrorLog represents a logged API error
type APIErrorLog struct {
	ID         uuid.UUID              `json:"id"`
	RequestID  string                 `json:"request_id"`
	Provider   string                 `json:"provider"`
	Endpoint   string                 `json:"endpoint"`
	ErrorType  string                 `json:"error_type"`
	ErrorCode  string                 `json:"error_code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"status_code"`
	Timestamp  time.Time              `json:"timestamp"`
	Retryable  bool                   `json:"retryable"`
	RetryCount int                    `json:"retry_count"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// APIMetrics represents aggregated API metrics
type APIMetrics struct {
	Providers      map[string]*ProviderMetrics `json:"providers"`
	TotalRequests  int64                       `json:"total_requests"`
	TotalErrors    int64                       `json:"total_errors"`
	TotalCost      float64                     `json:"total_cost"`
	TotalTokens    int64                       `json:"total_tokens"`
	AverageLatency float64                     `json:"average_latency_ms"`
	ErrorRate      float64                     `json:"error_rate"`
	LastUpdated    time.Time                   `json:"last_updated"`
	TimeWindow     time.Duration               `json:"time_window"`
}

// ProviderMetrics represents metrics for a specific provider
type ProviderMetrics struct {
	Provider       string                    `json:"provider"`
	RequestCount   int64                     `json:"request_count"`
	ErrorCount     int64                     `json:"error_count"`
	SuccessRate    float64                   `json:"success_rate"`
	AverageLatency float64                   `json:"average_latency_ms"`
	TokensUsed     int64                     `json:"tokens_used"`
	TotalCost      float64                   `json:"total_cost"`
	LastRequest    time.Time                 `json:"last_request"`
	Endpoints      map[string]*EndpointStats `json:"endpoints"`
	CircuitState   string                    `json:"circuit_state"`
	HealthStatus   string                    `json:"health_status"`
}

// EndpointStats represents statistics for a specific endpoint
type EndpointStats struct {
	Endpoint       string    `json:"endpoint"`
	RequestCount   int64     `json:"request_count"`
	ErrorCount     int64     `json:"error_count"`
	AverageLatency float64   `json:"average_latency_ms"`
	LastAccessed   time.Time `json:"last_accessed"`
}

// HealthStatus represents overall health status
type HealthStatus struct {
	Overall     string                     `json:"overall"`
	Providers   map[string]*ProviderHealth `json:"providers"`
	LastChecked time.Time                  `json:"last_checked"`
	Alerts      []*Alert                   `json:"active_alerts"`
}

// ProviderHealth represents health status for a provider
type ProviderHealth struct {
	Provider     string    `json:"provider"`
	Status       string    `json:"status"` // healthy, degraded, down
	LastChecked  time.Time `json:"last_checked"`
	Uptime       float64   `json:"uptime_percentage"`
	Issues       []string  `json:"issues,omitempty"`
	ResponseTime float64   `json:"avg_response_time_ms"`
}

// Alert represents an API monitoring alert
type Alert struct {
	ID         uuid.UUID              `json:"id"`
	Type       string                 `json:"type"`
	Severity   string                 `json:"severity"`
	Provider   string                 `json:"provider"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details"`
	CreatedAt  time.Time              `json:"created_at"`
	ResolvedAt *time.Time             `json:"resolved_at,omitempty"`
	Status     string                 `json:"status"`
}

// LogFilter represents filtering criteria for log queries
type LogFilter struct {
	Provider    string        `json:"provider,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
	SessionID   string        `json:"session_id,omitempty"`
	StartTime   time.Time     `json:"start_time,omitempty"`
	EndTime     time.Time     `json:"end_time,omitempty"`
	Success     *bool         `json:"success,omitempty"`
	MinDuration time.Duration `json:"min_duration,omitempty"`
	MaxDuration time.Duration `json:"max_duration,omitempty"`
	Limit       int           `json:"limit,omitempty"`
	Offset      int           `json:"offset,omitempty"`
}

// PerformanceStats represents performance statistics over time
type PerformanceStats struct {
	Provider       string             `json:"provider"`
	TimeRange      time.Duration      `json:"time_range"`
	TotalRequests  int64              `json:"total_requests"`
	SuccessRate    float64            `json:"success_rate"`
	Percentiles    map[string]float64 `json:"latency_percentiles"`
	ThroughputRPS  float64            `json:"throughput_rps"`
	ErrorBreakdown map[string]int64   `json:"error_breakdown"`
	CostAnalysis   *CostAnalysis      `json:"cost_analysis"`
	Trends         *PerformanceTrends `json:"trends"`
}

// CostAnalysis represents cost breakdown and analysis
type CostAnalysis struct {
	TotalCost    float64            `json:"total_cost"`
	CostPerToken float64            `json:"cost_per_token"`
	CostByModel  map[string]float64 `json:"cost_by_model"`
	DailyCosts   []DailyCost        `json:"daily_costs"`
	Projections  *CostProjections   `json:"projections"`
}

// DailyCostOld represents cost for a specific day (use DailyCost from cost_service.go instead)
// This struct is deprecated - using DailyCost from cost_service.go
type DailyCostOld struct {
	Date  time.Time `json:"date"`
	Cost  float64   `json:"cost"`
	Usage int64     `json:"token_usage"`
}

// CostProjections represents projected costs
type CostProjections struct {
	Monthly    float64 `json:"monthly_projection"`
	Yearly     float64 `json:"yearly_projection"`
	Trend      string  `json:"trend"` // increasing, decreasing, stable
	Confidence float64 `json:"confidence_percentage"`
}

// PerformanceTrends represents performance trends over time
type PerformanceTrends struct {
	LatencyTrend     string  `json:"latency_trend"`
	ErrorRateTrend   string  `json:"error_rate_trend"`
	ThroughputTrend  string  `json:"throughput_trend"`
	ReliabilityScore float64 `json:"reliability_score"`
}

// monitoringServiceImpl implements MonitoringService
type monitoringServiceImpl struct {
	auditSvc     audit.Service
	requestLogs  []*APIRequestLog
	responseLogs []*APIResponseLog
	errorLogs    []*APIErrorLog
	metrics      *APIMetrics
	alerts       []*Alert
	healthCache  *HealthStatus
	mu           sync.RWMutex
	config       *MonitoringConfig
}

// MonitoringConfig represents monitoring service configuration
type MonitoringConfig struct {
	RetentionDays     int                        `json:"retention_days"`
	MetricsInterval   time.Duration              `json:"metrics_interval"`
	HealthInterval    time.Duration              `json:"health_interval"`
	MaxLogEntries     int                        `json:"max_log_entries"`
	EnableDetailedLog bool                       `json:"enable_detailed_log"`
	AlertThresholds   *MonitoringAlertThresholds `json:"alert_thresholds"`
}

// MonitoringAlertThresholds defines monitoring-specific alert thresholds
type MonitoringAlertThresholds struct {
	ErrorRatePercent   float64       `json:"error_rate_percent"`
	LatencyThresholdMs float64       `json:"latency_threshold_ms"`
	CostThresholdDaily float64       `json:"cost_threshold_daily"`
	HealthCheckTimeout time.Duration `json:"health_check_timeout"`
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(auditSvc audit.Service, config *MonitoringConfig) MonitoringService {
	if config == nil {
		config = &MonitoringConfig{
			RetentionDays:     30,
			MetricsInterval:   1 * time.Minute,
			HealthInterval:    5 * time.Minute,
			MaxLogEntries:     10000,
			EnableDetailedLog: true,
			AlertThresholds: &MonitoringAlertThresholds{
				ErrorRatePercent:   10.0,
				LatencyThresholdMs: 5000.0,
				CostThresholdDaily: 100.0,
				HealthCheckTimeout: 30 * time.Second,
			},
		}
	}

	service := &monitoringServiceImpl{
		auditSvc:     auditSvc,
		requestLogs:  make([]*APIRequestLog, 0),
		responseLogs: make([]*APIResponseLog, 0),
		errorLogs:    make([]*APIErrorLog, 0),
		alerts:       make([]*Alert, 0),
		config:       config,
		metrics: &APIMetrics{
			Providers: make(map[string]*ProviderMetrics),
		},
		healthCache: &HealthStatus{
			Providers: make(map[string]*ProviderHealth),
		},
	}

	// Start background tasks
	go service.runMetricsCollection()
	go service.runHealthChecks()
	go service.runLogCleanup()

	log.Println("External API monitoring service initialized")
	return service
}

// LogRequest logs an outgoing API request
func (s *monitoringServiceImpl) LogRequest(ctx context.Context, req *APIRequestLog) error {
	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	s.mu.Lock()
	s.requestLogs = append(s.requestLogs, req)

	// Maintain log size limit
	if len(s.requestLogs) > s.config.MaxLogEntries {
		s.requestLogs = s.requestLogs[1:]
	}
	s.mu.Unlock()

	// Log to audit service if detailed logging is enabled
	if s.config.EnableDetailedLog && s.auditSvc != nil {
		uid := uuid.MustParse(req.UserID)
		auditReq := models.AuditLogRequest{
			UserID:           &uid,
			QueryText:        fmt.Sprintf("External API request: %s %s", req.Method, req.Endpoint),
			GeneratedCommand: fmt.Sprintf("Request to %s: %s %s", req.Provider, req.Method, req.Endpoint),
			SafetyLevel:      models.SafetyLevelSafe,
			ExecutionResult: map[string]interface{}{
				"request_id":   req.RequestID,
				"provider":     req.Provider,
				"endpoint":     req.Endpoint,
				"method":       req.Method,
				"request_size": req.RequestSize,
			},
			ExecutionStatus: models.ExecutionStatusSuccess,
		}

		if err := s.auditSvc.LogUserAction(ctx, auditReq); err != nil {
			log.Printf("Failed to log request audit: %v", err)
		}
	}

	return nil
}

// LogResponse logs an API response
func (s *monitoringServiceImpl) LogResponse(ctx context.Context, resp *APIResponseLog) error {
	if resp.ID == uuid.Nil {
		resp.ID = uuid.New()
	}
	if resp.Timestamp.IsZero() {
		resp.Timestamp = time.Now()
	}

	s.mu.Lock()
	s.responseLogs = append(s.responseLogs, resp)

	// Maintain log size limit
	if len(s.responseLogs) > s.config.MaxLogEntries {
		s.responseLogs = s.responseLogs[1:]
	}

	// Update metrics
	s.updateMetrics(resp)
	s.mu.Unlock()

	// Check for alerts
	go s.checkAlerts(resp)

	// Log to audit service
	if s.auditSvc != nil {
		safetyLevel := models.SafetyLevelSafe
		if !resp.Success {
			safetyLevel = models.SafetyLevelWarning
		}

		auditReq := models.AuditLogRequest{
			QueryText:        fmt.Sprintf("External API response from %s", resp.Provider),
			GeneratedCommand: fmt.Sprintf("Response from %s with status %d", resp.Provider, resp.StatusCode),
			SafetyLevel:      string(safetyLevel),
			ExecutionResult: map[string]interface{}{
				"request_id":    resp.RequestID,
				"provider":      resp.Provider,
				"status_code":   resp.StatusCode,
				"duration_ms":   resp.Duration.Milliseconds(),
				"response_size": resp.ResponseSize,
				"tokens_used":   resp.TokensUsed,
				"cost":          resp.Cost,
				"cache_hit":     resp.CacheHit,
				"retry_count":   resp.RetryCount,
			},
			ExecutionStatus: func() string {
				if resp.Success {
					return models.ExecutionStatusSuccess
				}
				return models.ExecutionStatusFailed
			}(),
		}

		if err := s.auditSvc.LogUserAction(ctx, auditReq); err != nil {
			log.Printf("Failed to log response audit: %v", err)
		}
	}

	return nil
}

// LogError logs an API error
func (s *monitoringServiceImpl) LogError(ctx context.Context, err *APIErrorLog) error {
	if err.ID == uuid.Nil {
		err.ID = uuid.New()
	}
	if err.Timestamp.IsZero() {
		err.Timestamp = time.Now()
	}

	s.mu.Lock()
	s.errorLogs = append(s.errorLogs, err)

	// Maintain log size limit
	if len(s.errorLogs) > s.config.MaxLogEntries {
		s.errorLogs = s.errorLogs[1:]
	}
	s.mu.Unlock()

	// Create alert for errors
	go s.createErrorAlert(err)

	// Log to audit service
	if s.auditSvc != nil {
		auditReq := models.AuditLogRequest{
			QueryText:        fmt.Sprintf("External API error from %s", err.Provider),
			GeneratedCommand: fmt.Sprintf("Error in %s at %s: %s", err.Provider, err.Endpoint, err.Message),
			SafetyLevel:      models.SafetyLevelDangerous,
			ExecutionResult: map[string]interface{}{
				"request_id":  err.RequestID,
				"provider":    err.Provider,
				"endpoint":    err.Endpoint,
				"error_type":  err.ErrorType,
				"error_code":  err.ErrorCode,
				"message":     err.Message,
				"status_code": err.StatusCode,
				"retryable":   err.Retryable,
				"retry_count": err.RetryCount,
			},
			ExecutionStatus: models.ExecutionStatusFailed,
		}

		if logErr := s.auditSvc.LogUserAction(ctx, auditReq); logErr != nil {
			log.Printf("Failed to log error audit: %v", logErr)
		}
	}

	return nil
}

// GetMetrics returns current metrics
func (s *monitoringServiceImpl) GetMetrics(ctx context.Context) (*APIMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a deep copy of metrics
	metricsJSON, err := json.Marshal(s.metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize metrics: %w", err)
	}

	var metrics APIMetrics
	if err := json.Unmarshal(metricsJSON, &metrics); err != nil {
		return nil, fmt.Errorf("failed to deserialize metrics: %w", err)
	}

	return &metrics, nil
}

// GetHealthStatus returns health status of all monitored APIs
func (s *monitoringServiceImpl) GetHealthStatus(ctx context.Context) (*HealthStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a deep copy of health status
	healthJSON, err := json.Marshal(s.healthCache)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize health status: %w", err)
	}

	var health HealthStatus
	if err := json.Unmarshal(healthJSON, &health); err != nil {
		return nil, fmt.Errorf("failed to deserialize health status: %w", err)
	}

	return &health, nil
}

// CreateAlert creates an alert for API issues
func (s *monitoringServiceImpl) CreateAlert(ctx context.Context, alert *Alert) error {
	if alert.ID == uuid.Nil {
		alert.ID = uuid.New()
	}
	if alert.CreatedAt.IsZero() {
		alert.CreatedAt = time.Now()
	}
	if alert.Status == "" {
		alert.Status = "active"
	}

	s.mu.Lock()
	s.alerts = append(s.alerts, alert)
	s.mu.Unlock()

	log.Printf("Alert created: %s - %s", alert.Type, alert.Message)
	return nil
}

// GetRequestLogs returns request logs with filtering
func (s *monitoringServiceImpl) GetRequestLogs(ctx context.Context, filter *LogFilter) ([]*APIRequestLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*APIRequestLog

	for _, logEntry := range s.requestLogs {
		if s.matchesFilter(logEntry, filter) {
			filtered = append(filtered, logEntry)
		}
	}

	// Apply limit and offset
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(filtered) {
			filtered = filtered[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(filtered) {
			filtered = filtered[:filter.Limit]
		}
	}

	return filtered, nil
}

// GetPerformanceStats returns performance statistics
func (s *monitoringServiceImpl) GetPerformanceStats(ctx context.Context, provider string, timeRange time.Duration) (*PerformanceStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cutoff := time.Now().Add(-timeRange)
	var relevantResponses []*APIResponseLog

	for _, resp := range s.responseLogs {
		if resp.Provider == provider && resp.Timestamp.After(cutoff) {
			relevantResponses = append(relevantResponses, resp)
		}
	}

	if len(relevantResponses) == 0 {
		return nil, fmt.Errorf("no data available for provider %s in time range", provider)
	}

	stats := &PerformanceStats{
		Provider:       provider,
		TimeRange:      timeRange,
		TotalRequests:  int64(len(relevantResponses)),
		ErrorBreakdown: make(map[string]int64),
		Percentiles:    make(map[string]float64),
		CostAnalysis: &CostAnalysis{
			CostByModel: make(map[string]float64),
			DailyCosts:  make([]DailyCost, 0),
		},
	}

	// Calculate statistics from relevant responses
	s.calculatePerformanceStats(stats, relevantResponses)

	return stats, nil
}

// Helper methods

func (s *monitoringServiceImpl) updateMetrics(resp *APIResponseLog) {
	if s.metrics.Providers == nil {
		s.metrics.Providers = make(map[string]*ProviderMetrics)
	}

	provider := resp.Provider
	if _, exists := s.metrics.Providers[provider]; !exists {
		s.metrics.Providers[provider] = &ProviderMetrics{
			Provider:  provider,
			Endpoints: make(map[string]*EndpointStats),
		}
	}

	pm := s.metrics.Providers[provider]
	pm.RequestCount++
	pm.LastRequest = resp.Timestamp
	pm.TokensUsed += int64(resp.TokensUsed)
	pm.TotalCost += resp.Cost

	if !resp.Success {
		pm.ErrorCount++
		s.metrics.TotalErrors++
	}

	// Update average latency
	totalRequests := float64(pm.RequestCount)
	pm.AverageLatency = (pm.AverageLatency*(totalRequests-1) +
		float64(resp.Duration.Milliseconds())) / totalRequests

	pm.SuccessRate = float64(pm.RequestCount-pm.ErrorCount) / float64(pm.RequestCount) * 100

	// Update overall metrics
	s.metrics.TotalRequests++
	s.metrics.TotalCost += resp.Cost
	s.metrics.TotalTokens += int64(resp.TokensUsed)

	totalReqs := float64(s.metrics.TotalRequests)
	s.metrics.AverageLatency = (s.metrics.AverageLatency*(totalReqs-1) +
		float64(resp.Duration.Milliseconds())) / totalReqs

	s.metrics.ErrorRate = float64(s.metrics.TotalErrors) / float64(s.metrics.TotalRequests) * 100
	s.metrics.LastUpdated = time.Now()
}

func (s *monitoringServiceImpl) checkAlerts(resp *APIResponseLog) {
	// Check error rate threshold
	if s.config.AlertThresholds != nil {
		provider := resp.Provider
		if pm, exists := s.metrics.Providers[provider]; exists {
			if pm.SuccessRate < (100 - s.config.AlertThresholds.ErrorRatePercent) {
				alert := &Alert{
					Type:     "high_error_rate",
					Severity: "warning",
					Provider: provider,
					Message:  fmt.Sprintf("High error rate detected: %.2f%%", 100-pm.SuccessRate),
					Details: map[string]interface{}{
						"error_rate":    100 - pm.SuccessRate,
						"threshold":     s.config.AlertThresholds.ErrorRatePercent,
						"request_count": pm.RequestCount,
					},
				}
				s.CreateAlert(context.Background(), alert)
			}
		}

		// Check latency threshold
		if resp.Duration.Milliseconds() > int64(s.config.AlertThresholds.LatencyThresholdMs) {
			alert := &Alert{
				Type:     "high_latency",
				Severity: "warning",
				Provider: resp.Provider,
				Message:  fmt.Sprintf("High latency detected: %dms", resp.Duration.Milliseconds()),
				Details: map[string]interface{}{
					"latency_ms": resp.Duration.Milliseconds(),
					"threshold":  s.config.AlertThresholds.LatencyThresholdMs,
					"request_id": resp.RequestID,
				},
			}
			s.CreateAlert(context.Background(), alert)
		}
	}
}

func (s *monitoringServiceImpl) createErrorAlert(err *APIErrorLog) {
	alert := &Alert{
		Type:     "api_error",
		Severity: "error",
		Provider: err.Provider,
		Message:  fmt.Sprintf("API error: %s", err.Message),
		Details: map[string]interface{}{
			"error_type":  err.ErrorType,
			"error_code":  err.ErrorCode,
			"endpoint":    err.Endpoint,
			"status_code": err.StatusCode,
			"retryable":   err.Retryable,
			"retry_count": err.RetryCount,
		},
	}
	s.CreateAlert(context.Background(), alert)
}

func (s *monitoringServiceImpl) matchesFilter(log *APIRequestLog, filter *LogFilter) bool {
	if filter == nil {
		return true
	}

	if filter.Provider != "" && log.Provider != filter.Provider {
		return false
	}

	if filter.UserID != "" && log.UserID != filter.UserID {
		return false
	}

	if filter.SessionID != "" && log.SessionID != filter.SessionID {
		return false
	}

	if !filter.StartTime.IsZero() && log.Timestamp.Before(filter.StartTime) {
		return false
	}

	if !filter.EndTime.IsZero() && log.Timestamp.After(filter.EndTime) {
		return false
	}

	return true
}

func (s *monitoringServiceImpl) calculatePerformanceStats(stats *PerformanceStats, responses []*APIResponseLog) {
	if len(responses) == 0 {
		return
	}

	var totalDuration time.Duration
	var successCount int64
	var latencies []float64

	for _, resp := range responses {
		totalDuration += resp.Duration
		latencies = append(latencies, float64(resp.Duration.Milliseconds()))
		stats.CostAnalysis.TotalCost += resp.Cost

		if resp.Success {
			successCount++
		}
	}

	stats.SuccessRate = float64(successCount) / float64(len(responses)) * 100
	stats.ThroughputRPS = float64(len(responses)) / stats.TimeRange.Seconds()

	// Calculate percentiles (simplified implementation)
	if len(latencies) > 0 {
		// Sort latencies for percentile calculation
		// This is a simplified implementation
		stats.Percentiles["p50"] = latencies[len(latencies)/2]
		stats.Percentiles["p95"] = latencies[int(float64(len(latencies))*0.95)]
		stats.Percentiles["p99"] = latencies[int(float64(len(latencies))*0.99)]
	}
}

func (s *monitoringServiceImpl) runMetricsCollection() {
	ticker := time.NewTicker(s.config.MetricsInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		s.metrics.LastUpdated = time.Now()
		s.metrics.TimeWindow = s.config.MetricsInterval
		s.mu.Unlock()
	}
}

func (s *monitoringServiceImpl) runHealthChecks() {
	ticker := time.NewTicker(s.config.HealthInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		s.healthCache.LastChecked = time.Now()
		// Update health status based on recent metrics
		s.updateHealthStatus()
		s.mu.Unlock()
	}
}

func (s *monitoringServiceImpl) updateHealthStatus() {
	overall := "healthy"

	for provider, metrics := range s.metrics.Providers {
		health := &ProviderHealth{
			Provider:     provider,
			Status:       "healthy",
			LastChecked:  time.Now(),
			Uptime:       metrics.SuccessRate,
			ResponseTime: metrics.AverageLatency,
		}

		if metrics.SuccessRate < 95 {
			health.Status = "degraded"
			health.Issues = append(health.Issues, "High error rate")
			overall = "degraded"
		}

		if metrics.AverageLatency > 5000 {
			health.Status = "degraded"
			health.Issues = append(health.Issues, "High latency")
			if overall == "healthy" {
				overall = "degraded"
			}
		}

		if time.Since(metrics.LastRequest) > 1*time.Hour {
			health.Status = "down"
			health.Issues = append(health.Issues, "No recent requests")
			overall = "down"
		}

		s.healthCache.Providers[provider] = health
	}

	s.healthCache.Overall = overall
}

func (s *monitoringServiceImpl) runLogCleanup() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().AddDate(0, 0, -s.config.RetentionDays)

		s.mu.Lock()
		// Clean up old logs
		s.cleanupOldLogs(cutoff)
		s.mu.Unlock()

		log.Printf("Cleaned up logs older than %d days", s.config.RetentionDays)
	}
}

func (s *monitoringServiceImpl) cleanupOldLogs(cutoff time.Time) {
	// Clean request logs
	var newRequestLogs []*APIRequestLog
	for _, log := range s.requestLogs {
		if log.Timestamp.After(cutoff) {
			newRequestLogs = append(newRequestLogs, log)
		}
	}
	s.requestLogs = newRequestLogs

	// Clean response logs
	var newResponseLogs []*APIResponseLog
	for _, log := range s.responseLogs {
		if log.Timestamp.After(cutoff) {
			newResponseLogs = append(newResponseLogs, log)
		}
	}
	s.responseLogs = newResponseLogs

	// Clean error logs
	var newErrorLogs []*APIErrorLog
	for _, log := range s.errorLogs {
		if log.Timestamp.After(cutoff) {
			newErrorLogs = append(newErrorLogs, log)
		}
	}
	s.errorLogs = newErrorLogs
}
