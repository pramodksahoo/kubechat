package security

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityValidationConfig holds security validation configuration
type SecurityValidationConfig struct {
	RequiredEnvVars     []string
	MinPasswordLength   int
	TokenExpiryMinutes  int
	RequireHTTPS        bool
	AllowedOrigins      []string
	MaxRequestSizeBytes int64
}

// SecurityValidator handles startup security validation
type SecurityValidator struct {
	config *SecurityValidationConfig
}

// NewSecurityValidator creates a new security validator
func NewSecurityValidator() *SecurityValidator {
	return &SecurityValidator{
		config: &SecurityValidationConfig{
			RequiredEnvVars: []string{
				"DB_HOST",
				"DB_NAME",
				"DB_USER",
				"DB_PASSWORD",
				"JWT_SECRET",
				"REDIS_HOST",
			},
			MinPasswordLength:   8,
			TokenExpiryMinutes:  1440, // 24 hours
			RequireHTTPS:        false, // Allow HTTP in development
			AllowedOrigins:      []string{"http://localhost:3000", "http://localhost:30001"},
			MaxRequestSizeBytes: 10 * 1024 * 1024, // 10MB
		},
	}
}

// ValidateSecurityConfiguration performs comprehensive security validation at startup
func (sv *SecurityValidator) ValidateSecurityConfiguration() error {
	validationErrors := []string{}

	// Validate required environment variables
	if err := sv.validateEnvironmentVariables(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate JWT configuration
	if err := sv.validateJWTConfiguration(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate database configuration
	if err := sv.validateDatabaseConfiguration(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate CORS configuration
	if err := sv.validateCORSConfiguration(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate security headers configuration
	if err := sv.validateSecurityHeaders(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("security validation failed:\n%s", strings.Join(validationErrors, "\n"))
	}

	return nil
}

// validateEnvironmentVariables checks for required environment variables
func (sv *SecurityValidator) validateEnvironmentVariables() error {
	missing := []string{}

	for _, envVar := range sv.config.RequiredEnvVars {
		if value := os.Getenv(envVar); value == "" {
			missing = append(missing, envVar)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// validateJWTConfiguration validates JWT security settings
func (sv *SecurityValidator) validateJWTConfiguration() error {
	jwtSecret := os.Getenv("JWT_SECRET")

	// Check JWT secret strength
	if len(jwtSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long for security")
	}

	// Check for weak secrets
	weakSecrets := []string{"secret", "password", "jwt", "token", "key"}
	for _, weak := range weakSecrets {
		if strings.Contains(strings.ToLower(jwtSecret), weak) {
			return fmt.Errorf("JWT_SECRET appears to contain weak patterns - use a strong random secret")
		}
	}

	// Validate token expiry configuration
	if expiryStr := os.Getenv("JWT_EXPIRY_MINUTES"); expiryStr != "" {
		if expiry, err := strconv.Atoi(expiryStr); err != nil {
			return fmt.Errorf("JWT_EXPIRY_MINUTES must be a valid integer")
		} else if expiry < 1 || expiry > 43200 { // Max 30 days
			return fmt.Errorf("JWT_EXPIRY_MINUTES must be between 1 and 43200 (30 days)")
		}
	}

	return nil
}

// validateDatabaseConfiguration validates database security settings
func (sv *SecurityValidator) validateDatabaseConfiguration() error {
	dbPassword := os.Getenv("DB_PASSWORD")

	// Check database password strength
	if len(dbPassword) < sv.config.MinPasswordLength {
		return fmt.Errorf("DB_PASSWORD must be at least %d characters long", sv.config.MinPasswordLength)
	}

	// Validate SSL mode if specified
	if sslMode := os.Getenv("DB_SSLMODE"); sslMode != "" {
		validSSLModes := []string{"disable", "require", "verify-ca", "verify-full"}
		isValid := false
		for _, mode := range validSSLModes {
			if sslMode == mode {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("DB_SSLMODE must be one of: %s", strings.Join(validSSLModes, ", "))
		}
	}

	return nil
}

// validateCORSConfiguration validates CORS settings
func (sv *SecurityValidator) validateCORSConfiguration() error {
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")

	// If CORS origins are specified, validate them
	if allowedOrigins != "" {
		origins := strings.Split(allowedOrigins, ",")
		for _, origin := range origins {
			origin = strings.TrimSpace(origin)
			if origin == "*" {
				return fmt.Errorf("CORS wildcard (*) origin is not allowed in production for security")
			}
			if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
				return fmt.Errorf("CORS origin must include protocol (http:// or https://): %s", origin)
			}
		}
	}

	return nil
}

// validateSecurityHeaders validates security header configuration
func (sv *SecurityValidator) validateSecurityHeaders() error {
	// Check if running in production mode
	if gin.Mode() == gin.ReleaseMode {
		// In production, ensure HTTPS is enforced
		if !sv.config.RequireHTTPS {
			// Check if HTTPS enforcement is disabled via environment
			if httpsRequired := os.Getenv("REQUIRE_HTTPS"); httpsRequired != "false" {
				sv.config.RequireHTTPS = true
			}
		}
	}

	// Validate Content Security Policy if specified
	if csp := os.Getenv("CONTENT_SECURITY_POLICY"); csp != "" {
		if !strings.Contains(csp, "default-src") {
			return fmt.Errorf("Content Security Policy must include default-src directive")
		}
	}

	return nil
}

// PerformSecurityHealthCheck runs security-focused health checks
func (sv *SecurityValidator) PerformSecurityHealthCheck() map[string]interface{} {
	checks := make(map[string]interface{})

	// Check environment validation
	checks["environment_variables"] = sv.checkEnvironmentHealth()

	// Check JWT configuration
	checks["jwt_configuration"] = sv.checkJWTHealth()

	// Check database connection security
	checks["database_security"] = sv.checkDatabaseSecurity()

	// Check security headers
	checks["security_headers"] = sv.checkSecurityHeaders()

	// Check rate limiting
	checks["rate_limiting"] = sv.checkRateLimiting()

	return checks
}

// checkEnvironmentHealth validates environment configuration health
func (sv *SecurityValidator) checkEnvironmentHealth() map[string]interface{} {
	result := map[string]interface{}{
		"status": "healthy",
		"details": make(map[string]string),
	}

	for _, envVar := range sv.config.RequiredEnvVars {
		if value := os.Getenv(envVar); value == "" {
			result["status"] = "unhealthy"
			result["details"].(map[string]string)[envVar] = "missing"
		} else {
			result["details"].(map[string]string)[envVar] = "present"
		}
	}

	return result
}

// checkJWTHealth validates JWT configuration health
func (sv *SecurityValidator) checkJWTHealth() map[string]interface{} {
	result := map[string]interface{}{
		"status": "healthy",
		"details": make(map[string]interface{}),
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		result["status"] = "unhealthy"
		result["details"].(map[string]interface{})["secret_strength"] = "weak"
	} else {
		result["details"].(map[string]interface{})["secret_strength"] = "strong"
	}

	result["details"].(map[string]interface{})["expiry_configured"] = os.Getenv("JWT_EXPIRY_MINUTES") != ""

	return result
}

// checkDatabaseSecurity validates database security configuration
func (sv *SecurityValidator) checkDatabaseSecurity() map[string]interface{} {
	result := map[string]interface{}{
		"status": "healthy",
		"details": make(map[string]interface{}),
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if len(dbPassword) < sv.config.MinPasswordLength {
		result["status"] = "unhealthy"
		result["details"].(map[string]interface{})["password_strength"] = "weak"
	} else {
		result["details"].(map[string]interface{})["password_strength"] = "strong"
	}

	sslMode := os.Getenv("DB_SSLMODE")
	result["details"].(map[string]interface{})["ssl_mode"] = sslMode
	if sslMode == "disable" {
		result["details"].(map[string]interface{})["ssl_warning"] = "SSL disabled - not recommended for production"
	}

	return result
}

// checkSecurityHeaders validates security headers configuration
func (sv *SecurityValidator) checkSecurityHeaders() map[string]interface{} {
	result := map[string]interface{}{
		"status": "healthy",
		"details": make(map[string]interface{}),
	}

	result["details"].(map[string]interface{})["gin_mode"] = gin.Mode()
	result["details"].(map[string]interface{})["https_required"] = sv.config.RequireHTTPS
	result["details"].(map[string]interface{})["csp_configured"] = os.Getenv("CONTENT_SECURITY_POLICY") != ""

	return result
}

// checkRateLimiting validates rate limiting configuration
func (sv *SecurityValidator) checkRateLimiting() map[string]interface{} {
	result := map[string]interface{}{
		"status": "healthy",
		"details": make(map[string]interface{}),
	}

	rateLimitEnabled := os.Getenv("RATE_LIMIT_ENABLED")
	result["details"].(map[string]interface{})["enabled"] = rateLimitEnabled == "true"

	if rateLimitRPS := os.Getenv("RATE_LIMIT_RPS"); rateLimitRPS != "" {
		if rps, err := strconv.Atoi(rateLimitRPS); err == nil {
			result["details"].(map[string]interface{})["requests_per_second"] = rps
		}
	}

	return result
}

// GetSecurityConfiguration returns the current security configuration
func (sv *SecurityValidator) GetSecurityConfiguration() *SecurityValidationConfig {
	return sv.config
}

// SetRequireHTTPS sets HTTPS requirement (for testing)
func (sv *SecurityValidator) SetRequireHTTPS(require bool) {
	sv.config.RequireHTTPS = require
}