package external

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pramodksahoo/kubechat/apps/api/internal/services/audit"
)

// RecoveryService provides API failure recovery procedures
type RecoveryService interface {
	// StartRecoveryMonitoring starts monitoring for failed APIs and attempts recovery
	StartRecoveryMonitoring(ctx context.Context) error

	// StopRecoveryMonitoring stops the recovery monitoring
	StopRecoveryMonitoring() error

	// TriggerRecovery manually triggers recovery procedures for a specific API
	TriggerRecovery(ctx context.Context, apiName string) (*RecoveryResult, error)

	// RegisterRecoveryProcedure registers a custom recovery procedure for an API
	RegisterRecoveryProcedure(apiName string, procedure *RecoveryProcedure) error

	// GetRecoveryStatus returns the current recovery status for all APIs
	GetRecoveryStatus() *RecoveryStatus

	// GetRecoveryMetrics returns recovery metrics and statistics
	GetRecoveryMetrics() *RecoveryMetrics
}

// RecoveryProcedure defines the recovery steps for a specific API
type RecoveryProcedure struct {
	APIName       string                 `json:"api_name"`
	Description   string                 `json:"description"`
	MaxAttempts   int                    `json:"max_attempts"`
	InitialDelay  time.Duration          `json:"initial_delay"`
	MaxDelay      time.Duration          `json:"max_delay"`
	BackoffFactor float64                `json:"backoff_factor"`
	Steps         []RecoveryStep         `json:"steps"`
	HealthCheck   func() error           `json:"-"` // Health check function
	OnSuccess     func()                 `json:"-"` // Success callback
	OnFailure     func(error)            `json:"-"` // Failure callback
	Config        map[string]interface{} `json:"config,omitempty"`
	Enabled       bool                   `json:"enabled"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// RecoveryStep represents a single step in the recovery procedure
type RecoveryStep struct {
	Name        string                 `json:"name"`
	Type        RecoveryStepType       `json:"type"`
	Action      func() error           `json:"-"` // Action function to execute
	Timeout     time.Duration          `json:"timeout"`
	RetryCount  int                    `json:"retry_count"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Description string                 `json:"description,omitempty"`
	Required    bool                   `json:"required"` // If true, failure stops recovery
}

// RecoveryStepType defines the type of recovery step
type RecoveryStepType int

const (
	RecoveryStepHealthCheck RecoveryStepType = iota
	RecoveryStepRestart
	RecoveryStepReconnect
	RecoveryStepClearCache
	RecoveryStepResetConfig
	RecoveryStepCustom
)

func (t RecoveryStepType) String() string {
	switch t {
	case RecoveryStepHealthCheck:
		return "health_check"
	case RecoveryStepRestart:
		return "restart"
	case RecoveryStepReconnect:
		return "reconnect"
	case RecoveryStepClearCache:
		return "clear_cache"
	case RecoveryStepResetConfig:
		return "reset_config"
	case RecoveryStepCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// RecoveryResult contains the result of a recovery attempt
type RecoveryResult struct {
	APIName       string        `json:"api_name"`
	Success       bool          `json:"success"`
	AttemptNumber int           `json:"attempt_number"`
	Duration      time.Duration `json:"duration"`
	StepsExecuted int           `json:"steps_executed"`
	ErrorMessage  string        `json:"error_message,omitempty"`
	ExecutedSteps []string      `json:"executed_steps"`
	Timestamp     time.Time     `json:"timestamp"`
}

// RecoveryStatus contains the current status of all recovery procedures
type RecoveryStatus struct {
	TotalAPIs        int                          `json:"total_apis"`
	HealthyAPIs      int                          `json:"healthy_apis"`
	RecoveringAPIs   int                          `json:"recovering_apis"`
	FailedAPIs       int                          `json:"failed_apis"`
	APIStatus        map[string]*APIRecoveryState `json:"api_status"`
	LastUpdated      time.Time                    `json:"last_updated"`
	MonitoringActive bool                         `json:"monitoring_active"`
}

// APIRecoveryState contains the recovery state for a specific API
type APIRecoveryState struct {
	APIName             string    `json:"api_name"`
	Status              APIStatus `json:"status"`
	LastHealthCheck     time.Time `json:"last_health_check"`
	LastRecoveryTry     time.Time `json:"last_recovery_try,omitempty"`
	RecoveryAttempts    int       `json:"recovery_attempts"`
	MaxAttempts         int       `json:"max_attempts"`
	NextRecoveryTime    time.Time `json:"next_recovery_time,omitempty"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
	LastError           string    `json:"last_error,omitempty"`
}

// APIStatus represents the current status of an API
type APIStatus int

const (
	APIStatusHealthy APIStatus = iota
	APIStatusDegraded
	APIStatusRecovering
	APIStatusFailed
	APIStatusUnknown
)

func (s APIStatus) String() string {
	switch s {
	case APIStatusHealthy:
		return "healthy"
	case APIStatusDegraded:
		return "degraded"
	case APIStatusRecovering:
		return "recovering"
	case APIStatusFailed:
		return "failed"
	case APIStatusUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// RecoveryMetrics contains metrics about recovery operations
type RecoveryMetrics struct {
	TotalRecoveryAttempts int64                          `json:"total_recovery_attempts"`
	SuccessfulRecoveries  int64                          `json:"successful_recoveries"`
	FailedRecoveries      int64                          `json:"failed_recoveries"`
	AverageRecoveryTime   time.Duration                  `json:"average_recovery_time"`
	APIMetrics            map[string]*APIRecoveryMetrics `json:"api_metrics"`
	LastRecoveryAttempt   time.Time                      `json:"last_recovery_attempt,omitempty"`
	MonitoringStartTime   time.Time                      `json:"monitoring_start_time,omitempty"`
	TotalDowntime         time.Duration                  `json:"total_downtime"`
	LastUpdated           time.Time                      `json:"last_updated"`
}

// APIRecoveryMetrics contains metrics for a specific API
type APIRecoveryMetrics struct {
	APIName              string        `json:"api_name"`
	RecoveryAttempts     int64         `json:"recovery_attempts"`
	SuccessfulRecoveries int64         `json:"successful_recoveries"`
	FailedRecoveries     int64         `json:"failed_recoveries"`
	AverageRecoveryTime  time.Duration `json:"average_recovery_time"`
	TotalDowntime        time.Duration `json:"total_downtime"`
	LastRecovery         time.Time     `json:"last_recovery,omitempty"`
	ConsecutiveFailures  int           `json:"consecutive_failures"`
	SuccessRate          float64       `json:"success_rate"`
}

// recoveryServiceImpl implements RecoveryService
type recoveryServiceImpl struct {
	procedures map[string]*RecoveryProcedure
	status     *RecoveryStatus
	metrics    *RecoveryMetrics
	auditSvc   audit.Service
	monitoring bool
	stopChan   chan struct{}
	mu         sync.RWMutex
}

// NewRecoveryService creates a new recovery service
func NewRecoveryService(auditSvc audit.Service) RecoveryService {
	return &recoveryServiceImpl{
		procedures: make(map[string]*RecoveryProcedure),
		status: &RecoveryStatus{
			APIStatus:   make(map[string]*APIRecoveryState),
			LastUpdated: time.Now(),
		},
		metrics: &RecoveryMetrics{
			APIMetrics:  make(map[string]*APIRecoveryMetrics),
			LastUpdated: time.Now(),
		},
		auditSvc: auditSvc,
		stopChan: make(chan struct{}),
	}
}

// StartRecoveryMonitoring starts monitoring for failed APIs and attempts recovery
func (s *recoveryServiceImpl) StartRecoveryMonitoring(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.monitoring {
		return fmt.Errorf("recovery monitoring is already running")
	}

	s.monitoring = true
	s.metrics.MonitoringStartTime = time.Now()
	s.status.MonitoringActive = true

	// Start monitoring goroutine
	go s.monitoringLoop(ctx)

	log.Println("API recovery monitoring started")
	return nil
}

// StopRecoveryMonitoring stops the recovery monitoring
func (s *recoveryServiceImpl) StopRecoveryMonitoring() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.monitoring {
		return fmt.Errorf("recovery monitoring is not running")
	}

	s.monitoring = false
	s.status.MonitoringActive = false
	close(s.stopChan)
	s.stopChan = make(chan struct{}) // Create new channel for next start

	log.Println("API recovery monitoring stopped")
	return nil
}

// monitoringLoop runs the main monitoring loop
func (s *recoveryServiceImpl) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAndRecoverAPIs(ctx)
		}
	}
}

// checkAndRecoverAPIs checks all APIs and triggers recovery if needed
func (s *recoveryServiceImpl) checkAndRecoverAPIs(ctx context.Context) {
	s.mu.RLock()
	procedures := make(map[string]*RecoveryProcedure)
	for name, proc := range s.procedures {
		if proc.Enabled {
			procedures[name] = proc
		}
	}
	s.mu.RUnlock()

	for apiName, procedure := range procedures {
		s.checkAPIHealth(ctx, apiName, procedure)
	}

	s.updateStatus()
}

// checkAPIHealth checks the health of a specific API and triggers recovery if needed
func (s *recoveryServiceImpl) checkAPIHealth(ctx context.Context, apiName string, procedure *RecoveryProcedure) {
	apiState := s.getOrCreateAPIState(apiName)

	// Perform health check
	var healthErr error
	if procedure.HealthCheck != nil {
		healthErr = procedure.HealthCheck()
	}

	apiState.LastHealthCheck = time.Now()

	if healthErr == nil {
		// API is healthy
		if apiState.Status != APIStatusHealthy {
			log.Printf("API %s recovered successfully", apiName)
			s.logRecoveryEvent(apiName, "api_recovered", nil)
		}
		apiState.Status = APIStatusHealthy
		apiState.ConsecutiveFailures = 0
		apiState.RecoveryAttempts = 0
		return
	}

	// API is unhealthy
	apiState.ConsecutiveFailures++
	apiState.LastError = healthErr.Error()

	// Check if we should attempt recovery
	if s.shouldAttemptRecovery(apiState, procedure) {
		log.Printf("Triggering automatic recovery for API %s (failure #%d): %v",
			apiName, apiState.ConsecutiveFailures, healthErr)

		go func() {
			result, err := s.executeRecovery(ctx, apiName, procedure)
			if err != nil {
				log.Printf("Auto recovery failed for API %s: %v", apiName, err)
			} else if result.Success {
				log.Printf("Auto recovery succeeded for API %s after %d steps",
					apiName, result.StepsExecuted)
			}
		}()
	}
}

// shouldAttemptRecovery determines if we should attempt recovery for an API
func (s *recoveryServiceImpl) shouldAttemptRecovery(apiState *APIRecoveryState, procedure *RecoveryProcedure) bool {
	// Don't attempt if we've exceeded max attempts
	if apiState.RecoveryAttempts >= procedure.MaxAttempts {
		apiState.Status = APIStatusFailed
		return false
	}

	// Don't attempt if we're currently recovering
	if apiState.Status == APIStatusRecovering {
		return false
	}

	// Check if enough time has passed since last recovery attempt
	if !apiState.LastRecoveryTry.IsZero() {
		minWait := s.calculateBackoffDelay(apiState.RecoveryAttempts, procedure)
		if time.Since(apiState.LastRecoveryTry) < minWait {
			return false
		}
	}

	return true
}

// calculateBackoffDelay calculates the delay before next recovery attempt using exponential backoff
func (s *recoveryServiceImpl) calculateBackoffDelay(attemptNumber int, procedure *RecoveryProcedure) time.Duration {
	if attemptNumber == 0 {
		return procedure.InitialDelay
	}

	delay := float64(procedure.InitialDelay)
	for i := 1; i < attemptNumber; i++ {
		delay *= procedure.BackoffFactor
	}

	if time.Duration(delay) > procedure.MaxDelay {
		return procedure.MaxDelay
	}

	return time.Duration(delay)
}

// TriggerRecovery manually triggers recovery procedures for a specific API
func (s *recoveryServiceImpl) TriggerRecovery(ctx context.Context, apiName string) (*RecoveryResult, error) {
	s.mu.RLock()
	procedure, exists := s.procedures[apiName]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no recovery procedure registered for API: %s", apiName)
	}

	if !procedure.Enabled {
		return nil, fmt.Errorf("recovery procedure for API %s is disabled", apiName)
	}

	return s.executeRecovery(ctx, apiName, procedure)
}

// executeRecovery executes the recovery procedure for an API
func (s *recoveryServiceImpl) executeRecovery(ctx context.Context, apiName string, procedure *RecoveryProcedure) (*RecoveryResult, error) {
	startTime := time.Now()
	apiState := s.getOrCreateAPIState(apiName)

	apiState.Status = APIStatusRecovering
	apiState.LastRecoveryTry = startTime
	apiState.RecoveryAttempts++

	result := &RecoveryResult{
		APIName:       apiName,
		AttemptNumber: apiState.RecoveryAttempts,
		Timestamp:     startTime,
		ExecutedSteps: make([]string, 0),
	}

	s.updateRecoveryMetrics(apiName, true)
	s.logRecoveryEvent(apiName, "recovery_started", nil)

	// Execute recovery steps
	for i, step := range procedure.Steps {
		stepCtx, cancel := context.WithTimeout(ctx, step.Timeout)
		stepErr := s.executeRecoveryStep(stepCtx, &step, result)
		cancel()

		if stepErr != nil {
			if step.Required {
				// Required step failed, abort recovery
				result.Success = false
				result.ErrorMessage = fmt.Sprintf("required step '%s' failed: %v", step.Name, stepErr)
				result.Duration = time.Since(startTime)

				apiState.Status = APIStatusFailed
				s.updateRecoveryMetrics(apiName, false)
				s.logRecoveryEvent(apiName, "recovery_failed", stepErr)

				return result, stepErr
			}

			// Optional step failed, continue
			log.Printf("Optional recovery step '%s' failed for API %s: %v", step.Name, apiName, stepErr)
		}

		result.StepsExecuted = i + 1
	}

	// Final health check
	var finalHealthErr error
	if procedure.HealthCheck != nil {
		finalHealthErr = procedure.HealthCheck()
	}

	result.Duration = time.Since(startTime)
	result.Success = finalHealthErr == nil

	if result.Success {
		apiState.Status = APIStatusHealthy
		apiState.ConsecutiveFailures = 0
		s.updateRecoveryMetrics(apiName, false)
		s.logRecoveryEvent(apiName, "recovery_succeeded", nil)

		if procedure.OnSuccess != nil {
			procedure.OnSuccess()
		}
	} else {
		apiState.Status = APIStatusFailed
		result.ErrorMessage = fmt.Sprintf("recovery completed but health check still fails: %v", finalHealthErr)
		s.updateRecoveryMetrics(apiName, false)
		s.logRecoveryEvent(apiName, "recovery_completed_unhealthy", finalHealthErr)

		if procedure.OnFailure != nil {
			procedure.OnFailure(finalHealthErr)
		}
	}

	return result, nil
}

// executeRecoveryStep executes a single recovery step
func (s *recoveryServiceImpl) executeRecoveryStep(ctx context.Context, step *RecoveryStep, result *RecoveryResult) error {
	result.ExecutedSteps = append(result.ExecutedSteps, step.Name)

	if step.Action == nil {
		return fmt.Errorf("no action defined for step: %s", step.Name)
	}

	var lastErr error
	for attempt := 0; attempt <= step.RetryCount; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		lastErr = step.Action()
		if lastErr == nil {
			return nil // Step succeeded
		}

		if attempt < step.RetryCount {
			// Wait before retry (simple linear backoff)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt+1) * time.Second):
			}
		}
	}

	return fmt.Errorf("step failed after %d attempts: %w", step.RetryCount+1, lastErr)
}

// RegisterRecoveryProcedure registers a custom recovery procedure for an API
func (s *recoveryServiceImpl) RegisterRecoveryProcedure(apiName string, procedure *RecoveryProcedure) error {
	if procedure == nil {
		return fmt.Errorf("procedure cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	procedure.APIName = apiName
	procedure.UpdatedAt = time.Now()
	if procedure.CreatedAt.IsZero() {
		procedure.CreatedAt = time.Now()
	}

	// Set defaults
	if procedure.MaxAttempts == 0 {
		procedure.MaxAttempts = 5
	}
	if procedure.InitialDelay == 0 {
		procedure.InitialDelay = 30 * time.Second
	}
	if procedure.MaxDelay == 0 {
		procedure.MaxDelay = 10 * time.Minute
	}
	if procedure.BackoffFactor == 0 {
		procedure.BackoffFactor = 2.0
	}

	s.procedures[apiName] = procedure

	// Initialize API state
	s.getOrCreateAPIState(apiName)

	log.Printf("Registered recovery procedure for API %s with %d steps", apiName, len(procedure.Steps))
	return nil
}

// getOrCreateAPIState gets or creates an API recovery state
func (s *recoveryServiceImpl) getOrCreateAPIState(apiName string) *APIRecoveryState {
	if state, exists := s.status.APIStatus[apiName]; exists {
		return state
	}

	state := &APIRecoveryState{
		APIName:     apiName,
		Status:      APIStatusUnknown,
		MaxAttempts: 5, // Default
	}

	s.status.APIStatus[apiName] = state

	// Initialize metrics
	if _, exists := s.metrics.APIMetrics[apiName]; !exists {
		s.metrics.APIMetrics[apiName] = &APIRecoveryMetrics{
			APIName: apiName,
		}
	}

	return state
}

// GetRecoveryStatus returns the current recovery status for all APIs
func (s *recoveryServiceImpl) GetRecoveryStatus() *RecoveryStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.updateStatus()

	// Create a deep copy
	status := &RecoveryStatus{
		TotalAPIs:        s.status.TotalAPIs,
		HealthyAPIs:      s.status.HealthyAPIs,
		RecoveringAPIs:   s.status.RecoveringAPIs,
		FailedAPIs:       s.status.FailedAPIs,
		APIStatus:        make(map[string]*APIRecoveryState),
		LastUpdated:      time.Now(),
		MonitoringActive: s.monitoring,
	}

	for name, state := range s.status.APIStatus {
		status.APIStatus[name] = &APIRecoveryState{
			APIName:             state.APIName,
			Status:              state.Status,
			LastHealthCheck:     state.LastHealthCheck,
			LastRecoveryTry:     state.LastRecoveryTry,
			RecoveryAttempts:    state.RecoveryAttempts,
			MaxAttempts:         state.MaxAttempts,
			NextRecoveryTime:    state.NextRecoveryTime,
			ConsecutiveFailures: state.ConsecutiveFailures,
			LastError:           state.LastError,
		}
	}

	return status
}

// updateStatus updates the overall status counters
func (s *recoveryServiceImpl) updateStatus() {
	healthy, recovering, failed := 0, 0, 0

	for _, state := range s.status.APIStatus {
		switch state.Status {
		case APIStatusHealthy:
			healthy++
		case APIStatusRecovering:
			recovering++
		case APIStatusFailed, APIStatusDegraded:
			failed++
		}
	}

	s.status.TotalAPIs = len(s.status.APIStatus)
	s.status.HealthyAPIs = healthy
	s.status.RecoveringAPIs = recovering
	s.status.FailedAPIs = failed
	s.status.LastUpdated = time.Now()
}

// GetRecoveryMetrics returns recovery metrics and statistics
func (s *recoveryServiceImpl) GetRecoveryMetrics() *RecoveryMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a deep copy
	metrics := &RecoveryMetrics{
		TotalRecoveryAttempts: s.metrics.TotalRecoveryAttempts,
		SuccessfulRecoveries:  s.metrics.SuccessfulRecoveries,
		FailedRecoveries:      s.metrics.FailedRecoveries,
		AverageRecoveryTime:   s.metrics.AverageRecoveryTime,
		APIMetrics:            make(map[string]*APIRecoveryMetrics),
		LastRecoveryAttempt:   s.metrics.LastRecoveryAttempt,
		MonitoringStartTime:   s.metrics.MonitoringStartTime,
		TotalDowntime:         s.metrics.TotalDowntime,
		LastUpdated:           time.Now(),
	}

	for name, apiMetrics := range s.metrics.APIMetrics {
		metrics.APIMetrics[name] = &APIRecoveryMetrics{
			APIName:              apiMetrics.APIName,
			RecoveryAttempts:     apiMetrics.RecoveryAttempts,
			SuccessfulRecoveries: apiMetrics.SuccessfulRecoveries,
			FailedRecoveries:     apiMetrics.FailedRecoveries,
			AverageRecoveryTime:  apiMetrics.AverageRecoveryTime,
			TotalDowntime:        apiMetrics.TotalDowntime,
			LastRecovery:         apiMetrics.LastRecovery,
			ConsecutiveFailures:  apiMetrics.ConsecutiveFailures,
			SuccessRate:          apiMetrics.SuccessRate,
		}
	}

	return metrics
}

// updateRecoveryMetrics updates recovery metrics
func (s *recoveryServiceImpl) updateRecoveryMetrics(apiName string, isStart bool) {
	if isStart {
		s.metrics.TotalRecoveryAttempts++
		s.metrics.LastRecoveryAttempt = time.Now()

		if apiMetrics, exists := s.metrics.APIMetrics[apiName]; exists {
			apiMetrics.RecoveryAttempts++
			apiMetrics.LastRecovery = time.Now()
		}
	}

	s.metrics.LastUpdated = time.Now()
}

// logRecoveryEvent logs recovery events to the audit service
func (s *recoveryServiceImpl) logRecoveryEvent(apiName, event string, err error) {
	if s.auditSvc == nil {
		return
	}

	description := fmt.Sprintf("API Recovery Event: %s - %s", apiName, event)
	if err != nil {
		description += fmt.Sprintf(" (error: %s)", err.Error())
	}

	severity := "info"
	if event == "recovery_failed" || err != nil {
		severity = "error"
	} else if event == "recovery_succeeded" {
		severity = "info"
	}

	metadata := map[string]interface{}{
		"api_name":   apiName,
		"event_type": event,
		"component":  "recovery_service",
	}
	if err != nil {
		metadata["error"] = err.Error()
	}

	if logErr := s.auditSvc.LogSecurityEvent(context.Background(), "api_recovery", description, nil, severity, nil); logErr != nil {
		log.Printf("Failed to log recovery event: %v", logErr)
	}
}
