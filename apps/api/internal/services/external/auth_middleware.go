package external

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/audit"
)

// AuthMiddleware provides authentication middleware for external API access
type AuthMiddleware interface {
	// RequireAPIKey validates API key authentication
	RequireAPIKey(provider string) gin.HandlerFunc

	// RequireJWT validates JWT token authentication
	RequireJWT(provider string, validateOptions *JWTValidationOptions) gin.HandlerFunc

	// RequireOAuth validates OAuth token authentication
	RequireOAuth(provider string, requiredScopes []string) gin.HandlerFunc

	// RequireAnyAuth validates any supported authentication method
	RequireAnyAuth(provider string, methods []AuthMethod) gin.HandlerFunc

	// InjectAuthContext injects authentication context into request
	InjectAuthContext() gin.HandlerFunc

	// RateLimitByToken applies rate limiting based on token
	RateLimitByToken(limit int, window time.Duration) gin.HandlerFunc
}

// AuthMethod represents different authentication methods
type AuthMethod string

const (
	AuthMethodAPIKey AuthMethod = "api_key"
	AuthMethodJWT    AuthMethod = "jwt"
	AuthMethodOAuth  AuthMethod = "oauth"
	AuthMethodBearer AuthMethod = "bearer"
)

// JWTValidationOptions represents JWT validation options
type JWTValidationOptions struct {
	ValidateExp bool   `json:"validate_exp"`
	ValidateIat bool   `json:"validate_iat"`
	ValidateNbf bool   `json:"validate_nbf"`
	ValidateAud bool   `json:"validate_aud"`
	ValidateIss bool   `json:"validate_iss"`
	ExpectedAud string `json:"expected_aud,omitempty"`
	ExpectedIss string `json:"expected_iss,omitempty"`
}

// AuthContext represents authentication context in requests
type AuthContext struct {
	TokenID       string                 `json:"token_id"`
	Provider      string                 `json:"provider"`
	TokenType     TokenType              `json:"token_type"`
	UserID        string                 `json:"user_id"`
	SessionID     string                 `json:"session_id,omitempty"`
	Scopes        []string               `json:"scopes,omitempty"`
	ExpiresAt     *time.Time             `json:"expires_at,omitempty"`
	IssuedAt      *time.Time             `json:"issued_at,omitempty"`
	Claims        map[string]interface{} `json:"claims,omitempty"`
	RateLimitInfo *RateLimitInfo         `json:"rate_limit_info,omitempty"`
}

// RateLimitInfo represents rate limiting information
type RateLimitInfo struct {
	Limit     int           `json:"limit"`
	Remaining int           `json:"remaining"`
	Reset     time.Time     `json:"reset"`
	Window    time.Duration `json:"window"`
}

// AuthError represents authentication errors
type AuthError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	Provider   string `json:"provider,omitempty"`
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// MiddlewareConfig represents middleware configuration
type MiddlewareConfig struct {
	TokenValidator      TokenValidator `json:"-"`
	TokenManager        TokenManager   `json:"-"`
	AuditService        audit.Service  `json:"-"`
	EnableRateLimit     bool           `json:"enable_rate_limit"`
	EnableAuditLog      bool           `json:"enable_audit_log"`
	ErrorResponseFormat string         `json:"error_response_format"`
	TokenHeaderName     string         `json:"token_header_name"`
	AlternativeHeaders  []string       `json:"alternative_headers"`
	CacheValidationTTL  time.Duration  `json:"cache_validation_ttl"`
}

// authMiddlewareImpl implements AuthMiddleware interface
type authMiddlewareImpl struct {
	config          *MiddlewareConfig
	tokenValidator  TokenValidator
	tokenManager    TokenManager
	auditSvc        audit.Service
	validationCache map[string]*cacheEntry
	rateLimiters    map[string]*rateLimiter
}

// cacheEntry represents cached validation result
type cacheEntry struct {
	result    *TokenValidationResponse
	expiresAt time.Time
}

// rateLimiter represents a token-based rate limiter
type rateLimiter struct {
	requests []time.Time
	limit    int
	window   time.Duration
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(tokenValidator TokenValidator, tokenManager TokenManager, auditSvc audit.Service, config *MiddlewareConfig) AuthMiddleware {
	if config == nil {
		config = &MiddlewareConfig{
			EnableRateLimit:     true,
			EnableAuditLog:      true,
			ErrorResponseFormat: "json",
			TokenHeaderName:     "Authorization",
			AlternativeHeaders:  []string{"X-API-Key", "X-Auth-Token"},
			CacheValidationTTL:  5 * time.Minute,
		}
	}

	middleware := &authMiddlewareImpl{
		config:          config,
		tokenValidator:  tokenValidator,
		tokenManager:    tokenManager,
		auditSvc:        auditSvc,
		validationCache: make(map[string]*cacheEntry),
		rateLimiters:    make(map[string]*rateLimiter),
	}

	log.Println("Authentication middleware initialized")
	return middleware
}

// RequireAPIKey validates API key authentication
func (am *authMiddlewareImpl) RequireAPIKey(provider string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := am.getRequestID(c)
		userID := am.getUserID(c)

		// Extract API key from headers
		apiKey := am.extractTokenFromHeaders(c, []string{"X-API-Key", "Authorization"})
		if apiKey == "" {
			am.handleAuthError(c, &AuthError{
				Code:       "MISSING_API_KEY",
				Message:    "API key is required",
				StatusCode: http.StatusUnauthorized,
				Provider:   provider,
			})
			return
		}

		// Clean Bearer prefix if present
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")

		// Validate API key
		validationReq := &ValidateAPIKeyRequest{
			Provider:  provider,
			APIKey:    apiKey,
			UserID:    userID,
			RequestID: requestID,
		}

		result, err := am.tokenValidator.ValidateAPIKey(c.Request.Context(), validationReq)
		if err != nil {
			am.handleAuthError(c, &AuthError{
				Code:       "VALIDATION_FAILED",
				Message:    fmt.Sprintf("Token validation failed: %v", err),
				StatusCode: http.StatusInternalServerError,
				Provider:   provider,
			})
			return
		}

		if !result.IsValid {
			am.handleAuthError(c, &AuthError{
				Code:       "INVALID_API_KEY",
				Message:    "Invalid API key",
				StatusCode: http.StatusUnauthorized,
				Provider:   provider,
			})
			return
		}

		// Create auth context
		authCtx := &AuthContext{
			TokenID:   result.TokenID,
			Provider:  provider,
			TokenType: TokenTypeAPIKey,
			UserID:    userID,
			Scopes:    result.Scopes,
			ExpiresAt: result.ExpiresAt,
		}

		// Set context and continue
		c.Set("auth_context", authCtx)
		am.logAuthSuccess(c, "api_key", provider, userID)
		c.Next()
	}
}

// RequireJWT validates JWT token authentication
func (am *authMiddlewareImpl) RequireJWT(provider string, validateOptions *JWTValidationOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := am.getRequestID(c)
		userID := am.getUserID(c)

		// Extract JWT token
		token := am.extractTokenFromHeaders(c, []string{"Authorization"})
		if token == "" {
			am.handleAuthError(c, &AuthError{
				Code:       "MISSING_JWT_TOKEN",
				Message:    "JWT token is required",
				StatusCode: http.StatusUnauthorized,
				Provider:   provider,
			})
			return
		}

		// Clean Bearer prefix
		token = strings.TrimPrefix(token, "Bearer ")

		// Validate JWT
		validationReq := &ValidateJWTRequest{
			Provider:    provider,
			Token:       token,
			UserID:      userID,
			RequestID:   requestID,
			ValidateExp: true,
			ValidateIat: true,
			ValidateNbf: true,
		}

		if validateOptions != nil {
			validationReq.ValidateAud = validateOptions.ValidateAud
			validationReq.ValidateIss = validateOptions.ValidateIss
			validationReq.ExpectedAud = validateOptions.ExpectedAud
			validationReq.ExpectedIss = validateOptions.ExpectedIss
		}

		result, err := am.tokenValidator.ValidateJWT(c.Request.Context(), validationReq)
		if err != nil {
			am.handleAuthError(c, &AuthError{
				Code:       "JWT_VALIDATION_FAILED",
				Message:    fmt.Sprintf("JWT validation failed: %v", err),
				StatusCode: http.StatusInternalServerError,
				Provider:   provider,
			})
			return
		}

		if !result.IsValid {
			am.handleAuthError(c, &AuthError{
				Code:       "INVALID_JWT_TOKEN",
				Message:    "Invalid JWT token",
				StatusCode: http.StatusUnauthorized,
				Provider:   provider,
			})
			return
		}

		// Create auth context
		authCtx := &AuthContext{
			TokenID:   result.TokenID,
			Provider:  provider,
			TokenType: TokenTypeJWT,
			UserID:    userID,
			Scopes:    result.Scopes,
			ExpiresAt: result.ExpiresAt,
			IssuedAt:  result.IssuedAt,
			Claims:    result.Claims,
		}

		// Set context and continue
		c.Set("auth_context", authCtx)
		am.logAuthSuccess(c, "jwt", provider, userID)
		c.Next()
	}
}

// RequireOAuth validates OAuth token authentication
func (am *authMiddlewareImpl) RequireOAuth(provider string, requiredScopes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := am.getRequestID(c)
		userID := am.getUserID(c)

		// Extract OAuth token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			am.handleAuthError(c, &AuthError{
				Code:       "MISSING_OAUTH_TOKEN",
				Message:    "OAuth token is required",
				StatusCode: http.StatusUnauthorized,
				Provider:   provider,
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			am.handleAuthError(c, &AuthError{
				Code:       "INVALID_AUTH_HEADER",
				Message:    "Authorization header must be 'Bearer <token>'",
				StatusCode: http.StatusUnauthorized,
				Provider:   provider,
			})
			return
		}

		accessToken := parts[1]

		// Validate OAuth token
		validationReq := &ValidateOAuthRequest{
			Provider:      provider,
			AccessToken:   accessToken,
			TokenType:     "Bearer",
			UserID:        userID,
			RequestID:     requestID,
			RequiredScope: requiredScopes,
		}

		result, err := am.tokenValidator.ValidateOAuth(c.Request.Context(), validationReq)
		if err != nil {
			am.handleAuthError(c, &AuthError{
				Code:       "OAUTH_VALIDATION_FAILED",
				Message:    fmt.Sprintf("OAuth validation failed: %v", err),
				StatusCode: http.StatusInternalServerError,
				Provider:   provider,
			})
			return
		}

		if !result.IsValid {
			am.handleAuthError(c, &AuthError{
				Code:       "INVALID_OAUTH_TOKEN",
				Message:    "Invalid OAuth token",
				StatusCode: http.StatusUnauthorized,
				Provider:   provider,
			})
			return
		}

		// Check required scopes
		if len(requiredScopes) > 0 {
			if !am.hasRequiredScopes(result.Scopes, requiredScopes) {
				am.handleAuthError(c, &AuthError{
					Code:       "INSUFFICIENT_SCOPE",
					Message:    "Token does not have required scopes",
					StatusCode: http.StatusForbidden,
					Provider:   provider,
				})
				return
			}
		}

		// Create auth context
		authCtx := &AuthContext{
			TokenID:   result.TokenID,
			Provider:  provider,
			TokenType: TokenTypeOAuth,
			UserID:    userID,
			Scopes:    result.Scopes,
			ExpiresAt: result.ExpiresAt,
		}

		// Set context and continue
		c.Set("auth_context", authCtx)
		am.logAuthSuccess(c, "oauth", provider, userID)
		c.Next()
	}
}

// RequireAnyAuth validates any supported authentication method
func (am *authMiddlewareImpl) RequireAnyAuth(provider string, methods []AuthMethod) gin.HandlerFunc {
	return func(c *gin.Context) {
		var lastError *AuthError

		for _, method := range methods {
			switch method {
			case AuthMethodAPIKey:
				handler := am.RequireAPIKey(provider)
				// Create a test context to check if auth succeeds
				if am.testAuthMethod(c, handler) {
					handler(c)
					return
				}
			case AuthMethodJWT:
				handler := am.RequireJWT(provider, nil)
				if am.testAuthMethod(c, handler) {
					handler(c)
					return
				}
			case AuthMethodOAuth, AuthMethodBearer:
				handler := am.RequireOAuth(provider, []string{})
				if am.testAuthMethod(c, handler) {
					handler(c)
					return
				}
			}
		}

		// If no method succeeded, return generic auth error
		if lastError == nil {
			lastError = &AuthError{
				Code:       "AUTHENTICATION_REQUIRED",
				Message:    "Valid authentication is required",
				StatusCode: http.StatusUnauthorized,
				Provider:   provider,
			}
		}

		am.handleAuthError(c, lastError)
	}
}

// InjectAuthContext injects authentication context into request
func (am *authMiddlewareImpl) InjectAuthContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract any available auth info and inject into context
		requestID := am.getRequestID(c)
		userID := am.getUserID(c)

		authCtx := &AuthContext{
			UserID: userID,
		}

		// Try to extract token info without validation
		if token := am.extractTokenFromHeaders(c, []string{"Authorization", "X-API-Key"}); token != "" {
			authCtx.TokenID = am.generateTokenID(token)
		}

		c.Set("auth_context", authCtx)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// RateLimitByToken applies rate limiting based on token
func (am *authMiddlewareImpl) RateLimitByToken(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !am.config.EnableRateLimit {
			c.Next()
			return
		}

		// Get token identifier
		tokenID := ""
		if authCtx, exists := c.Get("auth_context"); exists {
			if ctx, ok := authCtx.(*AuthContext); ok {
				tokenID = ctx.TokenID
			}
		}

		if tokenID == "" {
			// Use IP as fallback
			tokenID = c.ClientIP()
		}

		// Check rate limit
		if am.isRateLimited(tokenID, limit, window) {
			am.handleAuthError(c, &AuthError{
				Code:       "RATE_LIMIT_EXCEEDED",
				Message:    "Rate limit exceeded",
				StatusCode: http.StatusTooManyRequests,
			})
			return
		}

		c.Next()
	}
}

// Helper methods

func (am *authMiddlewareImpl) extractTokenFromHeaders(c *gin.Context, headerNames []string) string {
	for _, header := range headerNames {
		if value := c.GetHeader(header); value != "" {
			return value
		}
	}
	return ""
}

func (am *authMiddlewareImpl) getRequestID(c *gin.Context) string {
	if reqID := c.GetHeader("X-Request-ID"); reqID != "" {
		return reqID
	}
	return uuid.New().String()
}

func (am *authMiddlewareImpl) getUserID(c *gin.Context) string {
	if userID := c.GetHeader("X-User-ID"); userID != "" {
		return userID
	}
	// Could also extract from JWT claims or session
	return "anonymous"
}

func (am *authMiddlewareImpl) generateTokenID(token string) string {
	// Generate a hash or ID from token for identification
	return fmt.Sprintf("token_%x", len(token)) // Simplified
}

func (am *authMiddlewareImpl) hasRequiredScopes(granted, required []string) bool {
	if len(required) == 0 {
		return true
	}

	for _, req := range required {
		found := false
		for _, grant := range granted {
			if grant == req {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (am *authMiddlewareImpl) testAuthMethod(c *gin.Context, handler gin.HandlerFunc) bool {
	// This is a simplified test - in practice you'd create a test context
	// and check if the handler would succeed
	return am.extractTokenFromHeaders(c, []string{"Authorization", "X-API-Key"}) != ""
}

func (am *authMiddlewareImpl) isRateLimited(tokenID string, limit int, window time.Duration) bool {
	now := time.Now()

	// Get or create rate limiter for this token
	limiter, exists := am.rateLimiters[tokenID]
	if !exists {
		limiter = &rateLimiter{
			requests: make([]time.Time, 0),
			limit:    limit,
			window:   window,
		}
		am.rateLimiters[tokenID] = limiter
	}

	// Clean old requests
	var validRequests []time.Time
	for _, reqTime := range limiter.requests {
		if reqTime.Add(window).After(now) {
			validRequests = append(validRequests, reqTime)
		}
	}
	limiter.requests = validRequests

	// Check if limit exceeded
	if len(limiter.requests) >= limit {
		return true
	}

	// Add current request
	limiter.requests = append(limiter.requests, now)
	return false
}

func (am *authMiddlewareImpl) handleAuthError(c *gin.Context, authErr *AuthError) {
	// Log the authentication failure
	if am.config.EnableAuditLog && am.auditSvc != nil {
		am.logAuthFailure(c, authErr)
	}

	// Set response headers
	c.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"%s\"", authErr.Provider))

	// Return error response
	c.JSON(authErr.StatusCode, gin.H{
		"error": gin.H{
			"code":    authErr.Code,
			"message": authErr.Message,
			"type":    "authentication_error",
		},
		"timestamp": time.Now().UTC(),
	})
	c.Abort()
}

func (am *authMiddlewareImpl) logAuthSuccess(c *gin.Context, method, provider, userID string) {
	if !am.config.EnableAuditLog || am.auditSvc == nil {
		return
	}

	var uid uuid.UUID
	if userID != "" && userID != "anonymous" {
		uid = uuid.MustParse(userID)
	}

	auditReq := models.AuditLogRequest{
		UserID:           &uid,
		QueryText:        fmt.Sprintf("Authentication via %s for provider %s", method, provider),
		GeneratedCommand: fmt.Sprintf("Successful %s authentication for %s", method, provider),
		SafetyLevel:      models.SafetyLevelSafe,
		ExecutionResult: map[string]interface{}{
			"method":     method,
			"provider":   provider,
			"ip_address": c.ClientIP(),
			"user_agent": c.GetHeader("User-Agent"),
		},
		ExecutionStatus: models.ExecutionStatusSuccess,
	}

	if err := am.auditSvc.LogUserAction(c.Request.Context(), auditReq); err != nil {
		log.Printf("Failed to log auth success: %v", err)
	}
}

func (am *authMiddlewareImpl) logAuthFailure(c *gin.Context, authErr *AuthError) {
	if !am.config.EnableAuditLog || am.auditSvc == nil {
		return
	}

	auditReq := models.AuditLogRequest{
		QueryText:        fmt.Sprintf("Authentication failure for provider %s", authErr.Provider),
		GeneratedCommand: fmt.Sprintf("Failed authentication attempt for %s: %s", authErr.Provider, authErr.Message),
		SafetyLevel:      models.SafetyLevelWarning,
		ExecutionResult: map[string]interface{}{
			"error_code": authErr.Code,
			"message":    authErr.Message,
			"provider":   authErr.Provider,
			"ip_address": c.ClientIP(),
			"user_agent": c.GetHeader("User-Agent"),
		},
		ExecutionStatus: models.ExecutionStatusFailed,
	}

	if err := am.auditSvc.LogUserAction(c.Request.Context(), auditReq); err != nil {
		log.Printf("Failed to log auth failure: %v", err)
	}
}
