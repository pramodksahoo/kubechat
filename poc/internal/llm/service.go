package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pramodksahoo/kubechat/internal/config"
)

type Service interface {
	GenerateKubectlCommand(ctx context.Context, query string, contextInfo map[string]interface{}) (*CommandResponse, error)
	ExplainCommand(ctx context.Context, command string) (string, error)
}

type CommandResponse struct {
	Command     string `json:"command"`
	Explanation string `json:"explanation"`
	Safety      string `json:"safety"` // "safe", "warning", "dangerous"
	Namespace   string `json:"namespace,omitempty"`
}

type service struct {
	ollama  *OllamaClient
	openai  *OpenAIClient
	config  config.LLMConfig
}

func NewService(cfg config.LLMConfig) (Service, error) {
	s := &service{
		config: cfg,
	}

	// Initialize primary provider
	switch cfg.Provider {
	case "ollama":
		client, err := NewOllamaClient(cfg.OllamaURL, cfg.OllamaModel)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Ollama client: %w", err)
		}
		s.ollama = client

	case "openai":
		if cfg.OpenAIKey == "" {
			return nil, fmt.Errorf("OpenAI API key is required")
		}
		s.openai = NewOpenAIClient(cfg.OpenAIKey, cfg.OpenAIModel)

	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}

	// Initialize fallback if enabled
	if cfg.EnableFallback {
		if cfg.Provider == "ollama" && cfg.OpenAIKey != "" {
			s.openai = NewOpenAIClient(cfg.OpenAIKey, cfg.OpenAIModel)
		} else if cfg.Provider == "openai" {
			client, _ := NewOllamaClient(cfg.OllamaURL, cfg.OllamaModel)
			s.ollama = client
		}
	}

	return s, nil
}

func (s *service) GenerateKubectlCommand(ctx context.Context, query string, contextInfo map[string]interface{}) (*CommandResponse, error) {
	prompt := s.buildPrompt(query, contextInfo)

	// Try primary provider
	var response *CommandResponse
	var err error

	switch s.config.Provider {
	case "ollama":
		if s.ollama != nil {
			response, err = s.ollama.GenerateCommand(ctx, prompt)
		}
	case "openai":
		if s.openai != nil {
			response, err = s.openai.GenerateCommand(ctx, prompt)
		}
	}

	// Try fallback if primary fails
	if err != nil && s.config.EnableFallback {
		if s.config.Provider == "ollama" && s.openai != nil {
			response, err = s.openai.GenerateCommand(ctx, prompt)
		} else if s.config.Provider == "openai" && s.ollama != nil {
			response, err = s.ollama.GenerateCommand(ctx, prompt)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate command: %w", err)
	}

	return response, nil
}

func (s *service) ExplainCommand(ctx context.Context, command string) (string, error) {
	prompt := fmt.Sprintf(`Explain what this kubectl command does in simple terms:

Command: %s

Please explain:
1. What operation this performs
2. What resources it affects  
3. Any potential risks or side effects
4. Whether it's a read-only or write operation

Keep the explanation clear and concise for DevOps teams.`, command)

	switch s.config.Provider {
	case "ollama":
		if s.ollama != nil {
			return s.ollama.ExplainCommand(ctx, prompt)
		}
	case "openai":
		if s.openai != nil {
			return s.openai.ExplainCommand(ctx, prompt)
		}
	}

	return "", fmt.Errorf("no LLM provider available for explanation")
}

func (s *service) buildPrompt(query string, contextInfo map[string]interface{}) string {
	contextJSON, _ := json.Marshal(contextInfo)
	
	return fmt.Sprintf(`You are KubeChat, an AI assistant that converts natural language queries into kubectl commands for Kubernetes cluster management.

CONTEXT:
%s

USER QUERY: "%s"

IMPORTANT RULES:
1. Generate ONLY valid kubectl commands
2. Always include appropriate namespace flags if mentioned
3. Use proper resource names and selectors
4. For dangerous operations, set safety to "dangerous" 
5. For write operations, set safety to "warning"
6. For read-only operations, set safety to "safe"
7. Provide clear explanations of what the command does
8. If the query is unclear, ask for clarification

RESPONSE FORMAT (JSON):
{
  "command": "kubectl get pods -n default",
  "explanation": "This command lists all pods in the default namespace",
  "safety": "safe",
  "namespace": "default"
}

Generate the kubectl command for the user's query:`, string(contextJSON), query)
}