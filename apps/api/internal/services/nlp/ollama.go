package nlp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// OllamaConfig represents Ollama service configuration
type OllamaConfig struct {
	BaseURL        string  `json:"base_url"`
	Model          string  `json:"model"`
	TimeoutSeconds int     `json:"timeout_seconds"`
	MaxTokens      int     `json:"max_tokens"`
	Temperature    float32 `json:"temperature"`
	EnableStream   bool    `json:"enable_stream"`
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	Stream  bool           `json:"stream"`
	Options *OllamaOptions `json:"options,omitempty"`
}

// OllamaOptions represents Ollama generation options
type OllamaOptions struct {
	Temperature float32 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"` // max tokens
}

// OllamaResponse represents response from Ollama API
type OllamaResponse struct {
	Model              string `json:"model"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

// ollamaService implements OllamaService interface
type ollamaService struct {
	config     *OllamaConfig
	httpClient *http.Client
}

// NewOllamaService creates a new Ollama service instance
func NewOllamaService(config *OllamaConfig) OllamaService {
	if config == nil {
		config = &OllamaConfig{
			BaseURL:        "http://kubechat-dev-ollama:11434",
			Model:          "llama3.2:3b",
			TimeoutSeconds: 30,
			MaxTokens:      1000,
			Temperature:    0.1,
			EnableStream:   false,
		}
	}

	service := &ollamaService{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
		},
	}

	log.Printf("Ollama service initialized with model: %s, base URL: %s", config.Model, config.BaseURL)
	return service
}

// ProcessQuery processes a natural language query using Ollama
func (s *ollamaService) ProcessQuery(ctx context.Context, request *models.NLPRequest) (*models.NLPResponse, error) {
	startTime := time.Now()

	// Build the prompt for kubectl command generation
	prompt := s.buildKubectlPrompt(request)

	// Create Ollama request
	ollamaReq := &OllamaRequest{
		Model:  s.config.Model,
		Prompt: prompt,
		Stream: s.config.EnableStream,
		Options: &OllamaOptions{
			Temperature: s.config.Temperature,
			NumPredict:  s.config.MaxTokens,
		},
	}

	// Make request to Ollama
	ollamaResp, err := s.makeOllamaRequest(ctx, ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}

	// Parse the response to extract command and explanation
	command, explanation, confidence, err := s.parseOllamaResponse(ollamaResp.Response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ollama response: %w", err)
	}

	// Calculate processing time
	processingTime := time.Since(startTime).Milliseconds()

	// Create NLP response
	response := &models.NLPResponse{
		ID:               uuid.New(),
		RequestID:        request.ID,
		GeneratedCommand: command,
		Explanation:      explanation,
		SafetyLevel:      s.classifyCommandSafety(command),
		Confidence:       confidence,
		Provider:         models.ProviderOllama,
		ProcessingTimeMs: processingTime,
		TokensUsed:       ollamaResp.EvalCount,
		CreatedAt:        time.Now(),
	}

	return response, nil
}

// HealthCheck validates Ollama connectivity
func (s *ollamaService) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/tags", s.config.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ollama health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama health check returned status: %d", resp.StatusCode)
	}

	return nil
}

// IsAvailable checks if Ollama service is available
func (s *ollamaService) IsAvailable(ctx context.Context) bool {
	return s.HealthCheck(ctx) == nil
}

// Private methods

// buildKubectlPrompt constructs a prompt for kubectl command generation
func (s *ollamaService) buildKubectlPrompt(request *models.NLPRequest) string {
	var prompt strings.Builder

	prompt.WriteString("You are a Kubernetes expert assistant. Convert the following natural language query into a valid kubectl command.\n\n")

	// Add context if available
	if request.Context != "" {
		prompt.WriteString(fmt.Sprintf("Context: %s\n\n", request.Context))
	}

	if request.ClusterInfo != "" {
		prompt.WriteString(fmt.Sprintf("Cluster Information: %s\n\n", request.ClusterInfo))
	}

	prompt.WriteString(fmt.Sprintf("User Query: %s\n\n", request.Query))

	prompt.WriteString("Please provide your response in the following JSON format:\n")
	prompt.WriteString("{\n")
	prompt.WriteString("  \"command\": \"kubectl get pods -n default\",\n")
	prompt.WriteString("  \"explanation\": \"This command lists all pods in the default namespace\",\n")
	prompt.WriteString("  \"confidence\": 0.95\n")
	prompt.WriteString("}\n\n")

	prompt.WriteString("Requirements:\n")
	prompt.WriteString("- Only generate valid kubectl commands\n")
	prompt.WriteString("- Include namespace if not specified in query\n")
	prompt.WriteString("- Use appropriate kubectl flags and options\n")
	prompt.WriteString("- Provide confidence score between 0.0 and 1.0\n")
	prompt.WriteString("- If the query is unclear, generate the safest possible command\n")
	prompt.WriteString("- Always start kubectl commands with 'kubectl'\n")

	return prompt.String()
}

// makeOllamaRequest makes HTTP request to Ollama API
func (s *ollamaService) makeOllamaRequest(ctx context.Context, ollamaReq *OllamaRequest) (*OllamaResponse, error) {
	jsonBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", s.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle streaming vs non-streaming responses
	if ollamaReq.Stream {
		return s.parseStreamingResponse(body)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &ollamaResp, nil
}

// parseStreamingResponse handles streaming responses from Ollama
func (s *ollamaService) parseStreamingResponse(body []byte) (*OllamaResponse, error) {
	lines := bytes.Split(body, []byte("\n"))
	var finalResponse OllamaResponse
	var fullResponse strings.Builder

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		var resp OllamaResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			continue // Skip invalid JSON lines
		}

		fullResponse.WriteString(resp.Response)

		if resp.Done {
			finalResponse = resp
			finalResponse.Response = fullResponse.String()
			break
		}
	}

	return &finalResponse, nil
}

// parseOllamaResponse extracts command, explanation, and confidence from Ollama response
func (s *ollamaService) parseOllamaResponse(response string) (command, explanation string, confidence float64, err error) {
	// Default values
	command = ""
	explanation = ""
	confidence = 0.5

	// Try to parse as JSON first
	var jsonResp struct {
		Command     string  `json:"command"`
		Explanation string  `json:"explanation"`
		Confidence  float64 `json:"confidence"`
	}

	// Look for JSON in the response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonStr := response[jsonStart : jsonEnd+1]
		if err := json.Unmarshal([]byte(jsonStr), &jsonResp); err == nil {
			return jsonResp.Command, jsonResp.Explanation, jsonResp.Confidence, nil
		}
	}

	// Fallback: parse response text for kubectl commands
	lines := strings.Split(response, "\n")
	var foundCommand string
	var foundExplanation strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for kubectl command
		if strings.HasPrefix(line, "kubectl") {
			if foundCommand == "" {
				foundCommand = line
			}
		} else if line != "" && foundCommand != "" {
			// Treat subsequent lines as explanation
			if foundExplanation.Len() > 0 {
				foundExplanation.WriteString(" ")
			}
			foundExplanation.WriteString(line)
		}
	}

	if foundCommand == "" {
		// If no kubectl command found, generate a safe default
		foundCommand = "kubectl get pods --all-namespaces"
		foundExplanation.WriteString("Default command to list all pods across namespaces")
		confidence = 0.3
	} else {
		confidence = 0.8 // Higher confidence if we found a command
	}

	return foundCommand, foundExplanation.String(), confidence, nil
}

// classifyCommandSafety determines safety level of a kubectl command
func (s *ollamaService) classifyCommandSafety(command string) models.SafetyLevel {
	if command == "" {
		return models.NLPSafetyLevelDangerous
	}

	lowerCmd := strings.ToLower(command)

	// Dangerous operations
	dangerousPatterns := []string{
		"delete", "destroy", "rm", "--force", "--cascade=foreground",
		"drain", "cordon", "evict", "--grace-period=0",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerCmd, pattern) {
			return models.NLPSafetyLevelDangerous
		}
	}

	// Warning-level operations
	warningPatterns := []string{
		"create", "apply", "patch", "replace", "scale", "restart",
		"edit", "label", "annotate", "expose", "rollout",
	}

	for _, pattern := range warningPatterns {
		if strings.Contains(lowerCmd, pattern) {
			return models.NLPSafetyLevelWarning
		}
	}

	// Safe operations (read-only)
	return models.NLPSafetyLevelSafe
}
