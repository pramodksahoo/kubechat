package external

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/audit"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/credentials"
)

// CredentialValidator provides comprehensive credential validation and format checking
type CredentialValidator interface {
	// ValidateCredential validates a credential comprehensively
	ValidateCredential(ctx context.Context, req *ValidationRequest) (*ValidationResponse, error)

	// ValidateFormat validates credential format without testing connectivity
	ValidateFormat(ctx context.Context, cred *credentials.Credential) (*FormatValidation, error)

	// ValidateStrength analyzes credential strength and security
	ValidateStrength(ctx context.Context, cred *credentials.Credential) (*StrengthAnalysis, error)

	// ValidateConnectivity tests credential connectivity with target service
	ValidateConnectivity(ctx context.Context, cred *credentials.Credential, endpoint string) (*ConnectivityTest, error)

	// ValidateExpiration checks credential expiration and lifetime
	ValidateExpiration(ctx context.Context, cred *credentials.Credential) (*ExpirationCheck, error)

	// ValidateCompliance checks credential compliance with policies
	ValidateCompliance(ctx context.Context, cred *credentials.Credential, policies []string) (*ComplianceValidation, error)

	// BatchValidate validates multiple credentials efficiently
	BatchValidate(ctx context.Context, creds []*credentials.Credential) (*BatchValidationResult, error)

	// GetValidationRules returns available validation rules
	GetValidationRules(ctx context.Context) ([]*ValidationRule, error)

	// ConfigureValidationRules configures custom validation rules
	ConfigureValidationRules(ctx context.Context, rules []*ValidationRule) error
}

// ValidationRequest contains credential validation parameters
type ValidationRequest struct {
	Credential     *credentials.Credential `json:"credential"`
	ValidationType ValidationType          `json:"validation_type"`
	Options        *ValidationOptions      `json:"options,omitempty"`
	Policies       []string                `json:"policies,omitempty"`
	TestEndpoint   string                  `json:"test_endpoint,omitempty"`
	CustomRules    []*ValidationRule       `json:"custom_rules,omitempty"`
}

// ValidationResponse contains comprehensive validation results
type ValidationResponse struct {
	ValidationID     uuid.UUID             `json:"validation_id"`
	CredentialName   string                `json:"credential_name"`
	IsValid          bool                  `json:"is_valid"`
	OverallScore     int                   `json:"overall_score"`
	ValidationTime   time.Duration         `json:"validation_time"`
	ValidatedAt      time.Time             `json:"validated_at"`
	FormatCheck      *FormatValidation     `json:"format_check"`
	StrengthAnalysis *StrengthAnalysis     `json:"strength_analysis,omitempty"`
	ConnectivityTest *ConnectivityTest     `json:"connectivity_test,omitempty"`
	ExpirationCheck  *ExpirationCheck      `json:"expiration_check,omitempty"`
	ComplianceCheck  *ComplianceValidation `json:"compliance_check,omitempty"`
	Recommendations  []string              `json:"recommendations"`
	Errors           []ValidationError     `json:"errors"`
	Warnings         []ValidationWarning   `json:"warnings"`
}

// FormatValidation contains credential format validation results
type FormatValidation struct {
	IsValidFormat   bool                       `json:"is_valid_format"`
	DetectedType    credentials.CredentialType `json:"detected_type"`
	FormatScore     int                        `json:"format_score"`
	FormatErrors    []string                   `json:"format_errors"`
	FormatWarnings  []string                   `json:"format_warnings"`
	FormatDetails   *FormatDetails             `json:"format_details,omitempty"`
	ValidatedFields []ValidatedField           `json:"validated_fields"`
}

// StrengthAnalysis contains credential strength analysis
type StrengthAnalysis struct {
	StrengthScore   int                   `json:"strength_score"`
	StrengthLevel   StrengthLevel         `json:"strength_level"`
	Entropy         float64               `json:"entropy"`
	Length          int                   `json:"length"`
	Composition     *CharacterComposition `json:"composition"`
	Patterns        []SecurityPattern     `json:"patterns"`
	Vulnerabilities []string              `json:"vulnerabilities"`
	Recommendations []string              `json:"recommendations"`
}

// ConnectivityTest contains connectivity test results
type ConnectivityTest struct {
	Success         bool                   `json:"success"`
	ResponseTime    time.Duration          `json:"response_time"`
	StatusCode      int                    `json:"status_code,omitempty"`
	ResponseSize    int64                  `json:"response_size"`
	TestEndpoint    string                 `json:"test_endpoint"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	TestedAt        time.Time              `json:"tested_at"`
	CertificateInfo *CertificateInfo       `json:"certificate_info,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ExpirationCheck contains expiration validation results
type ExpirationCheck struct {
	HasExpiration   bool           `json:"has_expiration"`
	ExpiresAt       *time.Time     `json:"expires_at,omitempty"`
	IsExpired       bool           `json:"is_expired"`
	ExpiresIn       *time.Duration `json:"expires_in,omitempty"`
	ExpirationScore int            `json:"expiration_score"`
	Warnings        []string       `json:"warnings"`
	Recommendations []string       `json:"recommendations"`
}

// ComplianceValidation contains compliance validation results
type ComplianceValidation struct {
	IsCompliant     bool                     `json:"is_compliant"`
	ComplianceScore int                      `json:"compliance_score"`
	PolicyResults   []PolicyValidationResult `json:"policy_results"`
	Violations      []ComplianceViolation    `json:"violations"`
	RequiredActions []string                 `json:"required_actions"`
}

// BatchValidationResult contains results for batch validation
type BatchValidationResult struct {
	TotalCredentials   int                   `json:"total_credentials"`
	ValidCredentials   int                   `json:"valid_credentials"`
	InvalidCredentials int                   `json:"invalid_credentials"`
	ValidationTime     time.Duration         `json:"validation_time"`
	Results            []*ValidationResponse `json:"results"`
	Summary            *ValidationSummary    `json:"summary"`
	ValidatedAt        time.Time             `json:"validated_at"`
}

// ValidationRule defines custom validation rules
type ValidationRule struct {
	ID             uuid.UUID                  `json:"id"`
	Name           string                     `json:"name"`
	Description    string                     `json:"description"`
	CredentialType credentials.CredentialType `json:"credential_type"`
	Category       ValidationCategory         `json:"category"`
	Severity       ValidationSeverity         `json:"severity"`
	Enabled        bool                       `json:"enabled"`
	Conditions     []RuleCondition            `json:"conditions"`
	Actions        []RuleAction               `json:"actions"`
	Score          int                        `json:"score"`
	CustomLogic    string                     `json:"custom_logic,omitempty"`
	CreatedAt      time.Time                  `json:"created_at"`
	UpdatedAt      time.Time                  `json:"updated_at"`
}

// Supporting structs
type ValidationOptions struct {
	SkipConnectivityTest bool              `json:"skip_connectivity_test,omitempty"`
	SkipStrengthAnalysis bool              `json:"skip_strength_analysis,omitempty"`
	SkipComplianceCheck  bool              `json:"skip_compliance_check,omitempty"`
	Timeout              time.Duration     `json:"timeout,omitempty"`
	RetryAttempts        int               `json:"retry_attempts,omitempty"`
	CustomHeaders        map[string]string `json:"custom_headers,omitempty"`
	EnableDebug          bool              `json:"enable_debug,omitempty"`
}

type FormatDetails struct {
	Algorithm    string                 `json:"algorithm,omitempty"`
	KeyLength    int                    `json:"key_length,omitempty"`
	Version      string                 `json:"version,omitempty"`
	Issuer       string                 `json:"issuer,omitempty"`
	Subject      string                 `json:"subject,omitempty"`
	Fingerprint  string                 `json:"fingerprint,omitempty"`
	SerialNumber string                 `json:"serial_number,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

type ValidatedField struct {
	Name     string      `json:"name"`
	Value    interface{} `json:"value,omitempty"`
	IsValid  bool        `json:"is_valid"`
	Score    int         `json:"score"`
	Errors   []string    `json:"errors,omitempty"`
	Warnings []string    `json:"warnings,omitempty"`
}

type CharacterComposition struct {
	Uppercase  int     `json:"uppercase"`
	Lowercase  int     `json:"lowercase"`
	Digits     int     `json:"digits"`
	Special    int     `json:"special"`
	Whitespace int     `json:"whitespace"`
	Diversity  float64 `json:"diversity"`
}

type SecurityPattern struct {
	Type        string `json:"type"`
	Pattern     string `json:"pattern"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Count       int    `json:"count"`
}

type CertificateInfo struct {
	Subject      string    `json:"subject"`
	Issuer       string    `json:"issuer"`
	SerialNumber string    `json:"serial_number"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	KeyAlgorithm string    `json:"key_algorithm"`
	Fingerprint  string    `json:"fingerprint"`
	IsCA         bool      `json:"is_ca"`
	KeyUsage     []string  `json:"key_usage"`
}

type PolicyValidationResult struct {
	PolicyName  string   `json:"policy_name"`
	IsCompliant bool     `json:"is_compliant"`
	Score       int      `json:"score"`
	Violations  []string `json:"violations"`
	Passed      []string `json:"passed"`
}

type ComplianceViolation struct {
	ViolationType  string                 `json:"violation_type"`
	Severity       string                 `json:"severity"`
	Description    string                 `json:"description"`
	Recommendation string                 `json:"recommendation"`
	PolicyName     string                 `json:"policy_name"`
	Details        map[string]interface{} `json:"details,omitempty"`
}

type ValidationSummary struct {
	AverageScore         float64               `json:"average_score"`
	StrengthDistribution map[StrengthLevel]int `json:"strength_distribution"`
	CommonIssues         []string              `json:"common_issues"`
	ComplianceRate       float64               `json:"compliance_rate"`
	RecommendedActions   []string              `json:"recommended_actions"`
}

type RuleCondition struct {
	Field         string      `json:"field"`
	Operator      string      `json:"operator"`
	Value         interface{} `json:"value"`
	CaseSensitive bool        `json:"case_sensitive,omitempty"`
}

type RuleAction struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

type ValidationWarning struct {
	Code           string `json:"code"`
	Message        string `json:"message"`
	Field          string `json:"field,omitempty"`
	Recommendation string `json:"recommendation,omitempty"`
}

// Enums
type ValidationType string

const (
	ValidationTypeFormat       ValidationType = "format"
	ValidationTypeStrength     ValidationType = "strength"
	ValidationTypeConnectivity ValidationType = "connectivity"
	ValidationTypeExpiration   ValidationType = "expiration"
	ValidationTypeCompliance   ValidationType = "compliance"
	ValidationTypeFull         ValidationType = "full"
)

type StrengthLevel string

const (
	StrengthLevelWeak      StrengthLevel = "weak"
	StrengthLevelFair      StrengthLevel = "fair"
	StrengthLevelGood      StrengthLevel = "good"
	StrengthLevelStrong    StrengthLevel = "strong"
	StrengthLevelExcellent StrengthLevel = "excellent"
)

type ValidationCategory string

const (
	ValidationCategoryFormat     ValidationCategory = "format"
	ValidationCategorySecurity   ValidationCategory = "security"
	ValidationCategoryCompliance ValidationCategory = "compliance"
	ValidationCategoryCustom     ValidationCategory = "custom"
)

type ValidationSeverity string

const (
	ValidationSeverityInfo     ValidationSeverity = "info"
	ValidationSeverityWarning  ValidationSeverity = "warning"
	ValidationSeverityError    ValidationSeverity = "error"
	ValidationSeverityCritical ValidationSeverity = "critical"
)

// CredentialValidatorConfig contains configuration for the credential validator
type CredentialValidatorConfig struct {
	EnableConnectivityTests bool          `json:"enable_connectivity_tests"`
	DefaultTimeout          time.Duration `json:"default_timeout"`
	MaxRetryAttempts        int           `json:"max_retry_attempts"`
	EnableCaching           bool          `json:"enable_caching"`
	CacheDuration           time.Duration `json:"cache_duration"`
	EnableAuditLog          bool          `json:"enable_audit_log"`
	StrictValidation        bool          `json:"strict_validation"`
	EnableCustomRules       bool          `json:"enable_custom_rules"`
}

// credentialValidatorImpl implements CredentialValidator
type credentialValidatorImpl struct {
	auditSvc        audit.Service
	config          *CredentialValidatorConfig
	validationRules []*ValidationRule
	httpClient      *http.Client
}

// NewCredentialValidator creates a new credential validator
func NewCredentialValidator(auditSvc audit.Service, config *CredentialValidatorConfig) (CredentialValidator, error) {
	if config == nil {
		config = &CredentialValidatorConfig{
			EnableConnectivityTests: true,
			DefaultTimeout:          30 * time.Second,
			MaxRetryAttempts:        3,
			EnableCaching:           true,
			CacheDuration:           5 * time.Minute,
			EnableAuditLog:          true,
			StrictValidation:        false,
			EnableCustomRules:       true,
		}
	}

	validator := &credentialValidatorImpl{
		auditSvc:        auditSvc,
		config:          config,
		validationRules: make([]*ValidationRule, 0),
		httpClient: &http.Client{
			Timeout: config.DefaultTimeout,
		},
	}

	// Initialize built-in validation rules
	if err := validator.initializeBuiltInRules(); err != nil {
		return nil, fmt.Errorf("failed to initialize validation rules: %w", err)
	}

	log.Printf("Credential validator initialized with connectivity tests: %v", config.EnableConnectivityTests)
	return validator, nil
}

// ValidateCredential performs comprehensive credential validation
func (v *credentialValidatorImpl) ValidateCredential(ctx context.Context, req *ValidationRequest) (*ValidationResponse, error) {
	startTime := time.Now()
	validationID := uuid.New()

	response := &ValidationResponse{
		ValidationID:    validationID,
		CredentialName:  req.Credential.Name,
		ValidatedAt:     startTime,
		Recommendations: []string{},
		Errors:          []ValidationError{},
		Warnings:        []ValidationWarning{},
	}

	// Format validation (always performed)
	formatResult, err := v.ValidateFormat(ctx, req.Credential)
	if err != nil {
		return nil, fmt.Errorf("format validation failed: %w", err)
	}
	response.FormatCheck = formatResult

	// Perform additional validations based on type
	switch req.ValidationType {
	case ValidationTypeFull:
		// Perform all validations
		if !req.Options.SkipStrengthAnalysis {
			response.StrengthAnalysis, _ = v.ValidateStrength(ctx, req.Credential)
		}

		if !req.Options.SkipConnectivityTest && req.TestEndpoint != "" {
			response.ConnectivityTest, _ = v.ValidateConnectivity(ctx, req.Credential, req.TestEndpoint)
		}

		response.ExpirationCheck, _ = v.ValidateExpiration(ctx, req.Credential)

		if !req.Options.SkipComplianceCheck && len(req.Policies) > 0 {
			response.ComplianceCheck, _ = v.ValidateCompliance(ctx, req.Credential, req.Policies)
		}

	case ValidationTypeStrength:
		response.StrengthAnalysis, _ = v.ValidateStrength(ctx, req.Credential)
	case ValidationTypeConnectivity:
		if req.TestEndpoint != "" {
			response.ConnectivityTest, _ = v.ValidateConnectivity(ctx, req.Credential, req.TestEndpoint)
		}
	case ValidationTypeExpiration:
		response.ExpirationCheck, _ = v.ValidateExpiration(ctx, req.Credential)
	case ValidationTypeCompliance:
		if len(req.Policies) > 0 {
			response.ComplianceCheck, _ = v.ValidateCompliance(ctx, req.Credential, req.Policies)
		}
	}

	// Calculate overall score and validity
	response.OverallScore = v.calculateOverallScore(response)
	response.IsValid = response.OverallScore >= 70 // Configurable threshold

	// Generate recommendations
	response.Recommendations = v.generateRecommendations(response)

	response.ValidationTime = time.Since(startTime)

	// Log validation event
	v.logValidationEvent("validate_credential", req.Credential.Name, response.IsValid, nil)

	return response, nil
}

// ValidateFormat validates credential format without connectivity tests
func (v *credentialValidatorImpl) ValidateFormat(ctx context.Context, cred *credentials.Credential) (*FormatValidation, error) {
	result := &FormatValidation{
		FormatErrors:    []string{},
		FormatWarnings:  []string{},
		ValidatedFields: []ValidatedField{},
		FormatScore:     100,
	}

	// Detect credential type if not set
	detectedType := v.detectCredentialType(cred.Value)
	result.DetectedType = detectedType

	// Type-specific format validation
	switch cred.Type {
	case credentials.CredentialTypeAPIKey:
		v.validateAPIKeyFormat(cred.Value, result)
	case credentials.CredentialTypeToken:
		v.validateTokenFormat(cred.Value, result)
	case credentials.CredentialTypeCertificate:
		v.validateCertificateFormat(cred.Value, result)
	case credentials.CredentialTypeOAuth:
		v.validateOAuthFormat(cred.Value, result)
	case credentials.CredentialTypePassword:
		v.validatePasswordFormat(cred.Value, result)
	default:
		v.validateGenericFormat(cred.Value, result)
	}

	result.IsValidFormat = len(result.FormatErrors) == 0
	if len(result.FormatWarnings) > 0 {
		result.FormatScore -= len(result.FormatWarnings) * 5
	}
	if len(result.FormatErrors) > 0 {
		result.FormatScore -= len(result.FormatErrors) * 20
	}

	return result, nil
}

// ValidateStrength analyzes credential strength
func (v *credentialValidatorImpl) ValidateStrength(ctx context.Context, cred *credentials.Credential) (*StrengthAnalysis, error) {
	value := cred.Value

	analysis := &StrengthAnalysis{
		Length:          len(value),
		Vulnerabilities: []string{},
		Recommendations: []string{},
		Patterns:        []SecurityPattern{},
	}

	// Calculate character composition
	analysis.Composition = v.analyzeCharacterComposition(value)

	// Calculate entropy
	analysis.Entropy = v.calculateEntropy(value)

	// Detect security patterns
	analysis.Patterns = v.detectSecurityPatterns(value)

	// Calculate strength score
	analysis.StrengthScore = v.calculateStrengthScore(analysis)
	analysis.StrengthLevel = v.determineStrengthLevel(analysis.StrengthScore)

	// Generate recommendations
	if analysis.Length < 12 {
		analysis.Recommendations = append(analysis.Recommendations, "Consider using a longer credential")
	}
	if analysis.Composition.Diversity < 0.5 {
		analysis.Recommendations = append(analysis.Recommendations, "Increase character diversity")
	}
	if analysis.Entropy < 3.0 {
		analysis.Recommendations = append(analysis.Recommendations, "Use more random characters")
	}

	return analysis, nil
}

// ValidateConnectivity tests credential connectivity
func (v *credentialValidatorImpl) ValidateConnectivity(ctx context.Context, cred *credentials.Credential, endpoint string) (*ConnectivityTest, error) {
	startTime := time.Now()

	test := &ConnectivityTest{
		TestEndpoint: endpoint,
		TestedAt:     startTime,
		Metadata:     make(map[string]interface{}),
	}

	// Parse endpoint URL
	testURL, err := url.Parse(endpoint)
	if err != nil {
		test.Success = false
		test.ErrorMessage = fmt.Sprintf("Invalid endpoint URL: %v", err)
		return test, nil
	}

	// Create HTTP request with credential
	req, err := http.NewRequestWithContext(ctx, "GET", testURL.String(), nil)
	if err != nil {
		test.Success = false
		test.ErrorMessage = fmt.Sprintf("Failed to create request: %v", err)
		return test, nil
	}

	// Apply credential based on type
	switch cred.Type {
	case credentials.CredentialTypeAPIKey:
		req.Header.Set("Authorization", "Bearer "+cred.Value)
	case credentials.CredentialTypeToken:
		if strings.HasPrefix(cred.Value, "Bearer ") {
			req.Header.Set("Authorization", cred.Value)
		} else {
			req.Header.Set("Authorization", "Bearer "+cred.Value)
		}
	default:
		req.Header.Set("Authorization", "Bearer "+cred.Value)
	}

	// Make request
	resp, err := v.httpClient.Do(req)
	if err != nil {
		test.Success = false
		test.ErrorMessage = fmt.Sprintf("Request failed: %v", err)
		test.ResponseTime = time.Since(startTime)
		return test, nil
	}
	defer resp.Body.Close()

	test.ResponseTime = time.Since(startTime)
	test.StatusCode = resp.StatusCode
	test.ResponseSize = resp.ContentLength

	// Determine success based on status code
	test.Success = resp.StatusCode >= 200 && resp.StatusCode < 400

	if !test.Success {
		test.ErrorMessage = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return test, nil
}

// ValidateExpiration checks credential expiration
func (v *credentialValidatorImpl) ValidateExpiration(ctx context.Context, cred *credentials.Credential) (*ExpirationCheck, error) {
	check := &ExpirationCheck{
		Warnings:        []string{},
		Recommendations: []string{},
		ExpirationScore: 100,
	}

	// Check if credential has explicit expiration
	if cred.ExpiresAt != nil {
		check.HasExpiration = true
		check.ExpiresAt = cred.ExpiresAt
		check.IsExpired = cred.ExpiresAt.Before(time.Now())

		if !check.IsExpired {
			expiresIn := time.Until(*cred.ExpiresAt)
			check.ExpiresIn = &expiresIn

			// Generate warnings based on time to expiration
			if expiresIn < 24*time.Hour {
				check.Warnings = append(check.Warnings, "Credential expires within 24 hours")
				check.ExpirationScore = 20
			} else if expiresIn < 7*24*time.Hour {
				check.Warnings = append(check.Warnings, "Credential expires within 7 days")
				check.ExpirationScore = 50
			} else if expiresIn < 30*24*time.Hour {
				check.Warnings = append(check.Warnings, "Credential expires within 30 days")
				check.ExpirationScore = 80
			}
		} else {
			check.ExpirationScore = 0
		}
	} else {
		// Try to detect expiration from credential content (e.g., JWT)
		if expiration := v.detectIntrinsicExpiration(cred); expiration != nil {
			check.HasExpiration = true
			check.ExpiresAt = expiration
			check.IsExpired = expiration.Before(time.Now())
		}
	}

	// Generate recommendations
	if check.IsExpired {
		check.Recommendations = append(check.Recommendations, "Credential has expired and should be rotated immediately")
	} else if check.ExpiresIn != nil && *check.ExpiresIn < 7*24*time.Hour {
		check.Recommendations = append(check.Recommendations, "Plan credential rotation soon")
	}

	return check, nil
}

// Helper methods

func (v *credentialValidatorImpl) detectCredentialType(value string) credentials.CredentialType {
	// JWT Token
	if v.isJWTToken(value) {
		return credentials.CredentialTypeToken
	}

	// OAuth Bearer Token
	if strings.HasPrefix(value, "Bearer ") {
		return credentials.CredentialTypeOAuth
	}

	// Certificate (PEM format)
	if strings.Contains(value, "-----BEGIN CERTIFICATE-----") {
		return credentials.CredentialTypeCertificate
	}

	// API Key patterns
	if v.isAPIKeyPattern(value) {
		return credentials.CredentialTypeAPIKey
	}

	// Default to API Key
	return credentials.CredentialTypeAPIKey
}

func (v *credentialValidatorImpl) isJWTToken(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return false
	}

	// Try to decode the header
	if _, err := base64.RawURLEncoding.DecodeString(parts[0]); err != nil {
		return false
	}

	return true
}

func (v *credentialValidatorImpl) isAPIKeyPattern(value string) bool {
	// Common API key patterns
	patterns := []string{
		`^sk-[a-zA-Z0-9]{32,}$`, // OpenAI style
		`^[a-fA-F0-9]{32,}$`,    // Hex keys
		`^[a-zA-Z0-9_-]{20,}$`,  // Base64-like keys
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, value); matched {
			return true
		}
	}

	return false
}

func (v *credentialValidatorImpl) validateAPIKeyFormat(value string, result *FormatValidation) {
	// Length check
	if len(value) < 16 {
		result.FormatErrors = append(result.FormatErrors, "API key is too short (minimum 16 characters)")
	}

	// Pattern checks
	if strings.Contains(value, " ") {
		result.FormatWarnings = append(result.FormatWarnings, "API key contains spaces")
	}

	// Check for common prefixes
	if strings.HasPrefix(value, "sk-") {
		result.ValidatedFields = append(result.ValidatedFields, ValidatedField{
			Name:    "prefix",
			Value:   "sk-",
			IsValid: true,
			Score:   100,
		})
	}
}

func (v *credentialValidatorImpl) validateTokenFormat(value string, result *FormatValidation) {
	if v.isJWTToken(value) {
		v.validateJWTFormat(value, result)
	} else {
		v.validateBearerTokenFormat(value, result)
	}
}

func (v *credentialValidatorImpl) validateJWTFormat(value string, result *FormatValidation) {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		result.FormatErrors = append(result.FormatErrors, "JWT token must have exactly 3 parts")
		return
	}

	// Parse header
	headerData, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		result.FormatErrors = append(result.FormatErrors, "Invalid JWT header encoding")
		return
	}

	var header jwt.MapClaims
	if err := json.Unmarshal(headerData, &header); err != nil {
		result.FormatErrors = append(result.FormatErrors, "Invalid JWT header JSON")
		return
	}

	// Parse claims
	claimsData, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		result.FormatErrors = append(result.FormatErrors, "Invalid JWT claims encoding")
		return
	}

	var claims jwt.MapClaims
	if err := json.Unmarshal(claimsData, &claims); err != nil {
		result.FormatErrors = append(result.FormatErrors, "Invalid JWT claims JSON")
		return
	}

	// Create format details
	result.FormatDetails = &FormatDetails{
		Algorithm:    fmt.Sprintf("%v", header["alg"]),
		CustomFields: make(map[string]interface{}),
	}

	if iss, ok := claims["iss"]; ok {
		result.FormatDetails.Issuer = fmt.Sprintf("%v", iss)
	}

	if sub, ok := claims["sub"]; ok {
		result.FormatDetails.Subject = fmt.Sprintf("%v", sub)
	}
}

func (v *credentialValidatorImpl) validateCertificateFormat(value string, result *FormatValidation) {
	block, _ := pem.Decode([]byte(value))
	if block == nil {
		result.FormatErrors = append(result.FormatErrors, "Certificate is not in valid PEM format")
		return
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		result.FormatErrors = append(result.FormatErrors, fmt.Sprintf("Invalid certificate: %v", err))
		return
	}

	// Create format details
	result.FormatDetails = &FormatDetails{
		Algorithm:    cert.SignatureAlgorithm.String(),
		Subject:      cert.Subject.String(),
		Issuer:       cert.Issuer.String(),
		SerialNumber: cert.SerialNumber.String(),
	}
}

func (v *credentialValidatorImpl) analyzeCharacterComposition(value string) *CharacterComposition {
	comp := &CharacterComposition{}

	for _, char := range value {
		switch {
		case char >= 'A' && char <= 'Z':
			comp.Uppercase++
		case char >= 'a' && char <= 'z':
			comp.Lowercase++
		case char >= '0' && char <= '9':
			comp.Digits++
		case char == ' ' || char == '\t' || char == '\n':
			comp.Whitespace++
		default:
			comp.Special++
		}
	}

	// Calculate diversity score
	categories := 0
	if comp.Uppercase > 0 {
		categories++
	}
	if comp.Lowercase > 0 {
		categories++
	}
	if comp.Digits > 0 {
		categories++
	}
	if comp.Special > 0 {
		categories++
	}

	comp.Diversity = float64(categories) / 4.0

	return comp
}

func (v *credentialValidatorImpl) calculateEntropy(value string) float64 {
	if len(value) == 0 {
		return 0
	}

	// Count character frequency
	freq := make(map[rune]int)
	for _, char := range value {
		freq[char]++
	}

	// Calculate entropy
	entropy := 0.0
	length := float64(len(value))

	for _, count := range freq {
		p := float64(count) / length
		if p > 0 {
			entropy -= p * (float64(len(freq)) / length)
		}
	}

	return entropy
}

func (v *credentialValidatorImpl) calculateStrengthScore(analysis *StrengthAnalysis) int {
	score := 0

	// Length score (max 30 points)
	if analysis.Length >= 32 {
		score += 30
	} else if analysis.Length >= 20 {
		score += 20
	} else if analysis.Length >= 12 {
		score += 15
	} else if analysis.Length >= 8 {
		score += 10
	}

	// Diversity score (max 25 points)
	score += int(analysis.Composition.Diversity * 25)

	// Entropy score (max 25 points)
	if analysis.Entropy >= 4.0 {
		score += 25
	} else if analysis.Entropy >= 3.0 {
		score += 20
	} else if analysis.Entropy >= 2.0 {
		score += 15
	} else if analysis.Entropy >= 1.0 {
		score += 10
	}

	// Pattern penalties (max 20 points deduction)
	penaltyPoints := len(analysis.Patterns) * 5
	if penaltyPoints > 20 {
		penaltyPoints = 20
	}
	score -= penaltyPoints

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

func (v *credentialValidatorImpl) determineStrengthLevel(score int) StrengthLevel {
	switch {
	case score >= 90:
		return StrengthLevelExcellent
	case score >= 75:
		return StrengthLevelStrong
	case score >= 60:
		return StrengthLevelGood
	case score >= 40:
		return StrengthLevelFair
	default:
		return StrengthLevelWeak
	}
}

func (v *credentialValidatorImpl) detectIntrinsicExpiration(cred *credentials.Credential) *time.Time {
	// Try to extract expiration from JWT
	if v.isJWTToken(cred.Value) {
		parts := strings.Split(cred.Value, ".")
		if len(parts) >= 2 {
			claimsData, err := base64.RawURLEncoding.DecodeString(parts[1])
			if err == nil {
				var claims jwt.MapClaims
				if err := json.Unmarshal(claimsData, &claims); err == nil {
					if exp, ok := claims["exp"]; ok {
						if expFloat, ok := exp.(float64); ok {
							expTime := time.Unix(int64(expFloat), 0)
							return &expTime
						}
						if expStr, ok := exp.(string); ok {
							if expInt, err := strconv.ParseInt(expStr, 10, 64); err == nil {
								expTime := time.Unix(expInt, 0)
								return &expTime
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (v *credentialValidatorImpl) calculateOverallScore(response *ValidationResponse) int {
	totalScore := 0
	componentCount := 0

	if response.FormatCheck != nil {
		totalScore += response.FormatCheck.FormatScore
		componentCount++
	}

	if response.StrengthAnalysis != nil {
		totalScore += response.StrengthAnalysis.StrengthScore
		componentCount++
	}

	if response.ExpirationCheck != nil {
		totalScore += response.ExpirationCheck.ExpirationScore
		componentCount++
	}

	if response.ConnectivityTest != nil {
		if response.ConnectivityTest.Success {
			totalScore += 100
		} else {
			totalScore += 0
		}
		componentCount++
	}

	if response.ComplianceCheck != nil {
		totalScore += response.ComplianceCheck.ComplianceScore
		componentCount++
	}

	if componentCount == 0 {
		return 0
	}

	return totalScore / componentCount
}

// Placeholder implementations for remaining methods
func (v *credentialValidatorImpl) ValidateCompliance(ctx context.Context, cred *credentials.Credential, policies []string) (*ComplianceValidation, error) {
	return &ComplianceValidation{IsCompliant: true, ComplianceScore: 100}, nil
}

func (v *credentialValidatorImpl) BatchValidate(ctx context.Context, creds []*credentials.Credential) (*BatchValidationResult, error) {
	return &BatchValidationResult{}, nil
}

func (v *credentialValidatorImpl) GetValidationRules(ctx context.Context) ([]*ValidationRule, error) {
	return v.validationRules, nil
}

func (v *credentialValidatorImpl) ConfigureValidationRules(ctx context.Context, rules []*ValidationRule) error {
	v.validationRules = rules
	return nil
}

func (v *credentialValidatorImpl) validateOAuthFormat(value string, result *FormatValidation) {
	// OAuth format validation
}

func (v *credentialValidatorImpl) validatePasswordFormat(value string, result *FormatValidation) {
	// Password format validation
}

func (v *credentialValidatorImpl) validateGenericFormat(value string, result *FormatValidation) {
	// Generic format validation
}

func (v *credentialValidatorImpl) validateBearerTokenFormat(value string, result *FormatValidation) {
	// Bearer token format validation
}

func (v *credentialValidatorImpl) detectSecurityPatterns(value string) []SecurityPattern {
	patterns := []SecurityPattern{}

	// Check for common weak patterns
	commonPatterns := []string{"password", "123456", "qwerty", "admin"}
	for _, pattern := range commonPatterns {
		if strings.Contains(strings.ToLower(value), pattern) {
			patterns = append(patterns, SecurityPattern{
				Type:        "weak_pattern",
				Pattern:     pattern,
				Severity:    "high",
				Description: fmt.Sprintf("Contains common weak pattern: %s", pattern),
				Count:       1,
			})
		}
	}

	return patterns
}

func (v *credentialValidatorImpl) generateRecommendations(response *ValidationResponse) []string {
	recommendations := []string{}

	if response.FormatCheck != nil && !response.FormatCheck.IsValidFormat {
		recommendations = append(recommendations, "Fix credential format issues")
	}

	if response.StrengthAnalysis != nil && response.StrengthAnalysis.StrengthLevel == StrengthLevelWeak {
		recommendations = append(recommendations, "Use a stronger credential")
	}

	if response.ExpirationCheck != nil && response.ExpirationCheck.IsExpired {
		recommendations = append(recommendations, "Rotate expired credential immediately")
	}

	if response.ConnectivityTest != nil && !response.ConnectivityTest.Success {
		recommendations = append(recommendations, "Verify credential has correct permissions")
	}

	return recommendations
}

func (v *credentialValidatorImpl) initializeBuiltInRules() error {
	// Initialize built-in validation rules
	return nil
}

func (v *credentialValidatorImpl) logValidationEvent(action, credentialName string, success bool, err error) {
	if !v.config.EnableAuditLog || v.auditSvc == nil {
		return
	}

	safetyLevel := models.SafetyLevelSafe
	if err != nil {
		safetyLevel = models.SafetyLevelWarning
	}

	auditReq := models.AuditLogRequest{
		QueryText:        fmt.Sprintf("Credential validation: %s", action),
		GeneratedCommand: fmt.Sprintf("Validate credential %s with action %s", credentialName, action),
		SafetyLevel:      string(safetyLevel),
		ExecutionResult: map[string]interface{}{
			"credential_name": credentialName,
			"action":          action,
			"success":         success,
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

	if logErr := v.auditSvc.LogUserAction(context.Background(), auditReq); logErr != nil {
		log.Printf("Failed to log validation event: %v", logErr)
	}
}
