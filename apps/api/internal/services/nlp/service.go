package nlp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// NLPRequest represents enhanced request structure for command generation
type NLPRequest struct {
	Query       string            `json:"query"`
	Context     map[string]string `json:"context"`
	Namespace   string            `json:"namespace,omitempty"`
	UserRole    string            `json:"user_role"`
	Environment string            `json:"environment"`
}

// NLPResult represents enhanced response structure with safety features
type NLPResult struct {
	Command     string   `json:"command"`
	Explanation string   `json:"explanation"`
	SafetyLevel string   `json:"safety_level"`
	Confidence  float64  `json:"confidence"`
	Suggestions []string `json:"suggestions,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
}

// SafetyResult represents command safety validation result
type SafetyResult struct {
	SafetyLevel      string   `json:"safety_level"`
	SafetyReasons    []string `json:"safety_reasons"`
	IsBlocked        bool     `json:"is_blocked"`
	RequiresApproval bool     `json:"requires_approval"`
	Suggestions      []string `json:"suggestions,omitempty"`
}

// Service defines the enhanced NLP service interface with multi-provider support
type Service interface {
	// GenerateCommand converts natural language to kubectl commands with safety classification
	GenerateCommand(ctx context.Context, req NLPRequest) (*NLPResult, error)

	// ValidateCommand validates a kubectl command for safety and correctness (legacy support)
	ValidateCommand(ctx context.Context, command string) (*models.CommandValidationResult, error)

	// ValidateCommandSafety validates a kubectl command for enhanced safety and correctness
	ValidateCommandSafety(ctx context.Context, command string) (*SafetyResult, error)

	// GetProviders returns list of available providers
	GetProviders() []string

	// SetProvider sets the active provider
	SetProvider(provider string) error

	// ProcessQuery processes a natural language query and generates kubectl commands (legacy support)
	ProcessQuery(ctx context.Context, request *models.NLPRequest) (*models.NLPResponse, error)

	// GetSupportedProviders returns a list of available NLP providers (legacy support)
	GetSupportedProviders(ctx context.Context) ([]models.NLPProvider, error)

	// HealthCheck validates connectivity to NLP providers
	HealthCheck(ctx context.Context) error

	// GetMetrics returns NLP service metrics
	GetMetrics(ctx context.Context) (*models.NLPMetrics, error)
}

// serviceImpl implements the NLP Service interface
type serviceImpl struct {
	ollamaService  OllamaService
	openaiService  OpenAIService
	providers      []Provider
	config         *Config
	defaultTimeout time.Duration
}

// Config represents NLP service configuration
type Config struct {
	DefaultProvider    models.NLPProvider `json:"default_provider"`
	EnableFallback     bool               `json:"enable_fallback"`
	MaxRetries         int                `json:"max_retries"`
	TimeoutSeconds     int                `json:"timeout_seconds"`
	EnableCaching      bool               `json:"enable_caching"`
	CacheTTLMinutes    int                `json:"cache_ttl_minutes"`
	EnableRateLimiting bool               `json:"enable_rate_limiting"`
	RateLimit          int                `json:"rate_limit"` // requests per minute
}

// Provider represents a generic NLP provider interface
type Provider interface {
	Name() models.NLPProvider
	IsHealthy(ctx context.Context) bool
	ProcessQuery(ctx context.Context, request *models.NLPRequest) (*models.NLPResponse, error)
	HealthCheck(ctx context.Context) error
}

// OllamaService interface for Ollama integration
type OllamaService interface {
	ProcessQuery(ctx context.Context, request *models.NLPRequest) (*models.NLPResponse, error)
	HealthCheck(ctx context.Context) error
	IsAvailable(ctx context.Context) bool
}

// OpenAIService interface for OpenAI integration
type OpenAIService interface {
	ProcessQuery(ctx context.Context, request *models.NLPRequest) (*models.NLPResponse, error)
	HealthCheck(ctx context.Context) error
	IsAvailable(ctx context.Context) bool
}

// NewService creates a new NLP service instance
func NewService(ollamaService OllamaService, openaiService OpenAIService, config *Config) Service {
	if config == nil {
		config = &Config{
			DefaultProvider:    models.ProviderOllama,
			EnableFallback:     true,
			MaxRetries:         3,
			TimeoutSeconds:     30,
			EnableCaching:      true,
			CacheTTLMinutes:    60,
			EnableRateLimiting: true,
			RateLimit:          30, // 30 requests per minute
		}
	}

	service := &serviceImpl{
		ollamaService:  ollamaService,
		openaiService:  openaiService,
		config:         config,
		defaultTimeout: time.Duration(config.TimeoutSeconds) * time.Second,
	}

	// Initialize providers list
	service.initializeProviders()

	log.Printf("NLP Service initialized with default provider: %s", config.DefaultProvider)
	log.Printf("Fallback enabled: %v, Max retries: %d", config.EnableFallback, config.MaxRetries)

	return service
}

// initializeProviders sets up the provider list based on configuration
func (s *serviceImpl) initializeProviders() {
	s.providers = make([]Provider, 0)

	// Add providers based on availability and configuration
	if s.ollamaService != nil {
		s.providers = append(s.providers, &ollamaProvider{service: s.ollamaService})
	}

	if s.openaiService != nil {
		s.providers = append(s.providers, &openaiProvider{service: s.openaiService})
	}

	log.Printf("Initialized %d NLP providers", len(s.providers))
}

// ProcessQuery processes a natural language query with fallback support
func (s *serviceImpl) ProcessQuery(ctx context.Context, request *models.NLPRequest) (*models.NLPResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	startTime := time.Now()

	// Try primary provider first
	primaryProvider := s.getProvider(s.config.DefaultProvider)
	if primaryProvider != nil {
		log.Printf("Processing query with primary provider: %s", s.config.DefaultProvider)
		response, err := s.processWithProvider(ctx, primaryProvider, request)
		if err == nil && response != nil {
			response.ProcessingTimeMs = time.Since(startTime).Milliseconds()
			return response, nil
		}
		log.Printf("Primary provider failed: %v", err)
	}

	// Try fallback providers if enabled
	if s.config.EnableFallback {
		log.Printf("Attempting fallback providers")
		for _, provider := range s.providers {
			if provider.Name() == s.config.DefaultProvider {
				continue // Skip primary provider
			}

			if provider.IsHealthy(ctx) {
				log.Printf("Trying fallback provider: %s", provider.Name())
				response, err := s.processWithProvider(ctx, provider, request)
				if err == nil && response != nil {
					response.ProcessingTimeMs = time.Since(startTime).Milliseconds()
					return response, nil
				}
				log.Printf("Fallback provider %s failed: %v", provider.Name(), err)
			}
		}
	}

	return nil, fmt.Errorf("all NLP providers failed or unavailable")
}

// processWithProvider processes a request with a specific provider
func (s *serviceImpl) processWithProvider(ctx context.Context, provider Provider, request *models.NLPRequest) (*models.NLPResponse, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.defaultTimeout)
	defer cancel()

	// Set the provider in the request
	request.Provider = provider.Name()

	// Process the query
	response, err := provider.ProcessQuery(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("provider %s failed: %w", provider.Name(), err)
	}

	// Validate and enhance the response
	if response != nil {
		response.Provider = provider.Name()
		if response.CreatedAt.IsZero() {
			response.CreatedAt = time.Now()
		}

		// Auto-classify safety if not set
		if response.SafetyLevel == "" {
			response.SafetyLevel = s.classifyCommandSafety(response.GeneratedCommand)
		}
	}

	return response, nil
}

// ValidateCommand validates a kubectl command for safety and correctness
func (s *serviceImpl) ValidateCommand(ctx context.Context, command string) (*models.CommandValidationResult, error) {
	if command == "" {
		return &models.CommandValidationResult{
			IsValid:     false,
			SafetyLevel: models.NLPSafetyLevelDangerous,
			Errors:      []string{"Empty command provided"},
		}, nil
	}

	result := &models.CommandValidationResult{
		IsValid:     true,
		SafetyLevel: s.classifyCommandSafety(command),
		Warnings:    []string{},
		Errors:      []string{},
		Suggestions: []string{},
	}

	// Basic kubectl command validation
	if !s.isValidKubectlCommand(command) {
		result.IsValid = false
		result.SafetyLevel = models.NLPSafetyLevelDangerous
		result.Errors = append(result.Errors, "Invalid kubectl command format")
	}

	// Check for dangerous operations
	dangerousPatterns := []string{"delete", "destroy", "rm", "--force", "--cascade=foreground"}
	for _, pattern := range dangerousPatterns {
		if contains(command, pattern) {
			result.SafetyLevel = models.NLPSafetyLevelDangerous
			result.Warnings = append(result.Warnings, fmt.Sprintf("Contains potentially dangerous operation: %s", pattern))
		}
	}

	// Check for warning-level operations
	warningPatterns := []string{"create", "apply", "patch", "replace", "scale", "restart"}
	for _, pattern := range warningPatterns {
		if contains(command, pattern) && result.SafetyLevel == models.NLPSafetyLevelSafe {
			result.SafetyLevel = models.NLPSafetyLevelWarning
			result.Warnings = append(result.Warnings, fmt.Sprintf("Contains operation that may modify cluster state: %s", pattern))
		}
	}

	return result, nil
}

// GetSupportedProviders returns available NLP providers
func (s *serviceImpl) GetSupportedProviders(ctx context.Context) ([]models.NLPProvider, error) {
	providers := make([]models.NLPProvider, 0, len(s.providers))

	for _, provider := range s.providers {
		if provider.IsHealthy(ctx) {
			providers = append(providers, provider.Name())
		}
	}

	return providers, nil
}

// HealthCheck validates connectivity to NLP providers
func (s *serviceImpl) HealthCheck(ctx context.Context) error {
	healthyProviders := 0

	for _, provider := range s.providers {
		if err := provider.HealthCheck(ctx); err != nil {
			log.Printf("Provider %s health check failed: %v", provider.Name(), err)
		} else {
			healthyProviders++
		}
	}

	if healthyProviders == 0 {
		return fmt.Errorf("no healthy NLP providers available")
	}

	log.Printf("NLP service health check passed: %d/%d providers healthy", healthyProviders, len(s.providers))
	return nil
}

// GetMetrics returns NLP service metrics
func (s *serviceImpl) GetMetrics(ctx context.Context) (*models.NLPMetrics, error) {
	// For now, return basic metrics - this would be enhanced with actual metrics collection
	return &models.NLPMetrics{
		Provider:         s.config.DefaultProvider,
		RequestCount:     0,
		SuccessCount:     0,
		ErrorCount:       0,
		AverageLatencyMs: 0,
		TokensUsed:       0,
		LastUsed:         time.Now(),
	}, nil
}

// Helper methods

func (s *serviceImpl) getProvider(providerType models.NLPProvider) Provider {
	for _, provider := range s.providers {
		if provider.Name() == providerType {
			return provider
		}
	}
	return nil
}

func (s *serviceImpl) classifyCommandSafety(command string) models.SafetyLevel {
	if command == "" {
		return models.NLPSafetyLevelDangerous
	}

	// Dangerous operations
	dangerousPatterns := []string{
		"delete", "destroy", "rm", "--force", "--cascade=foreground",
		"drain", "cordon", "evict", "--grace-period=0",
	}

	for _, pattern := range dangerousPatterns {
		if contains(command, pattern) {
			return models.NLPSafetyLevelDangerous
		}
	}

	// Warning-level operations
	warningPatterns := []string{
		"create", "apply", "patch", "replace", "scale", "restart",
		"edit", "label", "annotate", "expose", "rollout",
	}

	for _, pattern := range warningPatterns {
		if contains(command, pattern) {
			return models.NLPSafetyLevelWarning
		}
	}

	// Default to safe for read-only operations
	return models.NLPSafetyLevelSafe
}

func (s *serviceImpl) isValidKubectlCommand(command string) bool {
	// Basic validation - should start with kubectl
	return len(command) > 7 && (command[:7] == "kubectl" || command[:8] == " kubectl")
}

// Utility functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Provider implementations

type ollamaProvider struct {
	service OllamaService
}

func (p *ollamaProvider) Name() models.NLPProvider {
	return models.ProviderOllama
}

func (p *ollamaProvider) IsHealthy(ctx context.Context) bool {
	return p.service.IsAvailable(ctx)
}

func (p *ollamaProvider) ProcessQuery(ctx context.Context, request *models.NLPRequest) (*models.NLPResponse, error) {
	return p.service.ProcessQuery(ctx, request)
}

func (p *ollamaProvider) HealthCheck(ctx context.Context) error {
	return p.service.HealthCheck(ctx)
}

type openaiProvider struct {
	service OpenAIService
}

func (p *openaiProvider) Name() models.NLPProvider {
	return models.ProviderOpenAI
}

func (p *openaiProvider) IsHealthy(ctx context.Context) bool {
	return p.service.IsAvailable(ctx)
}

func (p *openaiProvider) ProcessQuery(ctx context.Context, request *models.NLPRequest) (*models.NLPResponse, error) {
	return p.service.ProcessQuery(ctx, request)
}

func (p *openaiProvider) HealthCheck(ctx context.Context) error {
	return p.service.HealthCheck(ctx)
}

// Enhanced NLP Service Methods - Implementing new interface

// GenerateCommand converts natural language to kubectl commands with enhanced safety features
func (s *serviceImpl) GenerateCommand(ctx context.Context, req NLPRequest) (*NLPResult, error) {
	// Input sanitization - prevent prompt injection attacks
	sanitizedQuery := s.sanitizeInput(req.Query)

	// Convert to internal request format
	internalReq := &models.NLPRequest{
		Query:       sanitizedQuery,
		Context:     s.formatContextString(req.Context),
		ClusterInfo: req.Namespace,
		Provider:    s.config.DefaultProvider,
		CreatedAt:   time.Now(),
	}

	// Process with the existing ProcessQuery logic
	response, err := s.ProcessQuery(ctx, internalReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate command: %w", err)
	}

	// Convert to enhanced result format with safety classification
	result := &NLPResult{
		Command:     response.GeneratedCommand,
		Explanation: response.Explanation,
		SafetyLevel: string(response.SafetyLevel),
		Confidence:  response.Confidence,
		Suggestions: []string{},
		Warnings:    []string{},
	}

	// Apply context-aware safety adjustments
	result = s.applyContextAwareSafety(result, req)

	// Add safety warnings and suggestions
	if result.SafetyLevel == "dangerous" {
		result.Warnings = append(result.Warnings, "This command may cause significant cluster changes")
		result.Suggestions = append(result.Suggestions, "Consider using --dry-run first to preview changes")
	}

	return result, nil
}

// ValidateCommandSafety validates a kubectl command with enhanced safety features
func (s *serviceImpl) ValidateCommandSafety(ctx context.Context, command string) (*SafetyResult, error) {
	// Use existing validation logic
	validation, err := s.ProcessQuery(ctx, &models.NLPRequest{
		Query:     command,
		Provider:  s.config.DefaultProvider,
		CreatedAt: time.Now(),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to validate command: %w", err)
	}

	// Convert to enhanced safety result
	result := &SafetyResult{
		SafetyLevel:      string(validation.SafetyLevel),
		SafetyReasons:    []string{},
		IsBlocked:        false,
		RequiresApproval: false,
		Suggestions:      []string{},
	}

	// Apply comprehensive safety analysis
	result = s.analyzeCommandSafety(command, result)

	return result, nil
}

// GetProviders returns list of available providers
func (s *serviceImpl) GetProviders() []string {
	providers := make([]string, len(s.providers))
	for i, provider := range s.providers {
		providers[i] = string(provider.Name())
	}
	return providers
}

// SetProvider sets the active provider
func (s *serviceImpl) SetProvider(provider string) error {
	// Validate provider exists
	found := false
	for _, p := range s.providers {
		if string(p.Name()) == provider {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("provider %s not found", provider)
	}

	// Update configuration
	s.config.DefaultProvider = models.NLPProvider(provider)
	log.Printf("NLP service provider changed to: %s", provider)

	return nil
}

// Helper methods for enhanced functionality

// sanitizeInput prevents prompt injection attacks
func (s *serviceImpl) sanitizeInput(query string) string {
	// Remove potential prompt injection patterns
	dangerousPatterns := []string{
		"ignore previous instructions",
		"system:",
		"assistant:",
		"```",
		"<script>",
		"</script>",
	}

	sanitized := query
	for _, pattern := range dangerousPatterns {
		// Simple replacement - in production would use more sophisticated detection
		for contains(sanitized, pattern) {
			sanitized = strings.Replace(sanitized, pattern, "[FILTERED]", -1)
		}
	}

	// Limit length to prevent abuse
	if len(sanitized) > 2000 {
		sanitized = sanitized[:2000] + "..."
	}

	return sanitized
}

// formatContextString converts context map to string representation
func (s *serviceImpl) formatContextString(context map[string]string) string {
	if len(context) == 0 {
		return ""
	}

	var contextParts []string
	for k, v := range context {
		contextParts = append(contextParts, fmt.Sprintf("%s=%s", k, v))
	}

	return fmt.Sprintf("Context: %s", strings.Join(contextParts, ", "))
}

// applyContextAwareSafety adjusts safety based on user role and environment
func (s *serviceImpl) applyContextAwareSafety(result *NLPResult, req NLPRequest) *NLPResult {
	// Increase safety level for production environments
	if req.Environment == "production" {
		if result.SafetyLevel == "safe" {
			result.SafetyLevel = "warning"
			result.Warnings = append(result.Warnings, "Enhanced safety applied for production environment")
		} else if result.SafetyLevel == "warning" {
			result.SafetyLevel = "dangerous"
			result.Warnings = append(result.Warnings, "Command blocked for production environment")
		}
	}

	// Apply role-based restrictions
	if req.UserRole == "developer" && result.SafetyLevel == "dangerous" {
		result.Warnings = append(result.Warnings, "Developer role: dangerous operations require approval")
		result.Suggestions = append(result.Suggestions, "Contact cluster administrator for approval")
	}

	// Namespace-specific restrictions
	criticalNamespaces := []string{"kube-system", "kube-public", "istio-system"}
	for _, ns := range criticalNamespaces {
		if req.Namespace == ns && result.SafetyLevel != "safe" {
			result.SafetyLevel = "dangerous"
			result.Warnings = append(result.Warnings, fmt.Sprintf("Critical namespace %s requires special caution", ns))
		}
	}

	return result
}

// analyzeCommandSafety performs comprehensive command safety analysis
func (s *serviceImpl) analyzeCommandSafety(command string, result *SafetyResult) *SafetyResult {
	// Analyze for dangerous patterns with reasons
	dangerousPatterns := map[string]string{
		"delete":  "Destructive operation that removes resources",
		"--force": "Force flag bypasses safety checks",
		"cascade": "Cascade deletion affects dependent resources",
		"drain":   "Node drain affects running workloads",
	}

	for pattern, reason := range dangerousPatterns {
		if contains(command, pattern) {
			result.SafetyLevel = "dangerous"
			result.SafetyReasons = append(result.SafetyReasons, reason)
			result.RequiresApproval = true
		}
	}

	// Check if command should be blocked
	blockingPatterns := []string{
		"rm -rf", "destroy", "--cascade=foreground", "--grace-period=0",
	}

	for _, pattern := range blockingPatterns {
		if contains(command, pattern) {
			result.IsBlocked = true
			result.SafetyReasons = append(result.SafetyReasons, fmt.Sprintf("Blocked pattern: %s", pattern))
			result.Suggestions = append(result.Suggestions, "Use safer alternatives or seek administrator approval")
		}
	}

	return result
}
