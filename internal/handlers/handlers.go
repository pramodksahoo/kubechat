package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pramodksahoo/kubechat/internal/audit"
	"github.com/pramodksahoo/kubechat/internal/k8s"
	"github.com/pramodksahoo/kubechat/internal/llm"
	"github.com/pramodksahoo/kubechat/internal/translator"
)

type Handlers struct {
	k8sClient   *k8s.Client
	llmService  llm.Service
	translator  *translator.Engine
	auditLogger *audit.Logger
	upgrader    websocket.Upgrader
}

type QueryRequest struct {
	Query string `json:"query" binding:"required"`
}

type QueryResponse struct {
	Command     string      `json:"command"`
	Explanation string      `json:"explanation"`
	Safety      string      `json:"safety"`
	Preview     bool        `json:"preview"`
	Result      interface{} `json:"result,omitempty"`
	Error       string      `json:"error,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

type ExecuteRequest struct {
	Command   string `json:"command" binding:"required"`
	Namespace string `json:"namespace"`
}

func NewHandlers(k8sClient *k8s.Client, llmService llm.Service) *Handlers {
	translator := translator.NewEngine(llmService, k8sClient)
	auditLogger := audit.NewLogger()

	return &Handlers{
		k8sClient:   k8sClient,
		llmService:  llmService,
		translator:  translator,
		auditLogger: auditLogger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}
}

func (h *Handlers) HandleQuery(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process the natural language query
	result, err := h.translator.ProcessQuery(c.Request.Context(), req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log to audit trail
	h.auditLogger.LogQuery(req.Query, result.Command, result.Safety, "user")

	response := QueryResponse{
		Command:     result.Command,
		Explanation: result.Explanation,
		Safety:      result.Safety,
		Preview:     result.Preview,
		Result:      result.Result,
		Error:       result.Error,
		Timestamp:   time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) HandleExecute(c *gin.Context) {
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Execute the command
	result, err := h.translator.ExecuteCommand(c.Request.Context(), req.Command, req.Namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log to audit trail
	h.auditLogger.LogExecution(req.Command, req.Namespace, result.Error == "", "user")

	response := QueryResponse{
		Command:   result.Command,
		Result:    result.Result,
		Error:     result.Error,
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "0.1.0-poc",
	})
}

func (h *Handlers) HandleGetClusters(c *gin.Context) {
	namespaces, err := h.k8sClient.GetNamespaces()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	nodes, err := h.k8sClient.GetNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"namespaces": namespaces,
		"nodes":      nodes,
		"timestamp":  time.Now(),
	})
}

func (h *Handlers) HandleAuditLog(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	logs := h.auditLogger.GetRecentLogs(limit)
	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"count": len(logs),
	})
}

func (h *Handlers) HandleWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	// Handle WebSocket connections for real-time updates
	for {
		var message map[string]interface{}
		if err := conn.ReadJSON(&message); err != nil {
			break
		}

		// Echo back for now - in production, this would handle real-time cluster events
		response := map[string]interface{}{
			"type":      "response",
			"message":   "WebSocket connection established",
			"timestamp": time.Now(),
		}

		if err := conn.WriteJSON(response); err != nil {
			break
		}
	}
}