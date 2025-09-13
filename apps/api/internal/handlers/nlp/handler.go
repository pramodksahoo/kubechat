package nlp

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/auth"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/nlp"
)

// Handler handles NLP-related HTTP requests
type Handler struct {
	nlpService nlp.Service
}

// NewHandler creates a new NLP handler
func NewHandler(nlpService nlp.Service) *Handler {
	return &Handler{
		nlpService: nlpService,
	}
}

// RegisterRoutes registers NLP routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// All NLP routes require authentication
	nlpRoutes := router.Group("/nlp")
	nlpRoutes.Use(auth.RequireWritePermission()) // Require at least user role for NLP operations
	{
		// Query processing endpoints
		nlpRoutes.POST("/process", h.ProcessQuery)
		nlpRoutes.POST("/validate", h.ValidateCommand)

		// Service information endpoints
		nlpRoutes.GET("/providers", h.GetSupportedProviders)
		nlpRoutes.GET("/health", h.HealthCheck)
		nlpRoutes.GET("/metrics", h.GetMetrics)

		// Command classification
		nlpRoutes.POST("/classify", h.ClassifyCommand)
	}
}

// ProcessQuery handles natural language query processing requests
//
//	@Summary		Process natural language query
//	@Description	Processes a natural language query and converts it to Kubernetes commands
//	@Tags			AI & NLP Services
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{query=string,context=string,cluster_info=string,provider=string}	true	"NLP query request"
//	@Success		200		{object}	map[string]interface{}													"Query processed successfully"
//	@Failure		400		{object}	map[string]interface{}													"Invalid request format"
//	@Failure		500		{object}	map[string]interface{}													"NLP processing error"
//	@Security		BearerAuth
//	@Router			/nlp/process [post]
func (h *Handler) ProcessQuery(c *gin.Context) {
	var req struct {
		Query       string `json:"query" binding:"required"`
		Context     string `json:"context,omitempty"`
		ClusterInfo string `json:"cluster_info,omitempty"`
		Provider    string `json:"provider,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Get user context from authentication middleware
	userID, _, sessionID, _ := auth.ExtractUserContext(c)

	// Create NLP request
	nlpRequest := &models.NLPRequest{
		ID:          uuid.New(),
		UserID:      uuid.MustParse(userID),
		SessionID:   uuid.MustParse(sessionID),
		Query:       req.Query,
		Context:     req.Context,
		ClusterInfo: req.ClusterInfo,
		CreatedAt:   time.Now(),
	}

	// Set provider if specified
	if req.Provider != "" {
		nlpRequest.Provider = models.NLPProvider(req.Provider)
	}

	// Process the query
	response, err := h.nlpService.ProcessQuery(c.Request.Context(), nlpRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process query",
			"code":  "NLP_PROCESSING_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    response,
		"message": "Query processed successfully",
	})
}

// ValidateCommand validates a kubectl command for safety and correctness
//
//	@Summary		Validate kubectl command
//	@Description	Validates a kubectl command for safety, correctness, and potential risks
//	@Tags			AI & NLP Services
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{command=string}	true	"Command validation request"
//	@Success		200		{object}	map[string]interface{}	"Command validated successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request format"
//	@Failure		500		{object}	map[string]interface{}	"Command validation error"
//	@Security		BearerAuth
//	@Router			/nlp/validate [post]
func (h *Handler) ValidateCommand(c *gin.Context) {
	var req struct {
		Command string `json:"command" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Validate the command
	result, err := h.nlpService.ValidateCommand(c.Request.Context(), req.Command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to validate command",
			"code":  "COMMAND_VALIDATION_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    result,
		"message": "Command validated successfully",
	})
}

// GetSupportedProviders returns available NLP providers
//
//	@Summary		Get supported NLP providers
//	@Description	Returns a list of available NLP providers and their status
//	@Tags			AI & NLP Services
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of supported NLP providers"
//	@Failure		500	{object}	map[string]interface{}	"Failed to get supported providers"
//	@Security		BearerAuth
//	@Router			/nlp/providers [get]
func (h *Handler) GetSupportedProviders(c *gin.Context) {
	providers, err := h.nlpService.GetSupportedProviders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get supported providers",
			"code":  "PROVIDERS_ERROR",
		})
		return
	}

	// Convert to response format
	providerList := make([]gin.H, len(providers))
	for i, provider := range providers {
		providerList[i] = gin.H{
			"name":         provider,
			"display_name": provider.GetProviderDisplayName(),
			"available":    true,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  providerList,
		"count": len(providerList),
	})
}

// HealthCheck checks NLP service health
//
//	@Summary		Check NLP service health
//	@Description	Performs a health check on the NLP service and providers
//	@Tags			AI & NLP Services
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"NLP service is operational"
//	@Failure		503	{object}	map[string]interface{}	"NLP service health check failed"
//	@Security		BearerAuth
//	@Router			/nlp/health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	err := h.nlpService.HealthCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "NLP service health check failed",
			"code":  "NLP_HEALTH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "NLP service is operational",
	})
}

// GetMetrics returns NLP service metrics
//
//	@Summary		Get NLP service metrics
//	@Description	Returns performance metrics and statistics for NLP operations
//	@Tags			AI & NLP Services
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"NLP service metrics"
//	@Failure		500	{object}	map[string]interface{}	"Failed to get NLP metrics"
//	@Security		BearerAuth
//	@Router			/nlp/metrics [get]
func (h *Handler) GetMetrics(c *gin.Context) {
	metrics, err := h.nlpService.GetMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get NLP metrics",
			"code":  "METRICS_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": metrics,
	})
}

// ClassifyCommand classifies the safety level of a command without full processing
//
//	@Summary		Classify command safety level
//	@Description	Classifies the safety level and risk of a kubectl command without full processing
//	@Tags			AI & NLP Services
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{command=string}	true	"Command classification request"
//	@Success		200		{object}	map[string]interface{}	"Command classified successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request format"
//	@Failure		500		{object}	map[string]interface{}	"Classification error"
//	@Security		BearerAuth
//	@Router			/nlp/classify [post]
func (h *Handler) ClassifyCommand(c *gin.Context) {
	var req struct {
		Command string `json:"command" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Validate and classify the command
	result, err := h.nlpService.ValidateCommand(c.Request.Context(), req.Command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to classify command",
			"code":  "CLASSIFICATION_ERROR",
		})
		return
	}

	// Return classification information
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"command":      req.Command,
			"safety_level": result.SafetyLevel,
			"description":  result.SafetyLevel.GetSafetyDescription(),
			"color":        result.SafetyLevel.GetSafetyColor(),
			"is_valid":     result.IsValid,
			"warnings":     result.Warnings,
			"errors":       result.Errors,
			"suggestions":  result.Suggestions,
		},
		"message": "Command classified successfully",
	})
}
