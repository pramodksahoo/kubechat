package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/auth"
)

// Handler handles authentication-related HTTP requests
type Handler struct {
	authService auth.Service
}

// NewHandler creates a new authentication handler
func NewHandler(authService auth.Service) *Handler {
	return &Handler{
		authService: authService,
	}
}

// RegisterRoutes registers authentication routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Public routes (no authentication required)
	public := router.Group("/auth")
	{
		public.POST("/register", h.Register)
		public.POST("/login", h.Login)
		public.POST("/refresh", h.RefreshSession)
	}

	// Protected routes (authentication required)
	protected := router.Group("/auth")
	protected.Use(auth.AuthMiddleware(h.authService))
	{
		protected.POST("/logout", h.Logout)
		protected.GET("/me", h.GetCurrentUser)
		protected.GET("/profile", h.GetProfile)
	}

	// Admin routes
	admin := router.Group("/auth/admin")
	admin.Use(auth.AuthMiddleware(h.authService))
	admin.Use(auth.RequireRole("admin"))
	{
		admin.GET("/users", h.ListUsers)
		admin.GET("/users/:id", h.GetUser)
	}
}

// Register handles user registration
//
//	@Summary		Register new user account
//	@Description	Create a new user account with username, email and password
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.CreateUserRequest	true						"User registration data"
//	@Success		201		{object}	map[string]interface{}		"User created successfully"
//	@Failure		400		{object}	map[string]interface{}		"Invalid request format"
//	@Failure		409		{object}	map[string]interface{}		"User already exists"
//	@Failure		500		{object}	map[string]interface{}		"Registration failed"
//	@Router			/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"code":    "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	// Register user
	user, err := h.authService.RegisterUser(c.Request.Context(), &req)
	if err != nil {
		status := http.StatusInternalServerError
		code := "REGISTRATION_FAILED"

		if err == auth.ErrUserAlreadyExists {
			status = http.StatusConflict
			code = "USER_EXISTS"
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// Login handles user login
//
//	@Summary		User login authentication
//	@Description	Authenticate user with username/email and password, returns session token
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.LoginRequest		true	"Login credentials"
//	@Success		200		{object}	map[string]interface{}	"Login successful with session token"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request format"
//	@Failure		401		{object}	map[string]interface{}	"Invalid credentials"
//	@Failure		500		{object}	map[string]interface{}	"Login failed"
//	@Router			/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"code":    "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	// Get client IP and User-Agent
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Login user
	loginResponse, err := h.authService.LoginUser(c.Request.Context(), &req, &clientIP, &userAgent)
	if err != nil {
		status := http.StatusInternalServerError
		code := "LOGIN_FAILED"

		if err == auth.ErrInvalidCredentials {
			status = http.StatusUnauthorized
			code = "INVALID_CREDENTIALS"
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data": gin.H{
			"user": gin.H{
				"id":       loginResponse.User.ID,
				"username": loginResponse.User.Username,
				"email":    loginResponse.User.Email,
				"role":     loginResponse.User.Role,
			},
			"token":      loginResponse.SessionToken,
			"expires_at": loginResponse.ExpiresAt,
		},
	})
}

// Logout handles user logout
//
//	@Summary		User logout
//	@Description	Logs out the current user and invalidates their session token
//	@Tags			Authentication
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Logout successful"
//	@Failure		400	{object}	map[string]interface{}	"Authorization header required"
//	@Failure		500	{object}	map[string]interface{}	"Logout failed"
//	@Security		BearerAuth
//	@Router			/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Authorization header required",
			"code":  "AUTH_HEADER_MISSING",
		})
		return
	}

	// Extract token
	tokenString := authHeader[7:] // Remove "Bearer " prefix

	// Logout user
	if err := h.authService.LogoutUser(c.Request.Context(), tokenString); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to logout",
			"code":  "LOGOUT_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

// RefreshSession handles session refresh
//
//	@Summary		Refresh user session
//	@Description	Refreshes an existing session token with a new one
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{token=string}	true	"Session refresh request"
//	@Success		200		{object}	map[string]interface{}	"Session refreshed successfully"
//	@Failure		400		{object}	map[string]interface{}	"Token is required"
//	@Failure		401		{object}	map[string]interface{}	"Session expired or invalid"
//	@Failure		500		{object}	map[string]interface{}	"Refresh failed"
//	@Router			/auth/refresh [post]
func (h *Handler) RefreshSession(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token is required",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Refresh session
	loginResponse, err := h.authService.RefreshSession(c.Request.Context(), req.Token)
	if err != nil {
		status := http.StatusInternalServerError
		code := "REFRESH_FAILED"

		if err == auth.ErrSessionExpired || err == auth.ErrInvalidToken {
			status = http.StatusUnauthorized
			code = "SESSION_EXPIRED"
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session refreshed successfully",
		"data": gin.H{
			"user": gin.H{
				"id":       loginResponse.User.ID,
				"username": loginResponse.User.Username,
				"email":    loginResponse.User.Email,
				"role":     loginResponse.User.Role,
			},
			"token":      loginResponse.SessionToken,
			"expires_at": loginResponse.ExpiresAt,
		},
	})
}

// GetCurrentUser returns current authenticated user
//
//	@Summary		Get current user info
//	@Description	Returns basic information about the currently authenticated user
//	@Tags			Authentication
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Current user information"
//	@Failure		500	{object}	map[string]interface{}	"Failed to extract user context"
//	@Security		BearerAuth
//	@Router			/auth/me [get]
func (h *Handler) GetCurrentUser(c *gin.Context) {
	userID, username, role, ok := auth.ExtractUserContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to extract user context",
			"code":  "CONTEXT_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":       userID,
			"username": username,
			"role":     role,
		},
	})
}

// GetProfile returns detailed user profile
//
//	@Summary		Get user profile
//	@Description	Returns detailed profile information for the current user
//	@Tags			Authentication
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"User profile information"
//	@Failure		500	{object}	map[string]interface{}	"Failed to get user profile"
//	@Security		BearerAuth
//	@Router			/auth/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	tokenString := authHeader[7:] // Remove "Bearer " prefix

	// Get current user
	user, err := h.authService.GetCurrentUser(c.Request.Context(), tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user profile",
			"code":  "PROFILE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}

// ListUsers returns list of users (admin only)
//
//	@Summary		List all users
//	@Description	Returns a list of all users (admin only)
//	@Tags			Authentication
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of users"
//	@Security		BearerAuth
//	@Router			/auth/admin/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	// This is a placeholder - would need pagination implementation
	c.JSON(http.StatusOK, gin.H{
		"message": "List users endpoint",
		"data":    []interface{}{},
	})
}

// GetUser returns specific user by ID (admin only)
//
//	@Summary		Get user by ID
//	@Description	Returns detailed information about a specific user (admin only)
//	@Tags			Authentication
//	@Produce		json
//	@Param			id	path		string					true	"User ID"	format(uuid)
//	@Success		200	{object}	map[string]interface{}	"User information"
//	@Failure		400	{object}	map[string]interface{}	"Invalid user ID format"
//	@Failure		404	{object}	map[string]interface{}	"User not found"
//	@Security		BearerAuth
//	@Router			/auth/admin/users/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	// Parse UUID
	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
			"code":  "INVALID_USER_ID",
		})
		return
	}

	// Get user
	user, err := h.authService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}
