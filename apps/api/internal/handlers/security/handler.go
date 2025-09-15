package security

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	securityService "github.com/pramodksahoo/kubechat/apps/api/internal/services/security"
)

// Handler handles HTTP requests for security and performance management
type Handler struct {
	service securityService.Service
}

// NewHandler creates a new security handler
func NewHandler(service securityService.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers the security and performance routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	secRoutes := router.Group("/security")
	{
		// Security Management endpoints
		secRoutes.POST("/validate-password", h.ValidatePassword)
		secRoutes.POST("/hash-password", h.HashPassword)
		secRoutes.POST("/generate-token", h.GenerateToken)
		secRoutes.POST("/validate-headers", h.ValidateSecurityHeaders)
		secRoutes.POST("/analyze-request", h.AnalyzeRequest)

		// Session Management endpoints
		secRoutes.POST("/sessions", h.CreateSession)
		secRoutes.GET("/sessions/:token", h.ValidateSession)
		secRoutes.DELETE("/sessions/:token", h.RevokeSession)
		secRoutes.DELETE("/sessions/cleanup", h.CleanupSessions)

		// Rate Limiting endpoints
		secRoutes.GET("/rate-limit/:identifier", h.GetRateLimitStatus)
		secRoutes.POST("/rate-limit/:identifier/check", h.CheckRateLimit)
		secRoutes.DELETE("/rate-limit/:identifier", h.ResetRateLimit)

		// Security Events endpoints
		secRoutes.POST("/events", h.RecordSecurityEvent)
		secRoutes.GET("/events", h.GetSecurityEvents)
		secRoutes.GET("/alerts", h.GetSecurityAlerts)

		// Security Scanning endpoints
		secRoutes.POST("/scan", h.StartSecurityScan)
		secRoutes.POST("/validate-input", h.ValidateInput)
		secRoutes.GET("/ip-reputation/:ip", h.CheckIPReputation)

		// Security Health endpoints
		secRoutes.GET("/health", h.GetSecurityHealth)
	}

	perfRoutes := router.Group("/performance")
	{
		// Performance Monitoring endpoints
		perfRoutes.POST("/metrics", h.RecordPerformanceMetric)
		perfRoutes.GET("/metrics", h.GetPerformanceMetrics)
		perfRoutes.GET("/metrics/:service", h.GetServiceStats)

		// Cache Management endpoints
		perfRoutes.GET("/cache/:key", h.GetCacheValue)
		perfRoutes.POST("/cache/:key", h.SetCacheValue)
		perfRoutes.DELETE("/cache/:key", h.DeleteCacheValue)
		perfRoutes.GET("/cache/stats", h.GetCacheStats)
		perfRoutes.POST("/cache/warmup", h.WarmupCache)

		// Performance Optimization endpoints
		perfRoutes.POST("/optimize", h.OptimizePerformance)
		perfRoutes.POST("/compression/enable", h.EnableCompression)
		perfRoutes.POST("/connections/optimize", h.OptimizeConnections)
		perfRoutes.GET("/resource-usage", h.GetResourceUsage)

		// Performance Health endpoints
		perfRoutes.GET("/health", h.GetPerformanceHealth)
	}
}

// ValidatePassword validates password against security policy
//
//	@Summary		Validate password against policy
//	@Description	Validates a password against the configured security policy
//	@Tags			Security & Authorization
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{password=string,policy=models.PasswordPolicy}	true	"Password validation request"
//	@Success		200		{object}	map[string]interface{}									"Password validation result"
//	@Failure		400		{object}	map[string]interface{}									"Invalid password or policy"
//	@Security		BearerAuth
//	@Router			/security/validate-password [post]
func (h *Handler) ValidatePassword(c *gin.Context) {
	var request struct {
		Password string                 `json:"password" binding:"required"`
		Policy   *models.PasswordPolicy `json:"policy"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Use default policy if not provided
	if request.Policy == nil {
		request.Policy = &models.PasswordPolicy{
			MinLength:           8,
			RequireUppercase:    true,
			RequireLowercase:    true,
			RequireNumbers:      true,
			RequireSpecialChars: true,
		}
	}

	if err := h.service.ValidatePassword(request.Password, request.Policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "Password meets security requirements",
	})
}

// HashPassword creates a secure hash of the password
//
//	@Summary		Hash password securely
//	@Description	Creates a secure bcrypt hash of the provided password
//	@Tags			Security & Authorization
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{password=string}	true	"Password to hash"
//	@Success		200		{object}	map[string]interface{}	"Password hashed successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request format"
//	@Failure		500		{object}	map[string]interface{}	"Failed to hash password"
//	@Security		BearerAuth
//	@Router			/security/hash-password [post]
func (h *Handler) HashPassword(c *gin.Context) {
	var request struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	hash, err := h.service.HashPassword(request.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to hash password",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hash":      hash,
		"algorithm": "bcrypt",
		"timestamp": time.Now(),
	})
}

// GenerateToken generates a secure token
//
//	@Summary		Generate secure token
//	@Description	Generates a cryptographically secure random token
//	@Tags			Security & Authorization
//	@Produce		json
//	@Param			length	query		int						false	"Token length (1-128)"	default(32)
//	@Success		200		{object}	map[string]interface{}	"Secure token generated"
//	@Failure		400		{object}	map[string]interface{}	"Invalid length parameter"
//	@Failure		500		{object}	map[string]interface{}	"Failed to generate token"
//	@Security		BearerAuth
//	@Router			/security/generate-token [post]
func (h *Handler) GenerateToken(c *gin.Context) {
	lengthStr := c.DefaultQuery("length", "32")
	length, err := strconv.Atoi(lengthStr)
	if err != nil || length < 1 || length > 128 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid length parameter (1-128)",
		})
		return
	}

	token, err := h.service.GenerateSecureToken(length)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate token",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     token,
		"length":    len(token),
		"timestamp": time.Now(),
	})
}

// ValidateSecurityHeaders validates HTTP security headers
//
//	@Summary		Validate security headers
//	@Description	Validates HTTP security headers against security best practices
//	@Tags			Security & Authorization
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{headers=map[string]string}	true	"Headers to validate"
//	@Success		200		{object}	map[string]interface{}				"Header validation results"
//	@Failure		400		{object}	map[string]interface{}				"Invalid request format"
//	@Security		BearerAuth
//	@Router			/security/validate-headers [post]
func (h *Handler) ValidateSecurityHeaders(c *gin.Context) {
	var request struct {
		Headers map[string]string `json:"headers" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Convert map to http.Header
	headers := http.Header{}
	for key, value := range request.Headers {
		headers.Set(key, value)
	}

	violations := h.service.ValidateSecurityHeaders(headers)

	c.JSON(http.StatusOK, gin.H{
		"valid":            len(violations) == 0,
		"violations":       violations,
		"violations_count": len(violations),
		"timestamp":        time.Now(),
	})
}

// AnalyzeRequest analyzes request for suspicious activity
//
//	@Summary		Analyze request for threats
//	@Description	Analyzes an HTTP request for suspicious activity and security threats
//	@Tags			Security & Authorization
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{ip_address=string,user_agent=string,headers=map[string]string,path=string,method=string,body=string,user_id=string,session_id=string}	true	"Request analysis data"
//	@Success		200		{object}	map[string]interface{}																															"Request analysis results"
//	@Failure		400		{object}	map[string]interface{}																															"Invalid request format"
//	@Failure		500		{object}	map[string]interface{}																															"Analysis failed"
//	@Security		BearerAuth
//	@Router			/security/analyze-request [post]
func (h *Handler) AnalyzeRequest(c *gin.Context) {
	var request struct {
		IPAddress string            `json:"ip_address" binding:"required"`
		UserAgent string            `json:"user_agent"`
		Headers   map[string]string `json:"headers"`
		Path      string            `json:"path" binding:"required"`
		Method    string            `json:"method" binding:"required"`
		Body      string            `json:"body"`
		UserID    string            `json:"user_id"`
		SessionID string            `json:"session_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	securityRequest := &securityService.SecurityRequest{
		IPAddress: request.IPAddress,
		UserAgent: request.UserAgent,
		Headers:   request.Headers,
		Path:      request.Path,
		Method:    request.Method,
		Body:      request.Body,
		UserID:    request.UserID,
		SessionID: request.SessionID,
		Timestamp: time.Now(),
	}

	event, err := h.service.DetectSuspiciousActivity(securityRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to analyze request",
			"details": err.Error(),
		})
		return
	}

	if event == nil {
		c.JSON(http.StatusOK, gin.H{
			"suspicious": false,
			"message":    "No suspicious activity detected",
			"timestamp":  time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"suspicious": true,
		"event":      event,
		"risk_score": event.RiskScore,
		"blocked":    event.Blocked,
		"timestamp":  time.Now(),
	})
}

// CheckRateLimit checks rate limit for identifier
//
//	@Summary		Check rate limit status
//	@Description	Checks if a request should be rate limited for a specific identifier
//	@Tags			Security & Authorization
//	@Accept			json
//	@Produce		json
//	@Param			identifier	path		string								true	"Rate limit identifier (IP, user ID, etc.)"
//	@Param			request		body		object{limit=int,window=string}	true	"Rate limit parameters"
//	@Success		200			{object}	map[string]interface{}				"Rate limit check result"
//	@Failure		400			{object}	map[string]interface{}				"Invalid request format"
//	@Security		BearerAuth
//	@Router			/security/rate-limit/{identifier}/check [post]
func (h *Handler) CheckRateLimit(c *gin.Context) {
	identifier := c.Param("identifier")

	var request struct {
		Limit  int           `json:"limit" binding:"required"`
		Window time.Duration `json:"window" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	allowed := h.service.CheckRateLimit(identifier, request.Limit, request.Window)

	c.JSON(http.StatusOK, gin.H{
		"identifier": identifier,
		"allowed":    allowed,
		"limit":      request.Limit,
		"window":     request.Window,
		"timestamp":  time.Now(),
	})
}

// GetRateLimitStatus returns rate limit status
//
//	@Summary		Get rate limit status
//	@Description	Returns current rate limit status for a specific identifier
//	@Tags			Security & Authorization
//	@Produce		json
//	@Param			identifier	path		string					true	"Rate limit identifier"
//	@Success		200			{object}	map[string]interface{}	"Rate limit status"
//	@Failure		404			{object}	map[string]interface{}	"Rate limiter not found"
//	@Security		BearerAuth
//	@Router			/security/rate-limit/{identifier} [get]
func (h *Handler) GetRateLimitStatus(c *gin.Context) {
	identifier := c.Param("identifier")

	status := h.service.GetRateLimitStatus(identifier)
	if status == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "Rate limiter not found",
			"identifier": identifier,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    status,
		"timestamp": time.Now(),
	})
}

// ResetRateLimit resets rate limit for identifier
//
//	@Summary		Reset rate limit
//	@Description	Resets the rate limit counter for a specific identifier
//	@Tags			Security & Authorization
//	@Produce		json
//	@Param			identifier	path		string					true	"Rate limit identifier"
//	@Success		200			{object}	map[string]interface{}	"Rate limit reset successfully"
//	@Security		BearerAuth
//	@Router			/security/rate-limit/{identifier} [delete]
func (h *Handler) ResetRateLimit(c *gin.Context) {
	identifier := c.Param("identifier")

	h.service.ResetRateLimit(identifier)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Rate limit reset successfully",
		"identifier": identifier,
		"timestamp":  time.Now(),
	})
}

// RecordSecurityEvent records a security event
//
//	@Summary		Record security event
//	@Description	Records a security event for monitoring and analysis
//	@Tags			Security & Authorization
//	@Accept			json
//	@Produce		json
//	@Param			event	body		models.SecurityEvent	true	"Security event data"
//	@Success		201		{object}	map[string]interface{}	"Security event recorded successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid event format"
//	@Failure		500		{object}	map[string]interface{}	"Failed to record security event"
//	@Security		BearerAuth
//	@Router			/security/events [post]
func (h *Handler) RecordSecurityEvent(c *gin.Context) {
	var event models.SecurityEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid event format",
			"details": err.Error(),
		})
		return
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Generate ID if not provided
	if event.ID == "" {
		event.ID = "sec_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	// Calculate risk score
	event.RiskScore = event.CalculateRiskScore()

	if err := h.service.RecordSecurityEvent(&event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to record security event",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Security event recorded successfully",
		"event_id":   event.ID,
		"risk_score": event.RiskScore,
		"timestamp":  time.Now(),
	})
}

// GetSecurityEvents returns filtered security events
//
//	@Summary		Get security events
//	@Description	Returns security events with optional filtering
//	@Tags			Security & Authorization
//	@Produce		json
//	@Param			limit	query		int						false	"Limit number of results"
//	@Param			offset	query		int						false	"Offset for pagination"
//	@Success		200		{object}	map[string]interface{}	"Security events list"
//	@Failure		500		{object}	map[string]interface{}	"Failed to get security events"
//	@Security		BearerAuth
//	@Router			/security/events [get]
func (h *Handler) GetSecurityEvents(c *gin.Context) {
	filter := securityService.SecurityEventFilter{}

	// Parse query parameters
	if limit := c.Query("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			filter.Limit = val
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if val, err := strconv.Atoi(offset); err == nil {
			filter.Offset = val
		}
	}

	events, err := h.service.GetSecurityEvents(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get security events",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":    events,
		"count":     len(events),
		"timestamp": time.Now(),
	})
}

// GetSecurityAlerts returns security alerts
//
//	@Summary		Get security alerts
//	@Description	Returns current security alerts and warnings
//	@Tags			Security & Authorization
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Security alerts list"
//	@Failure		500	{object}	map[string]interface{}	"Failed to get security alerts"
//	@Security		BearerAuth
//	@Router			/security/alerts [get]
func (h *Handler) GetSecurityAlerts(c *gin.Context) {
	alerts, err := h.service.GetSecurityAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get security alerts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts":    alerts,
		"count":     len(alerts),
		"timestamp": time.Now(),
	})
}

// StartSecurityScan starts a vulnerability scan
//
//	@Summary		Start security vulnerability scan
//	@Description	Initiates a comprehensive security vulnerability scan
//	@Tags			Security & Authorization
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Security scan completed"
//	@Failure		500	{object}	map[string]interface{}	"Security scan failed"
//	@Security		BearerAuth
//	@Router			/security/scan [post]
func (h *Handler) StartSecurityScan(c *gin.Context) {
	result, err := h.service.ScanForVulnerabilities()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Security scan failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Security scan completed",
		"result":    result,
		"timestamp": time.Now(),
	})
}

// GetPerformanceMetrics returns performance metrics
//
//	@Summary		Get performance metrics
//	@Description	Returns system performance metrics and statistics
//	@Tags			Security & Authorization
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Performance metrics"
//	@Security		BearerAuth
//	@Router			/performance/metrics [get]
func (h *Handler) GetPerformanceMetrics(c *gin.Context) {
	metrics := h.service.GetPerformanceMetrics()

	c.JSON(http.StatusOK, gin.H{
		"metrics":   metrics,
		"timestamp": time.Now(),
	})
}

// GetCacheStats returns cache statistics
//
//	@Summary		Get cache statistics
//	@Description	Returns cache performance statistics and hit/miss ratios
//	@Tags			Security & Authorization
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Cache statistics"
//	@Security		BearerAuth
//	@Router			/performance/cache/stats [get]
func (h *Handler) GetCacheStats(c *gin.Context) {
	stats := h.service.CacheStats()

	c.JSON(http.StatusOK, gin.H{
		"stats":     stats,
		"timestamp": time.Now(),
	})
}

// GetSecurityHealth returns security system health
//
//	@Summary		Get security system health
//	@Description	Returns health status of security subsystems
//	@Tags			Security & Authorization
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Security system health"
//	@Security		BearerAuth
//	@Router			/security/health [get]
func (h *Handler) GetSecurityHealth(c *gin.Context) {
	health := h.service.GetSecurityHealth()

	c.JSON(http.StatusOK, gin.H{
		"health":    health,
		"timestamp": time.Now(),
	})
}

// GetPerformanceHealth returns performance system health
//
//	@Summary		Get performance system health
//	@Description	Returns health status of performance monitoring systems
//	@Tags			Security & Authorization
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Performance system health"
//	@Security		BearerAuth
//	@Router			/performance/health [get]
func (h *Handler) GetPerformanceHealth(c *gin.Context) {
	health := h.service.GetPerformanceHealth()

	c.JSON(http.StatusOK, gin.H{
		"health":    health,
		"timestamp": time.Now(),
	})
}

// Placeholder implementations for remaining endpoints

func (h *Handler) CreateSession(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}

func (h *Handler) ValidateSession(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}

func (h *Handler) RevokeSession(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}

// CleanupSessions cleans up expired sessions
//
//	@Summary		Cleanup expired sessions
//	@Description	Removes expired sessions from the security system
//	@Tags			Security & Authorization
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Sessions cleaned up"
//	@Security		BearerAuth
//	@Router			/security/sessions/cleanup [delete]
func (h *Handler) CleanupSessions(c *gin.Context) {
	cleaned := h.service.CleanupExpiredSessions()
	c.JSON(http.StatusOK, gin.H{
		"message":       "Sessions cleaned up",
		"cleaned_count": cleaned,
		"timestamp":     time.Now(),
	})
}

func (h *Handler) ValidateInput(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}

func (h *Handler) CheckIPReputation(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}

func (h *Handler) RecordPerformanceMetric(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}

func (h *Handler) GetServiceStats(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}

// GetCacheValue retrieves a value from cache
//
//	@Summary		Get cache value
//	@Description	Retrieves a value from the cache by key
//	@Tags			Security & Authorization
//	@Produce		json
//	@Param			key	path		string					true	"Cache key"
//	@Success		200	{object}	map[string]interface{}	"Cache value retrieved"
//	@Failure		404	{object}	map[string]interface{}	"Cache key not found"
//	@Security		BearerAuth
//	@Router			/performance/cache/{key} [get]
func (h *Handler) GetCacheValue(c *gin.Context) {
	key := c.Param("key")

	value, found := h.service.CacheGet(key)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Cache key not found",
			"key":   key,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":       key,
		"value":     value,
		"found":     true,
		"timestamp": time.Now(),
	})
}

// SetCacheValue stores a value in cache
//
//	@Summary		Set cache value
//	@Description	Stores a value in the cache with optional TTL
//	@Tags			Security & Authorization
//	@Accept			json
//	@Produce		json
//	@Param			key		path		string								true	"Cache key"
//	@Param			request	body		object{value=interface{},ttl=int}	true	"Cache value and TTL"
//	@Success		200		{object}	map[string]interface{}				"Cache value set successfully"
//	@Failure		400		{object}	map[string]interface{}				"Invalid request format"
//	@Failure		500		{object}	map[string]interface{}				"Failed to set cache value"
//	@Security		BearerAuth
//	@Router			/performance/cache/{key} [post]
func (h *Handler) SetCacheValue(c *gin.Context) {
	key := c.Param("key")

	var request struct {
		Value interface{} `json:"value" binding:"required"`
		TTL   int         `json:"ttl"` // seconds
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	ttl := time.Duration(request.TTL) * time.Second
	if request.TTL == 0 {
		ttl = time.Hour // default 1 hour
	}

	if err := h.service.CacheSet(key, request.Value, ttl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to set cache value",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Cache value set successfully",
		"key":       key,
		"ttl":       ttl,
		"timestamp": time.Now(),
	})
}

// DeleteCacheValue removes a value from cache
//
//	@Summary		Delete cache value
//	@Description	Removes a value from the cache by key
//	@Tags			Security & Authorization
//	@Produce		json
//	@Param			key	path		string					true	"Cache key"
//	@Success		200	{object}	map[string]interface{}	"Cache value deleted successfully"
//	@Failure		500	{object}	map[string]interface{}	"Failed to delete cache value"
//	@Security		BearerAuth
//	@Router			/performance/cache/{key} [delete]
func (h *Handler) DeleteCacheValue(c *gin.Context) {
	key := c.Param("key")

	if err := h.service.CacheDelete(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete cache value",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Cache value deleted successfully",
		"key":       key,
		"timestamp": time.Now(),
	})
}

func (h *Handler) WarmupCache(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}

func (h *Handler) OptimizePerformance(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}

func (h *Handler) EnableCompression(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}

func (h *Handler) OptimizeConnections(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}

func (h *Handler) GetResourceUsage(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}
