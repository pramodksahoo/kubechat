package command

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/safety"
)

// WebSocketService defines a minimal interface for WebSocket notifications
type WebSocketService interface {
	BroadcastToUser(userID uuid.UUID, message *models.WebSocketMessage) error
}

// ApprovalService defines the approval workflow interface
type ApprovalService interface {
	// Approval workflow management
	CreateApprovalRequest(ctx context.Context, req *models.CommandApprovalRequest) (*models.CommandApproval, error)
	ProcessApproval(ctx context.Context, req *models.CommandApprovalRequest) (*models.CommandApproval, error)
	GetApproval(ctx context.Context, approvalID uuid.UUID) (*models.CommandApproval, error)
	ListPendingApprovals(ctx context.Context, userID uuid.UUID) ([]*models.CommandApproval, error)

	// WebSocket notifications
	SendApprovalNotification(ctx context.Context, approval *models.CommandApproval, eventType string) error

	// Safety classification integration
	ClassifyCommandSafety(ctx context.Context, command *models.KubernetesOperation, userRole, environment string) (*safety.SafetyClassification, error)
	RequiresApproval(ctx context.Context, safetyLevel string) bool

	// Approval expiration management
	ExpireApprovals(ctx context.Context) (int, error)
	IsApprovalExpired(approval *models.CommandApproval) bool
}

// ApprovalRepository defines the approval repository interface
type ApprovalRepository interface {
	Create(ctx context.Context, approval *models.CommandApproval) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.CommandApproval, error)
	Update(ctx context.Context, approval *models.CommandApproval) error
	ListPendingByUser(ctx context.Context, userID uuid.UUID) ([]*models.CommandApproval, error)
	ListExpired(ctx context.Context) ([]*models.CommandApproval, error)
	UpdateExpiredApprovals(ctx context.Context) (int, error)
}

// approvalService implements the approval workflow service
type approvalService struct {
	approvalRepo     ApprovalRepository
	commandRepo      Repository
	safetyService    safety.Service
	websocketService WebSocketService
	defaultTTL       time.Duration // Default approval expiration time
}

// NewApprovalService creates a new approval workflow service
func NewApprovalService(approvalRepo ApprovalRepository, commandRepo Repository, safetyService safety.Service, websocketService WebSocketService) ApprovalService {
	return &approvalService{
		approvalRepo:     approvalRepo,
		commandRepo:      commandRepo,
		safetyService:    safetyService,
		websocketService: websocketService,
		defaultTTL:       1 * time.Hour, // Default 1 hour expiration
	}
}

// CreateApprovalRequest creates a new approval request for dangerous operations
func (s *approvalService) CreateApprovalRequest(ctx context.Context, req *models.CommandApprovalRequest) (*models.CommandApproval, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid approval request: %w", err)
	}

	// Get the command execution
	execution, err := s.commandRepo.GetByID(ctx, req.ExecutionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get command execution: %w", err)
	}

	// Verify the execution requires approval
	if !execution.RequiresApproval() {
		return nil, fmt.Errorf("command execution does not require approval (safety level: %s)", execution.SafetyLevel)
	}

	// Create approval record
	approval := &models.CommandApproval{
		ID:                 uuid.New(),
		CommandExecutionID: req.ExecutionID,
		RequestedByUserID:  req.UserID,
		Status:             "pending",
		Reason:             req.Reason,
		ExpiresAt:          time.Now().Add(s.defaultTTL),
		CreatedAt:          time.Now(),
	}

	// Store approval record
	if err := s.approvalRepo.Create(ctx, approval); err != nil {
		return nil, fmt.Errorf("failed to create approval request: %w", err)
	}

	// Update command execution with approval ID
	execution.ApprovalID = &approval.ID
	execution.Status = "pending_approval"
	if err := s.commandRepo.Update(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to update command execution: %w", err)
	}

	// Send notification for approval request
	if err := s.SendApprovalNotification(ctx, approval, "approval_requested"); err != nil {
		// Log error but don't fail the approval creation
		fmt.Printf("Failed to send approval notification: %v\n", err)
	}

	return approval, nil
}

// ProcessApproval processes an approval decision (approve/reject)
func (s *approvalService) ProcessApproval(ctx context.Context, req *models.CommandApprovalRequest) (*models.CommandApproval, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid approval request: %w", err)
	}

	// Get the command execution
	execution, err := s.commandRepo.GetByID(ctx, req.ExecutionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get command execution: %w", err)
	}

	// Check if approval exists
	if execution.ApprovalID == nil {
		return nil, fmt.Errorf("no approval request found for execution: %s", req.ExecutionID)
	}

	// Get the approval record
	approval, err := s.approvalRepo.GetByID(ctx, *execution.ApprovalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get approval record: %w", err)
	}

	// Check if approval is still pending
	if approval.Status != "pending" {
		return nil, fmt.Errorf("approval already processed with status: %s", approval.Status)
	}

	// Check if approval has expired
	if s.IsApprovalExpired(approval) {
		approval.Status = "expired"
		approval.DecidedAt = &time.Time{}
		*approval.DecidedAt = time.Now()

		if err := s.approvalRepo.Update(ctx, approval); err != nil {
			return nil, fmt.Errorf("failed to update expired approval: %w", err)
		}

		return nil, fmt.Errorf("approval request has expired")
	}

	// Process the approval decision
	decidedAt := time.Now()
	approval.ApprovedByUserID = &req.UserID
	approval.Status = req.Decision // "approve" or "reject"
	approval.Reason = req.Reason
	approval.DecidedAt = &decidedAt

	// Update approval record
	if err := s.approvalRepo.Update(ctx, approval); err != nil {
		return nil, fmt.Errorf("failed to update approval: %w", err)
	}

	// Update command execution status based on approval decision
	if req.Decision == "approve" {
		execution.Status = "approved"
	} else {
		execution.Status = "rejected"
	}

	if err := s.commandRepo.Update(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to update command execution status: %w", err)
	}

	// Send notification for approval decision
	if err := s.SendApprovalNotification(ctx, approval, "approval_decided"); err != nil {
		// Log error but don't fail the approval processing
		fmt.Printf("Failed to send approval decision notification: %v\n", err)
	}

	return approval, nil
}

// GetApproval retrieves an approval by ID
func (s *approvalService) GetApproval(ctx context.Context, approvalID uuid.UUID) (*models.CommandApproval, error) {
	return s.approvalRepo.GetByID(ctx, approvalID)
}

// ListPendingApprovals lists pending approval requests for a user
func (s *approvalService) ListPendingApprovals(ctx context.Context, userID uuid.UUID) ([]*models.CommandApproval, error) {
	return s.approvalRepo.ListPendingByUser(ctx, userID)
}

// ClassifyCommandSafety integrates with Story 1.5 safety classification
func (s *approvalService) ClassifyCommandSafety(ctx context.Context, command *models.KubernetesOperation, userRole, environment string) (*safety.SafetyClassification, error) {
	if s.safetyService == nil {
		// Fallback to basic safety classification
		safetyLevel := command.GetSafetyLevel()
		return &safety.SafetyClassification{
			Level:            safety.SafetyLevel(safetyLevel),
			Score:            getSafetyScore(safetyLevel),
			RequiresApproval: safetyLevel == "dangerous" || safetyLevel == "warning",
			Blocked:          safetyLevel == "dangerous" && environment == "production",
		}, nil
	}

	// Use comprehensive safety service from Story 1.5
	req := safety.ContextualSafetyRequest{
		Command:     command.GetDescription(),
		UserRole:    userRole,
		Environment: environment,
		Namespace:   command.Namespace,
		Context: map[string]string{
			"operation": command.Operation,
			"resource":  command.Resource,
			"name":      command.Name,
		},
	}

	return s.safetyService.ClassifyWithContext(ctx, req)
}

// RequiresApproval determines if a safety level requires approval
func (s *approvalService) RequiresApproval(ctx context.Context, safetyLevel string) bool {
	return safetyLevel == "dangerous" || safetyLevel == "warning"
}

// ExpireApprovals marks expired approvals and returns count
func (s *approvalService) ExpireApprovals(ctx context.Context) (int, error) {
	return s.approvalRepo.UpdateExpiredApprovals(ctx)
}

// IsApprovalExpired checks if an approval has expired
func (s *approvalService) IsApprovalExpired(approval *models.CommandApproval) bool {
	return time.Now().After(approval.ExpiresAt)
}

// SendApprovalNotification sends WebSocket notification for approval events
func (s *approvalService) SendApprovalNotification(ctx context.Context, approval *models.CommandApproval, eventType string) error {
	if s.websocketService == nil {
		return nil // Skip notifications if WebSocket service not available
	}

	// Get the command execution for context
	execution, err := s.commandRepo.GetByID(ctx, approval.CommandExecutionID)
	if err != nil {
		return fmt.Errorf("failed to get command execution for notification: %w", err)
	}

	// Create notification payload
	payload := map[string]any{
		"event_type":          eventType,
		"approval_id":         approval.ID,
		"execution_id":        approval.CommandExecutionID,
		"command_description": execution.GetDescription(),
		"safety_level":        execution.SafetyLevel,
		"status":              approval.Status,
		"requested_by":        approval.RequestedByUserID,
		"approved_by":         approval.ApprovedByUserID,
		"expires_at":          approval.ExpiresAt,
		"reason":              approval.Reason,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal notification payload: %w", err)
	}

	// Create notification message
	message := &models.WebSocketMessage{
		ID:        uuid.New().String(),
		Type:      models.WSMsgTypeNotification,
		Payload:   payloadJSON,
		Timestamp: time.Now(),
	}

	// Send notification based on event type
	switch eventType {
	case "approval_requested":
		// Notify potential approvers (for now, this would be admin users)
		// In a real implementation, this would query for users with approval permissions
		return s.websocketService.BroadcastToUser(approval.RequestedByUserID, message)

	case "approval_decided":
		// Notify the requester of the decision
		return s.websocketService.BroadcastToUser(approval.RequestedByUserID, message)

	case "approval_expired":
		// Notify the requester that approval expired
		return s.websocketService.BroadcastToUser(approval.RequestedByUserID, message)

	default:
		return fmt.Errorf("unknown approval notification event type: %s", eventType)
	}
}

// getSafetyScore returns a numeric safety score for basic classification
func getSafetyScore(safetyLevel string) float64 {
	switch safetyLevel {
	case "safe":
		return 90.0
	case "warning":
		return 60.0
	case "dangerous":
		return 20.0
	default:
		return 50.0
	}
}
