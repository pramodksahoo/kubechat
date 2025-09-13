package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// Service defines the Kubernetes service interface
type Service interface {
	// Cluster management
	GetClusterInfo(ctx context.Context) (*models.ClusterInfo, error)
	ListNamespaces(ctx context.Context) ([]*models.KubernetesNamespace, error)

	// Pod operations
	ListPods(ctx context.Context, namespace string) ([]*models.KubernetesPod, error)
	GetPod(ctx context.Context, namespace, name string) (*models.KubernetesPod, error)
	DeletePod(ctx context.Context, namespace, name string) error

	// Deployment operations
	ListDeployments(ctx context.Context, namespace string) ([]*models.KubernetesDeployment, error)
	GetDeployment(ctx context.Context, namespace, name string) (*models.KubernetesDeployment, error)
	ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error
	RestartDeployment(ctx context.Context, namespace, name string) error

	// Service operations
	ListServices(ctx context.Context, namespace string) ([]*models.KubernetesService, error)
	GetService(ctx context.Context, namespace, name string) (*models.KubernetesService, error)

	// ConfigMap operations
	ListConfigMaps(ctx context.Context, namespace string) ([]*models.KubernetesConfigMap, error)
	GetConfigMap(ctx context.Context, namespace, name string) (*models.KubernetesConfigMap, error)

	// Secret operations (safe exposure only)
	ListSecrets(ctx context.Context, namespace string) ([]*models.KubernetesSecret, error)

	// Log operations
	GetPodLogs(ctx context.Context, namespace, podName string, options *models.LogOptions) (string, error)

	// Resource validation
	ValidateOperation(ctx context.Context, operation *models.KubernetesOperation) error

	// Health check
	HealthCheck(ctx context.Context) error

	// Command execution integration (Story 1.6)
	ExecuteOperation(ctx context.Context, operation *models.KubernetesOperation) (*models.KubernetesOperationResult, error)

	// Rollback support operations (Story 1.6)
	ResourceExists(ctx context.Context, resource, name, namespace string) (bool, error)
	CreateResource(ctx context.Context, resource string, data map[string]interface{}) error
	DeleteResource(ctx context.Context, resource, name, namespace string) error
	PatchResource(ctx context.Context, resource, name, namespace string, data map[string]interface{}) error
	ScaleResource(ctx context.Context, resource, name, namespace string, replicas int) error
}

// service implements the Kubernetes service
type service struct {
	client            kubernetes.Interface
	config            *rest.Config
	allowedNamespaces []string
	maxLogLines       int64
	timeout           time.Duration
}

// NewService creates a new Kubernetes service
func NewService(kubeconfig string, allowedNamespaces []string) (Service, error) {
	config, err := buildConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build Kubernetes config: %w", err)
	}

	// Configure timeouts and limits for security
	config.Timeout = 30 * time.Second
	config.QPS = 20   // Limit queries per second
	config.Burst = 30 // Allow burst requests

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &service{
		client:            client,
		config:            config,
		allowedNamespaces: allowedNamespaces,
		maxLogLines:       1000, // Limit log lines for security
		timeout:           30 * time.Second,
	}, nil
}

// buildConfig builds Kubernetes client configuration
func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		// Use provided kubeconfig file
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	// Use in-cluster configuration when running in a pod
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to default kubeconfig
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		).ClientConfig()
	}

	return config, nil
}

// GetClusterInfo returns basic cluster information
func (s *service) GetClusterInfo(ctx context.Context) (*models.ClusterInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Get cluster version
	version, err := s.client.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	// Get node count
	nodes, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	return &models.ClusterInfo{
		Version:   version.String(),
		NodeCount: int32(len(nodes.Items)),
		APIServer: s.config.Host,
		CreatedAt: time.Now(),
	}, nil
}

// ListNamespaces returns allowed namespaces
func (s *service) ListNamespaces(ctx context.Context) ([]*models.KubernetesNamespace, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	namespaces, err := s.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var result []*models.KubernetesNamespace
	for _, ns := range namespaces.Items {
		// Filter to allowed namespaces for security
		if s.isNamespaceAllowed(ns.Name) {
			result = append(result, &models.KubernetesNamespace{
				Name:      ns.Name,
				Status:    string(ns.Status.Phase),
				CreatedAt: ns.CreationTimestamp.Time,
				Labels:    ns.Labels,
			})
		}
	}

	return result, nil
}

// ListPods returns pods in a namespace
func (s *service) ListPods(ctx context.Context, namespace string) ([]*models.KubernetesPod, error) {
	if !s.isNamespaceAllowed(namespace) {
		return nil, fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	pods, err := s.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var result []*models.KubernetesPod
	for _, pod := range pods.Items {
		result = append(result, convertPodToModel(&pod))
	}

	return result, nil
}

// GetPod returns a specific pod
func (s *service) GetPod(ctx context.Context, namespace, name string) (*models.KubernetesPod, error) {
	if !s.isNamespaceAllowed(namespace) {
		return nil, fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	pod, err := s.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	return convertPodToModel(pod), nil
}

// DeletePod safely deletes a pod
func (s *service) DeletePod(ctx context.Context, namespace, name string) error {
	if !s.isNamespaceAllowed(namespace) {
		return fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Use graceful deletion with timeout
	gracePeriod := int64(30)
	err := s.client.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	})

	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}

	return nil
}

// ListDeployments returns deployments in a namespace
func (s *service) ListDeployments(ctx context.Context, namespace string) ([]*models.KubernetesDeployment, error) {
	if !s.isNamespaceAllowed(namespace) {
		return nil, fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	deployments, err := s.client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	var result []*models.KubernetesDeployment
	for _, deploy := range deployments.Items {
		result = append(result, convertDeploymentToModel(&deploy))
	}

	return result, nil
}

// GetDeployment returns a specific deployment
func (s *service) GetDeployment(ctx context.Context, namespace, name string) (*models.KubernetesDeployment, error) {
	if !s.isNamespaceAllowed(namespace) {
		return nil, fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	deployment, err := s.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	return convertDeploymentToModel(deployment), nil
}

// ScaleDeployment scales a deployment
func (s *service) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error {
	if !s.isNamespaceAllowed(namespace) {
		return fmt.Errorf("access denied to namespace: %s", namespace)
	}

	// Safety check: limit max replicas
	if replicas > 10 {
		return fmt.Errorf("replica count %d exceeds maximum allowed (10)", replicas)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, err := s.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		deployment.Spec.Replicas = &replicas
		_, err = s.client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}

	return nil
}

// RestartDeployment restarts a deployment
func (s *service) RestartDeployment(ctx context.Context, namespace, name string) error {
	if !s.isNamespaceAllowed(namespace) {
		return fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, err := s.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		// Trigger restart by updating restart annotation
		if deployment.Spec.Template.Annotations == nil {
			deployment.Spec.Template.Annotations = make(map[string]string)
		}
		deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

		_, err = s.client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to restart deployment: %w", err)
	}

	return nil
}

// ListServices returns services in a namespace
func (s *service) ListServices(ctx context.Context, namespace string) ([]*models.KubernetesService, error) {
	if !s.isNamespaceAllowed(namespace) {
		return nil, fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	services, err := s.client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	var result []*models.KubernetesService
	for _, svc := range services.Items {
		result = append(result, convertServiceToModel(&svc))
	}

	return result, nil
}

// GetService returns a specific service
func (s *service) GetService(ctx context.Context, namespace, name string) (*models.KubernetesService, error) {
	if !s.isNamespaceAllowed(namespace) {
		return nil, fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	service, err := s.client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return convertServiceToModel(service), nil
}

// ListConfigMaps returns configmaps in a namespace
func (s *service) ListConfigMaps(ctx context.Context, namespace string) ([]*models.KubernetesConfigMap, error) {
	if !s.isNamespaceAllowed(namespace) {
		return nil, fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	configMaps, err := s.client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps: %w", err)
	}

	var result []*models.KubernetesConfigMap
	for _, cm := range configMaps.Items {
		result = append(result, convertConfigMapToModel(&cm))
	}

	return result, nil
}

// GetConfigMap returns a specific configmap
func (s *service) GetConfigMap(ctx context.Context, namespace, name string) (*models.KubernetesConfigMap, error) {
	if !s.isNamespaceAllowed(namespace) {
		return nil, fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	configMap, err := s.client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap: %w", err)
	}

	return convertConfigMapToModel(configMap), nil
}

// ListSecrets returns secrets in a namespace (metadata only for security)
func (s *service) ListSecrets(ctx context.Context, namespace string) ([]*models.KubernetesSecret, error) {
	if !s.isNamespaceAllowed(namespace) {
		return nil, fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	secrets, err := s.client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	var result []*models.KubernetesSecret
	for _, secret := range secrets.Items {
		// Only expose metadata for security
		result = append(result, &models.KubernetesSecret{
			Name:      secret.Name,
			Namespace: secret.Namespace,
			Type:      string(secret.Type),
			CreatedAt: secret.CreationTimestamp.Time,
			KeyCount:  len(secret.Data),
		})
	}

	return result, nil
}

// GetPodLogs returns pod logs with safety limits
func (s *service) GetPodLogs(ctx context.Context, namespace, podName string, options *models.LogOptions) (string, error) {
	if !s.isNamespaceAllowed(namespace) {
		return "", fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Apply safety limits
	tailLines := options.TailLines
	if tailLines > s.maxLogLines {
		tailLines = s.maxLogLines
	}

	logOptions := &corev1.PodLogOptions{
		Follow:     false, // Never follow for security
		TailLines:  &tailLines,
		Timestamps: options.Timestamps,
		Container:  options.Container,
	}

	req := s.client.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
	logs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer logs.Close()

	buf := make([]byte, 1024*1024) // 1MB limit
	n, _ := logs.Read(buf)

	return string(buf[:n]), nil
}

// ValidateOperation validates a Kubernetes operation for safety
func (s *service) ValidateOperation(ctx context.Context, operation *models.KubernetesOperation) error {
	// Check namespace access
	if !s.isNamespaceAllowed(operation.Namespace) {
		return fmt.Errorf("access denied to namespace: %s", operation.Namespace)
	}

	// Validate operation type
	allowedOperations := map[string]bool{
		"get":     true,
		"list":    true,
		"delete":  true,
		"scale":   true,
		"restart": true,
		"logs":    true,
	}

	if !allowedOperations[operation.Operation] {
		return fmt.Errorf("operation not allowed: %s", operation.Operation)
	}

	// Validate resource types
	allowedResources := map[string]bool{
		"pods":        true,
		"deployments": true,
		"services":    true,
		"configmaps":  true,
		"secrets":     true,
	}

	if !allowedResources[operation.Resource] {
		return fmt.Errorf("resource type not allowed: %s", operation.Resource)
	}

	// Additional safety checks for destructive operations
	if operation.Operation == "delete" && operation.Resource == "deployment" {
		return fmt.Errorf("deployment deletion is not allowed for safety")
	}

	return nil
}

// HealthCheck verifies the Kubernetes connection
func (s *service) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := s.client.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("kubernetes health check failed: %w", err)
	}

	return nil
}

// isNamespaceAllowed checks if namespace access is allowed
func (s *service) isNamespaceAllowed(namespace string) bool {
	// If no restrictions, allow all
	if len(s.allowedNamespaces) == 0 {
		return true
	}

	// Check against allowed list
	for _, allowed := range s.allowedNamespaces {
		if allowed == namespace || allowed == "*" {
			return true
		}
		// Support wildcard patterns
		if strings.HasSuffix(allowed, "*") {
			prefix := strings.TrimSuffix(allowed, "*")
			if strings.HasPrefix(namespace, prefix) {
				return true
			}
		}
	}

	return false
}

// ExecuteOperation executes a Kubernetes operation and returns structured result
func (s *service) ExecuteOperation(ctx context.Context, operation *models.KubernetesOperation) (*models.KubernetesOperationResult, error) {
	startTime := time.Now()

	// Validate the operation first
	if err := s.ValidateOperation(ctx, operation); err != nil {
		return &models.KubernetesOperationResult{
			OperationID: operation.ID,
			Success:     false,
			Error:       err.Error(),
			ExecutedAt:  startTime,
		}, err
	}

	var result interface{}
	var err error

	// Execute based on operation type
	switch operation.Operation {
	case "get":
		result, err = s.executeGetOperation(ctx, operation)
	case "list":
		result, err = s.executeListOperation(ctx, operation)
	case "delete":
		result, err = s.executeDeleteOperation(ctx, operation)
	case "scale":
		result, err = s.executeScaleOperation(ctx, operation)
	case "restart":
		result, err = s.executeRestartOperation(ctx, operation)
	case "logs":
		result, err = s.executeLogsOperation(ctx, operation)
	default:
		err = fmt.Errorf("unsupported operation: %s", operation.Operation)
	}

	executedAt := time.Now()

	if err != nil {
		return &models.KubernetesOperationResult{
			OperationID: operation.ID,
			Success:     false,
			Error:       err.Error(),
			ExecutedAt:  executedAt,
		}, err
	}

	return &models.KubernetesOperationResult{
		OperationID: operation.ID,
		Success:     true,
		Result:      result,
		ExecutedAt:  executedAt,
	}, nil
}

// executeGetOperation handles get operations
func (s *service) executeGetOperation(ctx context.Context, operation *models.KubernetesOperation) (interface{}, error) {
	switch operation.Resource {
	case "pods":
		if operation.Name != "" {
			return s.GetPod(ctx, operation.Namespace, operation.Name)
		}
		return s.ListPods(ctx, operation.Namespace)
	case "deployments":
		if operation.Name != "" {
			return s.GetDeployment(ctx, operation.Namespace, operation.Name)
		}
		return s.ListDeployments(ctx, operation.Namespace)
	case "services":
		if operation.Name != "" {
			return s.GetService(ctx, operation.Namespace, operation.Name)
		}
		return s.ListServices(ctx, operation.Namespace)
	case "configmaps":
		if operation.Name != "" {
			return s.GetConfigMap(ctx, operation.Namespace, operation.Name)
		}
		return s.ListConfigMaps(ctx, operation.Namespace)
	case "secrets":
		return s.ListSecrets(ctx, operation.Namespace)
	default:
		return nil, fmt.Errorf("unsupported resource for get operation: %s", operation.Resource)
	}
}

// executeListOperation handles list operations
func (s *service) executeListOperation(ctx context.Context, operation *models.KubernetesOperation) (interface{}, error) {
	switch operation.Resource {
	case "pods":
		return s.ListPods(ctx, operation.Namespace)
	case "deployments":
		return s.ListDeployments(ctx, operation.Namespace)
	case "services":
		return s.ListServices(ctx, operation.Namespace)
	case "configmaps":
		return s.ListConfigMaps(ctx, operation.Namespace)
	case "secrets":
		return s.ListSecrets(ctx, operation.Namespace)
	default:
		return nil, fmt.Errorf("unsupported resource for list operation: %s", operation.Resource)
	}
}

// executeDeleteOperation handles delete operations
func (s *service) executeDeleteOperation(ctx context.Context, operation *models.KubernetesOperation) (interface{}, error) {
	if operation.Name == "" {
		return nil, fmt.Errorf("resource name is required for delete operation")
	}

	switch operation.Resource {
	case "pods":
		err := s.DeletePod(ctx, operation.Namespace, operation.Name)
		if err != nil {
			return nil, err
		}
		return map[string]string{"message": fmt.Sprintf("Pod %s deleted successfully", operation.Name)}, nil
	default:
		return nil, fmt.Errorf("delete operation not supported for resource: %s", operation.Resource)
	}
}

// executeScaleOperation handles scale operations
func (s *service) executeScaleOperation(ctx context.Context, operation *models.KubernetesOperation) (interface{}, error) {
	if operation.Name == "" {
		return nil, fmt.Errorf("resource name is required for scale operation")
	}

	// Default to 1 replica if not specified in operation context
	replicas := int32(1)

	switch operation.Resource {
	case "deployments":
		err := s.ScaleDeployment(ctx, operation.Namespace, operation.Name, replicas)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message":  fmt.Sprintf("Deployment %s scaled successfully", operation.Name),
			"replicas": replicas,
		}, nil
	default:
		return nil, fmt.Errorf("scale operation not supported for resource: %s", operation.Resource)
	}
}

// executeRestartOperation handles restart operations
func (s *service) executeRestartOperation(ctx context.Context, operation *models.KubernetesOperation) (interface{}, error) {
	if operation.Name == "" {
		return nil, fmt.Errorf("resource name is required for restart operation")
	}

	switch operation.Resource {
	case "deployments":
		err := s.RestartDeployment(ctx, operation.Namespace, operation.Name)
		if err != nil {
			return nil, err
		}
		return map[string]string{"message": fmt.Sprintf("Deployment %s restarted successfully", operation.Name)}, nil
	default:
		return nil, fmt.Errorf("restart operation not supported for resource: %s", operation.Resource)
	}
}

// executeLogsOperation handles logs operations
func (s *service) executeLogsOperation(ctx context.Context, operation *models.KubernetesOperation) (interface{}, error) {
	if operation.Name == "" {
		return nil, fmt.Errorf("resource name is required for logs operation")
	}

	switch operation.Resource {
	case "pods":
		logOptions := &models.LogOptions{
			TailLines:  100,
			Timestamps: true,
		}

		logs, err := s.GetPodLogs(ctx, operation.Namespace, operation.Name, logOptions)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"logs":  logs,
			"pod":   operation.Name,
			"lines": strings.Count(logs, "\n"),
		}, nil
	default:
		return nil, fmt.Errorf("logs operation not supported for resource: %s", operation.Resource)
	}
}

// Helper functions to convert Kubernetes objects to models
func convertPodToModel(pod *corev1.Pod) *models.KubernetesPod {
	return &models.KubernetesPod{
		Name:        pod.Name,
		Namespace:   pod.Namespace,
		Status:      string(pod.Status.Phase),
		Ready:       isPodReady(pod),
		Restarts:    getPodRestartCount(pod),
		CreatedAt:   pod.CreationTimestamp.Time,
		NodeName:    pod.Spec.NodeName,
		PodIP:       pod.Status.PodIP,
		Labels:      pod.Labels,
		Annotations: pod.Annotations,
	}
}

func convertDeploymentToModel(deployment *appsv1.Deployment) *models.KubernetesDeployment {
	return &models.KubernetesDeployment{
		Name:              deployment.Name,
		Namespace:         deployment.Namespace,
		Replicas:          *deployment.Spec.Replicas,
		ReadyReplicas:     deployment.Status.ReadyReplicas,
		AvailableReplicas: deployment.Status.AvailableReplicas,
		CreatedAt:         deployment.CreationTimestamp.Time,
		Labels:            deployment.Labels,
		Annotations:       deployment.Annotations,
	}
}

func convertServiceToModel(service *corev1.Service) *models.KubernetesService {
	return &models.KubernetesService{
		Name:        service.Name,
		Namespace:   service.Namespace,
		Type:        string(service.Spec.Type),
		ClusterIP:   service.Spec.ClusterIP,
		ExternalIPs: service.Spec.ExternalIPs,
		Ports:       convertServicePorts(service.Spec.Ports),
		CreatedAt:   service.CreationTimestamp.Time,
		Labels:      service.Labels,
		Annotations: service.Annotations,
	}
}

func convertConfigMapToModel(configMap *corev1.ConfigMap) *models.KubernetesConfigMap {
	return &models.KubernetesConfigMap{
		Name:      configMap.Name,
		Namespace: configMap.Namespace,
		Data:      configMap.Data,
		CreatedAt: configMap.CreationTimestamp.Time,
		Labels:    configMap.Labels,
	}
}

func convertServicePorts(ports []corev1.ServicePort) []models.ServicePort {
	var result []models.ServicePort
	for _, port := range ports {
		result = append(result, models.ServicePort{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: port.TargetPort.String(),
			Protocol:   string(port.Protocol),
		})
	}
	return result
}

func isPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func getPodRestartCount(pod *corev1.Pod) int32 {
	var restarts int32
	for _, containerStatus := range pod.Status.ContainerStatuses {
		restarts += containerStatus.RestartCount
	}
	return restarts
}

// ResourceExists checks if a specific resource exists in the cluster
func (s *service) ResourceExists(ctx context.Context, resource, name, namespace string) (bool, error) {
	if !s.isNamespaceAllowed(namespace) {
		return false, fmt.Errorf("access denied to namespace: %s", namespace)
	}

	switch strings.ToLower(resource) {
	case "pod", "pods":
		_, err := s.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, nil // Resource doesn't exist
		}
		return true, nil

	case "service", "services", "svc":
		_, err := s.client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return true, nil

	case "deployment", "deployments", "deploy":
		_, err := s.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return true, nil

	case "configmap", "configmaps", "cm":
		_, err := s.client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return true, nil

	default:
		return false, fmt.Errorf("resource type %s not supported for existence check", resource)
	}
}

// CreateResource creates a new resource from data
func (s *service) CreateResource(ctx context.Context, resource string, data map[string]interface{}) error {
	// For rollback purposes, we primarily support recreating deleted resources
	// This would typically involve parsing the resource data and creating the appropriate Kubernetes object
	// For now, return not supported as resource creation from arbitrary data requires complex parsing
	return fmt.Errorf("create resource not yet supported for rollback operations")
}

// DeleteResource deletes a specific resource
func (s *service) DeleteResource(ctx context.Context, resource, name, namespace string) error {
	if !s.isNamespaceAllowed(namespace) {
		return fmt.Errorf("access denied to namespace: %s", namespace)
	}

	switch strings.ToLower(resource) {
	case "pod", "pods":
		return s.client.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})

	case "service", "services", "svc":
		return s.client.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})

	case "deployment", "deployments", "deploy":
		return s.client.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})

	case "configmap", "configmaps", "cm":
		return s.client.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})

	default:
		return fmt.Errorf("resource type %s not supported for deletion", resource)
	}
}

// PatchResource applies a patch to a specific resource
func (s *service) PatchResource(ctx context.Context, resource, name, namespace string, data map[string]interface{}) error {
	if !s.isNamespaceAllowed(namespace) {
		return fmt.Errorf("access denied to namespace: %s", namespace)
	}

	// For rollback purposes, patch operations are complex and would require
	// proper JSON patch or strategic merge patch implementation
	// For now, return not supported
	return fmt.Errorf("patch resource not yet supported for rollback operations")
}

// ScaleResource scales a resource to the specified replica count
func (s *service) ScaleResource(ctx context.Context, resource, name, namespace string, replicas int) error {
	if !s.isNamespaceAllowed(namespace) {
		return fmt.Errorf("access denied to namespace: %s", namespace)
	}

	switch strings.ToLower(resource) {
	case "deployment", "deployments", "deploy":
		return s.ScaleDeployment(ctx, namespace, name, int32(replicas))

	default:
		return fmt.Errorf("resource type %s does not support scaling", resource)
	}
}
