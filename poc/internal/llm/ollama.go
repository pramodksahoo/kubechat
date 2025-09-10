package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OllamaClient struct {
	baseURL string
	model   string
	client  *http.Client
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func NewOllamaClient(baseURL, model string) (*OllamaClient, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama2"
	}

	client := &OllamaClient{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}

	// Test connection
	if err := client.testConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}

	return client, nil
}

func (c *OllamaClient) testConnection() error {
	resp, err := c.client.Get(c.baseURL + "/api/tags")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama server returned status %d", resp.StatusCode)
	}

	return nil
}

func (c *OllamaClient) GenerateCommand(ctx context.Context, prompt string) (*CommandResponse, error) {
	request := OllamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Parse the response to extract JSON
	return c.parseCommandResponse(ollamaResp.Response)
}

func (c *OllamaClient) ExplainCommand(ctx context.Context, prompt string) (string, error) {
	request := OllamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return ollamaResp.Response, nil
}

func (c *OllamaClient) parseCommandResponse(response string) (*CommandResponse, error) {
	// Try to find JSON in the response
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	
	if start == -1 || end == -1 {
		// Fallback: create a basic response
		return &CommandResponse{
			Command:     "kubectl get pods",
			Explanation: "I couldn't parse a specific command from your request. Here's a basic pod listing command.",
			Safety:      "safe",
		}, nil
	}

	jsonStr := response[start : end+1]
	var cmdResp CommandResponse
	
	if err := json.Unmarshal([]byte(jsonStr), &cmdResp); err != nil {
		// Fallback on parse error
		return &CommandResponse{
			Command:     "kubectl get pods",
			Explanation: "I had trouble parsing your request. Here's a basic pod listing command.",
			Safety:      "safe",
		}, nil
	}

	// Validate the command
	if !strings.HasPrefix(cmdResp.Command, "kubectl") {
		cmdResp.Command = "kubectl " + cmdResp.Command
	}

	return &cmdResp, nil
}