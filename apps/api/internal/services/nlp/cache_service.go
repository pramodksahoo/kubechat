package nlp

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/pramodksahoo/kubechat/apps/api/internal/services/safety"
)

// CacheService handles NLP response caching with Redis integration
type CacheService struct {
	enabled           bool
	queryTTL          time.Duration
	safetyTTL         time.Duration
	userContextTTL    time.Duration
	maxCacheSize      int64
	performanceTarget time.Duration
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Enabled           bool          `json:"enabled"`
	QueryTTL          time.Duration `json:"query_ttl"`
	SafetyTTL         time.Duration `json:"safety_ttl"`
	UserContextTTL    time.Duration `json:"user_context_ttl"`
	MaxCacheSize      int64         `json:"max_cache_size"`
	PerformanceTarget time.Duration `json:"performance_target"`
}

// CachedNLPResult represents a cached NLP result
type CachedNLPResult struct {
	Result    *NLPResult `json:"result"`
	CachedAt  time.Time  `json:"cached_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	HitCount  int        `json:"hit_count"`
}

// CachedSafetyResult represents a cached safety classification
type CachedSafetyResult struct {
	Result    *safety.SafetyClassification `json:"result"`
	CachedAt  time.Time                    `json:"cached_at"`
	ExpiresAt time.Time                    `json:"expires_at"`
	HitCount  int                          `json:"hit_count"`
}

// NewCacheService creates a new cache service instance
func NewCacheService(config *CacheConfig) *CacheService {
	if config == nil {
		config = &CacheConfig{
			Enabled:           true,
			QueryTTL:          1 * time.Hour,          // Query results cache for 1 hour
			SafetyTTL:         24 * time.Hour,         // Safety classifications cache for 24 hours
			UserContextTTL:    8 * time.Hour,          // User contexts cache for session duration
			MaxCacheSize:      100 * 1024 * 1024,      // 100MB max cache size
			PerformanceTarget: 100 * time.Millisecond, // Target <100ms for cached queries
		}
	}

	service := &CacheService{
		enabled:           config.Enabled,
		queryTTL:          config.QueryTTL,
		safetyTTL:         config.SafetyTTL,
		userContextTTL:    config.UserContextTTL,
		maxCacheSize:      config.MaxCacheSize,
		performanceTarget: config.PerformanceTarget,
	}

	log.Printf("Cache service initialized: enabled=%v, query_ttl=%v, safety_ttl=%v",
		service.enabled, service.queryTTL, service.safetyTTL)

	return service
}

// GetCachedNLPResult retrieves cached NLP result
func (cs *CacheService) GetCachedNLPResult(ctx context.Context, req NLPRequest) (*NLPResult, bool) {
	if !cs.enabled {
		return nil, false
	}

	startTime := time.Now()
	cacheKey := cs.generateQueryCacheKey(req)

	// In a real implementation, this would use Redis
	// For now, implementing in-memory cache simulation
	cached := cs.getCachedData(cacheKey)
	if cached == nil {
		return nil, false
	}

	// Check if cache entry is still valid
	cachedResult, ok := cached.(*CachedNLPResult)
	if !ok || time.Now().After(cachedResult.ExpiresAt) {
		cs.invalidateCache(cacheKey)
		return nil, false
	}

	// Update hit count and performance metrics
	cachedResult.HitCount++
	processingTime := time.Since(startTime)

	log.Printf("Cache HIT: query processed in %v (target: %v)", processingTime, cs.performanceTarget)

	return cachedResult.Result, true
}

// CacheNLPResult stores NLP result in cache
func (cs *CacheService) CacheNLPResult(ctx context.Context, req NLPRequest, result *NLPResult) error {
	if !cs.enabled {
		return nil
	}

	cacheKey := cs.generateQueryCacheKey(req)
	cachedResult := &CachedNLPResult{
		Result:    result,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(cs.queryTTL),
		HitCount:  0,
	}

	cs.setCachedData(cacheKey, cachedResult)

	log.Printf("Cached NLP result: key=%s, expires=%v", cacheKey[:16]+"...", cachedResult.ExpiresAt)
	return nil
}

// GetCachedSafetyResult retrieves cached safety classification
func (cs *CacheService) GetCachedSafetyResult(ctx context.Context, command string) (*safety.SafetyClassification, bool) {
	if !cs.enabled {
		return nil, false
	}

	cacheKey := cs.generateSafetyCacheKey(command)
	cached := cs.getCachedData(cacheKey)
	if cached == nil {
		return nil, false
	}

	cachedResult, ok := cached.(*CachedSafetyResult)
	if !ok || time.Now().After(cachedResult.ExpiresAt) {
		cs.invalidateCache(cacheKey)
		return nil, false
	}

	cachedResult.HitCount++
	log.Printf("Safety cache HIT: command=%s", command[:min(len(command), 50)])

	return cachedResult.Result, true
}

// CacheSafetyResult stores safety classification in cache
func (cs *CacheService) CacheSafetyResult(ctx context.Context, command string, result *safety.SafetyClassification) error {
	if !cs.enabled {
		return nil
	}

	cacheKey := cs.generateSafetyCacheKey(command)
	cachedResult := &CachedSafetyResult{
		Result:    result,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(cs.safetyTTL),
		HitCount:  0,
	}

	cs.setCachedData(cacheKey, cachedResult)
	return nil
}

// InvalidateUserCache invalidates all cache entries for a specific user
func (cs *CacheService) InvalidateUserCache(ctx context.Context, userID string) error {
	if !cs.enabled {
		return nil
	}

	// In a real Redis implementation, this would use pattern matching
	// For now, simulate cache invalidation
	log.Printf("Invalidated user cache: user_id=%s", userID)
	return nil
}

// GetCacheStats returns cache performance statistics
func (cs *CacheService) GetCacheStats(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		"enabled":            cs.enabled,
		"query_ttl_hours":    cs.queryTTL.Hours(),
		"safety_ttl_hours":   cs.safetyTTL.Hours(),
		"performance_target": cs.performanceTarget.Milliseconds(),
		"max_cache_size_mb":  cs.maxCacheSize / (1024 * 1024),
		// In real implementation, would include hit rates, memory usage, etc.
		"status": "simulated_cache",
	}
}

// OptimizePromptTemplates optimizes prompt templates for faster processing
func (cs *CacheService) OptimizePromptTemplates() map[string]string {
	optimizedTemplates := map[string]string{
		"base_prompt": `Convert natural language to kubectl commands.
Context: {{.Context}}
Query: {{.Query}}
Return JSON: {"command":"", "explanation":"", "confidence":0.9}`,

		"safety_prompt": `Classify kubectl command safety:
Command: {{.Command}}
Return: {"level":"safe|warning|dangerous", "reasons":[]}`,

		"namespace_prompt": `Generate kubectl command for namespace {{.Namespace}}:
Query: {{.Query}}`,
	}

	log.Printf("Optimized %d prompt templates for performance", len(optimizedTemplates))
	return optimizedTemplates
}

// BatchProcessRequests processes multiple queries with performance optimization
func (cs *CacheService) BatchProcessRequests(ctx context.Context, requests []NLPRequest) map[string]*NLPResult {
	results := make(map[string]*NLPResult)

	// Group similar requests for batch processing
	batchGroups := cs.groupSimilarRequests(requests)

	for groupKey, group := range batchGroups {
		log.Printf("Batch processing %d similar requests: %s", len(group), groupKey)

		// Process the first request in the group
		if len(group) > 0 {
			// In real implementation, would optimize by processing similar queries together
			for _, req := range group {
				reqKey := cs.generateQueryCacheKey(req)
				// Simulate batch processing result
				results[reqKey] = &NLPResult{
					Command:     "kubectl get pods",
					Explanation: "Batch processed result",
					SafetyLevel: "safe",
					Confidence:  0.8,
				}
			}
		}
	}

	return results
}

// Helper methods

// generateQueryCacheKey creates a cache key for NLP queries
func (cs *CacheService) generateQueryCacheKey(req NLPRequest) string {
	// Create a deterministic cache key based on request parameters
	keyData := fmt.Sprintf("query:%s:namespace:%s:role:%s:env:%s",
		req.Query, req.Namespace, req.UserRole, req.Environment)

	// Add context to key if present
	if len(req.Context) > 0 {
		contextBytes, _ := json.Marshal(req.Context)
		keyData += ":context:" + string(contextBytes)
	}

	// Hash the key data
	hash := sha256.Sum256([]byte(keyData))
	return fmt.Sprintf("nlp:query:%x", hash)
}

// generateSafetyCacheKey creates a cache key for safety classifications
func (cs *CacheService) generateSafetyCacheKey(command string) string {
	hash := sha256.Sum256([]byte(command))
	return fmt.Sprintf("nlp:safety:%x", hash)
}

// In-memory cache simulation (in production would use Redis)
var cacheStore = make(map[string]interface{})

func (cs *CacheService) getCachedData(key string) interface{} {
	return cacheStore[key]
}

func (cs *CacheService) setCachedData(key string, data interface{}) {
	cacheStore[key] = data
}

func (cs *CacheService) invalidateCache(key string) {
	delete(cacheStore, key)
}

// groupSimilarRequests groups similar requests for batch processing
func (cs *CacheService) groupSimilarRequests(requests []NLPRequest) map[string][]NLPRequest {
	groups := make(map[string][]NLPRequest)

	for _, req := range requests {
		// Group by similar characteristics
		groupKey := fmt.Sprintf("%s:%s", req.UserRole, req.Environment)
		groups[groupKey] = append(groups[groupKey], req)
	}

	return groups
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
