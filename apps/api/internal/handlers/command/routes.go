package command

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all command execution routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// API v1 group
	v1 := router.Group("/api/v1")

	// Commands group
	commands := v1.Group("/commands")
	{
		// Command execution endpoints
		commands.POST("/execute", h.ExecuteCommand)
		commands.GET("/executions/:id", h.GetExecution)
		commands.GET("/executions", h.ListExecutions)
		commands.GET("/stats", h.GetExecutionStats)

		// Approval workflow endpoints
		commands.POST("/approve", h.ApproveCommand)
		commands.GET("/approvals/pending", h.GetPendingApprovals)

		// Rollback validation endpoint (nested under executions)
		commands.GET("/executions/:id/rollback/validate", h.ValidateRollback)
		commands.POST("/executions/:id/rollback/plan", h.CreateRollbackPlan)

		// Health check
		commands.GET("/health", h.HealthCheck)
	}

	// Rollback group (separate from commands for clarity)
	rollback := v1.Group("/rollback")
	{
		// Rollback plan execution and monitoring
		rollback.POST("/plans/:planId/execute", h.ExecuteRollback)
		rollback.GET("/executions/:rollbackId/status", h.GetRollbackStatus)
	}
}

// RegisterRoutesWithMiddleware registers routes with authentication and authorization middleware
func (h *Handler) RegisterRoutesWithMiddleware(router *gin.Engine, authMiddleware gin.HandlerFunc, rbacMiddleware gin.HandlerFunc) {
	// API v1 group with authentication
	v1 := router.Group("/api/v1")
	v1.Use(authMiddleware) // All routes require authentication

	// Commands group with RBAC
	commands := v1.Group("/commands")
	commands.Use(rbacMiddleware) // All command operations require proper permissions
	{
		// Command execution endpoints
		commands.POST("/execute", h.ExecuteCommand)
		commands.GET("/executions/:id", h.GetExecution)
		commands.GET("/executions", h.ListExecutions)
		commands.GET("/stats", h.GetExecutionStats)

		// Approval workflow endpoints (may need elevated permissions)
		commands.POST("/approve", h.ApproveCommand)
		commands.GET("/approvals/pending", h.GetPendingApprovals)

		// Rollback validation and planning
		commands.GET("/executions/:id/rollback/validate", h.ValidateRollback)
		commands.POST("/executions/:id/rollback/plan", h.CreateRollbackPlan)

		// Health check (no additional permissions needed)
		commands.GET("/health", h.HealthCheck)
	}

	// Rollback group with RBAC
	rollback := v1.Group("/rollback")
	rollback.Use(rbacMiddleware) // Rollback operations require proper permissions
	{
		// Rollback execution and monitoring
		rollback.POST("/plans/:planId/execute", h.ExecuteRollback)
		rollback.GET("/executions/:rollbackId/status", h.GetRollbackStatus)
	}
}
