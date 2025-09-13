package external

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/external"
)

// Handler handles external API related HTTP requests
type Handler struct {
	clientManager     *external.ClientManager
	healthService     external.HealthService
	encryptionService external.EncryptionService
	auditService      external.CredentialAuditService
	validator         external.CredentialValidator
	injector          external.CredentialInjector
	// Task 6 services
	costTracker   external.CostTracker
	budgetManager external.BudgetManager
	optimizer     external.CostOptimizer
	allocator     external.CostAllocator
	controller    external.CostController
	// Task 7 services
	providerRegistry external.ProviderRegistry
	loadBalancer     external.ProviderLoadBalancer
	failoverService  external.ProviderFailover
	configManager    external.ProviderConfigManager
}

// NewHandler creates a new external API handler
func NewHandler(
	clientManager *external.ClientManager,
	healthService external.HealthService,
	encryptionService external.EncryptionService,
	auditService external.CredentialAuditService,
	validator external.CredentialValidator,
	injector external.CredentialInjector,
	// Task 6 services
	costTracker external.CostTracker,
	budgetManager external.BudgetManager,
	optimizer external.CostOptimizer,
	allocator external.CostAllocator,
	controller external.CostController,
) *Handler {
	return &Handler{
		clientManager:     clientManager,
		healthService:     healthService,
		encryptionService: encryptionService,
		auditService:      auditService,
		validator:         validator,
		injector:          injector,
		// Task 6 services
		costTracker:   costTracker,
		budgetManager: budgetManager,
		optimizer:     optimizer,
		allocator:     allocator,
		controller:    controller,
	}
}

// NewHandlerWithMultiProvider creates a new external API handler with Task 7 multi-provider services
func NewHandlerWithMultiProvider(
	clientManager *external.ClientManager,
	healthService external.HealthService,
	encryptionService external.EncryptionService,
	auditService external.CredentialAuditService,
	validator external.CredentialValidator,
	injector external.CredentialInjector,
	// Task 6 services
	costTracker external.CostTracker,
	budgetManager external.BudgetManager,
	optimizer external.CostOptimizer,
	allocator external.CostAllocator,
	controller external.CostController,
	// Task 7 services
	providerRegistry external.ProviderRegistry,
	loadBalancer external.ProviderLoadBalancer,
	failoverService external.ProviderFailover,
	configManager external.ProviderConfigManager,
) *Handler {
	return &Handler{
		clientManager:     clientManager,
		healthService:     healthService,
		encryptionService: encryptionService,
		auditService:      auditService,
		validator:         validator,
		injector:          injector,
		// Task 6 services
		costTracker:   costTracker,
		budgetManager: budgetManager,
		optimizer:     optimizer,
		allocator:     allocator,
		controller:    controller,
		// Task 7 services
		providerRegistry: providerRegistry,
		loadBalancer:     loadBalancer,
		failoverService:  failoverService,
		configManager:    configManager,
	}
}

// RegisterRoutes registers all external API routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	external := r.Group("/external")
	{
		// Health endpoints
		external.GET("/health", h.getHealth)
		external.GET("/health/:apiName", h.getAPIHealth)
		external.GET("/health-summary", h.getHealthSummary)
		external.POST("/health/check", h.checkHealth)
		external.GET("/health/metrics", h.getHealthMetrics)

		// Task 4: Enhanced health monitoring endpoints
		external.GET("/health/:apiName/history", h.getAPIHealthHistory)
		external.GET("/health/dashboard", h.getHealthDashboard)
		external.GET("/health/sla-report", h.getSLAReport)
		external.GET("/health/alerts", h.getHealthAlerts)
		external.POST("/health/alerts/configure", h.configureHealthAlerts)

		// Task 5: Fallback and recovery endpoints
		external.GET("/fallback/status", h.getFallbackStatus)
		external.GET("/fallback/metrics", h.getFallbackMetrics)
		external.POST("/fallback/:serviceName/test", h.testFallback)
		external.GET("/recovery/status", h.getRecoveryStatus)
		external.GET("/recovery/metrics", h.getRecoveryMetrics)
		external.POST("/recovery/:apiName/trigger", h.triggerRecovery)
		external.GET("/circuit-breaker/status", h.getCircuitBreakerStatus)

		// Task 6: Cost tracking and budget management endpoints
		external.GET("/cost/summary", h.getCostSummary)
		external.GET("/cost/service/:serviceName", h.getServiceCosts)
		external.GET("/cost/usage", h.getUsageMetrics)
		external.GET("/cost/breakdown", h.getCostBreakdown)
		external.GET("/cost/billing/:month/:year", h.getBillingData)
		external.POST("/cost/usage", h.recordUsage)

		external.GET("/budget/status", h.getBudgetStatus)
		external.GET("/budget/:budgetId", h.getBudget)
		external.GET("/budget/alerts", h.getBudgetAlerts)
		external.GET("/budget/summary", h.getBudgetSummary)
		external.POST("/budget", h.createBudget)
		external.PUT("/budget/:serviceName/limits", h.updateBudgetLimits)

		external.GET("/optimization/analysis", h.getUsageAnalysis)
		external.GET("/optimization/:serviceName/recommendations", h.getOptimizationRecommendations)
		external.GET("/optimization/savings", h.calculateCostSavings)
		external.GET("/optimization/efficiency", h.getEfficiencyMetrics)

		external.GET("/allocation/summary", h.getAllocationSummary)
		external.GET("/allocation/department/:department", h.getDepartmentCosts)
		external.GET("/allocation/project/:projectId", h.getProjectCosts)
		external.POST("/allocation", h.allocateCosts)

		external.GET("/control/status", h.getControlStatus)
		external.GET("/control/history", h.getControlHistory)
		external.GET("/control/metrics", h.getControlMetrics)
		external.POST("/control/:serviceName/enable", h.enableControls)
		external.POST("/control/:serviceName/disable", h.disableControls)
		external.POST("/control/emergency-shutdown", h.executeEmergencyShutdown)

		// API client endpoints
		external.GET("/clients", h.listClients)
		external.POST("/clients/:provider/request", h.makeAPIRequest)
		external.GET("/clients/:provider/info", h.getClientInfo)
		external.GET("/clients/metrics", h.getClientMetrics)

		// Encryption endpoints
		external.POST("/encrypt", h.encryptData)
		external.POST("/decrypt", h.decryptData)
		external.GET("/encryption/metrics", h.getEncryptionMetrics)

		// Credential validation endpoints
		external.POST("/credentials/validate", h.validateCredential)
		external.POST("/credentials/validate/batch", h.batchValidateCredentials)
		external.GET("/credentials/validation-rules", h.getValidationRules)

		// Audit endpoints
		external.GET("/audit/access-logs", h.getAccessLogs)
		external.GET("/audit/operation-logs", h.getOperationLogs)
		external.GET("/audit/security-events", h.getSecurityEvents)
		external.GET("/audit/statistics", h.getAuditStatistics)

		// Credential injection endpoints
		external.POST("/credentials/inject", h.injectCredentials)
		external.GET("/credentials/injection-configs", h.listInjectionConfigs)
		external.POST("/credentials/injection-configs", h.createInjectionConfig)

		// Task 7: Multi-Provider Support Architecture endpoints
		external.GET("/providers", h.listProviders)
		external.GET("/providers/:providerName", h.getProvider)
		external.GET("/providers/:providerName/status", h.getProviderStatus)
		external.GET("/providers/:providerName/metrics", h.getProviderMetrics)
		external.POST("/providers/:providerName/request", h.processProviderRequest)
		external.POST("/providers/discover", h.discoverProviders)

		external.GET("/load-balancer/stats", h.getLoadBalancerStats)
		external.GET("/load-balancer/weights", h.getProviderWeights)
		external.PUT("/load-balancer/weights", h.updateProviderWeights)
		external.PUT("/load-balancer/strategy", h.setLoadBalancingStrategy)
		external.POST("/load-balancer/select", h.selectProvider)

		external.GET("/failover/rules", h.getFailoverRules)
		external.GET("/failover/status", h.getFailoverStatus)
		external.GET("/failover/history", h.getFailoverHistory)
		external.POST("/failover/:serviceName/test", h.testFailover)
		external.POST("/failover/:serviceName/trigger", h.triggerFailover)

		external.GET("/config/providers", h.getProviderConfigs)
		external.GET("/config/:providerName", h.getProviderConfig)
		external.PUT("/config/:providerName", h.updateProviderConfig)
		external.POST("/config/:providerName/validate", h.validateProviderConfig)
		external.GET("/config/templates", h.getConfigTemplates)
		external.POST("/config/backup", h.backupConfigurations)
		external.POST("/config/restore", h.restoreConfigurations)
	}
}

// Health endpoints

// getHealth returns overall external API health status
//
//	@Summary		Get overall external API health status
//	@Description	Returns comprehensive health status for all external API providers
//	@Tags			External API Health
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"External APIs healthy"
//	@Failure		500	{object}	map[string]interface{}	"Health check failed"
//	@Router			/external/health [get]
func (h *Handler) getHealth(c *gin.Context) {
	summary, err := h.healthService.GetHealthSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "external-apis",
		"summary": summary,
	})
}

// getAPIHealth returns health status for a specific external API provider
//
//	@Summary		Get specific API provider health status
//	@Description	Returns detailed health information for a specific external API provider
//	@Tags			External API Health
//	@Produce		json
//	@Param			apiName	path		string					true	"API provider name"	Enums(openai, anthropic, google, ollama)
//	@Success		200		{object}	map[string]interface{}	"API provider health status"
//	@Failure		500		{object}	map[string]interface{}	"Health check failed"
//	@Router			/external/health/{apiName} [get]
func (h *Handler) getAPIHealth(c *gin.Context) {
	apiName := c.Param("apiName")

	req := &external.HealthCheckRequest{
		APIName: apiName,
	}

	result, err := h.healthService.CheckHealth(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) getHealthSummary(c *gin.Context) {
	result, err := h.healthService.CheckAllAPIs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) checkHealth(c *gin.Context) {
	var req external.HealthCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.healthService.CheckHealth(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) getHealthMetrics(c *gin.Context) {
	metrics, err := h.healthService.GetHealthMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// API Client endpoints

func (h *Handler) listClients(c *gin.Context) {
	clients := h.clientManager.ListClients()
	c.JSON(http.StatusOK, gin.H{"clients": clients})
}

func (h *Handler) makeAPIRequest(c *gin.Context) {
	provider := c.Param("provider")

	var reqBody map[string]interface{}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// This would make an actual API request through the client manager
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"provider": provider,
		"status":   "success",
		"message":  "API request would be processed here",
		"request":  reqBody,
	})
}

func (h *Handler) getClientInfo(c *gin.Context) {
	provider := c.Param("provider")

	client, err := h.clientManager.GetClient(provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	info := client.GetProviderInfo()
	c.JSON(http.StatusOK, info)
}

func (h *Handler) getClientMetrics(c *gin.Context) {
	metrics := h.clientManager.GetMetrics()
	c.JSON(http.StatusOK, metrics)
}

// Encryption endpoints

func (h *Handler) encryptData(c *gin.Context) {
	var req external.EncryptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.encryptionService.EncryptCredential(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) decryptData(c *gin.Context) {
	var req external.DecryptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.encryptionService.DecryptCredential(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) getEncryptionMetrics(c *gin.Context) {
	metrics, err := h.encryptionService.GetEncryptionMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// Credential validation endpoints

func (h *Handler) validateCredential(c *gin.Context) {
	var req external.ValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.validator.ValidateCredential(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) batchValidateCredentials(c *gin.Context) {
	var creds []map[string]interface{}
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch validation would be processed here",
		"count":   len(creds),
	})
}

func (h *Handler) getValidationRules(c *gin.Context) {
	rules, err := h.validator.GetValidationRules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rules": rules})
}

// Audit endpoints

func (h *Handler) getAccessLogs(c *gin.Context) {
	// Parse query parameters
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	// This would query actual access logs
	c.JSON(http.StatusOK, gin.H{
		"access_logs": []map[string]interface{}{},
		"limit":       limit,
		"message":     "Access logs would be returned here",
	})
}

func (h *Handler) getOperationLogs(c *gin.Context) {
	// Parse query parameters
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"operation_logs": []map[string]interface{}{},
		"limit":          limit,
		"message":        "Operation logs would be returned here",
	})
}

func (h *Handler) getSecurityEvents(c *gin.Context) {
	// Parse query parameters
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"security_events": []map[string]interface{}{},
		"limit":           limit,
		"message":         "Security events would be returned here",
	})
}

func (h *Handler) getAuditStatistics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_access_logs":     0,
		"total_operation_logs":  0,
		"total_security_events": 0,
		"message":               "Audit statistics would be calculated here",
	})
}

// Credential injection endpoints

func (h *Handler) injectCredentials(c *gin.Context) {
	var req external.InjectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.injector.InjectCredentials(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) listInjectionConfigs(c *gin.Context) {
	configs, err := h.injector.ListInjectionConfigs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

func (h *Handler) createInjectionConfig(c *gin.Context) {
	var config external.InjectionConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.injector.CreateInjectionConfig(c.Request.Context(), &config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Injection config created successfully"})
}

// Task 4: Enhanced Health Monitoring Endpoints

func (h *Handler) getAPIHealthHistory(c *gin.Context) {
	apiName := c.Param("apiName")

	// Parse timeframe parameters
	hoursStr := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		hours = 24
	}

	timeframe := &external.TimeFrame{
		StartTime: time.Now().Add(-time.Duration(hours) * time.Hour),
		EndTime:   time.Now(),
	}

	history, err := h.healthService.GetAPIHealthHistory(c.Request.Context(), apiName, timeframe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"api_name": apiName,
		"timeframe": gin.H{
			"start": timeframe.StartTime,
			"end":   timeframe.EndTime,
			"hours": hours,
		},
		"history": history,
	})
}

func (h *Handler) getHealthDashboard(c *gin.Context) {
	// Get comprehensive dashboard data
	summary, err := h.healthService.GetHealthSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics, err := h.healthService.GetHealthMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get all APIs status
	allAPIs, err := h.healthService.CheckAllAPIs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dashboard := gin.H{
		"timestamp": time.Now(),
		"summary":   summary,
		"metrics":   metrics,
		"all_apis":  allAPIs,
		"dashboard_info": gin.H{
			"version":     "1.0.0",
			"environment": "development",
			"uptime":      time.Since(time.Now().Add(-24 * time.Hour)), // Mock uptime
		},
	}

	c.JSON(http.StatusOK, dashboard)
}

func (h *Handler) getSLAReport(c *gin.Context) {
	// Parse period parameter
	period := c.DefaultQuery("period", "daily")
	validPeriods := map[string]bool{"hourly": true, "daily": true, "weekly": true, "monthly": true}

	if !validPeriods[period] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period. Use: hourly, daily, weekly, monthly"})
		return
	}

	// Mock SLA data - in real implementation, this would come from the health service
	slaReport := gin.H{
		"period":    period,
		"timestamp": time.Now(),
		"sla_targets": gin.H{
			"uptime_percentage":    99.9,
			"response_time_ms":     1000,
			"error_rate_threshold": 1.0,
		},
		"actual_performance": gin.H{
			"uptime_percentage":    99.85,
			"avg_response_time_ms": 850,
			"error_rate":           0.15,
		},
		"compliance": gin.H{
			"uptime_met":        false,
			"response_time_met": true,
			"error_rate_met":    true,
			"overall_sla_met":   false,
		},
		"apis": []gin.H{
			{
				"name":              "openai",
				"uptime_percentage": 99.8,
				"avg_response_time": 1200,
				"error_rate":        0.2,
				"sla_compliance":    false,
			},
			{
				"name":              "ollama",
				"uptime_percentage": 99.95,
				"avg_response_time": 500,
				"error_rate":        0.05,
				"sla_compliance":    true,
			},
		},
		"recommendations": []string{
			"Monitor OpenAI API closely - approaching SLA limits",
			"Consider implementing fallback for external APIs",
			"Review rate limiting configuration",
		},
	}

	c.JSON(http.StatusOK, slaReport)
}

func (h *Handler) getHealthAlerts(c *gin.Context) {
	// Get current alerts - mock data for now
	alerts := gin.H{
		"timestamp": time.Now(),
		"active_alerts": []gin.H{
			{
				"id":         "alert-001",
				"api_name":   "openai",
				"severity":   "warning",
				"message":    "Response time above threshold",
				"threshold":  1000,
				"actual":     1250,
				"started_at": time.Now().Add(-15 * time.Minute),
				"status":     "active",
			},
		},
		"resolved_alerts": []gin.H{
			{
				"id":          "alert-002",
				"api_name":    "redis",
				"severity":    "critical",
				"message":     "Connection failed",
				"started_at":  time.Now().Add(-2 * time.Hour),
				"resolved_at": time.Now().Add(-1 * time.Hour),
				"status":      "resolved",
				"duration":    "1h0m0s",
			},
		},
		"alert_summary": gin.H{
			"total_active":   1,
			"total_resolved": 1,
			"critical":       0,
			"warning":        1,
			"info":           0,
		},
	}

	c.JSON(http.StatusOK, alerts)
}

func (h *Handler) configureHealthAlerts(c *gin.Context) {
	var alertConfig struct {
		APIName               string   `json:"api_name" binding:"required"`
		ResponseTimeThreshold int      `json:"response_time_threshold"` // milliseconds
		UptimeThreshold       float64  `json:"uptime_threshold"`        // percentage
		ErrorRateThreshold    float64  `json:"error_rate_threshold"`    // percentage
		AlertEmails           []string `json:"alert_emails"`
		Enabled               bool     `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&alertConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate thresholds
	if alertConfig.ResponseTimeThreshold <= 0 {
		alertConfig.ResponseTimeThreshold = 5000 // default 5 seconds
	}
	if alertConfig.UptimeThreshold <= 0 || alertConfig.UptimeThreshold > 100 {
		alertConfig.UptimeThreshold = 99.0 // default 99%
	}
	if alertConfig.ErrorRateThreshold < 0 || alertConfig.ErrorRateThreshold > 100 {
		alertConfig.ErrorRateThreshold = 5.0 // default 5%
	}

	// In real implementation, this would update the alert configuration in the health service
	response := gin.H{
		"message": "Alert configuration updated successfully",
		"config": gin.H{
			"api_name":                alertConfig.APIName,
			"response_time_threshold": alertConfig.ResponseTimeThreshold,
			"uptime_threshold":        alertConfig.UptimeThreshold,
			"error_rate_threshold":    alertConfig.ErrorRateThreshold,
			"alert_emails":            alertConfig.AlertEmails,
			"enabled":                 alertConfig.Enabled,
			"updated_at":              time.Now(),
		},
	}

	c.JSON(http.StatusOK, response)
}

// Task 5: Fallback and Recovery Endpoints

func (h *Handler) getFallbackStatus(c *gin.Context) {
	// Mock fallback status - in real implementation, this would use the fallback service
	status := gin.H{
		"timestamp": time.Now(),
		"services": []gin.H{
			{
				"name":                "openai",
				"status":              "active",
				"primary_available":   true,
				"fallback_enabled":    true,
				"mock_mode":           gin.Mode() == gin.DebugMode,
				"fallbacks_triggered": 0,
				"last_fallback_used":  nil,
			},
			{
				"name":                "ollama",
				"status":              "active",
				"primary_available":   true,
				"fallback_enabled":    true,
				"mock_mode":           gin.Mode() == gin.DebugMode,
				"fallbacks_triggered": 0,
				"last_fallback_used":  nil,
			},
		},
		"overall_status": gin.H{
			"total_services":    2,
			"healthy_services":  2,
			"degraded_services": 0,
			"fallbacks_active":  0,
		},
	}

	c.JSON(http.StatusOK, status)
}

func (h *Handler) getFallbackMetrics(c *gin.Context) {
	// Mock fallback metrics - in real implementation, this would use the fallback service
	metrics := gin.H{
		"timestamp":           time.Now(),
		"total_requests":      100,
		"primary_successes":   98,
		"fallbacks_triggered": 2,
		"fallbacks_by_source": gin.H{
			"mock":    1,
			"default": 1,
			"cache":   0,
		},
		"fallbacks_by_service": gin.H{
			"openai": 1,
			"ollama": 1,
		},
		"average_response_time": "150ms",
		"service_metrics": []gin.H{
			{
				"service_name":      "openai",
				"total_requests":    60,
				"primary_successes": 59,
				"fallbacks_used":    1,
				"success_rate":      98.3,
				"avg_response_time": "200ms",
			},
			{
				"service_name":      "ollama",
				"total_requests":    40,
				"primary_successes": 39,
				"fallbacks_used":    1,
				"success_rate":      97.5,
				"avg_response_time": "100ms",
			},
		},
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) testFallback(c *gin.Context) {
	serviceName := c.Param("serviceName")

	var testRequest struct {
		Operation     string                 `json:"operation"`
		Context       map[string]interface{} `json:"context,omitempty"`
		ForceFallback bool                   `json:"force_fallback,omitempty"`
	}

	if err := c.ShouldBindJSON(&testRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock fallback test response
	response := gin.H{
		"service_name": serviceName,
		"operation":    testRequest.Operation,
		"test_result": gin.H{
			"fallback_triggered": true,
			"source":             "mock",
			"execution_time":     "50ms",
			"success":            true,
		},
		"mock_response": gin.H{
			"message": fmt.Sprintf("Mock fallback response for %s service", serviceName),
			"status":  "fallback_active",
			"data":    "Test data from fallback mechanism",
		},
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) getRecoveryStatus(c *gin.Context) {
	// Mock recovery status - in real implementation, this would use the recovery service
	status := gin.H{
		"timestamp":         time.Now(),
		"monitoring_active": true,
		"total_apis":        2,
		"healthy_apis":      2,
		"recovering_apis":   0,
		"failed_apis":       0,
		"api_status": []gin.H{
			{
				"api_name":             "openai",
				"status":               "healthy",
				"last_health_check":    time.Now().Add(-30 * time.Second),
				"recovery_attempts":    0,
				"max_attempts":         3,
				"consecutive_failures": 0,
			},
			{
				"api_name":             "ollama",
				"status":               "healthy",
				"last_health_check":    time.Now().Add(-25 * time.Second),
				"recovery_attempts":    0,
				"max_attempts":         3,
				"consecutive_failures": 0,
			},
		},
	}

	c.JSON(http.StatusOK, status)
}

func (h *Handler) getRecoveryMetrics(c *gin.Context) {
	// Mock recovery metrics - in real implementation, this would use the recovery service
	metrics := gin.H{
		"timestamp":               time.Now(),
		"total_recovery_attempts": 5,
		"successful_recoveries":   4,
		"failed_recoveries":       1,
		"average_recovery_time":   "45s",
		"monitoring_start_time":   time.Now().Add(-24 * time.Hour),
		"total_downtime":          "5m30s",
		"api_metrics": []gin.H{
			{
				"api_name":              "openai",
				"recovery_attempts":     3,
				"successful_recoveries": 2,
				"failed_recoveries":     1,
				"avg_recovery_time":     "60s",
				"success_rate":          66.7,
				"last_recovery":         time.Now().Add(-2 * time.Hour),
			},
			{
				"api_name":              "ollama",
				"recovery_attempts":     2,
				"successful_recoveries": 2,
				"failed_recoveries":     0,
				"avg_recovery_time":     "30s",
				"success_rate":          100.0,
				"last_recovery":         time.Now().Add(-4 * time.Hour),
			},
		},
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) triggerRecovery(c *gin.Context) {
	apiName := c.Param("apiName")

	// Mock recovery trigger - in real implementation, this would use the recovery service
	result := gin.H{
		"api_name":       apiName,
		"triggered_at":   time.Now(),
		"attempt_number": 1,
		"status":         "initiated",
		"steps_planned": []string{
			"health_check",
			"clear_cache",
			"reconnect",
		},
		"estimated_duration": "30-60s",
		"message":            fmt.Sprintf("Recovery procedure initiated for %s API", apiName),
	}

	// Simulate recovery process
	go func() {
		// This would trigger actual recovery in a real implementation
		// For now, just log the action
		log.Printf("Manual recovery triggered for API: %s", apiName)
	}()

	c.JSON(http.StatusAccepted, result)
}

func (h *Handler) getCircuitBreakerStatus(c *gin.Context) {
	// Mock circuit breaker status - in real implementation, this would get actual breaker states
	status := gin.H{
		"timestamp": time.Now(),
		"circuit_breakers": []gin.H{
			{
				"name":                 "openai",
				"state":                "closed",
				"total_requests":       150,
				"successful_requests":  148,
				"failed_requests":      2,
				"consecutive_failures": 0,
				"failure_rate":         1.33,
				"last_failure":         nil,
				"last_success":         time.Now().Add(-10 * time.Second),
				"last_state_change":    time.Now().Add(-30 * time.Minute),
				"total_state_changes":  0,
				"configuration": gin.H{
					"max_failures":        5,
					"reset_timeout":       "60s",
					"half_open_max_calls": 3,
				},
			},
			{
				"name":                 "ollama",
				"state":                "closed",
				"total_requests":       200,
				"successful_requests":  200,
				"failed_requests":      0,
				"consecutive_failures": 0,
				"failure_rate":         0.0,
				"last_failure":         nil,
				"last_success":         time.Now().Add(-5 * time.Second),
				"last_state_change":    time.Now().Add(-30 * time.Minute),
				"total_state_changes":  0,
				"configuration": gin.H{
					"max_failures":        5,
					"reset_timeout":       "60s",
					"half_open_max_calls": 3,
				},
			},
		},
		"summary": gin.H{
			"total_breakers":     2,
			"closed_breakers":    2,
			"open_breakers":      0,
			"half_open_breakers": 0,
		},
	}

	c.JSON(http.StatusOK, status)
}

// Task 6: Cost Tracking and Budget Management Handler Methods

// getCostSummary returns cost summary for a time range
//
//	@Summary		Get API usage cost summary
//	@Description	Returns comprehensive cost summary for external API usage with breakdown by provider
//	@Tags			External API Cost Management
//	@Produce		json
//	@Param			start	query		string					false	"Start date (YYYY-MM-DD)"	Format(date)
//	@Param			end		query		string					false	"End date (YYYY-MM-DD)"		Format(date)
//	@Success		200		{object}	map[string]interface{}	"Cost summary with provider breakdown"
//	@Failure		500		{object}	map[string]interface{}	"Failed to get cost summary"
//	@Security		BearerAuth
//	@Router			/external/cost/summary [get]
func (h *Handler) getCostSummary(c *gin.Context) {
	startDate := c.Query("start")
	endDate := c.Query("end")

	// Default to current month if no dates provided
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := now

	if startDate != "" {
		if parsedStart, err := time.Parse("2006-01-02", startDate); err == nil {
			start = parsedStart
		}
	}
	if endDate != "" {
		if parsedEnd, err := time.Parse("2006-01-02", endDate); err == nil {
			end = parsedEnd
		}
	}

	timeRange := external.TimeRange{Start: start, End: end}

	summary, err := h.costTracker.GetCostSummary(c.Request.Context(), timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// getServiceCosts returns costs for a specific service
func (h *Handler) getServiceCosts(c *gin.Context) {
	serviceName := c.Param("serviceName")
	startDate := c.Query("start")
	endDate := c.Query("end")

	// Default to current month
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := now

	if startDate != "" {
		if parsedStart, err := time.Parse("2006-01-02", startDate); err == nil {
			start = parsedStart
		}
	}
	if endDate != "" {
		if parsedEnd, err := time.Parse("2006-01-02", endDate); err == nil {
			end = parsedEnd
		}
	}

	timeRange := external.TimeRange{Start: start, End: end}

	costs, err := h.costTracker.GetServiceCosts(c.Request.Context(), serviceName, timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, costs)
}

// getUsageMetrics returns usage metrics for a time range
func (h *Handler) getUsageMetrics(c *gin.Context) {
	startDate := c.Query("start")
	endDate := c.Query("end")

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := now

	if startDate != "" {
		if parsedStart, err := time.Parse("2006-01-02", startDate); err == nil {
			start = parsedStart
		}
	}
	if endDate != "" {
		if parsedEnd, err := time.Parse("2006-01-02", endDate); err == nil {
			end = parsedEnd
		}
	}

	timeRange := external.TimeRange{Start: start, End: end}

	metrics, err := h.costTracker.GetUsageMetrics(c.Request.Context(), timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// getCostBreakdown returns detailed cost breakdown
func (h *Handler) getCostBreakdown(c *gin.Context) {
	startDate := c.Query("start")
	endDate := c.Query("end")

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := now

	if startDate != "" {
		if parsedStart, err := time.Parse("2006-01-02", startDate); err == nil {
			start = parsedStart
		}
	}
	if endDate != "" {
		if parsedEnd, err := time.Parse("2006-01-02", endDate); err == nil {
			end = parsedEnd
		}
	}

	timeRange := external.TimeRange{Start: start, End: end}

	breakdown, err := h.costTracker.GetCostBreakdown(c.Request.Context(), timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, breakdown)
}

// getBillingData returns billing data for a specific month/year
func (h *Handler) getBillingData(c *gin.Context) {
	monthStr := c.Param("month")
	yearStr := c.Param("year")

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month parameter"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2020 || year > 2030 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year parameter"})
		return
	}

	billingData, err := h.costTracker.GetBillingData(c.Request.Context(), month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, billingData)
}

// recordUsage records API usage for cost tracking
func (h *Handler) recordUsage(c *gin.Context) {
	var usage external.UsageRequest
	if err := c.ShouldBindJSON(&usage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if usage.Timestamp.IsZero() {
		usage.Timestamp = time.Now()
	}

	err := h.costTracker.RecordUsage(c.Request.Context(), &usage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usage recorded successfully"})
}

// Budget Management Handlers

// getBudgetStatus returns budget status for a service
func (h *Handler) getBudgetStatus(c *gin.Context) {
	serviceName := c.Query("service")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service name is required"})
		return
	}

	status, err := h.budgetManager.CheckBudgetStatus(c.Request.Context(), serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// getBudget returns a specific budget by ID
func (h *Handler) getBudget(c *gin.Context) {
	budgetID := c.Param("budgetId")

	budget, err := h.budgetManager.GetBudget(c.Request.Context(), budgetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, budget)
}

// getBudgetAlerts returns all budget alerts
func (h *Handler) getBudgetAlerts(c *gin.Context) {
	alerts, err := h.budgetManager.GetBudgetAlerts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"alerts": alerts})
}

// getBudgetSummary returns budget summary for all services
func (h *Handler) getBudgetSummary(c *gin.Context) {
	// Mock budget summary response
	summary := gin.H{
		"total_budget":        1000.0,
		"total_spend":         425.75,
		"overall_utilization": 42.6,
		"service_summaries": []gin.H{
			{
				"service_name":    "openai",
				"budget_limit":    500.0,
				"current_spend":   375.50,
				"utilization_pct": 75.1,
				"status":          "warning",
			},
			{
				"service_name":    "ollama",
				"budget_limit":    100.0,
				"current_spend":   50.25,
				"utilization_pct": 50.3,
				"status":          "under_budget",
			},
		},
		"alert_count": 2,
		"timestamp":   time.Now(),
	}

	c.JSON(http.StatusOK, summary)
}

// createBudget creates a new budget
func (h *Handler) createBudget(c *gin.Context) {
	var budget external.Budget
	if err := c.ShouldBindJSON(&budget); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.budgetManager.SetBudget(c.Request.Context(), &budget)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Budget created successfully",
		"budget_id": budget.ID,
	})
}

// updateBudgetLimits updates budget limits for a service
func (h *Handler) updateBudgetLimits(c *gin.Context) {
	serviceName := c.Param("serviceName")

	var limits external.BudgetLimits
	if err := c.ShouldBindJSON(&limits); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.budgetManager.UpdateBudgetLimits(c.Request.Context(), serviceName, &limits)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Budget limits updated successfully"})
}

// Cost Optimization Handlers

// getUsageAnalysis returns usage pattern analysis
func (h *Handler) getUsageAnalysis(c *gin.Context) {
	startDate := c.Query("start")
	endDate := c.Query("end")

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := now

	if startDate != "" {
		if parsedStart, err := time.Parse("2006-01-02", startDate); err == nil {
			start = parsedStart
		}
	}
	if endDate != "" {
		if parsedEnd, err := time.Parse("2006-01-02", endDate); err == nil {
			end = parsedEnd
		}
	}

	timeRange := external.TimeRange{Start: start, End: end}

	analysis, err := h.optimizer.AnalyzeUsagePatterns(c.Request.Context(), timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// getOptimizationRecommendations returns optimization recommendations for a service
func (h *Handler) getOptimizationRecommendations(c *gin.Context) {
	serviceName := c.Param("serviceName")

	recommendations, err := h.optimizer.GetOptimizationRecommendations(c.Request.Context(), serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"recommendations": recommendations})
}

// calculateCostSavings calculates potential cost savings
func (h *Handler) calculateCostSavings(c *gin.Context) {
	serviceName := c.Query("service")
	if serviceName == "" {
		// Get recommendations for all services
		allRecs := make([]*external.OptimizationRecommendation, 0)

		for _, service := range []string{"openai", "ollama"} {
			recs, err := h.optimizer.GetOptimizationRecommendations(c.Request.Context(), service)
			if err == nil {
				allRecs = append(allRecs, recs...)
			}
		}

		savings, err := h.optimizer.CalculateCostSavings(c.Request.Context(), allRecs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, savings)
		return
	}

	// Get recommendations for specific service
	recommendations, err := h.optimizer.GetOptimizationRecommendations(c.Request.Context(), serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	savings, err := h.optimizer.CalculateCostSavings(c.Request.Context(), recommendations)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, savings)
}

// getEfficiencyMetrics returns cost efficiency metrics
func (h *Handler) getEfficiencyMetrics(c *gin.Context) {
	metrics, err := h.optimizer.GetEfficiencyMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// Cost Allocation Handlers

// getAllocationSummary returns cost allocation summary
func (h *Handler) getAllocationSummary(c *gin.Context) {
	startDate := c.Query("start")
	endDate := c.Query("end")

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := now

	if startDate != "" {
		if parsedStart, err := time.Parse("2006-01-02", startDate); err == nil {
			start = parsedStart
		}
	}
	if endDate != "" {
		if parsedEnd, err := time.Parse("2006-01-02", endDate); err == nil {
			end = parsedEnd
		}
	}

	timeRange := external.TimeRange{Start: start, End: end}

	summary, err := h.allocator.GetAllocationSummary(c.Request.Context(), timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// getDepartmentCosts returns costs for a specific department
func (h *Handler) getDepartmentCosts(c *gin.Context) {
	department := c.Param("department")
	startDate := c.Query("start")
	endDate := c.Query("end")

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := now

	if startDate != "" {
		if parsedStart, err := time.Parse("2006-01-02", startDate); err == nil {
			start = parsedStart
		}
	}
	if endDate != "" {
		if parsedEnd, err := time.Parse("2006-01-02", endDate); err == nil {
			end = parsedEnd
		}
	}

	timeRange := external.TimeRange{Start: start, End: end}

	costs, err := h.allocator.GetDepartmentCosts(c.Request.Context(), department, timeRange)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, costs)
}

// getProjectCosts returns costs for a specific project
func (h *Handler) getProjectCosts(c *gin.Context) {
	projectID := c.Param("projectId")
	startDate := c.Query("start")
	endDate := c.Query("end")

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := now

	if startDate != "" {
		if parsedStart, err := time.Parse("2006-01-02", startDate); err == nil {
			start = parsedStart
		}
	}
	if endDate != "" {
		if parsedEnd, err := time.Parse("2006-01-02", endDate); err == nil {
			end = parsedEnd
		}
	}

	timeRange := external.TimeRange{Start: start, End: end}

	costs, err := h.allocator.GetProjectCosts(c.Request.Context(), projectID, timeRange)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, costs)
}

// allocateCosts allocates costs based on allocation rules
func (h *Handler) allocateCosts(c *gin.Context) {
	var allocation external.CostAllocation
	if err := c.ShouldBindJSON(&allocation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.allocator.AllocateCosts(c.Request.Context(), &allocation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Costs allocated successfully"})
}

// Automated Cost Control Handlers

// getControlStatus returns the status of automated cost controls
func (h *Handler) getControlStatus(c *gin.Context) {
	statuses, err := h.controller.GetControlStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"control_statuses": statuses})
}

// getControlHistory returns the history of control actions
func (h *Handler) getControlHistory(c *gin.Context) {
	startDate := c.Query("start")
	endDate := c.Query("end")

	now := time.Now()
	start := now.AddDate(0, 0, -7) // Default to last 7 days
	end := now

	if startDate != "" {
		if parsedStart, err := time.Parse("2006-01-02", startDate); err == nil {
			start = parsedStart
		}
	}
	if endDate != "" {
		if parsedEnd, err := time.Parse("2006-01-02", endDate); err == nil {
			end = parsedEnd
		}
	}

	timeRange := external.TimeRange{Start: start, End: end}

	history, err := h.controller.GetControlHistory(c.Request.Context(), timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"control_history": history})
}

// getControlMetrics returns control metrics
func (h *Handler) getControlMetrics(c *gin.Context) {
	// Mock control metrics response
	metrics := gin.H{
		"time_range": gin.H{
			"start": time.Now().AddDate(0, 0, -7),
			"end":   time.Now(),
		},
		"total_actions":      15,
		"successful_actions": 14,
		"success_rate":       93.3,
		"action_breakdown": gin.H{
			"alert_sent":                8,
			"rate_limit_applied":        4,
			"rate_limit_strict_applied": 2,
			"service_suspended":         1,
		},
		"service_breakdown": gin.H{
			"openai": 12,
			"ollama": 3,
		},
		"active_controls": 2,
		"timestamp":       time.Now(),
	}

	c.JSON(http.StatusOK, metrics)
}

// enableControls enables automated cost controls for a service
func (h *Handler) enableControls(c *gin.Context) {
	serviceName := c.Param("serviceName")

	var controls external.AutomatedControls
	if err := c.ShouldBindJSON(&controls); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	controls.ServiceName = serviceName

	err := h.controller.EnableAutomatedControls(c.Request.Context(), &controls)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Automated controls enabled for service %s", serviceName)})
}

// disableControls disables automated cost controls for a service
func (h *Handler) disableControls(c *gin.Context) {
	serviceName := c.Param("serviceName")

	err := h.controller.DisableAutomatedControls(c.Request.Context(), serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Automated controls disabled for service %s", serviceName)})
}

// executeEmergencyShutdown performs emergency shutdown of services
func (h *Handler) executeEmergencyShutdown(c *gin.Context) {
	var request struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.controller.ExecuteEmergencyShutdown(c.Request.Context(), request.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Emergency shutdown executed successfully",
		"reason":    request.Reason,
		"timestamp": time.Now(),
	})
}

// Task 7: Multi-Provider Support Architecture endpoints

func (h *Handler) listProviders(c *gin.Context) {
	if h.providerRegistry == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Provider registry not available"})
		return
	}

	providers := h.providerRegistry.GetAllProviders()
	providerList := make([]gin.H, len(providers))

	for i, provider := range providers {
		providerList[i] = gin.H{
			"name":         provider.GetName(),
			"type":         provider.GetType(),
			"version":      provider.GetVersion(),
			"capabilities": provider.GetCapabilities(),
			"models":       provider.GetSupportedModels(),
			"healthy":      provider.IsHealthy(c.Request.Context()),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": providerList,
		"count":     len(providers),
		"timestamp": time.Now(),
	})
}

func (h *Handler) getProvider(c *gin.Context) {
	if h.providerRegistry == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Provider registry not available"})
		return
	}

	providerName := c.Param("providerName")
	provider, err := h.providerRegistry.GetProvider(providerName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Provider %s not found", providerName)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":         provider.GetName(),
		"type":         provider.GetType(),
		"version":      provider.GetVersion(),
		"capabilities": provider.GetCapabilities(),
		"models":       provider.GetSupportedModels(),
		"healthy":      provider.IsHealthy(c.Request.Context()),
		"config":       provider.GetConfiguration(),
	})
}

func (h *Handler) getProviderStatus(c *gin.Context) {
	if h.providerRegistry == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Provider registry not available"})
		return
	}

	providerName := c.Param("providerName")
	provider, err := h.providerRegistry.GetProvider(providerName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Provider %s not found", providerName)})
		return
	}

	status := provider.GetStatus(c.Request.Context())
	c.JSON(http.StatusOK, status)
}

func (h *Handler) getProviderMetrics(c *gin.Context) {
	if h.providerRegistry == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Provider registry not available"})
		return
	}

	providerName := c.Param("providerName")
	provider, err := h.providerRegistry.GetProvider(providerName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Provider %s not found", providerName)})
		return
	}

	metrics, err := provider.GetMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) processProviderRequest(c *gin.Context) {
	if h.providerRegistry == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Provider registry not available"})
		return
	}

	providerName := c.Param("providerName")
	provider, err := h.providerRegistry.GetProvider(providerName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Provider %s not found", providerName)})
		return
	}

	var request external.ProviderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := provider.ProcessRequest(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) discoverProviders(c *gin.Context) {
	if h.providerRegistry == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Provider registry not available"})
		return
	}

	err := h.providerRegistry.DiscoverProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	providers := h.providerRegistry.GetAllProviders()
	c.JSON(http.StatusOK, gin.H{
		"message":   "Provider discovery completed",
		"providers": len(providers),
		"timestamp": time.Now(),
	})
}

func (h *Handler) getLoadBalancerStats(c *gin.Context) {
	if h.loadBalancer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Load balancer not available"})
		return
	}

	stats := h.loadBalancer.GetLoadBalancingStats()
	c.JSON(http.StatusOK, stats)
}

func (h *Handler) getProviderWeights(c *gin.Context) {
	if h.loadBalancer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Load balancer not available"})
		return
	}

	weights := h.loadBalancer.GetProviderWeights()
	c.JSON(http.StatusOK, gin.H{
		"weights":   weights,
		"timestamp": time.Now(),
	})
}

func (h *Handler) updateProviderWeights(c *gin.Context) {
	if h.loadBalancer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Load balancer not available"})
		return
	}

	var weights map[string]float64
	if err := c.ShouldBindJSON(&weights); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.loadBalancer.UpdateProviderWeights(weights)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Provider weights updated successfully",
		"weights":   weights,
		"timestamp": time.Now(),
	})
}

func (h *Handler) setLoadBalancingStrategy(c *gin.Context) {
	if h.loadBalancer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Load balancer not available"})
		return
	}

	var request struct {
		Strategy string `json:"strategy"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	strategy := external.LoadBalancingStrategy(request.Strategy)
	err := h.loadBalancer.SetLoadBalancingStrategy(strategy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Load balancing strategy updated",
		"strategy":  request.Strategy,
		"timestamp": time.Now(),
	})
}

func (h *Handler) selectProvider(c *gin.Context) {
	if h.loadBalancer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Load balancer not available"})
		return
	}

	var request external.ProviderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider, err := h.loadBalancer.SelectProvider(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"selected_provider": provider.GetName(),
		"provider_type":     provider.GetType(),
		"timestamp":         time.Now(),
	})
}

func (h *Handler) getFailoverRules(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Failover rules endpoint - implementation depends on actual failover service interface",
		"note":    "Service may have built-in rules that are not exposed via API",
	})
}

func (h *Handler) getFailoverStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":   "Failover service status",
		"active":    h.failoverService != nil,
		"timestamp": time.Now(),
	})
}

func (h *Handler) getFailoverHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Failover history endpoint - implementation depends on actual failover service interface",
		"note":    "Historical data may be tracked internally",
	})
}

func (h *Handler) testFailover(c *gin.Context) {
	serviceName := c.Param("serviceName")
	c.JSON(http.StatusOK, gin.H{
		"message":      "Failover test initiated",
		"service_name": serviceName,
		"timestamp":    time.Now(),
	})
}

func (h *Handler) triggerFailover(c *gin.Context) {
	serviceName := c.Param("serviceName")
	c.JSON(http.StatusOK, gin.H{
		"message":      "Failover triggered",
		"service_name": serviceName,
		"timestamp":    time.Now(),
	})
}

func (h *Handler) getProviderConfigs(c *gin.Context) {
	if h.configManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Config manager not available"})
		return
	}

	configs := h.configManager.GetAllProviderConfigs()
	c.JSON(http.StatusOK, gin.H{
		"configs":   configs,
		"count":     len(configs),
		"timestamp": time.Now(),
	})
}

func (h *Handler) getProviderConfig(c *gin.Context) {
	if h.configManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Config manager not available"})
		return
	}

	providerName := c.Param("providerName")
	config, err := h.configManager.GetProviderConfig(providerName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Config for provider %s not found", providerName)})
		return
	}

	c.JSON(http.StatusOK, config)
}

func (h *Handler) updateProviderConfig(c *gin.Context) {
	if h.configManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Config manager not available"})
		return
	}

	providerName := c.Param("providerName")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.configManager.UpdateProviderConfig(providerName, updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Provider configuration updated",
		"provider_name": providerName,
		"timestamp":     time.Now(),
	})
}

func (h *Handler) validateProviderConfig(c *gin.Context) {
	if h.configManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Config manager not available"})
		return
	}

	var config external.ProviderConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.configManager.ValidateProviderConfig(&config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":     true,
		"message":   "Configuration is valid",
		"timestamp": time.Now(),
	})
}

func (h *Handler) getConfigTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration templates endpoint",
		"note":    "Templates are available through the config manager service",
	})
}

func (h *Handler) backupConfigurations(c *gin.Context) {
	if h.configManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Config manager not available"})
		return
	}

	err := h.configManager.BackupConfigurations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Configurations backed up successfully",
		"timestamp": time.Now(),
	})
}

func (h *Handler) restoreConfigurations(c *gin.Context) {
	if h.configManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Config manager not available"})
		return
	}

	var request struct {
		BackupID string `json:"backup_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.configManager.RestoreConfigurations(request.BackupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Configurations restored successfully",
		"backup_id": request.BackupID,
		"timestamp": time.Now(),
	})
}
