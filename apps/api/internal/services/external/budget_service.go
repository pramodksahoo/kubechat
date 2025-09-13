package external

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// budgetServiceImpl implements BudgetManager interface (Task 6.2)
type budgetServiceImpl struct {
	mu           sync.RWMutex
	budgets      map[string]*Budget
	budgetStatus map[string]*BudgetStatus
	alerts       []*BudgetAlert
	costTracker  CostTracker
}

// NewBudgetService creates a new budget management service
func NewBudgetService(costTracker CostTracker) BudgetManager {
	service := &budgetServiceImpl{
		budgets:      make(map[string]*Budget),
		budgetStatus: make(map[string]*BudgetStatus),
		alerts:       make([]*BudgetAlert, 0),
		costTracker:  costTracker,
	}

	// Initialize default budgets
	service.initializeDefaultBudgets()

	return service
}

// initializeDefaultBudgets creates default budgets for all services
func (s *budgetServiceImpl) initializeDefaultBudgets() {
	defaultBudgets := []*Budget{
		{
			ID:              "budget_openai_default",
			Name:            "OpenAI Default Budget",
			ServiceName:     "openai",
			Department:      "engineering",
			Project:         "kubechat",
			MonthlyLimit:    500.0,
			DailyLimit:      20.0,
			HourlyLimit:     2.0,
			AlertThresholds: []float64{50.0, 75.0, 90.0, 95.0},
			AutoActions: map[float64]string{
				90.0:  "alert_management",
				95.0:  "rate_limit",
				100.0: "suspend_service",
			},
			Currency:  "USD",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:              "budget_ollama_default",
			Name:            "Ollama Default Budget",
			ServiceName:     "ollama",
			Department:      "engineering",
			Project:         "kubechat",
			MonthlyLimit:    100.0, // Lower limit for local model
			DailyLimit:      5.0,
			HourlyLimit:     1.0,
			AlertThresholds: []float64{70.0, 85.0, 95.0},
			AutoActions: map[float64]string{
				85.0: "alert_management",
				95.0: "rate_limit",
			},
			Currency:  "USD",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, budget := range defaultBudgets {
		s.budgets[budget.ID] = budget
		s.budgetStatus[budget.ServiceName] = &BudgetStatus{
			BudgetID:       budget.ID,
			ServiceName:    budget.ServiceName,
			CurrentSpend:   0,
			MonthlyLimit:   budget.MonthlyLimit,
			DailyLimit:     budget.DailyLimit,
			HourlyLimit:    budget.HourlyLimit,
			UtilizationPct: 0,
			Status:         "under_budget",
			DaysRemaining:  s.getDaysRemainingInMonth(),
			ProjectedSpend: 0,
			Timestamp:      time.Now(),
		}
	}
}

func (s *budgetServiceImpl) getDaysRemainingInMonth() int {
	now := time.Now()
	year, month, _ := now.Date()
	nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, now.Location())
	return int(nextMonth.Sub(now).Hours() / 24)
}

// SetBudget creates or updates a budget (Task 6.2)
func (s *budgetServiceImpl) SetBudget(ctx context.Context, budget *Budget) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if budget.ID == "" {
		budget.ID = fmt.Sprintf("budget_%s_%v", budget.ServiceName, time.Now().Unix())
	}

	budget.UpdatedAt = time.Now()
	if budget.CreatedAt.IsZero() {
		budget.CreatedAt = time.Now()
	}

	// Validate budget limits
	if budget.MonthlyLimit <= 0 {
		return fmt.Errorf("monthly limit must be greater than 0")
	}

	if budget.DailyLimit > budget.MonthlyLimit {
		return fmt.Errorf("daily limit cannot exceed monthly limit")
	}

	s.budgets[budget.ID] = budget

	// Initialize or update budget status
	status := &BudgetStatus{
		BudgetID:       budget.ID,
		ServiceName:    budget.ServiceName,
		CurrentSpend:   0,
		MonthlyLimit:   budget.MonthlyLimit,
		DailyLimit:     budget.DailyLimit,
		HourlyLimit:    budget.HourlyLimit,
		UtilizationPct: 0,
		Status:         "under_budget",
		DaysRemaining:  s.getDaysRemainingInMonth(),
		ProjectedSpend: 0,
		Recommendations: []string{
			"Monitor usage patterns for optimization opportunities",
			"Consider implementing automated cost controls",
		},
		Timestamp: time.Now(),
	}

	s.budgetStatus[budget.ServiceName] = status

	return nil
}

// GetBudget retrieves a budget by ID (Task 6.2)
func (s *budgetServiceImpl) GetBudget(ctx context.Context, budgetID string) (*Budget, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	budget, exists := s.budgets[budgetID]
	if !exists {
		return nil, fmt.Errorf("budget with ID %s not found", budgetID)
	}

	return budget, nil
}

// CheckBudgetStatus checks current budget status for a service (Task 6.2)
func (s *budgetServiceImpl) CheckBudgetStatus(ctx context.Context, serviceName string) (*BudgetStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, exists := s.budgetStatus[serviceName]
	if !exists {
		return nil, fmt.Errorf("no budget status found for service %s", serviceName)
	}

	// Update current spending from cost tracker
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	timeRange := TimeRange{Start: monthStart, End: now}

	if s.costTracker != nil {
		serviceCosts, err := s.costTracker.GetServiceCosts(ctx, serviceName, timeRange)
		if err == nil && serviceCosts != nil {
			status.CurrentSpend = serviceCosts.TotalCost
			status.UtilizationPct = (status.CurrentSpend / status.MonthlyLimit) * 100

			// Update status based on utilization
			if status.UtilizationPct >= 100 {
				status.Status = "over_budget"
			} else if status.UtilizationPct >= 90 {
				status.Status = "warning"
			} else {
				status.Status = "under_budget"
			}

			// Calculate projected spend
			daysInMonth := s.getDaysInCurrentMonth()
			daysPassed := now.Day()
			if daysPassed > 0 {
				dailyAverage := status.CurrentSpend / float64(daysPassed)
				status.ProjectedSpend = dailyAverage * float64(daysInMonth)
			}

			// Update recommendations based on status
			status.Recommendations = s.generateRecommendations(status)
		}
	}

	status.Timestamp = time.Now()
	s.budgetStatus[serviceName] = status

	return status, nil
}

func (s *budgetServiceImpl) getDaysInCurrentMonth() int {
	now := time.Now()
	year, month, _ := now.Date()
	nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, now.Location())
	return nextMonth.AddDate(0, 0, -1).Day()
}

func (s *budgetServiceImpl) generateRecommendations(status *BudgetStatus) []string {
	recommendations := make([]string, 0)

	if status.UtilizationPct >= 95 {
		recommendations = append(recommendations,
			"URGENT: Budget critically exceeded - implement immediate cost controls",
			"Review and potentially suspend non-essential operations",
			"Consider emergency budget increase if justified")
	} else if status.UtilizationPct >= 90 {
		recommendations = append(recommendations,
			"Budget nearing limit - enable automated rate limiting",
			"Review recent usage patterns for anomalies",
			"Consider optimizing model selection and token usage")
	} else if status.UtilizationPct >= 75 {
		recommendations = append(recommendations,
			"Proactively monitor usage to avoid budget overrun",
			"Implement usage optimization strategies",
			"Review cost allocation and departmental usage")
	} else {
		recommendations = append(recommendations,
			"Budget usage is healthy - continue monitoring",
			"Consider analyzing usage patterns for optimization opportunities")
	}

	// Add projected spend recommendations
	if status.ProjectedSpend > status.MonthlyLimit*1.1 {
		recommendations = append(recommendations,
			fmt.Sprintf("Projected monthly spend ($%.2f) exceeds budget by %.1f%%",
				status.ProjectedSpend, ((status.ProjectedSpend/status.MonthlyLimit)-1)*100))
	}

	return recommendations
}

// GetBudgetAlerts retrieves all budget alerts (Task 6.2)
func (s *budgetServiceImpl) GetBudgetAlerts(ctx context.Context) ([]*BudgetAlert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Update alerts with recent budget status
	for serviceName := range s.budgetStatus {
		s.checkAndCreateAlerts(serviceName)
	}

	return s.alerts, nil
}

func (s *budgetServiceImpl) checkAndCreateAlerts(serviceName string) {
	status, exists := s.budgetStatus[serviceName]
	if !exists {
		return
	}

	budget := s.findBudgetByServiceName(serviceName)
	if budget == nil {
		return
	}

	// Check each threshold
	for _, threshold := range budget.AlertThresholds {
		if status.UtilizationPct >= threshold {
			// Check if alert already exists for this threshold
			if !s.alertExistsForThreshold(budget.ID, threshold) {
				alert := &BudgetAlert{
					ID:           fmt.Sprintf("alert_%s_%.0f_%v", budget.ID, threshold, time.Now().Unix()),
					BudgetID:     budget.ID,
					ServiceName:  serviceName,
					AlertType:    "threshold",
					Severity:     s.getSeverityForThreshold(threshold),
					Message:      fmt.Sprintf("Service %s has reached %.1f%% of monthly budget limit ($%.2f)", serviceName, threshold, status.MonthlyLimit),
					Threshold:    threshold,
					CurrentValue: status.UtilizationPct,
					ActionTaken:  s.getActionForThreshold(budget, threshold),
					Acknowledged: false,
					CreatedAt:    time.Now(),
				}
				s.alerts = append(s.alerts, alert)
			}
		}
	}
}

func (s *budgetServiceImpl) findBudgetByServiceName(serviceName string) *Budget {
	for _, budget := range s.budgets {
		if budget.ServiceName == serviceName && budget.Active {
			return budget
		}
	}
	return nil
}

func (s *budgetServiceImpl) alertExistsForThreshold(budgetID string, threshold float64) bool {
	for _, alert := range s.alerts {
		if alert.BudgetID == budgetID && alert.Threshold == threshold && !alert.Acknowledged {
			return true
		}
	}
	return false
}

func (s *budgetServiceImpl) getSeverityForThreshold(threshold float64) string {
	if threshold >= 95 {
		return "critical"
	} else if threshold >= 90 {
		return "high"
	} else if threshold >= 75 {
		return "medium"
	}
	return "low"
}

func (s *budgetServiceImpl) getActionForThreshold(budget *Budget, threshold float64) string {
	if action, exists := budget.AutoActions[threshold]; exists {
		return action
	}
	return "alert_only"
}

// UpdateBudgetLimits updates budget limits for a service (Task 6.2)
func (s *budgetServiceImpl) UpdateBudgetLimits(ctx context.Context, serviceName string, limits *BudgetLimits) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	budget := s.findBudgetByServiceName(serviceName)
	if budget == nil {
		return fmt.Errorf("no active budget found for service %s", serviceName)
	}

	// Validate new limits
	if limits.MonthlyLimit <= 0 {
		return fmt.Errorf("monthly limit must be greater than 0")
	}

	if limits.DailyLimit > limits.MonthlyLimit {
		return fmt.Errorf("daily limit cannot exceed monthly limit")
	}

	// Update budget
	budget.MonthlyLimit = limits.MonthlyLimit
	budget.DailyLimit = limits.DailyLimit
	budget.HourlyLimit = limits.HourlyLimit
	budget.UpdatedAt = time.Now()

	// Update budget status
	if status, exists := s.budgetStatus[serviceName]; exists {
		status.MonthlyLimit = limits.MonthlyLimit
		status.DailyLimit = limits.DailyLimit
		status.HourlyLimit = limits.HourlyLimit
		status.UtilizationPct = (status.CurrentSpend / status.MonthlyLimit) * 100
		status.Timestamp = time.Now()
	}

	return nil
}

// Additional budget management methods
func (s *budgetServiceImpl) GetAllBudgets(ctx context.Context) ([]*Budget, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	budgets := make([]*Budget, 0, len(s.budgets))
	for _, budget := range s.budgets {
		if budget.Active {
			budgets = append(budgets, budget)
		}
	}

	return budgets, nil
}

func (s *budgetServiceImpl) DeleteBudget(ctx context.Context, budgetID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	budget, exists := s.budgets[budgetID]
	if !exists {
		return fmt.Errorf("budget with ID %s not found", budgetID)
	}

	// Mark as inactive instead of deleting to preserve history
	budget.Active = false
	budget.UpdatedAt = time.Now()

	// Remove budget status
	delete(s.budgetStatus, budget.ServiceName)

	return nil
}

func (s *budgetServiceImpl) AcknowledgeAlert(ctx context.Context, alertID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, alert := range s.alerts {
		if alert.ID == alertID {
			alert.Acknowledged = true
			return nil
		}
	}

	return fmt.Errorf("alert with ID %s not found", alertID)
}

// Budget reporting methods
func (s *budgetServiceImpl) GetBudgetSummary(ctx context.Context) (*BudgetSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalBudget := 0.0
	totalSpend := 0.0
	serviceSummaries := make([]ServiceBudgetSummary, 0)

	for _, budget := range s.budgets {
		if !budget.Active {
			continue
		}

		totalBudget += budget.MonthlyLimit

		if status, exists := s.budgetStatus[budget.ServiceName]; exists {
			totalSpend += status.CurrentSpend
			serviceSummaries = append(serviceSummaries, ServiceBudgetSummary{
				ServiceName:    budget.ServiceName,
				BudgetLimit:    budget.MonthlyLimit,
				CurrentSpend:   status.CurrentSpend,
				UtilizationPct: status.UtilizationPct,
				Status:         status.Status,
			})
		}
	}

	overallUtilization := 0.0
	if totalBudget > 0 {
		overallUtilization = (totalSpend / totalBudget) * 100
	}

	return &BudgetSummary{
		TotalBudget:        totalBudget,
		TotalSpend:         totalSpend,
		OverallUtilization: overallUtilization,
		ServiceSummaries:   serviceSummaries,
		AlertCount:         len(s.getUnacknowledgedAlerts()),
		Timestamp:          time.Now(),
	}, nil
}

func (s *budgetServiceImpl) getUnacknowledgedAlerts() []*BudgetAlert {
	alerts := make([]*BudgetAlert, 0)
	for _, alert := range s.alerts {
		if !alert.Acknowledged {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// Additional data structures for budget management
type BudgetSummary struct {
	TotalBudget        float64                `json:"total_budget"`
	TotalSpend         float64                `json:"total_spend"`
	OverallUtilization float64                `json:"overall_utilization"`
	ServiceSummaries   []ServiceBudgetSummary `json:"service_summaries"`
	AlertCount         int                    `json:"alert_count"`
	Timestamp          time.Time              `json:"timestamp"`
}

type ServiceBudgetSummary struct {
	ServiceName    string  `json:"service_name"`
	BudgetLimit    float64 `json:"budget_limit"`
	CurrentSpend   float64 `json:"current_spend"`
	UtilizationPct float64 `json:"utilization_pct"`
	Status         string  `json:"status"`
}
