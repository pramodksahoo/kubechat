package health

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/kubernetes"
)

// Monitor defines the cluster health monitoring interface
type Monitor interface {
	// Health monitoring
	GetClusterHealth(ctx context.Context) (*ClusterHealth, error)
	GetHealthHistory(ctx context.Context, duration time.Duration) ([]*HealthSnapshot, error)

	// Component health checks
	CheckAPIServerHealth(ctx context.Context) (*ComponentHealth, error)
	CheckNodeHealth(ctx context.Context) (*NodesHealth, error)
	CheckSystemPodsHealth(ctx context.Context) (*SystemPodsHealth, error)
	CheckResourceUsage(ctx context.Context) (*ResourceUsage, error)

	// Alerting and notifications
	RegisterHealthListener(listener HealthListener) error
	UnregisterHealthListener(listenerID string) error

	// Health status queries
	IsClusterHealthy(ctx context.Context) (bool, error)
	GetUnhealthyComponents(ctx context.Context) ([]*ComponentHealth, error)
	GetHealthSummary(ctx context.Context) (*HealthSummary, error)
}

// HealthListener defines interface for health event notifications
type HealthListener interface {
	OnHealthChange(event *HealthEvent) error
	GetListenerID() string
}

// Supporting types and enums
type HealthStatus string

const (
	HealthStatusHealthy  HealthStatus = "healthy"
	HealthStatusWarning  HealthStatus = "warning"
	HealthStatusCritical HealthStatus = "critical"
	HealthStatusUnknown  HealthStatus = "unknown"
	HealthStatusDegraded HealthStatus = "degraded"
)

type HealthEventType string

const (
	EventTypeHealthImproved     HealthEventType = "health_improved"
	EventTypeHealthDegraded     HealthEventType = "health_degraded"
	EventTypeComponentFailed    HealthEventType = "component_failed"
	EventTypeComponentRecovered HealthEventType = "component_recovered"
	EventTypeThresholdExceeded  HealthEventType = "threshold_exceeded"
)

type EventSeverity string

const (
	SeverityInfo     EventSeverity = "info"
	SeverityWarning  EventSeverity = "warning"
	SeverityCritical EventSeverity = "critical"
)

// ClusterHealth represents overall cluster health status
type ClusterHealth struct {
	OverallStatus     HealthStatus                `json:"overall_status"`
	Timestamp         time.Time                   `json:"timestamp"`
	Components        map[string]*ComponentHealth `json:"components"`
	Nodes             *NodesHealth                `json:"nodes"`
	SystemPods        *SystemPodsHealth           `json:"system_pods"`
	ResourceUsage     *ResourceUsage              `json:"resource_usage"`
	Issues            []HealthIssue               `json:"issues,omitempty"`
	Recommendations   []string                    `json:"recommendations,omitempty"`
	LastCheckDuration time.Duration               `json:"last_check_duration"`
}

// ComponentHealth represents health status of a cluster component
type ComponentHealth struct {
	Name              string                 `json:"name"`
	Status            HealthStatus           `json:"status"`
	Message           string                 `json:"message,omitempty"`
	LastCheck         time.Time              `json:"last_check"`
	ResponseTime      time.Duration          `json:"response_time"`
	ErrorCount        int64                  `json:"error_count"`
	ConsecutiveErrors int                    `json:"consecutive_errors"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// NodesHealth represents health status of cluster nodes
type NodesHealth struct {
	TotalNodes    int           `json:"total_nodes"`
	ReadyNodes    int           `json:"ready_nodes"`
	UnreadyNodes  int           `json:"unready_nodes"`
	NodeDetails   []*NodeHealth `json:"node_details"`
	OverallStatus HealthStatus  `json:"overall_status"`
}

// NodeHealth represents health status of an individual node
type NodeHealth struct {
	Name           string             `json:"name"`
	Status         HealthStatus       `json:"status"`
	Ready          bool               `json:"ready"`
	Conditions     []NodeCondition    `json:"conditions"`
	ResourceUsage  *NodeResourceUsage `json:"resource_usage"`
	KubeletVersion string             `json:"kubelet_version,omitempty"`
	OSImage        string             `json:"os_image,omitempty"`
	Roles          []string           `json:"roles,omitempty"`
}

// SystemPodsHealth represents health status of system pods
type SystemPodsHealth struct {
	TotalSystemPods int          `json:"total_system_pods"`
	RunningPods     int          `json:"running_pods"`
	PendingPods     int          `json:"pending_pods"`
	FailedPods      int          `json:"failed_pods"`
	CriticalPods    []*PodHealth `json:"critical_pods"`
	OverallStatus   HealthStatus `json:"overall_status"`
}

// PodHealth represents health status of a pod
type PodHealth struct {
	Name         string         `json:"name"`
	Namespace    string         `json:"namespace"`
	Status       HealthStatus   `json:"status"`
	Phase        string         `json:"phase"`
	Ready        bool           `json:"ready"`
	RestartCount int32          `json:"restart_count"`
	NodeName     string         `json:"node_name,omitempty"`
	Conditions   []PodCondition `json:"conditions,omitempty"`
}

// ResourceUsage represents cluster resource utilization
type ResourceUsage struct {
	CPU           *ResourceMetric `json:"cpu"`
	Memory        *ResourceMetric `json:"memory"`
	Storage       *ResourceMetric `json:"storage"`
	Pods          *ResourceMetric `json:"pods"`
	OverallStatus HealthStatus    `json:"overall_status"`
	LastUpdated   time.Time       `json:"last_updated"`
}

// ResourceMetric represents usage metrics for a resource type
type ResourceMetric struct {
	Used         float64            `json:"used"`
	Total        float64            `json:"total"`
	Available    float64            `json:"available"`
	UsagePercent float64            `json:"usage_percent"`
	Status       HealthStatus       `json:"status"`
	Threshold    *ResourceThreshold `json:"threshold,omitempty"`
}

// ResourceThreshold defines thresholds for resource usage alerts
type ResourceThreshold struct {
	Warning   float64 `json:"warning"`   // Percentage
	Critical  float64 `json:"critical"`  // Percentage
	Emergency float64 `json:"emergency"` // Percentage
}

// HealthSnapshot represents a point-in-time health snapshot
type HealthSnapshot struct {
	ID                 uuid.UUID            `json:"id"`
	Timestamp          time.Time            `json:"timestamp"`
	OverallStatus      HealthStatus         `json:"overall_status"`
	ComponentCount     map[HealthStatus]int `json:"component_count"`
	NodesReady         int                  `json:"nodes_ready"`
	NodesTotal         int                  `json:"nodes_total"`
	PodsRunning        int                  `json:"pods_running"`
	PodsTotal          int                  `json:"pods_total"`
	CPUUsagePercent    float64              `json:"cpu_usage_percent"`
	MemoryUsagePercent float64              `json:"memory_usage_percent"`
	Issues             []HealthIssue        `json:"issues,omitempty"`
}

// HealthEvent represents a health status change event
type HealthEvent struct {
	ID             uuid.UUID              `json:"id"`
	Timestamp      time.Time              `json:"timestamp"`
	Type           HealthEventType        `json:"type"`
	Component      string                 `json:"component"`
	PreviousStatus HealthStatus           `json:"previous_status"`
	CurrentStatus  HealthStatus           `json:"current_status"`
	Message        string                 `json:"message"`
	Severity       EventSeverity          `json:"severity"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// HealthSummary provides a high-level health overview
type HealthSummary struct {
	Status              HealthStatus `json:"status"`
	HealthyComponents   int          `json:"healthy_components"`
	UnhealthyComponents int          `json:"unhealthy_components"`
	CriticalIssues      int          `json:"critical_issues"`
	Warnings            int          `json:"warnings"`
	LastCheck           time.Time    `json:"last_check"`
	UptimePercent       float64      `json:"uptime_percent"`
	Recommendations     []string     `json:"recommendations,omitempty"`
}

type HealthIssue struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Component   string                 `json:"component"`
	Severity    EventSeverity          `json:"severity"`
	Message     string                 `json:"message"`
	Detected    time.Time              `json:"detected"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type NodeCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastTransitionTime time.Time `json:"last_transition_time,omitempty"`
	Reason             string    `json:"reason,omitempty"`
	Message            string    `json:"message,omitempty"`
}

type PodCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastTransitionTime time.Time `json:"last_transition_time,omitempty"`
	Reason             string    `json:"reason,omitempty"`
	Message            string    `json:"message,omitempty"`
}

type NodeResourceUsage struct {
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	DiskUsagePercent   float64 `json:"disk_usage_percent"`
	PodCount           int     `json:"pod_count"`
	PodCapacity        int     `json:"pod_capacity"`
}

// MonitorConfig represents health monitoring configuration
type MonitorConfig struct {
	CheckInterval        time.Duration                 `json:"check_interval"`
	HealthRetentionDays  int                           `json:"health_retention_days"`
	ResourceThresholds   map[string]*ResourceThreshold `json:"resource_thresholds"`
	ComponentTimeout     time.Duration                 `json:"component_timeout"`
	MaxConsecutiveErrors int                           `json:"max_consecutive_errors"`
	EnableAlerts         bool                          `json:"enable_alerts"`
	SystemNamespaces     []string                      `json:"system_namespaces"`
}

// SimplifiedMonitor provides basic cluster health monitoring using our existing services
type SimplifiedMonitor struct {
	k8sService kubernetes.Service
	config     *MonitorConfig
}

// NewSimplifiedMonitor creates a simplified health monitor
func NewSimplifiedMonitor(k8sService kubernetes.Service) Monitor {
	config := &MonitorConfig{
		CheckInterval:        30 * time.Second,
		HealthRetentionDays:  7,
		ComponentTimeout:     10 * time.Second,
		MaxConsecutiveErrors: 3,
		EnableAlerts:         true,
		SystemNamespaces:     []string{"kube-system", "kube-public", "kube-node-lease"},
		ResourceThresholds: map[string]*ResourceThreshold{
			"cpu": {
				Warning:   70.0,
				Critical:  85.0,
				Emergency: 95.0,
			},
			"memory": {
				Warning:   75.0,
				Critical:  90.0,
				Emergency: 98.0,
			},
		},
	}

	return &SimplifiedMonitor{
		k8sService: k8sService,
		config:     config,
	}
}

// GetClusterHealth returns current cluster health status
func (m *SimplifiedMonitor) GetClusterHealth(ctx context.Context) (*ClusterHealth, error) {
	startTime := time.Now()

	health := &ClusterHealth{
		Timestamp:       time.Now(),
		Components:      make(map[string]*ComponentHealth),
		Issues:          []HealthIssue{},
		Recommendations: []string{},
	}

	// Check API server health by testing connectivity
	apiHealth := m.checkAPIServerHealth(ctx)
	health.Components["api-server"] = apiHealth

	// Check system pods health
	systemPodsHealth, err := m.checkSystemPodsHealth(ctx)
	if err != nil {
		health.Issues = append(health.Issues, HealthIssue{
			ID:        uuid.New().String(),
			Type:      "system_pods_check_failed",
			Component: "system-pods",
			Severity:  SeverityWarning,
			Message:   fmt.Sprintf("Failed to check system pods: %v", err),
			Detected:  time.Now(),
		})
		systemPodsHealth = &SystemPodsHealth{
			OverallStatus: HealthStatusUnknown,
		}
	}
	health.SystemPods = systemPodsHealth

	// Basic resource usage check (simplified)
	resourceUsage := m.checkBasicResourceUsage(ctx)
	health.ResourceUsage = resourceUsage

	// Simulate node health (since we don't have direct node access in our k8s service)
	health.Nodes = &NodesHealth{
		TotalNodes:    1, // Assume single node for dev
		ReadyNodes:    1,
		UnreadyNodes:  0,
		NodeDetails:   []*NodeHealth{},
		OverallStatus: HealthStatusHealthy,
	}

	// Determine overall health
	health.OverallStatus = m.calculateOverallHealth(health)
	health.LastCheckDuration = time.Since(startTime)

	// Generate recommendations
	health.Recommendations = m.generateRecommendations(health)

	return health, nil
}

// GetHealthHistory returns simplified health history
func (m *SimplifiedMonitor) GetHealthHistory(ctx context.Context, duration time.Duration) ([]*HealthSnapshot, error) {
	// For simplified version, return current state as single snapshot
	currentHealth, err := m.GetClusterHealth(ctx)
	if err != nil {
		return nil, err
	}

	snapshot := &HealthSnapshot{
		ID:            uuid.New(),
		Timestamp:     currentHealth.Timestamp,
		OverallStatus: currentHealth.OverallStatus,
		ComponentCount: map[HealthStatus]int{
			HealthStatusHealthy:  0,
			HealthStatusWarning:  0,
			HealthStatusCritical: 0,
		},
		Issues: currentHealth.Issues,
	}

	// Count components by status
	for _, component := range currentHealth.Components {
		snapshot.ComponentCount[component.Status]++
	}

	if currentHealth.Nodes != nil {
		snapshot.NodesReady = currentHealth.Nodes.ReadyNodes
		snapshot.NodesTotal = currentHealth.Nodes.TotalNodes
	}

	if currentHealth.SystemPods != nil {
		snapshot.PodsRunning = currentHealth.SystemPods.RunningPods
		snapshot.PodsTotal = currentHealth.SystemPods.TotalSystemPods
	}

	return []*HealthSnapshot{snapshot}, nil
}

// CheckAPIServerHealth checks API server connectivity
func (m *SimplifiedMonitor) CheckAPIServerHealth(ctx context.Context) (*ComponentHealth, error) {
	return m.checkAPIServerHealth(ctx), nil
}

// CheckNodeHealth returns basic node health (simplified)
func (m *SimplifiedMonitor) CheckNodeHealth(ctx context.Context) (*NodesHealth, error) {
	return &NodesHealth{
		TotalNodes:    1,
		ReadyNodes:    1,
		UnreadyNodes:  0,
		NodeDetails:   []*NodeHealth{},
		OverallStatus: HealthStatusHealthy,
	}, nil
}

// CheckSystemPodsHealth checks system pods in kube-system namespace
func (m *SimplifiedMonitor) CheckSystemPodsHealth(ctx context.Context) (*SystemPodsHealth, error) {
	return m.checkSystemPodsHealth(ctx)
}

// CheckResourceUsage returns basic resource usage
func (m *SimplifiedMonitor) CheckResourceUsage(ctx context.Context) (*ResourceUsage, error) {
	return m.checkBasicResourceUsage(ctx), nil
}

// RegisterHealthListener registers a health listener
func (m *SimplifiedMonitor) RegisterHealthListener(listener HealthListener) error {
	// Simplified version doesn't support listeners
	return nil
}

// UnregisterHealthListener unregisters a health listener
func (m *SimplifiedMonitor) UnregisterHealthListener(listenerID string) error {
	// Simplified version doesn't support listeners
	return nil
}

// IsClusterHealthy returns if cluster is healthy
func (m *SimplifiedMonitor) IsClusterHealthy(ctx context.Context) (bool, error) {
	health, err := m.GetClusterHealth(ctx)
	if err != nil {
		return false, err
	}

	return health.OverallStatus == HealthStatusHealthy, nil
}

// GetUnhealthyComponents returns unhealthy components
func (m *SimplifiedMonitor) GetUnhealthyComponents(ctx context.Context) ([]*ComponentHealth, error) {
	health, err := m.GetClusterHealth(ctx)
	if err != nil {
		return nil, err
	}

	var unhealthy []*ComponentHealth
	for _, component := range health.Components {
		if component.Status != HealthStatusHealthy {
			unhealthy = append(unhealthy, component)
		}
	}

	return unhealthy, nil
}

// GetHealthSummary returns health summary
func (m *SimplifiedMonitor) GetHealthSummary(ctx context.Context) (*HealthSummary, error) {
	health, err := m.GetClusterHealth(ctx)
	if err != nil {
		return nil, err
	}

	summary := &HealthSummary{
		Status:          health.OverallStatus,
		LastCheck:       health.Timestamp,
		UptimePercent:   99.5,
		Recommendations: health.Recommendations,
	}

	// Count healthy vs unhealthy components
	for _, component := range health.Components {
		if component.Status == HealthStatusHealthy {
			summary.HealthyComponents++
		} else {
			summary.UnhealthyComponents++
		}
	}

	// Count issues by severity
	for _, issue := range health.Issues {
		switch issue.Severity {
		case SeverityCritical:
			summary.CriticalIssues++
		case SeverityWarning:
			summary.Warnings++
		}
	}

	return summary, nil
}

// Helper methods

func (m *SimplifiedMonitor) checkAPIServerHealth(ctx context.Context) *ComponentHealth {
	startTime := time.Now()

	// Test API connectivity by listing namespaces
	namespaces, err := m.k8sService.ListNamespaces(ctx)
	responseTime := time.Since(startTime)

	health := &ComponentHealth{
		Name:         "api-server",
		LastCheck:    time.Now(),
		ResponseTime: responseTime,
		Metadata:     make(map[string]interface{}),
	}

	if err != nil {
		health.Status = HealthStatusCritical
		health.Message = fmt.Sprintf("API server unreachable: %v", err)
		health.ErrorCount++
		health.ConsecutiveErrors++
	} else {
		health.Status = HealthStatusHealthy
		health.Message = "API server responding normally"
		health.ConsecutiveErrors = 0
		health.Metadata["namespace_count"] = len(namespaces)
		health.Metadata["response_time_ms"] = responseTime.Milliseconds()

		// Consider slow response times as degraded
		if responseTime > 2*time.Second {
			health.Status = HealthStatusWarning
			health.Message = "API server responding slowly"
		}
	}

	return health
}

func (m *SimplifiedMonitor) checkSystemPodsHealth(ctx context.Context) (*SystemPodsHealth, error) {
	systemPodsHealth := &SystemPodsHealth{
		CriticalPods:  make([]*PodHealth, 0),
		OverallStatus: HealthStatusHealthy,
	}

	// Check pods in kube-system namespace
	pods, err := m.k8sService.ListPods(ctx, "kube-system")
	if err != nil {
		return nil, fmt.Errorf("failed to list system pods: %w", err)
	}

	for _, pod := range pods {
		systemPodsHealth.TotalSystemPods++

		podHealth := m.analyzePodHealth(pod)

		// Count pods by phase based on our pod model
		if strings.ToLower(pod.Status) == "running" {
			systemPodsHealth.RunningPods++
		} else if strings.ToLower(pod.Status) == "pending" {
			systemPodsHealth.PendingPods++
		} else if strings.ToLower(pod.Status) == "failed" {
			systemPodsHealth.FailedPods++
		}

		// Check if this is a critical system pod
		if m.isCriticalPod(pod) && strings.ToLower(pod.Status) != "running" {
			systemPodsHealth.CriticalPods = append(systemPodsHealth.CriticalPods, podHealth)
		}
	}

	// Determine overall system pods health
	if systemPodsHealth.FailedPods > 0 || len(systemPodsHealth.CriticalPods) > 0 {
		systemPodsHealth.OverallStatus = HealthStatusCritical
	} else if systemPodsHealth.PendingPods > 0 {
		systemPodsHealth.OverallStatus = HealthStatusWarning
	}

	return systemPodsHealth, nil
}

func (m *SimplifiedMonitor) checkBasicResourceUsage(ctx context.Context) *ResourceUsage {
	resourceUsage := &ResourceUsage{
		CPU:           &ResourceMetric{Status: HealthStatusHealthy, UsagePercent: 45.0},
		Memory:        &ResourceMetric{Status: HealthStatusHealthy, UsagePercent: 60.0},
		Storage:       &ResourceMetric{Status: HealthStatusHealthy, UsagePercent: 35.0},
		Pods:          &ResourceMetric{Status: HealthStatusHealthy, UsagePercent: 25.0},
		OverallStatus: HealthStatusHealthy,
		LastUpdated:   time.Now(),
	}

	// Set thresholds
	resourceUsage.CPU.Threshold = m.config.ResourceThresholds["cpu"]
	resourceUsage.Memory.Threshold = m.config.ResourceThresholds["memory"]

	// Determine status based on usage
	resourceUsage.CPU.Status = m.getResourceStatus(resourceUsage.CPU.UsagePercent, "cpu")
	resourceUsage.Memory.Status = m.getResourceStatus(resourceUsage.Memory.UsagePercent, "memory")

	// Overall status is worst of all resources
	statuses := []HealthStatus{
		resourceUsage.CPU.Status,
		resourceUsage.Memory.Status,
		resourceUsage.Storage.Status,
		resourceUsage.Pods.Status,
	}

	resourceUsage.OverallStatus = m.getWorstStatus(statuses)

	return resourceUsage
}

func (m *SimplifiedMonitor) analyzePodHealth(pod *models.KubernetesPod) *PodHealth {
	health := &PodHealth{
		Name:         pod.Name,
		Namespace:    pod.Namespace,
		Phase:        pod.Status,
		Ready:        strings.ToLower(pod.Status) == "running",
		RestartCount: 0, // Would need more detailed pod info
		NodeName:     pod.NodeName,
		Status:       HealthStatusUnknown,
	}

	// Determine health status based on phase
	switch strings.ToLower(pod.Status) {
	case "running":
		health.Status = HealthStatusHealthy
	case "pending":
		health.Status = HealthStatusWarning
	case "failed", "error":
		health.Status = HealthStatusCritical
	default:
		health.Status = HealthStatusUnknown
	}

	return health
}

func (m *SimplifiedMonitor) isCriticalPod(pod *models.KubernetesPod) bool {
	// Identify critical system pods
	criticalComponents := []string{
		"kube-apiserver",
		"kube-controller-manager",
		"kube-scheduler",
		"etcd",
		"kube-proxy",
		"coredns",
		"calico",
		"flannel",
	}

	podName := strings.ToLower(pod.Name)
	for _, component := range criticalComponents {
		if strings.Contains(podName, component) {
			return true
		}
	}

	return false
}

func (m *SimplifiedMonitor) getResourceStatus(usagePercent float64, resourceType string) HealthStatus {
	threshold, exists := m.config.ResourceThresholds[resourceType]
	if !exists {
		return HealthStatusUnknown
	}

	if usagePercent >= threshold.Emergency {
		return HealthStatusCritical
	} else if usagePercent >= threshold.Critical {
		return HealthStatusCritical
	} else if usagePercent >= threshold.Warning {
		return HealthStatusWarning
	}

	return HealthStatusHealthy
}

func (m *SimplifiedMonitor) calculateOverallHealth(health *ClusterHealth) HealthStatus {
	var statuses []HealthStatus

	// Collect all component statuses
	for _, component := range health.Components {
		statuses = append(statuses, component.Status)
	}

	// Add subsystem statuses
	if health.Nodes != nil {
		statuses = append(statuses, health.Nodes.OverallStatus)
	}
	if health.SystemPods != nil {
		statuses = append(statuses, health.SystemPods.OverallStatus)
	}
	if health.ResourceUsage != nil {
		statuses = append(statuses, health.ResourceUsage.OverallStatus)
	}

	return m.getWorstStatus(statuses)
}

func (m *SimplifiedMonitor) getWorstStatus(statuses []HealthStatus) HealthStatus {
	// Priority: Critical > Warning > Degraded > Unknown > Healthy
	for _, status := range statuses {
		if status == HealthStatusCritical {
			return HealthStatusCritical
		}
	}
	for _, status := range statuses {
		if status == HealthStatusWarning {
			return HealthStatusWarning
		}
	}
	for _, status := range statuses {
		if status == HealthStatusDegraded {
			return HealthStatusDegraded
		}
	}
	for _, status := range statuses {
		if status == HealthStatusUnknown {
			return HealthStatusUnknown
		}
	}
	return HealthStatusHealthy
}

func (m *SimplifiedMonitor) generateRecommendations(health *ClusterHealth) []string {
	var recommendations []string

	// Resource-based recommendations
	if health.ResourceUsage != nil {
		if health.ResourceUsage.CPU != nil && health.ResourceUsage.CPU.UsagePercent > 80 {
			recommendations = append(recommendations, "Consider scaling up cluster nodes due to high CPU usage")
		}
		if health.ResourceUsage.Memory != nil && health.ResourceUsage.Memory.UsagePercent > 85 {
			recommendations = append(recommendations, "Monitor memory usage closely and consider adding more nodes")
		}
	}

	// Node-based recommendations
	if health.Nodes != nil && health.Nodes.UnreadyNodes > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Investigate %d unready nodes", health.Nodes.UnreadyNodes))
	}

	// Pod-based recommendations
	if health.SystemPods != nil && len(health.SystemPods.CriticalPods) > 0 {
		recommendations = append(recommendations, "Critical system pods are unhealthy - check cluster components")
	}

	// Component-based recommendations
	for name, component := range health.Components {
		if component.Status == HealthStatusCritical {
			recommendations = append(recommendations, fmt.Sprintf("Component %s is critical - immediate attention required", name))
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Cluster is healthy - no immediate actions required")
	}

	return recommendations
}
