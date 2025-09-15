package config

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DatabasePerformanceConfig contains optimized database performance settings
type DatabasePerformanceConfig struct {
	// Connection Pool Settings
	MaxOpenConnections    int           `json:"max_open_connections" env:"DB_MAX_OPEN_CONNECTIONS" default:"25"`
	MaxIdleConnections    int           `json:"max_idle_connections" env:"DB_MAX_IDLE_CONNECTIONS" default:"10"`
	ConnectionMaxLifetime time.Duration `json:"connection_max_lifetime" env:"DB_CONNECTION_MAX_LIFETIME" default:"1h"`
	ConnectionMaxIdleTime time.Duration `json:"connection_max_idle_time" env:"DB_CONNECTION_MAX_IDLE_TIME" default:"15m"`

	// Query Performance Settings
	QueryTimeout          time.Duration `json:"query_timeout" env:"DB_QUERY_TIMEOUT" default:"30s"`
	SlowQueryThreshold    time.Duration `json:"slow_query_threshold" env:"DB_SLOW_QUERY_THRESHOLD" default:"1s"`
	EnableQueryLogging    bool          `json:"enable_query_logging" env:"DB_ENABLE_QUERY_LOGGING" default:"false"`
	EnableQueryStats      bool          `json:"enable_query_stats" env:"DB_ENABLE_QUERY_STATS" default:"true"`

	// Optimization Settings
	EnablePreparedStmts   bool `json:"enable_prepared_statements" env:"DB_ENABLE_PREPARED_STMTS" default:"true"`
	EnableBatchInserts    bool `json:"enable_batch_inserts" env:"DB_ENABLE_BATCH_INSERTS" default:"true"`
	BatchSize             int  `json:"batch_size" env:"DB_BATCH_SIZE" default:"100"`
	EnableAsyncLogging    bool `json:"enable_async_logging" env:"DB_ENABLE_ASYNC_LOGGING" default:"true"`

	// Monitoring Settings
	EnableConnectionMetrics bool `json:"enable_connection_metrics" env:"DB_ENABLE_CONNECTION_METRICS" default:"true"`
	MetricsRefreshInterval  time.Duration `json:"metrics_refresh_interval" env:"DB_METRICS_REFRESH_INTERVAL" default:"5m"`
}

// OptimizedDatabase wraps sqlx.DB with performance optimizations
type OptimizedDatabase struct {
	*sqlx.DB
	config *DatabasePerformanceConfig
	stats  *DatabaseStats
}

// DatabaseStats tracks database performance metrics
type DatabaseStats struct {
	TotalQueries       int64         `json:"total_queries"`
	SlowQueries        int64         `json:"slow_queries"`
	FailedQueries      int64         `json:"failed_queries"`
	AverageQueryTime   time.Duration `json:"average_query_time"`
	ConnectionsOpen    int           `json:"connections_open"`
	ConnectionsInUse   int           `json:"connections_in_use"`
	ConnectionsIdle    int           `json:"connections_idle"`
	LastStatsUpdate    time.Time     `json:"last_stats_update"`
}

// NewOptimizedDatabase creates a new optimized database connection
func NewOptimizedDatabase(databaseURL string, config *DatabasePerformanceConfig) (*OptimizedDatabase, error) {
	if config == nil {
		config = &DatabasePerformanceConfig{
			MaxOpenConnections:      25,
			MaxIdleConnections:      10,
			ConnectionMaxLifetime:   time.Hour,
			ConnectionMaxIdleTime:   15 * time.Minute,
			QueryTimeout:            30 * time.Second,
			SlowQueryThreshold:      time.Second,
			EnableQueryLogging:      false,
			EnableQueryStats:        true,
			EnablePreparedStmts:     true,
			EnableBatchInserts:      true,
			BatchSize:               100,
			EnableAsyncLogging:      true,
			EnableConnectionMetrics: true,
			MetricsRefreshInterval:  5 * time.Minute,
		}
	}

	// Open database connection
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Apply performance optimizations
	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetConnMaxLifetime(config.ConnectionMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnectionMaxIdleTime)

	// Create optimized database wrapper
	optimizedDB := &OptimizedDatabase{
		DB:     db,
		config: config,
		stats:  &DatabaseStats{},
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), config.QueryTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Apply PostgreSQL-specific optimizations
	if err := optimizedDB.applyPostgreSQLOptimizations(ctx); err != nil {
		return nil, fmt.Errorf("failed to apply PostgreSQL optimizations: %w", err)
	}

	// Start metrics collection if enabled
	if config.EnableConnectionMetrics {
		go optimizedDB.startMetricsCollection()
	}

	return optimizedDB, nil
}

// applyPostgreSQLOptimizations applies PostgreSQL-specific performance settings
func (db *OptimizedDatabase) applyPostgreSQLOptimizations(ctx context.Context) error {
	optimizations := []string{
		// Enable query plan caching
		"SET plan_cache_mode = 'auto'",

		// Optimize work memory for sorting and hashing
		"SET work_mem = '256MB'",

		// Optimize maintenance work memory
		"SET maintenance_work_mem = '512MB'",

		// Enable parallel queries
		"SET max_parallel_workers_per_gather = 4",

		// Optimize random page cost for SSD
		"SET random_page_cost = 1.1",

		// Enable query optimization
		"SET enable_hashjoin = on",
		"SET enable_mergejoin = on",
		"SET enable_nestloop = on",

		// Optimize checkpoint settings
		"SET checkpoint_completion_target = 0.9",

		// Enable query statistics
		"SET track_activities = on",
		"SET track_counts = on",
		"SET track_io_timing = on",
		"SET track_functions = 'all'",

		// Optimize autovacuum for audit logs
		"SET autovacuum_analyze_scale_factor = 0.05",
		"SET autovacuum_vacuum_scale_factor = 0.1",
	}

	for _, optimization := range optimizations {
		if _, err := db.ExecContext(ctx, optimization); err != nil {
			// Log warning but don't fail - some settings may require superuser
			continue
		}
	}

	return nil
}

// startMetricsCollection starts background metrics collection
func (db *OptimizedDatabase) startMetricsCollection() {
	ticker := time.NewTicker(db.config.MetricsRefreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		db.updateStats()
	}
}

// updateStats updates database performance statistics
func (db *OptimizedDatabase) updateStats() {
	dbStats := db.Stats()

	db.stats.ConnectionsOpen = dbStats.OpenConnections
	db.stats.ConnectionsInUse = dbStats.InUse
	db.stats.ConnectionsIdle = dbStats.Idle
	db.stats.LastStatsUpdate = time.Now()
}

// ExecWithStats executes a query and tracks performance metrics
func (db *OptimizedDatabase) ExecWithStats(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()

	// Apply query timeout
	queryCtx, cancel := context.WithTimeout(ctx, db.config.QueryTimeout)
	defer cancel()

	result, err := db.ExecContext(queryCtx, query, args...)

	// Update statistics
	duration := time.Since(start)
	db.updateQueryStats(duration, err)

	// Log slow queries if enabled
	if db.config.EnableQueryLogging && duration > db.config.SlowQueryThreshold {
		// Log slow query (implement logger as needed)
		fmt.Printf("SLOW QUERY [%v]: %s\n", duration, query)
	}

	return result, err
}

// QueryWithStats executes a query and tracks performance metrics
func (db *OptimizedDatabase) QueryWithStats(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()

	// Apply query timeout
	queryCtx, cancel := context.WithTimeout(ctx, db.config.QueryTimeout)
	defer cancel()

	rows, err := db.QueryContext(queryCtx, query, args...)

	// Update statistics
	duration := time.Since(start)
	db.updateQueryStats(duration, err)

	// Log slow queries if enabled
	if db.config.EnableQueryLogging && duration > db.config.SlowQueryThreshold {
		fmt.Printf("SLOW QUERY [%v]: %s\n", duration, query)
	}

	return rows, err
}

// QueryRowWithStats executes a query row and tracks performance metrics
func (db *OptimizedDatabase) QueryRowWithStats(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()

	// Apply query timeout
	queryCtx, cancel := context.WithTimeout(ctx, db.config.QueryTimeout)
	defer cancel()

	row := db.QueryRowContext(queryCtx, query, args...)

	// Update statistics
	duration := time.Since(start)
	db.updateQueryStats(duration, nil) // QueryRow doesn't return error until Scan

	// Log slow queries if enabled
	if db.config.EnableQueryLogging && duration > db.config.SlowQueryThreshold {
		fmt.Printf("SLOW QUERY [%v]: %s\n", duration, query)
	}

	return row
}

// updateQueryStats updates query performance statistics
func (db *OptimizedDatabase) updateQueryStats(duration time.Duration, err error) {
	if !db.config.EnableQueryStats {
		return
	}

	db.stats.TotalQueries++

	if err != nil {
		db.stats.FailedQueries++
	}

	if duration > db.config.SlowQueryThreshold {
		db.stats.SlowQueries++
	}

	// Update average query time
	if db.stats.TotalQueries == 1 {
		db.stats.AverageQueryTime = duration
	} else {
		// Calculate rolling average
		totalTime := db.stats.AverageQueryTime * time.Duration(db.stats.TotalQueries-1)
		db.stats.AverageQueryTime = (totalTime + duration) / time.Duration(db.stats.TotalQueries)
	}
}

// GetStats returns current database performance statistics
func (db *OptimizedDatabase) GetStats() *DatabaseStats {
	db.updateStats()
	return db.stats
}

// GetConfig returns the database performance configuration
func (db *OptimizedDatabase) GetConfig() *DatabasePerformanceConfig {
	return db.config
}

// BatchExecute executes multiple queries in a single transaction for better performance
func (db *OptimizedDatabase) BatchExecute(ctx context.Context, queries []string, argsList [][]interface{}) error {
	if !db.config.EnableBatchInserts || len(queries) == 0 {
		return fmt.Errorf("batch execution not enabled or no queries provided")
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for i, query := range queries {
		var args []interface{}
		if i < len(argsList) {
			args = argsList[i]
		}

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("failed to execute batch query %d: %w", i, err)
		}
	}

	return tx.Commit()
}

// OptimizeTable runs ANALYZE on the specified table to update statistics
func (db *OptimizedDatabase) OptimizeTable(ctx context.Context, tableName string) error {
	query := fmt.Sprintf("ANALYZE %s", tableName)

	_, err := db.ExecWithStats(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to analyze table %s: %w", tableName, err)
	}

	return nil
}

// RefreshMaterializedView refreshes a materialized view
func (db *OptimizedDatabase) RefreshMaterializedView(ctx context.Context, viewName string, concurrently bool) error {
	var query string
	if concurrently {
		query = fmt.Sprintf("REFRESH MATERIALIZED VIEW CONCURRENTLY %s", viewName)
	} else {
		query = fmt.Sprintf("REFRESH MATERIALIZED VIEW %s", viewName)
	}

	_, err := db.ExecWithStats(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to refresh materialized view %s: %w", viewName, err)
	}

	return nil
}

// Close closes the database connection and stops metrics collection
func (db *OptimizedDatabase) Close() error {
	return db.DB.Close()
}

// HealthCheck performs a comprehensive database health check
func (db *OptimizedDatabase) HealthCheck(ctx context.Context) (*DatabaseHealthStatus, error) {
	status := &DatabaseHealthStatus{
		Timestamp: time.Now(),
	}

	// Test basic connectivity
	start := time.Now()
	err := db.PingContext(ctx)
	status.PingLatency = time.Since(start)
	status.IsConnected = err == nil

	if !status.IsConnected {
		status.Issues = append(status.Issues, fmt.Sprintf("Database connection failed: %v", err))
		return status, nil
	}

	// Check connection pool health
	dbStats := db.Stats()
	status.ConnectionPoolStats = DatabaseConnectionStats{
		Open:    dbStats.OpenConnections,
		InUse:   dbStats.InUse,
		Idle:    dbStats.Idle,
		WaitCount: dbStats.WaitCount,
		WaitDuration: dbStats.WaitDuration,
	}

	// Check for connection pool issues
	maxConns := db.config.MaxOpenConnections
	if dbStats.OpenConnections > int(float64(maxConns)*0.8) {
		status.Issues = append(status.Issues, "Connection pool usage high (>80%)")
	}

	if dbStats.WaitCount > 100 {
		status.Issues = append(status.Issues, "High connection wait count detected")
	}

	// Check query performance
	if db.stats.SlowQueries > 0 {
		slowQueryRate := float64(db.stats.SlowQueries) / float64(db.stats.TotalQueries) * 100
		if slowQueryRate > 10 {
			status.Issues = append(status.Issues, fmt.Sprintf("High slow query rate: %.2f%%", slowQueryRate))
		}
	}

	// Check failed query rate
	if db.stats.FailedQueries > 0 {
		failedQueryRate := float64(db.stats.FailedQueries) / float64(db.stats.TotalQueries) * 100
		if failedQueryRate > 5 {
			status.Issues = append(status.Issues, fmt.Sprintf("High failed query rate: %.2f%%", failedQueryRate))
		}
	}

	status.IsHealthy = len(status.Issues) == 0
	return status, nil
}

// DatabaseHealthStatus represents the health status of the database
type DatabaseHealthStatus struct {
	Timestamp           time.Time                 `json:"timestamp"`
	IsConnected         bool                      `json:"is_connected"`
	IsHealthy           bool                      `json:"is_healthy"`
	PingLatency         time.Duration             `json:"ping_latency"`
	ConnectionPoolStats DatabaseConnectionStats   `json:"connection_pool_stats"`
	Issues              []string                  `json:"issues"`
}

// DatabaseConnectionStats represents connection pool statistics
type DatabaseConnectionStats struct {
	Open         int           `json:"open"`
	InUse        int           `json:"in_use"`
	Idle         int           `json:"idle"`
	WaitCount    int64         `json:"wait_count"`
	WaitDuration time.Duration `json:"wait_duration"`
}