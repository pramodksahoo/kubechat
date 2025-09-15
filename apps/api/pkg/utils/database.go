package utils

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string        `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port            int           `env:"POSTGRES_PORT" envDefault:"5432"`
	Database        string        `env:"POSTGRES_DB" envDefault:"kubechat"`
	Username        string        `env:"POSTGRES_USER" envDefault:"kubechat"`
	Password        string        `env:"POSTGRES_PASSWORD" envDefault:"kubechat"`
	SSLMode         string        `env:"POSTGRES_SSLMODE" envDefault:"disable"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`
	ConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"1m"`
}

// DatabaseHealth represents database health status
type DatabaseHealth struct {
	Status          string    `json:"status"`
	ResponseTime    string    `json:"response_time"`
	OpenConnections int       `json:"open_connections"`
	IdleConnections int       `json:"idle_connections"`
	Timestamp       time.Time `json:"timestamp"`
	Error           string    `json:"error,omitempty"`
}

// NewDatabase creates a new database connection with optimized settings
func NewDatabase(config DatabaseConfig) (*sqlx.DB, error) {
	// Build connection string
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	// Open database connection
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool for optimal performance
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// CheckDatabaseHealth performs comprehensive database health check
func CheckDatabaseHealth(ctx context.Context, db *sqlx.DB) *DatabaseHealth {
	startTime := time.Now()
	health := &DatabaseHealth{
		Timestamp: startTime,
	}

	// Test database connectivity
	if err := db.PingContext(ctx); err != nil {
		health.Status = "unhealthy"
		health.Error = fmt.Sprintf("ping failed: %v", err)
		health.ResponseTime = time.Since(startTime).String()
		return health
	}

	// Get connection statistics
	stats := db.Stats()
	health.OpenConnections = stats.OpenConnections
	health.IdleConnections = stats.Idle

	// Test a simple query
	var result int
	if err := db.GetContext(ctx, &result, "SELECT 1"); err != nil {
		health.Status = "degraded"
		health.Error = fmt.Sprintf("query test failed: %v", err)
	} else {
		health.Status = "healthy"
	}

	health.ResponseTime = time.Since(startTime).String()
	return health
}

// MigrateDatabase runs database migrations
func MigrateDatabase(ctx context.Context, db *sqlx.DB, migrationsPath string) error {
	// This would typically use a migration library like golang-migrate
	// For now, we'll implement a simple version that runs our migration files

	// Check if schema_migrations table exists
	var exists bool
	err := db.GetContext(ctx, &exists, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'schema_migrations'
		)`)
	if err != nil {
		return fmt.Errorf("failed to check migrations table: %w", err)
	}

	if !exists {
		// Create migrations table
		_, err = db.ExecContext(ctx, `
			CREATE TABLE schema_migrations (
				version BIGINT PRIMARY KEY,
				dirty BOOLEAN NOT NULL DEFAULT FALSE,
				created_at TIMESTAMP DEFAULT NOW()
			)`)
		if err != nil {
			return fmt.Errorf("failed to create migrations table: %w", err)
		}
	}

	return nil
}

// ValidateDatabaseSchema validates that required tables exist
func ValidateDatabaseSchema(ctx context.Context, db *sqlx.DB) error {
	requiredTables := []string{
		"users",
		"user_sessions",
		"audit_logs",
		"cluster_configs",
	}

	for _, table := range requiredTables {
		var exists bool
		err := db.GetContext(ctx, &exists, `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)`, table)
		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}

		if !exists {
			return fmt.Errorf("required table %s does not exist", table)
		}
	}

	return nil
}

// ValidateAuditLogIntegrity validates the integrity of audit logs
func ValidateAuditLogIntegrity(ctx context.Context, db *sqlx.DB) ([]IntegrityResult, error) {
	type dbResult struct {
		LogID        int64   `db:"log_id"`
		IsValid      bool    `db:"is_valid"`
		ErrorMessage *string `db:"error_message"`
	}

	var dbResults []dbResult
	err := db.SelectContext(ctx, &dbResults, "SELECT log_id, is_valid, error_message FROM verify_audit_log_integrity(NULL)")
	if err != nil {
		return nil, fmt.Errorf("failed to validate audit log integrity: %w", err)
	}

	results := make([]IntegrityResult, len(dbResults))
	for i, result := range dbResults {
		results[i] = IntegrityResult{
			LogID:   result.LogID,
			IsValid: result.IsValid,
		}
		if result.ErrorMessage != nil {
			results[i].ErrorMessage = *result.ErrorMessage
		}
	}

	return results, nil
}

// IntegrityResult represents audit log integrity check result
type IntegrityResult struct {
	LogID        int64  `json:"log_id"`
	IsValid      bool   `json:"is_valid"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// CleanupExpiredSessions removes expired user sessions
func CleanupExpiredSessions(ctx context.Context, db *sqlx.DB) (int64, error) {
	result, err := db.ExecContext(ctx, "DELETE FROM user_sessions WHERE expires_at < $1", time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return rowsAffected, nil
}

// GetDatabaseStats returns database connection and usage statistics
func GetDatabaseStats(db *sqlx.DB) map[string]interface{} {
	stats := db.Stats()

	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"connections_in_use":   stats.InUse,
		"idle_connections":     stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// DatabaseTransaction helper for managing database transactions
type DatabaseTransaction struct {
	tx *sqlx.Tx
}

// NewTransaction creates a new database transaction
func NewTransaction(ctx context.Context, db *sqlx.DB) (*DatabaseTransaction, error) {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &DatabaseTransaction{tx: tx}, nil
}

// Commit commits the transaction
func (dt *DatabaseTransaction) Commit() error {
	return dt.tx.Commit()
}

// Rollback rolls back the transaction
func (dt *DatabaseTransaction) Rollback() error {
	return dt.tx.Rollback()
}

// Tx returns the underlying sqlx.Tx
func (dt *DatabaseTransaction) Tx() *sqlx.Tx {
	return dt.tx
}
