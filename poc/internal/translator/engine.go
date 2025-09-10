package translator

import (
	"context"
	"fmt"
	"strings"

	"github.com/pramodksahoo/kubechat/internal/k8s"
	"github.com/pramodksahoo/kubechat/internal/llm"
)

type Engine struct {
	llmService llm.Service
	k8sClient  *k8s.Client
}

type QueryContext struct {
	Namespaces     []string `json:"namespaces"`
	CurrentContext string   `json:"current_context"`
	UserQuery      string   `json:"user_query"`
}

type TranslationResult struct {
	Command     string      `json:"command"`
	Explanation string      `json:"explanation"`
	Safety      string      `json:"safety"`
	Preview     bool        `json:"preview"`
	Result      interface{} `json:"result,omitempty"`
	Error       string      `json:"error,omitempty"`
}

func NewEngine(llmService llm.Service, k8sClient *k8s.Client) *Engine {
	return &Engine{
		llmService: llmService,
		k8sClient:  k8sClient,
	}
}

func (e *Engine) ProcessQuery(ctx context.Context, query string) (*TranslationResult, error) {
	// Build context information
	contextInfo, err := e.buildContext()
	if err != nil {
		return nil, fmt.Errorf("failed to build context: %w", err)
	}

	// Generate kubectl command using LLM
	cmdResp, err := e.llmService.GenerateKubectlCommand(ctx, query, contextInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to generate command: %w", err)
	}

	result := &TranslationResult{
		Command:     cmdResp.Command,
		Explanation: cmdResp.Explanation,
		Safety:      cmdResp.Safety,
		Preview:     e.requiresPreview(cmdResp.Safety),
	}

	// For safe read-only operations, execute immediately
	if cmdResp.Safety == "safe" {
		execResult, err := e.executeCommand(cmdResp.Command, cmdResp.Namespace)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Result = execResult
		}
	}

	return result, nil
}

func (e *Engine) ExecuteCommand(ctx context.Context, command, namespace string) (*TranslationResult, error) {
	result := &TranslationResult{
		Command: command,
		Preview: false,
	}

	execResult, err := e.executeCommand(command, namespace)
	if err != nil {
		result.Error = err.Error()
	} else {
		result.Result = execResult
	}

	return result, nil
}

func (e *Engine) buildContext() (map[string]interface{}, error) {
	namespaces, err := e.k8sClient.GetNamespaces()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"available_namespaces": namespaces,
		"default_namespace":    "default",
	}, nil
}

func (e *Engine) requiresPreview(safety string) bool {
	return safety == "warning" || safety == "dangerous"
}

func (e *Engine) executeCommand(command, namespace string) (interface{}, error) {
	// Parse the kubectl command and execute the appropriate Kubernetes API call
	parts := strings.Fields(command)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid kubectl command")
	}

	// Skip "kubectl"
	if parts[0] == "kubectl" {
		parts = parts[1:]
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("no kubectl subcommand specified")
	}

	action := parts[0]
	
	// Determine namespace from command or use provided namespace
	ns := e.extractNamespace(parts)
	if ns == "" && namespace != "" {
		ns = namespace
	}
	if ns == "" {
		ns = "default"
	}

	switch action {
	case "get":
		return e.handleGetCommand(parts[1:], ns)
	case "describe":
		return e.handleDescribeCommand(parts[1:], ns)
	case "logs":
		return e.handleLogsCommand(parts[1:], ns)
	default:
		return nil, fmt.Errorf("command '%s' not supported in PoC", action)
	}
}

func (e *Engine) extractNamespace(parts []string) string {
	for i, part := range parts {
		if (part == "-n" || part == "--namespace") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func (e *Engine) handleGetCommand(args []string, namespace string) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no resource type specified")
	}

	resource := args[0]
	
	switch resource {
	case "pods", "pod", "po":
		return e.k8sClient.GetPods(namespace)
	case "services", "service", "svc":
		return e.k8sClient.GetServices(namespace)
	case "deployments", "deployment", "deploy":
		return e.k8sClient.GetDeployments(namespace)
	case "nodes", "node", "no":
		return e.k8sClient.GetNodes()
	case "namespaces", "namespace", "ns":
		return e.k8sClient.GetNamespaces()
	default:
		return nil, fmt.Errorf("resource type '%s' not supported in PoC", resource)
	}
}

func (e *Engine) handleDescribeCommand(args []string, namespace string) (interface{}, error) {
	// For PoC, return basic information
	return map[string]string{
		"message": "Describe command recognized but not fully implemented in PoC",
		"command": fmt.Sprintf("kubectl describe %s", strings.Join(args, " ")),
	}, nil
}

func (e *Engine) handleLogsCommand(args []string, namespace string) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no pod name specified for logs")
	}

	podName := args[0]
	lines := 50 // default

	// Extract tail lines if specified
	for i, arg := range args {
		if arg == "--tail" && i+1 < len(args) {
			// In a real implementation, parse the number
			lines = 100
		}
	}

	logs, err := e.k8sClient.GetPodLogs(namespace, podName, lines)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"pod":       podName,
		"namespace": namespace,
		"logs":      logs,
	}, nil
}