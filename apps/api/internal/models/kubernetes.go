package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrInvalidInput represents a validation error for invalid input
type ErrInvalidInput struct {
	Field   string
	Message string
}

func (e ErrInvalidInput) Error() string {
	return fmt.Sprintf("invalid input for field '%s': %s", e.Field, e.Message)
}

// ClusterInfo represents basic cluster information
type ClusterInfo struct {
	Version   string    `json:"version" db:"version"`
	NodeCount int32     `json:"node_count" db:"node_count"`
	APIServer string    `json:"api_server" db:"api_server"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// KubernetesNamespace represents a Kubernetes namespace
type KubernetesNamespace struct {
	Name      string            `json:"name" db:"name"`
	Status    string            `json:"status" db:"status"`
	CreatedAt time.Time         `json:"created_at" db:"created_at"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// KubernetesPod represents a Kubernetes pod
type KubernetesPod struct {
	Name        string            `json:"name" db:"name"`
	Namespace   string            `json:"namespace" db:"namespace"`
	Status      string            `json:"status" db:"status"`
	Ready       bool              `json:"ready" db:"ready"`
	Restarts    int32             `json:"restarts" db:"restarts"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	NodeName    string            `json:"node_name,omitempty" db:"node_name"`
	PodIP       string            `json:"pod_ip,omitempty" db:"pod_ip"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// KubernetesDeployment represents a Kubernetes deployment
type KubernetesDeployment struct {
	Name              string            `json:"name" db:"name"`
	Namespace         string            `json:"namespace" db:"namespace"`
	Replicas          int32             `json:"replicas" db:"replicas"`
	ReadyReplicas     int32             `json:"ready_replicas" db:"ready_replicas"`
	AvailableReplicas int32             `json:"available_replicas" db:"available_replicas"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
}

// KubernetesService represents a Kubernetes service
type KubernetesService struct {
	Name        string            `json:"name" db:"name"`
	Namespace   string            `json:"namespace" db:"namespace"`
	Type        string            `json:"type" db:"type"`
	ClusterIP   string            `json:"cluster_ip" db:"cluster_ip"`
	ExternalIPs []string          `json:"external_ips,omitempty"`
	Ports       []ServicePort     `json:"ports,omitempty"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ServicePort represents a service port
type ServicePort struct {
	Name       string `json:"name,omitempty"`
	Port       int32  `json:"port"`
	TargetPort string `json:"target_port"`
	Protocol   string `json:"protocol"`
}

// KubernetesConfigMap represents a Kubernetes configmap
type KubernetesConfigMap struct {
	Name      string            `json:"name" db:"name"`
	Namespace string            `json:"namespace" db:"namespace"`
	Data      map[string]string `json:"data,omitempty"`
	CreatedAt time.Time         `json:"created_at" db:"created_at"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// KubernetesSecret represents a Kubernetes secret (metadata only for security)
type KubernetesSecret struct {
	Name      string    `json:"name" db:"name"`
	Namespace string    `json:"namespace" db:"namespace"`
	Type      string    `json:"type" db:"type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	KeyCount  int       `json:"key_count" db:"key_count"`
}

// LogOptions represents pod log retrieval options
type LogOptions struct {
	TailLines  int64  `json:"tail_lines"`
	Timestamps bool   `json:"timestamps"`
	Container  string `json:"container,omitempty"`
}

// KubernetesOperation represents a Kubernetes operation for validation
type KubernetesOperation struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	SessionID uuid.UUID `json:"session_id" db:"session_id"`
	Operation string    `json:"operation" db:"operation"`
	Resource  string    `json:"resource" db:"resource"`
	Namespace string    `json:"namespace" db:"namespace"`
	Name      string    `json:"name,omitempty" db:"name"`
	Context   string    `json:"context,omitempty" db:"context"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// KubernetesOperationResult represents the result of a Kubernetes operation
type KubernetesOperationResult struct {
	OperationID   uuid.UUID              `json:"operation_id" db:"operation_id"`
	Success       bool                   `json:"success" db:"success"`
	Result        interface{}            `json:"result,omitempty"`
	Error         string                 `json:"error,omitempty" db:"error"`
	ExecutedAt    time.Time              `json:"executed_at" db:"executed_at"`
	BackupData    map[string]interface{} `json:"backup_data,omitempty" db:"backup_data"`       // For rollback support
	PreviousState map[string]interface{} `json:"previous_state,omitempty" db:"previous_state"` // For rollback support
}

// ClusterContext represents a Kubernetes cluster context
type ClusterContext struct {
	ID                uuid.UUID `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	Cluster           string    `json:"cluster" db:"cluster"`
	User              string    `json:"user" db:"user"`
	Namespace         string    `json:"namespace" db:"namespace"`
	APIServer         string    `json:"api_server" db:"api_server"`
	IsActive          bool      `json:"is_active" db:"is_active"`
	AllowedNamespaces []string  `json:"allowed_namespaces,omitempty"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// ValidateKubernetesOperation validates a Kubernetes operation request
func (op *KubernetesOperation) Validate() error {
	if op.Operation == "" {
		return ErrInvalidInput{Field: "operation", Message: "operation is required"}
	}

	if op.Resource == "" {
		return ErrInvalidInput{Field: "resource", Message: "resource is required"}
	}

	if op.Namespace == "" {
		return ErrInvalidInput{Field: "namespace", Message: "namespace is required"}
	}

	// Validate operation types
	validOperations := map[string]bool{
		"get":     true,
		"list":    true,
		"delete":  true,
		"scale":   true,
		"restart": true,
		"logs":    true,
	}

	if !validOperations[op.Operation] {
		return ErrInvalidInput{
			Field:   "operation",
			Message: "invalid operation type",
		}
	}

	// Validate resource types
	validResources := map[string]bool{
		"pods":        true,
		"deployments": true,
		"services":    true,
		"configmaps":  true,
		"secrets":     true,
	}

	if !validResources[op.Resource] {
		return ErrInvalidInput{
			Field:   "resource",
			Message: "invalid resource type",
		}
	}

	return nil
}

// GetSafetyLevel returns the safety level of a Kubernetes operation
func (op *KubernetesOperation) GetSafetyLevel() string {
	// Read operations are safe
	if op.Operation == "get" || op.Operation == "list" || op.Operation == "logs" {
		return "safe"
	}

	// Pod deletion and scaling are warning level
	if (op.Operation == "delete" && op.Resource == "pods") || op.Operation == "scale" || op.Operation == "restart" {
		return "warning"
	}

	// Any other operations are dangerous
	return "dangerous"
}

// GetDescription returns a human-readable description of the operation
func (op *KubernetesOperation) GetDescription() string {
	switch op.Operation {
	case "get":
		if op.Name != "" {
			return fmt.Sprintf("Get %s '%s' in namespace '%s'", op.Resource, op.Name, op.Namespace)
		}
		return fmt.Sprintf("Get %s in namespace '%s'", op.Resource, op.Namespace)
	case "list":
		return fmt.Sprintf("List %s in namespace '%s'", op.Resource, op.Namespace)
	case "delete":
		return fmt.Sprintf("Delete %s '%s' in namespace '%s'", op.Resource, op.Name, op.Namespace)
	case "scale":
		return fmt.Sprintf("Scale %s '%s' in namespace '%s'", op.Resource, op.Name, op.Namespace)
	case "restart":
		return fmt.Sprintf("Restart %s '%s' in namespace '%s'", op.Resource, op.Name, op.Namespace)
	case "logs":
		return fmt.Sprintf("Get logs for %s '%s' in namespace '%s'", op.Resource, op.Name, op.Namespace)
	default:
		return fmt.Sprintf("Unknown operation '%s' on %s in namespace '%s'", op.Operation, op.Resource, op.Namespace)
	}
}

// ExecutionStats represents execution statistics for analytics
type ExecutionStats struct {
	UserID           uuid.UUID `json:"user_id" db:"user_id"`
	TotalExecutions  int       `json:"total_executions" db:"total_executions"`
	SuccessfulOnes   int       `json:"successful_ones" db:"successful_ones"`
	FailedOnes       int       `json:"failed_ones" db:"failed_ones"`
	AverageTime      float64   `json:"average_time_ms" db:"average_time_ms"`
	MostUsedResource string    `json:"most_used_resource" db:"most_used_resource"`
	From             time.Time `json:"from" db:"from"`
	To               time.Time `json:"to" db:"to"`
}
