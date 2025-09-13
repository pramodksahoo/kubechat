package command

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// ApprovalRepository implements command approval repository
type ApprovalRepository struct {
	db *sqlx.DB
}

// NewApprovalRepository creates a new command approval repository
func NewApprovalRepository(db *sqlx.DB) *ApprovalRepository {
	return &ApprovalRepository{
		db: db,
	}
}

// Create creates a new command approval record
func (r *ApprovalRepository) Create(ctx context.Context, approval *models.CommandApproval) error {
	query := `
		INSERT INTO command_approvals (
			id, command_execution_id, requested_by_user_id, approved_by_user_id,
			approval_status, approval_reason, expires_at, created_at, decided_at
		) VALUES (
			:id, :command_execution_id, :requested_by_user_id, :approved_by_user_id,
			:approval_status, :approval_reason, :expires_at, :created_at, :decided_at
		)`

	params := map[string]interface{}{
		"id":                   approval.ID,
		"command_execution_id": approval.CommandExecutionID,
		"requested_by_user_id": approval.RequestedByUserID,
		"approved_by_user_id":  approval.ApprovedByUserID,
		"approval_status":      approval.Status,
		"approval_reason":      approval.Reason,
		"expires_at":           approval.ExpiresAt,
		"created_at":           approval.CreatedAt,
		"decided_at":           approval.DecidedAt,
	}

	_, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to insert command approval: %w", err)
	}

	return nil
}

// GetByID retrieves a command approval by ID
func (r *ApprovalRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CommandApproval, error) {
	query := `
		SELECT 
			id, command_execution_id, requested_by_user_id, approved_by_user_id,
			approval_status, approval_reason, expires_at, created_at, decided_at
		FROM command_approvals
		WHERE id = $1`

	var row struct {
		ID                 uuid.UUID      `db:"id"`
		CommandExecutionID uuid.UUID      `db:"command_execution_id"`
		RequestedByUserID  uuid.UUID      `db:"requested_by_user_id"`
		ApprovedByUserID   *uuid.UUID     `db:"approved_by_user_id"`
		ApprovalStatus     string         `db:"approval_status"`
		ApprovalReason     sql.NullString `db:"approval_reason"`
		ExpiresAt          sql.NullTime   `db:"expires_at"`
		CreatedAt          time.Time      `db:"created_at"`
		DecidedAt          sql.NullTime   `db:"decided_at"`
	}

	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("command approval not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get command approval: %w", err)
	}

	approval := &models.CommandApproval{
		ID:                 row.ID,
		CommandExecutionID: row.CommandExecutionID,
		RequestedByUserID:  row.RequestedByUserID,
		ApprovedByUserID:   row.ApprovedByUserID,
		Status:             row.ApprovalStatus,
		CreatedAt:          row.CreatedAt,
	}

	if row.ApprovalReason.Valid {
		approval.Reason = row.ApprovalReason.String
	}

	if row.ExpiresAt.Valid {
		approval.ExpiresAt = row.ExpiresAt.Time
	}

	if row.DecidedAt.Valid {
		approval.DecidedAt = &row.DecidedAt.Time
	}

	return approval, nil
}

// Update updates a command approval record
func (r *ApprovalRepository) Update(ctx context.Context, approval *models.CommandApproval) error {
	query := `
		UPDATE command_approvals SET
			approved_by_user_id = :approved_by_user_id,
			approval_status = :approval_status,
			approval_reason = :approval_reason,
			expires_at = :expires_at,
			decided_at = :decided_at
		WHERE id = :id`

	params := map[string]interface{}{
		"id":                  approval.ID,
		"approved_by_user_id": approval.ApprovedByUserID,
		"approval_status":     approval.Status,
		"approval_reason":     approval.Reason,
		"expires_at":          approval.ExpiresAt,
		"decided_at":          approval.DecidedAt,
	}

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to update command approval: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("command approval not found: %s", approval.ID)
	}

	return nil
}

// ListPendingByUser retrieves pending approvals for a user (approver)
func (r *ApprovalRepository) ListPendingByUser(ctx context.Context, userID uuid.UUID) ([]*models.CommandApproval, error) {
	query := `
		SELECT 
			id, command_execution_id, requested_by_user_id, approved_by_user_id,
			approval_status, approval_reason, expires_at, created_at, decided_at
		FROM command_approvals
		WHERE approval_status = 'pending' 
		AND expires_at > NOW()
		AND requested_by_user_id != $1  -- Don't show own requests
		ORDER BY created_at ASC
		LIMIT 50`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending approvals: %w", err)
	}
	defer rows.Close()

	var approvals []*models.CommandApproval

	for rows.Next() {
		var row struct {
			ID                 uuid.UUID      `db:"id"`
			CommandExecutionID uuid.UUID      `db:"command_execution_id"`
			RequestedByUserID  uuid.UUID      `db:"requested_by_user_id"`
			ApprovedByUserID   *uuid.UUID     `db:"approved_by_user_id"`
			ApprovalStatus     string         `db:"approval_status"`
			ApprovalReason     sql.NullString `db:"approval_reason"`
			ExpiresAt          sql.NullTime   `db:"expires_at"`
			CreatedAt          time.Time      `db:"created_at"`
			DecidedAt          sql.NullTime   `db:"decided_at"`
		}

		err := rows.Scan(
			&row.ID, &row.CommandExecutionID, &row.RequestedByUserID, &row.ApprovedByUserID,
			&row.ApprovalStatus, &row.ApprovalReason, &row.ExpiresAt, &row.CreatedAt, &row.DecidedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan approval row: %w", err)
		}

		approval := &models.CommandApproval{
			ID:                 row.ID,
			CommandExecutionID: row.CommandExecutionID,
			RequestedByUserID:  row.RequestedByUserID,
			ApprovedByUserID:   row.ApprovedByUserID,
			Status:             row.ApprovalStatus,
			CreatedAt:          row.CreatedAt,
		}

		if row.ApprovalReason.Valid {
			approval.Reason = row.ApprovalReason.String
		}

		if row.ExpiresAt.Valid {
			approval.ExpiresAt = row.ExpiresAt.Time
		}

		if row.DecidedAt.Valid {
			approval.DecidedAt = &row.DecidedAt.Time
		}

		approvals = append(approvals, approval)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating approval rows: %w", err)
	}

	return approvals, nil
}

// ListExpired retrieves expired approval requests
func (r *ApprovalRepository) ListExpired(ctx context.Context) ([]*models.CommandApproval, error) {
	query := `
		SELECT 
			id, command_execution_id, requested_by_user_id, approved_by_user_id,
			approval_status, approval_reason, expires_at, created_at, decided_at
		FROM command_approvals
		WHERE approval_status = 'pending' 
		AND expires_at <= NOW()
		ORDER BY expires_at ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired approvals: %w", err)
	}
	defer rows.Close()

	var approvals []*models.CommandApproval

	for rows.Next() {
		var row struct {
			ID                 uuid.UUID      `db:"id"`
			CommandExecutionID uuid.UUID      `db:"command_execution_id"`
			RequestedByUserID  uuid.UUID      `db:"requested_by_user_id"`
			ApprovedByUserID   *uuid.UUID     `db:"approved_by_user_id"`
			ApprovalStatus     string         `db:"approval_status"`
			ApprovalReason     sql.NullString `db:"approval_reason"`
			ExpiresAt          sql.NullTime   `db:"expires_at"`
			CreatedAt          time.Time      `db:"created_at"`
			DecidedAt          sql.NullTime   `db:"decided_at"`
		}

		err := rows.Scan(
			&row.ID, &row.CommandExecutionID, &row.RequestedByUserID, &row.ApprovedByUserID,
			&row.ApprovalStatus, &row.ApprovalReason, &row.ExpiresAt, &row.CreatedAt, &row.DecidedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expired approval row: %w", err)
		}

		approval := &models.CommandApproval{
			ID:                 row.ID,
			CommandExecutionID: row.CommandExecutionID,
			RequestedByUserID:  row.RequestedByUserID,
			ApprovedByUserID:   row.ApprovedByUserID,
			Status:             row.ApprovalStatus,
			CreatedAt:          row.CreatedAt,
		}

		if row.ApprovalReason.Valid {
			approval.Reason = row.ApprovalReason.String
		}

		if row.ExpiresAt.Valid {
			approval.ExpiresAt = row.ExpiresAt.Time
		}

		if row.DecidedAt.Valid {
			approval.DecidedAt = &row.DecidedAt.Time
		}

		approvals = append(approvals, approval)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expired approval rows: %w", err)
	}

	return approvals, nil
}

// UpdateExpiredApprovals marks expired approvals and returns count
func (r *ApprovalRepository) UpdateExpiredApprovals(ctx context.Context) (int, error) {
	query := `
		UPDATE command_approvals 
		SET approval_status = 'expired', decided_at = NOW()
		WHERE approval_status = 'pending' 
		AND expires_at <= NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to update expired approvals: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}
