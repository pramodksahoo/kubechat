package models

import "time"

// HealthStatus represents the overall health status
type HealthStatus struct {
	Status     HealthStatusEnum            `json:"status"`
	Message    string                      `json:"message"`
	Timestamp  time.Time                   `json:"timestamp"`
	Components map[string]*ComponentHealth `json:"components"`
	Metadata   map[string]interface{}      `json:"metadata,omitempty"`
}

// ComponentHealth represents the health status of a single component
type ComponentHealth struct {
	Name      string                 `json:"name"`
	Status    HealthStatusEnum       `json:"status"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// HealthStatusEnum represents the possible health states
type HealthStatusEnum string

const (
	HealthStatusHealthy   HealthStatusEnum = "healthy"
	HealthStatusDegraded  HealthStatusEnum = "degraded"
	HealthStatusUnhealthy HealthStatusEnum = "unhealthy"
	HealthStatusUnknown   HealthStatusEnum = "unknown"
)

// HealthMetrics represents comprehensive health metrics
type HealthMetrics struct {
	StartTime        time.Time                    `json:"start_time"`
	Uptime           time.Duration                `json:"uptime" swaggertype:"string" format:"duration" example:"48h30m"`
	LastCheck        time.Time                    `json:"last_check"`
	TotalChecks      int64                        `json:"total_checks"`
	FailedChecks     int64                        `json:"failed_checks"`
	SuccessRate      float64                      `json:"success_rate"`
	ComponentMetrics map[string]*ComponentMetrics `json:"component_metrics"`
}

// ComponentMetrics represents metrics for a single component
type ComponentMetrics struct {
	Name                string           `json:"name"`
	TotalChecks         int64            `json:"total_checks"`
	FailedChecks        int64            `json:"failed_checks"`
	SuccessRate         float64          `json:"success_rate"`
	LastCheckTime       time.Time        `json:"last_check_time"`
	AverageResponseTime time.Duration    `json:"average_response_time" swaggertype:"string" format:"duration" example:"15ms"`
	LastStatus          HealthStatusEnum `json:"last_status"`
}

// HealthCheckConfig represents the configuration for health checks
type HealthCheckConfig struct {
	Enabled          bool          `json:"enabled"`
	Interval         time.Duration `json:"interval" swaggertype:"string" format:"duration" example:"30s"`
	Timeout          time.Duration `json:"timeout" swaggertype:"string" format:"duration" example:"10s"`
	RetryAttempts    int           `json:"retry_attempts"`
	FailureThreshold int           `json:"failure_threshold"`
}

// AlertConfig represents the configuration for health alerts
type AlertConfig struct {
	Enabled         bool          `json:"enabled"`
	Threshold       time.Duration `json:"threshold" swaggertype:"string" format:"duration" example:"5m"`
	NotificationURL string        `json:"notification_url,omitempty"`
	EmailRecipients []string      `json:"email_recipients,omitempty"`
	SlackWebhookURL string        `json:"slack_webhook_url,omitempty"`
}

// DatabaseHealth represents database-specific health information
type DatabaseHealth struct {
	ConnectionStatus string        `json:"connection_status"`
	OpenConnections  int           `json:"open_connections"`
	InUseConnections int           `json:"in_use_connections"`
	IdleConnections  int           `json:"idle_connections"`
	WaitCount        int64         `json:"wait_count"`
	WaitDuration     time.Duration `json:"wait_duration" swaggertype:"string" format:"duration" example:"1ms"`
	MaxOpenConns     int           `json:"max_open_conns"`
	MaxIdleConns     int           `json:"max_idle_conns"`
	ConnMaxLifetime  time.Duration `json:"conn_max_lifetime" swaggertype:"string" format:"duration" example:"30m"`
	ResponseTime     time.Duration `json:"response_time" swaggertype:"string" format:"duration" example:"5ms"`
}

// RedisHealth represents Redis-specific health information
type RedisHealth struct {
	ConnectionStatus string        `json:"connection_status"`
	ResponseTime     time.Duration `json:"response_time" swaggertype:"string" format:"duration" example:"5ms"`
	Version          string        `json:"version,omitempty"`
	ConnectedClients int           `json:"connected_clients,omitempty"`
	UsedMemory       int64         `json:"used_memory,omitempty"`
	KeyspaceHits     int64         `json:"keyspace_hits,omitempty"`
	KeyspaceMisses   int64         `json:"keyspace_misses,omitempty"`
}

// ServiceHealth represents microservice-specific health information
type ServiceHealth struct {
	ServiceName   string           `json:"service_name"`
	Version       string           `json:"version"`
	Status        HealthStatusEnum `json:"status"`
	Dependencies  []string         `json:"dependencies"`
	Endpoints     []EndpointHealth `json:"endpoints,omitempty"`
	ResourceUsage *ResourceUsage   `json:"resource_usage,omitempty"`
}

// EndpointHealth represents the health of individual endpoints
type EndpointHealth struct {
	Path         string           `json:"path"`
	Method       string           `json:"method"`
	Status       HealthStatusEnum `json:"status"`
	ResponseTime time.Duration    `json:"response_time"`
	LastChecked  time.Time        `json:"last_checked"`
	ErrorRate    float64          `json:"error_rate"`
}

// ResourceUsage represents system resource usage
type ResourceUsage struct {
	CPUPercent     float64 `json:"cpu_percent"`
	MemoryPercent  float64 `json:"memory_percent"`
	MemoryUsed     int64   `json:"memory_used_bytes"`
	MemoryTotal    int64   `json:"memory_total_bytes"`
	DiskUsed       int64   `json:"disk_used_bytes"`
	DiskTotal      int64   `json:"disk_total_bytes"`
	GoroutineCount int     `json:"goroutine_count"`
}

// HealthAlert represents a health alert
type HealthAlert struct {
	ID         string                 `json:"id"`
	Component  string                 `json:"component"`
	Status     HealthStatusEnum       `json:"status"`
	Message    string                 `json:"message"`
	Severity   AlertSeverity          `json:"severity"`
	Timestamp  time.Time              `json:"timestamp"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt *time.Time             `json:"resolved_at,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AlertSeverity represents the severity level of alerts
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// HealthHistoryEntry represents a historical health check entry
type HealthHistoryEntry struct {
	Timestamp    time.Time              `json:"timestamp"`
	Status       HealthStatusEnum       `json:"status"`
	Component    string                 `json:"component"`
	Message      string                 `json:"message"`
	ResponseTime time.Duration          `json:"response_time,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// HealthSummary represents a summary of health status over time
type HealthSummary struct {
	Period              string             `json:"period"`
	StartTime           time.Time          `json:"start_time"`
	EndTime             time.Time          `json:"end_time"`
	OverallUptime       float64            `json:"overall_uptime_percent"`
	ComponentUptime     map[string]float64 `json:"component_uptime_percent"`
	IncidentCount       int                `json:"incident_count"`
	AverageResponseTime time.Duration      `json:"average_response_time" swaggertype:"string" format:"duration" example:"20ms"`
	SloViolations       int                `json:"slo_violations"`
}

// SLAConfig represents Service Level Agreement configuration
type SLAConfig struct {
	UptimeTarget       float64       `json:"uptime_target_percent"`                                                       // e.g., 99.9
	ResponseTimeTarget time.Duration `json:"response_time_target" swaggertype:"string" format:"duration" example:"200ms"` // e.g., 200ms
	ErrorRateTarget    float64       `json:"error_rate_target_percent"`                                                   // e.g., 1.0
	MeasurementPeriod  string        `json:"measurement_period"`                                                          // e.g., "monthly"
}

// SLAStatus represents current SLA compliance status
type SLAStatus struct {
	Config              SLAConfig     `json:"config"`
	CurrentUptime       float64       `json:"current_uptime_percent"`
	CurrentResponseTime time.Duration `json:"current_avg_response_time" swaggertype:"string" format:"duration" example:"150ms"`
	CurrentErrorRate    float64       `json:"current_error_rate_percent"`
	Compliant           bool          `json:"compliant"`
	ViolationCount      int           `json:"violation_count"`
	LastViolation       *time.Time    `json:"last_violation,omitempty"`
}

// HealthCheckResult represents the result of a health check operation
type HealthCheckResult struct {
	Success      bool                   `json:"success"`
	Status       HealthStatusEnum       `json:"status"`
	Message      string                 `json:"message"`
	ResponseTime time.Duration          `json:"response_time"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Error        error                  `json:"error,omitempty"`
}

// Methods for calculating derived values

// CalculateSuccessRate calculates the success rate for health metrics
func (hm *HealthMetrics) CalculateSuccessRate() {
	if hm.TotalChecks > 0 {
		hm.SuccessRate = float64(hm.TotalChecks-hm.FailedChecks) / float64(hm.TotalChecks) * 100
	} else {
		hm.SuccessRate = 0
	}
}

// CalculateSuccessRate calculates the success rate for component metrics
func (cm *ComponentMetrics) CalculateSuccessRate() {
	if cm.TotalChecks > 0 {
		cm.SuccessRate = float64(cm.TotalChecks-cm.FailedChecks) / float64(cm.TotalChecks) * 100
	} else {
		cm.SuccessRate = 0
	}
}

// IsHealthy returns true if the status is healthy
func (s HealthStatusEnum) IsHealthy() bool {
	return s == HealthStatusHealthy
}

// IsDegraded returns true if the status is degraded
func (s HealthStatusEnum) IsDegraded() bool {
	return s == HealthStatusDegraded
}

// IsUnhealthy returns true if the status is unhealthy
func (s HealthStatusEnum) IsUnhealthy() bool {
	return s == HealthStatusUnhealthy
}

// GetSeverityLevel returns the alert severity for a health status
func (s HealthStatusEnum) GetSeverityLevel() AlertSeverity {
	switch s {
	case HealthStatusHealthy:
		return AlertSeverityInfo
	case HealthStatusDegraded:
		return AlertSeverityWarning
	case HealthStatusUnhealthy:
		return AlertSeverityError
	default:
		return AlertSeverityCritical
	}
}
