package external

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// optimizationServiceImpl implements CostOptimizer interface (Task 6.3)
type optimizationServiceImpl struct {
	mu              sync.RWMutex
	costTracker     CostTracker
	patterns        map[string]*UsagePattern
	recommendations map[string][]*OptimizationRecommendation
	analyses        []*UsageAnalysis
}

// NewOptimizationService creates a new cost optimization service
func NewOptimizationService(costTracker CostTracker) CostOptimizer {
	service := &optimizationServiceImpl{
		costTracker:     costTracker,
		patterns:        make(map[string]*UsagePattern),
		recommendations: make(map[string][]*OptimizationRecommendation),
		analyses:        make([]*UsageAnalysis, 0),
	}

	// Initialize with sample optimization recommendations
	service.initializeRecommendations()

	return service
}

// initializeRecommendations creates initial optimization recommendations
func (s *optimizationServiceImpl) initializeRecommendations() {
	openaiRecs := []*OptimizationRecommendation{
		{
			ID:              "opt_openai_001",
			ServiceName:     "openai",
			Category:        "model_optimization",
			Priority:        "high",
			Title:           "Switch to GPT-3.5-Turbo for Simple Tasks",
			Description:     "Analysis shows 60% of requests could use GPT-3.5-Turbo instead of GPT-4, reducing costs by 95%",
			PotentialSaving: 450.0,
			Implementation:  "Implement request classification to route simple queries to GPT-3.5-Turbo",
			Effort:          "medium",
			Timeline:        "1-2 weeks",
			RiskLevel:       "low",
			CreatedAt:       time.Now(),
		},
		{
			ID:              "opt_openai_002",
			ServiceName:     "openai",
			Category:        "token_optimization",
			Priority:        "medium",
			Title:           "Implement Response Caching",
			Description:     "Cache responses for similar queries to reduce API calls by 30%",
			PotentialSaving: 150.0,
			Implementation:  "Implement semantic similarity matching and response caching",
			Effort:          "high",
			Timeline:        "3-4 weeks",
			RiskLevel:       "low",
			CreatedAt:       time.Now(),
		},
		{
			ID:              "opt_openai_003",
			ServiceName:     "openai",
			Category:        "usage_pattern",
			Priority:        "medium",
			Title:           "Optimize Token Usage",
			Description:     "Reduce average tokens per request by 20% through prompt optimization",
			PotentialSaving: 100.0,
			Implementation:  "Review and optimize prompts, implement token counting and limits",
			Effort:          "medium",
			Timeline:        "2-3 weeks",
			RiskLevel:       "low",
			CreatedAt:       time.Now(),
		},
	}

	ollamaRecs := []*OptimizationRecommendation{
		{
			ID:              "opt_ollama_001",
			ServiceName:     "ollama",
			Category:        "infrastructure",
			Priority:        "low",
			Title:           "Optimize Local Model Performance",
			Description:     "Improve response times to reduce compute costs and resource usage",
			PotentialSaving: 25.0,
			Implementation:  "Optimize model loading, implement model quantization if available",
			Effort:          "high",
			Timeline:        "4-6 weeks",
			RiskLevel:       "medium",
			CreatedAt:       time.Now(),
		},
	}

	s.recommendations["openai"] = openaiRecs
	s.recommendations["ollama"] = ollamaRecs
}

// AnalyzeUsagePatterns analyzes usage patterns for optimization opportunities (Task 6.3)
func (s *optimizationServiceImpl) AnalyzeUsagePatterns(ctx context.Context, timeRange TimeRange) (*UsageAnalysis, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get usage metrics from cost tracker
	usageMetrics, err := s.costTracker.GetUsageMetrics(ctx, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage metrics: %v", err)
	}

	// Analyze patterns
	patterns := s.analyzePatterns(usageMetrics, timeRange)
	inefficiencies := s.detectInefficiencies(usageMetrics)
	trends := s.calculateTrends(usageMetrics)
	seasonality := s.detectSeasonality(timeRange)

	analysis := &UsageAnalysis{
		TimeRange:       timeRange,
		Patterns:        patterns,
		Inefficiencies:  inefficiencies,
		Trends:          trends,
		Seasonality:     seasonality,
		Recommendations: s.generateAnalysisRecommendations(patterns, inefficiencies),
		Timestamp:       time.Now(),
	}

	s.analyses = append(s.analyses, analysis)
	return analysis, nil
}

func (s *optimizationServiceImpl) analyzePatterns(metrics *UsageMetrics, timeRange TimeRange) []UsagePattern {
	patterns := make([]UsagePattern, 0)

	// Peak usage pattern
	if metrics.PeakUsageHour >= 0 {
		patterns = append(patterns, UsagePattern{
			PatternType:     "peak_usage",
			Description:     fmt.Sprintf("Peak usage occurs at hour %d", metrics.PeakUsageHour),
			Frequency:       "daily",
			ImpactScore:     8.5,
			CostImplication: 150.0,
		})
	}

	// Service usage distribution pattern
	if len(metrics.UsageByService) > 1 {
		var dominantService string
		var maxUsage int64 = 0
		for service, usage := range metrics.UsageByService {
			if usage > maxUsage {
				maxUsage = usage
				dominantService = service
			}
		}

		if maxUsage > metrics.TotalRequests/2 {
			patterns = append(patterns, UsagePattern{
				PatternType:     "service_dominance",
				Description:     fmt.Sprintf("Service %s accounts for %d%% of total usage", dominantService, (maxUsage*100)/metrics.TotalRequests),
				Frequency:       "continuous",
				ImpactScore:     7.0,
				CostImplication: 200.0,
			})
		}
	}

	// Model usage efficiency pattern
	for model, usage := range metrics.UsageByModel {
		if usage > 0 {
			patterns = append(patterns, UsagePattern{
				PatternType:     "model_usage",
				Description:     fmt.Sprintf("Model %s usage: %d requests", model, usage),
				Frequency:       "continuous",
				ImpactScore:     6.0,
				CostImplication: 75.0,
			})
		}
	}

	return patterns
}

func (s *optimizationServiceImpl) detectInefficiencies(metrics *UsageMetrics) []Inefficiency {
	inefficiencies := make([]Inefficiency, 0)

	// High token usage inefficiency
	if metrics.AverageTokens > 2000 {
		inefficiencies = append(inefficiencies, Inefficiency{
			Type:            "high_token_usage",
			ServiceName:     "openai",
			Description:     fmt.Sprintf("Average token usage (%.0f) is higher than recommended threshold (2000)", metrics.AverageTokens),
			WastedCost:      100.0,
			PotentialSaving: 80.0,
			Recommendation:  "Implement prompt optimization and token limits",
		})
	}

	// Model selection inefficiency
	for service, usage := range metrics.UsageByService {
		if service == "openai" && usage > int64(float64(metrics.TotalRequests)*0.7) {
			inefficiencies = append(inefficiencies, Inefficiency{
				Type:            "model_selection",
				ServiceName:     service,
				Description:     "Heavy reliance on expensive OpenAI models for potentially simple tasks",
				WastedCost:      300.0,
				PotentialSaving: 250.0,
				Recommendation:  "Implement request classification to route simple queries to cheaper models",
			})
		}
	}

	// Duplicate request inefficiency
	if metrics.TotalRequests > 0 {
		estimatedDuplicates := float64(metrics.TotalRequests) * 0.15 // Assume 15% duplication
		if estimatedDuplicates > 100 {
			inefficiencies = append(inefficiencies, Inefficiency{
				Type:            "duplicate_requests",
				ServiceName:     "all",
				Description:     fmt.Sprintf("Estimated %.0f duplicate or similar requests detected", estimatedDuplicates),
				WastedCost:      50.0,
				PotentialSaving: 40.0,
				Recommendation:  "Implement request deduplication and caching mechanisms",
			})
		}
	}

	return inefficiencies
}

func (s *optimizationServiceImpl) calculateTrends(metrics *UsageMetrics) map[string]float64 {
	trends := make(map[string]float64)

	// Mock trend calculations (in real implementation, this would analyze historical data)
	trends["usage_growth"] = 15.5       // 15.5% growth
	trends["cost_efficiency"] = -8.2    // 8.2% improvement in efficiency
	trends["token_optimization"] = 12.1 // 12.1% reduction in average tokens
	trends["request_frequency"] = 22.3  // 22.3% increase in request frequency

	return trends
}

func (s *optimizationServiceImpl) detectSeasonality(timeRange TimeRange) map[string]float64 {
	seasonality := make(map[string]float64)

	// Mock seasonality data (in real implementation, this would analyze usage patterns)
	seasonality["morning_peak"] = 1.8    // 80% higher usage in morning
	seasonality["weekend_drop"] = -0.4   // 40% lower usage on weekends
	seasonality["month_end_spike"] = 1.3 // 30% higher usage at month end

	return seasonality
}

func (s *optimizationServiceImpl) generateAnalysisRecommendations(patterns []UsagePattern, inefficiencies []Inefficiency) []string {
	recommendations := make([]string, 0)

	if len(patterns) > 0 {
		recommendations = append(recommendations, "Implement usage pattern-based optimization strategies")
	}

	if len(inefficiencies) > 0 {
		recommendations = append(recommendations, "Address identified cost inefficiencies to reduce waste")
	}

	recommendations = append(recommendations,
		"Consider implementing automated cost optimization rules",
		"Regular monitoring and analysis of usage patterns recommended",
		"Implement caching and request optimization mechanisms",
	)

	return recommendations
}

// GetOptimizationRecommendations returns optimization recommendations for a service (Task 6.3)
func (s *optimizationServiceImpl) GetOptimizationRecommendations(ctx context.Context, serviceName string) ([]*OptimizationRecommendation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	recommendations, exists := s.recommendations[serviceName]
	if !exists {
		return []*OptimizationRecommendation{}, nil
	}

	// Sort by priority and potential savings
	sortedRecs := make([]*OptimizationRecommendation, len(recommendations))
	copy(sortedRecs, recommendations)

	sort.Slice(sortedRecs, func(i, j int) bool {
		priorityOrder := map[string]int{"critical": 4, "high": 3, "medium": 2, "low": 1}
		iPriority := priorityOrder[sortedRecs[i].Priority]
		jPriority := priorityOrder[sortedRecs[j].Priority]

		if iPriority != jPriority {
			return iPriority > jPriority
		}
		return sortedRecs[i].PotentialSaving > sortedRecs[j].PotentialSaving
	})

	return sortedRecs, nil
}

// CalculateCostSavings calculates potential cost savings from recommendations (Task 6.3)
func (s *optimizationServiceImpl) CalculateCostSavings(ctx context.Context, recommendations []*OptimizationRecommendation) (*CostSavings, error) {
	if len(recommendations) == 0 {
		return &CostSavings{
			TotalSavings:       0,
			SavingsByService:   make(map[string]float64),
			SavingsBreakdown:   make([]SavingsBreakdown, 0),
			ImplementationCost: 0,
			NetSavings:         0,
			ROI:                0,
			Timestamp:          time.Now(),
		}, nil
	}

	totalSavings := 0.0
	savingsByService := make(map[string]float64)
	savingsByCategory := make(map[string]float64)
	implementationCost := 0.0

	for _, rec := range recommendations {
		totalSavings += rec.PotentialSaving
		savingsByService[rec.ServiceName] += rec.PotentialSaving
		savingsByCategory[rec.Category] += rec.PotentialSaving

		// Estimate implementation cost based on effort
		switch rec.Effort {
		case "low":
			implementationCost += 500.0
		case "medium":
			implementationCost += 2000.0
		case "high":
			implementationCost += 8000.0
		default:
			implementationCost += 1000.0
		}
	}

	// Create savings breakdown
	savingsBreakdown := make([]SavingsBreakdown, 0)
	for category, amount := range savingsByCategory {
		percentage := (amount / totalSavings) * 100
		savingsBreakdown = append(savingsBreakdown, SavingsBreakdown{
			Category:   category,
			Amount:     math.Round(amount*100) / 100,
			Percentage: math.Round(percentage*100) / 100,
		})
	}

	// Sort breakdown by amount
	sort.Slice(savingsBreakdown, func(i, j int) bool {
		return savingsBreakdown[i].Amount > savingsBreakdown[j].Amount
	})

	netSavings := totalSavings - implementationCost
	roi := 0.0
	if implementationCost > 0 {
		roi = (netSavings / implementationCost) * 100
	}

	return &CostSavings{
		TotalSavings:       math.Round(totalSavings*100) / 100,
		SavingsByService:   savingsByService,
		SavingsBreakdown:   savingsBreakdown,
		ImplementationCost: math.Round(implementationCost*100) / 100,
		NetSavings:         math.Round(netSavings*100) / 100,
		ROI:                math.Round(roi*100) / 100,
		Timestamp:          time.Now(),
	}, nil
}

// GetEfficiencyMetrics returns cost efficiency metrics (Task 6.3)
func (s *optimizationServiceImpl) GetEfficiencyMetrics(ctx context.Context) (*EfficiencyMetrics, error) {
	// Get current usage metrics
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	timeRange := TimeRange{Start: monthStart, End: now}

	usageMetrics, err := s.costTracker.GetUsageMetrics(ctx, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage metrics: %v", err)
	}

	costSummary, err := s.costTracker.GetCostSummary(ctx, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost summary: %v", err)
	}

	// Calculate efficiency metrics
	costPerToken := 0.0
	if usageMetrics.TotalTokens > 0 {
		costPerToken = costSummary.TotalCost / float64(usageMetrics.TotalTokens) * 1000 // Cost per 1K tokens
	}

	costPerRequest := 0.0
	if usageMetrics.TotalRequests > 0 {
		costPerRequest = costSummary.TotalCost / float64(usageMetrics.TotalRequests)
	}

	// Calculate utilization rate (mock calculation)
	utilizationRate := s.calculateUtilizationRate(usageMetrics)

	// Calculate efficiency score (0-100)
	efficiencyScore := s.calculateEfficiencyScore(costPerToken, costPerRequest, utilizationRate)

	// Calculate trends (mock data showing improvement)
	trends := map[string]float64{
		"cost_per_token_trend":   -5.2, // 5.2% improvement
		"cost_per_request_trend": -3.8, // 3.8% improvement
		"utilization_trend":      8.1,  // 8.1% improvement
		"efficiency_score_trend": 6.5,  // 6.5% improvement
	}

	return &EfficiencyMetrics{
		CostPerToken:    math.Round(costPerToken*10000) / 10000,
		CostPerRequest:  math.Round(costPerRequest*100) / 100,
		UtilizationRate: math.Round(utilizationRate*100) / 100,
		EfficiencyScore: math.Round(efficiencyScore*100) / 100,
		Trends:          trends,
		Timestamp:       time.Now(),
	}, nil
}

func (s *optimizationServiceImpl) calculateUtilizationRate(metrics *UsageMetrics) float64 {
	// Mock calculation - in real implementation this would consider:
	// - Resource allocation vs actual usage
	// - Peak capacity vs average usage
	// - Service availability vs requests

	if metrics.TotalRequests == 0 {
		return 0.0
	}

	// Simple utilization based on request distribution
	peakHourRequests := float64(metrics.TotalRequests) * 0.15 // Assume 15% in peak hour
	averageHourlyRequests := float64(metrics.TotalRequests) / 24.0

	utilization := averageHourlyRequests / peakHourRequests
	if utilization > 1.0 {
		utilization = 1.0
	}

	return utilization * 100
}

func (s *optimizationServiceImpl) calculateEfficiencyScore(costPerToken, costPerRequest, utilizationRate float64) float64 {
	// Efficiency score calculation (0-100)
	// Higher utilization = better efficiency
	// Lower costs = better efficiency

	utilizationScore := utilizationRate // Already 0-100

	// Cost efficiency score (inverse relationship)
	costScore := 100.0
	if costPerToken > 0.002 { // Threshold for good cost per token
		costScore -= (costPerToken - 0.002) * 10000
	}
	if costPerRequest > 0.01 { // Threshold for good cost per request
		costScore -= (costPerRequest - 0.01) * 1000
	}

	if costScore < 0 {
		costScore = 0
	}

	// Weighted average
	efficiencyScore := (utilizationScore * 0.4) + (costScore * 0.6)

	if efficiencyScore > 100 {
		efficiencyScore = 100
	}

	return efficiencyScore
}

// Additional optimization methods
func (s *optimizationServiceImpl) AddRecommendation(ctx context.Context, recommendation *OptimizationRecommendation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if recommendation.ID == "" {
		recommendation.ID = fmt.Sprintf("opt_%s_%v", recommendation.ServiceName, time.Now().Unix())
	}

	if recommendation.CreatedAt.IsZero() {
		recommendation.CreatedAt = time.Now()
	}

	if s.recommendations[recommendation.ServiceName] == nil {
		s.recommendations[recommendation.ServiceName] = make([]*OptimizationRecommendation, 0)
	}

	s.recommendations[recommendation.ServiceName] = append(s.recommendations[recommendation.ServiceName], recommendation)
	return nil
}

func (s *optimizationServiceImpl) GetAllRecommendations(ctx context.Context) (map[string][]*OptimizationRecommendation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modifications
	result := make(map[string][]*OptimizationRecommendation)
	for service, recs := range s.recommendations {
		result[service] = make([]*OptimizationRecommendation, len(recs))
		copy(result[service], recs)
	}

	return result, nil
}

func (s *optimizationServiceImpl) GetAnalysisHistory(ctx context.Context) ([]*UsageAnalysis, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy of analyses
	analyses := make([]*UsageAnalysis, len(s.analyses))
	copy(analyses, s.analyses)

	return analyses, nil
}
