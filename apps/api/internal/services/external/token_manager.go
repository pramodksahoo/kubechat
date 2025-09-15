package external

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/audit"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/credentials"
)

// TokenManager manages external API tokens with secure storage and validation
type TokenManager interface {
	// StoreToken securely stores an API token
	StoreToken(ctx context.Context, req *StoreTokenRequest) (*StoredToken, error)

	// RetrieveToken retrieves a stored token by ID
	RetrieveToken(ctx context.Context, tokenID string) (*StoredToken, error)

	// ValidateToken validates a token's authenticity and expiration
	ValidateToken(ctx context.Context, tokenID string) (*ValidationResult, error)

	// RefreshToken refreshes an expired token
	RefreshToken(ctx context.Context, tokenID string) (*StoredToken, error)

	// RevokeToken revokes a token making it invalid
	RevokeToken(ctx context.Context, tokenID string) error

	// ListTokens returns all tokens for a provider/user
	ListTokens(ctx context.Context, filter *TokenFilter) ([]*TokenInfo, error)

	// RotateToken rotates a token with a new value
	RotateToken(ctx context.Context, tokenID string, newToken string) (*StoredToken, error)

	// GetTokenUsage returns usage statistics for a token
	GetTokenUsage(ctx context.Context, tokenID string) (*TokenUsageStats, error)
}

// StoreTokenRequest represents a request to store a new token
type StoreTokenRequest struct {
	Provider     string                 `json:"provider"`
	TokenType    TokenType              `json:"token_type"`
	TokenValue   string                 `json:"token_value"`
	UserID       string                 `json:"user_id"`
	SessionID    string                 `json:"session_id,omitempty"`
	Description  string                 `json:"description"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
	RefreshToken string                 `json:"refresh_token,omitempty"`
	Scopes       []string               `json:"scopes,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// StoredToken represents a securely stored token
type StoredToken struct {
	ID          string                 `json:"id"`
	Provider    string                 `json:"provider"`
	TokenType   TokenType              `json:"token_type"`
	UserID      string                 `json:"user_id"`
	SessionID   string                 `json:"session_id,omitempty"`
	Description string                 `json:"description"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	LastUsed    *time.Time             `json:"last_used,omitempty"`
	UsageCount  int64                  `json:"usage_count"`
	IsActive    bool                   `json:"is_active"`
	Scopes      []string               `json:"scopes,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	TokenHash   string                 `json:"token_hash"` // For validation without exposing token
}

// TokenType represents different types of API tokens
type TokenType string

const (
	TokenTypeBearer    TokenType = "bearer"
	TokenTypeAPIKey    TokenType = "api_key"
	TokenTypeOAuth     TokenType = "oauth"
	TokenTypeJWT       TokenType = "jwt"
	TokenTypeBasicAuth TokenType = "basic_auth"
	TokenTypeCustom    TokenType = "custom"
)

// TokenFilter represents filtering criteria for token queries
type TokenFilter struct {
	Provider     string     `json:"provider,omitempty"`
	UserID       string     `json:"user_id,omitempty"`
	SessionID    string     `json:"session_id,omitempty"`
	TokenType    TokenType  `json:"token_type,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
	ExpiresAfter *time.Time `json:"expires_after,omitempty"`
	ExpiresBefor *time.Time `json:"expires_before,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

// TokenInfo contains non-sensitive information about a token
type TokenInfo struct {
	ID          string     `json:"id"`
	Provider    string     `json:"provider"`
	TokenType   TokenType  `json:"token_type"`
	UserID      string     `json:"user_id"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsed    *time.Time `json:"last_used,omitempty"`
	UsageCount  int64      `json:"usage_count"`
	IsActive    bool       `json:"is_active"`
	Status      string     `json:"status"`
}

// ValidationResult represents token validation results
type ValidationResult struct {
	IsValid        bool           `json:"is_valid"`
	TokenID        string         `json:"token_id"`
	Provider       string         `json:"provider"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	ExpiresIn      *time.Duration `json:"expires_in,omitempty"`
	IsExpired      bool           `json:"is_expired"`
	IsRevoked      bool           `json:"is_revoked"`
	Scopes         []string       `json:"scopes,omitempty"`
	ValidationTime time.Time      `json:"validation_time"`
	Errors         []string       `json:"errors,omitempty"`
	Warnings       []string       `json:"warnings,omitempty"`
}

// TokenUsageStats represents usage statistics for a token
type TokenUsageStats struct {
	TokenID       string       `json:"token_id"`
	TotalRequests int64        `json:"total_requests"`
	SuccessCount  int64        `json:"success_count"`
	ErrorCount    int64        `json:"error_count"`
	LastUsed      time.Time    `json:"last_used"`
	FirstUsed     time.Time    `json:"first_used"`
	DailyUsage    []DailyUsage `json:"daily_usage"`
	AverageRPS    float64      `json:"average_rps"`
}

// DailyUsage represents token usage for a specific day
type DailyUsage struct {
	Date     time.Time `json:"date"`
	Requests int64     `json:"requests"`
	Errors   int64     `json:"errors"`
}

// TokenManagerConfig represents token manager configuration
type TokenManagerConfig struct {
	DefaultTokenTTL     time.Duration `json:"default_token_ttl"`
	RefreshThreshold    time.Duration `json:"refresh_threshold"`
	EnableAutoRefresh   bool          `json:"enable_auto_refresh"`
	MaxTokensPerUser    int           `json:"max_tokens_per_user"`
	EnableUsageTracking bool          `json:"enable_usage_tracking"`
	EncryptionEnabled   bool          `json:"encryption_enabled"`
	AuditEnabled        bool          `json:"audit_enabled"`
}

// tokenManagerImpl implements TokenManager interface
type tokenManagerImpl struct {
	credSvc   credentials.Service
	auditSvc  audit.Service
	config    *TokenManagerConfig
	tokens    map[string]*StoredToken
	usage     map[string]*TokenUsageStats
	jwtSecret []byte
	mu        sync.RWMutex
}

// NewTokenManager creates a new token manager
func NewTokenManager(credSvc credentials.Service, auditSvc audit.Service, config *TokenManagerConfig) (TokenManager, error) {
	if config == nil {
		config = &TokenManagerConfig{
			DefaultTokenTTL:     24 * time.Hour,
			RefreshThreshold:    2 * time.Hour,
			EnableAutoRefresh:   true,
			MaxTokensPerUser:    10,
			EnableUsageTracking: true,
			EncryptionEnabled:   true,
			AuditEnabled:        true,
		}
	}

	// Generate JWT secret
	jwtSecret := make([]byte, 32)
	if _, err := rand.Read(jwtSecret); err != nil {
		return nil, fmt.Errorf("failed to generate JWT secret: %w", err)
	}

	tm := &tokenManagerImpl{
		credSvc:   credSvc,
		auditSvc:  auditSvc,
		config:    config,
		tokens:    make(map[string]*StoredToken),
		usage:     make(map[string]*TokenUsageStats),
		jwtSecret: jwtSecret,
	}

	log.Println("Token manager initialized")
	return tm, nil
}

// StoreToken securely stores an API token
func (tm *tokenManagerImpl) StoreToken(ctx context.Context, req *StoreTokenRequest) (*StoredToken, error) {
	if err := tm.validateStoreRequest(req); err != nil {
		return nil, fmt.Errorf("invalid store request: %w", err)
	}

	// Check user token limit
	if err := tm.checkTokenLimit(ctx, req.UserID); err != nil {
		return nil, fmt.Errorf("token limit exceeded: %w", err)
	}

	tokenID := uuid.New().String()
	now := time.Now()

	// Calculate expiry if not provided
	expiresAt := req.ExpiresAt
	if expiresAt == nil {
		expires := now.Add(tm.config.DefaultTokenTTL)
		expiresAt = &expires
	}

	// Create token hash for validation
	tokenHash, err := tm.generateTokenHash(req.TokenValue)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token hash: %w", err)
	}

	storedToken := &StoredToken{
		ID:          tokenID,
		Provider:    req.Provider,
		TokenType:   req.TokenType,
		UserID:      req.UserID,
		SessionID:   req.SessionID,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiresAt:   expiresAt,
		UsageCount:  0,
		IsActive:    true,
		Scopes:      req.Scopes,
		Metadata:    req.Metadata,
		TokenHash:   tokenHash,
	}

	// Store token using credentials service
	cred := &credentials.Credential{
		Name:        fmt.Sprintf("token-%s", tokenID),
		Type:        credentials.CredentialTypeToken,
		Value:       req.TokenValue,
		Description: fmt.Sprintf("API token for %s (%s)", req.Provider, req.Description),
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiresAt:   expiresAt,
		Tags:        []string{req.Provider, string(req.TokenType), req.UserID},
		Metadata: map[string]string{
			"token_id":   tokenID,
			"provider":   req.Provider,
			"token_type": string(req.TokenType),
			"user_id":    req.UserID,
			"session_id": req.SessionID,
		},
	}

	if req.RefreshToken != "" {
		cred.Metadata["refresh_token"] = req.RefreshToken
	}

	if err := tm.credSvc.SetCredential(ctx, cred); err != nil {
		return nil, fmt.Errorf("failed to store token: %w", err)
	}

	// Store token metadata
	tm.mu.Lock()
	tm.tokens[tokenID] = storedToken

	// Initialize usage tracking
	if tm.config.EnableUsageTracking {
		tm.usage[tokenID] = &TokenUsageStats{
			TokenID:   tokenID,
			FirstUsed: now,
		}
	}
	tm.mu.Unlock()

	// Log audit entry
	tm.logTokenOperation(ctx, tokenID, "store", req.UserID, nil)

	log.Printf("Token stored successfully: %s for provider %s", tokenID, req.Provider)
	return storedToken, nil
}

// RetrieveToken retrieves a stored token by ID
func (tm *tokenManagerImpl) RetrieveToken(ctx context.Context, tokenID string) (*StoredToken, error) {
	// Get token metadata
	tm.mu.RLock()
	storedToken, exists := tm.tokens[tokenID]
	tm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("token %s not found", tokenID)
	}

	// Retrieve actual token value from credentials service
	credName := fmt.Sprintf("token-%s", tokenID)
	_, err := tm.credSvc.GetCredential(ctx, credName)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve token: %w", err)
	}

	// Update last used
	tm.mu.Lock()
	now := time.Now()
	storedToken.LastUsed = &now
	storedToken.UsageCount++

	if tm.config.EnableUsageTracking {
		if usage, exists := tm.usage[tokenID]; exists {
			usage.TotalRequests++
			usage.LastUsed = now
		}
	}
	tm.mu.Unlock()

	// Don't expose the actual token value in the response for security
	result := *storedToken

	// Log audit entry for token access
	tm.logTokenOperation(ctx, tokenID, "retrieve", "", nil)

	return &result, nil
}

// ValidateToken validates a token's authenticity and expiration
func (tm *tokenManagerImpl) ValidateToken(ctx context.Context, tokenID string) (*ValidationResult, error) {
	tm.mu.RLock()
	storedToken, exists := tm.tokens[tokenID]
	tm.mu.RUnlock()

	result := &ValidationResult{
		TokenID:        tokenID,
		ValidationTime: time.Now(),
		IsValid:        false,
		Errors:         []string{},
		Warnings:       []string{},
	}

	if !exists {
		result.Errors = append(result.Errors, "Token not found")
		return result, nil
	}

	result.Provider = storedToken.Provider
	result.ExpiresAt = storedToken.ExpiresAt
	result.Scopes = storedToken.Scopes

	// Check if token is active
	if !storedToken.IsActive {
		result.IsRevoked = true
		result.Errors = append(result.Errors, "Token has been revoked")
		return result, nil
	}

	// Check expiration
	if storedToken.ExpiresAt != nil && storedToken.ExpiresAt.Before(time.Now()) {
		result.IsExpired = true
		result.Errors = append(result.Errors, "Token has expired")
		return result, nil
	}

	// Calculate time until expiration
	if storedToken.ExpiresAt != nil {
		expiresIn := time.Until(*storedToken.ExpiresAt)
		result.ExpiresIn = &expiresIn

		// Add warning if token expires soon
		if expiresIn < tm.config.RefreshThreshold {
			result.Warnings = append(result.Warnings, "Token expires soon")
		}
	}

	result.IsValid = true

	// Log validation attempt
	tm.logTokenOperation(ctx, tokenID, "validate", "", nil)

	return result, nil
}

// RefreshToken refreshes an expired token
func (tm *tokenManagerImpl) RefreshToken(ctx context.Context, tokenID string) (*StoredToken, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	storedToken, exists := tm.tokens[tokenID]
	if !exists {
		return nil, fmt.Errorf("token %s not found", tokenID)
	}

	// Get the credential to access refresh token
	credName := fmt.Sprintf("token-%s", tokenID)
	cred, err := tm.credSvc.GetCredential(ctx, credName)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential for refresh: %w", err)
	}

	refreshToken, hasRefresh := cred.Metadata["refresh_token"]
	if !hasRefresh || refreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	// Here you would typically make a call to the provider's refresh endpoint
	// For now, we'll simulate a refresh by extending the expiry
	now := time.Now()
	newExpiry := now.Add(tm.config.DefaultTokenTTL)

	storedToken.UpdatedAt = now
	storedToken.ExpiresAt = &newExpiry

	// Update credential
	cred.UpdatedAt = now
	cred.ExpiresAt = &newExpiry

	if err := tm.credSvc.UpdateCredential(ctx, cred); err != nil {
		return nil, fmt.Errorf("failed to update credential: %w", err)
	}

	// Log refresh operation
	tm.logTokenOperation(ctx, tokenID, "refresh", storedToken.UserID, nil)

	log.Printf("Token refreshed successfully: %s", tokenID)
	return storedToken, nil
}

// RevokeToken revokes a token making it invalid
func (tm *tokenManagerImpl) RevokeToken(ctx context.Context, tokenID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	storedToken, exists := tm.tokens[tokenID]
	if !exists {
		return fmt.Errorf("token %s not found", tokenID)
	}

	// Mark as inactive
	storedToken.IsActive = false
	storedToken.UpdatedAt = time.Now()

	// Delete the actual credential
	credName := fmt.Sprintf("token-%s", tokenID)
	if err := tm.credSvc.DeleteCredential(ctx, credName); err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	// Log revocation
	tm.logTokenOperation(ctx, tokenID, "revoke", storedToken.UserID, nil)

	log.Printf("Token revoked successfully: %s", tokenID)
	return nil
}

// ListTokens returns all tokens for a provider/user
func (tm *tokenManagerImpl) ListTokens(ctx context.Context, filter *TokenFilter) ([]*TokenInfo, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var tokens []*TokenInfo

	for _, token := range tm.tokens {
		if tm.matchesFilter(token, filter) {
			info := &TokenInfo{
				ID:          token.ID,
				Provider:    token.Provider,
				TokenType:   token.TokenType,
				UserID:      token.UserID,
				Description: token.Description,
				CreatedAt:   token.CreatedAt,
				ExpiresAt:   token.ExpiresAt,
				LastUsed:    token.LastUsed,
				UsageCount:  token.UsageCount,
				IsActive:    token.IsActive,
				Status:      tm.getTokenStatus(token),
			}
			tokens = append(tokens, info)
		}
	}

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(tokens) {
			tokens = tokens[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(tokens) {
			tokens = tokens[:filter.Limit]
		}
	}

	return tokens, nil
}

// RotateToken rotates a token with a new value
func (tm *tokenManagerImpl) RotateToken(ctx context.Context, tokenID string, newToken string) (*StoredToken, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	storedToken, exists := tm.tokens[tokenID]
	if !exists {
		return nil, fmt.Errorf("token %s not found", tokenID)
	}

	// Generate new token hash
	tokenHash, err := tm.generateTokenHash(newToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token hash: %w", err)
	}

	// Update stored token
	now := time.Now()
	storedToken.TokenHash = tokenHash
	storedToken.UpdatedAt = now

	// Update credential with new token value
	credName := fmt.Sprintf("token-%s", tokenID)
	cred, err := tm.credSvc.GetCredential(ctx, credName)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	cred.Value = newToken
	cred.UpdatedAt = now
	cred.RotatedAt = &now

	if err := tm.credSvc.UpdateCredential(ctx, cred); err != nil {
		return nil, fmt.Errorf("failed to update credential: %w", err)
	}

	// Log rotation
	tm.logTokenOperation(ctx, tokenID, "rotate", storedToken.UserID, nil)

	log.Printf("Token rotated successfully: %s", tokenID)
	return storedToken, nil
}

// GetTokenUsage returns usage statistics for a token
func (tm *tokenManagerImpl) GetTokenUsage(ctx context.Context, tokenID string) (*TokenUsageStats, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if !tm.config.EnableUsageTracking {
		return nil, fmt.Errorf("usage tracking is disabled")
	}

	usage, exists := tm.usage[tokenID]
	if !exists {
		return nil, fmt.Errorf("usage statistics not found for token %s", tokenID)
	}

	return usage, nil
}

// Helper methods

func (tm *tokenManagerImpl) validateStoreRequest(req *StoreTokenRequest) error {
	if req.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	if req.TokenValue == "" {
		return fmt.Errorf("token value is required")
	}

	if req.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if req.TokenType == "" {
		req.TokenType = TokenTypeBearer // Default
	}

	return nil
}

func (tm *tokenManagerImpl) checkTokenLimit(ctx context.Context, userID string) error {
	if tm.config.MaxTokensPerUser <= 0 {
		return nil // No limit
	}

	count := 0
	for _, token := range tm.tokens {
		if token.UserID == userID && token.IsActive {
			count++
		}
	}

	if count >= tm.config.MaxTokensPerUser {
		return fmt.Errorf("maximum tokens per user exceeded (%d)", tm.config.MaxTokensPerUser)
	}

	return nil
}

func (tm *tokenManagerImpl) generateTokenHash(token string) (string, error) {
	// Simple hash for now - in production you'd use a more secure method
	return base64.StdEncoding.EncodeToString([]byte(token)), nil
}

func (tm *tokenManagerImpl) matchesFilter(token *StoredToken, filter *TokenFilter) bool {
	if filter == nil {
		return true
	}

	if filter.Provider != "" && token.Provider != filter.Provider {
		return false
	}

	if filter.UserID != "" && token.UserID != filter.UserID {
		return false
	}

	if filter.SessionID != "" && token.SessionID != filter.SessionID {
		return false
	}

	if filter.TokenType != "" && token.TokenType != filter.TokenType {
		return false
	}

	if filter.IsActive != nil && token.IsActive != *filter.IsActive {
		return false
	}

	if filter.ExpiresAfter != nil && (token.ExpiresAt == nil || token.ExpiresAt.Before(*filter.ExpiresAfter)) {
		return false
	}

	if filter.ExpiresBefor != nil && (token.ExpiresAt != nil && token.ExpiresAt.After(*filter.ExpiresBefor)) {
		return false
	}

	return true
}

func (tm *tokenManagerImpl) getTokenStatus(token *StoredToken) string {
	if !token.IsActive {
		return "revoked"
	}

	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		return "expired"
	}

	if token.ExpiresAt != nil && time.Until(*token.ExpiresAt) < tm.config.RefreshThreshold {
		return "expiring_soon"
	}

	return "active"
}

func (tm *tokenManagerImpl) logTokenOperation(ctx context.Context, tokenID, operation, userID string, err error) {
	if !tm.config.AuditEnabled || tm.auditSvc == nil {
		return
	}

	safetyLevel := models.SafetyLevelWarning // Token operations are always security-relevant
	if err != nil {
		safetyLevel = models.SafetyLevelDangerous
	}

	var uid uuid.UUID
	if userID != "" {
		uid = uuid.MustParse(userID)
	}

	auditReq := models.AuditLogRequest{
		UserID:           &uid,
		QueryText:        fmt.Sprintf("Token management: %s", operation),
		GeneratedCommand: fmt.Sprintf("%s operation on token %s", operation, tokenID),
		SafetyLevel:      string(safetyLevel),
		ExecutionResult: map[string]interface{}{
			"token_id":  tokenID,
			"operation": operation,
		},
		ExecutionStatus: func() string {
			if err == nil {
				return models.ExecutionStatusSuccess
			}
			return models.ExecutionStatusFailed
		}(),
	}

	if err != nil {
		auditReq.ExecutionResult["error"] = err.Error()
	}

	if logErr := tm.auditSvc.LogUserAction(ctx, auditReq); logErr != nil {
		log.Printf("Failed to log token operation: %v", logErr)
	}
}
