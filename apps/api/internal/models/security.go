package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Security and Performance Integration Models

// SecurityConfig represents comprehensive security configuration
type SecurityConfig struct {
	EnableEncryption      bool              `json:"enable_encryption"`
	EnableRateLimiting    bool              `json:"enable_rate_limiting"`
	EnableInputValidation bool              `json:"enable_input_validation"`
	EnableAuditLogging    bool              `json:"enable_audit_logging"`
	EnableSecurityHeaders bool              `json:"enable_security_headers"`
	EnableCSRF            bool              `json:"enable_csrf"`
	EnableContentSecurity bool              `json:"enable_content_security"`
	PasswordPolicy        *PasswordPolicy   `json:"password_policy"`
	SessionConfig         *SessionConfig    `json:"session_config"`
	EncryptionConfig      *EncryptionConfig `json:"encryption_config"`
	RateLimitConfig       *RateLimitConfig  `json:"rate_limit_config"`
	SecurityHeaders       map[string]string `json:"security_headers"`
	AllowedOrigins        []string          `json:"allowed_origins"`
	BlockedIPs            []string          `json:"blocked_ips"`
	TrustedProxies        []string          `json:"trusted_proxies"`
	SecurityEventHandlers map[string]string `json:"security_event_handlers"`
}

// PasswordPolicy represents password security requirements
type PasswordPolicy struct {
	MinLength           int  `json:"min_length"`
	RequireUppercase    bool `json:"require_uppercase"`
	RequireLowercase    bool `json:"require_lowercase"`
	RequireNumbers      bool `json:"require_numbers"`
	RequireSpecialChars bool `json:"require_special_chars"`
	MaxAge              int  `json:"max_age_days"`
	PreventReuse        int  `json:"prevent_reuse_count"`
	MaxAttempts         int  `json:"max_attempts"`
	LockoutDuration     int  `json:"lockout_duration_minutes"`
}

// SessionConfig represents session security configuration
type SessionConfig struct {
	Timeout           time.Duration `json:"timeout" swaggertype:"string" format:"duration" example:"30m"`
	RefreshThreshold  time.Duration `json:"refresh_threshold" swaggertype:"string" format:"duration" example:"5m"`
	SecureCookie      bool          `json:"secure_cookie"`
	HTTPOnlyCookie    bool          `json:"http_only_cookie"`
	SameSiteCookie    string        `json:"same_site_cookie"` // Strict, Lax, None
	CookieDomain      string        `json:"cookie_domain"`
	CookiePath        string        `json:"cookie_path"`
	MaxSessions       int           `json:"max_sessions_per_user"`
	EnableFingerprint bool          `json:"enable_fingerprint"`
}

// EncryptionConfig represents encryption configuration
type EncryptionConfig struct {
	Algorithm           string `json:"algorithm"` // AES-256-GCM, ChaCha20-Poly1305
	KeyRotationInterval string `json:"key_rotation_interval"`
	EnableKeyEscrow     bool   `json:"enable_key_escrow"`
	EncryptionAtRest    bool   `json:"encryption_at_rest"`
	EncryptionInTransit bool   `json:"encryption_in_transit"`
	HashingAlgorithm    string `json:"hashing_algorithm"` // bcrypt, argon2id
	HashingCost         int    `json:"hashing_cost"`
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID          string                 `json:"id"`
	Type        SecurityEventType      `json:"type"`
	Severity    SecuritySeverity       `json:"severity"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Location    *GeoLocation           `json:"location,omitempty"`
	RiskScore   int                    `json:"risk_score"` // 0-100
	Blocked     bool                   `json:"blocked"`
	Action      SecurityAction         `json:"action"`
	Resolution  string                 `json:"resolution,omitempty"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

// SecurityEventType represents types of security events
type SecurityEventType string

const (
	SecurityEventLoginAttempt          SecurityEventType = "login_attempt"
	SecurityEventLoginSuccess          SecurityEventType = "login_success"
	SecurityEventLoginFailure          SecurityEventType = "login_failure"
	SecurityEventLogout                SecurityEventType = "logout"
	SecurityEventPasswordChange        SecurityEventType = "password_change"
	SecurityEventAccountLockout        SecurityEventType = "account_lockout"
	SecurityEventSuspiciousActivity    SecurityEventType = "suspicious_activity"
	SecurityEventRateLimitExceeded     SecurityEventType = "rate_limit_exceeded"
	SecurityEventUnauthorizedAccess    SecurityEventType = "unauthorized_access"
	SecurityEventDataBreach            SecurityEventType = "data_breach"
	SecurityEventInjectionAttempt      SecurityEventType = "injection_attempt"
	SecurityEventBruteForceAttempt     SecurityEventType = "brute_force_attempt"
	SecurityEventCSRFAttempt           SecurityEventType = "csrf_attempt"
	SecurityEventXSSAttempt            SecurityEventType = "xss_attempt"
	SecurityEventSecurityHeaderMissing SecurityEventType = "security_header_missing"
	SecurityEventEncryptionFailure     SecurityEventType = "encryption_failure"
	SecurityEventCertificateExpiry     SecurityEventType = "certificate_expiry"
)

// SecuritySeverity represents the severity level of security events
type SecuritySeverity string

const (
	SecuritySeverityLow      SecuritySeverity = "low"
	SecuritySeverityMedium   SecuritySeverity = "medium"
	SecuritySeverityHigh     SecuritySeverity = "high"
	SecuritySeverityCritical SecuritySeverity = "critical"
)

// SecurityAction represents actions taken in response to security events
type SecurityAction string

const (
	SecurityActionAllow   SecurityAction = "allow"
	SecurityActionBlock   SecurityAction = "block"
	SecurityActionFlag    SecurityAction = "flag"
	SecurityActionAlert   SecurityAction = "alert"
	SecurityActionAudit   SecurityAction = "audit"
	SecurityActionLockout SecurityAction = "lockout"
)

// GeoLocation represents geographical location data
type GeoLocation struct {
	Country    string  `json:"country"`
	Region     string  `json:"region"`
	City       string  `json:"city"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Timezone   string  `json:"timezone"`
	ISP        string  `json:"isp"`
	ASN        string  `json:"asn"`
	VPN        bool    `json:"vpn"`
	TOR        bool    `json:"tor"`
	Proxy      bool    `json:"proxy"`
	Suspicious bool    `json:"suspicious"`
}

// PerformanceConfig represents performance optimization configuration
type PerformanceConfig struct {
	EnableCaching           bool                    `json:"enable_caching"`
	EnableCompression       bool                    `json:"enable_compression"`
	EnableHTTP2             bool                    `json:"enable_http2"`
	EnableConnectionPooling bool                    `json:"enable_connection_pooling"`
	EnableRequestBatching   bool                    `json:"enable_request_batching"`
	EnableResponseStreaming bool                    `json:"enable_response_streaming"`
	CacheConfig             *CacheConfig            `json:"cache_config"`
	CompressionConfig       *CompressionConfig      `json:"compression_config"`
	ConnectionPoolConfig    *ConnectionPoolConfig   `json:"connection_pool_config"`
	PerformanceMonitoring   *PerformanceMonitoring  `json:"performance_monitoring"`
	OptimizationThresholds  *OptimizationThresholds `json:"optimization_thresholds"`
	LoadBalancingConfig     *LoadBalancingConfig    `json:"load_balancing_config"`
}

// CacheConfig represents caching configuration
type CacheConfig struct {
	DefaultTTL        time.Duration       `json:"default_ttl" swaggertype:"string" format:"duration" example:"1h"`
	MaxSize           int64               `json:"max_size_bytes"`
	MaxEntries        int                 `json:"max_entries"`
	EvictionPolicy    string              `json:"eviction_policy"` // LRU, LFU, FIFO
	EnableDistributed bool                `json:"enable_distributed"`
	RedisConfig       *RedisCacheConfig   `json:"redis_config,omitempty"`
	MemcacheConfig    *MemcacheConfig     `json:"memcache_config,omitempty"`
	CacheLayers       []CacheLayer        `json:"cache_layers"`
	WarmupStrategies  []WarmupStrategy    `json:"warmup_strategies"`
	InvalidationRules map[string][]string `json:"invalidation_rules"`
}

// RedisCacheConfig represents Redis-specific cache configuration
type RedisCacheConfig struct {
	Address       string        `json:"address"`
	Password      string        `json:"password"`
	DB            int           `json:"db"`
	PoolSize      int           `json:"pool_size"`
	MinIdleConns  int           `json:"min_idle_conns"`
	MaxRetries    int           `json:"max_retries"`
	RetryDelay    time.Duration `json:"retry_delay" swaggertype:"string" format:"duration" example:"1s"`
	EnableCluster bool          `json:"enable_cluster"`
	ClusterNodes  []string      `json:"cluster_nodes,omitempty"`
}

// MemcacheConfig represents Memcache configuration
type MemcacheConfig struct {
	Servers           []string      `json:"servers"`
	MaxIdleConns      int           `json:"max_idle_conns"`
	Timeout           time.Duration `json:"timeout" swaggertype:"string" format:"duration" example:"30s"`
	EnableCompression bool          `json:"enable_compression"`
}

// CacheLayer represents a layer in multi-tier caching
type CacheLayer struct {
	Name     string        `json:"name"`
	Type     string        `json:"type"` // memory, redis, memcache
	TTL      time.Duration `json:"ttl" swaggertype:"string" format:"duration" example:"1h"`
	MaxSize  int64         `json:"max_size"`
	Priority int           `json:"priority"` // Lower number = higher priority
}

// WarmupStrategy represents cache warmup configuration
type WarmupStrategy struct {
	Name       string   `json:"name"`
	Endpoints  []string `json:"endpoints"`
	Schedule   string   `json:"schedule"` // cron format
	BatchSize  int      `json:"batch_size"`
	Concurrent bool     `json:"concurrent"`
}

// CompressionConfig represents compression configuration
type CompressionConfig struct {
	Algorithm       string   `json:"algorithm"`      // gzip, brotli, zstd
	Level           int      `json:"level"`          // compression level
	MinSize         int      `json:"min_size_bytes"` // minimum size to compress
	EnableStreaming bool     `json:"enable_streaming"`
	MimeTypes       []string `json:"mime_types"`    // types to compress
	ExcludePaths    []string `json:"exclude_paths"` // paths to skip compression
}

// ConnectionPoolConfig represents connection pooling configuration
type ConnectionPoolConfig struct {
	MaxOpenConns        int           `json:"max_open_conns"`
	MaxIdleConns        int           `json:"max_idle_conns"`
	ConnMaxLifetime     time.Duration `json:"conn_max_lifetime" swaggertype:"string" format:"duration" example:"1h"`
	ConnMaxIdleTime     time.Duration `json:"conn_max_idle_time" swaggertype:"string" format:"duration" example:"30m"`
	EnableHealthCheck   bool          `json:"enable_health_check"`
	HealthCheckInterval time.Duration `json:"health_check_interval" swaggertype:"string" format:"duration" example:"30s"`
	RetryAttempts       int           `json:"retry_attempts"`
	RetryDelay          time.Duration `json:"retry_delay" swaggertype:"string" format:"duration" example:"1s"`
}

// PerformanceMonitoring represents performance monitoring configuration
type PerformanceMonitoring struct {
	EnableMetrics      bool          `json:"enable_metrics"`
	EnableTracing      bool          `json:"enable_tracing"`
	EnableProfiling    bool          `json:"enable_profiling"`
	SampleRate         float64       `json:"sample_rate"`
	MetricsInterval    time.Duration `json:"metrics_interval" swaggertype:"string" format:"duration" example:"1m"`
	SlowQueryThreshold time.Duration `json:"slow_query_threshold" swaggertype:"string" format:"duration" example:"2s"`
	MemoryThreshold    float64       `json:"memory_threshold_percent"`
	CPUThreshold       float64       `json:"cpu_threshold_percent"`
	DiskThreshold      float64       `json:"disk_threshold_percent"`
	EnableAlerts       bool          `json:"enable_alerts"`
	AlertEndpoints     []string      `json:"alert_endpoints"`
}

// OptimizationThresholds represents thresholds for automatic optimization
type OptimizationThresholds struct {
	ResponseTimeThreshold time.Duration `json:"response_time_threshold" swaggertype:"string" format:"duration" example:"500ms"`
	ThroughputThreshold   int64         `json:"throughput_threshold_rps"`
	ErrorRateThreshold    float64       `json:"error_rate_threshold_percent"`
	MemoryUsageThreshold  float64       `json:"memory_usage_threshold_percent"`
	CPUUsageThreshold     float64       `json:"cpu_usage_threshold_percent"`
	CacheHitRateThreshold float64       `json:"cache_hit_rate_threshold_percent"`
	AutoScalingEnabled    bool          `json:"auto_scaling_enabled"`
	LoadSheddingEnabled   bool          `json:"load_shedding_enabled"`
	CircuitBreakerEnabled bool          `json:"circuit_breaker_enabled"`
}

// LoadBalancingConfig represents load balancing configuration
type LoadBalancingConfig struct {
	Strategy            LoadBalancingStrategy `json:"strategy"`
	HealthCheckPath     string                `json:"health_check_path"`
	HealthCheckInterval time.Duration         `json:"health_check_interval" swaggertype:"string" format:"duration" example:"30s"`
	MaxRetries          int                   `json:"max_retries"`
	RetryDelay          time.Duration         `json:"retry_delay" swaggertype:"string" format:"duration" example:"1s"`
	StickySession       bool                  `json:"sticky_session"`
	SessionAffinityTTL  time.Duration         `json:"session_affinity_ttl" swaggertype:"string" format:"duration" example:"1h"`
	WeightedTargets     map[string]int        `json:"weighted_targets,omitempty"`
}

// PerformanceMetrics represents performance metrics
type PerformanceMetrics struct {
	RequestCount     int64                   `json:"request_count"`
	ResponseTime     PerformanceStats        `json:"response_time"`
	Throughput       float64                 `json:"throughput_rps"`
	ErrorRate        float64                 `json:"error_rate_percent"`
	CacheHitRate     float64                 `json:"cache_hit_rate_percent"`
	CacheMissRate    float64                 `json:"cache_miss_rate_percent"`
	CompressionRatio float64                 `json:"compression_ratio"`
	MemoryUsage      SystemResourceUsage     `json:"memory_usage"`
	CPUUsage         SystemResourceUsage     `json:"cpu_usage"`
	DiskUsage        SystemResourceUsage     `json:"disk_usage"`
	NetworkUsage     NetworkUsage            `json:"network_usage"`
	DatabaseMetrics  DatabasePerformance     `json:"database_metrics"`
	ServiceMetrics   map[string]ServiceStats `json:"service_metrics"`
	SecurityMetrics  SecurityMetrics         `json:"security_metrics"`
	StartTime        time.Time               `json:"start_time"`
	Uptime           time.Duration           `json:"uptime" swaggertype:"string" format:"duration" example:"24h"`
	LastUpdate       time.Time               `json:"last_update"`
}

// PerformanceStats represents statistical performance data
type PerformanceStats struct {
	Min         time.Duration `json:"min" swaggertype:"string" format:"duration" example:"1ms"`
	Max         time.Duration `json:"max" swaggertype:"string" format:"duration" example:"5s"`
	Mean        time.Duration `json:"mean" swaggertype:"string" format:"duration" example:"100ms"`
	Median      time.Duration `json:"median" swaggertype:"string" format:"duration" example:"50ms"`
	P95         time.Duration `json:"p95" swaggertype:"string" format:"duration" example:"300ms"`
	P99         time.Duration `json:"p99" swaggertype:"string" format:"duration" example:"1s"`
	StdDev      time.Duration `json:"std_dev" swaggertype:"string" format:"duration" example:"50ms"`
	TotalTime   time.Duration `json:"total_time" swaggertype:"string" format:"duration" example:"10m"`
	SampleCount int64         `json:"sample_count"`
}

// SystemResourceUsage represents system resource usage
type SystemResourceUsage struct {
	Current    float64   `json:"current_percent"`
	Peak       float64   `json:"peak_percent"`
	Average    float64   `json:"average_percent"`
	Allocated  int64     `json:"allocated_bytes"`
	Used       int64     `json:"used_bytes"`
	Available  int64     `json:"available_bytes"`
	LastUpdate time.Time `json:"last_update"`
}

// NetworkUsage represents network usage statistics
type NetworkUsage struct {
	BytesSent         int64     `json:"bytes_sent"`
	BytesReceived     int64     `json:"bytes_received"`
	PacketsSent       int64     `json:"packets_sent"`
	PacketsReceived   int64     `json:"packets_received"`
	ConnectionsActive int       `json:"connections_active"`
	ConnectionsTotal  int64     `json:"connections_total"`
	Bandwidth         float64   `json:"bandwidth_mbps"`
	LastUpdate        time.Time `json:"last_update"`
}

// DatabasePerformance represents database performance metrics
type DatabasePerformance struct {
	QueryCount          int64                 `json:"query_count"`
	SlowQueryCount      int64                 `json:"slow_query_count"`
	ConnectionCount     int                   `json:"connection_count"`
	ConnectionPoolUsage float64               `json:"connection_pool_usage_percent"`
	TransactionCount    int64                 `json:"transaction_count"`
	DeadlockCount       int64                 `json:"deadlock_count"`
	CacheHitRate        float64               `json:"cache_hit_rate_percent"`
	ReplicationLag      time.Duration         `json:"replication_lag" swaggertype:"string" format:"duration" example:"100ms"`
	QueryStats          map[string]QueryStats `json:"query_stats"`
}

// ServiceStats represents per-service performance statistics
type ServiceStats struct {
	RequestCount    int64            `json:"request_count"`
	ErrorCount      int64            `json:"error_count"`
	ResponseTime    PerformanceStats `json:"response_time"`
	Availability    float64          `json:"availability_percent"`
	Status          string           `json:"status"`
	LastHealthCheck time.Time        `json:"last_health_check"`
}

// SecurityMetrics represents security-related metrics
type SecurityMetrics struct {
	TotalSecurityEvents      int64                       `json:"total_security_events"`
	BlockedAttempts          int64                       `json:"blocked_attempts"`
	SecurityEventsByType     map[SecurityEventType]int64 `json:"security_events_by_type"`
	SecurityEventsBySeverity map[SecuritySeverity]int64  `json:"security_events_by_severity"`
	FailedLoginAttempts      int64                       `json:"failed_login_attempts"`
	SuccessfulLogins         int64                       `json:"successful_logins"`
	RateLimitViolations      int64                       `json:"rate_limit_violations"`
	SuspiciousActivities     int64                       `json:"suspicious_activities"`
	ActiveSessions           int64                       `json:"active_sessions"`
	ExpiredSessions          int64                       `json:"expired_sessions"`
	PasswordPolicyViolations int64                       `json:"password_policy_violations"`
	LastSecurityScan         time.Time                   `json:"last_security_scan"`
	VulnerabilityCount       int                         `json:"vulnerability_count"`
}

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    SecuritySeverity       `json:"severity"`
	Category    string                 `json:"category"`
	Source      string                 `json:"source"`
	Events      []SecurityEvent        `json:"events"`
	Status      string                 `json:"status"`     // open, investigating, resolved
	Priority    int                    `json:"priority"`   // 1-5
	RiskScore   int                    `json:"risk_score"` // 0-100
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Assignee    string                 `json:"assignee,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
}

// Helper methods

// GenerateSecurityToken generates a cryptographically secure random token
func GenerateSecurityToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CalculateRiskScore calculates a risk score based on security event
func (se *SecurityEvent) CalculateRiskScore() int {
	score := 0

	// Base score by event type
	switch se.Type {
	case SecurityEventDataBreach, SecurityEventUnauthorizedAccess:
		score += 40
	case SecurityEventBruteForceAttempt, SecurityEventInjectionAttempt:
		score += 30
	case SecurityEventSuspiciousActivity, SecurityEventRateLimitExceeded:
		score += 20
	case SecurityEventLoginFailure:
		score += 10
	default:
		score += 5
	}

	// Severity multiplier
	switch se.Severity {
	case SecuritySeverityCritical:
		score *= 2
	case SecuritySeverityHigh:
		score = int(float64(score) * 1.5)
	case SecuritySeverityMedium:
		// no change
	case SecuritySeverityLow:
		score = int(float64(score) * 0.7)
	}

	// Location risk (basic check for suspicious locations)
	if se.Location != nil && (se.Location.VPN || se.Location.TOR || se.Location.Proxy) {
		score += 15
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// IsHighRisk determines if a security event is high risk
func (se *SecurityEvent) IsHighRisk() bool {
	return se.RiskScore >= 70
}

// String methods for enums

func (set SecurityEventType) String() string {
	return string(set)
}

func (ss SecuritySeverity) String() string {
	return string(ss)
}

func (sa SecurityAction) String() string {
	return string(sa)
}

// CalculateAvailability calculates service availability percentage
func (ss *ServiceStats) CalculateAvailability() float64 {
	if ss.RequestCount == 0 {
		return 100.0
	}
	successCount := ss.RequestCount - ss.ErrorCount
	return float64(successCount) / float64(ss.RequestCount) * 100.0
}

// GetCacheEfficiency calculates cache efficiency
func (pm *PerformanceMetrics) GetCacheEfficiency() float64 {
	totalCacheRequests := pm.CacheHitRate + pm.CacheMissRate
	if totalCacheRequests == 0 {
		return 0.0
	}
	return (pm.CacheHitRate / totalCacheRequests) * 100.0
}
