package chat

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/chat"
)

// Handler handles chat-related HTTP requests
type Handler struct {
	chatService chat.Service
}

// NewHandler creates a new chat handler
func NewHandler(chatService chat.Service) *Handler {
	return &Handler{
		chatService: chatService,
	}
}

// RegisterRoutes registers chat routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	chatGroup := router.Group("/chat")
	{
		// Session routes
		chatGroup.POST("/sessions", h.CreateSession)
		chatGroup.GET("/sessions", h.GetSessions)
		chatGroup.GET("/sessions/:sessionId", h.GetSession)
		chatGroup.PUT("/sessions/:sessionId", h.UpdateSession)
		chatGroup.DELETE("/sessions/:sessionId", h.DeleteSession)

		// Message routes
		chatGroup.GET("/sessions/:sessionId/messages", h.GetMessages)
		chatGroup.POST("/sessions/:sessionId/messages", h.SendMessage)

		// Command routes
		chatGroup.POST("/commands/preview", h.GenerateCommandPreview)
		chatGroup.GET("/sessions/:sessionId/commands", h.GetCommandPreviews)

		// Context routes
		chatGroup.GET("/sessions/:sessionId/context", h.GetChatContext)

		// WebSocket route (will be handled separately by websocket handler)
		// chatGroup.GET("/sessions/:sessionId/ws", h.WebSocketConnection)
	}
}

// CreateSession creates a new chat session
//	@Summary		Create a new chat session
//	@Description	Creates a new chat conversation session
//	@Tags			chat
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.ChatSessionRequest	true	"Chat session request"
//	@Success		201		{object}	models.ChatResponse{data=models.ChatSession}
//	@Failure		400		{object}	models.ChatResponse
//	@Failure		401		{object}	models.ChatResponse
//	@Failure		500		{object}	models.ChatResponse
//	@Router			/chat/sessions [post]
func (h *Handler) CreateSession(c *gin.Context) {
	var req models.ChatSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Invalid request format: " + err.Error()),
		})
		return
	}

	// Get user ID from context (would be set by auth middleware)
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		if uid, err := uuid.Parse(userIDStr.(string)); err == nil {
			userID = &uid
		}
	}

	session, err := h.chatService.CreateSession(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Failed to create session: " + err.Error()),
		})
		return
	}

	c.JSON(http.StatusCreated, models.ChatResponse{
		Success: true,
		Data:    session,
		Message: stringPtr("Session created successfully"),
	})
}

// GetSessions retrieves all chat sessions for a user
//	@Summary		Get all chat sessions
//	@Description	Retrieves all chat sessions for the current user
//	@Tags			chat
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	models.ChatResponse{data=[]models.ChatSession}
//	@Failure		401	{object}	models.ChatResponse
//	@Failure		500	{object}	models.ChatResponse
//	@Router			/chat/sessions [get]
func (h *Handler) GetSessions(c *gin.Context) {
	// Get user ID from context
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		if uid, err := uuid.Parse(userIDStr.(string)); err == nil {
			userID = &uid
		}
	}

	sessions, err := h.chatService.GetSessions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Failed to retrieve sessions: " + err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.ChatResponse{
		Success: true,
		Data:    sessions,
	})
}

// GetSession retrieves a specific chat session
//	@Summary		Get a chat session
//	@Description	Retrieves a specific chat session by ID
//	@Tags			chat
//	@Accept			json
//	@Produce		json
//	@Param			sessionId	path		string	true	"Session ID"
//	@Success		200			{object}	models.ChatResponse{data=models.ChatSessionResponse}
//	@Failure		400			{object}	models.ChatResponse
//	@Failure		401			{object}	models.ChatResponse
//	@Failure		404			{object}	models.ChatResponse
//	@Failure		500			{object}	models.ChatResponse
//	@Router			/chat/sessions/{sessionId} [get]
func (h *Handler) GetSession(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Invalid session ID format"),
		})
		return
	}

	// Get user ID from context
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		if uid, err := uuid.Parse(userIDStr.(string)); err == nil {
			userID = &uid
		}
	}

	session, err := h.chatService.GetSession(c.Request.Context(), sessionID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Session not found: " + err.Error()),
		})
		return
	}

	// Also get recent messages and context
	messages, _ := h.chatService.GetMessages(c.Request.Context(), sessionID, 20, 0, userID)
	context, _ := h.chatService.GetChatContext(c.Request.Context(), sessionID, userID)

	response := models.ChatSessionResponse{
		Session:  session,
		Messages: convertToChatMessageSlice(messages), // Convert []*ChatMessage to []ChatMessage
		Context:  context,
	}

	c.JSON(http.StatusOK, models.ChatResponse{
		Success: true,
		Data:    response,
	})
}

// UpdateSession updates a chat session
//	@Summary		Update a chat session
//	@Description	Updates a chat session's properties
//	@Tags			chat
//	@Accept			json
//	@Produce		json
//	@Param			sessionId	path		string						true	"Session ID"
//	@Param			updates		body		map[string]interface{}		true	"Fields to update"
//	@Success		200			{object}	models.ChatResponse{data=models.ChatSession}
//	@Failure		400			{object}	models.ChatResponse
//	@Failure		401			{object}	models.ChatResponse
//	@Failure		404			{object}	models.ChatResponse
//	@Failure		500			{object}	models.ChatResponse
//	@Router			/chat/sessions/{sessionId} [put]
func (h *Handler) UpdateSession(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Invalid session ID format"),
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Invalid request format: " + err.Error()),
		})
		return
	}

	// Get user ID from context
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		if uid, err := uuid.Parse(userIDStr.(string)); err == nil {
			userID = &uid
		}
	}

	session, err := h.chatService.UpdateSession(c.Request.Context(), sessionID, updates, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Failed to update session: " + err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.ChatResponse{
		Success: true,
		Data:    session,
		Message: stringPtr("Session updated successfully"),
	})
}

// DeleteSession deletes a chat session
//	@Summary		Delete a chat session
//	@Description	Deletes a chat session and all its messages
//	@Tags			chat
//	@Accept			json
//	@Produce		json
//	@Param			sessionId	path		string	true	"Session ID"
//	@Success		200			{object}	models.ChatResponse
//	@Failure		400			{object}	models.ChatResponse
//	@Failure		401			{object}	models.ChatResponse
//	@Failure		404			{object}	models.ChatResponse
//	@Failure		500			{object}	models.ChatResponse
//	@Router			/chat/sessions/{sessionId} [delete]
func (h *Handler) DeleteSession(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Invalid session ID format"),
		})
		return
	}

	// Get user ID from context
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		if uid, err := uuid.Parse(userIDStr.(string)); err == nil {
			userID = &uid
		}
	}

	err = h.chatService.DeleteSession(c.Request.Context(), sessionID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Failed to delete session: " + err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.ChatResponse{
		Success: true,
		Message: stringPtr("Session deleted successfully"),
	})
}

// GetMessages retrieves messages for a chat session
//	@Summary		Get messages for a session
//	@Description	Retrieves messages for a specific chat session
//	@Tags			chat
//	@Accept			json
//	@Produce		json
//	@Param			sessionId	path		string	true	"Session ID"
//	@Param			limit		query		int		false	"Number of messages to retrieve"	default(50)
//	@Param			offset		query		int		false	"Number of messages to skip"		default(0)
//	@Success		200			{object}	models.ChatResponse{data=[]models.ChatMessage}
//	@Failure		400			{object}	models.ChatResponse
//	@Failure		401			{object}	models.ChatResponse
//	@Failure		500			{object}	models.ChatResponse
//	@Router			/chat/sessions/{sessionId}/messages [get]
func (h *Handler) GetMessages(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Invalid session ID format"),
		})
		return
	}

	// Parse query parameters
	limit := 50
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get user ID from context
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		if uid, err := uuid.Parse(userIDStr.(string)); err == nil {
			userID = &uid
		}
	}

	messages, err := h.chatService.GetMessages(c.Request.Context(), sessionID, limit, offset, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Failed to retrieve messages: " + err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.ChatResponse{
		Success: true,
		Data:    messages,
	})
}

// SendMessage sends a message in a chat session
//	@Summary		Send a message
//	@Description	Sends a message in a chat session and processes it
//	@Tags			chat
//	@Accept			json
//	@Produce		json
//	@Param			sessionId	path		string						true	"Session ID"
//	@Param			message		body		models.ChatMessageRequest	true	"Message to send"
//	@Success		201			{object}	models.ChatResponse{data=models.ChatMessage}
//	@Failure		400			{object}	models.ChatResponse
//	@Failure		401			{object}	models.ChatResponse
//	@Failure		500			{object}	models.ChatResponse
//	@Router			/chat/sessions/{sessionId}/messages [post]
func (h *Handler) SendMessage(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Invalid session ID format"),
		})
		return
	}

	var req models.ChatMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Invalid request format: " + err.Error()),
		})
		return
	}

	// Get user ID from context
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		if uid, err := uuid.Parse(userIDStr.(string)); err == nil {
			userID = &uid
		}
	}

	message, err := h.chatService.SendMessage(c.Request.Context(), sessionID, &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Failed to send message: " + err.Error()),
		})
		return
	}

	c.JSON(http.StatusCreated, models.ChatResponse{
		Success: true,
		Data:    message,
		Message: stringPtr("Message sent successfully"),
	})
}

// GenerateCommandPreview generates a command preview from natural language
//	@Summary		Generate command preview
//	@Description	Generates a Kubernetes command preview from natural language input
//	@Tags			chat
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.CommandPreviewRequest	true	"Command preview request"
//	@Success		201		{object}	models.ChatResponse{data=models.CommandPreview}
//	@Failure		400		{object}	models.ChatResponse
//	@Failure		401		{object}	models.ChatResponse
//	@Failure		500		{object}	models.ChatResponse
//	@Router			/chat/commands/preview [post]
func (h *Handler) GenerateCommandPreview(c *gin.Context) {
	var req models.CommandPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Invalid request format: " + err.Error()),
		})
		return
	}

	// Get user ID from context
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		if uid, err := uuid.Parse(userIDStr.(string)); err == nil {
			userID = &uid
		}
	}

	preview, err := h.chatService.GenerateCommandPreview(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Failed to generate command preview: " + err.Error()),
		})
		return
	}

	c.JSON(http.StatusCreated, models.ChatResponse{
		Success: true,
		Data:    preview,
		Message: stringPtr("Command preview generated successfully"),
	})
}

// GetCommandPreviews retrieves command previews for a session
//	@Summary		Get command previews
//	@Description	Retrieves command previews for a specific chat session
//	@Tags			chat
//	@Accept			json
//	@Produce		json
//	@Param			sessionId	path		string	true	"Session ID"
//	@Success		200			{object}	models.ChatResponse{data=[]models.CommandPreview}
//	@Failure		400			{object}	models.ChatResponse
//	@Failure		401			{object}	models.ChatResponse
//	@Failure		500			{object}	models.ChatResponse
//	@Router			/chat/sessions/{sessionId}/commands [get]
func (h *Handler) GetCommandPreviews(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Invalid session ID format"),
		})
		return
	}

	// Get user ID from context
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		if uid, err := uuid.Parse(userIDStr.(string)); err == nil {
			userID = &uid
		}
	}

	previews, err := h.chatService.GetCommandPreviews(c.Request.Context(), sessionID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Failed to retrieve command previews: " + err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.ChatResponse{
		Success: true,
		Data:    previews,
	})
}

// GetChatContext retrieves the chat context for a session
//	@Summary		Get chat context
//	@Description	Retrieves the current context for a chat session
//	@Tags			chat
//	@Accept			json
//	@Produce		json
//	@Param			sessionId	path		string	true	"Session ID"
//	@Success		200			{object}	models.ChatResponse{data=models.ChatContext}
//	@Failure		400			{object}	models.ChatResponse
//	@Failure		401			{object}	models.ChatResponse
//	@Failure		500			{object}	models.ChatResponse
//	@Router			/chat/sessions/{sessionId}/context [get]
func (h *Handler) GetChatContext(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Invalid session ID format"),
		})
		return
	}

	// Get user ID from context
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		if uid, err := uuid.Parse(userIDStr.(string)); err == nil {
			userID = &uid
		}
	}

	context, err := h.chatService.GetChatContext(c.Request.Context(), sessionID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ChatResponse{
			Success: false,
			Error:   stringPtr("Failed to retrieve chat context: " + err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.ChatResponse{
		Success: true,
		Data:    context,
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func convertToChatMessageSlice(messages []*models.ChatMessage) []models.ChatMessage {
	result := make([]models.ChatMessage, len(messages))
	for i, msg := range messages {
		result[i] = *msg
	}
	return result
}