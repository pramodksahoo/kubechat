package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	client *openai.Client
	model  string
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	if model == "" {
		model = openai.GPT4
	}

	return &OpenAIClient{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

func (c *OpenAIClient) GenerateCommand(ctx context.Context, prompt string) (*CommandResponse, error) {
	req := openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are KubeChat, an expert Kubernetes assistant. Always respond with valid JSON containing kubectl commands.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   500,
		Temperature: 0.1, // Low temperature for more deterministic responses
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	return c.parseCommandResponse(resp.Choices[0].Message.Content)
}

func (c *OpenAIClient) ExplainCommand(ctx context.Context, prompt string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are KubeChat, an expert Kubernetes assistant. Provide clear, concise explanations of kubectl commands.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   300,
		Temperature: 0.3,
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) parseCommandResponse(response string) (*CommandResponse, error) {
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

	// Validate and clean up the command
	if !strings.HasPrefix(cmdResp.Command, "kubectl") {
		cmdResp.Command = "kubectl " + cmdResp.Command
	}

	// Validate safety level
	if cmdResp.Safety == "" {
		cmdResp.Safety = "safe"
	}

	return &cmdResp, nil
}