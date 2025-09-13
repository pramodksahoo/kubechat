package command

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/command"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/rollback"
)

// Handler handles command execution HTTP requests
type Handler struct {
	commandService  command.Service
	rollbackService rollback.Service
}

// NewHandler creates a new command handler
func NewHandler(commandService command.Service, rollbackService rollback.Service) *Handler {
	return &Handler{
		commandService:  commandService,
		rollbackService: rollbackService,
	}
}

// ExecuteCommand handles command execution requests
// @Summary Execute a Kubernetes command
// @Description Execute a Kubernetes command with safety checks and approval workflow
// @Tags Commands
// @Accept json
// @Produce json
// @Param command body models.CommandExecutionRequest true "Command execution request"
// @Success 201 {object} models.KubernetesCommandExecution
// @Success 202 {object} models.CommandApproval "Command requires approval"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/commands/execute [post]
func (h *Handler) ExecuteCommand(c *gin.Context) {
	var req models.CommandExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get user ID from context (set by authentication middleware)
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
		return
	}

	req.UserID = userID

	// Get session ID from context or generate one
	sessionIDStr := c.GetString("session_id")
	if sessionIDStr == "" {
		req.SessionID = uuid.New()
	} else {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID format"})
			return
		}
		req.SessionID = sessionID
	}

	// Execute command with service
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	execution, approval, err := h.commandService.ExecuteCommand(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute command", "details": err.Error()})
		return
	}

	// If approval is required, return approval object
	if approval != nil {
		c.JSON(http.StatusAccepted, approval)
		return
	}

	// Return execution result
	c.JSON(http.StatusCreated, execution)
}

// GetExecution retrieves a command execution by ID
// @Summary Get command execution
// @Description Retrieve details of a command execution by ID
// @Tags Commands
// @Produce json
// @Param id path string true "Execution ID"
// @Success 200 {object} models.KubernetesCommandExecution
// @Failure 400 {object} map[string]string "Invalid execution ID"
// @Failure 404 {object} map[string]string "Execution not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/commands/executions/{id} [get]
func (h *Handler) GetExecution(c *gin.Context) {
	executionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	execution, err := h.commandService.GetExecution(ctx, executionID)
	if err != nil {
		if err.Error() == "execution not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get execution", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// ListExecutions lists command executions with filtering
// @Summary List command executions
// @Description List command executions with optional filtering by status and user
// @Tags Commands
// @Produce json
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{} "List of executions with pagination info"
// @Failure 400 {object} map[string]string "Invalid query parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/commands/executions [get]
func (h *Handler) ListExecutions(c *gin.Context) {
	// Parse query parameters
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter (must be 1-100)"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter (must be >= 0)"})
		return
	}

	// Get user ID from context
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	executions, total, err := h.commandService.ListExecutions(ctx, userID, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list executions", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"executions": executions,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
			"count":  len(executions),
		},
	})
}

// ApproveCommand approves a pending command execution
// @Summary Approve command execution
// @Description Approve a command execution that requires approval
// @Tags Commands
// @Accept json
// @Produce json
// @Param approval body models.CommandApprovalRequest true "Approval request"
// @Success 200 {object} models.CommandApproval
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Approval not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/commands/approve [post]
func (h *Handler) ApproveCommand(c *gin.Context) {
	var req models.CommandApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get user ID from context
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	approverID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	approval, err := h.commandService.ApproveCommand(ctx, req.ExecutionID, approverID, req.Decision, req.Reason)
	if err != nil {
		if err.Error() == "approval not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Approval not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve command", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, approval)
}

// GetPendingApprovals lists pending approvals for a user
// @Summary Get pending approvals
// @Description List all pending command approvals that require user decision
// @Tags Commands
// @Produce json
// @Success 200 {array} models.CommandApproval
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/commands/approvals/pending [get]
func (h *Handler) GetPendingApprovals(c *gin.Context) {
	// Get user ID from context
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	approvals, err := h.commandService.GetPendingApprovals(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pending approvals", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, approvals)
}

// CreateRollbackPlan creates a rollback plan for a command execution
// @Summary Create rollback plan
// @Description Create a rollback plan for a completed command execution
// @Tags Rollback
// @Accept json
// @Produce json
// @Param executionId path string true "Command Execution ID"
// @Success 201 {object} models.RollbackPlan
// @Failure 400 {object} map[string]string "Invalid execution ID or execution not eligible for rollback"
// @Failure 404 {object} map[string]string "Execution not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/commands/executions/{executionId}/rollback/plan [post]
func (h *Handler) CreateRollbackPlan(c *gin.Context) {
	executionID, err := uuid.Parse(c.Param("executionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	plan, err := h.rollbackService.CreateRollbackPlan(ctx, executionID)
	if err != nil {
		if err.Error() == "execution not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create rollback plan", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, plan)
}

// ExecuteRollback executes a rollback plan
// @Summary Execute rollback plan
// @Description Execute a rollback plan to revert a command execution
// @Tags Rollback
// @Accept json
// @Produce json
// @Param planId path string true "Rollback Plan ID"
// @Success 202 {object} models.RollbackExecution
// @Failure 400 {object} map[string]string "Invalid plan ID or plan not executable"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/rollback/plans/{planId}/execute [post]
func (h *Handler) ExecuteRollback(c *gin.Context) {
	planID, err := uuid.Parse(c.Param("planId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID format"})
		return
	}

	// Get user ID from context
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	execution, err := h.rollbackService.ExecuteRollback(ctx, planID, userID)
	if err != nil {
		if err.Error() == "rollback plan not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rollback plan not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to execute rollback", "details": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, execution)
}

// GetRollbackStatus gets the status of a rollback execution
// @Summary Get rollback status
// @Description Get the current status and progress of a rollback execution
// @Tags Rollback
// @Produce json
// @Param rollbackId path string true "Rollback Execution ID"
// @Success 200 {object} models.RollbackStatus
// @Failure 400 {object} map[string]string "Invalid rollback ID"
// @Failure 404 {object} map[string]string "Rollback not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/rollback/executions/{rollbackId}/status [get]
func (h *Handler) GetRollbackStatus(c *gin.Context) {
	rollbackID, err := uuid.Parse(c.Param("rollbackId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rollback ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	status, err := h.rollbackService.GetRollbackStatus(ctx, rollbackID)
	if err != nil {
		if err.Error() == "rollback execution not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rollback execution not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rollback status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// ValidateRollback validates if an execution can be rolled back
// @Summary Validate rollback eligibility
// @Description Check if a command execution is eligible for rollback
// @Tags Rollback
// @Produce json
// @Param executionId path string true "Command Execution ID"
// @Success 200 {object} models.RollbackValidation
// @Failure 400 {object} map[string]string "Invalid execution ID"
// @Failure 404 {object} map[string]string "Execution not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/commands/executions/{executionId}/rollback/validate [get]
func (h *Handler) ValidateRollback(c *gin.Context) {
	executionID, err := uuid.Parse(c.Param("executionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	validation, err := h.rollbackService.ValidateRollback(ctx, executionID)
	if err != nil {
		if err.Error() == "execution not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate rollback", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, validation)
}

// GetExecutionStats gets execution statistics for a user
// @Summary Get execution statistics
// @Description Get command execution statistics and analytics for a user
// @Tags Analytics
// @Produce json
// @Param from query string false "Start date (RFC3339 format)"
// @Param to query string false "End date (RFC3339 format)"
// @Success 200 {object} models.ExecutionStats
// @Failure 400 {object} map[string]string "Invalid date format"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/commands/stats [get]
func (h *Handler) GetExecutionStats(c *gin.Context) {
	// Get user ID from context
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Parse date range
	var from, to time.Time

	if fromStr := c.Query("from"); fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'from' date format (use RFC3339)"})
			return
		}
	} else {
		from = time.Now().AddDate(0, -1, 0) // Default to 1 month ago
	}

	if toStr := c.Query("to"); toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'to' date format (use RFC3339)"})
			return
		}
	} else {
		to = time.Now() // Default to now
	}

	if from.After(to) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "'from' date must be before 'to' date"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	stats, err := h.commandService.GetExecutionStats(ctx, userID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get execution stats", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// HealthCheck provides a health check endpoint for the command service
// @Summary Health check
// @Description Check the health of command execution services
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string "Service unavailable"
// @Router /api/v1/commands/health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Check command service health
	if err := h.commandService.HealthCheck(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"service": "command",
			"error":   err.Error(),
		})
		return
	}

	// Check rollback service health
	if err := h.rollbackService.HealthCheck(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"service": "rollback",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "command-execution",
	})
}
