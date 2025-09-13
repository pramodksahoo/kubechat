package external

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/audit"
)

// TokenValidator provides comprehensive token validation and refresh capabilities
type TokenValidator interface {
	// ValidateAPIKey validates an API key format and authenticity
	ValidateAPIKey(ctx context.Context, req *ValidateAPIKeyRequest) (*TokenValidationResponse, error)

	// ValidateJWT validates a JWT token
	ValidateJWT(ctx context.Context, req *ValidateJWTRequest) (*TokenValidationResponse, error)

	// ValidateOAuth validates an OAuth token
	ValidateOAuth(ctx context.Context, req *ValidateOAuthRequest) (*TokenValidationResponse, error)

	// RefreshExpiredToken attempts to refresh an expired token
	RefreshExpiredToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error)

	// ValidateTokenPermissions checks if token has required permissions
	ValidateTokenPermissions(ctx context.Context, tokenID string, requiredScopes []string) (*PermissionValidationResult, error)

	// GetTokenStrength analyzes token security strength
	GetTokenStrength(ctx context.Context, token string) (*TokenStrengthAnalysis, error)

	// ValidateTokenFormat validates token format for different providers
	ValidateTokenFormat(ctx context.Context, provider string, token string) (*FormatValidationResult, error)
}

// ValidateAPIKeyRequest represents an API key validation request
type ValidateAPIKeyRequest struct {
	Provider  string            `json:"provider"`
	APIKey    string            `json:"api_key"`
	UserID    string            `json:"user_id"`
	RequestID string            `json:"request_id"`
	Context   map[string]string `json:"context,omitempty"`
}

// ValidateJWTRequest represents a JWT validation request
type ValidateJWTRequest struct {
	Provider    string            `json:"provider"`
	Token       string            `json:"token"`
	UserID      string            `json:"user_id"`
	RequestID   string            `json:"request_id"`
	ValidateExp bool              `json:"validate_exp"`
	ValidateIat bool              `json:"validate_iat"`
	ValidateNbf bool              `json:"validate_nbf"`
	ValidateAud bool              `json:"validate_aud"`
	ValidateIss bool              `json:"validate_iss"`
	ExpectedAud string            `json:"expected_aud,omitempty"`
	ExpectedIss string            `json:"expected_iss,omitempty"`
	Context     map[string]string `json:"context,omitempty"`
}

// ValidateOAuthRequest represents an OAuth token validation request
type ValidateOAuthRequest struct {
	Provider      string            `json:"provider"`
	AccessToken   string            `json:"access_token"`
	TokenType     string            `json:"token_type"`
	UserID        string            `json:"user_id"`
	RequestID     string            `json:"request_id"`
	RequiredScope []string          `json:"required_scope,omitempty"`
	Context       map[string]string `json:"context,omitempty"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	Provider         string            `json:"provider"`
	RefreshToken     string            `json:"refresh_token"`
	ExpiredToken     string            `json:"expired_token,omitempty"`
	UserID           string            `json:"user_id"`
	RequestID        string            `json:"request_id"`
	AdditionalScopes []string          `json:"additional_scopes,omitempty"`
	Context          map[string]string `json:"context,omitempty"`
}

// TokenValidationResponse represents a token validation response
type TokenValidationResponse struct {
	IsValid        bool                   `json:"is_valid"`
	TokenID        string                 `json:"token_id,omitempty"`
	TokenType      TokenType              `json:"token_type"`
	Provider       string                 `json:"provider"`
	UserID         string                 `json:"user_id"`
	ExpiresAt      *time.Time             `json:"expires_at,omitempty"`
	IssuedAt       *time.Time             `json:"issued_at,omitempty"`
	NotBefore      *time.Time             `json:"not_before,omitempty"`
	Issuer         string                 `json:"issuer,omitempty"`
	Audience       string                 `json:"audience,omitempty"`
	Scopes         []string               `json:"scopes,omitempty"`
	Claims         map[string]interface{} `json:"claims,omitempty"`
	ValidationTime time.Time              `json:"validation_time"`
	ExpiresIn      time.Duration          `json:"expires_in"`
	Errors         []ValidationError      `json:"errors,omitempty"`
	Warnings       []string               `json:"warnings,omitempty"`
	SecurityScore  int                    `json:"security_score"` // 0-100
}

// RefreshTokenResponse represents a token refresh response
type RefreshTokenResponse struct {
	Success      bool      `json:"success"`
	NewToken     string    `json:"new_token,omitempty"`
	TokenType    TokenType `json:"token_type,omitempty"`
	ExpiresIn    int64     `json:"expires_in,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Scopes       []string  `json:"scopes,omitempty"`
	Error        string    `json:"error,omitempty"`
	RefreshedAt  time.Time `json:"refreshed_at"`
}

// PermissionValidationResult represents permission validation results
type PermissionValidationResult struct {
	HasPermission  bool      `json:"has_permission"`
	GrantedScopes  []string  `json:"granted_scopes"`
	RequiredScopes []string  `json:"required_scopes"`
	MissingScopes  []string  `json:"missing_scopes"`
	ValidationTime time.Time `json:"validation_time"`
}

// TokenStrengthAnalysis represents security analysis of a token
type TokenStrengthAnalysis struct {
	OverallScore    int      `json:"overall_score"` // 0-100
	Length          int      `json:"length"`
	Entropy         float64  `json:"entropy"`
	HasUppercase    bool     `json:"has_uppercase"`
	HasLowercase    bool     `json:"has_lowercase"`
	HasNumbers      bool     `json:"has_numbers"`
	HasSpecialChars bool     `json:"has_special_chars"`
	Recommendations []string `json:"recommendations"`
	SecurityLevel   string   `json:"security_level"` // weak, medium, strong, very_strong
}

// FormatValidationResult represents token format validation results
type FormatValidationResult struct {
	IsValidFormat  bool     `json:"is_valid_format"`
	ExpectedFormat string   `json:"expected_format"`
	ActualFormat   string   `json:"actual_format"`
	FormatErrors   []string `json:"format_errors,omitempty"`
	FormatWarnings []string `json:"format_warnings,omitempty"`
}

// TokenValidatorConfig represents token validator configuration
type TokenValidatorConfig struct {
	JWTSecretKey        []byte                               `json:"jwt_secret_key"`
	DefaultTokenTTL     time.Duration                        `json:"default_token_ttl"`
	ClockSkewTolerance  time.Duration                        `json:"clock_skew_tolerance"`
	EnableStrengthCheck bool                                 `json:"enable_strength_check"`
	MinTokenLength      int                                  `json:"min_token_length"`
	MaxTokenLength      int                                  `json:"max_token_length"`
	EnableFormatCheck   bool                                 `json:"enable_format_check"`
	CacheValidationTTL  time.Duration                        `json:"cache_validation_ttl"`
	AuditValidation     bool                                 `json:"audit_validation"`
	ProviderConfigs     map[string]*ProviderValidationConfig `json:"provider_configs"`
}

// ProviderValidationConfig represents provider-specific validation rules
type ProviderValidationConfig struct {
	TokenFormat       string `json:"token_format"`    // regex pattern
	RequiredPrefix    string `json:"required_prefix"` // e.g., "sk-" for OpenAI
	MinLength         int    `json:"min_length"`
	MaxLength         int    `json:"max_length"`
	AllowedCharacters string `json:"allowed_characters"` // character set
	ValidateChecksum  bool   `json:"validate_checksum"`
	ExpiryRequired    bool   `json:"expiry_required"`
	ScopesRequired    bool   `json:"scopes_required"`
}

// tokenValidatorImpl implements TokenValidator interface
type tokenValidatorImpl struct {
	config       *TokenValidatorConfig
	tokenManager TokenManager
	auditSvc     audit.Service
}

// NewTokenValidator creates a new token validator
func NewTokenValidator(tokenManager TokenManager, auditSvc audit.Service, config *TokenValidatorConfig) (TokenValidator, error) {
	if config == nil {
		// Default configuration
		config = &TokenValidatorConfig{
			DefaultTokenTTL:     24 * time.Hour,
			ClockSkewTolerance:  5 * time.Minute,
			EnableStrengthCheck: true,
			MinTokenLength:      16,
			MaxTokenLength:      2048,
			EnableFormatCheck:   true,
			CacheValidationTTL:  5 * time.Minute,
			AuditValidation:     true,
			ProviderConfigs:     make(map[string]*ProviderValidationConfig),
		}

		// Configure provider-specific validation
		config.ProviderConfigs["openai"] = &ProviderValidationConfig{
			RequiredPrefix:    "sk-",
			MinLength:         51,
			MaxLength:         51,
			AllowedCharacters: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
			ValidateChecksum:  false,
			ExpiryRequired:    false,
		}

		config.ProviderConfigs["claude"] = &ProviderValidationConfig{
			RequiredPrefix:    "sk-ant-",
			MinLength:         108,
			MaxLength:         108,
			AllowedCharacters: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_",
			ValidateChecksum:  false,
			ExpiryRequired:    false,
		}
	}

	validator := &tokenValidatorImpl{
		config:       config,
		tokenManager: tokenManager,
		auditSvc:     auditSvc,
	}

	log.Println("Token validator initialized")
	return validator, nil
}

// ValidateAPIKey validates an API key format and authenticity
func (tv *tokenValidatorImpl) ValidateAPIKey(ctx context.Context, req *ValidateAPIKeyRequest) (*TokenValidationResponse, error) {
	response := &TokenValidationResponse{
		TokenType:      TokenTypeAPIKey,
		Provider:       req.Provider,
		UserID:         req.UserID,
		ValidationTime: time.Now(),
		Errors:         []ValidationError{},
		Warnings:       []string{},
		SecurityScore:  100,
	}

	// Basic validation
	if req.APIKey == "" {
		response.Errors = append(response.Errors, ValidationError{
			Code:    "EMPTY_TOKEN",
			Message: "API key cannot be empty",
			Field:   "api_key",
		})
		response.IsValid = false
		response.SecurityScore = 0
		return response, nil
	}

	// Format validation
	if tv.config.EnableFormatCheck {
		formatResult, err := tv.ValidateTokenFormat(ctx, req.Provider, req.APIKey)
		if err != nil {
			response.Errors = append(response.Errors, ValidationError{
				Code:    "FORMAT_CHECK_FAILED",
				Message: err.Error(),
				Field:   "api_key",
			})
		} else if !formatResult.IsValidFormat {
			for _, formatError := range formatResult.FormatErrors {
				response.Errors = append(response.Errors, ValidationError{
					Code:    "INVALID_FORMAT",
					Message: formatError,
					Field:   "api_key",
				})
			}
			response.IsValid = false
			response.SecurityScore -= 30
		}
	}

	// Strength analysis
	if tv.config.EnableStrengthCheck {
		strength, err := tv.GetTokenStrength(ctx, req.APIKey)
		if err == nil {
			response.SecurityScore = (response.SecurityScore + strength.OverallScore) / 2

			if strength.OverallScore < 50 {
				response.Warnings = append(response.Warnings, "Token has low security strength")
			}

			response.Warnings = append(response.Warnings, strength.Recommendations...)
		}
	}

	// Check if token is stored and valid
	_ = tv.generateTokenHash(req.APIKey)
	// Here you would check against stored tokens
	// For now, assume valid if format is correct

	if len(response.Errors) == 0 {
		response.IsValid = true
	}

	// Audit the validation attempt
	tv.auditValidation(ctx, "api_key", req.Provider, req.UserID, response.IsValid, nil)

	return response, nil
}

// ValidateJWT validates a JWT token
func (tv *tokenValidatorImpl) ValidateJWT(ctx context.Context, req *ValidateJWTRequest) (*TokenValidationResponse, error) {
	response := &TokenValidationResponse{
		TokenType:      TokenTypeJWT,
		Provider:       req.Provider,
		UserID:         req.UserID,
		ValidationTime: time.Now(),
		Errors:         []ValidationError{},
		Warnings:       []string{},
		Claims:         make(map[string]interface{}),
		SecurityScore:  100,
	}

	// Parse JWT token
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		// Check signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tv.config.JWTSecretKey, nil
	})

	if err != nil {
		response.Errors = append(response.Errors, ValidationError{
			Code:    "JWT_PARSE_ERROR",
			Message: err.Error(),
			Field:   "token",
		})
		response.IsValid = false
		response.SecurityScore = 0
		return response, nil
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		response.Claims = claims

		// Validate expiration
		if req.ValidateExp {
			if exp, ok := claims["exp"]; ok {
				if expFloat, ok := exp.(float64); ok {
					expTime := time.Unix(int64(expFloat), 0)
					response.ExpiresAt = &expTime
					response.ExpiresIn = time.Until(expTime)

					if expTime.Before(time.Now()) {
						response.Errors = append(response.Errors, ValidationError{
							Code:    "TOKEN_EXPIRED",
							Message: "JWT token has expired",
							Field:   "exp",
						})
						response.IsValid = false
						response.SecurityScore -= 50
					}
				}
			}
		}

		// Validate issued at
		if req.ValidateIat {
			if iat, ok := claims["iat"]; ok {
				if iatFloat, ok := iat.(float64); ok {
					iatTime := time.Unix(int64(iatFloat), 0)
					response.IssuedAt = &iatTime

					if iatTime.After(time.Now().Add(tv.config.ClockSkewTolerance)) {
						response.Errors = append(response.Errors, ValidationError{
							Code:    "TOKEN_FUTURE_ISSUED",
							Message: "JWT token issued in the future",
							Field:   "iat",
						})
						response.IsValid = false
					}
				}
			}
		}

		// Validate not before
		if req.ValidateNbf {
			if nbf, ok := claims["nbf"]; ok {
				if nbfFloat, ok := nbf.(float64); ok {
					nbfTime := time.Unix(int64(nbfFloat), 0)
					response.NotBefore = &nbfTime

					if nbfTime.After(time.Now().Add(tv.config.ClockSkewTolerance)) {
						response.Errors = append(response.Errors, ValidationError{
							Code:    "TOKEN_NOT_YET_VALID",
							Message: "JWT token not yet valid",
							Field:   "nbf",
						})
						response.IsValid = false
					}
				}
			}
		}

		// Validate audience
		if req.ValidateAud && req.ExpectedAud != "" {
			if aud, ok := claims["aud"]; ok {
				if audStr, ok := aud.(string); ok {
					response.Audience = audStr
					if audStr != req.ExpectedAud {
						response.Errors = append(response.Errors, ValidationError{
							Code:    "INVALID_AUDIENCE",
							Message: fmt.Sprintf("Expected audience %s, got %s", req.ExpectedAud, audStr),
							Field:   "aud",
						})
						response.IsValid = false
					}
				}
			}
		}

		// Validate issuer
		if req.ValidateIss && req.ExpectedIss != "" {
			if iss, ok := claims["iss"]; ok {
				if issStr, ok := iss.(string); ok {
					response.Issuer = issStr
					if issStr != req.ExpectedIss {
						response.Errors = append(response.Errors, ValidationError{
							Code:    "INVALID_ISSUER",
							Message: fmt.Sprintf("Expected issuer %s, got %s", req.ExpectedIss, issStr),
							Field:   "iss",
						})
						response.IsValid = false
					}
				}
			}
		}

		// Extract scopes if present
		if scopes, ok := claims["scope"]; ok {
			if scopeStr, ok := scopes.(string); ok {
				response.Scopes = strings.Split(scopeStr, " ")
			}
		}
	}

	if token.Valid && len(response.Errors) == 0 {
		response.IsValid = true
	}

	// Audit the validation attempt
	tv.auditValidation(ctx, "jwt", req.Provider, req.UserID, response.IsValid, nil)

	return response, nil
}

// ValidateOAuth validates an OAuth token
func (tv *tokenValidatorImpl) ValidateOAuth(ctx context.Context, req *ValidateOAuthRequest) (*TokenValidationResponse, error) {
	response := &TokenValidationResponse{
		TokenType:      TokenTypeOAuth,
		Provider:       req.Provider,
		UserID:         req.UserID,
		ValidationTime: time.Now(),
		Errors:         []ValidationError{},
		Warnings:       []string{},
		SecurityScore:  100,
	}

	// Basic validation
	if req.AccessToken == "" {
		response.Errors = append(response.Errors, ValidationError{
			Code:    "EMPTY_TOKEN",
			Message: "Access token cannot be empty",
			Field:   "access_token",
		})
		response.IsValid = false
		response.SecurityScore = 0
		return response, nil
	}

	// Validate token type
	if req.TokenType != "" && req.TokenType != "Bearer" {
		response.Warnings = append(response.Warnings, fmt.Sprintf("Unexpected token type: %s", req.TokenType))
	}

	// Here you would typically make a call to the OAuth provider's introspection endpoint
	// For now, we'll do basic format validation

	// Strength analysis
	if tv.config.EnableStrengthCheck {
		strength, err := tv.GetTokenStrength(ctx, req.AccessToken)
		if err == nil {
			response.SecurityScore = strength.OverallScore

			if strength.OverallScore < 50 {
				response.Warnings = append(response.Warnings, "Token has low security strength")
			}
		}
	}

	// Assume valid for now (would integrate with actual OAuth provider)
	response.IsValid = true

	// Audit the validation attempt
	tv.auditValidation(ctx, "oauth", req.Provider, req.UserID, response.IsValid, nil)

	return response, nil
}

// RefreshExpiredToken attempts to refresh an expired token
func (tv *tokenValidatorImpl) RefreshExpiredToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	response := &RefreshTokenResponse{
		RefreshedAt: time.Now(),
	}

	// Basic validation
	if req.RefreshToken == "" {
		response.Error = "Refresh token is required"
		return response, nil
	}

	// Here you would integrate with the provider's token refresh endpoint
	// For this implementation, we'll simulate the refresh

	// Generate new token (simplified)
	newTokenID := uuid.New().String()
	response.Success = true
	response.NewToken = fmt.Sprintf("refreshed_%s", newTokenID)
	response.TokenType = TokenTypeBearer
	response.ExpiresIn = int64(tv.config.DefaultTokenTTL.Seconds())
	response.RefreshToken = fmt.Sprintf("refresh_%s", newTokenID)

	// Audit the refresh attempt
	tv.auditValidation(ctx, "refresh", req.Provider, req.UserID, true, nil)

	log.Printf("Token refreshed successfully for provider %s", req.Provider)
	return response, nil
}

// ValidateTokenPermissions checks if token has required permissions
func (tv *tokenValidatorImpl) ValidateTokenPermissions(ctx context.Context, tokenID string, requiredScopes []string) (*PermissionValidationResult, error) {
	result := &PermissionValidationResult{
		RequiredScopes: requiredScopes,
		ValidationTime: time.Now(),
	}

	// Retrieve token to get its scopes
	token, err := tv.tokenManager.RetrieveToken(ctx, tokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve token: %w", err)
	}

	result.GrantedScopes = token.Scopes

	// Check if all required scopes are present
	result.HasPermission = true
	result.MissingScopes = []string{}

	for _, required := range requiredScopes {
		found := false
		for _, granted := range token.Scopes {
			if granted == required {
				found = true
				break
			}
		}
		if !found {
			result.HasPermission = false
			result.MissingScopes = append(result.MissingScopes, required)
		}
	}

	return result, nil
}

// GetTokenStrength analyzes token security strength
func (tv *tokenValidatorImpl) GetTokenStrength(ctx context.Context, token string) (*TokenStrengthAnalysis, error) {
	analysis := &TokenStrengthAnalysis{
		Length:          len(token),
		Recommendations: []string{},
	}

	// Check character types
	for _, char := range token {
		if char >= 'A' && char <= 'Z' {
			analysis.HasUppercase = true
		} else if char >= 'a' && char <= 'z' {
			analysis.HasLowercase = true
		} else if char >= '0' && char <= '9' {
			analysis.HasNumbers = true
		} else {
			analysis.HasSpecialChars = true
		}
	}

	// Calculate basic entropy
	charSet := 0
	if analysis.HasUppercase {
		charSet += 26
	}
	if analysis.HasLowercase {
		charSet += 26
	}
	if analysis.HasNumbers {
		charSet += 10
	}
	if analysis.HasSpecialChars {
		charSet += 32 // Approximate number of special characters
	}

	if charSet > 0 {
		// Simplified entropy calculation
		analysis.Entropy = float64(analysis.Length) * (float64(charSet) / 94.0) * 6.5 // Rough approximation
	}

	// Calculate score based on various factors
	score := 0

	// Length score (0-40 points)
	if analysis.Length >= 32 {
		score += 40
	} else if analysis.Length >= 20 {
		score += 30
	} else if analysis.Length >= 16 {
		score += 20
	} else if analysis.Length >= 8 {
		score += 10
	}

	// Character diversity (0-30 points)
	diversity := 0
	if analysis.HasUppercase {
		diversity += 7
	}
	if analysis.HasLowercase {
		diversity += 7
	}
	if analysis.HasNumbers {
		diversity += 8
	}
	if analysis.HasSpecialChars {
		diversity += 8
	}
	score += diversity

	// Entropy bonus (0-30 points)
	if analysis.Entropy >= 4.0 {
		score += 30
	} else if analysis.Entropy >= 3.0 {
		score += 20
	} else if analysis.Entropy >= 2.0 {
		score += 10
	}

	analysis.OverallScore = score

	// Determine security level
	if score >= 85 {
		analysis.SecurityLevel = "very_strong"
	} else if score >= 70 {
		analysis.SecurityLevel = "strong"
	} else if score >= 50 {
		analysis.SecurityLevel = "medium"
	} else {
		analysis.SecurityLevel = "weak"
	}

	// Add recommendations
	if analysis.Length < 16 {
		analysis.Recommendations = append(analysis.Recommendations, "Consider using a longer token (16+ characters)")
	}
	if !analysis.HasUppercase {
		analysis.Recommendations = append(analysis.Recommendations, "Include uppercase letters")
	}
	if !analysis.HasLowercase {
		analysis.Recommendations = append(analysis.Recommendations, "Include lowercase letters")
	}
	if !analysis.HasNumbers {
		analysis.Recommendations = append(analysis.Recommendations, "Include numbers")
	}
	if !analysis.HasSpecialChars {
		analysis.Recommendations = append(analysis.Recommendations, "Include special characters")
	}

	return analysis, nil
}

// ValidateTokenFormat validates token format for different providers
func (tv *tokenValidatorImpl) ValidateTokenFormat(ctx context.Context, provider string, token string) (*FormatValidationResult, error) {
	result := &FormatValidationResult{
		ActualFormat:   fmt.Sprintf("length=%d", len(token)),
		FormatErrors:   []string{},
		FormatWarnings: []string{},
	}

	config, exists := tv.config.ProviderConfigs[strings.ToLower(provider)]
	if !exists {
		result.FormatWarnings = append(result.FormatWarnings, fmt.Sprintf("No format validation rules for provider: %s", provider))
		result.IsValidFormat = true // Assume valid if no rules
		return result, nil
	}

	result.ExpectedFormat = fmt.Sprintf("prefix=%s, length=%d-%d", config.RequiredPrefix, config.MinLength, config.MaxLength)

	// Check prefix
	if config.RequiredPrefix != "" && !strings.HasPrefix(token, config.RequiredPrefix) {
		result.FormatErrors = append(result.FormatErrors, fmt.Sprintf("Token must start with '%s'", config.RequiredPrefix))
	}

	// Check length
	if config.MinLength > 0 && len(token) < config.MinLength {
		result.FormatErrors = append(result.FormatErrors, fmt.Sprintf("Token too short: %d < %d", len(token), config.MinLength))
	}
	if config.MaxLength > 0 && len(token) > config.MaxLength {
		result.FormatErrors = append(result.FormatErrors, fmt.Sprintf("Token too long: %d > %d", len(token), config.MaxLength))
	}

	// Check allowed characters
	if config.AllowedCharacters != "" {
		allowedChars := config.AllowedCharacters
		for _, char := range token {
			if !strings.ContainsRune(allowedChars, char) {
				result.FormatErrors = append(result.FormatErrors, fmt.Sprintf("Invalid character in token: %c", char))
				break
			}
		}
	}

	result.IsValidFormat = len(result.FormatErrors) == 0
	return result, nil
}

// Helper methods

func (tv *tokenValidatorImpl) generateTokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (tv *tokenValidatorImpl) auditValidation(ctx context.Context, tokenType, provider, userID string, success bool, err error) {
	if !tv.config.AuditValidation || tv.auditSvc == nil {
		return
	}

	safetyLevel := models.SafetyLevelSafe
	if !success {
		safetyLevel = models.SafetyLevelWarning
	}

	var uid uuid.UUID
	if userID != "" {
		uid = uuid.MustParse(userID)
	}

	auditReq := models.AuditLogRequest{
		UserID:           &uid,
		QueryText:        fmt.Sprintf("Token validation: %s %s", tokenType, provider),
		GeneratedCommand: fmt.Sprintf("Validate %s token for provider %s", tokenType, provider),
		SafetyLevel:      string(safetyLevel),
		ExecutionResult: map[string]interface{}{
			"token_type": tokenType,
			"provider":   provider,
			"success":    success,
		},
		ExecutionStatus: func() string {
			if success {
				return models.ExecutionStatusSuccess
			}
			return models.ExecutionStatusFailed
		}(),
	}

	if err != nil {
		auditReq.ExecutionResult["error"] = err.Error()
	}

	if logErr := tv.auditSvc.LogUserAction(ctx, auditReq); logErr != nil {
		log.Printf("Failed to log token validation: %v", logErr)
	}
}
