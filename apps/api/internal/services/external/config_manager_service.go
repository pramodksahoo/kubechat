package external

import (
	"fmt"
	"sync"
	"time"
)

// configManagerImpl implements ProviderConfigManager interface (Task 7.5)
type configManagerImpl struct {
	mu             sync.RWMutex
	configurations map[string]*ProviderConfig
	configHistory  map[string][]*ConfigurationChange
	templates      map[string]*ProviderConfigTemplate
	validators     map[string]ConfigValidator
	backups        []*ConfigurationBackup
}

// ConfigurationChange tracks configuration changes
type ConfigurationChange struct {
	ID           string                 `json:"id"`
	ProviderName string                 `json:"provider_name"`
	ChangeType   string                 `json:"change_type"` // "create", "update", "delete"
	OldConfig    *ProviderConfig        `json:"old_config,omitempty"`
	NewConfig    *ProviderConfig        `json:"new_config,omitempty"`
	Changes      map[string]interface{} `json:"changes"`
	User         string                 `json:"user,omitempty"`
	Reason       string                 `json:"reason,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// ProviderConfigTemplate defines configuration templates for different provider types
type ProviderConfigTemplate struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	ProviderType string              `json:"provider_type"`
	Version      string              `json:"version"`
	Description  string              `json:"description"`
	Template     *ProviderConfig     `json:"template"`
	Variables    map[string]Variable `json:"variables"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

// Variable defines template variables
type Variable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value"`
	Validation   string      `json:"validation,omitempty"`
}

// ConfigValidator validates provider configurations
type ConfigValidator func(*ProviderConfig) error

// ConfigurationBackup stores configuration backups
type ConfigurationBackup struct {
	ID             string                     `json:"id"`
	Name           string                     `json:"name"`
	Description    string                     `json:"description"`
	Configurations map[string]*ProviderConfig `json:"configurations"`
	Timestamp      time.Time                  `json:"timestamp"`
}

// NewProviderConfigManager creates a new provider configuration manager
func NewProviderConfigManager() ProviderConfigManager {
	manager := &configManagerImpl{
		configurations: make(map[string]*ProviderConfig),
		configHistory:  make(map[string][]*ConfigurationChange),
		templates:      make(map[string]*ProviderConfigTemplate),
		validators:     make(map[string]ConfigValidator),
		backups:        make([]*ConfigurationBackup, 0),
	}

	// Initialize default templates and validators
	manager.initializeDefaultTemplates()
	manager.initializeDefaultValidators()

	return manager
}

func (c *configManagerImpl) initializeDefaultTemplates() {
	templates := []*ProviderConfigTemplate{
		{
			ID:           "openai_template",
			Name:         "OpenAI Provider Template",
			ProviderType: "ai",
			Version:      "1.0.0",
			Description:  "Template for OpenAI API configuration",
			Template: &ProviderConfig{
				Type:           "ai",
				Enabled:        false,
				BaseURL:        "https://api.openai.com/v1",
				Models:         []string{"gpt-3.5-turbo", "gpt-4"},
				Capabilities:   []string{"text_generation", "conversation"},
				MaxConcurrency: 10,
				RateLimits: &RateLimits{
					RequestsPerSecond: 10,
					RequestsPerMinute: 60,
					TokensPerMinute:   90000,
				},
				Timeouts: &TimeoutConfig{
					RequestTimeout:    30 * time.Second,
					ConnectionTimeout: 10 * time.Second,
					ReadTimeout:       60 * time.Second,
					WriteTimeout:      30 * time.Second,
				},
				HealthCheck: &HealthCheckConfig{
					Enabled:          true,
					Interval:         5 * time.Minute,
					Timeout:          10 * time.Second,
					FailureThreshold: 3,
					SuccessThreshold: 2,
					Endpoint:         "/models",
				},
			},
			Variables: map[string]Variable{
				"api_key": {
					Name:        "api_key",
					Type:        "string",
					Description: "OpenAI API key",
					Required:    true,
				},
				"organization": {
					Name:        "organization",
					Type:        "string",
					Description: "OpenAI organization ID",
					Required:    false,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:           "anthropic_template",
			Name:         "Anthropic Claude Template",
			ProviderType: "ai",
			Version:      "1.0.0",
			Description:  "Template for Anthropic Claude API configuration",
			Template: &ProviderConfig{
				Type:           "ai",
				Enabled:        false,
				BaseURL:        "https://api.anthropic.com",
				Models:         []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"},
				Capabilities:   []string{"text_generation", "conversation", "reasoning"},
				MaxConcurrency: 5,
				RateLimits: &RateLimits{
					RequestsPerSecond: 5,
					RequestsPerMinute: 50,
					TokensPerMinute:   50000,
				},
				Timeouts: &TimeoutConfig{
					RequestTimeout:    45 * time.Second,
					ConnectionTimeout: 10 * time.Second,
					ReadTimeout:       90 * time.Second,
					WriteTimeout:      30 * time.Second,
				},
				HealthCheck: &HealthCheckConfig{
					Enabled:          true,
					Interval:         5 * time.Minute,
					Timeout:          15 * time.Second,
					FailureThreshold: 3,
					SuccessThreshold: 2,
				},
			},
			Variables: map[string]Variable{
				"api_key": {
					Name:        "api_key",
					Type:        "string",
					Description: "Anthropic API key",
					Required:    true,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:           "local_ai_template",
			Name:         "Local AI Provider Template",
			ProviderType: "ai",
			Version:      "1.0.0",
			Description:  "Template for local AI models (Ollama, etc.)",
			Template: &ProviderConfig{
				Type:           "ai",
				Enabled:        true,
				BaseURL:        "http://localhost:11434",
				Models:         []string{"llama3.2:3b", "mistral", "codellama"},
				Capabilities:   []string{"text_generation", "code_generation"},
				MaxConcurrency: 3,
				RateLimits: &RateLimits{
					RequestsPerSecond:  5,
					RequestsPerMinute:  100,
					ConcurrentRequests: 3,
				},
				Timeouts: &TimeoutConfig{
					RequestTimeout:    120 * time.Second,
					ConnectionTimeout: 5 * time.Second,
					ReadTimeout:       180 * time.Second,
					WriteTimeout:      30 * time.Second,
				},
				HealthCheck: &HealthCheckConfig{
					Enabled:          true,
					Interval:         2 * time.Minute,
					Timeout:          5 * time.Second,
					FailureThreshold: 2,
					SuccessThreshold: 1,
					Endpoint:         "/api/tags",
				},
			},
			Variables: map[string]Variable{
				"base_url": {
					Name:         "base_url",
					Type:         "string",
					Description:  "Base URL for the local AI service",
					Required:     true,
					DefaultValue: "http://localhost:11434",
				},
				"model": {
					Name:         "model",
					Type:         "string",
					Description:  "Default model to use",
					Required:     false,
					DefaultValue: "llama3.2:3b",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, template := range templates {
		c.templates[template.ID] = template
	}
}

func (c *configManagerImpl) initializeDefaultValidators() {
	// OpenAI validator
	c.validators["openai"] = func(config *ProviderConfig) error {
		if config.APIKey == "" && config.SecretKeyName == "" {
			return fmt.Errorf("OpenAI provider requires either api_key or secret_key_name")
		}
		if config.BaseURL == "" {
			return fmt.Errorf("OpenAI provider requires base_url")
		}
		if len(config.Models) == 0 {
			return fmt.Errorf("OpenAI provider requires at least one model")
		}
		return nil
	}

	// Anthropic validator
	c.validators["anthropic_claude"] = func(config *ProviderConfig) error {
		if config.APIKey == "" && config.SecretKeyName == "" {
			return fmt.Errorf("Anthropic provider requires either api_key or secret_key_name")
		}
		if config.BaseURL == "" {
			return fmt.Errorf("Anthropic provider requires base_url")
		}
		return nil
	}

	// Local AI validator
	c.validators["ollama"] = func(config *ProviderConfig) error {
		if config.BaseURL == "" {
			return fmt.Errorf("Local AI provider requires base_url")
		}
		if config.MaxConcurrency > 5 {
			return fmt.Errorf("Local AI provider concurrency should not exceed 5 for optimal performance")
		}
		return nil
	}

	// Generic validator
	c.validators["generic"] = func(config *ProviderConfig) error {
		if config.Name == "" {
			return fmt.Errorf("provider name is required")
		}
		if config.Type == "" {
			return fmt.Errorf("provider type is required")
		}
		if config.Timeouts != nil {
			if config.Timeouts.RequestTimeout < 1*time.Second {
				return fmt.Errorf("request timeout must be at least 1 second")
			}
		}
		if config.RateLimits != nil {
			if config.RateLimits.RequestsPerSecond < 1 {
				return fmt.Errorf("requests per second must be at least 1")
			}
		}
		return nil
	}
}

// SetProviderConfig sets configuration for a provider
func (c *configManagerImpl) SetProviderConfig(providerName string, config *ProviderConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	// Validate configuration
	if err := c.ValidateProviderConfig(config); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}

	// Set provider name if not set
	if config.Name == "" {
		config.Name = providerName
	}

	// Set timestamps
	oldConfig := c.configurations[providerName]
	if oldConfig == nil {
		config.CreatedAt = time.Now()
	} else {
		config.CreatedAt = oldConfig.CreatedAt
	}
	config.UpdatedAt = time.Now()

	// Store configuration
	c.configurations[providerName] = config

	// Record change
	c.recordConfigurationChange(providerName, "update", oldConfig, config, "configuration_updated")

	return nil
}

// GetProviderConfig gets configuration for a provider
func (c *configManagerImpl) GetProviderConfig(providerName string) (*ProviderConfig, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	config, exists := c.configurations[providerName]
	if !exists {
		return nil, fmt.Errorf("configuration for provider %s not found", providerName)
	}

	// Return a copy to prevent external modifications
	configCopy := *config
	return &configCopy, nil
}

// ValidateProviderConfig validates a provider configuration
func (c *configManagerImpl) ValidateProviderConfig(config *ProviderConfig) error {
	// Generic validation
	if validator, exists := c.validators["generic"]; exists {
		if err := validator(config); err != nil {
			return err
		}
	}

	// Provider-specific validation
	if validator, exists := c.validators[config.Name]; exists {
		if err := validator(config); err != nil {
			return err
		}
	}

	return nil
}

// UpdateProviderConfig updates specific fields in a provider configuration
func (c *configManagerImpl) UpdateProviderConfig(providerName string, updates map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	config, exists := c.configurations[providerName]
	if !exists {
		return fmt.Errorf("configuration for provider %s not found", providerName)
	}

	// Create a copy for modification
	oldConfig := *config
	newConfig := *config

	// Apply updates
	for field, value := range updates {
		switch field {
		case "enabled":
			if enabled, ok := value.(bool); ok {
				newConfig.Enabled = enabled
			}
		case "base_url":
			if baseURL, ok := value.(string); ok {
				newConfig.BaseURL = baseURL
			}
		case "max_concurrency":
			if maxConcurrency, ok := value.(int); ok {
				newConfig.MaxConcurrency = maxConcurrency
			}
		case "models":
			if models, ok := value.([]string); ok {
				newConfig.Models = models
			}
		case "capabilities":
			if capabilities, ok := value.([]string); ok {
				newConfig.Capabilities = capabilities
			}
		case "custom_parameters":
			if params, ok := value.(map[string]interface{}); ok {
				if newConfig.CustomParameters == nil {
					newConfig.CustomParameters = make(map[string]interface{})
				}
				for k, v := range params {
					newConfig.CustomParameters[k] = v
				}
			}
		default:
			return fmt.Errorf("unknown configuration field: %s", field)
		}
	}

	newConfig.UpdatedAt = time.Now()

	// Validate updated configuration
	if err := c.ValidateProviderConfig(&newConfig); err != nil {
		return fmt.Errorf("updated configuration validation failed: %v", err)
	}

	// Store updated configuration
	c.configurations[providerName] = &newConfig

	// Record change
	c.recordConfigurationChange(providerName, "update", &oldConfig, &newConfig, "partial_update")

	return nil
}

// GetAllProviderConfigs returns all provider configurations
func (c *configManagerImpl) GetAllProviderConfigs() map[string]*ProviderConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	configs := make(map[string]*ProviderConfig)
	for name, config := range c.configurations {
		// Return copies to prevent external modifications
		configCopy := *config
		configs[name] = &configCopy
	}

	return configs
}

// BackupConfigurations creates a backup of all configurations
func (c *configManagerImpl) BackupConfigurations() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create configuration copies
	configCopies := make(map[string]*ProviderConfig)
	for name, config := range c.configurations {
		configCopy := *config
		configCopies[name] = &configCopy
	}

	backup := &ConfigurationBackup{
		ID:             fmt.Sprintf("backup_%v", time.Now().Unix()),
		Name:           fmt.Sprintf("Backup %s", time.Now().Format("2006-01-02 15:04:05")),
		Description:    "Automated configuration backup",
		Configurations: configCopies,
		Timestamp:      time.Now(),
	}

	// Keep last 10 backups
	if len(c.backups) >= 10 {
		c.backups = c.backups[1:]
	}

	c.backups = append(c.backups, backup)

	return nil
}

// RestoreConfigurations restores configurations from a backup
func (c *configManagerImpl) RestoreConfigurations(backupID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var backup *ConfigurationBackup
	for _, b := range c.backups {
		if b.ID == backupID {
			backup = b
			break
		}
	}

	if backup == nil {
		return fmt.Errorf("backup %s not found", backupID)
	}

	// Store old configurations for change tracking
	oldConfigurations := make(map[string]*ProviderConfig)
	for name, config := range c.configurations {
		configCopy := *config
		oldConfigurations[name] = &configCopy
	}

	// Restore configurations
	c.configurations = make(map[string]*ProviderConfig)
	for name, config := range backup.Configurations {
		configCopy := *config
		c.configurations[name] = &configCopy

		// Record restoration
		oldConfig := oldConfigurations[name]
		c.recordConfigurationChange(name, "restore", oldConfig, &configCopy, fmt.Sprintf("restored_from_backup_%s", backupID))
	}

	return nil
}

func (c *configManagerImpl) recordConfigurationChange(providerName, changeType string, oldConfig, newConfig *ProviderConfig, reason string) {
	change := &ConfigurationChange{
		ID:           fmt.Sprintf("change_%s_%v", providerName, time.Now().Unix()),
		ProviderName: providerName,
		ChangeType:   changeType,
		OldConfig:    oldConfig,
		NewConfig:    newConfig,
		Changes:      c.calculateChanges(oldConfig, newConfig),
		Reason:       reason,
		Timestamp:    time.Now(),
	}

	// Initialize history if needed
	if c.configHistory[providerName] == nil {
		c.configHistory[providerName] = make([]*ConfigurationChange, 0)
	}

	// Keep last 50 changes per provider
	if len(c.configHistory[providerName]) >= 50 {
		c.configHistory[providerName] = c.configHistory[providerName][1:]
	}

	c.configHistory[providerName] = append(c.configHistory[providerName], change)
}

func (c *configManagerImpl) calculateChanges(oldConfig, newConfig *ProviderConfig) map[string]interface{} {
	changes := make(map[string]interface{})

	if oldConfig == nil {
		changes["action"] = "created"
		return changes
	}

	if newConfig == nil {
		changes["action"] = "deleted"
		return changes
	}

	// Compare configurations
	if oldConfig.Enabled != newConfig.Enabled {
		changes["enabled"] = map[string]interface{}{
			"from": oldConfig.Enabled,
			"to":   newConfig.Enabled,
		}
	}

	if oldConfig.BaseURL != newConfig.BaseURL {
		changes["base_url"] = map[string]interface{}{
			"from": oldConfig.BaseURL,
			"to":   newConfig.BaseURL,
		}
	}

	if oldConfig.MaxConcurrency != newConfig.MaxConcurrency {
		changes["max_concurrency"] = map[string]interface{}{
			"from": oldConfig.MaxConcurrency,
			"to":   newConfig.MaxConcurrency,
		}
	}

	// Could add more detailed comparisons for models, capabilities, etc.

	return changes
}

// Additional configuration management methods

// CreateProviderFromTemplate creates a provider configuration from a template
func (c *configManagerImpl) CreateProviderFromTemplate(providerName, templateID string, variables map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	template, exists := c.templates[templateID]
	if !exists {
		return fmt.Errorf("template %s not found", templateID)
	}

	// Create configuration from template
	config := *template.Template
	config.Name = providerName

	// Apply variables
	for varName, value := range variables {
		switch varName {
		case "api_key":
			if apiKey, ok := value.(string); ok {
				config.APIKey = apiKey
			}
		case "secret_key_name":
			if secretKeyName, ok := value.(string); ok {
				config.SecretKeyName = secretKeyName
			}
		case "base_url":
			if baseURL, ok := value.(string); ok {
				config.BaseURL = baseURL
			}
		case "organization":
			if org, ok := value.(string); ok {
				if config.CustomParameters == nil {
					config.CustomParameters = make(map[string]interface{})
				}
				config.CustomParameters["organization"] = org
			}
		}
	}

	// Validate required variables
	for varName, varDef := range template.Variables {
		if varDef.Required {
			if _, exists := variables[varName]; !exists {
				return fmt.Errorf("required variable %s not provided", varName)
			}
		}
	}

	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	// Validate configuration
	if err := c.ValidateProviderConfig(&config); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}

	// Store configuration
	c.configurations[providerName] = &config

	// Record change
	c.recordConfigurationChange(providerName, "create", nil, &config, fmt.Sprintf("created_from_template_%s", templateID))

	return nil
}

// GetConfigurationHistory returns configuration change history for a provider
func (c *configManagerImpl) GetConfigurationHistory(providerName string) ([]*ConfigurationChange, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	history, exists := c.configHistory[providerName]
	if !exists {
		return make([]*ConfigurationChange, 0), nil
	}

	// Return a copy
	historyCopy := make([]*ConfigurationChange, len(history))
	copy(historyCopy, history)

	return historyCopy, nil
}

// GetAvailableTemplates returns all available configuration templates
func (c *configManagerImpl) GetAvailableTemplates() []*ProviderConfigTemplate {
	c.mu.RLock()
	defer c.mu.RUnlock()

	templates := make([]*ProviderConfigTemplate, 0, len(c.templates))
	for _, template := range c.templates {
		templates = append(templates, template)
	}

	return templates
}

// GetConfigurationBackups returns all available configuration backups
func (c *configManagerImpl) GetConfigurationBackups() []*ConfigurationBackup {
	c.mu.RLock()
	defer c.mu.RUnlock()

	backups := make([]*ConfigurationBackup, len(c.backups))
	copy(backups, c.backups)

	return backups
}

// DeleteProviderConfig deletes a provider configuration
func (c *configManagerImpl) DeleteProviderConfig(providerName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	config, exists := c.configurations[providerName]
	if !exists {
		return fmt.Errorf("configuration for provider %s not found", providerName)
	}

	// Record deletion
	c.recordConfigurationChange(providerName, "delete", config, nil, "configuration_deleted")

	// Delete configuration
	delete(c.configurations, providerName)

	return nil
}
