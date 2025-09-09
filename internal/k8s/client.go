package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Client struct {
	clientset     *kubernetes.Clientset
	metricsClient *metricsv1beta1.Clientset
	config        *rest.Config
}

type PodInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Status    string            `json:"status"`
	Ready     string            `json:"ready"`
	Age       string            `json:"age"`
	Node      string            `json:"node"`
	Labels    map[string]string `json:"labels"`
}

type NodeInfo struct {
	Name   string            `json:"name"`
	Status string            `json:"status"`
	Roles  []string          `json:"roles"`
	Age    string            `json:"age"`
	Labels map[string]string `json:"labels"`
}

type ServiceInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	ClusterIP string            `json:"clusterIP"`
	Ports     []string          `json:"ports"`
	Labels    map[string]string `json:"labels"`
}

func NewClient(kubeConfigPath string) (*Client, error) {
	var config *rest.Config
	var err error

	if kubeConfigPath == "" || !fileExists(kubeConfigPath) {
		// Try in-cluster config first
		config, err = rest.InClusterConfig()
		if err != nil {
			// Fallback to default kubeconfig location
			kubeConfigPath = filepath.Join(homeDir(), ".kube", "config")
		}
	}

	if config == nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	metricsClient, err := metricsv1beta1.NewForConfig(config)
	if err != nil {
		// Metrics server might not be available, continue without it
		metricsClient = nil
	}

	return &Client{
		clientset:     clientset,
		metricsClient: metricsClient,
		config:        config,
	}, nil
}

func (c *Client) GetPods(namespace string) ([]PodInfo, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	var podInfos []PodInfo
	for _, pod := range pods.Items {
		ready := fmt.Sprintf("%d/%d", readyContainers(pod), len(pod.Spec.Containers))
		
		podInfo := PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
			Ready:     ready,
			Age:       calculateAge(pod.CreationTimestamp.Time),
			Node:      pod.Spec.NodeName,
			Labels:    pod.Labels,
		}
		podInfos = append(podInfos, podInfo)
	}

	return podInfos, nil
}

func (c *Client) GetNodes() ([]NodeInfo, error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	var nodeInfos []NodeInfo
	for _, node := range nodes.Items {
		roles := []string{}
		if _, exists := node.Labels["node-role.kubernetes.io/control-plane"]; exists {
			roles = append(roles, "control-plane")
		}
		if _, exists := node.Labels["node-role.kubernetes.io/master"]; exists {
			roles = append(roles, "master")
		}
		if len(roles) == 0 {
			roles = append(roles, "worker")
		}

		status := "Unknown"
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				if condition.Status == corev1.ConditionTrue {
					status = "Ready"
				} else {
					status = "NotReady"
				}
				break
			}
		}

		nodeInfo := NodeInfo{
			Name:   node.Name,
			Status: status,
			Roles:  roles,
			Age:    calculateAge(node.CreationTimestamp.Time),
			Labels: node.Labels,
		}
		nodeInfos = append(nodeInfos, nodeInfo)
	}

	return nodeInfos, nil
}

func (c *Client) GetServices(namespace string) ([]ServiceInfo, error) {
	services, err := c.clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	var serviceInfos []ServiceInfo
	for _, svc := range services.Items {
		ports := []string{}
		for _, port := range svc.Spec.Ports {
			portStr := fmt.Sprintf("%d/%s", port.Port, port.Protocol)
			if port.NodePort != 0 {
				portStr += fmt.Sprintf(":%d", port.NodePort)
			}
			ports = append(ports, portStr)
		}

		serviceInfo := ServiceInfo{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Type:      string(svc.Spec.Type),
			ClusterIP: svc.Spec.ClusterIP,
			Ports:     ports,
			Labels:    svc.Labels,
		}
		serviceInfos = append(serviceInfos, serviceInfo)
	}

	return serviceInfos, nil
}

func (c *Client) GetDeployments(namespace string) ([]appsv1.Deployment, error) {
	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments: %w", err)
	}

	return deployments.Items, nil
}

func (c *Client) GetPodLogs(namespace, podName string, lines int) (string, error) {
	req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		TailLines: int64Ptr(int64(lines)),
	})

	logs, err := req.Stream(context.TODO())
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer logs.Close()

	buf := make([]byte, 2048)
	n, err := logs.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return "", fmt.Errorf("failed to read pod logs: %w", err)
	}

	return string(buf[:n]), nil
}

func (c *Client) GetNamespaces() ([]string, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces: %w", err)
	}

	var names []string
	for _, ns := range namespaces.Items {
		names = append(names, ns.Name)
	}

	return names, nil
}

// Helper functions
func readyContainers(pod corev1.Pod) int {
	ready := 0
	for _, status := range pod.Status.ContainerStatuses {
		if status.Ready {
			ready++
		}
	}
	return ready
}

func calculateAge(creationTime time.Time) string {
	duration := time.Since(creationTime)
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	
	if days > 0 {
		return fmt.Sprintf("%dd", days)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", int(duration.Minutes()))
}

func int64Ptr(i int64) *int64 {
	return &i
}

func homeDir() string {
	if h := getEnv("HOME"); h != "" {
		return h
	}
	return getEnv("USERPROFILE") // Windows
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func getEnv(key string) string {
	return os.Getenv(key)
}