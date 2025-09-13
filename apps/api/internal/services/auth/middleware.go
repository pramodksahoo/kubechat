package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityMiddleware adds security headers to all responses
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Content Security Policy - prevents XSS attacks
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data: https:; "+
				"connect-src 'self' wss: ws:; "+
				"frame-ancestors 'none'")

		// HSTS - forces HTTPS connections
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// X-Frame-Options - prevents clickjacking
		c.Header("X-Frame-Options", "DENY")

		// X-Content-Type-Options - prevents MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection - enables XSS filtering
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer Policy - controls referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy - controls browser features
		c.Header("Permissions-Policy",
			"camera=(), microphone=(), geolocation=(), interest-cohort=()")

		c.Next()
	}
}

// AuthMiddleware validates JWT tokens for protected routes
func AuthMiddleware(authService Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "AUTH_HEADER_MISSING",
			})
			c.Abort()
			return
		}

		// Check Bearer token format
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Use: Bearer <token>",
				"code":  "INVALID_AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Validate JWT token
		claims, err := authService.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Validate session in database
		session, err := authService.ValidateSession(c.Request.Context(), tokenString)
		if err != nil {
			status := http.StatusUnauthorized
			code := "SESSION_INVALID"

			if err == ErrSessionExpired {
				code = "SESSION_EXPIRED"
			}

			c.JSON(status, gin.H{
				"error": err.Error(),
				"code":  code,
			})
			c.Abort()
			return
		}

		// Store user information in context for handlers to use
		c.Set("user_id", claims.UserID.String())
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("session_id", session.ID.String())
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireRole middleware ensures user has required role
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found in context",
				"code":  "ROLE_CONTEXT_MISSING",
			})
			c.Abort()
			return
		}

		userRoleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user role format",
				"code":  "INVALID_ROLE_FORMAT",
			})
			c.Abort()
			return
		}

		// Check role hierarchy
		hasPermission := false
		switch role {
		case "viewer":
			hasPermission = userRoleStr == "admin" || userRoleStr == "user" || userRoleStr == "viewer"
		case "user":
			hasPermission = userRoleStr == "admin" || userRoleStr == "user"
		case "admin":
			hasPermission = userRoleStr == "admin"
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Insufficient permissions",
				"code":          "INSUFFICIENT_PERMISSIONS",
				"required_role": role,
				"user_role":     userRoleStr,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireWritePermission middleware ensures user can perform write operations
func RequireWritePermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found in context",
				"code":  "ROLE_CONTEXT_MISSING",
			})
			c.Abort()
			return
		}

		userRoleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user role format",
				"code":  "INVALID_ROLE_FORMAT",
			})
			c.Abort()
			return
		}

		// Check if user has write permissions (admin or user)
		if userRoleStr != "admin" && userRoleStr != "user" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":     "Write permissions required",
				"code":      "WRITE_PERMISSION_REQUIRED",
				"user_role": userRoleStr,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ExtractUserContext helper function to get user info from context
func ExtractUserContext(c *gin.Context) (userID string, username string, role string, ok bool) {
	userIDValue, exists1 := c.Get("user_id")
	usernameValue, exists2 := c.Get("username")
	roleValue, exists3 := c.Get("user_role")

	if !exists1 || !exists2 || !exists3 {
		return "", "", "", false
	}

	userID, ok1 := userIDValue.(string)
	username, ok2 := usernameValue.(string)
	role, ok3 := roleValue.(string)

	if !ok1 || !ok2 || !ok3 {
		return "", "", "", false
	}

	return userID, username, role, true
}
