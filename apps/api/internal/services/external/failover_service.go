package external

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// failoverImpl implements ProviderFailover interface (Task 7.4)
type failoverImpl struct {
	mu               sync.RWMutex
	registry         ProviderRegistry
	loadBalancer     ProviderLoadBalancer
	failoverRules    map[string]*FailoverRule
	activeFailovers  map[string]*ActiveFailover
	failoverHistory  []*FailoverEvent
	providerHealth   map[string]bool
	monitoringActive bool
	stopMonitoring   chan bool
}

// NewProviderFailover creates a new provider failover service
func NewProviderFailover(registry ProviderRegistry, loadBalancer ProviderLoadBalancer) ProviderFailover {
	failover := &failoverImpl{
		registry:         registry,
		loadBalancer:     loadBalancer,
		failoverRules:    make(map[string]*FailoverRule),
		activeFailovers:  make(map[string]*ActiveFailover),
		failoverHistory:  make([]*FailoverEvent, 0),
		providerHealth:   make(map[string]bool),
		monitoringActive: false,
		stopMonitoring:   make(chan bool, 1),
	}

	// Initialize default failover rules
	failover.initializeDefaultRules()

	// Start monitoring
	go failover.startHealthMonitoring()

	return failover
}

func (f *failoverImpl) initializeDefaultRules() {
	defaultRules := []*FailoverRule{
		{
			ID:              "openai_failover",
			Name:            "OpenAI Failover to Ollama",
			PrimaryProvider: "openai",
			BackupProviders: []string{"ollama", "anthropic_claude"},
			TriggerConditions: []*TriggerCondition{
				{
					Type:      "error_rate",
					Threshold: 50.0, // 50% error rate
					Duration:  2 * time.Minute,
					Operator:  ">=",
				},
				{
					Type:      "response_time",
					Threshold: 10000, // 10 seconds
					Duration:  1 * time.Minute,
					Operator:  ">=",
				},
			},
			FailoverStrategy: "immediate",
			Enabled:          true,
			Priority:         1,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			ID:              "ollama_failover",
			Name:            "Ollama Failover to OpenAI",
			PrimaryProvider: "ollama",
			BackupProviders: []string{"openai", "huggingface_inference"},
			TriggerConditions: []*TriggerCondition{
				{
					Type:      "availability",
					Threshold: false,
					Duration:  30 * time.Second,
					Operator:  "==",
				},
				{
					Type:      "error_rate",
					Threshold: 75.0, // 75% error rate
					Duration:  1 * time.Minute,
					Operator:  ">=",
				},
			},
			FailoverStrategy: "gradual",
			Enabled:          true,
			Priority:         2,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	for _, rule := range defaultRules {
		f.failoverRules[rule.ID] = rule
	}
}

// ExecuteWithFailover executes a request with automatic failover support
func (f *failoverImpl) ExecuteWithFailover(ctx context.Context, request *ProviderRequest, primaryProvider string) (*ProviderResponse, error) {
	f.mu.RLock()

	// Check if there's an active failover for this provider
	if activeFailover, exists := f.activeFailovers[primaryProvider]; exists {
		primaryProvider = activeFailover.CurrentProvider
	}

	f.mu.RUnlock()

	// Try primary provider
	provider, err := f.registry.GetProvider(primaryProvider)
	if err == nil && provider.IsHealthy(ctx) {
		response, err := provider.ProcessRequest(ctx, request)
		if err == nil {
			f.recordSuccessfulRequest(primaryProvider)
			return response, nil
		}
		f.recordFailedRequest(primaryProvider, err)
	}

	// Primary failed, try failover
	return f.executeFailover(ctx, request, primaryProvider)
}

func (f *failoverImpl) executeFailover(ctx context.Context, request *ProviderRequest, primaryProvider string) (*ProviderResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Find applicable failover rules
	var applicableRule *FailoverRule
	for _, rule := range f.failoverRules {
		if rule.PrimaryProvider == primaryProvider && rule.Enabled {
			if applicableRule == nil || rule.Priority < applicableRule.Priority {
				applicableRule = rule
			}
		}
	}

	if applicableRule == nil {
		return nil, fmt.Errorf("no failover rule found for provider %s", primaryProvider)
	}

	// Try backup providers in order
	for _, backupProvider := range applicableRule.BackupProviders {
		provider, err := f.registry.GetProvider(backupProvider)
		if err != nil {
			continue
		}

		if !provider.IsHealthy(ctx) {
			continue
		}

		response, err := provider.ProcessRequest(ctx, request)
		if err == nil {
			// Successful failover
			f.activateFailover(primaryProvider, backupProvider, applicableRule, "automatic_failover")
			f.recordFailoverEvent(primaryProvider, backupProvider, "automatic_failover", true)
			return response, nil
		}
	}

	// All providers failed
	f.recordFailoverEvent(primaryProvider, "", "all_providers_failed", false)
	return nil, fmt.Errorf("all providers failed for request %s", request.ID)
}

func (f *failoverImpl) activateFailover(primaryProvider, currentProvider string, rule *FailoverRule, reason string) {
	failoverID := fmt.Sprintf("failover_%s_%v", primaryProvider, time.Now().Unix())

	activeFailover := &ActiveFailover{
		ID:               failoverID,
		PrimaryProvider:  primaryProvider,
		CurrentProvider:  currentProvider,
		FailoverReason:   reason,
		StartTime:        time.Now(),
		ExpectedDuration: 10 * time.Minute, // Default duration
		Status:           "active",
	}

	f.activeFailovers[primaryProvider] = activeFailover
}

func (f *failoverImpl) recordFailoverEvent(fromProvider, toProvider, reason string, success bool) {
	event := &FailoverEvent{
		ID:               fmt.Sprintf("event_%s_%v", fromProvider, time.Now().Unix()),
		FromProvider:     fromProvider,
		ToProvider:       toProvider,
		Reason:           reason,
		TriggerCondition: reason,
		Duration:         0, // Will be updated when failover ends
		Success:          success,
		Impact:           f.calculateImpact(fromProvider, toProvider),
		Timestamp:        time.Now(),
	}

	// Keep last 1000 events
	if len(f.failoverHistory) >= 1000 {
		f.failoverHistory = f.failoverHistory[1:]
	}

	f.failoverHistory = append(f.failoverHistory, event)
}

func (f *failoverImpl) calculateImpact(fromProvider, toProvider string) string {
	if toProvider == "" {
		return "service_unavailable"
	}

	// Calculate impact based on provider capabilities and performance
	from, err1 := f.registry.GetProvider(fromProvider)
	to, err2 := f.registry.GetProvider(toProvider)

	if err1 != nil || err2 != nil {
		return "unknown"
	}

	fromCaps := from.GetCapabilities()
	toCaps := to.GetCapabilities()

	// Simple capability comparison
	if len(fromCaps) > len(toCaps) {
		return "reduced_functionality"
	} else if len(fromCaps) < len(toCaps) {
		return "enhanced_functionality"
	}

	return "similar_functionality"
}

func (f *failoverImpl) recordSuccessfulRequest(providerName string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Update provider health
	f.providerHealth[providerName] = true

	// Check if we should recover from failover
	if activeFailover, exists := f.activeFailovers[providerName]; exists && activeFailover.CurrentProvider != providerName {
		// Primary provider is healthy again, consider recovery
		if f.shouldRecover(providerName) {
			f.recoverFromFailover(providerName)
		}
	}
}

func (f *failoverImpl) recordFailedRequest(providerName string, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.providerHealth[providerName] = false

	// Check if we should trigger failover
	f.evaluateFailoverConditions(providerName)
}

func (f *failoverImpl) shouldRecover(providerName string) bool {
	// Simple recovery logic - could be enhanced with more sophisticated conditions
	return f.providerHealth[providerName] && time.Since(time.Now()) > 5*time.Minute
}

func (f *failoverImpl) recoverFromFailover(providerName string) {
	if activeFailover, exists := f.activeFailovers[providerName]; exists {
		// Calculate duration
		duration := time.Since(activeFailover.StartTime)

		// Update failover history
		for i := len(f.failoverHistory) - 1; i >= 0; i-- {
			if f.failoverHistory[i].FromProvider == providerName && f.failoverHistory[i].Duration == 0 {
				f.failoverHistory[i].Duration = duration
				break
			}
		}

		// Record recovery event
		f.recordFailoverEvent(activeFailover.CurrentProvider, providerName, "automatic_recovery", true)

		// Remove active failover
		delete(f.activeFailovers, providerName)
	}
}

func (f *failoverImpl) evaluateFailoverConditions(providerName string) {
	// Find rules for this provider
	for _, rule := range f.failoverRules {
		if rule.PrimaryProvider == providerName && rule.Enabled {
			if f.checkTriggerConditions(rule, providerName) {
				// Trigger failover
				if len(rule.BackupProviders) > 0 {
					f.activateFailover(providerName, rule.BackupProviders[0], rule, "condition_triggered")
					f.recordFailoverEvent(providerName, rule.BackupProviders[0], "condition_triggered", true)
				}
			}
		}
	}
}

func (f *failoverImpl) checkTriggerConditions(rule *FailoverRule, providerName string) bool {
	// Simplified condition checking - in real implementation would track metrics over time
	for _, condition := range rule.TriggerConditions {
		switch condition.Type {
		case "availability":
			if !f.providerHealth[providerName] {
				return true
			}
		case "error_rate":
			// Would check actual error rate metrics
			if !f.providerHealth[providerName] {
				return true
			}
		case "response_time":
			// Would check actual response time metrics
			provider, err := f.registry.GetProvider(providerName)
			if err == nil {
				status := provider.GetStatus(context.Background())
				if status != nil && status.ResponseTime > time.Duration(condition.Threshold.(int))*time.Millisecond {
					return true
				}
			}
		}
	}
	return false
}

// ConfigureFailoverRules configures failover rules
func (f *failoverImpl) ConfigureFailoverRules(rules []*FailoverRule) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Validate rules
	for _, rule := range rules {
		if rule.PrimaryProvider == "" {
			return fmt.Errorf("primary provider cannot be empty")
		}
		if len(rule.BackupProviders) == 0 {
			return fmt.Errorf("backup providers cannot be empty")
		}
		if len(rule.TriggerConditions) == 0 {
			return fmt.Errorf("trigger conditions cannot be empty")
		}
	}

	// Clear existing rules and add new ones
	f.failoverRules = make(map[string]*FailoverRule)
	for _, rule := range rules {
		if rule.ID == "" {
			rule.ID = fmt.Sprintf("rule_%s_%v", rule.PrimaryProvider, time.Now().Unix())
		}
		rule.UpdatedAt = time.Now()
		f.failoverRules[rule.ID] = rule
	}

	return nil
}

// GetFailoverStatus returns current failover status
func (f *failoverImpl) GetFailoverStatus() *FailoverStatus {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Create copies to avoid concurrent access
	activeFailovers := make(map[string]*ActiveFailover)
	for k, v := range f.activeFailovers {
		activeFailovers[k] = v
	}

	failoverHistory := make([]*FailoverEvent, 0, len(f.failoverHistory))
	// Return last 50 events
	startIndex := 0
	if len(f.failoverHistory) > 50 {
		startIndex = len(f.failoverHistory) - 50
	}
	failoverHistory = append(failoverHistory, f.failoverHistory[startIndex:]...)

	providerHealth := make(map[string]bool)
	for k, v := range f.providerHealth {
		providerHealth[k] = v
	}

	return &FailoverStatus{
		ActiveFailovers:  activeFailovers,
		FailoverHistory:  failoverHistory,
		ProviderHealth:   providerHealth,
		MonitoringActive: f.monitoringActive,
		Timestamp:        time.Now(),
	}
}

// TriggerManualFailover triggers a manual failover from one provider to another
func (f *failoverImpl) TriggerManualFailover(fromProvider, toProvider string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Validate providers exist and are available
	_, err := f.registry.GetProvider(fromProvider)
	if err != nil {
		return fmt.Errorf("source provider %s not found: %v", fromProvider, err)
	}

	to, err := f.registry.GetProvider(toProvider)
	if err != nil {
		return fmt.Errorf("target provider %s not found: %v", toProvider, err)
	}

	if !to.IsHealthy(context.Background()) {
		return fmt.Errorf("target provider %s is not healthy", toProvider)
	}

	// Create a manual failover rule
	manualRule := &FailoverRule{
		ID:              fmt.Sprintf("manual_%s_to_%s_%v", fromProvider, toProvider, time.Now().Unix()),
		Name:            fmt.Sprintf("Manual failover from %s to %s", fromProvider, toProvider),
		PrimaryProvider: fromProvider,
		BackupProviders: []string{toProvider},
		TriggerConditions: []*TriggerCondition{
			{
				Type:      "manual",
				Threshold: true,
				Duration:  0,
				Operator:  "==",
			},
		},
		FailoverStrategy: "immediate",
		Enabled:          true,
		Priority:         0, // Highest priority
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Activate failover
	f.activateFailover(fromProvider, toProvider, manualRule, "manual_failover")
	f.recordFailoverEvent(fromProvider, toProvider, "manual_failover", true)

	return nil
}

// GetFailoverHistory returns failover history
func (f *failoverImpl) GetFailoverHistory() []*FailoverEvent {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Return a copy of the history
	history := make([]*FailoverEvent, len(f.failoverHistory))
	copy(history, f.failoverHistory)

	return history
}

// startHealthMonitoring starts continuous health monitoring for failover decisions
func (f *failoverImpl) startHealthMonitoring() {
	f.mu.Lock()
	f.monitoringActive = true
	f.mu.Unlock()

	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			f.performHealthChecks()
		case <-f.stopMonitoring:
			f.mu.Lock()
			f.monitoringActive = false
			f.mu.Unlock()
			return
		}
	}
}

func (f *failoverImpl) performHealthChecks() {
	providers := f.registry.GetAllProviders()

	for _, provider := range providers {
		providerName := provider.GetName()
		isHealthy := provider.IsHealthy(context.Background())

		f.mu.Lock()
		previousHealth := f.providerHealth[providerName]
		f.providerHealth[providerName] = isHealthy

		// Check for health state changes
		if previousHealth != isHealthy {
			if isHealthy {
				// Provider recovered
				if f.shouldRecover(providerName) {
					f.recoverFromFailover(providerName)
				}
			} else {
				// Provider became unhealthy
				f.evaluateFailoverConditions(providerName)
			}
		}

		f.mu.Unlock()
	}
}

// StopMonitoring stops the health monitoring
func (f *failoverImpl) StopMonitoring() {
	if f.monitoringActive {
		f.stopMonitoring <- true
	}
}

// Additional helper methods for failover management

// GetFailoverRules returns all configured failover rules
func (f *failoverImpl) GetFailoverRules() []*FailoverRule {
	f.mu.RLock()
	defer f.mu.RUnlock()

	rules := make([]*FailoverRule, 0, len(f.failoverRules))
	for _, rule := range f.failoverRules {
		rules = append(rules, rule)
	}

	return rules
}

// UpdateFailoverRule updates a specific failover rule
func (f *failoverImpl) UpdateFailoverRule(ruleID string, updates map[string]interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	rule, exists := f.failoverRules[ruleID]
	if !exists {
		return fmt.Errorf("failover rule %s not found", ruleID)
	}

	// Apply updates
	if enabled, ok := updates["enabled"].(bool); ok {
		rule.Enabled = enabled
	}
	if priority, ok := updates["priority"].(int); ok {
		rule.Priority = priority
	}
	if strategy, ok := updates["failover_strategy"].(string); ok {
		rule.FailoverStrategy = strategy
	}

	rule.UpdatedAt = time.Now()

	return nil
}

// DeleteFailoverRule deletes a failover rule
func (f *failoverImpl) DeleteFailoverRule(ruleID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.failoverRules[ruleID]; !exists {
		return fmt.Errorf("failover rule %s not found", ruleID)
	}

	delete(f.failoverRules, ruleID)
	return nil
}
