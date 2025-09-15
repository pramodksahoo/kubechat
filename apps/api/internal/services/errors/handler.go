package errors

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// Handler defines the error handling interface
type Handler interface {
	// Error processing and classification
	ProcessError(ctx context.Context, err error, execution *models.KubernetesCommandExecution) (*ProcessedError, error)
	ClassifyError(err error) *ErrorClassification

	// Error recovery and suggestions
	SuggestRecovery(ctx context.Context, processedError *ProcessedError) *RecoveryRecommendation
	GetCommonSolutions(errorType ErrorType) []string

	// Error tracking and analytics
	TrackError(ctx context.Context, processedError *ProcessedError) error
	GetErrorStatistics(ctx context.Context, timeframe time.Duration) (*ErrorStatistics, error)

	// Context enrichment
	EnrichWithContext(processedError *ProcessedError, execution *models.KubernetesCommandExecution) error
	AddStackTrace(processedError *ProcessedError, err error) error
}

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeAuthentication     ErrorType = "authentication"
	ErrorTypeAuthorization      ErrorType = "authorization"
	ErrorTypeResourceNotFound   ErrorType = "resource_not_found"
	ErrorTypeTimeout            ErrorType = "timeout"
	ErrorTypeNetworkError       ErrorType = "network_error"
	ErrorTypeClusterUnavailable ErrorType = "cluster_unavailable"
	ErrorTypeInvalidRequest     ErrorType = "invalid_request"
	ErrorTypeRateLimited        ErrorType = "rate_limited"
	ErrorTypeServiceError       ErrorType = "service_error"
	ErrorTypeUnknown            ErrorType = "unknown"
)

// ErrorSeverity represents the severity level of errors
type ErrorSeverity string

const (
	SeverityCritical ErrorSeverity = "critical"
	SeverityHigh     ErrorSeverity = "high"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityLow      ErrorSeverity = "low"
	SeverityInfo     ErrorSeverity = "info"
)

// ProcessedError represents a processed and enriched error
type ProcessedError struct {
	ID             uuid.UUID               `json:"id"`
	OriginalError  string                  `json:"original_error"`
	Classification *ErrorClassification    `json:"classification"`
	Context        *ErrorContext           `json:"context"`
	Recovery       *RecoveryRecommendation `json:"recovery,omitempty"`
	StackTrace     []string                `json:"stack_trace,omitempty"`
	Metadata       map[string]interface{}  `json:"metadata"`
	ProcessedAt    time.Time               `json:"processed_at"`

	// Execution context
	ExecutionID uuid.UUID `json:"execution_id,omitempty"`
	UserID      uuid.UUID `json:"user_id,omitempty"`
	Command     string    `json:"command,omitempty"`
}

// ErrorClassification provides detailed error categorization
type ErrorClassification struct {
	Type        ErrorType     `json:"type"`
	Severity    ErrorSeverity `json:"severity"`
	Category    string        `json:"category"`
	SubCategory string        `json:"sub_category,omitempty"`
	Code        string        `json:"code,omitempty"`
	Message     string        `json:"message"`
	Retryable   bool          `json:"retryable"`
	UserFacing  bool          `json:"user_facing"`
	Tags        []string      `json:"tags,omitempty"`
}

// ErrorContext provides contextual information about when/where error occurred
type ErrorContext struct {
	Timestamp  time.Time              `json:"timestamp"`
	Operation  string                 `json:"operation,omitempty"`
	Resource   string                 `json:"resource,omitempty"`
	Namespace  string                 `json:"namespace,omitempty"`
	Cluster    string                 `json:"cluster,omitempty"`
	UserRole   string                 `json:"user_role,omitempty"`
	Component  string                 `json:"component"`
	Function   string                 `json:"function,omitempty"`
	LineNumber int                    `json:"line_number,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	Additional map[string]interface{} `json:"additional,omitempty"`
}

// RecoveryRecommendation provides suggestions for error recovery
type RecoveryRecommendation struct {
	AutoRecoverable bool                `json:"auto_recoverable"`
	Solutions       []RecoverySolution  `json:"solutions"`
	PreventionTips  []string            `json:"prevention_tips,omitempty"`
	Documentation   []DocumentationLink `json:"documentation,omitempty"`
	Urgency         string              `json:"urgency"` // "immediate", "high", "medium", "low"
}

// RecoverySolution represents a potential solution to an error
type RecoverySolution struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Steps         []string `json:"steps"`
	Automated     bool     `json:"automated"`
	Difficulty    string   `json:"difficulty"` // "easy", "medium", "hard"
	EstimatedTime string   `json:"estimated_time,omitempty"`
}

// DocumentationLink provides links to relevant documentation
type DocumentationLink struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"` // "official", "community", "troubleshooting"
}

// ErrorStatistics provides analytics about errors over time
type ErrorStatistics struct {
	TimeFrame         time.Duration           `json:"time_frame"`
	TotalErrors       int64                   `json:"total_errors"`
	ErrorsByType      map[ErrorType]int64     `json:"errors_by_type"`
	ErrorsBySeverity  map[ErrorSeverity]int64 `json:"errors_by_severity"`
	TopErrors         []ErrorPattern          `json:"top_errors"`
	ResolutionRate    float64                 `json:"resolution_rate"`
	AvgResolutionTime time.Duration           `json:"avg_resolution_time"`
	TrendAnalysis     *ErrorTrend             `json:"trend_analysis"`
}

// ErrorPattern represents a pattern of recurring errors
type ErrorPattern struct {
	Pattern   string    `json:"pattern"`
	Count     int64     `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	Trend     string    `json:"trend"` // "increasing", "decreasing", "stable"
}

// ErrorTrend provides trend analysis of errors
type ErrorTrend struct {
	Direction      string  `json:"direction"` // "up", "down", "stable"
	ChangePercent  float64 `json:"change_percent"`
	PredictedCount int64   `json:"predicted_count"`
	Confidence     float64 `json:"confidence"`
}

// handler implements the error Handler interface
type handler struct {
	errorPatterns   map[string]*ErrorPattern
	commonSolutions map[ErrorType][]string
}

// NewHandler creates a new error handler
func NewHandler() Handler {
	h := &handler{
		errorPatterns:   make(map[string]*ErrorPattern),
		commonSolutions: initializeCommonSolutions(),
	}

	return h
}

// ProcessError processes an error and enriches it with context and classification
func (h *handler) ProcessError(ctx context.Context, err error, execution *models.KubernetesCommandExecution) (*ProcessedError, error) {
	if err == nil {
		return nil, fmt.Errorf("cannot process nil error")
	}

	processedError := &ProcessedError{
		ID:            uuid.New(),
		OriginalError: err.Error(),
		ProcessedAt:   time.Now(),
		Metadata:      make(map[string]interface{}),
	}

	// Classify the error
	processedError.Classification = h.ClassifyError(err)

	// Create error context
	processedError.Context = &ErrorContext{
		Timestamp:  time.Now(),
		Component:  "command-execution",
		Additional: make(map[string]interface{}),
	}

	// Add stack trace
	h.AddStackTrace(processedError, err)

	// Enrich with execution context if provided
	if execution != nil {
		h.EnrichWithContext(processedError, execution)
	}

	// Get recovery recommendations
	processedError.Recovery = h.SuggestRecovery(ctx, processedError)

	return processedError, nil
}

// ClassifyError classifies an error into categories and determines its properties
func (h *handler) ClassifyError(err error) *ErrorClassification {
	errorMsg := strings.ToLower(err.Error())

	classification := &ErrorClassification{
		Type:       ErrorTypeUnknown,
		Severity:   SeverityMedium,
		Message:    err.Error(),
		UserFacing: true,
		Tags:       []string{},
	}

	// Authentication errors
	if strings.Contains(errorMsg, "unauthorized") || strings.Contains(errorMsg, "authentication") {
		classification.Type = ErrorTypeAuthentication
		classification.Category = "security"
		classification.Severity = SeverityHigh
		classification.Retryable = false
		classification.Tags = append(classification.Tags, "auth", "security")
	}

	// Authorization errors
	if strings.Contains(errorMsg, "forbidden") || strings.Contains(errorMsg, "permission") {
		classification.Type = ErrorTypeAuthorization
		classification.Category = "security"
		classification.Severity = SeverityHigh
		classification.Retryable = false
		classification.Tags = append(classification.Tags, "auth", "rbac")
	}

	// Resource not found errors
	if strings.Contains(errorMsg, "not found") || strings.Contains(errorMsg, "does not exist") {
		classification.Type = ErrorTypeResourceNotFound
		classification.Category = "resource"
		classification.Severity = SeverityMedium
		classification.Retryable = false
		classification.Tags = append(classification.Tags, "resource", "missing")
	}

	// Timeout errors
	if strings.Contains(errorMsg, "timeout") || strings.Contains(errorMsg, "deadline exceeded") {
		classification.Type = ErrorTypeTimeout
		classification.Category = "performance"
		classification.Severity = SeverityHigh
		classification.Retryable = true
		classification.Tags = append(classification.Tags, "timeout", "performance")
	}

	// Network errors
	if strings.Contains(errorMsg, "connection") || strings.Contains(errorMsg, "network") {
		classification.Type = ErrorTypeNetworkError
		classification.Category = "network"
		classification.Severity = SeverityHigh
		classification.Retryable = true
		classification.Tags = append(classification.Tags, "network", "connectivity")
	}

	// Cluster unavailable errors
	if strings.Contains(errorMsg, "cluster") && (strings.Contains(errorMsg, "unavailable") || strings.Contains(errorMsg, "unreachable")) {
		classification.Type = ErrorTypeClusterUnavailable
		classification.Category = "infrastructure"
		classification.Severity = SeverityCritical
		classification.Retryable = true
		classification.Tags = append(classification.Tags, "cluster", "infrastructure")
	}

	// Rate limiting errors
	if strings.Contains(errorMsg, "rate limit") || strings.Contains(errorMsg, "too many requests") {
		classification.Type = ErrorTypeRateLimited
		classification.Category = "throttling"
		classification.Severity = SeverityMedium
		classification.Retryable = true
		classification.Tags = append(classification.Tags, "rate-limit", "throttling")
	}

	// Invalid request errors
	if strings.Contains(errorMsg, "invalid") || strings.Contains(errorMsg, "bad request") {
		classification.Type = ErrorTypeInvalidRequest
		classification.Category = "validation"
		classification.Severity = SeverityLow
		classification.Retryable = false
		classification.Tags = append(classification.Tags, "validation", "input")
	}

	return classification
}

// SuggestRecovery provides recovery recommendations based on error analysis
func (h *handler) SuggestRecovery(ctx context.Context, processedError *ProcessedError) *RecoveryRecommendation {
	recommendation := &RecoveryRecommendation{
		Solutions: []RecoverySolution{},
	}

	switch processedError.Classification.Type {
	case ErrorTypeAuthentication:
		recommendation.AutoRecoverable = false
		recommendation.Urgency = "high"
		recommendation.Solutions = append(recommendation.Solutions, RecoverySolution{
			ID:            "reauth",
			Title:         "Re-authenticate",
			Description:   "Your authentication token has expired or is invalid",
			Steps:         []string{"Log out", "Log back in with valid credentials", "Retry the operation"},
			Automated:     false,
			Difficulty:    "easy",
			EstimatedTime: "2-3 minutes",
		})

	case ErrorTypeAuthorization:
		recommendation.AutoRecoverable = false
		recommendation.Urgency = "high"
		recommendation.Solutions = append(recommendation.Solutions, RecoverySolution{
			ID:            "check_permissions",
			Title:         "Check Permissions",
			Description:   "You don't have sufficient permissions for this operation",
			Steps:         []string{"Contact your administrator", "Request appropriate permissions", "Verify RBAC settings"},
			Automated:     false,
			Difficulty:    "medium",
			EstimatedTime: "5-10 minutes",
		})

	case ErrorTypeTimeout:
		recommendation.AutoRecoverable = true
		recommendation.Urgency = "medium"
		recommendation.Solutions = append(recommendation.Solutions,
			RecoverySolution{
				ID:            "retry",
				Title:         "Retry Operation",
				Description:   "The operation timed out and can be retried",
				Steps:         []string{"Wait a moment", "Retry the same operation"},
				Automated:     true,
				Difficulty:    "easy",
				EstimatedTime: "1 minute",
			},
			RecoverySolution{
				ID:            "increase_timeout",
				Title:         "Increase Timeout",
				Description:   "Consider increasing the timeout for similar operations",
				Steps:         []string{"Contact administrator", "Request timeout adjustment"},
				Automated:     false,
				Difficulty:    "medium",
				EstimatedTime: "5 minutes",
			},
		)

	case ErrorTypeResourceNotFound:
		recommendation.AutoRecoverable = false
		recommendation.Urgency = "low"
		recommendation.Solutions = append(recommendation.Solutions, RecoverySolution{
			ID:            "verify_resource",
			Title:         "Verify Resource Exists",
			Description:   "The requested resource was not found",
			Steps:         []string{"Check resource name spelling", "Verify namespace", "List available resources"},
			Automated:     false,
			Difficulty:    "easy",
			EstimatedTime: "2-3 minutes",
		})

	case ErrorTypeNetworkError:
		recommendation.AutoRecoverable = true
		recommendation.Urgency = "high"
		recommendation.Solutions = append(recommendation.Solutions,
			RecoverySolution{
				ID:            "retry_network",
				Title:         "Retry After Network Recovery",
				Description:   "Network connectivity issue detected",
				Steps:         []string{"Check network connectivity", "Wait for network recovery", "Retry operation"},
				Automated:     true,
				Difficulty:    "easy",
				EstimatedTime: "2-5 minutes",
			},
		)

	default:
		recommendation.AutoRecoverable = false
		recommendation.Urgency = "medium"
		recommendation.Solutions = append(recommendation.Solutions, RecoverySolution{
			ID:            "contact_support",
			Title:         "Contact Support",
			Description:   "This error requires manual investigation",
			Steps:         []string{"Collect error details", "Contact system administrator", "Provide error context"},
			Automated:     false,
			Difficulty:    "medium",
			EstimatedTime: "10-15 minutes",
		})
	}

	// Add common prevention tips
	recommendation.PreventionTips = []string{
		"Verify your permissions before executing commands",
		"Check resource names and namespaces carefully",
		"Monitor cluster health regularly",
		"Keep authentication tokens updated",
	}

	// Add documentation links based on error type
	recommendation.Documentation = []DocumentationLink{
		{
			Title: "Kubernetes Troubleshooting Guide",
			URL:   "https://kubernetes.io/docs/tasks/debug-application-cluster/",
			Type:  "official",
		},
	}

	return recommendation
}

// GetCommonSolutions returns common solutions for specific error types
func (h *handler) GetCommonSolutions(errorType ErrorType) []string {
	if solutions, exists := h.commonSolutions[errorType]; exists {
		return solutions
	}
	return []string{"Contact system administrator for assistance"}
}

// TrackError tracks an error for analytics and pattern recognition
func (h *handler) TrackError(ctx context.Context, processedError *ProcessedError) error {
	// Create or update error pattern
	patternKey := string(processedError.Classification.Type)

	if pattern, exists := h.errorPatterns[patternKey]; exists {
		pattern.Count++
		pattern.LastSeen = time.Now()
	} else {
		h.errorPatterns[patternKey] = &ErrorPattern{
			Pattern:   patternKey,
			Count:     1,
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
		}
	}

	// In a real implementation, this would persist to a database
	return nil
}

// GetErrorStatistics returns error statistics for the specified timeframe
func (h *handler) GetErrorStatistics(ctx context.Context, timeframe time.Duration) (*ErrorStatistics, error) {
	stats := &ErrorStatistics{
		TimeFrame:        timeframe,
		ErrorsByType:     make(map[ErrorType]int64),
		ErrorsBySeverity: make(map[ErrorSeverity]int64),
		TopErrors:        []ErrorPattern{},
	}

	// Aggregate statistics from error patterns
	for _, pattern := range h.errorPatterns {
		errorType := ErrorType(pattern.Pattern)
		stats.ErrorsByType[errorType] = pattern.Count
		stats.TotalErrors += pattern.Count

		stats.TopErrors = append(stats.TopErrors, *pattern)
	}

	// Calculate resolution rate (simplified)
	stats.ResolutionRate = 0.75 // 75% placeholder
	stats.AvgResolutionTime = 5 * time.Minute

	// Basic trend analysis
	stats.TrendAnalysis = &ErrorTrend{
		Direction:      "stable",
		ChangePercent:  0.0,
		PredictedCount: stats.TotalErrors,
		Confidence:     0.8,
	}

	return stats, nil
}

// EnrichWithContext enriches the processed error with execution context
func (h *handler) EnrichWithContext(processedError *ProcessedError, execution *models.KubernetesCommandExecution) error {
	if processedError.Context == nil {
		processedError.Context = &ErrorContext{
			Additional: make(map[string]interface{}),
		}
	}

	// Add execution details
	processedError.ExecutionID = execution.ID
	processedError.UserID = execution.UserID
	processedError.Command = execution.Command.Operation

	processedError.Context.Operation = execution.Command.Operation
	processedError.Context.Resource = execution.Command.Resource
	processedError.Context.Namespace = execution.Command.Namespace

	// Add additional metadata
	processedError.Metadata["safety_level"] = execution.SafetyLevel
	processedError.Metadata["execution_status"] = execution.Status
	processedError.Metadata["created_at"] = execution.CreatedAt

	return nil
}

// AddStackTrace adds stack trace information to the processed error
func (h *handler) AddStackTrace(processedError *ProcessedError, err error) error {
	// Get stack trace
	stackTrace := make([]string, 0)

	// Capture current stack
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	pc = pc[:n]

	frames := runtime.CallersFrames(pc)
	for {
		frame, more := frames.Next()
		stackTrace = append(stackTrace, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))

		if !more {
			break
		}
	}

	processedError.StackTrace = stackTrace

	// Add caller info to context
	if len(stackTrace) > 0 {
		parts := strings.Split(stackTrace[0], " ")
		if len(parts) >= 2 {
			processedError.Context.Function = parts[1]
			fileParts := strings.Split(parts[0], ":")
			if len(fileParts) >= 2 {
				// Extract line number from file path
				var lineNum int
				fmt.Sscanf(fileParts[len(fileParts)-1], "%d", &lineNum)
				processedError.Context.LineNumber = lineNum
			}
		}
	}

	return nil
}

// initializeCommonSolutions initializes common solutions for different error types
func initializeCommonSolutions() map[ErrorType][]string {
	return map[ErrorType][]string{
		ErrorTypeAuthentication: {
			"Re-authenticate with valid credentials",
			"Check if authentication token has expired",
			"Verify API server connectivity",
		},
		ErrorTypeAuthorization: {
			"Contact your administrator for proper permissions",
			"Verify RBAC configuration",
			"Check if you have access to the specified namespace",
		},
		ErrorTypeResourceNotFound: {
			"Verify the resource name and namespace",
			"Check if the resource exists using list commands",
			"Ensure you're connected to the correct cluster",
		},
		ErrorTypeTimeout: {
			"Retry the operation",
			"Check cluster performance and load",
			"Consider increasing timeout values",
		},
		ErrorTypeNetworkError: {
			"Check network connectivity",
			"Verify cluster endpoint accessibility",
			"Check firewall and proxy settings",
		},
		ErrorTypeClusterUnavailable: {
			"Wait for cluster to become available",
			"Check cluster status and health",
			"Contact cluster administrator",
		},
		ErrorTypeInvalidRequest: {
			"Review command syntax and parameters",
			"Check resource specifications",
			"Validate input data format",
		},
		ErrorTypeRateLimited: {
			"Reduce request frequency",
			"Implement exponential backoff",
			"Contact administrator about rate limits",
		},
	}
}
