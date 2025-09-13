package models

import (
	"database/sql"
	"time"
)

// Database Connection Management Models

// DatabaseConfig represents database connection configuration
type DatabaseConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Database        string        `json:"database"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	SSLMode         string        `json:"ssl_mode"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" swaggertype:"string" format:"duration" example:"30m"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" swaggertype:"string" format:"duration" example:"10m"`
	Timeout         time.Duration `json:"timeout" swaggertype:"string" format:"duration" example:"30s"`
	ReadTimeout     time.Duration `json:"read_timeout" swaggertype:"string" format:"duration" example:"30s"`
	WriteTimeout    time.Duration `json:"write_timeout" swaggertype:"string" format:"duration" example:"30s"`
}

// DatabaseInstance represents a database instance in the cluster
type DatabaseInstance struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Host        string              `json:"host"`
	Port        int                 `json:"port"`
	Role        DatabaseRole        `json:"role"` // "primary", "replica", "standby"
	Status      DatabaseStatus      `json:"status"`
	Health      DatabaseHealthCheck `json:"health"`
	Config      *DatabaseConfig     `json:"config"`
	LastChecked time.Time           `json:"last_checked"`
	Metadata    map[string]string   `json:"metadata"`
}

// DatabaseRole represents the role of a database instance
type DatabaseRole string

const (
	DatabaseRolePrimary DatabaseRole = "primary"
	DatabaseRoleReplica DatabaseRole = "replica"
	DatabaseRoleStandby DatabaseRole = "standby"
)

// DatabaseStatus represents the status of a database instance
type DatabaseStatus string

const (
	DatabaseStatusHealthy     DatabaseStatus = "healthy"
	DatabaseStatusUnhealthy   DatabaseStatus = "unhealthy"
	DatabaseStatusMaintenance DatabaseStatus = "maintenance"
	DatabaseStatusUnknown     DatabaseStatus = "unknown"
)

// DatabaseHealthCheck represents health check information
type DatabaseHealthCheck struct {
	IsConnected      bool          `json:"is_connected"`
	ResponseTime     time.Duration `json:"response_time" swaggertype:"string" format:"duration" example:"10ms"`
	ActiveConns      int           `json:"active_connections"`
	IdleConns        int           `json:"idle_connections"`
	WaitingConns     int           `json:"waiting_connections"`
	LastError        string        `json:"last_error,omitempty"`
	CheckedAt        time.Time     `json:"checked_at"`
	ConsecutiveFails int           `json:"consecutive_fails"`
}

// ConnectionPool represents a database connection pool
type ConnectionPool struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	DatabaseID      string            `json:"database_id"`
	DB              *sql.DB           `json:"-"`
	Config          *DatabaseConfig   `json:"config"`
	Stats           *PoolStats        `json:"stats"`
	CreatedAt       time.Time         `json:"created_at"`
	LastHealthCheck time.Time         `json:"last_health_check"`
	Metadata        map[string]string `json:"metadata"`
}

// PoolStats represents connection pool statistics
type PoolStats struct {
	MaxOpenConnections int           `json:"max_open_connections"`
	OpenConnections    int           `json:"open_connections"`
	InUse              int           `json:"in_use"`
	Idle               int           `json:"idle"`
	WaitCount          int64         `json:"wait_count"`
	WaitDuration       time.Duration `json:"wait_duration" swaggertype:"string" format:"duration" example:"1ms"`
	MaxIdleClosed      int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed  int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed  int64         `json:"max_lifetime_closed"`
}

// DatabaseCluster represents a database cluster with primary/replica setup
type DatabaseCluster struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Primary        *DatabaseInstance   `json:"primary"`
	Replicas       []*DatabaseInstance `json:"replicas"`
	LoadBalancer   *LoadBalancer       `json:"load_balancer"`
	FailoverConfig *FailoverConfig     `json:"failover_config"`
	Health         *ClusterHealth      `json:"health"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// LoadBalancer represents database load balancing configuration
type LoadBalancer struct {
	Strategy            LoadBalancingStrategy `json:"strategy"`
	ReadWriteSplit      bool                  `json:"read_write_split"`
	ReplicaWeight       map[string]int        `json:"replica_weight"`
	HealthCheckInterval time.Duration         `json:"health_check_interval" swaggertype:"string" format:"duration" example:"30s"`
	MaxRetries          int                   `json:"max_retries"`
}

// FailoverConfig represents failover configuration
type FailoverConfig struct {
	Enabled               bool          `json:"enabled"`
	AutoFailover          bool          `json:"auto_failover"`
	FailoverTimeout       time.Duration `json:"failover_timeout" swaggertype:"string" format:"duration" example:"5m"`
	MaxFailoverAttempts   int           `json:"max_failover_attempts"`
	HealthCheckInterval   time.Duration `json:"health_check_interval" swaggertype:"string" format:"duration" example:"30s"`
	FailureThreshold      int           `json:"failure_threshold"`
	RecoveryCheckInterval time.Duration `json:"recovery_check_interval" swaggertype:"string" format:"duration" example:"60s"`
}

// ClusterHealth represents the overall health of a database cluster
type ClusterHealth struct {
	Status           DatabaseStatus `json:"status"`
	PrimaryHealthy   bool           `json:"primary_healthy"`
	ReplicasHealthy  int            `json:"replicas_healthy"`
	TotalReplicas    int            `json:"total_replicas"`
	FailoverActive   bool           `json:"failover_active"`
	LastFailover     *time.Time     `json:"last_failover,omitempty"`
	ConnectionPools  int            `json:"connection_pools"`
	TotalConnections int            `json:"total_connections"`
	CheckedAt        time.Time      `json:"checked_at"`
}

// DatabaseMetrics represents database performance metrics
type DatabaseMetrics struct {
	TotalQueries        int64         `json:"total_queries"`
	ReadQueries         int64         `json:"read_queries"`
	WriteQueries        int64         `json:"write_queries"`
	FailedQueries       int64         `json:"failed_queries"`
	AverageResponseTime time.Duration `json:"average_response_time" swaggertype:"string" format:"duration" example:"5ms"`
	SlowQueries         int64         `json:"slow_queries"`
	ConnectionErrors    int64         `json:"connection_errors"`
	PoolExhaustionCount int64         `json:"pool_exhaustion_count"`
	FailoverCount       int64         `json:"failover_count"`
	ReplicationLag      time.Duration `json:"replication_lag" swaggertype:"string" format:"duration" example:"100ms"`
	ActiveConnections   int           `json:"active_connections"`
	IdleConnections     int           `json:"idle_connections"`
	StartTime           time.Time     `json:"start_time"`
	Uptime              time.Duration `json:"uptime" swaggertype:"string" format:"duration" example:"72h30m"`
}

// QueryStats represents statistics for database queries
type QueryStats struct {
	QueryType       string        `json:"query_type"` // SELECT, INSERT, UPDATE, DELETE
	Count           int64         `json:"count"`
	TotalDuration   time.Duration `json:"total_duration" swaggertype:"string" format:"duration" example:"500ms"`
	AverageDuration time.Duration `json:"average_duration" swaggertype:"string" format:"duration" example:"5ms"`
	MinDuration     time.Duration `json:"min_duration" swaggertype:"string" format:"duration" example:"1ms"`
	MaxDuration     time.Duration `json:"max_duration" swaggertype:"string" format:"duration" example:"100ms"`
	Errors          int64         `json:"errors"`
}

// Transaction represents a database transaction
type Transaction struct {
	ID        string            `json:"id"`
	StartTime time.Time         `json:"start_time"`
	EndTime   *time.Time        `json:"end_time,omitempty"`
	Duration  time.Duration     `json:"duration" swaggertype:"string" format:"duration" example:"250ms"`
	Status    TransactionStatus `json:"status"`
	Queries   []QueryInfo       `json:"queries"`
	Error     string            `json:"error,omitempty"`
}

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusActive    TransactionStatus = "active"
	TransactionStatusCommitted TransactionStatus = "committed"
	TransactionStatusAborted   TransactionStatus = "aborted"
)

// QueryInfo represents information about a query in a transaction
type QueryInfo struct {
	SQL      string        `json:"sql"`
	Args     []interface{} `json:"args,omitempty"`
	Duration time.Duration `json:"duration" swaggertype:"string" format:"duration" example:"10ms"`
	Error    string        `json:"error,omitempty"`
}

// DatabaseMigration represents a database migration
type DatabaseMigration struct {
	Version     string     `json:"version"`
	Description string     `json:"description"`
	Applied     bool       `json:"applied"`
	AppliedAt   *time.Time `json:"applied_at,omitempty"`
	Checksum    string     `json:"checksum"`
}

// DatabaseBackup represents a database backup
type DatabaseBackup struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	DatabaseID  string                 `json:"database_id"`
	Type        BackupType             `json:"type"`
	Status      BackupStatus           `json:"status"`
	Size        int64                  `json:"size"`
	Compressed  bool                   `json:"compressed"`
	Encrypted   bool                   `json:"encrypted"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Location    string                 `json:"location"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BackupType represents the type of backup
type BackupType string

const (
	BackupTypeFull         BackupType = "full"
	BackupTypeIncremental  BackupType = "incremental"
	BackupTypeDifferential BackupType = "differential"
)

// BackupStatus represents the status of a backup
type BackupStatus string

const (
	BackupStatusRunning   BackupStatus = "running"
	BackupStatusCompleted BackupStatus = "completed"
	BackupStatusFailed    BackupStatus = "failed"
	BackupStatusExpired   BackupStatus = "expired"
)

// Methods for calculating derived values

// CalculateSuccessRate calculates the success rate for database operations
func (dm *DatabaseMetrics) CalculateSuccessRate() float64 {
	if dm.TotalQueries == 0 {
		return 0.0
	}
	successQueries := dm.TotalQueries - dm.FailedQueries
	return float64(successQueries) / float64(dm.TotalQueries) * 100.0
}

// CalculateErrorRate calculates the error rate for database operations
func (dm *DatabaseMetrics) CalculateErrorRate() float64 {
	if dm.TotalQueries == 0 {
		return 0.0
	}
	return float64(dm.FailedQueries) / float64(dm.TotalQueries) * 100.0
}

// IsHealthy returns true if the database instance is healthy
func (di *DatabaseInstance) IsHealthy() bool {
	return di.Status == DatabaseStatusHealthy && di.Health.IsConnected
}

// IsPrimary returns true if the database instance is primary
func (di *DatabaseInstance) IsPrimary() bool {
	return di.Role == DatabaseRolePrimary
}

// IsReplica returns true if the database instance is a replica
func (di *DatabaseInstance) IsReplica() bool {
	return di.Role == DatabaseRoleReplica
}

// GetUtilization calculates connection pool utilization percentage
func (ps *PoolStats) GetUtilization() float64 {
	if ps.MaxOpenConnections == 0 {
		return 0.0
	}
	return float64(ps.InUse) / float64(ps.MaxOpenConnections) * 100.0
}

// String methods for enums

func (dr DatabaseRole) String() string {
	return string(dr)
}

func (ds DatabaseStatus) String() string {
	return string(ds)
}

func (ts TransactionStatus) String() string {
	return string(ts)
}

func (bt BackupType) String() string {
	return string(bt)
}

func (bs BackupStatus) String() string {
	return string(bs)
}
