package command

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// Repository implements command execution repository
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new command execution repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// Create creates a new command execution record
func (r *Repository) Create(ctx context.Context, execution *models.KubernetesCommandExecution) error {
	// Serialize the command to JSON
	commandJSON, err := json.Marshal(execution.Command)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	// Serialize result if present
	var resultJSON []byte
	if execution.Result != nil {
		resultJSON, err = json.Marshal(execution.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
	}

	query := `
		INSERT INTO command_executions (
			id, user_id, session_id, operation_type, resource_type, namespace, 
			resource_name, safety_level, command_text, execution_status,
			result_data, error_message, execution_time_ms, created_at, executed_at, completed_at
		) VALUES (
			:id, :user_id, :session_id, :operation_type, :resource_type, :namespace,
			:resource_name, :safety_level, :command_text, :execution_status,
			:result_data, :error_message, :execution_time_ms, :created_at, :executed_at, :completed_at
		)`

	params := map[string]interface{}{
		"id":                execution.ID,
		"user_id":           execution.UserID,
		"session_id":        execution.SessionID,
		"operation_type":    execution.Command.Operation,
		"resource_type":     execution.Command.Resource,
		"namespace":         execution.Command.Namespace,
		"resource_name":     execution.Command.Name,
		"safety_level":      execution.SafetyLevel,
		"command_text":      string(commandJSON),
		"execution_status":  execution.Status,
		"result_data":       resultJSON,
		"error_message":     execution.ErrorMessage,
		"execution_time_ms": execution.ExecutionTimeMS,
		"created_at":        execution.CreatedAt,
		"executed_at":       execution.ExecutedAt,
		"completed_at":      execution.CompletedAt,
	}

	_, err = r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to insert command execution: %w", err)
	}

	return nil
}

// GetByID retrieves a command execution by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.KubernetesCommandExecution, error) {
	query := `
		SELECT 
			id, user_id, session_id, operation_type, resource_type, namespace,
			resource_name, safety_level, command_text, execution_status,
			result_data, error_message, execution_time_ms, created_at, executed_at, completed_at
		FROM command_executions
		WHERE id = $1`

	var row struct {
		ID              uuid.UUID      `db:"id"`
		UserID          uuid.UUID      `db:"user_id"`
		SessionID       uuid.UUID      `db:"session_id"`
		OperationType   string         `db:"operation_type"`
		ResourceType    string         `db:"resource_type"`
		Namespace       string         `db:"namespace"`
		ResourceName    sql.NullString `db:"resource_name"`
		SafetyLevel     string         `db:"safety_level"`
		CommandText     string         `db:"command_text"`
		ExecutionStatus string         `db:"execution_status"`
		ResultData      []byte         `db:"result_data"`
		ErrorMessage    sql.NullString `db:"error_message"`
		ExecutionTimeMS sql.NullInt32  `db:"execution_time_ms"`
		CreatedAt       time.Time      `db:"created_at"`
		ExecutedAt      sql.NullTime   `db:"executed_at"`
		CompletedAt     sql.NullTime   `db:"completed_at"`
	}

	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("command execution not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get command execution: %w", err)
	}

	// Deserialize command
	var command models.KubernetesOperation
	if err := json.Unmarshal([]byte(row.CommandText), &command); err != nil {
		return nil, fmt.Errorf("failed to unmarshal command: %w", err)
	}

	// Deserialize result if present
	var result *models.KubernetesOperationResult
	if len(row.ResultData) > 0 {
		result = &models.KubernetesOperationResult{}
		if err := json.Unmarshal(row.ResultData, result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	execution := &models.KubernetesCommandExecution{
		ID:          row.ID,
		UserID:      row.UserID,
		SessionID:   row.SessionID,
		Command:     command,
		SafetyLevel: row.SafetyLevel,
		Status:      row.ExecutionStatus,
		Result:      result,
		CreatedAt:   row.CreatedAt,
	}

	if row.ResourceName.Valid {
		execution.Command.Name = row.ResourceName.String
	}

	if row.ErrorMessage.Valid {
		execution.ErrorMessage = row.ErrorMessage.String
	}

	if row.ExecutionTimeMS.Valid {
		executionTimeMS := int(row.ExecutionTimeMS.Int32)
		execution.ExecutionTimeMS = &executionTimeMS
	}

	if row.ExecutedAt.Valid {
		execution.ExecutedAt = &row.ExecutedAt.Time
	}

	if row.CompletedAt.Valid {
		execution.CompletedAt = &row.CompletedAt.Time
	}

	return execution, nil
}

// Update updates a command execution record
func (r *Repository) Update(ctx context.Context, execution *models.KubernetesCommandExecution) error {
	// Serialize the command to JSON
	commandJSON, err := json.Marshal(execution.Command)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	// Serialize result if present
	var resultJSON []byte
	if execution.Result != nil {
		resultJSON, err = json.Marshal(execution.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
	}

	query := `
		UPDATE command_executions SET
			operation_type = :operation_type,
			resource_type = :resource_type,
			namespace = :namespace,
			resource_name = :resource_name,
			safety_level = :safety_level,
			command_text = :command_text,
			execution_status = :execution_status,
			result_data = :result_data,
			error_message = :error_message,
			execution_time_ms = :execution_time_ms,
			executed_at = :executed_at,
			completed_at = :completed_at
		WHERE id = :id`

	params := map[string]interface{}{
		"id":                execution.ID,
		"operation_type":    execution.Command.Operation,
		"resource_type":     execution.Command.Resource,
		"namespace":         execution.Command.Namespace,
		"resource_name":     execution.Command.Name,
		"safety_level":      execution.SafetyLevel,
		"command_text":      string(commandJSON),
		"execution_status":  execution.Status,
		"result_data":       resultJSON,
		"error_message":     execution.ErrorMessage,
		"execution_time_ms": execution.ExecutionTimeMS,
		"executed_at":       execution.ExecutedAt,
		"completed_at":      execution.CompletedAt,
	}

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to update command execution: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("command execution not found: %s", execution.ID)
	}

	return nil
}

// ListByUser retrieves command executions for a user
func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*models.KubernetesCommandExecution, error) {
	query := `
		SELECT 
			id, user_id, session_id, operation_type, resource_type, namespace,
			resource_name, safety_level, command_text, execution_status,
			result_data, error_message, execution_time_ms, created_at, executed_at, completed_at
		FROM command_executions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query command executions: %w", err)
	}
	defer rows.Close()

	var executions []*models.KubernetesCommandExecution

	for rows.Next() {
		var row struct {
			ID              uuid.UUID      `db:"id"`
			UserID          uuid.UUID      `db:"user_id"`
			SessionID       uuid.UUID      `db:"session_id"`
			OperationType   string         `db:"operation_type"`
			ResourceType    string         `db:"resource_type"`
			Namespace       string         `db:"namespace"`
			ResourceName    sql.NullString `db:"resource_name"`
			SafetyLevel     string         `db:"safety_level"`
			CommandText     string         `db:"command_text"`
			ExecutionStatus string         `db:"execution_status"`
			ResultData      []byte         `db:"result_data"`
			ErrorMessage    sql.NullString `db:"error_message"`
			ExecutionTimeMS sql.NullInt32  `db:"execution_time_ms"`
			CreatedAt       time.Time      `db:"created_at"`
			ExecutedAt      sql.NullTime   `db:"executed_at"`
			CompletedAt     sql.NullTime   `db:"completed_at"`
		}

		err := rows.Scan(
			&row.ID, &row.UserID, &row.SessionID, &row.OperationType, &row.ResourceType,
			&row.Namespace, &row.ResourceName, &row.SafetyLevel, &row.CommandText,
			&row.ExecutionStatus, &row.ResultData, &row.ErrorMessage, &row.ExecutionTimeMS,
			&row.CreatedAt, &row.ExecutedAt, &row.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan command execution row: %w", err)
		}

		// Deserialize command
		var command models.KubernetesOperation
		if err := json.Unmarshal([]byte(row.CommandText), &command); err != nil {
			return nil, fmt.Errorf("failed to unmarshal command: %w", err)
		}

		// Deserialize result if present
		var result *models.KubernetesOperationResult
		if len(row.ResultData) > 0 {
			result = &models.KubernetesOperationResult{}
			if err := json.Unmarshal(row.ResultData, result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal result: %w", err)
			}
		}

		execution := &models.KubernetesCommandExecution{
			ID:          row.ID,
			UserID:      row.UserID,
			SessionID:   row.SessionID,
			Command:     command,
			SafetyLevel: row.SafetyLevel,
			Status:      row.ExecutionStatus,
			Result:      result,
			CreatedAt:   row.CreatedAt,
		}

		if row.ResourceName.Valid {
			execution.Command.Name = row.ResourceName.String
		}

		if row.ErrorMessage.Valid {
			execution.ErrorMessage = row.ErrorMessage.String
		}

		if row.ExecutionTimeMS.Valid {
			executionTimeMS := int(row.ExecutionTimeMS.Int32)
			execution.ExecutionTimeMS = &executionTimeMS
		}

		if row.ExecutedAt.Valid {
			execution.ExecutedAt = &row.ExecutedAt.Time
		}

		if row.CompletedAt.Valid {
			execution.CompletedAt = &row.CompletedAt.Time
		}

		executions = append(executions, execution)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating command execution rows: %w", err)
	}

	return executions, nil
}
