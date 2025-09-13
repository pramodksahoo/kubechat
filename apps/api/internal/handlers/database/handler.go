package database

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	databaseService "github.com/pramodksahoo/kubechat/apps/api/internal/services/database"
)

// Handler handles HTTP requests for database management
type Handler struct {
	service databaseService.Service
}

// NewHandler creates a new database handler
func NewHandler(service databaseService.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers the database routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	dbRoutes := router.Group("/database")
	{
		// Connection Pool endpoints
		dbRoutes.POST("/pools", h.CreateConnectionPool)
		dbRoutes.GET("/pools", h.GetAllConnectionPools)
		dbRoutes.GET("/pools/:poolId", h.GetConnectionPool)
		dbRoutes.DELETE("/pools/:poolId", h.DestroyConnectionPool)
		dbRoutes.GET("/pools/:poolId/stats", h.GetPoolStats)

		// Database Instance endpoints
		dbRoutes.POST("/instances", h.RegisterDatabaseInstance)
		dbRoutes.GET("/instances", h.GetAllDatabaseInstances)
		dbRoutes.GET("/instances/:instanceId", h.GetDatabaseInstance)
		dbRoutes.GET("/instances/:instanceId/health", h.CheckDatabaseHealth)

		// Cluster Management endpoints
		dbRoutes.POST("/clusters", h.CreateDatabaseCluster)
		dbRoutes.GET("/clusters", h.GetAllClusters)
		dbRoutes.GET("/clusters/:clusterId", h.GetDatabaseCluster)
		dbRoutes.GET("/clusters/:clusterId/health", h.GetClusterHealth)
		dbRoutes.POST("/clusters/:clusterId/failover", h.TriggerFailover)
		dbRoutes.GET("/clusters/:clusterId/failover-status", h.CheckFailoverStatus)

		// Query execution endpoints
		dbRoutes.POST("/pools/:poolId/query", h.ExecuteQuery)
		dbRoutes.POST("/pools/:poolId/transaction", h.ExecuteTransaction)

		// Connection endpoints
		dbRoutes.GET("/clusters/:clusterId/connection/read", h.GetReadOnlyConnection)
		dbRoutes.GET("/clusters/:clusterId/connection/write", h.GetReadWriteConnection)

		// Metrics and monitoring endpoints
		dbRoutes.GET("/metrics", h.GetDatabaseMetrics)
		dbRoutes.GET("/health", h.GetHealthStatus)

		// Migration endpoints
		dbRoutes.GET("/migrations", h.GetMigrationStatus)
		dbRoutes.POST("/migrations", h.ApplyMigration)

		// Backup endpoints
		dbRoutes.POST("/backups", h.CreateBackup)
		dbRoutes.GET("/backups/:backupId", h.GetBackupStatus)
		dbRoutes.POST("/backups/:backupId/restore", h.RestoreFromBackup)
	}
}

// CreateConnectionPool creates a new connection pool
//
//	@Summary		Create database connection pool
//	@Description	Creates a new database connection pool with specified configuration
//	@Tags			Database Management
//	@Accept			json
//	@Produce		json
//	@Param			config	body		models.DatabaseConfig	true	"Database configuration"
//	@Success		201		{object}	map[string]interface{}	"Connection pool created successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid configuration format"
//	@Failure		500		{object}	map[string]interface{}	"Failed to create connection pool"
//	@Security		BearerAuth
//	@Router			/database/pools [post]
func (h *Handler) CreateConnectionPool(c *gin.Context) {
	var config models.DatabaseConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid configuration format",
			"details": err.Error(),
		})
		return
	}

	// Set defaults if not provided
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 25
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 5
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = time.Hour
	}
	if config.ConnMaxIdleTime == 0 {
		config.ConnMaxIdleTime = 30 * time.Minute
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	pool, err := h.service.CreateConnectionPool(&config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create connection pool",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Connection pool created successfully",
		"pool_id":   pool.ID,
		"pool":      pool,
		"timestamp": time.Now(),
	})
}

// GetAllConnectionPools returns all connection pools
//
//	@Summary		Get all connection pools
//	@Description	Returns information about all database connection pools
//	@Tags			Database Management
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of connection pools"
//	@Failure		500	{object}	map[string]interface{}	"Failed to get connection pools"
//	@Security		BearerAuth
//	@Router			/database/pools [get]
func (h *Handler) GetAllConnectionPools(c *gin.Context) {
	pools, err := h.service.GetAllConnectionPools()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get connection pools",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pools":     pools,
		"count":     len(pools),
		"timestamp": time.Now(),
	})
}

// GetConnectionPool returns a specific connection pool
//
//	@Summary		Get specific connection pool
//	@Description	Returns information about a specific database connection pool
//	@Tags			Database Management
//	@Produce		json
//	@Param			poolId	path		string					true	"Connection pool ID"
//	@Success		200		{object}	map[string]interface{}	"Connection pool information"
//	@Failure		404		{object}	map[string]interface{}	"Connection pool not found"
//	@Security		BearerAuth
//	@Router			/database/pools/{poolId} [get]
func (h *Handler) GetConnectionPool(c *gin.Context) {
	poolID := c.Param("poolId")

	pool, err := h.service.GetConnectionPool(poolID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Connection pool not found",
			"pool_id": poolID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pool":      pool,
		"timestamp": time.Now(),
	})
}

// DestroyConnectionPool destroys a connection pool
//
//	@Summary		Destroy connection pool
//	@Description	Destroys a database connection pool and closes all connections
//	@Tags			Database Management
//	@Produce		json
//	@Param			poolId	path		string					true	"Connection pool ID"
//	@Success		200		{object}	map[string]interface{}	"Connection pool destroyed successfully"
//	@Failure		500		{object}	map[string]interface{}	"Failed to destroy connection pool"
//	@Security		BearerAuth
//	@Router			/database/pools/{poolId} [delete]
func (h *Handler) DestroyConnectionPool(c *gin.Context) {
	poolID := c.Param("poolId")

	if err := h.service.DestroyConnectionPool(poolID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to destroy connection pool",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Connection pool destroyed successfully",
		"pool_id":   poolID,
		"timestamp": time.Now(),
	})
}

// GetPoolStats returns connection pool statistics
//
//	@Summary		Get connection pool statistics
//	@Description	Returns detailed statistics for a specific connection pool
//	@Tags			Database Management
//	@Produce		json
//	@Param			poolId	path		string					true	"Connection pool ID"
//	@Success		200		{object}	map[string]interface{}	"Connection pool statistics"
//	@Failure		404		{object}	map[string]interface{}	"Pool not found"
//	@Security		BearerAuth
//	@Router			/database/pools/{poolId}/stats [get]
func (h *Handler) GetPoolStats(c *gin.Context) {
	poolID := c.Param("poolId")

	stats, err := h.service.GetPoolStats(poolID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Failed to get pool statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pool_id":             poolID,
		"stats":               stats,
		"utilization_percent": stats.GetUtilization(),
		"timestamp":           time.Now(),
	})
}

// RegisterDatabaseInstance registers a new database instance
//
//	@Summary		Register database instance
//	@Description	Registers a new database instance in the system
//	@Tags			Database Management
//	@Accept			json
//	@Produce		json
//	@Param			instance	body		models.DatabaseInstance	true	"Database instance configuration"
//	@Success		201			{object}	map[string]interface{}	"Database instance registered successfully"
//	@Failure		400			{object}	map[string]interface{}	"Invalid instance format"
//	@Failure		500			{object}	map[string]interface{}	"Failed to register database instance"
//	@Security		BearerAuth
//	@Router			/database/instances [post]
func (h *Handler) RegisterDatabaseInstance(c *gin.Context) {
	var instance models.DatabaseInstance
	if err := c.ShouldBindJSON(&instance); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid instance format",
			"details": err.Error(),
		})
		return
	}

	if err := h.service.RegisterDatabaseInstance(&instance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to register database instance",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Database instance registered successfully",
		"instance_id": instance.ID,
		"instance":    instance,
		"timestamp":   time.Now(),
	})
}

// GetAllDatabaseInstances returns all database instances
//
//	@Summary		Get all database instances
//	@Description	Returns information about all registered database instances
//	@Tags			Database Management
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of database instances"
//	@Failure		500	{object}	map[string]interface{}	"Failed to get database instances"
//	@Security		BearerAuth
//	@Router			/database/instances [get]
func (h *Handler) GetAllDatabaseInstances(c *gin.Context) {
	instances, err := h.service.GetAllDatabaseInstances()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get database instances",
			"details": err.Error(),
		})
		return
	}

	// Count healthy instances
	healthyCount := 0
	for _, instance := range instances {
		if instance.IsHealthy() {
			healthyCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"instances":     instances,
		"count":         len(instances),
		"healthy_count": healthyCount,
		"timestamp":     time.Now(),
	})
}

// GetDatabaseInstance returns a specific database instance
//
//	@Summary		Get specific database instance
//	@Description	Returns information about a specific database instance
//	@Tags			Database Management
//	@Produce		json
//	@Param			instanceId	path		string					true	"Database instance ID"
//	@Success		200			{object}	map[string]interface{}	"Database instance information"
//	@Failure		404			{object}	map[string]interface{}	"Database instance not found"
//	@Security		BearerAuth
//	@Router			/database/instances/{instanceId} [get]
func (h *Handler) GetDatabaseInstance(c *gin.Context) {
	instanceID := c.Param("instanceId")

	instance, err := h.service.GetDatabaseInstance(instanceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":       "Database instance not found",
			"instance_id": instanceID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"instance":   instance,
		"is_healthy": instance.IsHealthy(),
		"timestamp":  time.Now(),
	})
}

// CheckDatabaseHealth performs health check on a database instance
//
//	@Summary		Check database instance health
//	@Description	Performs a health check on a specific database instance
//	@Tags			Database Management
//	@Produce		json
//	@Param			instanceId	path		string					true	"Database instance ID"
//	@Success		200			{object}	map[string]interface{}	"Database health status"
//	@Failure		500			{object}	map[string]interface{}	"Health check failed"
//	@Security		BearerAuth
//	@Router			/database/instances/{instanceId}/health [get]
func (h *Handler) CheckDatabaseHealth(c *gin.Context) {
	instanceID := c.Param("instanceId")

	health, err := h.service.CheckDatabaseHealth(instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Health check failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"instance_id": instanceID,
		"health":      health,
		"is_healthy":  health.IsConnected && health.ConsecutiveFails == 0,
		"timestamp":   time.Now(),
	})
}

// CreateDatabaseCluster creates a new database cluster
//
//	@Summary		Create database cluster
//	@Description	Creates a new database cluster configuration
//	@Tags			Database Management
//	@Accept			json
//	@Produce		json
//	@Param			cluster	body		models.DatabaseCluster	true	"Database cluster configuration"
//	@Success		201		{object}	map[string]interface{}	"Database cluster created successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid cluster format"
//	@Failure		500		{object}	map[string]interface{}	"Failed to create database cluster"
//	@Security		BearerAuth
//	@Router			/database/clusters [post]
func (h *Handler) CreateDatabaseCluster(c *gin.Context) {
	var cluster models.DatabaseCluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid cluster format",
			"details": err.Error(),
		})
		return
	}

	if err := h.service.CreateDatabaseCluster(&cluster); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create database cluster",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Database cluster created successfully",
		"cluster_id": cluster.ID,
		"cluster":    cluster,
		"timestamp":  time.Now(),
	})
}

// GetAllClusters returns all database clusters
//
//	@Summary		Get all database clusters
//	@Description	Returns information about all database clusters
//	@Tags			Database Management
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of database clusters"
//	@Failure		500	{object}	map[string]interface{}	"Failed to get database clusters"
//	@Security		BearerAuth
//	@Router			/database/clusters [get]
func (h *Handler) GetAllClusters(c *gin.Context) {
	clusters, err := h.service.GetAllClusters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get database clusters",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clusters":  clusters,
		"count":     len(clusters),
		"timestamp": time.Now(),
	})
}

// GetDatabaseCluster returns a specific database cluster
//
//	@Summary		Get specific database cluster
//	@Description	Returns information about a specific database cluster
//	@Tags			Database Management
//	@Produce		json
//	@Param			clusterId	path		string					true	"Database cluster ID"
//	@Success		200			{object}	map[string]interface{}	"Database cluster information"
//	@Failure		404			{object}	map[string]interface{}	"Database cluster not found"
//	@Security		BearerAuth
//	@Router			/database/clusters/{clusterId} [get]
func (h *Handler) GetDatabaseCluster(c *gin.Context) {
	clusterID := c.Param("clusterId")

	cluster, err := h.service.GetDatabaseCluster(clusterID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "Database cluster not found",
			"cluster_id": clusterID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cluster":   cluster,
		"timestamp": time.Now(),
	})
}

// GetClusterHealth returns cluster health information
//
//	@Summary		Get cluster health status
//	@Description	Returns health status for a database cluster
//	@Tags			Database Management
//	@Produce		json
//	@Param			clusterId	path		string					true	"Database cluster ID"
//	@Success		200			{object}	map[string]interface{}	"Cluster health information"
//	@Failure		500			{object}	map[string]interface{}	"Failed to get cluster health"
//	@Security		BearerAuth
//	@Router			/database/clusters/{clusterId}/health [get]
func (h *Handler) GetClusterHealth(c *gin.Context) {
	clusterID := c.Param("clusterId")

	health, err := h.service.GetClusterHealth(clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get cluster health",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cluster_id": clusterID,
		"health":     health,
		"timestamp":  time.Now(),
	})
}

// TriggerFailover triggers a manual failover
//
//	@Summary		Trigger database failover
//	@Description	Triggers a manual failover for a database cluster
//	@Tags			Database Management
//	@Produce		json
//	@Param			clusterId	path		string					true	"Database cluster ID"
//	@Success		202			{object}	map[string]interface{}	"Failover initiated successfully"
//	@Failure		500			{object}	map[string]interface{}	"Failed to trigger failover"
//	@Security		BearerAuth
//	@Router			/database/clusters/{clusterId}/failover [post]
func (h *Handler) TriggerFailover(c *gin.Context) {
	clusterID := c.Param("clusterId")

	if err := h.service.TriggerFailover(clusterID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to trigger failover",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Failover initiated successfully",
		"cluster_id": clusterID,
		"timestamp":  time.Now(),
	})
}

// CheckFailoverStatus checks if failover is in progress
//
//	@Summary		Check failover status
//	@Description	Checks if a failover is currently in progress for a cluster
//	@Tags			Database Management
//	@Produce		json
//	@Param			clusterId	path		string					true	"Database cluster ID"
//	@Success		200			{object}	map[string]interface{}	"Failover status information"
//	@Failure		500			{object}	map[string]interface{}	"Failed to check failover status"
//	@Security		BearerAuth
//	@Router			/database/clusters/{clusterId}/failover-status [get]
func (h *Handler) CheckFailoverStatus(c *gin.Context) {
	clusterID := c.Param("clusterId")

	inProgress, err := h.service.CheckFailoverStatus(clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check failover status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cluster_id":           clusterID,
		"failover_in_progress": inProgress,
		"timestamp":            time.Now(),
	})
}

// ExecuteQuery executes a database query
//
//	@Summary		Execute database query
//	@Description	Executes a SQL query on a specific connection pool
//	@Tags			Database Management
//	@Accept			json
//	@Produce		json
//	@Param			poolId	path		string									true	"Connection pool ID"
//	@Param			request	body		object{query=string,args=[]interface{}}	true	"Query request"
//	@Success		200		{object}	map[string]interface{}					"Query executed successfully"
//	@Failure		400		{object}	map[string]interface{}					"Invalid query request"
//	@Failure		500		{object}	map[string]interface{}					"Query execution failed"
//	@Security		BearerAuth
//	@Router			/database/pools/{poolId}/query [post]
func (h *Handler) ExecuteQuery(c *gin.Context) {
	poolID := c.Param("poolId")

	var request struct {
		Query string        `json:"query" binding:"required"`
		Args  []interface{} `json:"args,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query request",
			"details": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := h.service.ExecuteQuery(ctx, poolID, request.Query, request.Args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Query execution failed",
			"details": err.Error(),
		})
		return
	}

	// Get result info
	rowsAffected, _ := result.RowsAffected()
	lastInsertID, _ := result.LastInsertId()

	c.JSON(http.StatusOK, gin.H{
		"message":        "Query executed successfully",
		"rows_affected":  rowsAffected,
		"last_insert_id": lastInsertID,
		"timestamp":      time.Now(),
	})
}

// ExecuteTransaction executes a database transaction
//
//	@Summary		Execute database transaction
//	@Description	Executes multiple SQL operations as a single transaction
//	@Tags			Database Management
//	@Accept			json
//	@Produce		json
//	@Param			poolId	path		string									true	"Connection pool ID"
//	@Param			request	body		object{operations=[]models.QueryInfo}	true	"Transaction request"
//	@Success		200		{object}	map[string]interface{}					"Transaction executed successfully"
//	@Failure		400		{object}	map[string]interface{}					"Invalid transaction request"
//	@Failure		500		{object}	map[string]interface{}					"Transaction execution failed"
//	@Security		BearerAuth
//	@Router			/database/pools/{poolId}/transaction [post]
func (h *Handler) ExecuteTransaction(c *gin.Context) {
	poolID := c.Param("poolId")

	var request struct {
		Operations []models.QueryInfo `json:"operations" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid transaction request",
			"details": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	transaction, err := h.service.ExecuteTransaction(ctx, poolID, request.Operations)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Transaction execution failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Transaction executed successfully",
		"transaction": transaction,
		"timestamp":   time.Now(),
	})
}

// GetReadOnlyConnection tests read-only connection
//
//	@Summary		Test read-only connection
//	@Description	Tests connectivity to read-only database instances
//	@Tags			Database Management
//	@Produce		json
//	@Param			clusterId	path		string					true	"Database cluster ID"
//	@Success		200			{object}	map[string]interface{}	"Read-only connection successful"
//	@Failure		500			{object}	map[string]interface{}	"Read-only connection failed"
//	@Security		BearerAuth
//	@Router			/database/clusters/{clusterId}/connection/read [get]
func (h *Handler) GetReadOnlyConnection(c *gin.Context) {
	clusterID := c.Param("clusterId")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := h.service.GetReadOnlyConnection(ctx, clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get read-only connection",
			"details": err.Error(),
		})
		return
	}

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Read-only connection test failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Read-only connection successful",
		"cluster_id": clusterID,
		"timestamp":  time.Now(),
	})
}

// GetReadWriteConnection tests read-write connection
//
//	@Summary		Test read-write connection
//	@Description	Tests connectivity to read-write database instances
//	@Tags			Database Management
//	@Produce		json
//	@Param			clusterId	path		string					true	"Database cluster ID"
//	@Success		200			{object}	map[string]interface{}	"Read-write connection successful"
//	@Failure		500			{object}	map[string]interface{}	"Read-write connection failed"
//	@Security		BearerAuth
//	@Router			/database/clusters/{clusterId}/connection/write [get]
func (h *Handler) GetReadWriteConnection(c *gin.Context) {
	clusterID := c.Param("clusterId")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := h.service.GetReadWriteConnection(ctx, clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get read-write connection",
			"details": err.Error(),
		})
		return
	}

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Read-write connection test failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Read-write connection successful",
		"cluster_id": clusterID,
		"timestamp":  time.Now(),
	})
}

// GetDatabaseMetrics returns database metrics
//
//	@Summary		Get database metrics
//	@Description	Returns performance metrics and statistics for database operations
//	@Tags			Database Management
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Database metrics and statistics"
//	@Security		BearerAuth
//	@Router			/database/metrics [get]
func (h *Handler) GetDatabaseMetrics(c *gin.Context) {
	metrics := h.service.GetDatabaseMetrics()

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
		"summary": gin.H{
			"success_rate_percent": metrics.CalculateSuccessRate(),
			"error_rate_percent":   metrics.CalculateErrorRate(),
			"total_queries":        metrics.TotalQueries,
			"uptime":               metrics.Uptime,
		},
		"timestamp": time.Now(),
	})
}

// GetHealthStatus returns database service health
//
//	@Summary		Get database service health
//	@Description	Returns health status and feature information for the database service
//	@Tags			Database Management
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Database service health status"
//	@Router			/database/health [get]
func (h *Handler) GetHealthStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "database",
		"status":  "healthy",
		"version": "1.0.0",
		"features": gin.H{
			"connection_pooling":   true,
			"failover_management":  true,
			"health_monitoring":    true,
			"cluster_management":   true,
			"query_execution":      true,
			"transaction_support":  true,
			"backup_restore":       true,
			"migration_management": true,
		},
		"timestamp": time.Now(),
	})
}

// GetMigrationStatus returns migration status
//
//	@Summary		Get migration status
//	@Description	Returns status of all database migrations
//	@Tags			Database Management
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Migration status information"
//	@Failure		500	{object}	map[string]interface{}	"Failed to get migration status"
//	@Security		BearerAuth
//	@Router			/database/migrations [get]
func (h *Handler) GetMigrationStatus(c *gin.Context) {
	migrations, err := h.service.GetMigrationStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get migration status",
			"details": err.Error(),
		})
		return
	}

	appliedCount := 0
	for _, migration := range migrations {
		if migration.Applied {
			appliedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"migrations":         migrations,
		"total_migrations":   len(migrations),
		"applied_migrations": appliedCount,
		"timestamp":          time.Now(),
	})
}

// ApplyMigration applies a database migration
//
//	@Summary		Apply database migration
//	@Description	Applies a database migration to update schema
//	@Tags			Database Management
//	@Accept			json
//	@Produce		json
//	@Param			migration	body		models.DatabaseMigration	true	"Migration configuration"
//	@Success		200			{object}	map[string]interface{}		"Migration applied successfully"
//	@Failure		400			{object}	map[string]interface{}		"Invalid migration format"
//	@Failure		500			{object}	map[string]interface{}		"Migration failed"
//	@Security		BearerAuth
//	@Router			/database/migrations [post]
func (h *Handler) ApplyMigration(c *gin.Context) {
	var migration models.DatabaseMigration
	if err := c.ShouldBindJSON(&migration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid migration format",
			"details": err.Error(),
		})
		return
	}

	if err := h.service.ApplyMigration(&migration); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Migration failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Migration applied successfully",
		"migration": migration,
		"timestamp": time.Now(),
	})
}

// CreateBackup creates a database backup
//
//	@Summary		Create database backup
//	@Description	Initiates a database backup operation
//	@Tags			Database Management
//	@Accept			json
//	@Produce		json
//	@Param			backup	body		models.DatabaseBackup	true	"Backup configuration"
//	@Success		202		{object}	map[string]interface{}	"Backup creation initiated"
//	@Failure		400		{object}	map[string]interface{}	"Invalid backup configuration"
//	@Failure		500		{object}	map[string]interface{}	"Backup creation failed"
//	@Security		BearerAuth
//	@Router			/database/backups [post]
func (h *Handler) CreateBackup(c *gin.Context) {
	var backup models.DatabaseBackup
	if err := c.ShouldBindJSON(&backup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid backup configuration",
			"details": err.Error(),
		})
		return
	}

	if err := h.service.CreateBackup(&backup); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Backup creation failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":   "Backup creation initiated",
		"backup":    backup,
		"timestamp": time.Now(),
	})
}

// GetBackupStatus returns backup status
//
//	@Summary		Get backup status
//	@Description	Returns status information for a specific backup
//	@Tags			Database Management
//	@Produce		json
//	@Param			backupId	path		string					true	"Backup ID"
//	@Success		200			{object}	map[string]interface{}	"Backup status information"
//	@Failure		404			{object}	map[string]interface{}	"Backup not found"
//	@Security		BearerAuth
//	@Router			/database/backups/{backupId} [get]
func (h *Handler) GetBackupStatus(c *gin.Context) {
	backupID := c.Param("backupId")

	backup, err := h.service.GetBackupStatus(backupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":     "Backup not found",
			"backup_id": backupID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"backup":    backup,
		"timestamp": time.Now(),
	})
}

// RestoreFromBackup restores from a backup
//
//	@Summary		Restore from backup
//	@Description	Restores a database from a backup to a target instance
//	@Tags			Database Management
//	@Accept			json
//	@Produce		json
//	@Param			backupId	path		string								true	"Backup ID"
//	@Param			request		body		object{target_instance_id=string}	true	"Restore request"
//	@Success		202			{object}	map[string]interface{}				"Restore initiated successfully"
//	@Failure		400			{object}	map[string]interface{}				"Invalid restore request"
//	@Failure		500			{object}	map[string]interface{}				"Restore failed"
//	@Security		BearerAuth
//	@Router			/database/backups/{backupId}/restore [post]
func (h *Handler) RestoreFromBackup(c *gin.Context) {
	backupID := c.Param("backupId")

	var request struct {
		TargetInstanceID string `json:"target_instance_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid restore request",
			"details": err.Error(),
		})
		return
	}

	if err := h.service.RestoreFromBackup(backupID, request.TargetInstanceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Restore failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":            "Restore initiated successfully",
		"backup_id":          backupID,
		"target_instance_id": request.TargetInstanceID,
		"timestamp":          time.Now(),
	})
}
