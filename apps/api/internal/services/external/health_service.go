package external

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/audit"
)

// HealthService provides health checking for external APIs
type HealthService interface {
	// CheckHealth performs health check on a specific API
	CheckHealth(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error)

	// CheckAllAPIs performs health checks on all registered APIs
	CheckAllAPIs(ctx context.Context) (*AllAPIsHealthResponse, error)

	// RegisterAPI registers an API for health monitoring
	RegisterAPI(ctx context.Context, config *APIHealthConfig) error

	// UpdateAPIConfig updates health check configuration for an API
	UpdateAPIConfig(ctx context.Context, config *APIHealthConfig) error

	// RemoveAPI removes an API from health monitoring
	RemoveAPI(ctx context.Context, apiName string) error

	// GetAPIHealthHistory returns health history for an API
	GetAPIHealthHistory(ctx context.Context, apiName string, timeframe *TimeFrame) (*HealthHistoryResponse, error)

	// GetHealthSummary returns overall health summary
	GetHealthSummary(ctx context.Context) (*HealthSummary, error)

	// StartContinuousMonitoring starts background health monitoring
	StartContinuousMonitoring(ctx context.Context) error

	// StopContinuousMonitoring stops background health monitoring
	StopContinuousMonitoring() error

	// GetHealthMetrics returns health monitoring metrics
	GetHealthMetrics(ctx context.Context) (*HealthMetrics, error)
}

// HealthCheckRequest contains parameters for health check
type HealthCheckRequest struct {
	APIName         string            `json:"api_name"`
	Endpoint        string            `json:"endpoint,omitempty"`
	Method          string            `json:"method,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	Timeout         time.Duration     `json:"timeout,omitempty"`
	ExpectedStatus  []int             `json:"expected_status,omitempty"`
	ValidateContent bool              `json:"validate_content,omitempty"`
	ContentChecks   []ContentCheck    `json:"content_checks,omitempty"`
}

// HealthCheckResponse contains health check results
type HealthCheckResponse struct {
	CheckID        uuid.UUID              `json:"check_id"`
	APIName        string                 `json:"api_name"`
	Endpoint       string                 `json:"endpoint"`
	Status         HealthServiceStatus    `json:"status"`
	ResponseTime   time.Duration          `json:"response_time"`
	StatusCode     int                    `json:"status_code,omitempty"`
	ResponseSize   int64                  `json:"response_size"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	CheckedAt      time.Time              `json:"checked_at"`
	ContentResults []ContentResult        `json:"content_results,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Warnings       []string               `json:"warnings,omitempty"`
}

// AllAPIsHealthResponse contains health status for all APIs
type AllAPIsHealthResponse struct {
	OverallStatus HealthServiceStatus    `json:"overall_status"`
	TotalAPIs     int                    `json:"total_apis"`
	HealthyAPIs   int                    `json:"healthy_apis"`
	UnhealthyAPIs int                    `json:"unhealthy_apis"`
	DegradedAPIs  int                    `json:"degraded_apis"`
	APIResults    []*HealthCheckResponse `json:"api_results"`
	Summary       *HealthSummary         `json:"summary"`
	CheckedAt     time.Time              `json:"checked_at"`
}

// APIHealthConfig contains health monitoring configuration for an API
type APIHealthConfig struct {
	Name                string            `json:"name"`
	DisplayName         string            `json:"display_name"`
	Description         string            `json:"description"`
	Endpoint            string            `json:"endpoint"`
	Method              string            `json:"method"`
	Headers             map[string]string `json:"headers,omitempty"`
	Timeout             time.Duration     `json:"timeout"`
	Interval            time.Duration     `json:"interval"`
	RetryAttempts       int               `json:"retry_attempts"`
	RetryInterval       time.Duration     `json:"retry_interval"`
	ExpectedStatusCodes []int             `json:"expected_status_codes"`
	ContentChecks       []ContentCheck    `json:"content_checks,omitempty"`
	AlertThresholds     *AlertThresholds  `json:"alert_thresholds,omitempty"`
	Enabled             bool              `json:"enabled"`
	Tags                []string          `json:"tags,omitempty"`
	Metadata            map[string]string `json:"metadata,omitempty"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

// ContentCheck defines content validation rules
type ContentCheck struct {
	Type        ContentCheckType `json:"type"`
	Pattern     string           `json:"pattern,omitempty"`
	JsonPath    string           `json:"json_path,omitempty"`
	Expected    interface{}      `json:"expected,omitempty"`
	Required    bool             `json:"required"`
	Description string           `json:"description,omitempty"`
}

// ContentResult contains content validation results
type ContentResult struct {
	CheckType    ContentCheckType `json:"check_type"`
	Pattern      string           `json:"pattern,omitempty"`
	JsonPath     string           `json:"json_path,omitempty"`
	Expected     interface{}      `json:"expected,omitempty"`
	Actual       interface{}      `json:"actual,omitempty"`
	Passed       bool             `json:"passed"`
	ErrorMessage string           `json:"error_message,omitempty"`
}

// HealthHistoryResponse contains health history data
type HealthHistoryResponse struct {
	APIName     string                 `json:"api_name"`
	TimeFrame   *TimeFrame             `json:"time_frame"`
	TotalChecks int64                  `json:"total_checks"`
	History     []*HealthCheckResponse `json:"history"`
	Statistics  *HealthStatistics      `json:"statistics"`
	Trends      *HealthTrends          `json:"trends"`
}

// HealthSummary contains overall health summary
type HealthSummary struct {
	OverallStatus       HealthServiceStatus         `json:"overall_status"`
	TotalAPIs           int                         `json:"total_apis"`
	APIsByStatus        map[HealthServiceStatus]int `json:"apis_by_status"`
	AverageResponseTime time.Duration               `json:"average_response_time"`
	UptimePercentage    float64                     `json:"uptime_percentage"`
	TotalChecks         int64                       `json:"total_checks"`
	TotalFailures       int64                       `json:"total_failures"`
	LastUpdated         time.Time                   `json:"last_updated"`
	CriticalIssues      []CriticalIssue             `json:"critical_issues,omitempty"`
	RecentAlerts        []HealthAlert               `json:"recent_alerts,omitempty"`
}

// HealthMetrics contains health monitoring metrics
type HealthMetrics struct {
	TotalHealthChecks   int64                    `json:"total_health_checks"`
	SuccessfulChecks    int64                    `json:"successful_checks"`
	FailedChecks        int64                    `json:"failed_checks"`
	AverageResponseTime time.Duration            `json:"average_response_time"`
	ChecksByAPI         map[string]int64         `json:"checks_by_api"`
	ResponseTimesByAPI  map[string]time.Duration `json:"response_times_by_api"`
	UptimeByAPI         map[string]float64       `json:"uptime_by_api"`
	AlertsTriggered     int64                    `json:"alerts_triggered"`
	LastCheckTime       time.Time                `json:"last_check_time"`
	MonitoringStartedAt time.Time                `json:"monitoring_started_at"`
}

// HealthStatistics contains statistical health data
type HealthStatistics struct {
	UptimePercentage    float64       `json:"uptime_percentage"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`
	P50ResponseTime     time.Duration `json:"p50_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	ErrorRate           float64       `json:"error_rate"`
}

// HealthTrends contains health trend analysis
type HealthTrends struct {
	UptimeTrend       TrendDirection `json:"uptime_trend"`
	ResponseTimeTrend TrendDirection `json:"response_time_trend"`
	ErrorRateTrend    TrendDirection `json:"error_rate_trend"`
	TrendPeriod       time.Duration  `json:"trend_period"`
	Confidence        float64        `json:"confidence"`
	Recommendations   []string       `json:"recommendations,omitempty"`
}

// CriticalIssue represents a critical health issue
type CriticalIssue struct {
	APIName     string        `json:"api_name"`
	IssueType   string        `json:"issue_type"`
	Severity    IssueSeverity `json:"severity"`
	Description string        `json:"description"`
	DetectedAt  time.Time     `json:"detected_at"`
	Duration    time.Duration `json:"duration,omitempty"`
	Impact      string        `json:"impact,omitempty"`
}

// HealthAlert represents a health monitoring alert
type HealthAlert struct {
	AlertID     uuid.UUID              `json:"alert_id"`
	APIName     string                 `json:"api_name"`
	AlertType   AlertType              `json:"alert_type"`
	Severity    AlertSeverity          `json:"severity"`
	Message     string                 `json:"message"`
	TriggeredAt time.Time              `json:"triggered_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Status      AlertStatus            `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Enums for health service
type HealthServiceStatus string

const (
	HealthServiceStatusHealthy   HealthServiceStatus = "healthy"
	HealthServiceStatusDegraded  HealthServiceStatus = "degraded"
	HealthServiceStatusUnhealthy HealthServiceStatus = "unhealthy"
	HealthServiceStatusUnknown   HealthServiceStatus = "unknown"
)

type ContentCheckType string

const (
	ContentCheckTypeContains    ContentCheckType = "contains"
	ContentCheckTypeRegex       ContentCheckType = "regex"
	ContentCheckTypeJSONPath    ContentCheckType = "json_path"
	ContentCheckTypeEquals      ContentCheckType = "equals"
	ContentCheckTypeNotContains ContentCheckType = "not_contains"
)

type TrendDirection string

const (
	TrendDirectionImproving TrendDirection = "improving"
	TrendDirectionStable    TrendDirection = "stable"
	TrendDirectionDeclining TrendDirection = "declining"
	TrendDirectionUnknown   TrendDirection = "unknown"
)

type IssueSeverity string

const (
	IssueSeverityLow      IssueSeverity = "low"
	IssueSeverityMedium   IssueSeverity = "medium"
	IssueSeverityHigh     IssueSeverity = "high"
	IssueSeverityCritical IssueSeverity = "critical"
)

type AlertType string

const (
	AlertTypeResponseTime        AlertType = "response_time"
	AlertTypeFailureRate         AlertType = "failure_rate"
	AlertTypeConsecutiveFailures AlertType = "consecutive_failures"
	AlertTypeContentValidation   AlertType = "content_validation"
	AlertTypeEndpointDown        AlertType = "endpoint_down"
)

type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

type AlertStatus string

const (
	AlertStatusActive   AlertStatus = "active"
	AlertStatusResolved AlertStatus = "resolved"
	AlertStatusMuted    AlertStatus = "muted"
)

// HealthServiceConfig contains configuration for health service
type HealthServiceConfig struct {
	DefaultTimeout       time.Duration `json:"default_timeout"`
	DefaultInterval      time.Duration `json:"default_interval"`
	DefaultRetryAttempts int           `json:"default_retry_attempts"`
	MaxConcurrentChecks  int           `json:"max_concurrent_checks"`
	HistoryRetentionDays int           `json:"history_retention_days"`
	EnableAuditLogging   bool          `json:"enable_audit_logging"`
	EnableAlerts         bool          `json:"enable_alerts"`
	AlertCooldownPeriod  time.Duration `json:"alert_cooldown_period"`
}

// healthServiceImpl implements HealthService
type healthServiceImpl struct {
	auditSvc         audit.Service
	config           *HealthServiceConfig
	httpClient       *http.Client
	registeredAPIs   map[string]*APIHealthConfig
	apisMutex        sync.RWMutex
	monitoringActive bool
	cancelMonitoring context.CancelFunc
	healthHistory    map[string][]*HealthCheckResponse
	historyMutex     sync.RWMutex
	metrics          *HealthMetrics
	metricsMutex     sync.RWMutex
}

// NewHealthService creates a new health monitoring service
func NewHealthService(auditSvc audit.Service, config *HealthServiceConfig) (HealthService, error) {
	if config == nil {
		config = &HealthServiceConfig{
			DefaultTimeout:       30 * time.Second,
			DefaultInterval:      5 * time.Minute,
			DefaultRetryAttempts: 3,
			MaxConcurrentChecks:  10,
			HistoryRetentionDays: 7,
			EnableAuditLogging:   true,
			EnableAlerts:         true,
			AlertCooldownPeriod:  5 * time.Minute,
		}
	}

	service := &healthServiceImpl{
		auditSvc:       auditSvc,
		config:         config,
		registeredAPIs: make(map[string]*APIHealthConfig),
		healthHistory:  make(map[string][]*HealthCheckResponse),
		httpClient: &http.Client{
			Timeout: config.DefaultTimeout,
		},
		metrics: &HealthMetrics{
			ChecksByAPI:         make(map[string]int64),
			ResponseTimesByAPI:  make(map[string]time.Duration),
			UptimeByAPI:         make(map[string]float64),
			MonitoringStartedAt: time.Now(),
		},
	}

	log.Printf("Health service initialized with %d second timeout", int(config.DefaultTimeout.Seconds()))
	return service, nil
}

// CheckHealth performs health check on a specific API
func (s *healthServiceImpl) CheckHealth(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
	startTime := time.Now()
	checkID := uuid.New()

	// Get API config if not provided in request
	var apiConfig *APIHealthConfig
	if req.APIName != "" {
		s.apisMutex.RLock()
		apiConfig = s.registeredAPIs[req.APIName]
		s.apisMutex.RUnlock()
	}

	// Use request parameters or fall back to API config
	endpoint := req.Endpoint
	method := req.Method
	headers := req.Headers
	timeout := req.Timeout
	expectedStatus := req.ExpectedStatus

	if apiConfig != nil {
		if endpoint == "" {
			endpoint = apiConfig.Endpoint
		}
		if method == "" {
			method = apiConfig.Method
		}
		if headers == nil {
			headers = apiConfig.Headers
		}
		if timeout == 0 {
			timeout = apiConfig.Timeout
		}
		if expectedStatus == nil {
			expectedStatus = apiConfig.ExpectedStatusCodes
		}
	}

	// Set defaults
	if method == "" {
		method = "GET"
	}
	if timeout == 0 {
		timeout = s.config.DefaultTimeout
	}
	if expectedStatus == nil {
		expectedStatus = []int{200, 201, 204}
	}

	response := &HealthCheckResponse{
		CheckID:   checkID,
		APIName:   req.APIName,
		Endpoint:  endpoint,
		CheckedAt: startTime,
		Metadata:  make(map[string]interface{}),
		Warnings:  []string{},
	}

	// Create HTTP request with timeout context
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(checkCtx, method, endpoint, nil)
	if err != nil {
		response.Status = HealthServiceStatusUnhealthy
		response.ErrorMessage = fmt.Sprintf("Failed to create request: %v", err)
		response.ResponseTime = time.Since(startTime)
		return response, nil
	}

	// Add headers
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	// Set User-Agent for identification
	httpReq.Header.Set("User-Agent", "KubeChat-HealthMonitor/1.0")

	// Execute request
	httpResp, err := s.httpClient.Do(httpReq)
	response.ResponseTime = time.Since(startTime)

	if err != nil {
		response.Status = HealthServiceStatusUnhealthy
		response.ErrorMessage = fmt.Sprintf("Request failed: %v", err)
	} else {
		defer httpResp.Body.Close()

		response.StatusCode = httpResp.StatusCode
		response.ResponseSize = httpResp.ContentLength

		// Check if status code is expected
		statusOK := false
		for _, expected := range expectedStatus {
			if httpResp.StatusCode == expected {
				statusOK = true
				break
			}
		}

		if statusOK {
			response.Status = HealthServiceStatusHealthy

			// Perform content validation if requested
			if req.ValidateContent && len(req.ContentChecks) > 0 {
				contentResults, err := s.performContentChecks(httpResp, req.ContentChecks)
				if err != nil {
					response.Warnings = append(response.Warnings, fmt.Sprintf("Content validation error: %v", err))
				} else {
					response.ContentResults = contentResults

					// Check if any content validation failed
					for _, result := range contentResults {
						if !result.Passed {
							response.Status = HealthServiceStatusDegraded
							break
						}
					}
				}
			}
		} else {
			response.Status = HealthServiceStatusUnhealthy
			response.ErrorMessage = fmt.Sprintf("Unexpected status code: %d", httpResp.StatusCode)
		}

		// Add response time warning if slow
		if response.ResponseTime > 5*time.Second {
			response.Warnings = append(response.Warnings, "Response time is slow")
		}

		// Store response metadata
		response.Metadata["content_type"] = httpResp.Header.Get("Content-Type")
		response.Metadata["server"] = httpResp.Header.Get("Server")
		if httpResp.Header.Get("X-RateLimit-Remaining") != "" {
			response.Metadata["rate_limit_remaining"] = httpResp.Header.Get("X-RateLimit-Remaining")
		}
	}

	// Update metrics
	s.updateHealthMetrics(req.APIName, response)

	// Store in history
	s.storeHealthHistory(req.APIName, response)

	// Log health check event
	s.logHealthEvent("health_check", req.APIName, response.Status == HealthServiceStatusHealthy, nil)

	return response, nil
}

// RegisterAPI registers an API for health monitoring
func (s *healthServiceImpl) RegisterAPI(ctx context.Context, config *APIHealthConfig) error {
	if config.Name == "" {
		return fmt.Errorf("API name is required")
	}

	if config.Endpoint == "" {
		return fmt.Errorf("API endpoint is required")
	}

	// Set defaults
	if config.Method == "" {
		config.Method = "GET"
	}
	if config.Timeout == 0 {
		config.Timeout = s.config.DefaultTimeout
	}
	if config.Interval == 0 {
		config.Interval = s.config.DefaultInterval
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = s.config.DefaultRetryAttempts
	}
	if config.RetryInterval == 0 {
		config.RetryInterval = 30 * time.Second
	}
	if len(config.ExpectedStatusCodes) == 0 {
		config.ExpectedStatusCodes = []int{200, 201, 204}
	}

	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	config.Enabled = true

	s.apisMutex.Lock()
	s.registeredAPIs[config.Name] = config
	s.apisMutex.Unlock()

	s.logHealthEvent("register_api", config.Name, true, nil)
	log.Printf("Registered API for health monitoring: %s -> %s", config.Name, config.Endpoint)

	return nil
}

// CheckAllAPIs performs health checks on all registered APIs
func (s *healthServiceImpl) CheckAllAPIs(ctx context.Context) (*AllAPIsHealthResponse, error) {
	s.apisMutex.RLock()
	apis := make([]*APIHealthConfig, 0, len(s.registeredAPIs))
	for _, api := range s.registeredAPIs {
		if api.Enabled {
			apis = append(apis, api)
		}
	}
	s.apisMutex.RUnlock()

	response := &AllAPIsHealthResponse{
		TotalAPIs:  len(apis),
		APIResults: make([]*HealthCheckResponse, 0, len(apis)),
		CheckedAt:  time.Now(),
	}

	if len(apis) == 0 {
		response.OverallStatus = HealthServiceStatusUnknown
		return response, nil
	}

	// Perform health checks concurrently
	sem := make(chan struct{}, s.config.MaxConcurrentChecks)
	results := make(chan *HealthCheckResponse, len(apis))
	errors := make(chan error, len(apis))

	for _, api := range apis {
		go func(api *APIHealthConfig) {
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			req := &HealthCheckRequest{
				APIName:         api.Name,
				Endpoint:        api.Endpoint,
				Method:          api.Method,
				Headers:         api.Headers,
				Timeout:         api.Timeout,
				ExpectedStatus:  api.ExpectedStatusCodes,
				ValidateContent: len(api.ContentChecks) > 0,
				ContentChecks:   api.ContentChecks,
			}

			result, err := s.CheckHealth(ctx, req)
			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}(api)
	}

	// Collect results
	for i := 0; i < len(apis); i++ {
		select {
		case result := <-results:
			response.APIResults = append(response.APIResults, result)

			switch result.Status {
			case HealthServiceStatusHealthy:
				response.HealthyAPIs++
			case HealthServiceStatusDegraded:
				response.DegradedAPIs++
			case HealthServiceStatusUnhealthy:
				response.UnhealthyAPIs++
			}
		case err := <-errors:
			log.Printf("Health check error: %v", err)
			response.UnhealthyAPIs++
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Determine overall status
	if response.UnhealthyAPIs == 0 && response.DegradedAPIs == 0 {
		response.OverallStatus = HealthServiceStatusHealthy
	} else if response.UnhealthyAPIs > response.HealthyAPIs {
		response.OverallStatus = HealthServiceStatusUnhealthy
	} else {
		response.OverallStatus = HealthServiceStatusDegraded
	}

	// Generate summary
	response.Summary = s.generateHealthSummary(response)

	return response, nil
}

// Helper methods

func (s *healthServiceImpl) performContentChecks(resp *http.Response, checks []ContentCheck) ([]ContentResult, error) {
	results := make([]ContentResult, 0, len(checks))

	// Read response body
	body := make([]byte, resp.ContentLength)
	if resp.ContentLength > 0 {
		resp.Body.Read(body)
	}

	for _, check := range checks {
		result := ContentResult{
			CheckType: check.Type,
			Pattern:   check.Pattern,
			JsonPath:  check.JsonPath,
			Expected:  check.Expected,
		}

		switch check.Type {
		case ContentCheckTypeContains:
			result.Passed = strings.Contains(string(body), check.Pattern)
		case ContentCheckTypeNotContains:
			result.Passed = !strings.Contains(string(body), check.Pattern)
		case ContentCheckTypeEquals:
			result.Actual = string(body)
			result.Passed = string(body) == fmt.Sprintf("%v", check.Expected)
		case ContentCheckTypeJSONPath:
			// JSON path checking would be implemented here
			result.Passed = true // Placeholder
		case ContentCheckTypeRegex:
			// Regex checking would be implemented here
			result.Passed = true // Placeholder
		}

		if !result.Passed && check.Required {
			result.ErrorMessage = fmt.Sprintf("Required content check failed: %s", check.Description)
		}

		results = append(results, result)
	}

	return results, nil
}

func (s *healthServiceImpl) updateHealthMetrics(apiName string, response *HealthCheckResponse) {
	s.metricsMutex.Lock()
	defer s.metricsMutex.Unlock()

	s.metrics.TotalHealthChecks++
	s.metrics.ChecksByAPI[apiName]++
	s.metrics.LastCheckTime = response.CheckedAt

	if response.Status == HealthServiceStatusHealthy {
		s.metrics.SuccessfulChecks++
	} else {
		s.metrics.FailedChecks++
	}

	// Update response time
	totalTime := s.metrics.AverageResponseTime * time.Duration(s.metrics.TotalHealthChecks-1)
	s.metrics.AverageResponseTime = (totalTime + response.ResponseTime) / time.Duration(s.metrics.TotalHealthChecks)

	s.metrics.ResponseTimesByAPI[apiName] = response.ResponseTime

	// Update uptime percentage
	apiChecks := s.metrics.ChecksByAPI[apiName]
	if apiChecks > 0 {
		successRate := float64(1)
		if response.Status != HealthServiceStatusHealthy {
			successRate = 0
		}

		currentUptime := s.metrics.UptimeByAPI[apiName]
		s.metrics.UptimeByAPI[apiName] = (currentUptime*float64(apiChecks-1) + successRate) / float64(apiChecks)
	}
}

func (s *healthServiceImpl) storeHealthHistory(apiName string, response *HealthCheckResponse) {
	s.historyMutex.Lock()
	defer s.historyMutex.Unlock()

	if _, exists := s.healthHistory[apiName]; !exists {
		s.healthHistory[apiName] = make([]*HealthCheckResponse, 0)
	}

	s.healthHistory[apiName] = append(s.healthHistory[apiName], response)

	// Keep only recent history based on retention policy
	maxHistory := s.config.HistoryRetentionDays * 24 * 12 // Assuming 5-minute intervals
	if len(s.healthHistory[apiName]) > maxHistory {
		s.healthHistory[apiName] = s.healthHistory[apiName][1:]
	}
}

func (s *healthServiceImpl) generateHealthSummary(response *AllAPIsHealthResponse) *HealthSummary {
	summary := &HealthSummary{
		OverallStatus: response.OverallStatus,
		TotalAPIs:     response.TotalAPIs,
		APIsByStatus: map[HealthServiceStatus]int{
			HealthServiceStatusHealthy:   response.HealthyAPIs,
			HealthServiceStatusDegraded:  response.DegradedAPIs,
			HealthServiceStatusUnhealthy: response.UnhealthyAPIs,
		},
		TotalChecks:      int64(response.TotalAPIs),
		TotalFailures:    int64(response.UnhealthyAPIs + response.DegradedAPIs),
		UptimePercentage: float64(response.HealthyAPIs) / float64(response.TotalAPIs) * 100,
		LastUpdated:      response.CheckedAt,
		CriticalIssues:   []CriticalIssue{},
		RecentAlerts:     []HealthAlert{},
	}

	// Calculate average response time
	if len(response.APIResults) > 0 {
		totalTime := time.Duration(0)
		for _, result := range response.APIResults {
			totalTime += result.ResponseTime
		}
		summary.AverageResponseTime = totalTime / time.Duration(len(response.APIResults))
	}

	return summary
}

func (s *healthServiceImpl) logHealthEvent(action, apiName string, success bool, err error) {
	if !s.config.EnableAuditLogging || s.auditSvc == nil {
		return
	}

	safetyLevel := models.SafetyLevelSafe
	if err != nil {
		safetyLevel = models.SafetyLevelWarning
	}

	auditReq := models.AuditLogRequest{
		QueryText:        fmt.Sprintf("Health monitoring: %s", action),
		GeneratedCommand: fmt.Sprintf("Health %s for API %s", action, apiName),
		SafetyLevel:      string(safetyLevel),
		ExecutionResult: map[string]interface{}{
			"api_name": apiName,
			"action":   action,
			"success":  success,
		},
		ExecutionStatus: func() string {
			if success {
				return models.ExecutionStatusSuccess
			}
			return models.ExecutionStatusFailed
		}(),
	}

	if err != nil {
		auditReq.ExecutionResult["error"] = err.Error()
	}

	if logErr := s.auditSvc.LogUserAction(context.Background(), auditReq); logErr != nil {
		log.Printf("Failed to log health event: %v", logErr)
	}
}

// Placeholder implementations for remaining interface methods
func (s *healthServiceImpl) UpdateAPIConfig(ctx context.Context, config *APIHealthConfig) error {
	return nil
}

func (s *healthServiceImpl) RemoveAPI(ctx context.Context, apiName string) error {
	return nil
}

func (s *healthServiceImpl) GetAPIHealthHistory(ctx context.Context, apiName string, timeframe *TimeFrame) (*HealthHistoryResponse, error) {
	return &HealthHistoryResponse{}, nil
}

func (s *healthServiceImpl) GetHealthSummary(ctx context.Context) (*HealthSummary, error) {
	return &HealthSummary{}, nil
}

func (s *healthServiceImpl) StartContinuousMonitoring(ctx context.Context) error {
	return nil
}

func (s *healthServiceImpl) StopContinuousMonitoring() error {
	return nil
}

func (s *healthServiceImpl) GetHealthMetrics(ctx context.Context) (*HealthMetrics, error) {
	s.metricsMutex.RLock()
	defer s.metricsMutex.RUnlock()

	// Return a copy of metrics
	metrics := *s.metrics
	return &metrics, nil
}
