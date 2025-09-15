package external

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// allocationServiceImpl implements CostAllocator interface (Task 6.4)
type allocationServiceImpl struct {
	mu              sync.RWMutex
	allocations     map[string]*CostAllocation
	departments     map[string]*DepartmentCosts
	projects        map[string]*ProjectCosts
	allocationRules map[string]*AllocationRule
	costTracker     CostTracker
}

// AllocationRule defines how costs should be allocated
type AllocationRule struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	ServiceName    string                 `json:"service_name"`
	AllocationType string                 `json:"allocation_type"` // "percentage", "usage_based", "fixed"
	Rules          map[string]interface{} `json:"rules"`
	Priority       int                    `json:"priority"`
	Active         bool                   `json:"active"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// NewAllocationService creates a new cost allocation service
func NewAllocationService(costTracker CostTracker) CostAllocator {
	service := &allocationServiceImpl{
		allocations:     make(map[string]*CostAllocation),
		departments:     make(map[string]*DepartmentCosts),
		projects:        make(map[string]*ProjectCosts),
		allocationRules: make(map[string]*AllocationRule),
		costTracker:     costTracker,
	}

	// Initialize default allocation rules
	service.initializeDefaultRules()

	return service
}

// initializeDefaultRules creates default cost allocation rules
func (s *allocationServiceImpl) initializeDefaultRules() {
	defaultRules := []*AllocationRule{
		{
			ID:             "rule_dept_engineering",
			Name:           "Engineering Department Allocation",
			ServiceName:    "all",
			AllocationType: "percentage",
			Rules: map[string]interface{}{
				"engineering": 70.0,
				"product":     20.0,
				"operations":  10.0,
			},
			Priority:  1,
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:             "rule_project_kubechat",
			Name:           "KubeChat Project Allocation",
			ServiceName:    "all",
			AllocationType: "usage_based",
			Rules: map[string]interface{}{
				"kubechat_core":    50.0,
				"kubechat_ai":      30.0,
				"kubechat_infra":   15.0,
				"kubechat_testing": 5.0,
			},
			Priority:  2,
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:             "rule_openai_allocation",
			Name:           "OpenAI Service Specific Allocation",
			ServiceName:    "openai",
			AllocationType: "usage_based",
			Rules: map[string]interface{}{
				"ai_research":      40.0,
				"product_features": 35.0,
				"customer_support": 15.0,
				"testing_qa":       10.0,
			},
			Priority:  3,
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, rule := range defaultRules {
		s.allocationRules[rule.ID] = rule
	}

	// Initialize sample departments and projects
	s.initializeDepartmentsAndProjects()
}

func (s *allocationServiceImpl) initializeDepartmentsAndProjects() {
	// Initialize sample departments
	departments := []struct {
		name     string
		services []string
		budget   float64
	}{
		{"engineering", []string{"openai", "ollama"}, 800.0},
		{"product", []string{"openai"}, 200.0},
		{"operations", []string{"ollama"}, 100.0},
	}

	for _, dept := range departments {
		s.departments[dept.name] = &DepartmentCosts{
			Department:   dept.name,
			TotalCost:    0,
			ServiceCosts: make(map[string]float64),
			TimeRange: TimeRange{
				Start: time.Now().AddDate(0, -1, 0),
				End:   time.Now(),
			},
			Timestamp: time.Now(),
		}
	}

	// Initialize sample projects
	projects := []struct {
		id       string
		name     string
		services []string
		budget   float64
	}{
		{"proj_kubechat_core", "KubeChat Core", []string{"openai", "ollama"}, 600.0},
		{"proj_kubechat_ai", "KubeChat AI Features", []string{"openai"}, 300.0},
		{"proj_kubechat_infra", "KubeChat Infrastructure", []string{"ollama"}, 150.0},
		{"proj_kubechat_testing", "KubeChat Testing", []string{"openai", "ollama"}, 50.0},
	}

	for _, proj := range projects {
		s.projects[proj.id] = &ProjectCosts{
			ProjectID:    proj.id,
			ProjectName:  proj.name,
			TotalCost:    0,
			ServiceCosts: make(map[string]float64),
			TimeRange: TimeRange{
				Start: time.Now().AddDate(0, -1, 0),
				End:   time.Now(),
			},
			Timestamp: time.Now(),
		}
	}
}

// AllocateCosts allocates costs based on allocation rules (Task 6.4)
func (s *allocationServiceImpl) AllocateCosts(ctx context.Context, allocation *CostAllocation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate allocation
	if allocation.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	if allocation.Amount <= 0 {
		return fmt.Errorf("allocation amount must be greater than 0")
	}

	if allocation.Percentage < 0 || allocation.Percentage > 100 {
		return fmt.Errorf("allocation percentage must be between 0 and 100")
	}

	// Generate allocation ID if not provided
	allocationID := fmt.Sprintf("alloc_%s_%s_%v", allocation.ServiceName, allocation.Department, time.Now().Unix())
	s.allocations[allocationID] = allocation

	// Apply allocation to department
	if allocation.Department != "" {
		s.allocateToDepartment(allocation)
	}

	// Apply allocation to project
	if allocation.Project != "" {
		s.allocateToProject(allocation)
	}

	return nil
}

func (s *allocationServiceImpl) allocateToDepartment(allocation *CostAllocation) {
	dept, exists := s.departments[allocation.Department]
	if !exists {
		// Create new department
		dept = &DepartmentCosts{
			Department:   allocation.Department,
			TotalCost:    0,
			ServiceCosts: make(map[string]float64),
			TimeRange: TimeRange{
				Start: time.Now().AddDate(0, -1, 0),
				End:   time.Now(),
			},
			Timestamp: time.Now(),
		}
		s.departments[allocation.Department] = dept
	}

	// Add cost to department
	dept.TotalCost += allocation.Amount
	if dept.ServiceCosts == nil {
		dept.ServiceCosts = make(map[string]float64)
	}
	dept.ServiceCosts[allocation.ServiceName] += allocation.Amount
	dept.Timestamp = time.Now()
}

func (s *allocationServiceImpl) allocateToProject(allocation *CostAllocation) {
	proj, exists := s.projects[allocation.Project]
	if !exists {
		// Create new project
		proj = &ProjectCosts{
			ProjectID:    allocation.Project,
			ProjectName:  allocation.Project,
			TotalCost:    0,
			ServiceCosts: make(map[string]float64),
			TimeRange: TimeRange{
				Start: time.Now().AddDate(0, -1, 0),
				End:   time.Now(),
			},
			Timestamp: time.Now(),
		}
		s.projects[allocation.Project] = proj
	}

	// Add cost to project
	proj.TotalCost += allocation.Amount
	if proj.ServiceCosts == nil {
		proj.ServiceCosts = make(map[string]float64)
	}
	proj.ServiceCosts[allocation.ServiceName] += allocation.Amount
	proj.Timestamp = time.Now()
}

// GetAllocationSummary returns cost allocation summary (Task 6.4)
func (s *allocationServiceImpl) GetAllocationSummary(ctx context.Context, timeRange TimeRange) (*AllocationSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get total costs for the time range
	costSummary, err := s.costTracker.GetCostSummary(ctx, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost summary: %v", err)
	}

	// Apply allocation rules to calculate allocations
	departmentAllocations := s.calculateDepartmentAllocations(costSummary)
	projectAllocations := s.calculateProjectAllocations(costSummary)

	return &AllocationSummary{
		TimeRange:       timeRange,
		TotalCost:       costSummary.TotalCost,
		DepartmentCosts: departmentAllocations,
		ProjectCosts:    projectAllocations,
		Timestamp:       time.Now(),
	}, nil
}

func (s *allocationServiceImpl) calculateDepartmentAllocations(costSummary *CostSummary) map[string]float64 {
	departmentCosts := make(map[string]float64)

	// Find department allocation rule
	var deptRule *AllocationRule
	for _, rule := range s.allocationRules {
		if rule.AllocationType == "percentage" && rule.Active {
			deptRule = rule
			break
		}
	}

	if deptRule != nil {
		for dept, percentageInterface := range deptRule.Rules {
			if percentage, ok := percentageInterface.(float64); ok {
				allocation := costSummary.TotalCost * (percentage / 100.0)
				departmentCosts[dept] = math.Round(allocation*100) / 100
			}
		}
	} else {
		// Default equal allocation
		numDepartments := len(s.departments)
		if numDepartments > 0 {
			equalAllocation := costSummary.TotalCost / float64(numDepartments)
			for dept := range s.departments {
				departmentCosts[dept] = math.Round(equalAllocation*100) / 100
			}
		}
	}

	return departmentCosts
}

func (s *allocationServiceImpl) calculateProjectAllocations(costSummary *CostSummary) map[string]float64 {
	projectCosts := make(map[string]float64)

	// Find project allocation rule
	var projRule *AllocationRule
	for _, rule := range s.allocationRules {
		if rule.AllocationType == "usage_based" && rule.Active && rule.ServiceName == "all" {
			projRule = rule
			break
		}
	}

	if projRule != nil {
		for proj, percentageInterface := range projRule.Rules {
			if percentage, ok := percentageInterface.(float64); ok {
				allocation := costSummary.TotalCost * (percentage / 100.0)
				projectCosts[proj] = math.Round(allocation*100) / 100
			}
		}
	} else {
		// Default equal allocation
		numProjects := len(s.projects)
		if numProjects > 0 {
			equalAllocation := costSummary.TotalCost / float64(numProjects)
			for proj := range s.projects {
				projectCosts[proj] = math.Round(equalAllocation*100) / 100
			}
		}
	}

	return projectCosts
}

// GetDepartmentCosts returns costs for a specific department (Task 6.4)
func (s *allocationServiceImpl) GetDepartmentCosts(ctx context.Context, department string, timeRange TimeRange) (*DepartmentCosts, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get current department costs
	_, exists := s.departments[department]
	if !exists {
		return nil, fmt.Errorf("department %s not found", department)
	}

	// Calculate real-time costs based on allocation rules
	costSummary, err := s.costTracker.GetCostSummary(ctx, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost summary: %v", err)
	}

	// Apply department allocation rules
	departmentAllocations := s.calculateDepartmentAllocations(costSummary)

	allocatedCost, exists := departmentAllocations[department]
	if !exists {
		allocatedCost = 0
	}

	// Create updated department costs
	updatedDept := &DepartmentCosts{
		Department:   department,
		TotalCost:    allocatedCost,
		ServiceCosts: s.calculateDepartmentServiceCosts(department, costSummary),
		TimeRange:    timeRange,
		Timestamp:    time.Now(),
	}

	return updatedDept, nil
}

func (s *allocationServiceImpl) calculateDepartmentServiceCosts(department string, costSummary *CostSummary) map[string]float64 {
	serviceCosts := make(map[string]float64)

	// Find department allocation percentage
	var deptPercentage float64 = 0
	for _, rule := range s.allocationRules {
		if rule.AllocationType == "percentage" && rule.Active {
			if percentage, ok := rule.Rules[department].(float64); ok {
				deptPercentage = percentage / 100.0
				break
			}
		}
	}

	// Allocate service costs proportionally
	for service, cost := range costSummary.ServiceCosts {
		serviceCosts[service] = math.Round(cost*deptPercentage*100) / 100
	}

	return serviceCosts
}

// GetProjectCosts returns costs for a specific project (Task 6.4)
func (s *allocationServiceImpl) GetProjectCosts(ctx context.Context, projectID string, timeRange TimeRange) (*ProjectCosts, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get current project costs
	proj, exists := s.projects[projectID]
	if !exists {
		return nil, fmt.Errorf("project %s not found", projectID)
	}

	// Calculate real-time costs based on allocation rules
	costSummary, err := s.costTracker.GetCostSummary(ctx, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost summary: %v", err)
	}

	// Apply project allocation rules
	projectAllocations := s.calculateProjectAllocations(costSummary)

	allocatedCost, exists := projectAllocations[projectID]
	if !exists {
		allocatedCost = 0
	}

	// Create updated project costs
	updatedProj := &ProjectCosts{
		ProjectID:    projectID,
		ProjectName:  proj.ProjectName,
		TotalCost:    allocatedCost,
		ServiceCosts: s.calculateProjectServiceCosts(projectID, costSummary),
		TimeRange:    timeRange,
		Timestamp:    time.Now(),
	}

	return updatedProj, nil
}

func (s *allocationServiceImpl) calculateProjectServiceCosts(projectID string, costSummary *CostSummary) map[string]float64 {
	serviceCosts := make(map[string]float64)

	// Find project allocation percentage
	var projPercentage float64 = 0
	for _, rule := range s.allocationRules {
		if rule.AllocationType == "usage_based" && rule.Active && rule.ServiceName == "all" {
			if percentage, ok := rule.Rules[projectID].(float64); ok {
				projPercentage = percentage / 100.0
				break
			}
		}
	}

	// Allocate service costs proportionally
	for service, cost := range costSummary.ServiceCosts {
		serviceCosts[service] = math.Round(cost*projPercentage*100) / 100
	}

	return serviceCosts
}

// Additional allocation management methods
func (s *allocationServiceImpl) CreateAllocationRule(ctx context.Context, rule *AllocationRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rule.ID == "" {
		rule.ID = fmt.Sprintf("rule_%s_%v", rule.ServiceName, time.Now().Unix())
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	// Validate rule
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if rule.AllocationType == "" {
		return fmt.Errorf("allocation type is required")
	}

	// Validate percentages sum to 100 for percentage-based rules
	if rule.AllocationType == "percentage" {
		totalPercentage := 0.0
		for _, value := range rule.Rules {
			if percentage, ok := value.(float64); ok {
				totalPercentage += percentage
			}
		}
		if math.Abs(totalPercentage-100.0) > 0.01 {
			return fmt.Errorf("percentage allocation rules must sum to 100%%")
		}
	}

	s.allocationRules[rule.ID] = rule
	return nil
}

func (s *allocationServiceImpl) UpdateAllocationRule(ctx context.Context, ruleID string, rule *AllocationRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existingRule, exists := s.allocationRules[ruleID]
	if !exists {
		return fmt.Errorf("allocation rule %s not found", ruleID)
	}

	// Update fields
	existingRule.Name = rule.Name
	existingRule.ServiceName = rule.ServiceName
	existingRule.AllocationType = rule.AllocationType
	existingRule.Rules = rule.Rules
	existingRule.Priority = rule.Priority
	existingRule.Active = rule.Active
	existingRule.UpdatedAt = time.Now()

	return nil
}

func (s *allocationServiceImpl) DeleteAllocationRule(ctx context.Context, ruleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.allocationRules[ruleID]
	if !exists {
		return fmt.Errorf("allocation rule %s not found", ruleID)
	}

	delete(s.allocationRules, ruleID)
	return nil
}

func (s *allocationServiceImpl) GetAllocationRules(ctx context.Context) ([]*AllocationRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rules := make([]*AllocationRule, 0, len(s.allocationRules))
	for _, rule := range s.allocationRules {
		if rule.Active {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// Reporting and analytics methods
func (s *allocationServiceImpl) GetAllocationReport(ctx context.Context, timeRange TimeRange) (*AllocationReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get cost summary
	costSummary, err := s.costTracker.GetCostSummary(ctx, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost summary: %v", err)
	}

	// Calculate allocations
	departmentAllocations := s.calculateDepartmentAllocations(costSummary)
	projectAllocations := s.calculateProjectAllocations(costSummary)

	// Generate allocation breakdown
	breakdown := s.generateAllocationBreakdown(departmentAllocations, projectAllocations)

	return &AllocationReport{
		TimeRange:             timeRange,
		TotalCost:             costSummary.TotalCost,
		DepartmentAllocations: departmentAllocations,
		ProjectAllocations:    projectAllocations,
		AllocationBreakdown:   breakdown,
		RulesApplied:          s.getActiveRulesSummary(),
		Timestamp:             time.Now(),
	}, nil
}

func (s *allocationServiceImpl) generateAllocationBreakdown(deptAllocs, projAllocs map[string]float64) []AllocationBreakdownItem {
	breakdown := make([]AllocationBreakdownItem, 0)

	// Add department breakdown
	for dept, cost := range deptAllocs {
		breakdown = append(breakdown, AllocationBreakdownItem{
			Type:       "department",
			Name:       dept,
			Amount:     cost,
			Percentage: 0, // Will be calculated based on total
		})
	}

	// Add project breakdown
	for proj, cost := range projAllocs {
		breakdown = append(breakdown, AllocationBreakdownItem{
			Type:       "project",
			Name:       proj,
			Amount:     cost,
			Percentage: 0, // Will be calculated based on total
		})
	}

	return breakdown
}

func (s *allocationServiceImpl) getActiveRulesSummary() []string {
	summary := make([]string, 0)
	for _, rule := range s.allocationRules {
		if rule.Active {
			summary = append(summary, fmt.Sprintf("%s (%s)", rule.Name, rule.AllocationType))
		}
	}
	return summary
}

// Additional data structures for allocation reporting
type AllocationReport struct {
	TimeRange             TimeRange                 `json:"time_range"`
	TotalCost             float64                   `json:"total_cost"`
	DepartmentAllocations map[string]float64        `json:"department_allocations"`
	ProjectAllocations    map[string]float64        `json:"project_allocations"`
	AllocationBreakdown   []AllocationBreakdownItem `json:"allocation_breakdown"`
	RulesApplied          []string                  `json:"rules_applied"`
	Timestamp             time.Time                 `json:"timestamp"`
}

type AllocationBreakdownItem struct {
	Type       string  `json:"type"` // "department", "project", "service"
	Name       string  `json:"name"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}
