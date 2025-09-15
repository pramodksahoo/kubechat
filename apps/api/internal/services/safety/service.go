package safety

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
)

// SafetyLevel represents command safety classification
type SafetyLevel string

const (
	SafetyLevelSafe      SafetyLevel = "safe"
	SafetyLevelWarning   SafetyLevel = "warning"
	SafetyLevelDangerous SafetyLevel = "dangerous"
)

// SafetyClassification represents detailed safety analysis
type SafetyClassification struct {
	Level            SafetyLevel `json:"level"`
	Score            float64     `json:"score"`             // 0-100 safety score
	Reasons          []string    `json:"reasons"`           // List of safety concerns
	Blocked          bool        `json:"blocked"`           // Whether command should be blocked
	RequiresApproval bool        `json:"requires_approval"` // Whether command needs approval
	Suggestions      []string    `json:"suggestions"`       // Safer alternatives
	Warnings         []string    `json:"warnings"`          // Warning messages
}

// ContextualSafetyRequest represents context-aware safety analysis request
type ContextualSafetyRequest struct {
	Command     string            `json:"command"`
	UserRole    string            `json:"user_role"`
	Environment string            `json:"environment"`
	Namespace   string            `json:"namespace,omitempty"`
	Context     map[string]string `json:"context,omitempty"`
}

// Service defines the safety classification service interface
type Service interface {
	// ClassifyCommand performs basic command safety classification
	ClassifyCommand(ctx context.Context, command string) (*SafetyClassification, error)

	// ClassifyWithContext performs context-aware safety classification
	ClassifyWithContext(ctx context.Context, req ContextualSafetyRequest) (*SafetyClassification, error)

	// ValidatePrompt checks for prompt injection attacks
	ValidatePrompt(ctx context.Context, prompt string) (bool, []string, error)

	// HealthCheck validates safety service health
	HealthCheck(ctx context.Context) error
}

// Config represents safety service configuration
type Config struct {
	// Safety thresholds
	SafeThreshold      float64 `json:"safe_threshold"`      // Below this is safe (default: 30)
	WarningThreshold   float64 `json:"warning_threshold"`   // Below this is warning (default: 70)
	DangerousThreshold float64 `json:"dangerous_threshold"` // Above this is dangerous (default: 70)

	// Environment settings
	ProductionMode  bool `json:"production_mode"`  // Stricter rules in production
	BlockDangerous  bool `json:"block_dangerous"`  // Block dangerous commands
	RequireApproval bool `json:"require_approval"` // Require approval for risky ops

	// Namespace protections
	ProtectedNamespaces []string `json:"protected_namespaces"`
}

// serviceImpl implements the Safety Service interface
type serviceImpl struct {
	config           *Config
	patterns         *SafetyPatterns
	promptValidators []PromptValidator
}

// SafetyPatterns holds regex patterns for command classification
type SafetyPatterns struct {
	// Dangerous patterns with weights
	Dangerous map[string]float64
	// Warning patterns with weights
	Warning map[string]float64
	// Safe patterns (negative weights)
	Safe map[string]float64
}

// PromptValidator interface for prompt injection detection
type PromptValidator interface {
	Validate(prompt string) (bool, string)
}

// NewService creates a new safety classification service
func NewService(config *Config) Service {
	if config == nil {
		config = &Config{
			SafeThreshold:       30.0,
			WarningThreshold:    70.0,
			DangerousThreshold:  70.0,
			ProductionMode:      false,
			BlockDangerous:      true,
			RequireApproval:     true,
			ProtectedNamespaces: []string{"kube-system", "kube-public", "istio-system", "monitoring"},
		}
	}

	patterns := initializeSafetyPatterns()
	validators := initializePromptValidators()

	service := &serviceImpl{
		config:           config,
		patterns:         patterns,
		promptValidators: validators,
	}

	log.Printf("Safety service initialized with thresholds: safe=%.1f, warning=%.1f, dangerous=%.1f",
		config.SafeThreshold, config.WarningThreshold, config.DangerousThreshold)
	log.Printf("Production mode: %v, Block dangerous: %v", config.ProductionMode, config.BlockDangerous)

	return service
}

// ClassifyCommand performs basic command safety classification
func (s *serviceImpl) ClassifyCommand(ctx context.Context, command string) (*SafetyClassification, error) {
	if command == "" {
		return &SafetyClassification{
			Level:            SafetyLevelDangerous,
			Score:            100.0,
			Reasons:          []string{"Empty command provided"},
			Blocked:          true,
			RequiresApproval: false,
			Suggestions:      []string{"Provide a valid kubectl command"},
			Warnings:         []string{"No command specified"},
		}, nil
	}

	// Calculate safety score
	score := s.calculateSafetyScore(command)
	level := s.determineLevel(score)

	// Build classification result
	classification := &SafetyClassification{
		Level:            level,
		Score:            score,
		Reasons:          s.identifyReasons(command),
		Blocked:          s.shouldBlock(level, score),
		RequiresApproval: s.requiresApproval(level, score),
		Suggestions:      s.generateSuggestions(command, level),
		Warnings:         s.generateWarnings(command, level),
	}

	return classification, nil
}

// ClassifyWithContext performs context-aware safety classification
func (s *serviceImpl) ClassifyWithContext(ctx context.Context, req ContextualSafetyRequest) (*SafetyClassification, error) {
	// Start with basic classification
	baseClassification, err := s.ClassifyCommand(ctx, req.Command)
	if err != nil {
		return nil, fmt.Errorf("base classification failed: %w", err)
	}

	// Apply contextual adjustments
	contextClassification := s.applyContextualRules(baseClassification, req)

	return contextClassification, nil
}

// ValidatePrompt checks for prompt injection attacks
func (s *serviceImpl) ValidatePrompt(ctx context.Context, prompt string) (bool, []string, error) {
	issues := make([]string, 0)

	for _, validator := range s.promptValidators {
		if valid, issue := validator.Validate(prompt); !valid {
			issues = append(issues, issue)
		}
	}

	isValid := len(issues) == 0
	return isValid, issues, nil
}

// HealthCheck validates safety service health
func (s *serviceImpl) HealthCheck(ctx context.Context) error {
	// Test basic classification with a known safe command
	testCmd := "kubectl get pods"
	classification, err := s.ClassifyCommand(ctx, testCmd)
	if err != nil {
		return fmt.Errorf("basic classification test failed: %w", err)
	}

	if classification.Level != SafetyLevelSafe {
		return fmt.Errorf("safety classification inconsistency: expected safe, got %s", classification.Level)
	}

	// Test prompt validation
	testPrompt := "show me the pods"
	valid, _, err := s.ValidatePrompt(ctx, testPrompt)
	if err != nil {
		return fmt.Errorf("prompt validation test failed: %w", err)
	}

	if !valid {
		return fmt.Errorf("prompt validation inconsistency: expected valid prompt to pass")
	}

	log.Println("Safety service health check passed")
	return nil
}

// Helper methods

func (s *serviceImpl) calculateSafetyScore(command string) float64 {
	score := 0.0
	lowerCmd := strings.ToLower(command)

	// Check dangerous patterns (positive score = more dangerous)
	for pattern, weight := range s.patterns.Dangerous {
		if matched, _ := regexp.MatchString(pattern, lowerCmd); matched {
			score += weight
		}
	}

	// Check warning patterns
	for pattern, weight := range s.patterns.Warning {
		if matched, _ := regexp.MatchString(pattern, lowerCmd); matched {
			score += weight
		}
	}

	// Check safe patterns (negative score = safer)
	for pattern, weight := range s.patterns.Safe {
		if matched, _ := regexp.MatchString(pattern, lowerCmd); matched {
			score -= weight
		}
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return score
}

func (s *serviceImpl) determineLevel(score float64) SafetyLevel {
	if score >= s.config.DangerousThreshold {
		return SafetyLevelDangerous
	} else if score >= s.config.WarningThreshold {
		return SafetyLevelWarning
	} else {
		return SafetyLevelSafe
	}
}

func (s *serviceImpl) identifyReasons(command string) []string {
	reasons := make([]string, 0)
	lowerCmd := strings.ToLower(command)

	// Check for specific dangerous operations
	dangerousOps := map[string]string{
		"delete":           "Destructive operation that removes resources",
		"destroy":          "Destructive operation that destroys resources",
		"--force":          "Force flag bypasses safety checks",
		"--cascade":        "Cascade operation affects dependent resources",
		"--grace-period=0": "Zero grace period forces immediate termination",
		"drain":            "Node drain operation affects running workloads",
		"cordon":           "Node cordon operation prevents scheduling",
		"evict":            "Pod eviction operation disrupts services",
	}

	for pattern, reason := range dangerousOps {
		if strings.Contains(lowerCmd, pattern) {
			reasons = append(reasons, reason)
		}
	}

	return reasons
}

func (s *serviceImpl) shouldBlock(level SafetyLevel, score float64) bool {
	if !s.config.BlockDangerous {
		return false
	}

	// Block dangerous operations in production mode
	if s.config.ProductionMode && level == SafetyLevelDangerous {
		return true
	}

	// Block extremely dangerous operations (score > 90)
	if score > 90 {
		return true
	}

	return false
}

func (s *serviceImpl) requiresApproval(level SafetyLevel, score float64) bool {
	if !s.config.RequireApproval {
		return false
	}

	return level == SafetyLevelDangerous || score > 50
}

func (s *serviceImpl) generateSuggestions(command string, level SafetyLevel) []string {
	suggestions := make([]string, 0)

	if level == SafetyLevelDangerous {
		suggestions = append(suggestions, "Consider using --dry-run=client to preview changes")
		suggestions = append(suggestions, "Use specific resource names instead of --all")
		suggestions = append(suggestions, "Add --grace-period=30 for safer deletion")
	}

	if level != SafetyLevelSafe {
		suggestions = append(suggestions, "Test in development environment first")
		suggestions = append(suggestions, "Create resource backup before modification")
	}

	return suggestions
}

func (s *serviceImpl) generateWarnings(command string, level SafetyLevel) []string {
	warnings := make([]string, 0)

	if level == SafetyLevelDangerous {
		warnings = append(warnings, "This command may cause significant cluster changes")
		warnings = append(warnings, "Service disruption may occur")
	}

	if level == SafetyLevelWarning {
		warnings = append(warnings, "This operation modifies cluster state")
	}

	return warnings
}

func (s *serviceImpl) applyContextualRules(base *SafetyClassification, req ContextualSafetyRequest) *SafetyClassification {
	result := *base // Copy base classification

	// Environment-based adjustments
	if req.Environment == "production" {
		result = s.adjustForProduction(result, req.Command)
	}

	// Role-based adjustments
	result = s.adjustForUserRole(result, req.UserRole, req.Command)

	// Namespace-based adjustments
	if req.Namespace != "" {
		result = s.adjustForNamespace(result, req.Namespace, req.Command)
	}

	return &result
}

func (s *serviceImpl) adjustForProduction(classification SafetyClassification, command string) SafetyClassification {
	// Increase severity in production
	if classification.Level == SafetyLevelSafe {
		classification.Level = SafetyLevelWarning
		classification.Score += 20
		classification.Warnings = append(classification.Warnings, "Enhanced safety applied for production environment")
	} else if classification.Level == SafetyLevelWarning {
		classification.Level = SafetyLevelDangerous
		classification.Score += 30
		classification.RequiresApproval = true
		classification.Warnings = append(classification.Warnings, "Production environment requires extra caution")
	} else {
		// Already dangerous - make it blocked
		classification.Blocked = true
		classification.Warnings = append(classification.Warnings, "Dangerous operations blocked in production")
	}

	return classification
}

func (s *serviceImpl) adjustForUserRole(classification SafetyClassification, role, command string) SafetyClassification {
	switch strings.ToLower(role) {
	case "developer", "dev":
		if classification.Level == SafetyLevelDangerous {
			classification.RequiresApproval = true
			classification.Warnings = append(classification.Warnings, "Developer role: dangerous operations require approval")
			classification.Suggestions = append(classification.Suggestions, "Contact cluster administrator for approval")
		}
	case "readonly", "viewer":
		// Readonly users should only have safe operations
		if classification.Level != SafetyLevelSafe {
			classification.Level = SafetyLevelDangerous
			classification.Blocked = true
			classification.Reasons = append(classification.Reasons, "User role limited to read-only operations")
		}
	}

	return classification
}

func (s *serviceImpl) adjustForNamespace(classification SafetyClassification, namespace, command string) SafetyClassification {
	// Check if namespace is protected
	for _, protectedNs := range s.config.ProtectedNamespaces {
		if namespace == protectedNs {
			if classification.Level != SafetyLevelSafe {
				classification.Level = SafetyLevelDangerous
				classification.RequiresApproval = true
				classification.Warnings = append(classification.Warnings,
					fmt.Sprintf("Critical namespace %s requires special caution", namespace))
			}
			break
		}
	}

	return classification
}

// Initialize safety patterns
func initializeSafetyPatterns() *SafetyPatterns {
	return &SafetyPatterns{
		Dangerous: map[string]float64{
			`delete.*--all`:                95.0,  // Delete all resources
			`delete.*--force`:              90.0,  // Force delete
			`delete.*--grace-period=0`:     85.0,  // Immediate delete
			`delete.*--cascade=foreground`: 80.0,  // Cascading delete
			`destroy`:                      95.0,  // Destroy command
			`rm.*-rf`:                      100.0, // Force recursive remove
			`drain.*--force`:               75.0,  // Force node drain
			`delete.*node`:                 85.0,  // Delete nodes
			`delete.*namespace`:            80.0,  // Delete namespaces
			`delete.*pv|persistentvolume`:  75.0,  // Delete persistent volumes
		},
		Warning: map[string]float64{
			`create`:                      30.0, // Create resources
			`apply`:                       25.0, // Apply configurations
			`patch`:                       35.0, // Patch resources
			`replace`:                     40.0, // Replace resources
			`scale.*--replicas=0`:         50.0, // Scale to zero
			`scale.*replicas.*[5-9][0-9]`: 45.0, // Scale to high numbers
			`edit`:                        30.0, // Edit resources
			`label`:                       20.0, // Label resources
			`annotate`:                    20.0, // Annotate resources
			`expose`:                      25.0, // Expose services
			`rollout.*restart`:            40.0, // Restart rollouts
			`drain`:                       60.0, // Node drain
			`cordon`:                      45.0, // Node cordon
			`uncordon`:                    30.0, // Node uncordon
		},
		Safe: map[string]float64{
			`get`:                 20.0, // Get resources
			`describe`:            15.0, // Describe resources
			`logs`:                10.0, // View logs
			`top`:                 10.0, // Resource usage
			`explain`:             15.0, // Explain resources
			`version`:             20.0, // Version info
			`cluster-info`:        15.0, // Cluster info
			`config.*view`:        10.0, // View config
			`auth.*can-i`:         15.0, // Check permissions
			`--dry-run`:           15.0, // Dry run operations
			`--output.*yaml|json`: 10.0, // Output formatting
		},
	}
}

// Initialize prompt validators
func initializePromptValidators() []PromptValidator {
	return []PromptValidator{
		&PromptInjectionValidator{},
		&ScriptValidator{},
		&SystemCommandValidator{},
	}
}

// Prompt injection validator
type PromptInjectionValidator struct{}

func (v *PromptInjectionValidator) Validate(prompt string) (bool, string) {
	lowerPrompt := strings.ToLower(prompt)

	dangerousPatterns := []string{
		"ignore previous instructions",
		"ignore all previous instructions",
		"system:",
		"assistant:",
		"forget everything",
		"new instructions:",
		"override",
		"bypass",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerPrompt, pattern) {
			return false, fmt.Sprintf("Potential prompt injection detected: %s", pattern)
		}
	}

	return true, ""
}

// Script injection validator
type ScriptValidator struct{}

func (v *ScriptValidator) Validate(prompt string) (bool, string) {
	lowerPrompt := strings.ToLower(prompt)

	scriptPatterns := []string{
		"<script>",
		"</script>",
		"javascript:",
		"eval(",
		"exec(",
		"system(",
		"`;",
		"&&",
		"||",
		"|",
	}

	for _, pattern := range scriptPatterns {
		if strings.Contains(lowerPrompt, pattern) {
			return false, fmt.Sprintf("Potential script injection detected: %s", pattern)
		}
	}

	return true, ""
}

// System command validator
type SystemCommandValidator struct{}

func (v *SystemCommandValidator) Validate(prompt string) (bool, string) {
	// Check for system commands that shouldn't be in kubectl queries
	systemCommands := []string{
		"rm ", "mv ", "cp ", "chmod ", "chown ",
		"kill ", "killall ", "ps ", "top ",
		"wget ", "curl ", "ssh ", "scp ",
		"/bin/", "/usr/bin/", "/sbin/",
	}

	for _, cmd := range systemCommands {
		if strings.Contains(prompt, cmd) {
			return false, fmt.Sprintf("System command detected in prompt: %s", cmd)
		}
	}

	return true, ""
}
