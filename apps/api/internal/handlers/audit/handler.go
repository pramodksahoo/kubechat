package audit

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/audit"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/auth"
)

// Handler handles audit-related HTTP requests
type Handler struct {
	auditService audit.Service
}

// NewHandler creates a new audit handler
func NewHandler(auditService audit.Service) *Handler {
	return &Handler{
		auditService: auditService,
	}
}

// RegisterRoutes registers audit routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	auditRoutes := router.Group("/audit")
	auditRoutes.Use(auth.RequireWritePermission()) // Require user role for audit operations
	{
		// Audit log endpoints
		auditRoutes.GET("/logs", h.GetAuditLogs)
		auditRoutes.GET("/logs/:id", h.GetAuditLogByID)
		auditRoutes.POST("/logs", h.CreateAuditLog)

		// Summary and analytics endpoints
		auditRoutes.GET("/summary", h.GetAuditLogSummary)
		auditRoutes.GET("/dangerous", h.GetDangerousOperations)
		auditRoutes.GET("/failed", h.GetFailedOperations)

		// Integrity and health endpoints
		auditRoutes.POST("/verify-integrity", h.VerifyIntegrity)
		auditRoutes.GET("/health", h.HealthCheck)
		auditRoutes.GET("/metrics", h.GetMetrics)
	}
}

// GetAuditLogs retrieves audit logs with optional filtering
//
//	@Summary		Get audit logs with filtering
//	@Description	Retrieves audit logs with optional filtering by user, session, safety level, status, and time range
//	@Tags			Audit & Compliance
//	@Produce		json
//	@Param			user_id			query		string					false	"User ID to filter by"
//	@Param			session_id		query		string					false	"Session ID to filter by"
//	@Param			safety_level	query		string					false	"Safety level (safe, warning, dangerous)"
//	@Param			status			query		string					false	"Execution status (success, failed, cancelled)"
//	@Param			start_time		query		string					false	"Start time (RFC3339 format)"
//	@Param			end_time		query		string					false	"End time (RFC3339 format)"
//	@Param			limit			query		int						false	"Limit number of results (1-1000)"	default(50)
//	@Param			offset			query		int						false	"Offset for pagination"				default(0)
//	@Success		200				{object}	map[string]interface{}	"Audit logs retrieved successfully"
//	@Failure		400				{object}	map[string]interface{}	"Invalid query parameters"
//	@Failure		500				{object}	map[string]interface{}	"Failed to retrieve audit logs"
//	@Security		BearerAuth
//	@Router			/audit/logs [get]
func (h *Handler) GetAuditLogs(c *gin.Context) {
	filter := models.AuditLogFilter{}

	// Parse query parameters for filtering
	if userID := c.Query("user_id"); userID != "" {
		if uid, err := uuid.Parse(userID); err == nil {
			filter.UserID = &uid
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user_id format",
				"code":  "INVALID_USER_ID",
			})
			return
		}
	}

	if sessionID := c.Query("session_id"); sessionID != "" {
		if sid, err := uuid.Parse(sessionID); err == nil {
			filter.SessionID = &sid
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid session_id format",
				"code":  "INVALID_SESSION_ID",
			})
			return
		}
	}

	if safetyLevel := c.Query("safety_level"); safetyLevel != "" {
		if models.IsValidSafetyLevel(safetyLevel) {
			filter.SafetyLevel = &safetyLevel
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid safety_level. Must be: safe, warning, or dangerous",
				"code":  "INVALID_SAFETY_LEVEL",
			})
			return
		}
	}

	if status := c.Query("status"); status != "" {
		if models.IsValidExecutionStatus(status) {
			filter.Status = &status
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid status. Must be: success, failed, or cancelled",
				"code":  "INVALID_STATUS",
			})
			return
		}
	}

	// Parse time range
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filter.StartTime = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid start_time format. Use RFC3339 format",
				"code":  "INVALID_START_TIME",
			})
			return
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filter.EndTime = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid end_time format. Use RFC3339 format",
				"code":  "INVALID_END_TIME",
			})
			return
		}
	}

	// Parse pagination
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 1000 {
			filter.Limit = l
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid limit. Must be between 1 and 1000",
				"code":  "INVALID_LIMIT",
			})
			return
		}
	} else {
		filter.Limit = 50 // Default limit
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid offset. Must be >= 0",
				"code":  "INVALID_OFFSET",
			})
			return
		}
	}

	// Retrieve audit logs
	logs, err := h.auditService.GetAuditLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve audit logs",
			"code":  "AUDIT_LOGS_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   logs,
		"count":  len(logs),
		"filter": filter,
	})
}

// GetAuditLogByID retrieves a specific audit log by ID
//
//	@Summary		Get audit log by ID
//	@Description	Retrieves a specific audit log entry by its ID
//	@Tags			Audit & Compliance
//	@Produce		json
//	@Param			id	path		int						true	"Audit log ID"
//	@Success		200	{object}	map[string]interface{}	"Audit log retrieved successfully"
//	@Failure		400	{object}	map[string]interface{}	"Invalid audit log ID"
//	@Failure		404	{object}	map[string]interface{}	"Audit log not found"
//	@Security		BearerAuth
//	@Router			/audit/logs/{id} [get]
func (h *Handler) GetAuditLogByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid audit log ID",
			"code":  "INVALID_AUDIT_LOG_ID",
		})
		return
	}

	log, err := h.auditService.GetAuditLogByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Audit log not found",
			"code":  "AUDIT_LOG_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": log,
	})
}

// CreateAuditLog creates a new audit log entry
//
//	@Summary		Create a new audit log entry
//	@Description	Creates a new audit log entry for tracking user actions
//	@Tags			Audit & Compliance
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.AuditLogRequest	true	"Audit log data"
//	@Success		201		{object}	map[string]interface{}	"Audit log created successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request format"
//	@Failure		500		{object}	map[string]interface{}	"Failed to create audit log"
//	@Security		BearerAuth
//	@Router			/audit/logs [post]
func (h *Handler) CreateAuditLog(c *gin.Context) {
	var req models.AuditLogRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Get user context from authentication middleware
	userID, _, sessionID, _ := auth.ExtractUserContext(c)

	// Override user and session IDs with authenticated context
	if userID != "" {
		if uid, err := uuid.Parse(userID); err == nil {
			req.UserID = &uid
		}
	}

	if sessionID != "" {
		if sid, err := uuid.Parse(sessionID); err == nil {
			req.SessionID = &sid
		}
	}

	if err := h.auditService.LogUserAction(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create audit log",
			"code":  "AUDIT_LOG_CREATION_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Audit log created successfully",
	})
}

// GetAuditLogSummary returns summary statistics for audit logs
//
//	@Summary		Get audit log summary statistics
//	@Description	Returns summary statistics and analytics for audit logs within a time range
//	@Tags			Audit & Compliance
//	@Produce		json
//	@Param			start_time	query		string					false	"Start time for summary (RFC3339 format)"
//	@Param			end_time	query		string					false	"End time for summary (RFC3339 format)"
//	@Success		200			{object}	map[string]interface{}	"Audit log summary retrieved successfully"
//	@Failure		500			{object}	map[string]interface{}	"Failed to get audit log summary"
//	@Security		BearerAuth
//	@Router			/audit/summary [get]
func (h *Handler) GetAuditLogSummary(c *gin.Context) {
	filter := models.AuditLogFilter{}

	// Parse optional filter parameters (similar to GetAuditLogs)
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filter.StartTime = &t
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filter.EndTime = &t
		}
	}

	summary, err := h.auditService.GetAuditLogSummary(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get audit log summary",
			"code":  "AUDIT_SUMMARY_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": summary,
	})
}

// GetDangerousOperations retrieves dangerous operations
//
//	@Summary		Get dangerous operations
//	@Description	Retrieves audit logs for operations classified as dangerous
//	@Tags			Audit & Compliance
//	@Produce		json
//	@Param			start_time	query		string					false	"Start time (RFC3339 format)"
//	@Param			end_time	query		string					false	"End time (RFC3339 format)"
//	@Param			limit		query		int						false	"Limit number of results"	default(50)
//	@Param			offset		query		int						false	"Offset for pagination"		default(0)
//	@Success		200			{object}	map[string]interface{}	"Dangerous operations retrieved successfully"
//	@Failure		500			{object}	map[string]interface{}	"Failed to retrieve dangerous operations"
//	@Security		BearerAuth
//	@Router			/audit/dangerous [get]
func (h *Handler) GetDangerousOperations(c *gin.Context) {
	filter := models.AuditLogFilter{}

	// Parse time range and pagination parameters
	h.parseCommonFilterParams(c, &filter)

	logs, err := h.auditService.GetDangerousOperations(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve dangerous operations",
			"code":  "DANGEROUS_OPERATIONS_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"count": len(logs),
	})
}

// GetFailedOperations retrieves failed operations
//
//	@Summary		Get failed operations
//	@Description	Retrieves audit logs for operations that failed during execution
//	@Tags			Audit & Compliance
//	@Produce		json
//	@Param			start_time	query		string					false	"Start time (RFC3339 format)"
//	@Param			end_time	query		string					false	"End time (RFC3339 format)"
//	@Param			limit		query		int						false	"Limit number of results"	default(50)
//	@Param			offset		query		int						false	"Offset for pagination"		default(0)
//	@Success		200			{object}	map[string]interface{}	"Failed operations retrieved successfully"
//	@Failure		500			{object}	map[string]interface{}	"Failed to retrieve failed operations"
//	@Security		BearerAuth
//	@Router			/audit/failed [get]
func (h *Handler) GetFailedOperations(c *gin.Context) {
	filter := models.AuditLogFilter{}

	// Parse time range and pagination parameters
	h.parseCommonFilterParams(c, &filter)

	logs, err := h.auditService.GetFailedOperations(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve failed operations",
			"code":  "FAILED_OPERATIONS_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"count": len(logs),
	})
}

// VerifyIntegrity verifies audit log integrity
//
//	@Summary		Verify audit log integrity
//	@Description	Verifies the integrity of audit logs within a specified ID range
//	@Tags			Audit & Compliance
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{start_id=int,end_id=int}	true	"Integrity verification request"
//	@Success		200		{object}	map[string]interface{}			"Integrity verification completed"
//	@Failure		400		{object}	map[string]interface{}			"Invalid request format or ID range"
//	@Failure		500		{object}	map[string]interface{}			"Integrity check failed"
//	@Security		BearerAuth
//	@Router			/audit/verify-integrity [post]
func (h *Handler) VerifyIntegrity(c *gin.Context) {
	var req struct {
		StartID int64 `json:"start_id" binding:"required"`
		EndID   int64 `json:"end_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	if req.StartID > req.EndID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start_id must be less than or equal to end_id",
			"code":  "INVALID_ID_RANGE",
		})
		return
	}

	results, err := h.auditService.VerifyIntegrity(c.Request.Context(), req.StartID, req.EndID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify integrity",
			"code":  "INTEGRITY_CHECK_ERROR",
		})
		return
	}

	// Calculate summary statistics
	totalChecked := len(results)
	passed := 0
	failed := 0

	for _, result := range results {
		if result.IsValid {
			passed++
		} else {
			failed++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"results": results,
			"summary": gin.H{
				"total_checked": totalChecked,
				"passed":        passed,
				"failed":        failed,
				"success_rate":  float64(passed) / float64(totalChecked) * 100,
			},
		},
	})
}

// HealthCheck checks audit service health
//
//	@Summary		Check audit service health
//	@Description	Performs a health check on the audit service and returns operational status
//	@Tags			Audit & Compliance
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Audit service is operational"
//	@Failure		503	{object}	map[string]interface{}	"Audit service health check failed"
//	@Security		BearerAuth
//	@Router			/audit/health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	err := h.auditService.HealthCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Audit service health check failed",
			"code":  "AUDIT_HEALTH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "Audit service is operational",
	})
}

// GetMetrics returns audit service metrics
//
//	@Summary		Get audit service metrics
//	@Description	Returns performance metrics and statistics for the audit service
//	@Tags			Audit & Compliance
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Audit service metrics"
//	@Failure		500	{object}	map[string]interface{}	"Failed to get audit metrics"
//	@Security		BearerAuth
//	@Router			/audit/metrics [get]
func (h *Handler) GetMetrics(c *gin.Context) {
	metrics, err := h.auditService.GetMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get audit metrics",
			"code":  "AUDIT_METRICS_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": metrics,
	})
}

// parseCommonFilterParams parses common filter parameters
func (h *Handler) parseCommonFilterParams(c *gin.Context, filter *models.AuditLogFilter) {
	// Parse time range
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filter.StartTime = &t
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filter.EndTime = &t
		}
	}

	// Parse pagination
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 1000 {
			filter.Limit = l
		}
	} else {
		filter.Limit = 50 // Default limit
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}
}
