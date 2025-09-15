package external

import "time"

// HealthStatus represents the health status as a string
type HealthStatusString string

const (
	HealthStatusHealthy   HealthStatusString = "healthy"
	HealthStatusUnhealthy HealthStatusString = "unhealthy"
	HealthStatusUnknown   HealthStatusString = "unknown"
)

// HealthStatusStruct represents the health status with detailed information
type HealthStatusStruct struct {
	Status       HealthStatusString `json:"status"`
	Message      string             `json:"message,omitempty"`
	LastChecked  time.Time          `json:"last_checked"`
	ResponseTime time.Duration      `json:"response_time"`
	Details      map[string]any     `json:"details,omitempty"`
}

// AlertThresholds defines thresholds for alerting
type AlertThresholds struct {
	ErrorRateThreshold    float64       `json:"error_rate_threshold"`
	ResponseTimeThreshold time.Duration `json:"response_time_threshold"`
	ConsecutiveFailures   int           `json:"consecutive_failures"`
	TimeWindow            time.Duration `json:"time_window"`
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidatorConfig represents configuration for validators
type ValidatorConfig struct {
	EnableStrictMode   bool              `json:"enable_strict_mode"`
	MaxValidationTime  time.Duration     `json:"max_validation_time"`
	CustomRules        map[string]string `json:"custom_rules,omitempty"`
	EnableCaching      bool              `json:"enable_caching"`
	CacheExpiration    time.Duration     `json:"cache_expiration"`
	RateLimitPerMinute int               `json:"rate_limit_per_minute"`
}
