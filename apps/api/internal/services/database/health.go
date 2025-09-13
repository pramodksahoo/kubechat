package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// healthChecker manages database health checking
type healthChecker struct {
	service  *service
	ctx      context.Context
	cancel   context.CancelFunc
	running  bool
	interval time.Duration
}

// newHealthChecker creates a new health checker
func newHealthChecker(s *service) *healthChecker {
	return &healthChecker{
		service:  s,
		interval: s.config.HealthCheckInterval,
	}
}

// start begins health checking
func (hc *healthChecker) start(ctx context.Context) error {
	if hc.running {
		return fmt.Errorf("health checker is already running")
	}

	hc.ctx, hc.cancel = context.WithCancel(ctx)
	hc.running = true

	go hc.healthCheckLoop()
	log.Println("Database health checker started")
	return nil
}

// stop stops health checking
func (hc *healthChecker) stop() {
	if !hc.running {
		return
	}

	hc.cancel()
	hc.running = false
	log.Println("Database health checker stopped")
}

// healthCheckLoop runs the health checking loop
func (hc *healthChecker) healthCheckLoop() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			hc.performHealthChecks()
		}
	}
}

// performHealthChecks checks health of all database instances and connection pools
func (hc *healthChecker) performHealthChecks() {
	// Check database instances
	instances, err := hc.service.GetAllDatabaseInstances()
	if err != nil {
		log.Printf("Failed to get database instances for health check: %v", err)
		return
	}

	for _, instance := range instances {
		health, err := hc.checkInstanceHealth(instance.ID)
		if err != nil {
			log.Printf("Health check failed for instance %s: %v", instance.ID, err)
			continue
		}

		if err := hc.service.UpdateInstanceHealth(instance.ID, health); err != nil {
			log.Printf("Failed to update health for instance %s: %v", instance.ID, err)
		}
	}

	// Check connection pools
	pools, err := hc.service.GetAllConnectionPools()
	if err != nil {
		log.Printf("Failed to get connection pools for health check: %v", err)
		return
	}

	for _, pool := range pools {
		hc.checkPoolHealth(pool)
	}

	// Check clusters
	clusters, err := hc.service.GetAllClusters()
	if err != nil {
		log.Printf("Failed to get clusters for health check: %v", err)
		return
	}

	for _, cluster := range clusters {
		if err := hc.service.UpdateClusterHealth(cluster.ID); err != nil {
			log.Printf("Failed to update cluster health %s: %v", cluster.ID, err)
		}
	}
}

// checkInstanceHealth performs health check on a database instance
func (hc *healthChecker) checkInstanceHealth(instanceID string) (*models.DatabaseHealthCheck, error) {
	hc.service.mu.RLock()
	instance, exists := hc.service.databaseInstances[instanceID]
	hc.service.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("database instance not found: %s", instanceID)
	}

	health := &models.DatabaseHealthCheck{
		CheckedAt:        time.Now(),
		ConsecutiveFails: instance.Health.ConsecutiveFails,
	}

	// Find connection pool for this instance
	var pool *models.ConnectionPool
	databaseID := fmt.Sprintf("%s:%d", instance.Host, instance.Port)

	hc.service.mu.RLock()
	for _, p := range hc.service.connectionPools {
		if p.DatabaseID == databaseID {
			pool = p
			break
		}
	}
	hc.service.mu.RUnlock()

	if pool == nil {
		health.IsConnected = false
		health.LastError = "no connection pool found"
		health.ConsecutiveFails++
		return health, nil
	}

	// Perform ping test
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.DB.PingContext(ctx); err != nil {
		health.IsConnected = false
		health.LastError = err.Error()
		health.ConsecutiveFails++
		health.ResponseTime = time.Since(start)
		return health, nil
	}

	health.IsConnected = true
	health.ResponseTime = time.Since(start)
	health.ConsecutiveFails = 0
	health.LastError = ""

	// Get connection statistics
	if pool.Stats != nil {
		health.ActiveConns = pool.Stats.InUse
		health.IdleConns = pool.Stats.Idle
		health.WaitingConns = int(pool.Stats.WaitCount)
	}

	return health, nil
}

// checkPoolHealth performs health check on a connection pool
func (hc *healthChecker) checkPoolHealth(pool *models.ConnectionPool) {
	if pool.DB == nil {
		return
	}

	// Update pool statistics
	hc.service.updatePoolStats(pool)

	// Update last health check time
	pool.LastHealthCheck = time.Now()

	// Check if pool is experiencing issues
	if pool.Stats != nil {
		utilization := pool.Stats.GetUtilization()
		if utilization > 90.0 { // Pool utilization above 90%
			log.Printf("High utilization detected on pool %s: %.2f%%", pool.ID, utilization)
		}

		// Check for connection wait issues
		if pool.Stats.WaitCount > 0 && pool.Stats.WaitDuration > time.Second {
			log.Printf("Connection wait issues detected on pool %s: %d waits, avg duration: %v",
				pool.ID, pool.Stats.WaitCount, pool.Stats.WaitDuration)
		}
	}

	// Perform basic connectivity test
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := pool.DB.PingContext(ctx); err != nil {
		log.Printf("Connection pool %s failed ping test: %v", pool.ID, err)
	}
}

// performAdvancedHealthCheck performs more detailed health checks
func (hc *healthChecker) performAdvancedHealthCheck(instanceID string) (*models.DatabaseHealthCheck, error) {
	instance, err := hc.service.GetDatabaseInstance(instanceID)
	if err != nil {
		return nil, err
	}

	// Find connection pool
	databaseID := fmt.Sprintf("%s:%d", instance.Host, instance.Port)
	var pool *models.ConnectionPool

	hc.service.mu.RLock()
	for _, p := range hc.service.connectionPools {
		if p.DatabaseID == databaseID {
			pool = p
			break
		}
	}
	hc.service.mu.RUnlock()

	if pool == nil {
		return nil, fmt.Errorf("no connection pool found for instance %s", instanceID)
	}

	health := &models.DatabaseHealthCheck{
		CheckedAt: time.Now(),
	}

	// Test basic connectivity
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.DB.PingContext(ctx); err != nil {
		health.IsConnected = false
		health.LastError = err.Error()
		health.ResponseTime = time.Since(start)
		return health, nil
	}

	health.IsConnected = true
	health.ResponseTime = time.Since(start)

	// Test with a simple query
	queryStart := time.Now()
	var result int
	if err := pool.DB.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
		health.LastError = fmt.Sprintf("Query test failed: %v", err)
	} else {
		queryDuration := time.Since(queryStart)
		if queryDuration > health.ResponseTime {
			health.ResponseTime = queryDuration
		}
	}

	// Get detailed connection stats
	stats := pool.DB.Stats()
	health.ActiveConns = stats.InUse
	health.IdleConns = stats.Idle
	health.WaitingConns = int(stats.WaitCount)

	return health, nil
}

// checkReplicationLag checks replication lag for replica instances
func (hc *healthChecker) checkReplicationLag(replicaID string) (time.Duration, error) {
	// This would typically query replication status tables
	// For now, return a simulated lag
	return 100 * time.Millisecond, nil
}

// validateConnectionPool validates a connection pool configuration
func (hc *healthChecker) validateConnectionPool(pool *models.ConnectionPool) error {
	if pool.DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	if pool.Config == nil {
		return fmt.Errorf("pool configuration is nil")
	}

	// Test connection limits
	if pool.Config.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections must be positive")
	}

	if pool.Config.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections cannot be negative")
	}

	if pool.Config.ConnMaxLifetime <= 0 {
		return fmt.Errorf("connection max lifetime must be positive")
	}

	return nil
}

// generateHealthReport generates a comprehensive health report
func (hc *healthChecker) generateHealthReport() (*DatabaseHealthReport, error) {
	report := &DatabaseHealthReport{
		GeneratedAt: time.Now(),
		Instances:   make(map[string]*models.DatabaseHealthCheck),
		Pools:       make(map[string]*models.PoolStats),
		Clusters:    make(map[string]*models.ClusterHealth),
	}

	// Get all instances health
	instances, err := hc.service.GetAllDatabaseInstances()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instances: %w", err)
	}

	healthyInstances := 0
	for _, instance := range instances {
		if instance.IsHealthy() {
			healthyInstances++
		}
		report.Instances[instance.ID] = &instance.Health
	}

	// Get all pools stats
	pools, err := hc.service.GetAllConnectionPools()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection pools: %w", err)
	}

	for _, pool := range pools {
		if pool.Stats != nil {
			report.Pools[pool.ID] = pool.Stats
		}
	}

	// Get all clusters health
	clusters, err := hc.service.GetAllClusters()
	if err != nil {
		return nil, fmt.Errorf("failed to get clusters: %w", err)
	}

	for _, cluster := range clusters {
		health, err := hc.service.GetClusterHealth(cluster.ID)
		if err != nil {
			continue
		}
		report.Clusters[cluster.ID] = health
	}

	// Calculate overall health score
	if len(instances) > 0 {
		report.OverallHealthScore = float64(healthyInstances) / float64(len(instances)) * 100.0
	}

	report.TotalInstances = len(instances)
	report.HealthyInstances = healthyInstances
	report.TotalPools = len(pools)
	report.TotalClusters = len(clusters)

	return report, nil
}

// DatabaseHealthReport represents a comprehensive health report
type DatabaseHealthReport struct {
	GeneratedAt        time.Time                              `json:"generated_at"`
	OverallHealthScore float64                                `json:"overall_health_score"`
	TotalInstances     int                                    `json:"total_instances"`
	HealthyInstances   int                                    `json:"healthy_instances"`
	TotalPools         int                                    `json:"total_pools"`
	TotalClusters      int                                    `json:"total_clusters"`
	Instances          map[string]*models.DatabaseHealthCheck `json:"instances"`
	Pools              map[string]*models.PoolStats           `json:"pools"`
	Clusters           map[string]*models.ClusterHealth       `json:"clusters"`
}
