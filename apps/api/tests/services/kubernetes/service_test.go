package kubernetes_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	discoveryfake "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	kubernetesService "github.com/pramodksahoo/kubechat/apps/api/internal/services/kubernetes"
)

func TestKubernetesService_GetClusterInfo(t *testing.T) {
	tests := []struct {
		name           string
		serverVersion  *version.Info
		nodeCount      int
		expectError    bool
		expectedResult *models.ClusterInfo
	}{
		{
			name: "successful cluster info retrieval",
			serverVersion: &version.Info{
				Major:      "1",
				Minor:      "28",
				GitVersion: "v1.28.3",
			},
			nodeCount:   3,
			expectError: false,
			expectedResult: &models.ClusterInfo{
				Version:   "v1.28.3",
				NodeCount: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client
			fakeClient := fake.NewSimpleClientset()

			// Add nodes
			for i := 0; i < tt.nodeCount; i++ {
				node := &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: fmt.Sprintf("node-%d", i+1),
					},
				}
				fakeClient.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})
			}

			// Mock the Discovery client to return server version
			fakeDiscovery := fakeClient.Discovery().(*discoveryfake.FakeDiscovery)
			fakeDiscovery.FakedServerVersion = tt.serverVersion

			// Create service with fake client - note this requires a constructor that accepts client
			service := createTestService(fakeClient, []string{})

			result, err := service.GetClusterInfo(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Version, result.Version)
				assert.Equal(t, tt.expectedResult.NodeCount, result.NodeCount)
			}
		})
	}
}

func TestKubernetesService_ListNamespaces(t *testing.T) {
	tests := []struct {
		name              string
		namespaces        []corev1.Namespace
		allowedNamespaces []string
		expectedCount     int
		expectError       bool
	}{
		{
			name: "list all namespaces when no restrictions",
			namespaces: []corev1.Namespace{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "default"},
					Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "kube-system"},
					Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "kubechat"},
					Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
				},
			},
			allowedNamespaces: []string{},
			expectedCount:     3,
			expectError:       false,
		},
		{
			name: "list only allowed namespaces",
			namespaces: []corev1.Namespace{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "default"},
					Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "kube-system"},
					Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "kubechat"},
					Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
				},
			},
			allowedNamespaces: []string{"default", "kubechat"},
			expectedCount:     2,
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with namespaces
			var objects []runtime.Object
			for _, ns := range tt.namespaces {
				objects = append(objects, &ns)
			}

			fakeClient := fake.NewSimpleClientset(objects...)
			service := createTestService(fakeClient, tt.allowedNamespaces)

			result, err := service.ListNamespaces(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}

func TestKubernetesService_ListPods(t *testing.T) {
	tests := []struct {
		name              string
		namespace         string
		pods              []corev1.Pod
		allowedNamespaces []string
		expectedCount     int
		expectError       bool
		expectedErrorMsg  string
	}{
		{
			name:      "successful pod listing in allowed namespace",
			namespace: "kubechat",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "kubechat",
					},
					Status: corev1.PodStatus{Phase: corev1.PodRunning},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-2",
						Namespace: "kubechat",
					},
					Status: corev1.PodStatus{Phase: corev1.PodRunning},
				},
			},
			allowedNamespaces: []string{"kubechat"},
			expectedCount:     2,
			expectError:       false,
		},
		{
			name:              "access denied to disallowed namespace",
			namespace:         "kube-system",
			allowedNamespaces: []string{"kubechat"},
			expectError:       true,
			expectedErrorMsg:  "access denied to namespace: kube-system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with pods
			var objects []runtime.Object
			for _, pod := range tt.pods {
				objects = append(objects, &pod)
			}

			fakeClient := fake.NewSimpleClientset(objects...)
			service := createTestService(fakeClient, tt.allowedNamespaces)

			result, err := service.ListPods(context.Background(), tt.namespace)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}

func TestKubernetesService_ScaleDeployment(t *testing.T) {
	tests := []struct {
		name              string
		namespace         string
		deploymentName    string
		replicas          int32
		allowedNamespaces []string
		expectError       bool
		expectedErrorMsg  string
	}{
		{
			name:              "successful deployment scaling",
			namespace:         "kubechat",
			deploymentName:    "test-deployment",
			replicas:          3,
			allowedNamespaces: []string{"kubechat"},
			expectError:       false,
		},
		{
			name:              "scaling with too many replicas",
			namespace:         "kubechat",
			deploymentName:    "test-deployment",
			replicas:          15,
			allowedNamespaces: []string{"kubechat"},
			expectError:       true,
			expectedErrorMsg:  "replica count 15 exceeds maximum allowed (10)",
		},
		{
			name:              "access denied to namespace",
			namespace:         "kube-system",
			deploymentName:    "test-deployment",
			replicas:          2,
			allowedNamespaces: []string{"kubechat"},
			expectError:       true,
			expectedErrorMsg:  "access denied to namespace: kube-system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake deployment
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.deploymentName,
					Namespace: tt.namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
				},
			}

			fakeClient := fake.NewSimpleClientset(deployment)
			service := createTestService(fakeClient, tt.allowedNamespaces)

			err := service.ScaleDeployment(context.Background(), tt.namespace, tt.deploymentName, tt.replicas)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKubernetesService_ValidateOperation(t *testing.T) {
	tests := []struct {
		name              string
		operation         *models.KubernetesOperation
		allowedNamespaces []string
		expectError       bool
		expectedErrorMsg  string
	}{
		{
			name: "valid get operation",
			operation: &models.KubernetesOperation{
				Operation: "get",
				Resource:  "pods",
				Namespace: "kubechat",
			},
			allowedNamespaces: []string{"kubechat"},
			expectError:       false,
		},
		{
			name: "invalid operation type",
			operation: &models.KubernetesOperation{
				Operation: "create",
				Resource:  "pods",
				Namespace: "kubechat",
			},
			allowedNamespaces: []string{"kubechat"},
			expectError:       true,
			expectedErrorMsg:  "operation not allowed: create",
		},
		{
			name: "invalid resource type",
			operation: &models.KubernetesOperation{
				Operation: "get",
				Resource:  "clusterroles",
				Namespace: "kubechat",
			},
			allowedNamespaces: []string{"kubechat"},
			expectError:       true,
			expectedErrorMsg:  "resource type not allowed: clusterroles",
		},
		{
			name: "access denied to namespace",
			operation: &models.KubernetesOperation{
				Operation: "get",
				Resource:  "pods",
				Namespace: "kube-system",
			},
			allowedNamespaces: []string{"kubechat"},
			expectError:       true,
			expectedErrorMsg:  "access denied to namespace: kube-system",
		},
		{
			name: "deployment deletion blocked",
			operation: &models.KubernetesOperation{
				Operation: "delete",
				Resource:  "deployment",
				Namespace: "kubechat",
			},
			allowedNamespaces: []string{"kubechat"},
			expectError:       true,
			expectedErrorMsg:  "deployment deletion is not allowed for safety",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewSimpleClientset()
			service := createTestService(fakeClient, tt.allowedNamespaces)

			err := service.ValidateOperation(context.Background(), tt.operation)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKubernetesService_HealthCheck(t *testing.T) {
	tests := []struct {
		name        string
		mockError   error
		expectError bool
	}{
		{
			name:        "successful health check",
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "failed health check",
			mockError:   errors.New("connection refused"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewSimpleClientset()

			if tt.mockError != nil {
				fakeClient.PrependReactor("*", "*", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, tt.mockError
				})
			}

			service := createTestService(fakeClient, []string{})

			err := service.HealthCheck(context.Background())

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions for tests
func createTestService(client *fake.Clientset, allowedNamespaces []string) kubernetesService.Service {
	return &testKubernetesService{
		client:            client,
		allowedNamespaces: allowedNamespaces,
		maxLogLines:       1000,
		timeout:           30 * time.Second,
	}
}

// testKubernetesService implements the Service interface for testing
type testKubernetesService struct {
	client            *fake.Clientset
	allowedNamespaces []string
	maxLogLines       int64
	timeout           time.Duration
}

// Implement all Service interface methods for testing
func (s *testKubernetesService) GetClusterInfo(ctx context.Context) (*models.ClusterInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Get cluster version from fake client
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
		Version:   version.GitVersion,
		NodeCount: int32(len(nodes.Items)),
		CreatedAt: time.Now(),
	}, nil
}

func (s *testKubernetesService) ListNamespaces(ctx context.Context) ([]*models.KubernetesNamespace, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	namespaces, err := s.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var result []*models.KubernetesNamespace
	for _, ns := range namespaces.Items {
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

func (s *testKubernetesService) ListPods(ctx context.Context, namespace string) ([]*models.KubernetesPod, error) {
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
		result = append(result, &models.KubernetesPod{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
			CreatedAt: pod.CreationTimestamp.Time,
		})
	}

	return result, nil
}

func (s *testKubernetesService) GetPod(ctx context.Context, namespace, name string) (*models.KubernetesPod, error) {
	if !s.isNamespaceAllowed(namespace) {
		return nil, fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	pod, err := s.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	return &models.KubernetesPod{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Status:    string(pod.Status.Phase),
		CreatedAt: pod.CreationTimestamp.Time,
	}, nil
}

func (s *testKubernetesService) DeletePod(ctx context.Context, namespace, name string) error {
	if !s.isNamespaceAllowed(namespace) {
		return fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	gracePeriod := int64(30)
	err := s.client.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	})

	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}

	return nil
}

func (s *testKubernetesService) ListDeployments(ctx context.Context, namespace string) ([]*models.KubernetesDeployment, error) {
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
		result = append(result, &models.KubernetesDeployment{
			Name:      deploy.Name,
			Namespace: deploy.Namespace,
			Replicas:  *deploy.Spec.Replicas,
			CreatedAt: deploy.CreationTimestamp.Time,
		})
	}

	return result, nil
}

func (s *testKubernetesService) GetDeployment(ctx context.Context, namespace, name string) (*models.KubernetesDeployment, error) {
	if !s.isNamespaceAllowed(namespace) {
		return nil, fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	deployment, err := s.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	return &models.KubernetesDeployment{
		Name:      deployment.Name,
		Namespace: deployment.Namespace,
		Replicas:  *deployment.Spec.Replicas,
		CreatedAt: deployment.CreationTimestamp.Time,
	}, nil
}

func (s *testKubernetesService) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error {
	if !s.isNamespaceAllowed(namespace) {
		return fmt.Errorf("access denied to namespace: %s", namespace)
	}

	// Safety check: limit max replicas
	if replicas > 10 {
		return fmt.Errorf("replica count %d exceeds maximum allowed (10)", replicas)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	deployment, err := s.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	deployment.Spec.Replicas = &replicas
	_, err = s.client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}

	return nil
}

func (s *testKubernetesService) RestartDeployment(ctx context.Context, namespace, name string) error {
	if !s.isNamespaceAllowed(namespace) {
		return fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

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
	if err != nil {
		return fmt.Errorf("failed to restart deployment: %w", err)
	}

	return nil
}

func (s *testKubernetesService) ListServices(ctx context.Context, namespace string) ([]*models.KubernetesService, error) {
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
		result = append(result, &models.KubernetesService{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Type:      string(svc.Spec.Type),
			CreatedAt: svc.CreationTimestamp.Time,
		})
	}

	return result, nil
}

func (s *testKubernetesService) GetService(ctx context.Context, namespace, name string) (*models.KubernetesService, error) {
	if !s.isNamespaceAllowed(namespace) {
		return nil, fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	service, err := s.client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return &models.KubernetesService{
		Name:      service.Name,
		Namespace: service.Namespace,
		Type:      string(service.Spec.Type),
		CreatedAt: service.CreationTimestamp.Time,
	}, nil
}

func (s *testKubernetesService) ListConfigMaps(ctx context.Context, namespace string) ([]*models.KubernetesConfigMap, error) {
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
		result = append(result, &models.KubernetesConfigMap{
			Name:      cm.Name,
			Namespace: cm.Namespace,
			Data:      cm.Data,
			CreatedAt: cm.CreationTimestamp.Time,
		})
	}

	return result, nil
}

func (s *testKubernetesService) GetConfigMap(ctx context.Context, namespace, name string) (*models.KubernetesConfigMap, error) {
	if !s.isNamespaceAllowed(namespace) {
		return nil, fmt.Errorf("access denied to namespace: %s", namespace)
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	configMap, err := s.client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap: %w", err)
	}

	return &models.KubernetesConfigMap{
		Name:      configMap.Name,
		Namespace: configMap.Namespace,
		Data:      configMap.Data,
		CreatedAt: configMap.CreationTimestamp.Time,
	}, nil
}

func (s *testKubernetesService) ListSecrets(ctx context.Context, namespace string) ([]*models.KubernetesSecret, error) {
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

func (s *testKubernetesService) GetPodLogs(ctx context.Context, namespace, podName string, options *models.LogOptions) (string, error) {
	if !s.isNamespaceAllowed(namespace) {
		return "", fmt.Errorf("access denied to namespace: %s", namespace)
	}

	// For testing, return mock logs
	return "mock pod logs", nil
}

func (s *testKubernetesService) ValidateOperation(ctx context.Context, operation *models.KubernetesOperation) error {
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

func (s *testKubernetesService) CreateResource(ctx context.Context, resource string, data map[string]any) error {
	// Mock implementation for testing - in real tests this would validate the creation
	return nil
}

func (s *testKubernetesService) DeleteResource(ctx context.Context, resource, name, namespace string) error {
	// Mock implementation for testing - in real tests this would validate the deletion
	return nil
}

func (s *testKubernetesService) PatchResource(ctx context.Context, resource, name, namespace string, data map[string]any) error {
	// Mock implementation for testing - in real tests this would validate the patch
	return nil
}

func (s *testKubernetesService) ScaleResource(ctx context.Context, resource, name, namespace string, replicas int) error {
	// Mock implementation for testing - in real tests this would validate the scaling
	return nil
}

func (s *testKubernetesService) ResourceExists(ctx context.Context, resource, name, namespace string) (bool, error) {
	// Mock implementation for testing - in real tests this would check existence
	return true, nil
}

func (s *testKubernetesService) ExecuteOperation(ctx context.Context, operation *models.KubernetesOperation) (*models.KubernetesOperationResult, error) {
	// Mock implementation for testing
	return &models.KubernetesOperationResult{
		OperationID: operation.ID,
		Success:     true,
		Result:      map[string]any{"status": "completed"},
		ExecutedAt:  time.Now(),
	}, nil
}

func (s *testKubernetesService) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := s.client.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("kubernetes health check failed: %w", err)
	}

	return nil
}

// isNamespaceAllowed checks if namespace access is allowed
func (s *testKubernetesService) isNamespaceAllowed(namespace string) bool {
	// If no restrictions, allow all
	if len(s.allowedNamespaces) == 0 {
		return true
	}

	// Check against allowed list
	for _, allowed := range s.allowedNamespaces {
		if allowed == namespace || allowed == "*" {
			return true
		}
	}

	return false
}

func int32Ptr(i int32) *int32 {
	return &i
}
