package results

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// Processor defines the result processing interface
type Processor interface {
	// Process and format command results
	ProcessCommandResult(result *models.KubernetesOperationResult, execution *models.KubernetesCommandExecution) (*ProcessedResult, error)

	// Format results for different output types
	FormatAsJSON(result *ProcessedResult) (string, error)
	FormatAsTable(result *ProcessedResult) (string, error)
	FormatAsYAML(result *ProcessedResult) (string, error)

	// Extract structured data from raw output
	ExtractResourceMetadata(rawOutput string, resourceType string) (map[string]interface{}, error)

	// Performance analysis
	AnalyzeCommandPerformance(executions []*models.KubernetesCommandExecution) (*PerformanceAnalysis, error)
}

// ProcessedResult represents a structured, processed command result
type ProcessedResult struct {
	// Basic execution info
	ExecutionID   string    `json:"execution_id"`
	Command       string    `json:"command"`
	Status        string    `json:"status"`
	ExecutionTime float64   `json:"execution_time_seconds"`
	Timestamp     time.Time `json:"timestamp"`

	// Structured output data
	ResourceType string                 `json:"resource_type,omitempty"`
	Resources    []ResourceInfo         `json:"resources,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`

	// Raw output for fallback
	RawOutput   string `json:"raw_output,omitempty"`
	ErrorOutput string `json:"error_output,omitempty"`

	// Processing metadata
	ProcessedAt time.Time `json:"processed_at"`
	Format      string    `json:"format"` // "structured", "raw", "error"
}

// ResourceInfo represents structured information about a Kubernetes resource
type ResourceInfo struct {
	Name        string                 `json:"name"`
	Namespace   string                 `json:"namespace,omitempty"`
	Kind        string                 `json:"kind"`
	Status      string                 `json:"status,omitempty"`
	Age         string                 `json:"age,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Annotations map[string]string      `json:"annotations,omitempty"`
	Spec        map[string]interface{} `json:"spec,omitempty"`
	Conditions  []ConditionInfo        `json:"conditions,omitempty"`

	// Resource-specific fields
	Ready        string `json:"ready,omitempty"`         // For pods, deployments
	Restarts     int    `json:"restarts,omitempty"`      // For pods
	NodeName     string `json:"node_name,omitempty"`     // For pods
	ClusterIP    string `json:"cluster_ip,omitempty"`    // For services
	ExternalIP   string `json:"external_ip,omitempty"`   // For services
	Replicas     string `json:"replicas,omitempty"`      // For deployments, replicasets
	Capacity     string `json:"capacity,omitempty"`      // For nodes, PVs
	StorageClass string `json:"storage_class,omitempty"` // For PVCs
}

// ConditionInfo represents a Kubernetes resource condition
type ConditionInfo struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastTransitionTime time.Time `json:"last_transition_time,omitempty"`
	Reason             string    `json:"reason,omitempty"`
	Message            string    `json:"message,omitempty"`
}

// PerformanceAnalysis represents analysis of command execution performance
type PerformanceAnalysis struct {
	TotalExecutions      int     `json:"total_executions"`
	AverageExecutionTime float64 `json:"average_execution_time_seconds"`
	SuccessRate          float64 `json:"success_rate"`
	ErrorRate            float64 `json:"error_rate"`

	// Performance by command type
	CommandStats map[string]*CommandStats `json:"command_stats"`

	// Performance by resource type
	ResourceTypeStats map[string]*ResourceTypeStats `json:"resource_type_stats"`

	// Time-based analysis
	HourlyDistribution map[int]int `json:"hourly_distribution"`

	// Error analysis
	CommonErrors []ErrorPattern `json:"common_errors"`

	// Recommendations
	Recommendations []string `json:"recommendations"`
}

// CommandStats represents statistics for a specific command type
type CommandStats struct {
	Count            int      `json:"count"`
	SuccessRate      float64  `json:"success_rate"`
	AverageTime      float64  `json:"average_time_seconds"`
	MinTime          float64  `json:"min_time_seconds"`
	MaxTime          float64  `json:"max_time_seconds"`
	MostCommonErrors []string `json:"most_common_errors"`
}

// ResourceTypeStats represents statistics for a specific resource type
type ResourceTypeStats struct {
	Count               int     `json:"count"`
	AverageResponseTime float64 `json:"average_response_time_seconds"`
	CacheHitRate        float64 `json:"cache_hit_rate,omitempty"`
}

// ErrorPattern represents a common error pattern
type ErrorPattern struct {
	Pattern     string    `json:"pattern"`
	Count       int       `json:"count"`
	LastSeen    time.Time `json:"last_seen"`
	Description string    `json:"description"`
}

// processor implements the Processor interface
type processor struct{}

// NewProcessor creates a new result processor
func NewProcessor() Processor {
	return &processor{}
}

// ProcessCommandResult processes and structures a command result
func (p *processor) ProcessCommandResult(result *models.KubernetesOperationResult, execution *models.KubernetesCommandExecution) (*ProcessedResult, error) {
	processed := &ProcessedResult{
		ExecutionID:   result.OperationID.String(),
		Command:       execution.Command.Operation,
		Status:        "success",
		ExecutionTime: time.Since(result.ExecutedAt).Seconds(), // Calculate duration from executed time
		Timestamp:     result.ExecutedAt,
		ProcessedAt:   time.Now(),
		Format:        "structured",
	}

	// Handle error cases
	if result.Error != "" {
		processed.Status = "error"
		processed.ErrorOutput = result.Error
		processed.Format = "error"
		return processed, nil
	}

	// Handle success case with result data
	if result.Success && result.Result != nil {
		// Try to convert result to string for parsing
		var outputStr string
		if resultBytes, err := json.Marshal(result.Result); err == nil {
			outputStr = string(resultBytes)
		} else if str, ok := result.Result.(string); ok {
			outputStr = str
		}

		processed.RawOutput = outputStr

		// Try to extract resource information
		resourceType := execution.Command.Resource
		if resourceInfo, err := p.parseKubernetesOutput(outputStr, resourceType); err == nil {
			processed.ResourceType = resourceType
			processed.Resources = resourceInfo
			processed.Format = "structured"
		} else {
			// Fallback to raw format
			processed.Format = "raw"
		}

		// Extract metadata
		if metadata, err := p.ExtractResourceMetadata(outputStr, resourceType); err == nil {
			processed.Metadata = metadata
		}
	}

	return processed, nil
}

// FormatAsJSON formats the result as JSON
func (p *processor) FormatAsJSON(result *ProcessedResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result as JSON: %w", err)
	}
	return string(data), nil
}

// FormatAsTable formats the result as a table
func (p *processor) FormatAsTable(result *ProcessedResult) (string, error) {
	if result.Format == "error" {
		return fmt.Sprintf("ERROR: %s\n", result.ErrorOutput), nil
	}

	if len(result.Resources) == 0 {
		return result.RawOutput, nil
	}

	// Build table based on resource type
	var table strings.Builder

	switch result.ResourceType {
	case "pods":
		table.WriteString("NAME\tREADY\tSTATUS\tRESTARTS\tAGE\tNODE\n")
		for _, res := range result.Resources {
			table.WriteString(fmt.Sprintf("%s\t%s\t%s\t%d\t%s\t%s\n",
				res.Name, res.Ready, res.Status, res.Restarts, res.Age, res.NodeName))
		}
	case "deployments":
		table.WriteString("NAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE\n")
		for _, res := range result.Resources {
			table.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n",
				res.Name, res.Ready, res.Replicas, res.Status, res.Age))
		}
	case "services":
		table.WriteString("NAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tAGE\n")
		for _, res := range result.Resources {
			table.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n",
				res.Name, res.Kind, res.ClusterIP, res.ExternalIP, res.Age))
		}
	default:
		// Generic table format
		table.WriteString("NAME\tKIND\tSTATUS\tAGE\n")
		for _, res := range result.Resources {
			table.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\n",
				res.Name, res.Kind, res.Status, res.Age))
		}
	}

	return table.String(), nil
}

// FormatAsYAML formats the result as YAML (simplified)
func (p *processor) FormatAsYAML(result *ProcessedResult) (string, error) {
	// For now, return JSON format as YAML is more complex
	// In production, you'd use a YAML library like gopkg.in/yaml.v3
	return p.FormatAsJSON(result)
}

// ExtractResourceMetadata extracts metadata from raw output
func (p *processor) ExtractResourceMetadata(rawOutput string, resourceType string) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})

	// Try to parse as JSON first (for kubectl -o json)
	var jsonData interface{}
	if err := json.Unmarshal([]byte(rawOutput), &jsonData); err == nil {
		if jsonMap, ok := jsonData.(map[string]interface{}); ok {
			// Extract common metadata fields
			if items, exists := jsonMap["items"]; exists {
				metadata["item_count"] = len(items.([]interface{}))
			}
			if kind, exists := jsonMap["kind"]; exists {
				metadata["kind"] = kind
			}
			if apiVersion, exists := jsonMap["apiVersion"]; exists {
				metadata["api_version"] = apiVersion
			}
		}
		return metadata, nil
	}

	// Extract from table output
	lines := strings.Split(rawOutput, "\n")
	if len(lines) > 1 {
		// Count data rows (excluding header)
		dataRows := 0
		for i, line := range lines {
			if i == 0 {
				continue // Skip header
			}
			if strings.TrimSpace(line) != "" {
				dataRows++
			}
		}
		metadata["resource_count"] = dataRows
	}

	metadata["resource_type"] = resourceType
	metadata["output_format"] = "table"

	return metadata, nil
}

// AnalyzeCommandPerformance analyzes performance across multiple executions
func (p *processor) AnalyzeCommandPerformance(executions []*models.KubernetesCommandExecution) (*PerformanceAnalysis, error) {
	if len(executions) == 0 {
		return &PerformanceAnalysis{}, nil
	}

	analysis := &PerformanceAnalysis{
		TotalExecutions:    len(executions),
		CommandStats:       make(map[string]*CommandStats),
		ResourceTypeStats:  make(map[string]*ResourceTypeStats),
		HourlyDistribution: make(map[int]int),
		CommonErrors:       []ErrorPattern{},
		Recommendations:    []string{},
	}

	var totalTime float64
	var successCount int
	errorPatterns := make(map[string]*ErrorPattern)

	for _, exec := range executions {
		// Calculate execution time (approximate from created/completed time difference)
		var execTime float64
		if exec.CompletedAt != nil {
			execTime = exec.CompletedAt.Sub(exec.CreatedAt).Seconds()
			totalTime += execTime
		}

		// Track success/failure
		if exec.Status == "completed" && (exec.Result == nil || exec.Result.Error == "") {
			successCount++
		}

		// Track by command type
		cmdType := exec.Command.Operation
		if _, exists := analysis.CommandStats[cmdType]; !exists {
			analysis.CommandStats[cmdType] = &CommandStats{
				Count:            0,
				AverageTime:      0,
				MinTime:          execTime,
				MaxTime:          execTime,
				MostCommonErrors: []string{},
			}
		}

		stats := analysis.CommandStats[cmdType]
		stats.Count++
		if execTime > 0 {
			if stats.MinTime == 0 || execTime < stats.MinTime {
				stats.MinTime = execTime
			}
			if execTime > stats.MaxTime {
				stats.MaxTime = execTime
			}
		}

		// Track by resource type
		resType := exec.Command.Resource
		if _, exists := analysis.ResourceTypeStats[resType]; !exists {
			analysis.ResourceTypeStats[resType] = &ResourceTypeStats{
				Count: 0,
			}
		}
		analysis.ResourceTypeStats[resType].Count++

		// Track hourly distribution
		hour := exec.CreatedAt.Hour()
		analysis.HourlyDistribution[hour]++

		// Track error patterns
		if exec.Result != nil && exec.Result.Error != "" {
			pattern := p.categorizeError(exec.Result.Error)
			if _, exists := errorPatterns[pattern]; !exists {
				errorPatterns[pattern] = &ErrorPattern{
					Pattern:     pattern,
					Count:       0,
					Description: p.describeErrorPattern(pattern),
				}
			}
			errorPatterns[pattern].Count++
			errorPatterns[pattern].LastSeen = exec.CreatedAt
		}
	}

	// Calculate averages
	analysis.AverageExecutionTime = totalTime / float64(len(executions))
	analysis.SuccessRate = float64(successCount) / float64(len(executions))
	analysis.ErrorRate = 1.0 - analysis.SuccessRate

	// Calculate command stats averages
	for _, stats := range analysis.CommandStats {
		if stats.Count > 0 {
			stats.SuccessRate = float64(successCount) / float64(stats.Count)
			stats.AverageTime = totalTime / float64(stats.Count)
		}
	}

	// Convert error patterns to slice
	for _, pattern := range errorPatterns {
		analysis.CommonErrors = append(analysis.CommonErrors, *pattern)
	}

	// Generate recommendations
	analysis.Recommendations = p.generateRecommendations(analysis)

	return analysis, nil
}

// Helper methods

func (p *processor) parseKubernetesOutput(output, resourceType string) ([]ResourceInfo, error) {
	// Try JSON format first
	var jsonData interface{}
	if err := json.Unmarshal([]byte(output), &jsonData); err == nil {
		return p.parseJSONOutput(jsonData, resourceType)
	}

	// Parse table format
	return p.parseTableOutput(output, resourceType)
}

func (p *processor) parseJSONOutput(jsonData interface{}, resourceType string) ([]ResourceInfo, error) {
	var resources []ResourceInfo

	jsonMap, ok := jsonData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid JSON structure")
	}

	// Handle single resource
	if kind, exists := jsonMap["kind"].(string); exists && kind != "List" {
		resource, err := p.parseJSONResource(jsonMap, resourceType)
		if err != nil {
			return nil, err
		}
		return []ResourceInfo{*resource}, nil
	}

	// Handle list of resources
	if items, exists := jsonMap["items"].([]interface{}); exists {
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				resource, err := p.parseJSONResource(itemMap, resourceType)
				if err != nil {
					continue // Skip invalid resources
				}
				resources = append(resources, *resource)
			}
		}
	}

	return resources, nil
}

func (p *processor) parseJSONResource(data map[string]interface{}, resourceType string) (*ResourceInfo, error) {
	resource := &ResourceInfo{
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
		Spec:        make(map[string]interface{}),
	}

	// Extract metadata
	if metadata, exists := data["metadata"].(map[string]interface{}); exists {
		if name, ok := metadata["name"].(string); ok {
			resource.Name = name
		}
		if namespace, ok := metadata["namespace"].(string); ok {
			resource.Namespace = namespace
		}
		if labels, ok := metadata["labels"].(map[string]interface{}); ok {
			for k, v := range labels {
				if str, ok := v.(string); ok {
					resource.Labels[k] = str
				}
			}
		}
		if annotations, ok := metadata["annotations"].(map[string]interface{}); ok {
			for k, v := range annotations {
				if str, ok := v.(string); ok {
					resource.Annotations[k] = str
				}
			}
		}
	}

	// Extract kind
	if kind, ok := data["kind"].(string); ok {
		resource.Kind = kind
	}

	// Extract spec
	if spec, exists := data["spec"].(map[string]interface{}); exists {
		resource.Spec = spec
	}

	// Extract status information
	if status, exists := data["status"].(map[string]interface{}); exists {
		p.extractStatusInfo(resource, status, resourceType)
	}

	return resource, nil
}

func (p *processor) parseTableOutput(output, resourceType string) ([]ResourceInfo, error) {
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("insufficient table data")
	}

	var resources []ResourceInfo
	header := strings.Fields(lines[0])

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < len(header) {
			continue // Skip incomplete lines
		}

		resource := p.parseTableRow(fields, header, resourceType)
		if resource != nil {
			resources = append(resources, *resource)
		}
	}

	return resources, nil
}

func (p *processor) parseTableRow(fields, header []string, resourceType string) *ResourceInfo {
	resource := &ResourceInfo{
		Labels: make(map[string]string),
		Spec:   make(map[string]interface{}),
	}

	// Map fields based on header
	for i, headerField := range header {
		if i >= len(fields) {
			break
		}

		value := fields[i]
		switch strings.ToLower(headerField) {
		case "name":
			resource.Name = value
		case "namespace":
			resource.Namespace = value
		case "ready":
			resource.Ready = value
		case "status":
			resource.Status = value
		case "age":
			resource.Age = value
		case "node":
			resource.NodeName = value
		case "restarts":
			fmt.Sscanf(value, "%d", &resource.Restarts)
		case "cluster-ip":
			resource.ClusterIP = value
		case "external-ip":
			resource.ExternalIP = value
		}
	}

	// Set default kind based on resource type
	if resource.Kind == "" {
		resource.Kind = resourceType
	}

	return resource
}

func (p *processor) extractStatusInfo(resource *ResourceInfo, status map[string]interface{}, resourceType string) {
	switch resourceType {
	case "pods":
		if phase, ok := status["phase"].(string); ok {
			resource.Status = phase
		}
		// Extract conditions
		if conditions, ok := status["conditions"].([]interface{}); ok {
			for _, cond := range conditions {
				if condMap, ok := cond.(map[string]interface{}); ok {
					condition := ConditionInfo{}
					if t, ok := condMap["type"].(string); ok {
						condition.Type = t
					}
					if s, ok := condMap["status"].(string); ok {
						condition.Status = s
					}
					if r, ok := condMap["reason"].(string); ok {
						condition.Reason = r
					}
					if m, ok := condMap["message"].(string); ok {
						condition.Message = m
					}
					resource.Conditions = append(resource.Conditions, condition)
				}
			}
		}
	case "deployments":
		if readyReplicas, ok := status["readyReplicas"].(float64); ok {
			if replicas, ok := status["replicas"].(float64); ok {
				resource.Ready = fmt.Sprintf("%.0f/%.0f", readyReplicas, replicas)
			}
		}
	case "services":
		if loadBalancer, ok := status["loadBalancer"].(map[string]interface{}); ok {
			if ingress, ok := loadBalancer["ingress"].([]interface{}); ok && len(ingress) > 0 {
				if ingressMap, ok := ingress[0].(map[string]interface{}); ok {
					if ip, ok := ingressMap["ip"].(string); ok {
						resource.ExternalIP = ip
					}
				}
			}
		}
	}
}

func (p *processor) categorizeError(errorMsg string) string {
	errorLower := strings.ToLower(errorMsg)

	if strings.Contains(errorLower, "not found") {
		return "resource_not_found"
	}
	if strings.Contains(errorLower, "forbidden") || strings.Contains(errorLower, "permission") {
		return "permission_denied"
	}
	if strings.Contains(errorLower, "timeout") {
		return "timeout"
	}
	if strings.Contains(errorLower, "connection") {
		return "connection_error"
	}
	if strings.Contains(errorLower, "invalid") {
		return "invalid_request"
	}
	return "unknown_error"
}

func (p *processor) describeErrorPattern(pattern string) string {
	descriptions := map[string]string{
		"resource_not_found": "Resource does not exist in the specified namespace or cluster",
		"permission_denied":  "Insufficient permissions to perform the requested operation",
		"timeout":            "Operation timed out, possibly due to cluster or network issues",
		"connection_error":   "Unable to connect to the Kubernetes cluster",
		"invalid_request":    "Request contains invalid parameters or syntax",
		"unknown_error":      "Uncategorized error that needs further investigation",
	}

	if desc, exists := descriptions[pattern]; exists {
		return desc
	}
	return "No description available for this error pattern"
}

func (p *processor) generateRecommendations(analysis *PerformanceAnalysis) []string {
	var recommendations []string

	// Performance recommendations
	if analysis.AverageExecutionTime > 5.0 {
		recommendations = append(recommendations, "Consider optimizing slow-running commands or increasing cluster resources")
	}

	// Error rate recommendations
	if analysis.ErrorRate > 0.1 {
		recommendations = append(recommendations, "High error rate detected - review command parameters and cluster connectivity")
	}

	// Cache recommendations
	for resType, stats := range analysis.ResourceTypeStats {
		if stats.Count > 10 && stats.CacheHitRate < 0.5 {
			recommendations = append(recommendations,
				fmt.Sprintf("Low cache hit rate for %s resources - consider adjusting cache TTL", resType))
		}
	}

	return recommendations
}
