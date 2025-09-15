package external

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// loadBalancerImpl implements ProviderLoadBalancer interface (Task 7.3)
type loadBalancerImpl struct {
	mu                 sync.RWMutex
	registry           ProviderRegistry
	strategy           LoadBalancingStrategy
	providerWeights    map[string]float64
	roundRobinIndex    int
	selectionHistory   []*SelectionEvent
	performanceMetrics map[string]*ProviderMetrics
	stats              *LoadBalancingStats
}

// NewProviderLoadBalancer creates a new provider load balancer
func NewProviderLoadBalancer(registry ProviderRegistry) ProviderLoadBalancer {
	lb := &loadBalancerImpl{
		registry:           registry,
		strategy:           LoadBalancingWeighted,
		providerWeights:    make(map[string]float64),
		roundRobinIndex:    0,
		selectionHistory:   make([]*SelectionEvent, 0),
		performanceMetrics: make(map[string]*ProviderMetrics),
		stats: &LoadBalancingStats{
			Strategy:           LoadBalancingWeighted,
			TotalRequests:      0,
			ProviderRequests:   make(map[string]int64),
			ProviderWeights:    make(map[string]float64),
			SelectionHistory:   make([]*SelectionEvent, 0),
			PerformanceMetrics: make(map[string]*ProviderMetrics),
			Timestamp:          time.Now(),
		},
	}

	// Initialize default weights for existing providers
	lb.initializeDefaultWeights()

	return lb
}

func (lb *loadBalancerImpl) initializeDefaultWeights() {
	providers := lb.registry.GetAllProviders()

	// Set default weights based on provider capabilities and type
	defaultWeights := map[string]float64{
		"openai":                0.4, // High weight for established provider
		"ollama":                0.3, // Medium weight for local model
		"anthropic_claude":      0.1, // Low weight - new provider
		"google_gemini":         0.1, // Low weight - new provider
		"huggingface_inference": 0.1, // Low weight - experimental
	}

	for _, provider := range providers {
		name := provider.GetName()
		if weight, exists := defaultWeights[name]; exists {
			lb.providerWeights[name] = weight
		} else {
			// Default weight for unknown providers
			lb.providerWeights[name] = 0.05
		}

		// Initialize provider request counters
		lb.stats.ProviderRequests[name] = 0
		lb.stats.ProviderWeights[name] = lb.providerWeights[name]
	}
}

// SelectProvider selects the best provider based on the configured strategy
func (lb *loadBalancerImpl) SelectProvider(ctx context.Context, request *ProviderRequest) (ProviderInterface, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// If a specific provider is requested, try to use it
	if request.ProviderName != "" {
		provider, err := lb.registry.GetProvider(request.ProviderName)
		if err == nil && provider.IsHealthy(ctx) {
			lb.recordSelection(provider, request, "specific_request")
			return provider, nil
		}
		// Fall back to load balancing if specific provider is unavailable
	}

	providers := lb.getHealthyProviders(ctx)
	if len(providers) == 0 {
		return nil, fmt.Errorf("no healthy providers available")
	}

	var selectedProvider ProviderInterface
	var selectionReason string
	var loadFactor float64

	switch lb.strategy {
	case LoadBalancingRoundRobin:
		selectedProvider, selectionReason, loadFactor = lb.selectRoundRobin(providers)
	case LoadBalancingWeighted:
		selectedProvider, selectionReason, loadFactor = lb.selectWeighted(providers)
	case LoadBalancingLeastLoaded:
		selectedProvider, selectionReason, loadFactor = lb.selectLeastLoaded(ctx, providers)
	case LoadBalancingResponseTime:
		selectedProvider, selectionReason, loadFactor = lb.selectByResponseTime(ctx, providers)
	case LoadBalancingCost:
		selectedProvider, selectionReason, loadFactor = lb.selectByCost(ctx, providers, request)
	case LoadBalancingAvailability:
		selectedProvider, selectionReason, loadFactor = lb.selectByAvailability(ctx, providers)
	default:
		selectedProvider, selectionReason, loadFactor = lb.selectWeighted(providers)
	}

	if selectedProvider == nil {
		return nil, fmt.Errorf("failed to select provider using strategy %s", lb.strategy)
	}

	// Record the selection
	lb.recordSelection(selectedProvider, request, selectionReason)
	lb.updateSelectionStats(selectedProvider, loadFactor)

	return selectedProvider, nil
}

func (lb *loadBalancerImpl) getHealthyProviders(ctx context.Context) []ProviderInterface {
	allProviders := lb.registry.GetAllProviders()
	healthyProviders := make([]ProviderInterface, 0)

	for _, provider := range allProviders {
		if provider.IsHealthy(ctx) {
			healthyProviders = append(healthyProviders, provider)
		}
	}

	return healthyProviders
}

func (lb *loadBalancerImpl) selectRoundRobin(providers []ProviderInterface) (ProviderInterface, string, float64) {
	if len(providers) == 0 {
		return nil, "no_providers", 0
	}

	selectedProvider := providers[lb.roundRobinIndex%len(providers)]
	lb.roundRobinIndex++

	return selectedProvider, "round_robin", 1.0 / float64(len(providers))
}

func (lb *loadBalancerImpl) selectWeighted(providers []ProviderInterface) (ProviderInterface, string, float64) {
	if len(providers) == 0 {
		return nil, "no_providers", 0
	}

	// Calculate total weight
	totalWeight := 0.0
	providerWeights := make(map[string]float64)

	for _, provider := range providers {
		name := provider.GetName()
		weight := lb.providerWeights[name]
		if weight <= 0 {
			weight = 0.1 // Default minimum weight
		}
		providerWeights[name] = weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		// Fallback to round robin if no weights
		return lb.selectRoundRobin(providers)
	}

	// Generate random number and select based on weights
	random := rand.Float64() * totalWeight
	currentWeight := 0.0

	for _, provider := range providers {
		name := provider.GetName()
		weight := providerWeights[name]
		currentWeight += weight

		if random <= currentWeight {
			return provider, "weighted_selection", weight / totalWeight
		}
	}

	// Fallback to last provider if something goes wrong
	lastProvider := providers[len(providers)-1]
	name := lastProvider.GetName()
	return lastProvider, "weighted_fallback", providerWeights[name] / totalWeight
}

func (lb *loadBalancerImpl) selectLeastLoaded(ctx context.Context, providers []ProviderInterface) (ProviderInterface, string, float64) {
	if len(providers) == 0 {
		return nil, "no_providers", 0
	}

	var selectedProvider ProviderInterface
	lowestLoad := float64(100) // Start with 100% load

	for _, provider := range providers {
		status := provider.GetStatus(ctx)
		if status != nil && status.Load < lowestLoad {
			lowestLoad = status.Load
			selectedProvider = provider
		}
	}

	if selectedProvider == nil {
		// Fallback to first provider
		selectedProvider = providers[0]
		lowestLoad = 0
	}

	return selectedProvider, "least_loaded", 1.0 - lowestLoad/100.0
}

func (lb *loadBalancerImpl) selectByResponseTime(ctx context.Context, providers []ProviderInterface) (ProviderInterface, string, float64) {
	if len(providers) == 0 {
		return nil, "no_providers", 0
	}

	var selectedProvider ProviderInterface
	fastestTime := time.Duration(0)

	for _, provider := range providers {
		status := provider.GetStatus(ctx)
		if status != nil && (fastestTime == 0 || status.ResponseTime < fastestTime) {
			fastestTime = status.ResponseTime
			selectedProvider = provider
		}
	}

	if selectedProvider == nil {
		selectedProvider = providers[0]
		fastestTime = 100 * time.Millisecond
	}

	// Convert response time to selection factor (lower time = higher factor)
	factor := 1.0 / (float64(fastestTime.Milliseconds()) + 1)

	return selectedProvider, "response_time", factor
}

func (lb *loadBalancerImpl) selectByCost(ctx context.Context, providers []ProviderInterface, request *ProviderRequest) (ProviderInterface, string, float64) {
	if len(providers) == 0 {
		return nil, "no_providers", 0
	}

	var selectedProvider ProviderInterface
	lowestCost := float64(-1)

	for _, provider := range providers {
		estimate, err := provider.EstimateCost(request)
		if err != nil {
			continue
		}

		if lowestCost == -1 || estimate.TotalCost < lowestCost {
			lowestCost = estimate.TotalCost
			selectedProvider = provider
		}
	}

	if selectedProvider == nil {
		selectedProvider = providers[0]
		lowestCost = 0.01
	}

	// Convert cost to selection factor (lower cost = higher factor)
	factor := 1.0 / (lowestCost + 0.001)

	return selectedProvider, "lowest_cost", factor
}

func (lb *loadBalancerImpl) selectByAvailability(ctx context.Context, providers []ProviderInterface) (ProviderInterface, string, float64) {
	if len(providers) == 0 {
		return nil, "no_providers", 0
	}

	// Sort providers by error rate (ascending) and availability
	type providerScore struct {
		provider  ProviderInterface
		errorRate float64
		available bool
	}

	scores := make([]providerScore, 0, len(providers))

	for _, provider := range providers {
		status := provider.GetStatus(ctx)
		score := providerScore{
			provider:  provider,
			errorRate: 0.0,
			available: true,
		}

		if status != nil {
			score.errorRate = status.ErrorRate
			score.available = status.Available
		}

		scores = append(scores, score)
	}

	// Sort by availability (true first) then by error rate (ascending)
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].available != scores[j].available {
			return scores[i].available // true comes first
		}
		return scores[i].errorRate < scores[j].errorRate
	})

	selectedProvider := scores[0].provider
	factor := 1.0 - scores[0].errorRate // Higher availability = higher factor

	return selectedProvider, "highest_availability", factor
}

func (lb *loadBalancerImpl) recordSelection(provider ProviderInterface, request *ProviderRequest, reason string) {
	event := &SelectionEvent{
		ProviderName:    provider.GetName(),
		RequestID:       request.ID,
		SelectionReason: reason,
		LoadFactor:      0, // Will be set by updateSelectionStats
		ResponseTime:    0, // Will be updated after request completion
		Success:         true,
		Metadata: map[string]interface{}{
			"strategy":     lb.strategy,
			"request_type": request.RequestType,
			"model":        request.Model,
		},
		Timestamp: time.Now(),
	}

	// Keep last 1000 selection events
	if len(lb.selectionHistory) >= 1000 {
		lb.selectionHistory = lb.selectionHistory[1:]
	}

	lb.selectionHistory = append(lb.selectionHistory, event)
}

func (lb *loadBalancerImpl) updateSelectionStats(provider ProviderInterface, loadFactor float64) {
	providerName := provider.GetName()

	// Update stats
	lb.stats.TotalRequests++
	lb.stats.ProviderRequests[providerName]++

	// Update load factor in the latest selection event
	if len(lb.selectionHistory) > 0 {
		lb.selectionHistory[len(lb.selectionHistory)-1].LoadFactor = loadFactor
	}

	lb.stats.Timestamp = time.Now()
}

// UpdateProviderWeights updates the weights for weighted load balancing
func (lb *loadBalancerImpl) UpdateProviderWeights(weights map[string]float64) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Validate weights
	totalWeight := 0.0
	for name, weight := range weights {
		if weight < 0 {
			return fmt.Errorf("provider weight for %s cannot be negative", name)
		}
		totalWeight += weight
	}

	if totalWeight == 0 {
		return fmt.Errorf("total provider weights cannot be zero")
	}

	// Update weights
	for name, weight := range weights {
		lb.providerWeights[name] = weight
		lb.stats.ProviderWeights[name] = weight
	}

	return nil
}

// GetProviderWeights returns current provider weights
func (lb *loadBalancerImpl) GetProviderWeights() map[string]float64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	weights := make(map[string]float64)
	for name, weight := range lb.providerWeights {
		weights[name] = weight
	}

	return weights
}

// SetLoadBalancingStrategy sets the load balancing strategy
func (lb *loadBalancerImpl) SetLoadBalancingStrategy(strategy LoadBalancingStrategy) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Validate strategy
	validStrategies := map[LoadBalancingStrategy]bool{
		LoadBalancingRoundRobin:   true,
		LoadBalancingWeighted:     true,
		LoadBalancingLeastLoaded:  true,
		LoadBalancingResponseTime: true,
		LoadBalancingCost:         true,
		LoadBalancingAvailability: true,
	}

	if !validStrategies[strategy] {
		return fmt.Errorf("invalid load balancing strategy: %s", strategy)
	}

	lb.strategy = strategy
	lb.stats.Strategy = strategy

	return nil
}

// GetLoadBalancingStats returns current load balancing statistics
func (lb *loadBalancerImpl) GetLoadBalancingStats() *LoadBalancingStats {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	// Update performance metrics
	lb.updatePerformanceMetrics()

	// Create a copy of the stats to avoid concurrent access issues
	stats := &LoadBalancingStats{
		Strategy:           lb.stats.Strategy,
		TotalRequests:      lb.stats.TotalRequests,
		ProviderRequests:   make(map[string]int64),
		ProviderWeights:    make(map[string]float64),
		SelectionHistory:   make([]*SelectionEvent, len(lb.selectionHistory)),
		PerformanceMetrics: make(map[string]*ProviderMetrics),
		Timestamp:          time.Now(),
	}

	// Copy provider requests
	for name, count := range lb.stats.ProviderRequests {
		stats.ProviderRequests[name] = count
	}

	// Copy provider weights
	for name, weight := range lb.stats.ProviderWeights {
		stats.ProviderWeights[name] = weight
	}

	// Copy selection history (last 100 events)
	startIndex := 0
	if len(lb.selectionHistory) > 100 {
		startIndex = len(lb.selectionHistory) - 100
	}
	copy(stats.SelectionHistory, lb.selectionHistory[startIndex:])

	// Copy performance metrics
	for name, metrics := range lb.performanceMetrics {
		stats.PerformanceMetrics[name] = metrics
	}

	return stats
}

func (lb *loadBalancerImpl) updatePerformanceMetrics() {
	providers := lb.registry.GetAllProviders()

	for _, provider := range providers {
		multiMetrics, err := provider.GetMetrics(context.Background())
		if err == nil && multiMetrics != nil {
			// Convert MultiProviderMetrics to ProviderMetrics for load balancer stats
			convertedMetrics := &ProviderMetrics{
				Provider:       multiMetrics.ProviderName,
				RequestCount:   multiMetrics.TotalRequests,
				ErrorCount:     multiMetrics.FailedRequests,
				SuccessRate:    multiMetrics.SuccessRate,
				AverageLatency: float64(multiMetrics.AverageResponseTime.Milliseconds()),
				TokensUsed:     multiMetrics.TokensProcessed,
				TotalCost:      multiMetrics.TotalCost,
				LastRequest:    multiMetrics.Timestamp,
				Endpoints:      make(map[string]*EndpointStats),
				CircuitState:   "closed", // Default state
				HealthStatus:   multiMetrics.HealthStatus,
			}
			lb.performanceMetrics[provider.GetName()] = convertedMetrics
		}
	}
}

// UpdateProviderPerformance updates performance metrics for a provider after request completion
func (lb *loadBalancerImpl) UpdateProviderPerformance(providerName string, responseTime time.Duration, success bool) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Update the corresponding selection event
	for i := len(lb.selectionHistory) - 1; i >= 0; i-- {
		event := lb.selectionHistory[i]
		if event.ProviderName == providerName && event.ResponseTime == 0 {
			event.ResponseTime = responseTime
			event.Success = success
			break
		}
	}

	// Update performance-based weights if using response time strategy
	if lb.strategy == LoadBalancingResponseTime {
		lb.adjustWeightBasedOnPerformance(providerName, responseTime, success)
	}
}

func (lb *loadBalancerImpl) adjustWeightBasedOnPerformance(providerName string, responseTime time.Duration, success bool) {
	currentWeight := lb.providerWeights[providerName]

	// Adjust weight based on performance
	if success {
		// Faster response = higher weight
		if responseTime < 200*time.Millisecond {
			currentWeight *= 1.05 // Increase weight by 5%
		} else if responseTime > 1*time.Second {
			currentWeight *= 0.95 // Decrease weight by 5%
		}
	} else {
		// Failed request = lower weight
		currentWeight *= 0.9 // Decrease weight by 10%
	}

	// Keep weight within bounds
	if currentWeight > 1.0 {
		currentWeight = 1.0
	}
	if currentWeight < 0.01 {
		currentWeight = 0.01
	}

	lb.providerWeights[providerName] = currentWeight
	lb.stats.ProviderWeights[providerName] = currentWeight
}
