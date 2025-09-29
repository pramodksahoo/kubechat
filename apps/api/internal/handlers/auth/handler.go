package auth

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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
		admin.POST("/users", h.CreateAdminUser)
		admin.GET("/users/:id", h.GetUser)
		admin.PUT("/users/:id", h.UpdateAdminUser)
		admin.DELETE("/users/:id", h.DeleteAdminUser)
		admin.GET("/roles", h.ListRoles)
		admin.GET("/permissions", h.ListPermissions)
		admin.POST("/credentials/sync/from-k8s", h.SyncAdminCredentialsFromK8s)
		admin.GET("/credentials/sync/status", h.GetAdminCredentialStatus)
	}

	// Secure token cookie management routes
	cookieRoutes := router.Group("/auth")
	{
		cookieRoutes.POST("/set-tokens", h.SetTokensCookie)
		cookieRoutes.GET("/get-token", h.GetTokenCookie)
		cookieRoutes.GET("/get-refresh-token", h.GetRefreshTokenCookie)
		cookieRoutes.POST("/clear-tokens", h.ClearTokensCookie)
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

	// Set secure HTTP-only cookie with the token
	c.SetCookie("kubechat_token", loginResponse.SessionToken, 24*60*60, "/", "", c.Request.TLS != nil, true)

	// Return response with token included for development and proper frontend integration
	responseData := gin.H{
		"message": "Login successful",
		"data": gin.H{
			"user": gin.H{
				"id":       loginResponse.User.ID,
				"username": loginResponse.User.Username,
				"email":    loginResponse.User.Email,
				"role":     loginResponse.User.Role,
			},
			"expires_at": loginResponse.ExpiresAt,
			"token":      loginResponse.SessionToken, // Include token in response for frontend
		},
	}

	c.JSON(http.StatusOK, responseData)
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

	// Clear authentication cookies
	c.SetCookie("kubechat_token", "", -1, "/", "", c.Request.TLS != nil, true)
	c.SetCookie("kubechat_refresh_token", "", -1, "/", "", c.Request.TLS != nil, true)

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

	// Set secure HTTP-only cookie with the new token
	c.SetCookie("kubechat_token", loginResponse.SessionToken, 24*60*60, "/", "", c.Request.TLS != nil, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Session refreshed successfully",
		"data": gin.H{
			"user": gin.H{
				"id":       loginResponse.User.ID,
				"username": loginResponse.User.Username,
				"email":    loginResponse.User.Email,
				"role":     loginResponse.User.Role,
			},
			"expires_at": loginResponse.ExpiresAt,
			"token":      loginResponse.SessionToken, // Include token in response for frontend
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
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit <= 0 {
		limit = 50
	}

	filters := models.UserListFilters{
		Role:   c.Query("role"),
		Status: c.Query("status"),
		Search: c.Query("search"),
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	users, total, svcErr := h.authService.ListAdminUsers(c.Request.Context(), filters)
	if svcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list users",
			"code":  "LIST_USERS_ERROR",
		})
		return
	}

	response := make([]gin.H, 0, len(users))
	for _, user := range users {
		response = append(response, toAdminUserResponse(user))
	}

	c.JSON(http.StatusOK, gin.H{
		"page":  page,
		"limit": limit,
		"total": total,
		"users": response,
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
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, toAdminUserResponse(user))
}

// ListRoles provides available RBAC roles for the admin console
func (h *Handler) ListRoles(c *gin.Context) {
	stamp := time.Now().UTC().Format(time.RFC3339)
	roles := []gin.H{
		{
			"id":           "admin",
			"name":         "Administrator",
			"description":  "Full system access with capability to manage infrastructure and security",
			"permissions":  []string{"system:*,security:*,cluster:*"},
			"isSystemRole": true,
			"createdAt":    stamp,
			"updatedAt":    stamp,
			"userCount":    0,
		},
		{
			"id":           "user",
			"name":         "Operator",
			"description":  "Standard operator access for day-to-day cluster tasks",
			"permissions":  []string{"cluster:read", "cluster:write", "audit:read"},
			"isSystemRole": true,
			"createdAt":    stamp,
			"updatedAt":    stamp,
			"userCount":    0,
		},
		{
			"id":           "viewer",
			"name":         "ReadOnly",
			"description":  "View-only access for compliance and audit teams",
			"permissions":  []string{"cluster:read", "audit:read"},
			"isSystemRole": true,
			"createdAt":    stamp,
			"updatedAt":    stamp,
			"userCount":    0,
		},
		{
			"id":           "auditor",
			"name":         "Auditor",
			"description":  "Audit-focused read access with enhanced reporting visibility",
			"permissions":  []string{"audit:read", "security:read"},
			"isSystemRole": true,
			"createdAt":    stamp,
			"updatedAt":    stamp,
			"userCount":    0,
		},
		{
			"id":           "compliance_officer",
			"name":         "Compliance Officer",
			"description":  "Compliance oversight with ability to manage legal holds",
			"permissions":  []string{"audit:read", "compliance:manage"},
			"isSystemRole": true,
			"createdAt":    stamp,
			"updatedAt":    stamp,
			"userCount":    0,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
		"total": len(roles),
	})
}

// ListPermissions returns static permission catalog for RBAC UI
func (h *Handler) ListPermissions(c *gin.Context) {
	permissions := []gin.H{
		{"id": "system:manage", "name": "Manage Platform", "description": "Full system management", "resource": "system", "action": "manage", "category": "system"},
		{"id": "cluster:read", "name": "View Cluster", "description": "Read-only cluster visibility", "resource": "cluster", "action": "read", "category": "cluster"},
		{"id": "cluster:write", "name": "Modify Cluster", "description": "Apply cluster changes", "resource": "cluster", "action": "write", "category": "cluster"},
		{"id": "audit:read", "name": "View Audit Logs", "description": "Access audit trail records", "resource": "audit", "action": "read", "category": "user"},
		{"id": "security:read", "name": "View Security Events", "description": "Read security alerts and posture", "resource": "security", "action": "read", "category": "system"},
		{"id": "compliance:manage", "name": "Manage Compliance", "description": "Administer compliance holds", "resource": "compliance", "action": "manage", "category": "user"},
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
	})
}

// SyncAdminCredentialsFromK8s triggers a credential sync placeholder
func (h *Handler) SyncAdminCredentialsFromK8s(c *gin.Context) {
	c.JSON(http.StatusOK, buildCredentialSyncStatus("k8s_secret"))
}

// GetAdminCredentialStatus returns the latest credential sync metadata placeholder
func (h *Handler) GetAdminCredentialStatus(c *gin.Context) {
	c.JSON(http.StatusOK, buildCredentialSyncStatus("k8s_secret"))
}

func buildCredentialSyncStatus(source string) gin.H {
	now := time.Now().UTC()
	return gin.H{
		"sync_status":          "success",
		"sync_timestamp":       now.Format(time.RFC3339),
		"sync_source":          source,
		"sync_type":            "bootstrap",
		"username":             "admin",
		"email":                "admin@kubechat.dev",
		"password_expires_at":  now.Add(90 * 24 * time.Hour).Format(time.RFC3339),
		"rotation_count":       0,
		"compliance_status":    "compliant",
		"k8s_resource_version": "1",
	}
}

// CreateAdminUser creates a new user via admin console
func (h *Handler) CreateAdminUser(c *gin.Context) {
	var req models.AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	user, err := h.authService.CreateAdminUser(c.Request.Context(), &req)
	if err != nil {
		status := http.StatusBadRequest
		code := "CREATE_USER_ERROR"
		msg := err.Error()
		if strings.Contains(msg, "already exists") || strings.Contains(msg, "already in use") {
			status = http.StatusConflict
		} else if err == auth.ErrUserNotFound {
			status = http.StatusNotFound
			code = "USER_NOT_FOUND"
		}

		c.JSON(status, gin.H{
			"error": msg,
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusCreated, toAdminUserResponse(user))
}

// UpdateAdminUser updates an existing user record
func (h *Handler) UpdateAdminUser(c *gin.Context) {
	userID := c.Param("id")
	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
			"code":  "INVALID_USER_ID",
		})
		return
	}

	var req models.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	user, svcErr := h.authService.UpdateAdminUser(c.Request.Context(), id, &req)
	if svcErr != nil {
		status := http.StatusBadRequest
		code := "UPDATE_USER_ERROR"
		msg := svcErr.Error()
		if svcErr == auth.ErrUserNotFound {
			status = http.StatusNotFound
			code = "USER_NOT_FOUND"
		} else if strings.Contains(msg, "already exists") || strings.Contains(msg, "already in use") {
			status = http.StatusConflict
		}

		c.JSON(status, gin.H{
			"error": msg,
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, toAdminUserResponse(user))
}

// DeleteAdminUser removes a user via soft delete
func (h *Handler) DeleteAdminUser(c *gin.Context) {
	userID := c.Param("id")
	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
			"code":  "INVALID_USER_ID",
		})
		return
	}

	err = h.authService.DeleteAdminUser(c.Request.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		code := "DELETE_USER_ERROR"
		if err == auth.ErrUserNotFound {
			status = http.StatusNotFound
			code = "USER_NOT_FOUND"
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
			"code":  code,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func toAdminUserResponse(user *models.User) gin.H {
	var lastLogin string
	if user.LastLogin != nil {
		lastLogin = user.LastLogin.UTC().Format(time.RFC3339)
	}

	var lockUntil string
	locked := false
	if user.AccountLockedUntil != nil {
		lockUntil = user.AccountLockedUntil.UTC().Format(time.RFC3339)
		if user.AccountLockedUntil.After(time.Now()) {
			locked = true
		}
	}

	return gin.H{
		"id":                  user.ID,
		"username":            user.Username,
		"email":               user.Email,
		"role":                user.Role,
		"permissions":         []string{},
		"clusters":            []string{},
		"createdAt":           user.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":           user.UpdatedAt.UTC().Format(time.RFC3339),
		"lastLoginAt":         lastLogin,
		"isActive":            user.IsActive,
		"accountLocked":       locked,
		"accountLockedUntil":  lockUntil,
		"failedLoginAttempts": user.FailedLoginAttempts,
		"mfaEnabled":          false,
		"lastPasswordChange":  user.PasswordChangedAt.UTC().Format(time.RFC3339),
		"mustChangePassword":  user.MustChangePassword,
	}
}

// SetTokensCookie handles setting secure HTTP-only cookies for tokens
func (h *Handler) SetTokensCookie(c *gin.Context) {
	var req struct {
		AccessToken  string `json:"accessToken" binding:"required"`
		RefreshToken string `json:"refreshToken"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Set access token cookie
	c.SetCookie("kubechat_token", req.AccessToken, 24*60*60, "/", "", c.Request.TLS != nil, true)

	// Set refresh token cookie if provided
	if req.RefreshToken != "" {
		c.SetCookie("kubechat_refresh_token", req.RefreshToken, 24*60*60*7, "/", "", c.Request.TLS != nil, true)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tokens set successfully",
	})
}

// GetTokenCookie retrieves the access token from HTTP-only cookie
func (h *Handler) GetTokenCookie(c *gin.Context) {
	token, err := c.Cookie("kubechat_token")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"accessToken": "",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken": token,
	})
}

// GetRefreshTokenCookie retrieves the refresh token from HTTP-only cookie
func (h *Handler) GetRefreshTokenCookie(c *gin.Context) {
	token, err := c.Cookie("kubechat_refresh_token")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"refreshToken": "",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refreshToken": token,
	})
}

// ClearTokensCookie clears all authentication cookies
func (h *Handler) ClearTokensCookie(c *gin.Context) {
	// Clear access token cookie
	c.SetCookie("kubechat_token", "", -1, "/", "", c.Request.TLS != nil, true)

	// Clear refresh token cookie
	c.SetCookie("kubechat_refresh_token", "", -1, "/", "", c.Request.TLS != nil, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Tokens cleared successfully",
	})
}
