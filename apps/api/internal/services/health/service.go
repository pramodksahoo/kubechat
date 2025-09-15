package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// Service defines the health check service interface
type Service interface {
	// Health checks
	CheckOverallHealth(ctx context.Context) (*models.HealthStatus, error)
	CheckDatabaseHealth(ctx context.Context) (*models.ComponentHealth, error)
	CheckRedisHealth(ctx context.Context) (*models.ComponentHealth, error)
	CheckExternalServicesHealth(ctx context.Context) (*models.ComponentHealth, error)

	// Component registration
	RegisterComponent(name string, checker ComponentChecker)
	UnregisterComponent(name string)

	// Monitoring
	GetHealthMetrics() *models.HealthMetrics
	StartHealthMonitoring(ctx context.Context) error
	StopHealthMonitoring()

	// Configuration
	SetHealthCheckInterval(interval time.Duration)
	SetUnhealthyThreshold(threshold int)
}

// ComponentChecker defines the interface for component health checkers
type ComponentChecker interface {
	CheckHealth(ctx context.Context) (*models.ComponentHealth, error)
	GetComponentName() string
}

// Config represents the health check service configuration
type Config struct {
	CheckInterval      time.Duration `json:"check_interval"`
	UnhealthyThreshold int           `json:"unhealthy_threshold"`
	EnableMonitoring   bool          `json:"enable_monitoring"`
	DatabaseTimeout    time.Duration `json:"database_timeout"`
	RedisTimeout       time.Duration `json:"redis_timeout"`
	ExternalTimeout    time.Duration `json:"external_timeout"`
	AlertThreshold     time.Duration `json:"alert_threshold"`
	EnableAlerts       bool          `json:"enable_alerts"`
	LogHealthStatus    bool          `json:"log_health_status"`
	MetricsRetention   time.Duration `json:"metrics_retention"`
}

// service implements the Health service interface
type service struct {
	config     *Config
	db         *sqlx.DB
	components map[string]ComponentChecker
	metrics    *models.HealthMetrics
	mutex      sync.RWMutex
	ticker     *time.Ticker
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewService creates a new health check service
func NewService(db *sqlx.DB, config *Config) Service {
	if config == nil {
		config = &Config{
			CheckInterval:      30 * time.Second,
			UnhealthyThreshold: 3,
			EnableMonitoring:   true,
			DatabaseTimeout:    5 * time.Second,
			RedisTimeout:       5 * time.Second,
			ExternalTimeout:    10 * time.Second,
			AlertThreshold:     2 * time.Minute,
			EnableAlerts:       true,
			LogHealthStatus:    true,
			MetricsRetention:   24 * time.Hour,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	svc := &service{
		config:     config,
		db:         db,
		components: make(map[string]ComponentChecker),
		metrics: &models.HealthMetrics{
			StartTime:        time.Now(),
			LastCheck:        time.Time{},
			TotalChecks:      0,
			FailedChecks:     0,
			ComponentMetrics: make(map[string]*models.ComponentMetrics),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Register default components
	svc.registerDefaultComponents()

	return svc
}

// CheckOverallHealth performs a comprehensive health check of all components
func (s *service) CheckOverallHealth(ctx context.Context) (*models.HealthStatus, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	startTime := time.Now()
	status := &models.HealthStatus{
		Status:     models.HealthStatusHealthy,
		Timestamp:  startTime,
		Components: make(map[string]*models.ComponentHealth),
		Metadata: map[string]interface{}{
			"check_duration":   time.Duration(0),
			"total_components": len(s.components),
		},
	}

	// Check all registered components
	var wg sync.WaitGroup
	var componentMutex sync.Mutex
	healthyComponents := 0

	for name, checker := range s.components {
		wg.Add(1)
		go func(name string, checker ComponentChecker) {
			defer wg.Done()

			componentCtx, cancel := context.WithTimeout(ctx, s.config.ExternalTimeout)
			defer cancel()

			componentHealth, err := checker.CheckHealth(componentCtx)

			componentMutex.Lock()
			defer componentMutex.Unlock()

			if err != nil {
				componentHealth = &models.ComponentHealth{
					Name:      name,
					Status:    models.HealthStatusUnhealthy,
					Message:   err.Error(),
					Timestamp: time.Now(),
					Metadata:  map[string]interface{}{"error": err.Error()},
				}
			}

			status.Components[name] = componentHealth

			if componentHealth.Status == models.HealthStatusHealthy {
				healthyComponents++
			}
		}(name, checker)
	}

	wg.Wait()

	// Determine overall status
	totalComponents := len(s.components)
	if totalComponents == 0 {
		status.Status = models.HealthStatusHealthy
		status.Message = "No components registered"
	} else if healthyComponents == totalComponents {
		status.Status = models.HealthStatusHealthy
		status.Message = "All components healthy"
	} else if healthyComponents > totalComponents/2 {
		status.Status = models.HealthStatusDegraded
		status.Message = fmt.Sprintf("%d of %d components healthy", healthyComponents, totalComponents)
	} else {
		status.Status = models.HealthStatusUnhealthy
		status.Message = fmt.Sprintf("Only %d of %d components healthy", healthyComponents, totalComponents)
	}

	duration := time.Since(startTime)
	status.Metadata["check_duration"] = duration

	// Update metrics
	s.updateMetrics(status)

	return status, nil
}

// CheckDatabaseHealth checks the database connection and basic functionality
func (s *service) CheckDatabaseHealth(ctx context.Context) (*models.ComponentHealth, error) {
	if s.db == nil {
		return &models.ComponentHealth{
			Name:      "database",
			Status:    models.HealthStatusUnhealthy,
			Message:   "Database connection not available",
			Timestamp: time.Now(),
		}, nil
	}

	checkCtx, cancel := context.WithTimeout(ctx, s.config.DatabaseTimeout)
	defer cancel()

	startTime := time.Now()

	// Test connection
	if err := s.db.PingContext(checkCtx); err != nil {
		return &models.ComponentHealth{
			Name:      "database",
			Status:    models.HealthStatusUnhealthy,
			Message:   fmt.Sprintf("Database ping failed: %v", err),
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"error":         err.Error(),
				"response_time": time.Since(startTime),
			},
		}, nil
	}

	// Check pool stats
	stats := s.db.Stats()
	responseTime := time.Since(startTime)

	health := &models.ComponentHealth{
		Name:      "database",
		Status:    models.HealthStatusHealthy,
		Message:   "Database connection healthy",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"response_time":    responseTime,
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
			"wait_count":       stats.WaitCount,
			"wait_duration":    stats.WaitDuration,
		},
	}

	// Check for potential issues
	if stats.WaitCount > 10 {
		health.Status = models.HealthStatusDegraded
		health.Message = fmt.Sprintf("High connection wait count: %d", stats.WaitCount)
	}

	if responseTime > time.Second {
		health.Status = models.HealthStatusDegraded
		health.Message = fmt.Sprintf("Slow database response: %v", responseTime)
	}

	return health, nil
}

// CheckRedisHealth checks the Redis connection and basic functionality
func (s *service) CheckRedisHealth(ctx context.Context) (*models.ComponentHealth, error) {
	// Redis health check not implemented - would require Redis client
	return &models.ComponentHealth{
		Name:      "redis",
		Status:    models.HealthStatusUnknown,
		Message:   "Redis health check not implemented",
		Timestamp: time.Now(),
	}, nil
}

// CheckExternalServicesHealth checks external service dependencies
func (s *service) CheckExternalServicesHealth(ctx context.Context) (*models.ComponentHealth, error) {
	// This is a placeholder for external service health checks
	// In a real implementation, this would check services like Ollama, OpenAI, etc.

	health := &models.ComponentHealth{
		Name:      "external_services",
		Status:    models.HealthStatusHealthy,
		Message:   "External services check not implemented",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"checked_services": []string{},
		},
	}

	return health, nil
}

// RegisterComponent registers a component for health checking
func (s *service) RegisterComponent(name string, checker ComponentChecker) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.components[name] = checker

	// Initialize metrics for the component
	if s.metrics.ComponentMetrics[name] == nil {
		s.metrics.ComponentMetrics[name] = &models.ComponentMetrics{
			Name:                name,
			TotalChecks:         0,
			FailedChecks:        0,
			LastCheckTime:       time.Time{},
			AverageResponseTime: 0,
		}
	}
}

// UnregisterComponent removes a component from health checking
func (s *service) UnregisterComponent(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.components, name)
	delete(s.metrics.ComponentMetrics, name)
}

// GetHealthMetrics returns current health metrics
func (s *service) GetHealthMetrics() *models.HealthMetrics {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Create a copy to avoid race conditions
	metrics := *s.metrics
	metrics.ComponentMetrics = make(map[string]*models.ComponentMetrics)

	for name, componentMetrics := range s.metrics.ComponentMetrics {
		componentCopy := *componentMetrics
		metrics.ComponentMetrics[name] = &componentCopy
	}

	metrics.Uptime = time.Since(s.metrics.StartTime)

	return &metrics
}

// StartHealthMonitoring starts the health monitoring goroutine
func (s *service) StartHealthMonitoring(ctx context.Context) error {
	if !s.config.EnableMonitoring {
		return nil
	}

	s.ticker = time.NewTicker(s.config.CheckInterval)

	go s.monitoringLoop()

	return nil
}

// StopHealthMonitoring stops the health monitoring goroutine
func (s *service) StopHealthMonitoring() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.cancel()
}

// SetHealthCheckInterval sets the health check interval
func (s *service) SetHealthCheckInterval(interval time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.config.CheckInterval = interval

	if s.ticker != nil {
		s.ticker.Reset(interval)
	}
}

// SetUnhealthyThreshold sets the unhealthy threshold
func (s *service) SetUnhealthyThreshold(threshold int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.config.UnhealthyThreshold = threshold
}

// Helper methods

func (s *service) registerDefaultComponents() {
	// Register database checker
	if s.db != nil {
		s.RegisterComponent("database", &DatabaseChecker{
			db:      s.db,
			timeout: s.config.DatabaseTimeout,
		})
	}

	// Redis checker would be registered here if Redis is available
}

func (s *service) updateMetrics(status *models.HealthStatus) {
	s.metrics.LastCheck = status.Timestamp
	s.metrics.TotalChecks++

	if status.Status != models.HealthStatusHealthy {
		s.metrics.FailedChecks++
	}

	// Update component metrics
	for name, componentHealth := range status.Components {
		if metrics, exists := s.metrics.ComponentMetrics[name]; exists {
			metrics.TotalChecks++
			metrics.LastCheckTime = componentHealth.Timestamp

			if componentHealth.Status != models.HealthStatusHealthy {
				metrics.FailedChecks++
			}

			// Update average response time if available
			if responseTime, ok := componentHealth.Metadata["response_time"].(time.Duration); ok {
				if metrics.AverageResponseTime == 0 {
					metrics.AverageResponseTime = responseTime
				} else {
					// Exponential moving average
					metrics.AverageResponseTime = time.Duration(
						float64(metrics.AverageResponseTime)*0.9 + float64(responseTime)*0.1,
					)
				}
			}
		}
	}
}

func (s *service) monitoringLoop() {
	for {
		select {
		case <-s.ticker.C:
			// Perform health check
			ctx, cancel := context.WithTimeout(s.ctx, s.config.CheckInterval/2)
			status, err := s.CheckOverallHealth(ctx)
			cancel()

			if err != nil && s.config.LogHealthStatus {
				fmt.Printf("Health check error: %v\n", err)
			}

			if status != nil && s.config.LogHealthStatus {
				if status.Status != models.HealthStatusHealthy {
					fmt.Printf("Health check alert: %s - %s\n", status.Status, status.Message)
				}
			}

		case <-s.ctx.Done():
			return
		}
	}
}

// Default component checkers

type DatabaseChecker struct {
	db      *sqlx.DB
	timeout time.Duration
}

func (d *DatabaseChecker) CheckHealth(ctx context.Context) (*models.ComponentHealth, error) {
	checkCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	startTime := time.Now()

	if err := d.db.PingContext(checkCtx); err != nil {
		return &models.ComponentHealth{
			Name:      "database",
			Status:    models.HealthStatusUnhealthy,
			Message:   fmt.Sprintf("Database ping failed: %v", err),
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"error":         err.Error(),
				"response_time": time.Since(startTime),
			},
		}, nil
	}

	stats := d.db.Stats()
	responseTime := time.Since(startTime)

	status := models.HealthStatusHealthy
	message := "Database healthy"

	if stats.WaitCount > 10 {
		status = models.HealthStatusDegraded
		message = fmt.Sprintf("High wait count: %d", stats.WaitCount)
	}

	return &models.ComponentHealth{
		Name:      "database",
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"response_time":    responseTime,
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
			"wait_count":       stats.WaitCount,
		},
	}, nil
}

func (d *DatabaseChecker) GetComponentName() string {
	return "database"
}

// RedisChecker would be implemented here if Redis client is available
