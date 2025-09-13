package external

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// controlServiceImpl implements CostController interface (Task 6.5)
type controlServiceImpl struct {
	mu               sync.RWMutex
	controls         map[string]*AutomatedControls
	controlStatuses  map[string]*ControlStatus
	controlHistory   []*ControlAction
	budgetManager    BudgetManager
	costTracker      CostTracker
	monitoringActive bool
	stopChannel      chan bool
}

// NewControlService creates a new automated cost control service
func NewControlService(budgetManager BudgetManager, costTracker CostTracker) CostController {
	service := &controlServiceImpl{
		controls:         make(map[string]*AutomatedControls),
		controlStatuses:  make(map[string]*ControlStatus),
		controlHistory:   make([]*ControlAction, 0),
		budgetManager:    budgetManager,
		costTracker:      costTracker,
		monitoringActive: false,
		stopChannel:      make(chan bool),
	}

	// Initialize default automated controls
	service.initializeDefaultControls()

	return service
}

// initializeDefaultControls creates default automated cost control rules
func (s *controlServiceImpl) initializeDefaultControls() {
	defaultControls := []*AutomatedControls{
		{
			ServiceName: "openai",
			Enabled:     true,
			ThresholdRules: []ThresholdRule{
				{
					Threshold: 75.0,
					Action:    "alert_management",
					Enabled:   true,
				},
				{
					Threshold: 90.0,
					Action:    "rate_limit_moderate",
					Enabled:   true,
				},
				{
					Threshold: 95.0,
					Action:    "rate_limit_strict",
					Enabled:   true,
				},
				{
					Threshold: 100.0,
					Action:    "suspend_service",
					Enabled:   false, // Disabled by default for safety
				},
			},
			Actions: []AutomatedAction{
				{
					Type: "rate_limit",
					Parameters: map[string]interface{}{
						"requests_per_minute": 10,
						"burst_limit":         5,
					},
					Timestamp: time.Now(),
				},
			},
			CreatedAt: time.Now(),
		},
		{
			ServiceName: "ollama",
			Enabled:     true,
			ThresholdRules: []ThresholdRule{
				{
					Threshold: 80.0,
					Action:    "alert_management",
					Enabled:   true,
				},
				{
					Threshold: 95.0,
					Action:    "rate_limit_moderate",
					Enabled:   true,
				},
				{
					Threshold: 100.0,
					Action:    "rate_limit_strict",
					Enabled:   true,
				},
			},
			Actions: []AutomatedAction{
				{
					Type: "resource_limit",
					Parameters: map[string]interface{}{
						"max_concurrent_requests": 3,
						"timeout_seconds":         30,
					},
					Timestamp: time.Now(),
				},
			},
			CreatedAt: time.Now(),
		},
	}

	for _, control := range defaultControls {
		s.controls[control.ServiceName] = control
		s.controlStatuses[control.ServiceName] = &ControlStatus{
			ServiceName: control.ServiceName,
			Enabled:     control.Enabled,
			Status:      "active",
			LastAction:  time.Now(),
		}
	}
}

// EnableAutomatedControls enables automated cost controls for a service (Task 6.5)
func (s *controlServiceImpl) EnableAutomatedControls(ctx context.Context, controls *AutomatedControls) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if controls.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	controls.Enabled = true
	s.controls[controls.ServiceName] = controls

	// Update control status
	s.controlStatuses[controls.ServiceName] = &ControlStatus{
		ServiceName: controls.ServiceName,
		Enabled:     true,
		Status:      "active",
		LastAction:  time.Now(),
	}

	// Log control action
	s.logControlAction(&ControlAction{
		ID:          fmt.Sprintf("ctrl_%s_%v", controls.ServiceName, time.Now().Unix()),
		ServiceName: controls.ServiceName,
		Action:      "enable_controls",
		Reason:      "Automated cost controls enabled",
		Result:      "success",
		Timestamp:   time.Now(),
	})

	// Start monitoring if not already active
	if !s.monitoringActive {
		go s.startAutomatedMonitoring()
	}

	return nil
}

// DisableAutomatedControls disables automated cost controls for a service (Task 6.5)
func (s *controlServiceImpl) DisableAutomatedControls(ctx context.Context, serviceName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	control, exists := s.controls[serviceName]
	if !exists {
		return fmt.Errorf("no automated controls found for service %s", serviceName)
	}

	control.Enabled = false

	// Update control status
	if status, exists := s.controlStatuses[serviceName]; exists {
		status.Enabled = false
		status.Status = "disabled"
		status.LastAction = time.Now()
	}

	// Log control action
	s.logControlAction(&ControlAction{
		ID:          fmt.Sprintf("ctrl_%s_%v", serviceName, time.Now().Unix()),
		ServiceName: serviceName,
		Action:      "disable_controls",
		Reason:      "Automated cost controls disabled",
		Result:      "success",
		Timestamp:   time.Now(),
	})

	return nil
}

// GetControlStatus returns the status of automated controls (Task 6.5)
func (s *controlServiceImpl) GetControlStatus(ctx context.Context) ([]*ControlStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	statuses := make([]*ControlStatus, 0, len(s.controlStatuses))
	for _, status := range s.controlStatuses {
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// ExecuteEmergencyShutdown performs emergency shutdown of services (Task 6.5)
func (s *controlServiceImpl) ExecuteEmergencyShutdown(ctx context.Context, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	emergencyActions := make([]*ControlAction, 0)

	// Shutdown all services with automated controls
	for serviceName, control := range s.controls {
		if control.Enabled {
			// Execute emergency shutdown
			action := &ControlAction{
				ID:          fmt.Sprintf("emergency_%s_%v", serviceName, time.Now().Unix()),
				ServiceName: serviceName,
				Action:      "emergency_shutdown",
				Reason:      fmt.Sprintf("Emergency shutdown: %s", reason),
				Result:      "executed",
				Timestamp:   time.Now(),
			}

			emergencyActions = append(emergencyActions, action)
			s.logControlAction(action)

			// Update control status
			if status, exists := s.controlStatuses[serviceName]; exists {
				status.Status = "emergency_shutdown"
				status.LastAction = time.Now()
			}

			// Execute actual shutdown logic here
			// This would integrate with service management systems
			s.executeShutdownAction(serviceName, reason)
		}
	}

	if len(emergencyActions) == 0 {
		return fmt.Errorf("no services with automated controls found for emergency shutdown")
	}

	return nil
}

func (s *controlServiceImpl) executeShutdownAction(serviceName, reason string) {
	// In a real implementation, this would:
	// 1. Disable API endpoints
	// 2. Stop accepting new requests
	// 3. Complete existing requests
	// 4. Update load balancer configurations
	// 5. Send notifications to relevant teams

	// For demo purposes, we'll just log the action
	fmt.Printf("EMERGENCY SHUTDOWN: Service %s shutdown due to: %s\n", serviceName, reason)
}

// GetControlHistory returns the history of control actions (Task 6.5)
func (s *controlServiceImpl) GetControlHistory(ctx context.Context, timeRange TimeRange) ([]*ControlAction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filteredHistory := make([]*ControlAction, 0)
	for _, action := range s.controlHistory {
		if action.Timestamp.After(timeRange.Start) && action.Timestamp.Before(timeRange.End) {
			filteredHistory = append(filteredHistory, action)
		}
	}

	return filteredHistory, nil
}

// startAutomatedMonitoring starts the automated cost control monitoring
func (s *controlServiceImpl) startAutomatedMonitoring() {
	s.mu.Lock()
	s.monitoringActive = true
	s.mu.Unlock()

	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.performAutomatedChecks()
		case <-s.stopChannel:
			s.mu.Lock()
			s.monitoringActive = false
			s.mu.Unlock()
			return
		}
	}
}

func (s *controlServiceImpl) performAutomatedChecks() {
	s.mu.RLock()
	controls := make(map[string]*AutomatedControls)
	for k, v := range s.controls {
		if v.Enabled {
			controls[k] = v
		}
	}
	s.mu.RUnlock()

	for serviceName, control := range controls {
		s.checkServiceThresholds(serviceName, control)
	}
}

func (s *controlServiceImpl) checkServiceThresholds(serviceName string, control *AutomatedControls) {
	// Get current budget status
	budgetStatus, err := s.budgetManager.CheckBudgetStatus(context.Background(), serviceName)
	if err != nil {
		return
	}

	// Check each threshold rule
	for _, rule := range control.ThresholdRules {
		if !rule.Enabled {
			continue
		}

		if budgetStatus.UtilizationPct >= rule.Threshold {
			s.executeControlAction(serviceName, rule, budgetStatus)
		}
	}
}

func (s *controlServiceImpl) executeControlAction(serviceName string, rule ThresholdRule, budgetStatus *BudgetStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	actionID := fmt.Sprintf("auto_%s_%s_%v", serviceName, rule.Action, time.Now().Unix())

	// Check if we've already executed this action recently (prevent spam)
	if s.hasRecentAction(serviceName, rule.Action, 5*time.Minute) {
		return
	}

	var result string
	var actualAction string

	switch rule.Action {
	case "alert_management":
		result = s.sendAlert(serviceName, rule.Threshold, budgetStatus)
		actualAction = "alert_sent"
	case "rate_limit_moderate":
		result = s.applyRateLimit(serviceName, "moderate")
		actualAction = "rate_limit_applied"
	case "rate_limit_strict":
		result = s.applyRateLimit(serviceName, "strict")
		actualAction = "rate_limit_strict_applied"
	case "suspend_service":
		result = s.suspendService(serviceName)
		actualAction = "service_suspended"
	default:
		result = "unknown_action"
		actualAction = rule.Action
	}

	// Log the control action
	action := &ControlAction{
		ID:          actionID,
		ServiceName: serviceName,
		Action:      actualAction,
		Reason:      fmt.Sprintf("Threshold %.1f%% exceeded (current: %.1f%%)", rule.Threshold, budgetStatus.UtilizationPct),
		Result:      result,
		Timestamp:   time.Now(),
	}

	s.logControlAction(action)

	// Update control status
	if status, exists := s.controlStatuses[serviceName]; exists {
		status.Status = fmt.Sprintf("action_taken_%s", actualAction)
		status.LastAction = time.Now()
	}
}

func (s *controlServiceImpl) hasRecentAction(serviceName, action string, duration time.Duration) bool {
	cutoff := time.Now().Add(-duration)
	for _, historyAction := range s.controlHistory {
		if historyAction.ServiceName == serviceName &&
			historyAction.Action == action &&
			historyAction.Timestamp.After(cutoff) {
			return true
		}
	}
	return false
}

func (s *controlServiceImpl) sendAlert(serviceName string, threshold float64, budgetStatus *BudgetStatus) string {
	// In a real implementation, this would send alerts via:
	// - Email
	// - Slack/Teams
	// - PagerDuty
	// - SMS

	alertMessage := fmt.Sprintf(
		"COST ALERT: Service %s has exceeded %.1f%% of budget (Current: %.1f%%, $%.2f/$%.2f)",
		serviceName, threshold, budgetStatus.UtilizationPct,
		budgetStatus.CurrentSpend, budgetStatus.MonthlyLimit,
	)

	fmt.Printf("ALERT SENT: %s\n", alertMessage)
	return "alert_sent_successfully"
}

func (s *controlServiceImpl) applyRateLimit(serviceName, severity string) string {
	// In a real implementation, this would:
	// 1. Update API gateway configuration
	// 2. Apply rate limiting rules
	// 3. Update load balancer settings

	var limitConfig map[string]interface{}
	switch severity {
	case "moderate":
		limitConfig = map[string]interface{}{
			"requests_per_minute": 30,
			"burst_limit":         10,
		}
	case "strict":
		limitConfig = map[string]interface{}{
			"requests_per_minute": 10,
			"burst_limit":         3,
		}
	}

	fmt.Printf("RATE LIMIT APPLIED: Service %s - %s limits: %v\n", serviceName, severity, limitConfig)
	return fmt.Sprintf("rate_limit_%s_applied", severity)
}

func (s *controlServiceImpl) suspendService(serviceName string) string {
	// In a real implementation, this would:
	// 1. Gracefully stop accepting new requests
	// 2. Allow existing requests to complete
	// 3. Update service discovery
	// 4. Send critical alerts

	fmt.Printf("SERVICE SUSPENDED: %s has been suspended due to budget overrun\n", serviceName)
	return "service_suspended_successfully"
}

func (s *controlServiceImpl) logControlAction(action *ControlAction) {
	s.controlHistory = append(s.controlHistory, action)

	// Keep only last 1000 actions to prevent memory bloat
	if len(s.controlHistory) > 1000 {
		s.controlHistory = s.controlHistory[len(s.controlHistory)-1000:]
	}
}

// Additional control management methods
func (s *controlServiceImpl) UpdateControlRules(ctx context.Context, serviceName string, rules []ThresholdRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	control, exists := s.controls[serviceName]
	if !exists {
		return fmt.Errorf("no automated controls found for service %s", serviceName)
	}

	control.ThresholdRules = rules

	// Log the update
	s.logControlAction(&ControlAction{
		ID:          fmt.Sprintf("update_%s_%v", serviceName, time.Now().Unix()),
		ServiceName: serviceName,
		Action:      "update_rules",
		Reason:      "Control rules updated",
		Result:      "success",
		Timestamp:   time.Now(),
	})

	return nil
}

func (s *controlServiceImpl) GetControlConfiguration(ctx context.Context, serviceName string) (*AutomatedControls, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	control, exists := s.controls[serviceName]
	if !exists {
		return nil, fmt.Errorf("no automated controls found for service %s", serviceName)
	}

	return control, nil
}

func (s *controlServiceImpl) TestControlAction(ctx context.Context, serviceName, actionType string) (*ControlActionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Simulate control action without actually executing it
	testAction := &ControlAction{
		ID:          fmt.Sprintf("test_%s_%s_%v", serviceName, actionType, time.Now().Unix()),
		ServiceName: serviceName,
		Action:      fmt.Sprintf("test_%s", actionType),
		Reason:      "Control action test",
		Result:      "test_successful",
		Timestamp:   time.Now(),
	}

	result := &ControlActionResult{
		Action:    testAction,
		Executed:  false,
		TestMode:  true,
		Message:   fmt.Sprintf("Test of %s action for service %s completed successfully", actionType, serviceName),
		Timestamp: time.Now(),
	}

	return result, nil
}

func (s *controlServiceImpl) StopMonitoring() {
	if s.monitoringActive {
		s.stopChannel <- true
	}
}

// Control reporting methods
func (s *controlServiceImpl) GetControlMetrics(ctx context.Context, timeRange TimeRange) (*ControlMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter actions within time range
	relevantActions := make([]*ControlAction, 0)
	for _, action := range s.controlHistory {
		if action.Timestamp.After(timeRange.Start) && action.Timestamp.Before(timeRange.End) {
			relevantActions = append(relevantActions, action)
		}
	}

	// Calculate metrics
	actionCounts := make(map[string]int)
	serviceCounts := make(map[string]int)
	successfulActions := 0

	for _, action := range relevantActions {
		actionCounts[action.Action]++
		serviceCounts[action.ServiceName]++
		if action.Result == "success" || action.Result == "executed" ||
			action.Result == "alert_sent_successfully" || action.Result == "rate_limit_moderate_applied" {
			successfulActions++
		}
	}

	successRate := 0.0
	if len(relevantActions) > 0 {
		successRate = float64(successfulActions) / float64(len(relevantActions)) * 100
	}

	return &ControlMetrics{
		TimeRange:         timeRange,
		TotalActions:      len(relevantActions),
		SuccessfulActions: successfulActions,
		SuccessRate:       successRate,
		ActionBreakdown:   actionCounts,
		ServiceBreakdown:  serviceCounts,
		ActiveControls:    len(s.getActiveControls()),
		Timestamp:         time.Now(),
	}, nil
}

func (s *controlServiceImpl) getActiveControls() map[string]*AutomatedControls {
	activeControls := make(map[string]*AutomatedControls)
	for name, control := range s.controls {
		if control.Enabled {
			activeControls[name] = control
		}
	}
	return activeControls
}

// Additional data structures for control management
type ControlActionResult struct {
	Action    *ControlAction `json:"action"`
	Executed  bool           `json:"executed"`
	TestMode  bool           `json:"test_mode"`
	Message   string         `json:"message"`
	Timestamp time.Time      `json:"timestamp"`
}

type ControlMetrics struct {
	TimeRange         TimeRange      `json:"time_range"`
	TotalActions      int            `json:"total_actions"`
	SuccessfulActions int            `json:"successful_actions"`
	SuccessRate       float64        `json:"success_rate"`
	ActionBreakdown   map[string]int `json:"action_breakdown"`
	ServiceBreakdown  map[string]int `json:"service_breakdown"`
	ActiveControls    int            `json:"active_controls"`
	Timestamp         time.Time      `json:"timestamp"`
}
