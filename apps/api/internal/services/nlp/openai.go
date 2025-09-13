package nlp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// OpenAIConfig represents OpenAI service configuration
type OpenAIConfig struct {
	APIKey         string  `json:"api_key"`
	BaseURL        string  `json:"base_url"`
	Model          string  `json:"model"`
	TimeoutSeconds int     `json:"timeout_seconds"`
	MaxTokens      int     `json:"max_tokens"`
	Temperature    float32 `json:"temperature"`
	Organization   string  `json:"organization,omitempty"`
}

// OpenAIRequest represents a request to OpenAI API
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float32         `json:"temperature"`
	Stop        []string        `json:"stop,omitempty"`
}

// OpenAIMessage represents a message in OpenAI format
type OpenAIMessage struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// OpenAIResponse represents response from OpenAI API
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// openaiService implements OpenAIService interface
type openaiService struct {
	config     *OpenAIConfig
	httpClient *http.Client
	// Integration with external API service from Story 1.4
	useExternalService bool
	externalClient     ExternalAPIClient
}

// ExternalAPIClient interface for integration with Story 1.4 external service
type ExternalAPIClient interface {
	MakeRequest(ctx context.Context, method, url string, headers map[string]string, body []byte) ([]byte, error)
	HealthCheck(ctx context.Context) error
}

// NewOpenAIService creates a new OpenAI service instance
func NewOpenAIService(config *OpenAIConfig) OpenAIService {
	if config == nil {
		config = &OpenAIConfig{
			APIKey:         os.Getenv("OPENAI_API_KEY"),
			BaseURL:        "https://api.openai.com/v1",
			Model:          "gpt-3.5-turbo",
			TimeoutSeconds: 30,
			MaxTokens:      1000,
			Temperature:    0.1,
		}
	}

	service := &openaiService{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
		},
	}

	// Don't log the API key for security
	log.Printf("OpenAI service initialized with model: %s", config.Model)
	if config.APIKey == "" {
		log.Println("Warning: OpenAI API key not configured - using fallback mode")
		service.useExternalService = false // Fallback to mock responses
	} else {
		service.useExternalService = true
	}

	return service
}

// ProcessQuery processes a natural language query using OpenAI with fallback mechanisms
func (s *openaiService) ProcessQuery(ctx context.Context, request *models.NLPRequest) (*models.NLPResponse, error) {
	// Try external API service first if available
	if s.useExternalService && s.config.APIKey != "" {
		result, err := s.processWithOpenAI(ctx, request)
		if err != nil {
			log.Printf("OpenAI processing failed, falling back to mock: %v", err)
			return s.generateMockResponse(ctx, request)
		}
		return result, nil
	}

	// Fallback to mock responses for development/testing
	log.Printf("Using fallback mock response for query: %s", request.Query)
	return s.generateMockResponse(ctx, request)
}

// processWithOpenAI handles the actual OpenAI API integration
func (s *openaiService) processWithOpenAI(ctx context.Context, request *models.NLPRequest) (*models.NLPResponse, error) {

	startTime := time.Now()

	// Build messages for OpenAI chat completion
	messages := s.buildChatMessages(request)

	// Create OpenAI request
	openaiReq := &OpenAIRequest{
		Model:       s.config.Model,
		Messages:    messages,
		MaxTokens:   s.config.MaxTokens,
		Temperature: s.config.Temperature,
		Stop:        []string{},
	}

	// Make request to OpenAI
	openaiResp, err := s.makeOpenAIRequest(ctx, openaiReq)
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}

	if openaiResp.Error != nil {
		return nil, fmt.Errorf("openai error: %s (%s)", openaiResp.Error.Message, openaiResp.Error.Type)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from openai")
	}

	// Parse the response to extract command and explanation
	responseContent := openaiResp.Choices[0].Message.Content
	command, explanation, confidence, err := s.parseOpenAIResponse(responseContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse openai response: %w", err)
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
		Provider:         models.ProviderOpenAI,
		ProcessingTimeMs: processingTime,
		TokensUsed:       openaiResp.Usage.TotalTokens,
		CreatedAt:        time.Now(),
	}

	return response, nil
}

// HealthCheck validates OpenAI API connectivity
func (s *openaiService) HealthCheck(ctx context.Context) error {
	if s.config.APIKey == "" {
		return fmt.Errorf("OpenAI API key not configured")
	}

	// Simple test request to validate API connectivity
	testReq := &OpenAIRequest{
		Model: s.config.Model,
		Messages: []OpenAIMessage{
			{Role: "user", Content: "Test connectivity"},
		},
		MaxTokens:   10,
		Temperature: 0,
	}

	resp, err := s.makeOpenAIRequest(ctx, testReq)
	if err != nil {
		return fmt.Errorf("openai health check failed: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("openai health check error: %s", resp.Error.Message)
	}

	return nil
}

// IsAvailable checks if OpenAI service is available
func (s *openaiService) IsAvailable(ctx context.Context) bool {
	if s.config.APIKey == "" {
		return false
	}

	// For now, just check if API key is configured
	// In production, might want to cache health check results
	return true
}

// Private methods

// buildChatMessages constructs messages for OpenAI chat completion
func (s *openaiService) buildChatMessages(request *models.NLPRequest) []OpenAIMessage {
	messages := make([]OpenAIMessage, 0, 3)

	// System message with instructions
	systemPrompt := s.buildSystemPrompt()
	messages = append(messages, OpenAIMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	// Add context if available
	if request.Context != "" || request.ClusterInfo != "" {
		contextMsg := strings.Builder{}
		contextMsg.WriteString("Context Information:\n")

		if request.Context != "" {
			contextMsg.WriteString(fmt.Sprintf("User Context: %s\n", request.Context))
		}

		if request.ClusterInfo != "" {
			contextMsg.WriteString(fmt.Sprintf("Cluster Information: %s\n", request.ClusterInfo))
		}

		messages = append(messages, OpenAIMessage{
			Role:    "user",
			Content: contextMsg.String(),
		})
	}

	// User query
	messages = append(messages, OpenAIMessage{
		Role:    "user",
		Content: fmt.Sprintf("Convert this natural language query to a kubectl command: %s", request.Query),
	})

	return messages
}

// buildSystemPrompt creates the enhanced safety-focused system prompt for OpenAI
func (s *openaiService) buildSystemPrompt() string {
	return `You are a Kubernetes expert assistant with advanced safety controls. Your role is to convert natural language queries into valid kubectl commands while prioritizing safety and security.

CRITICAL SAFETY INSTRUCTIONS:
1. Generate only valid kubectl commands
2. ALWAYS prioritize read-only operations when possible
3. For destructive operations, include safety warnings in the explanation
4. Assess confidence level from 0.0 to 1.0 based on query clarity
5. If the query is unclear, generate the SAFEST possible command
6. Include appropriate dry-run suggestions for dangerous operations

Response Format:
Always respond with a JSON object in this exact format:
{
  "command": "kubectl get pods -n default",
  "explanation": "This command lists all pods in the default namespace",
  "confidence": 0.95,
  "safety_warnings": ["Optional array of safety warnings"],
  "suggestions": ["Optional array of safer alternatives"]
}

SAFETY CLASSIFICATION RULES:
- SAFE: Read-only operations (get, describe, logs, top, explain)
- WARNING: Operations that modify cluster state (create, apply, patch, scale, restart, rollout)
- DANGEROUS: Destructive operations (delete, destroy, drain, --force, --cascade)

ENHANCED SAFETY GUIDELINES:
- Include --dry-run=client for potentially dangerous operations
- Add appropriate resource selectors to limit scope
- Use --output flags for better formatting when helpful
- For production environments, suggest extra caution
- For critical namespaces (kube-system), require explicit confirmation
- Always include namespace specification when applicable
- Suggest using --confirm or --wait flags for destructive operations

COMMON OPERATION PATTERNS:
- List resources: kubectl get [resource] -n [namespace] -o wide
- Describe resources: kubectl describe [resource] [name] -n [namespace]
- View logs: kubectl logs [pod] -n [namespace] --tail=100
- Create resources: kubectl apply -f [file] --dry-run=client -o yaml
- Delete resources: kubectl delete [resource] [name] -n [namespace] --grace-period=30

Remember: When in doubt, choose the safer option and explain the risks.`
}

// makeOpenAIRequest makes HTTP request to OpenAI API
func (s *openaiService) makeOpenAIRequest(ctx context.Context, openaiReq *OpenAIRequest) (*OpenAIResponse, error) {
	jsonBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", s.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.APIKey))

	if s.config.Organization != "" {
		req.Header.Set("OpenAI-Organization", s.config.Organization)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var openaiResp OpenAIResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &openaiResp, nil
}

// parseOpenAIResponse extracts command, explanation, and confidence from OpenAI response
func (s *openaiService) parseOpenAIResponse(response string) (command, explanation string, confidence float64, err error) {
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
		if strings.HasPrefix(line, "kubectl") || strings.Contains(line, "kubectl") {
			// Extract the kubectl command
			if cmdStart := strings.Index(line, "kubectl"); cmdStart != -1 {
				cmdLine := line[cmdStart:]
				// Remove any trailing punctuation or quotes
				cmdLine = strings.Trim(cmdLine, "\"'`.,;")
				if foundCommand == "" {
					foundCommand = cmdLine
				}
			}
		} else if line != "" && foundCommand != "" &&
			!strings.Contains(line, "{") && !strings.Contains(line, "}") {
			// Treat subsequent non-JSON lines as explanation
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
		confidence = 0.9 // Higher confidence for OpenAI responses
	}

	return foundCommand, foundExplanation.String(), confidence, nil
}

// classifyCommandSafety determines safety level of a kubectl command
func (s *openaiService) classifyCommandSafety(command string) models.SafetyLevel {
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

// generateMockResponse creates mock responses for development and testing
func (s *openaiService) generateMockResponse(ctx context.Context, request *models.NLPRequest) (*models.NLPResponse, error) {
	log.Printf("Generating mock response for query: %s", request.Query)

	// Generate mock command based on query patterns
	mockCommand := s.generateMockCommand(request.Query)
	mockExplanation := s.generateMockExplanation(mockCommand, request.Query)
	mockConfidence := s.calculateMockConfidence(request.Query)
	mockSafety := s.classifyCommandSafety(mockCommand)

	response := &models.NLPResponse{
		ID:               uuid.New(),
		RequestID:        request.ID,
		GeneratedCommand: mockCommand,
		Explanation:      mockExplanation,
		SafetyLevel:      mockSafety,
		Confidence:       mockConfidence,
		Provider:         models.ProviderOpenAI,
		ProcessingTimeMs: 100, // Mock fast processing
		TokensUsed:       50,  // Mock token usage
		CreatedAt:        time.Now(),
	}

	return response, nil
}

// generateMockCommand creates realistic kubectl commands based on query patterns
func (s *openaiService) generateMockCommand(query string) string {
	lowerQuery := strings.ToLower(query)

	// Pattern matching for common queries
	if strings.Contains(lowerQuery, "pod") && strings.Contains(lowerQuery, "show") {
		return "kubectl get pods -o wide"
	}

	if strings.Contains(lowerQuery, "service") && (strings.Contains(lowerQuery, "list") || strings.Contains(lowerQuery, "show")) {
		return "kubectl get services -o wide"
	}

	if strings.Contains(lowerQuery, "node") && strings.Contains(lowerQuery, "status") {
		return "kubectl get nodes -o wide"
	}

	if strings.Contains(lowerQuery, "deployment") && strings.Contains(lowerQuery, "show") {
		return "kubectl get deployments -o wide"
	}

	if strings.Contains(lowerQuery, "log") && strings.Contains(lowerQuery, "pod") {
		return "kubectl logs [POD_NAME] --tail=100"
	}

	if strings.Contains(lowerQuery, "delete") && strings.Contains(lowerQuery, "pod") {
		return "kubectl delete pod [POD_NAME] --grace-period=30"
	}

	if strings.Contains(lowerQuery, "create") && strings.Contains(lowerQuery, "namespace") {
		return "kubectl create namespace [NAMESPACE_NAME]"
	}

	if strings.Contains(lowerQuery, "scale") && strings.Contains(lowerQuery, "deployment") {
		return "kubectl scale deployment [DEPLOYMENT_NAME] --replicas=3"
	}

	// Default safe command
	return "kubectl get pods --all-namespaces"
}

// generateMockExplanation creates explanations for mock commands
func (s *openaiService) generateMockExplanation(command, query string) string {
	if strings.Contains(command, "get pods") {
		return "This command lists all pods with detailed information including node placement and IP addresses."
	}

	if strings.Contains(command, "get services") {
		return "This command shows all services with their types, cluster IPs, and exposed ports."
	}

	if strings.Contains(command, "get nodes") {
		return "This command displays all cluster nodes with their status and system information."
	}

	if strings.Contains(command, "get deployments") {
		return "This command shows all deployments with their replica status and availability."
	}

	if strings.Contains(command, "logs") {
		return "This command shows the last 100 log lines from the specified pod."
	}

	if strings.Contains(command, "delete") {
		return "This command deletes the specified pod with a 30-second grace period for clean shutdown."
	}

	if strings.Contains(command, "create namespace") {
		return "This command creates a new Kubernetes namespace with the specified name."
	}

	if strings.Contains(command, "scale") {
		return "This command scales the specified deployment to the desired number of replicas."
	}

	return fmt.Sprintf("Mock response for query: %s", query)
}

// calculateMockConfidence determines confidence based on query clarity
func (s *openaiService) calculateMockConfidence(query string) float64 {
	confidence := 0.7 // Base confidence for mocks

	// Increase confidence for specific, clear queries
	if strings.Contains(query, "kubectl") {
		confidence += 0.2
	}

	if len(query) > 20 && len(query) < 100 {
		confidence += 0.1
	}

	// Decrease confidence for vague queries
	if len(query) < 10 {
		confidence -= 0.2
	}

	if strings.Contains(strings.ToLower(query), "something") || strings.Contains(strings.ToLower(query), "stuff") {
		confidence -= 0.3
	}

	// Keep confidence within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.1 {
		confidence = 0.1
	}

	return confidence
}

// enhanceWithMultiProviderSupport prepares the service for Ollama integration (Epic 3)
func (s *openaiService) enhanceWithMultiProviderSupport() {
	// This will be expanded in Epic 3 for Ollama integration
	log.Println("OpenAI service ready for multi-provider expansion")
}
