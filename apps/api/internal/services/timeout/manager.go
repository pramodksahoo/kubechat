package timeout

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager defines the timeout management interface
type Manager interface {
	// Timeout management
	SetCommandTimeout(ctx context.Context, executionID uuid.UUID, duration time.Duration) error
	CancelCommand(ctx context.Context, executionID uuid.UUID, reason string) error
	IsTimedOut(ctx context.Context, executionID uuid.UUID) (bool, error)

	// Context management
	CreateTimedContext(parent context.Context, duration time.Duration) (context.Context, context.CancelFunc)
	WithDeadline(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc)

	// Monitoring and cleanup
	GetActiveTimeouts() []*TimeoutInfo
	CleanupExpiredTimeouts(ctx context.Context) (int, error)

	// Health monitoring
	HealthCheck(ctx context.Context) error
}

// TimeoutInfo represents information about an active timeout
type TimeoutInfo struct {
	ExecutionID  uuid.UUID     `json:"execution_id"`
	StartTime    time.Time     `json:"start_time"`
	Timeout      time.Duration `json:"timeout"`
	Deadline     time.Time     `json:"deadline"`
	Status       string        `json:"status"` // "active", "expired", "cancelled"
	CancelReason string        `json:"cancel_reason,omitempty"`
	CreatedBy    uuid.UUID     `json:"created_by,omitempty"`
}

// TimeoutConfig represents timeout configuration
type TimeoutConfig struct {
	// Default timeouts for different operation types
	DefaultCommandTimeout time.Duration `json:"default_command_timeout"`
	MaxCommandTimeout     time.Duration `json:"max_command_timeout"`
	LongRunningTimeout    time.Duration `json:"long_running_timeout"`

	// Specific operation timeouts
	GetOperationTimeout     time.Duration `json:"get_operation_timeout"`
	ListOperationTimeout    time.Duration `json:"list_operation_timeout"`
	DeleteOperationTimeout  time.Duration `json:"delete_operation_timeout"`
	ScaleOperationTimeout   time.Duration `json:"scale_operation_timeout"`
	RestartOperationTimeout time.Duration `json:"restart_operation_timeout"`
	LogsOperationTimeout    time.Duration `json:"logs_operation_timeout"`

	// Cleanup and monitoring
	CleanupInterval    time.Duration `json:"cleanup_interval"`
	TimeoutGracePeriod time.Duration `json:"timeout_grace_period"`
	MaxActiveTimeouts  int           `json:"max_active_timeouts"`
}

// manager implements the timeout Manager interface
type manager struct {
	config      *TimeoutConfig
	timeouts    map[uuid.UUID]*TimeoutInfo
	mutex       sync.RWMutex
	cancelFuncs map[uuid.UUID]context.CancelFunc

	// Cleanup goroutine control
	cleanupCtx    context.Context
	cleanupCancel context.CancelFunc
	cleanupWg     sync.WaitGroup
}

// NewManager creates a new timeout manager
func NewManager(config *TimeoutConfig) Manager {
	if config == nil {
		config = &TimeoutConfig{
			DefaultCommandTimeout:   30 * time.Second,
			MaxCommandTimeout:       10 * time.Minute,
			LongRunningTimeout:      5 * time.Minute,
			GetOperationTimeout:     30 * time.Second,
			ListOperationTimeout:    60 * time.Second,
			DeleteOperationTimeout:  2 * time.Minute,
			ScaleOperationTimeout:   3 * time.Minute,
			RestartOperationTimeout: 5 * time.Minute,
			LogsOperationTimeout:    2 * time.Minute,
			CleanupInterval:         1 * time.Minute,
			TimeoutGracePeriod:      10 * time.Second,
			MaxActiveTimeouts:       1000,
		}
	}

	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())

	m := &manager{
		config:        config,
		timeouts:      make(map[uuid.UUID]*TimeoutInfo),
		cancelFuncs:   make(map[uuid.UUID]context.CancelFunc),
		cleanupCtx:    cleanupCtx,
		cleanupCancel: cleanupCancel,
	}

	// Start cleanup goroutine
	m.cleanupWg.Add(1)
	go m.runCleanupRoutine()

	return m
}

// SetCommandTimeout sets a timeout for a specific command execution
func (m *manager) SetCommandTimeout(ctx context.Context, executionID uuid.UUID, duration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Validate timeout duration
	if duration <= 0 {
		return fmt.Errorf("timeout duration must be positive")
	}
	if duration > m.config.MaxCommandTimeout {
		return fmt.Errorf("timeout duration %v exceeds maximum allowed %v", duration, m.config.MaxCommandTimeout)
	}

	// Check if we're at the limit of active timeouts
	if len(m.timeouts) >= m.config.MaxActiveTimeouts {
		return fmt.Errorf("maximum active timeouts (%d) reached", m.config.MaxActiveTimeouts)
	}

	// Create timeout info
	timeoutInfo := &TimeoutInfo{
		ExecutionID: executionID,
		StartTime:   time.Now(),
		Timeout:     duration,
		Deadline:    time.Now().Add(duration),
		Status:      "active",
	}

	m.timeouts[executionID] = timeoutInfo
	return nil
}

// CancelCommand cancels a command and its associated timeout
func (m *manager) CancelCommand(ctx context.Context, executionID uuid.UUID, reason string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	timeoutInfo, exists := m.timeouts[executionID]
	if !exists {
		return fmt.Errorf("timeout not found for execution ID: %s", executionID)
	}

	// Update timeout info
	timeoutInfo.Status = "cancelled"
	timeoutInfo.CancelReason = reason

	// Cancel the context if it exists
	if cancelFunc, exists := m.cancelFuncs[executionID]; exists {
		cancelFunc()
		delete(m.cancelFuncs, executionID)
	}

	return nil
}

// IsTimedOut checks if a command has timed out
func (m *manager) IsTimedOut(ctx context.Context, executionID uuid.UUID) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	timeoutInfo, exists := m.timeouts[executionID]
	if !exists {
		return false, fmt.Errorf("timeout not found for execution ID: %s", executionID)
	}

	if timeoutInfo.Status == "cancelled" {
		return true, nil
	}

	if time.Now().After(timeoutInfo.Deadline) {
		// Mark as expired
		timeoutInfo.Status = "expired"
		return true, nil
	}

	return false, nil
}

// CreateTimedContext creates a context with timeout and registers it
func (m *manager) CreateTimedContext(parent context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	timedCtx, cancel := context.WithTimeout(parent, duration)

	// Generate execution ID for tracking
	executionID := uuid.New()

	// Register the timeout
	m.mutex.Lock()
	m.cancelFuncs[executionID] = cancel
	m.mutex.Unlock()

	// Create a custom cancel function that cleans up our tracking
	customCancel := func() {
		cancel()
		m.mutex.Lock()
		delete(m.cancelFuncs, executionID)
		delete(m.timeouts, executionID)
		m.mutex.Unlock()
	}

	return timedCtx, customCancel
}

// WithDeadline creates a context with a specific deadline
func (m *manager) WithDeadline(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	deadlineCtx, cancel := context.WithDeadline(parent, deadline)

	// Generate execution ID for tracking
	executionID := uuid.New()

	// Register the timeout
	m.mutex.Lock()
	m.cancelFuncs[executionID] = cancel
	m.mutex.Unlock()

	// Create a custom cancel function that cleans up our tracking
	customCancel := func() {
		cancel()
		m.mutex.Lock()
		delete(m.cancelFuncs, executionID)
		delete(m.timeouts, executionID)
		m.mutex.Unlock()
	}

	return deadlineCtx, customCancel
}

// GetActiveTimeouts returns all active timeout information
func (m *manager) GetActiveTimeouts() []*TimeoutInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var activeTimeouts []*TimeoutInfo
	for _, timeoutInfo := range m.timeouts {
		// Create a copy to avoid mutation
		infoCopy := *timeoutInfo
		activeTimeouts = append(activeTimeouts, &infoCopy)
	}

	return activeTimeouts
}

// CleanupExpiredTimeouts removes expired timeout entries
func (m *manager) CleanupExpiredTimeouts(ctx context.Context) (int, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var cleanedCount int
	now := time.Now()

	for executionID, timeoutInfo := range m.timeouts {
		// Check if timeout has expired beyond grace period
		if now.After(timeoutInfo.Deadline.Add(m.config.TimeoutGracePeriod)) {
			// Cancel associated context if exists
			if cancelFunc, exists := m.cancelFuncs[executionID]; exists {
				cancelFunc()
				delete(m.cancelFuncs, executionID)
			}

			// Remove timeout info
			delete(m.timeouts, executionID)
			cleanedCount++
		}
	}

	return cleanedCount, nil
}

// HealthCheck performs health check on the timeout manager
func (m *manager) HealthCheck(ctx context.Context) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Check if cleanup routine is still running
	select {
	case <-m.cleanupCtx.Done():
		return fmt.Errorf("cleanup routine has stopped")
	default:
	}

	// Check for any issues with active timeouts
	activeCount := len(m.timeouts)
	if activeCount > m.config.MaxActiveTimeouts {
		return fmt.Errorf("active timeout count (%d) exceeds maximum (%d)", activeCount, m.config.MaxActiveTimeouts)
	}

	return nil
}

// GetTimeoutForOperation returns appropriate timeout duration for operation type
func (m *manager) GetTimeoutForOperation(operation string) time.Duration {
	switch operation {
	case "get", "describe":
		return m.config.GetOperationTimeout
	case "list":
		return m.config.ListOperationTimeout
	case "delete":
		return m.config.DeleteOperationTimeout
	case "scale":
		return m.config.ScaleOperationTimeout
	case "restart":
		return m.config.RestartOperationTimeout
	case "logs":
		return m.config.LogsOperationTimeout
	default:
		return m.config.DefaultCommandTimeout
	}
}

// Shutdown gracefully shuts down the timeout manager
func (m *manager) Shutdown(ctx context.Context) error {
	// Cancel cleanup routine
	m.cleanupCancel()

	// Wait for cleanup routine to finish with timeout
	done := make(chan struct{})
	go func() {
		m.cleanupWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Cleanup completed normally
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out waiting for cleanup routine")
	}

	// Cancel all active contexts
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, cancelFunc := range m.cancelFuncs {
		cancelFunc()
	}

	// Clear all maps
	m.timeouts = make(map[uuid.UUID]*TimeoutInfo)
	m.cancelFuncs = make(map[uuid.UUID]context.CancelFunc)

	return nil
}

// runCleanupRoutine runs the background cleanup routine
func (m *manager) runCleanupRoutine() {
	defer m.cleanupWg.Done()

	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.cleanupCtx.Done():
			return
		case <-ticker.C:
			if cleanedCount, err := m.CleanupExpiredTimeouts(m.cleanupCtx); err == nil && cleanedCount > 0 {
				fmt.Printf("Cleaned up %d expired timeouts\n", cleanedCount)
			}
		}
	}
}
