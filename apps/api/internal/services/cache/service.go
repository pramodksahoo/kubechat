package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// Service defines the cache service interface
type Service interface {
	// Command result caching
	StoreCommandResult(ctx context.Context, executionID uuid.UUID, result *models.KubernetesOperationResult) error
	GetCommandResult(ctx context.Context, executionID uuid.UUID) (*models.KubernetesOperationResult, error)

	// Resource-based caching with intelligent invalidation
	StoreResourceData(ctx context.Context, key string, data interface{}, expiration time.Duration) error
	GetResourceData(ctx context.Context, key string, dest interface{}) error
	InvalidateResourceCache(ctx context.Context, resourceType, namespace, name string) error
	InvalidateNamespaceCache(ctx context.Context, namespace string) error

	// Performance metrics
	GetCacheMetrics(ctx context.Context) (*CacheMetrics, error)
	RecordCacheHit(ctx context.Context, cacheType string)
	RecordCacheMiss(ctx context.Context, cacheType string)

	// Health and management
	HealthCheck(ctx context.Context) error
	FlushCache(ctx context.Context) error
}

// CacheMetrics represents cache performance metrics
type CacheMetrics struct {
	// Hit/Miss ratios
	CommandResultHits    int64   `json:"command_result_hits"`
	CommandResultMisses  int64   `json:"command_result_misses"`
	CommandResultHitRate float64 `json:"command_result_hit_rate"`

	ResourceDataHits    int64   `json:"resource_data_hits"`
	ResourceDataMisses  int64   `json:"resource_data_misses"`
	ResourceDataHitRate float64 `json:"resource_data_hit_rate"`

	// Cache size and performance
	TotalKeys        int64         `json:"total_keys"`
	TotalMemoryUsage int64         `json:"total_memory_usage_bytes"`
	AverageLatency   time.Duration `json:"average_latency"`

	// Invalidation metrics
	InvalidationCount int64 `json:"invalidation_count"`

	// Uptime
	Uptime time.Duration `json:"uptime"`
}

// Config represents cache service configuration
type Config struct {
	RedisURL               string        `json:"redis_url"`
	CommandResultTTL       time.Duration `json:"command_result_ttl"`
	ResourceDataTTL        time.Duration `json:"resource_data_ttl"`
	MetricsRetentionPeriod time.Duration `json:"metrics_retention_period"`
	MaxMemoryUsage         string        `json:"max_memory_usage"`
	EnableMetrics          bool          `json:"enable_metrics"`
}

// service implements the cache Service interface
type service struct {
	redis     *redis.Client
	config    *Config
	startTime time.Time
}

// NewService creates a new cache service
func NewService(config *Config) (Service, error) {
	if config == nil {
		config = &Config{
			RedisURL:               "redis://localhost:6379",
			CommandResultTTL:       15 * time.Minute,
			ResourceDataTTL:        5 * time.Minute,
			MetricsRetentionPeriod: 24 * time.Hour,
			MaxMemoryUsage:         "256mb",
			EnableMetrics:          true,
		}
	}

	// Parse Redis options from URL
	opts, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Create Redis client
	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &service{
		redis:     client,
		config:    config,
		startTime: time.Now(),
	}, nil
}

// StoreCommandResult stores command execution result with TTL
func (s *service) StoreCommandResult(ctx context.Context, executionID uuid.UUID, result *models.KubernetesOperationResult) error {
	key := s.getCommandResultKey(executionID)

	// Serialize result to JSON
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal command result: %w", err)
	}

	// Store with TTL
	if err := s.redis.Set(ctx, key, data, s.config.CommandResultTTL).Err(); err != nil {
		return fmt.Errorf("failed to store command result in cache: %w", err)
	}

	// Record cache operation metrics
	if s.config.EnableMetrics {
		s.recordMetric(ctx, "command_result_stored", 1)
	}

	return nil
}

// GetCommandResult retrieves cached command result
func (s *service) GetCommandResult(ctx context.Context, executionID uuid.UUID) (*models.KubernetesOperationResult, error) {
	key := s.getCommandResultKey(executionID)

	// Get data from cache
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Record cache miss
			if s.config.EnableMetrics {
				s.RecordCacheMiss(ctx, "command_result")
			}
			return nil, nil // Cache miss, not an error
		}
		return nil, fmt.Errorf("failed to get command result from cache: %w", err)
	}

	// Deserialize result
	var result models.KubernetesOperationResult
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached command result: %w", err)
	}

	// Record cache hit
	if s.config.EnableMetrics {
		s.RecordCacheHit(ctx, "command_result")
	}

	return &result, nil
}

// StoreResourceData stores generic resource data with expiration
func (s *service) StoreResourceData(ctx context.Context, key string, data interface{}, expiration time.Duration) error {
	// Serialize data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal resource data: %w", err)
	}

	// Use provided expiration or default
	if expiration == 0 {
		expiration = s.config.ResourceDataTTL
	}

	// Store with expiration
	if err := s.redis.Set(ctx, key, jsonData, expiration).Err(); err != nil {
		return fmt.Errorf("failed to store resource data in cache: %w", err)
	}

	// Record metrics
	if s.config.EnableMetrics {
		s.recordMetric(ctx, "resource_data_stored", 1)
	}

	return nil
}

// GetResourceData retrieves cached resource data
func (s *service) GetResourceData(ctx context.Context, key string, dest interface{}) error {
	// Get data from cache
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Record cache miss
			if s.config.EnableMetrics {
				s.RecordCacheMiss(ctx, "resource_data")
			}
			return nil // Cache miss, not an error
		}
		return fmt.Errorf("failed to get resource data from cache: %w", err)
	}

	// Deserialize into destination
	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cached resource data: %w", err)
	}

	// Record cache hit
	if s.config.EnableMetrics {
		s.RecordCacheHit(ctx, "resource_data")
	}

	return nil
}

// InvalidateResourceCache invalidates cache entries for specific resource
func (s *service) InvalidateResourceCache(ctx context.Context, resourceType, namespace, name string) error {
	// Build pattern for resource-specific keys
	patterns := []string{
		fmt.Sprintf("resource:%s:%s:%s:*", resourceType, namespace, name),
		fmt.Sprintf("list:%s:%s:*", resourceType, namespace),
		fmt.Sprintf("describe:%s:%s:%s:*", resourceType, namespace, name),
	}

	var deletedCount int64
	for _, pattern := range patterns {
		keys, err := s.redis.Keys(ctx, pattern).Result()
		if err != nil {
			continue // Continue with other patterns on error
		}

		if len(keys) > 0 {
			deleted, err := s.redis.Del(ctx, keys...).Result()
			if err == nil {
				deletedCount += deleted
			}
		}
	}

	// Record invalidation metrics
	if s.config.EnableMetrics {
		s.recordMetric(ctx, "cache_invalidations", deletedCount)
	}

	return nil
}

// InvalidateNamespaceCache invalidates all cache entries for a namespace
func (s *service) InvalidateNamespaceCache(ctx context.Context, namespace string) error {
	pattern := fmt.Sprintf("*:%s:*", namespace)

	keys, err := s.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for namespace invalidation: %w", err)
	}

	if len(keys) > 0 {
		deletedCount, err := s.redis.Del(ctx, keys...).Result()
		if err != nil {
			return fmt.Errorf("failed to delete keys for namespace invalidation: %w", err)
		}

		// Record invalidation metrics
		if s.config.EnableMetrics {
			s.recordMetric(ctx, "cache_invalidations", deletedCount)
		}
	}

	return nil
}

// GetCacheMetrics returns cache performance metrics
func (s *service) GetCacheMetrics(ctx context.Context) (*CacheMetrics, error) {
	metrics := &CacheMetrics{
		Uptime: time.Since(s.startTime),
	}

	// Get hit/miss statistics from Redis
	if s.config.EnableMetrics {
		// Command result metrics
		cmdHits, _ := s.getMetricValue(ctx, "command_result_hits")
		cmdMisses, _ := s.getMetricValue(ctx, "command_result_misses")
		metrics.CommandResultHits = cmdHits
		metrics.CommandResultMisses = cmdMisses
		if cmdHits+cmdMisses > 0 {
			metrics.CommandResultHitRate = float64(cmdHits) / float64(cmdHits+cmdMisses)
		}

		// Resource data metrics
		resHits, _ := s.getMetricValue(ctx, "resource_data_hits")
		resMisses, _ := s.getMetricValue(ctx, "resource_data_misses")
		metrics.ResourceDataHits = resHits
		metrics.ResourceDataMisses = resMisses
		if resHits+resMisses > 0 {
			metrics.ResourceDataHitRate = float64(resHits) / float64(resHits+resMisses)
		}

		// Invalidation count
		metrics.InvalidationCount, _ = s.getMetricValue(ctx, "cache_invalidations")
	}

	// Get Redis info
	info, err := s.redis.Info(ctx, "memory", "keyspace").Result()
	if err == nil {
		// Parse memory usage and key count from info
		lines := strings.Split(info, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "used_memory:") {
				fmt.Sscanf(line, "used_memory:%d", &metrics.TotalMemoryUsage)
			}
			if strings.Contains(line, "keys=") {
				var keys int64
				fmt.Sscanf(line, "db0:keys=%d", &keys)
				metrics.TotalKeys = keys
			}
		}
	}

	return metrics, nil
}

// RecordCacheHit records a cache hit for metrics
func (s *service) RecordCacheHit(ctx context.Context, cacheType string) {
	if s.config.EnableMetrics {
		key := fmt.Sprintf("%s_hits", cacheType)
		s.redis.Incr(ctx, key)
		s.redis.Expire(ctx, key, s.config.MetricsRetentionPeriod)
	}
}

// RecordCacheMiss records a cache miss for metrics
func (s *service) RecordCacheMiss(ctx context.Context, cacheType string) {
	if s.config.EnableMetrics {
		key := fmt.Sprintf("%s_misses", cacheType)
		s.redis.Incr(ctx, key)
		s.redis.Expire(ctx, key, s.config.MetricsRetentionPeriod)
	}
}

// HealthCheck performs cache service health check
func (s *service) HealthCheck(ctx context.Context) error {
	// Test Redis connection
	if err := s.redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	// Test basic operations
	testKey := "health_check_test"
	if err := s.redis.Set(ctx, testKey, "test", time.Second).Err(); err != nil {
		return fmt.Errorf("Redis write test failed: %w", err)
	}

	if err := s.redis.Del(ctx, testKey).Err(); err != nil {
		return fmt.Errorf("Redis delete test failed: %w", err)
	}

	return nil
}

// FlushCache clears all cached data
func (s *service) FlushCache(ctx context.Context) error {
	if err := s.redis.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("failed to flush cache: %w", err)
	}
	return nil
}

// Helper methods

func (s *service) getCommandResultKey(executionID uuid.UUID) string {
	return fmt.Sprintf("command_result:%s", executionID)
}

func (s *service) recordMetric(ctx context.Context, metric string, value int64) {
	s.redis.IncrBy(ctx, metric, value)
	s.redis.Expire(ctx, metric, s.config.MetricsRetentionPeriod)
}

func (s *service) getMetricValue(ctx context.Context, metric string) (int64, error) {
	val, err := s.redis.Get(ctx, metric).Int64()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	return val, nil
}
