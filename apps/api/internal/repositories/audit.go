package repositories

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// AuditRepository handles database operations for audit logs
type AuditRepository interface {
	// CreateAuditLog creates a new audit log entry
	CreateAuditLog(ctx context.Context, auditLog *models.AuditLog) error

	// GetAuditLogByID retrieves an audit log by its ID
	GetAuditLogByID(ctx context.Context, id int64) (*models.AuditLog, error)

	// GetAuditLogs retrieves audit logs with optional filtering
	GetAuditLogs(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error)

	// GetAuditLogSummary returns summary statistics for audit logs
	GetAuditLogSummary(ctx context.Context, filter models.AuditLogFilter) (*models.AuditLogSummary, error)

	// GetLastChecksum retrieves the checksum of the last audit log entry
	GetLastChecksum(ctx context.Context) (*string, error)

	// VerifyIntegrity verifies the integrity of audit log entries
	VerifyIntegrity(ctx context.Context, startID, endID int64) ([]models.IntegrityCheckResult, error)

	// GetAuditLogsByUserID retrieves audit logs for a specific user
	GetAuditLogsByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.AuditLog, error)

	// GetDangerousOperations retrieves all dangerous operations within a time range
	GetDangerousOperations(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error)

	// GetFailedOperations retrieves all failed operations within a time range
	GetFailedOperations(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error)

	// CountAuditLogs counts the total number of audit logs matching the filter
	CountAuditLogs(ctx context.Context, filter models.AuditLogFilter) (int64, error)
}

// auditRepository implements AuditRepository interface
type auditRepository struct {
	db *sqlx.DB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *sqlx.DB) AuditRepository {
	return &auditRepository{
		db: db,
	}
}

// JSONB is a custom type to handle PostgreSQL JSONB columns
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return err
	}

	*j = JSONB(data)
	return nil
}

// CreateAuditLog creates a new audit log entry
func (r *auditRepository) CreateAuditLog(ctx context.Context, auditLog *models.AuditLog) error {
	// Get the last checksum to maintain audit trail integrity
	lastChecksum, err := r.GetLastChecksum(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last checksum: %w", err)
	}

	// Set checksum for integrity verification
	auditLog.SetChecksum(lastChecksum)

	query := `
		INSERT INTO audit_logs (
			user_id, session_id, query_text, generated_command, safety_level,
			execution_result, execution_status, cluster_context, namespace_context,
			timestamp, ip_address, user_agent, checksum, previous_checksum
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id`

	var executionResultJSON []byte
	if auditLog.ExecutionResult != nil {
		executionResultJSON, err = json.Marshal(auditLog.ExecutionResult)
		if err != nil {
			return fmt.Errorf("failed to marshal execution result: %w", err)
		}
	}

	err = r.db.QueryRowxContext(ctx, query,
		auditLog.UserID,
		auditLog.SessionID,
		auditLog.QueryText,
		auditLog.GeneratedCommand,
		auditLog.SafetyLevel,
		executionResultJSON,
		auditLog.ExecutionStatus,
		auditLog.ClusterContext,
		auditLog.NamespaceContext,
		auditLog.Timestamp,
		auditLog.IPAddress,
		auditLog.UserAgent,
		auditLog.Checksum,
		auditLog.PreviousChecksum,
	).Scan(&auditLog.ID)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetAuditLogByID retrieves an audit log by its ID
func (r *auditRepository) GetAuditLogByID(ctx context.Context, id int64) (*models.AuditLog, error) {
	query := `
		SELECT id, user_id, session_id, query_text, generated_command, safety_level,
			   execution_result, execution_status, cluster_context, namespace_context,
			   timestamp, ip_address, user_agent, checksum, previous_checksum
		FROM audit_logs 
		WHERE id = $1`

	var auditLog models.AuditLog
	var executionResultJSON []byte

	err := r.db.QueryRowxContext(ctx, query, id).Scan(
		&auditLog.ID,
		&auditLog.UserID,
		&auditLog.SessionID,
		&auditLog.QueryText,
		&auditLog.GeneratedCommand,
		&auditLog.SafetyLevel,
		&executionResultJSON,
		&auditLog.ExecutionStatus,
		&auditLog.ClusterContext,
		&auditLog.NamespaceContext,
		&auditLog.Timestamp,
		&auditLog.IPAddress,
		&auditLog.UserAgent,
		&auditLog.Checksum,
		&auditLog.PreviousChecksum,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("audit log with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	// Unmarshal execution result if it exists
	if executionResultJSON != nil {
		if err := json.Unmarshal(executionResultJSON, &auditLog.ExecutionResult); err != nil {
			return nil, fmt.Errorf("failed to unmarshal execution result: %w", err)
		}
	}

	return &auditLog, nil
}

// GetAuditLogs retrieves audit logs with optional filtering
func (r *auditRepository) GetAuditLogs(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error) {
	query, args := r.buildFilterQuery("SELECT id, user_id, session_id, query_text, generated_command, safety_level, execution_result, execution_status, cluster_context, namespace_context, timestamp, ip_address, user_agent, checksum, previous_checksum FROM audit_logs", filter, true)

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var auditLogs []*models.AuditLog
	for rows.Next() {
		var auditLog models.AuditLog
		var executionResultJSON []byte

		err := rows.Scan(
			&auditLog.ID,
			&auditLog.UserID,
			&auditLog.SessionID,
			&auditLog.QueryText,
			&auditLog.GeneratedCommand,
			&auditLog.SafetyLevel,
			&executionResultJSON,
			&auditLog.ExecutionStatus,
			&auditLog.ClusterContext,
			&auditLog.NamespaceContext,
			&auditLog.Timestamp,
			&auditLog.IPAddress,
			&auditLog.UserAgent,
			&auditLog.Checksum,
			&auditLog.PreviousChecksum,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// Unmarshal execution result if it exists
		if executionResultJSON != nil {
			if err := json.Unmarshal(executionResultJSON, &auditLog.ExecutionResult); err != nil {
				return nil, fmt.Errorf("failed to unmarshal execution result: %w", err)
			}
		}

		auditLogs = append(auditLogs, &auditLog)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return auditLogs, nil
}

// GetAuditLogSummary returns summary statistics for audit logs
func (r *auditRepository) GetAuditLogSummary(ctx context.Context, filter models.AuditLogFilter) (*models.AuditLogSummary, error) {
	query, args := r.buildFilterQuery(`
		SELECT 
			COUNT(*) as total_entries,
			COUNT(CASE WHEN safety_level = 'safe' THEN 1 END) as safe_operations,
			COUNT(CASE WHEN safety_level = 'warning' THEN 1 END) as warning_operations,
			COUNT(CASE WHEN safety_level = 'dangerous' THEN 1 END) as dangerous_operations,
			COUNT(CASE WHEN execution_status = 'success' THEN 1 END) as successful_operations,
			COUNT(CASE WHEN execution_status = 'failed' THEN 1 END) as failed_operations,
			COUNT(CASE WHEN execution_status = 'cancelled' THEN 1 END) as cancelled_operations
		FROM audit_logs`, filter, false)

	var summary models.AuditLogSummary
	err := r.db.QueryRowxContext(ctx, query, args...).Scan(
		&summary.TotalEntries,
		&summary.SafeOperations,
		&summary.WarningOps,
		&summary.DangerousOps,
		&summary.SuccessfulOps,
		&summary.FailedOps,
		&summary.CancelledOps,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log summary: %w", err)
	}

	return &summary, nil
}

// GetLastChecksum retrieves the checksum of the last audit log entry
func (r *auditRepository) GetLastChecksum(ctx context.Context) (*string, error) {
	query := `SELECT checksum FROM audit_logs ORDER BY id DESC LIMIT 1`

	var checksum sql.NullString
	err := r.db.QueryRowxContext(ctx, query).Scan(&checksum)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No previous entries
		}
		return nil, fmt.Errorf("failed to get last checksum: %w", err)
	}

	if checksum.Valid {
		return &checksum.String, nil
	}
	return nil, nil
}

// VerifyIntegrity verifies the integrity of audit log entries
func (r *auditRepository) VerifyIntegrity(ctx context.Context, startID, endID int64) ([]models.IntegrityCheckResult, error) {
	query := `
		SELECT id, user_id, session_id, query_text, generated_command, safety_level,
			   execution_result, execution_status, cluster_context, namespace_context,
			   timestamp, ip_address, user_agent, checksum, previous_checksum
		FROM audit_logs 
		WHERE id BETWEEN $1 AND $2
		ORDER BY id ASC`

	rows, err := r.db.QueryxContext(ctx, query, startID, endID)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs for integrity check: %w", err)
	}
	defer rows.Close()

	var results []models.IntegrityCheckResult
	for rows.Next() {
		var auditLog models.AuditLog
		var executionResultJSON []byte

		err := rows.Scan(
			&auditLog.ID,
			&auditLog.UserID,
			&auditLog.SessionID,
			&auditLog.QueryText,
			&auditLog.GeneratedCommand,
			&auditLog.SafetyLevel,
			&executionResultJSON,
			&auditLog.ExecutionStatus,
			&auditLog.ClusterContext,
			&auditLog.NamespaceContext,
			&auditLog.Timestamp,
			&auditLog.IPAddress,
			&auditLog.UserAgent,
			&auditLog.Checksum,
			&auditLog.PreviousChecksum,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log for integrity check: %w", err)
		}

		// Unmarshal execution result if it exists
		if executionResultJSON != nil {
			if err := json.Unmarshal(executionResultJSON, &auditLog.ExecutionResult); err != nil {
				return nil, fmt.Errorf("failed to unmarshal execution result: %w", err)
			}
		}

		// Verify integrity
		result := models.IntegrityCheckResult{
			LogID:   auditLog.ID,
			IsValid: auditLog.VerifyIntegrity(),
		}

		if !result.IsValid {
			result.ErrorMessage = "Checksum verification failed"
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}

// GetAuditLogsByUserID retrieves audit logs for a specific user
func (r *auditRepository) GetAuditLogsByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.AuditLog, error) {
	filter := models.AuditLogFilter{
		Limit:  limit,
		Offset: offset,
	}

	// Parse userID as UUID
	if userID != "" {
		if uid, err := parseUUID(userID); err == nil {
			filter.UserID = &uid
		}
	}

	return r.GetAuditLogs(ctx, filter)
}

// GetDangerousOperations retrieves all dangerous operations within a time range
func (r *auditRepository) GetDangerousOperations(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error) {
	dangerousLevel := models.SafetyLevelDangerous
	filter.SafetyLevel = &dangerousLevel
	return r.GetAuditLogs(ctx, filter)
}

// GetFailedOperations retrieves all failed operations within a time range
func (r *auditRepository) GetFailedOperations(ctx context.Context, filter models.AuditLogFilter) ([]*models.AuditLog, error) {
	failedStatus := models.ExecutionStatusFailed
	filter.Status = &failedStatus
	return r.GetAuditLogs(ctx, filter)
}

// CountAuditLogs counts the total number of audit logs matching the filter
func (r *auditRepository) CountAuditLogs(ctx context.Context, filter models.AuditLogFilter) (int64, error) {
	query, args := r.buildFilterQuery("SELECT COUNT(*) FROM audit_logs", filter, false)

	var count int64
	err := r.db.QueryRowxContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return count, nil
}

// buildFilterQuery constructs a SQL query with WHERE clauses based on the filter
func (r *auditRepository) buildFilterQuery(baseQuery string, filter models.AuditLogFilter, withPagination bool) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.SessionID != nil {
		conditions = append(conditions, fmt.Sprintf("session_id = $%d", argIndex))
		args = append(args, *filter.SessionID)
		argIndex++
	}

	if filter.SafetyLevel != nil {
		conditions = append(conditions, fmt.Sprintf("safety_level = $%d", argIndex))
		args = append(args, *filter.SafetyLevel)
		argIndex++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("execution_status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIndex))
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argIndex))
		args = append(args, *filter.EndTime)
		argIndex++
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + joinConditions(conditions, " AND ")
	}

	if withPagination {
		query += " ORDER BY timestamp DESC, id DESC"

		if filter.Limit > 0 {
			query += fmt.Sprintf(" LIMIT $%d", argIndex)
			args = append(args, filter.Limit)
			argIndex++
		}

		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, filter.Offset)
			argIndex++
		}
	}

	return query, args
}

// Helper function to join conditions with a separator
func joinConditions(conditions []string, separator string) string {
	if len(conditions) == 0 {
		return ""
	}

	result := conditions[0]
	for i := 1; i < len(conditions); i++ {
		result += separator + conditions[i]
	}
	return result
}

// Helper function to parse UUID string
func parseUUID(uuidStr string) (uuid.UUID, error) {
	return uuid.Parse(uuidStr)
}
