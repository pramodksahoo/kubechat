package external

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/audit"
	"golang.org/x/time/rate"
)

// OpenAIClient implements APIClient for OpenAI integration
type OpenAIClient struct {
	config         *ClientConfig
	httpClient     *http.Client
	rateLimiter    *rate.Limiter
	circuitBreaker *CircuitBreaker
	auditSvc       audit.Service
	metrics        *OpenAIMetrics
	mu             sync.RWMutex
}

// OpenAIMetrics tracks OpenAI-specific metrics
type OpenAIMetrics struct {
	RequestCount        int64            `json:"request_count"`
	SuccessfulRequests  int64            `json:"successful_requests"`
	FailedRequests      int64            `json:"failed_requests"`
	TotalTokensUsed     int64            `json:"total_tokens_used"`
	TotalCost           float64          `json:"total_cost"`
	AverageLatency      float64          `json:"average_latency_ms"`
	RateLimitHits       int64            `json:"rate_limit_hits"`
	CircuitBreakerTrips int64            `json:"circuit_breaker_trips"`
	LastRequestTime     time.Time        `json:"last_request_time"`
	ModelUsage          map[string]int64 `json:"model_usage"`
	mu                  sync.RWMutex
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	maxFailures   int
	resetTimeout  time.Duration
	failures      int
	lastFailTime  time.Time
	state         CircuitState
	mu            sync.RWMutex
	halfOpenCalls int
	maxHalfOpen   int
}

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// OpenAI API request/response structures
type OpenAIChatRequest struct {
	Model            string                `json:"model"`
	Messages         []OpenAIChatMessage   `json:"messages"`
	MaxTokens        int                   `json:"max_tokens,omitempty"`
	Temperature      float32               `json:"temperature,omitempty"`
	TopP             float32               `json:"top_p,omitempty"`
	FrequencyPenalty float32               `json:"frequency_penalty,omitempty"`
	PresencePenalty  float32               `json:"presence_penalty,omitempty"`
	Stop             []string              `json:"stop,omitempty"`
	User             string                `json:"user,omitempty"`
	Stream           bool                  `json:"stream,omitempty"`
	Tools            []OpenAITool          `json:"tools,omitempty"`
	ToolChoice       interface{}           `json:"tool_choice,omitempty"`
	ResponseFormat   *OpenAIResponseFormat `json:"response_format,omitempty"`
}

type OpenAIChatMessage struct {
	Role       string           `json:"role"`
	Content    interface{}      `json:"content"`
	Name       string           `json:"name,omitempty"`
	ToolCalls  []OpenAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type OpenAITool struct {
	Type     string         `json:"type"`
	Function OpenAIFunction `json:"function"`
}

type OpenAIFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type OpenAIToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function OpenAIFunctionCall `json:"function"`
}

type OpenAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type OpenAIResponseFormat struct {
	Type string `json:"type"`
}

type OpenAIChatResponse struct {
	ID                string         `json:"id"`
	Object            string         `json:"object"`
	Created           int64          `json:"created"`
	Model             string         `json:"model"`
	SystemFingerprint string         `json:"system_fingerprint"`
	Choices           []OpenAIChoice `json:"choices"`
	Usage             OpenAIUsage    `json:"usage"`
	Error             *OpenAIError   `json:"error,omitempty"`
}

type OpenAIChoice struct {
	Index        int               `json:"index"`
	Message      OpenAIChatMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
	Logprobs     interface{}       `json:"logprobs"`
}

type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIError struct {
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Param   interface{} `json:"param"`
	Code    interface{} `json:"code"`
}

// NewOpenAIClient creates a new OpenAI API client
func NewOpenAIClient(config *ClientConfig, auditSvc audit.Service) (*OpenAIClient, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if config.APIKey == "" && config.SecretKeyName == "" {
		return nil, fmt.Errorf("API key or secret key name is required")
	}

	// Create HTTP client with custom transport
	transport := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	// Add proxy support if configured
	if config.ProxyURL != "" {
		proxyURL, err := url.Parse(config.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	// Initialize rate limiter
	rateLimit := rate.Every(time.Minute / time.Duration(config.RateLimitPerMin))
	rateLimiter := rate.NewLimiter(rateLimit, config.RateLimitBurst)

	// Initialize circuit breaker
	circuitBreaker := &CircuitBreaker{
		maxFailures:  config.FailureThreshold,
		resetTimeout: config.RecoveryTimeout,
		state:        CircuitClosed,
		maxHalfOpen:  config.HalfOpenMaxCalls,
	}

	// Initialize metrics
	metrics := &OpenAIMetrics{
		ModelUsage: make(map[string]int64),
	}

	client := &OpenAIClient{
		config:         config,
		httpClient:     httpClient,
		rateLimiter:    rateLimiter,
		circuitBreaker: circuitBreaker,
		auditSvc:       auditSvc,
		metrics:        metrics,
	}

	log.Printf("OpenAI client initialized for %s", config.BaseURL)
	return client, nil
}

// Request makes an HTTP request to OpenAI API with full feature support
func (c *OpenAIClient) Request(ctx context.Context, req *APIRequest) (*APIResponse, error) {
	startTime := time.Now()
	requestID := req.RequestID
	if requestID == "" {
		requestID = uuid.New().String()
	}

	// Check circuit breaker
	if !c.circuitBreaker.allowRequest() {
		c.updateMetrics(false, 0, 0, 0, time.Since(startTime))
		return nil, fmt.Errorf("circuit breaker is open")
	}

	// Apply rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		c.updateMetrics(false, 0, 0, 0, time.Since(startTime))
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	var lastErr error
	maxRetries := c.config.MaxRetries

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoffDuration := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoffDuration):
				log.Printf("Retrying OpenAI request (attempt %d/%d) after %v",
					attempt+1, maxRetries+1, backoffDuration)
			}
		}

		resp, err := c.makeHTTPRequest(ctx, req, requestID, attempt)
		if err == nil {
			c.circuitBreaker.onSuccess()
			c.updateMetrics(true, resp.TokensUsed, int64(resp.RequestSize), int64(resp.ResponseSize), resp.ResponseTime)
			c.logAuditTrail(req, resp, nil)
			return resp, nil
		}

		lastErr = err

		// Check if error is retryable
		if !c.isRetryableError(err) {
			break
		}
	}

	// All attempts failed
	c.circuitBreaker.onFailure()
	c.updateMetrics(false, 0, 0, 0, time.Since(startTime))
	c.logAuditTrail(req, nil, lastErr)

	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries+1, lastErr)
}

// makeHTTPRequest performs the actual HTTP request
func (c *OpenAIClient) makeHTTPRequest(ctx context.Context, req *APIRequest, requestID string, attempt int) (*APIResponse, error) {
	startTime := time.Now()

	// Build complete URL
	fullURL := strings.TrimSuffix(c.config.BaseURL, "/") + "/" + strings.TrimPrefix(req.Endpoint, "/")

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, fullURL, bytes.NewBuffer(req.Body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	c.setRequestHeaders(httpReq, req, requestID)

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	responseTime := time.Since(startTime)

	// Parse response headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Create API response
	apiResp := &APIResponse{
		StatusCode:   resp.StatusCode,
		Headers:      headers,
		Body:         body,
		ResponseTime: responseTime,
		RequestSize:  int64(len(req.Body)),
		ResponseSize: int64(len(body)),
		Retries:      attempt,
		CircuitState: c.circuitBreaker.getState().String(),
	}

	// Handle HTTP errors
	if resp.StatusCode >= 400 {
		return apiResp, c.handleHTTPError(resp.StatusCode, body)
	}

	// Parse tokens and cost for successful responses
	if resp.StatusCode == 200 {
		c.parseTokensAndCost(apiResp, body)
	}

	return apiResp, nil
}

// setRequestHeaders sets all required headers for OpenAI requests
func (c *OpenAIClient) setRequestHeaders(httpReq *http.Request, req *APIRequest, requestID string) {
	// Standard headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	// Optional organization header
	if c.config.Organization != "" {
		httpReq.Header.Set("OpenAI-Organization", c.config.Organization)
	}

	// User agent
	userAgent := c.config.UserAgent
	if userAgent == "" {
		userAgent = "KubeChat/1.0"
	}
	httpReq.Header.Set("User-Agent", userAgent)

	// Request ID for tracing
	httpReq.Header.Set("X-Request-ID", requestID)

	// Custom headers from config
	for key, value := range c.config.CustomHeaders {
		httpReq.Header.Set(key, value)
	}

	// Headers from request
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}
}

// parseTokensAndCost extracts token usage and calculates cost from response
func (c *OpenAIClient) parseTokensAndCost(apiResp *APIResponse, body []byte) {
	var chatResp OpenAIChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return
	}

	if chatResp.Usage.TotalTokens > 0 {
		apiResp.TokensUsed = chatResp.Usage.TotalTokens

		// Calculate cost based on model and token usage
		apiResp.Cost = c.calculateCost(chatResp.Model,
			chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens)
	}
}

// calculateCost calculates the cost based on OpenAI pricing
func (c *OpenAIClient) calculateCost(model string, promptTokens, completionTokens int) float64 {
	// OpenAI pricing per 1K tokens (as of 2024)
	modelPricing := map[string][2]float64{
		"gpt-4":             {0.03, 0.06}, // input, output per 1K tokens
		"gpt-4-turbo":       {0.01, 0.03},
		"gpt-4o":            {0.005, 0.015},
		"gpt-3.5-turbo":     {0.0015, 0.002},
		"gpt-3.5-turbo-16k": {0.003, 0.004},
	}

	pricing, exists := modelPricing[model]
	if !exists {
		// Default to GPT-3.5 pricing for unknown models
		pricing = modelPricing["gpt-3.5-turbo"]
	}

	inputCost := (float64(promptTokens) / 1000.0) * pricing[0]
	outputCost := (float64(completionTokens) / 1000.0) * pricing[1]

	return inputCost + outputCost
}

// handleHTTPError creates appropriate error messages for HTTP status codes
func (c *OpenAIClient) handleHTTPError(statusCode int, body []byte) error {
	var openaiErr OpenAIError
	if json.Unmarshal(body, &struct {
		Error *OpenAIError `json:"error"`
	}{Error: &openaiErr}) == nil && openaiErr.Message != "" {
		return &APIError{
			StatusCode: statusCode,
			Message:    openaiErr.Message,
			Type:       openaiErr.Type,
			Retryable:  c.isStatusRetryable(statusCode),
		}
	}

	return &APIError{
		StatusCode: statusCode,
		Message:    fmt.Sprintf("HTTP %d: %s", statusCode, string(body)),
		Retryable:  c.isStatusRetryable(statusCode),
	}
}

// isRetryableError determines if an error warrants a retry
func (c *OpenAIClient) isRetryableError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Retryable
	}

	// Network errors are generally retryable
	return true
}

// isStatusRetryable determines if an HTTP status code is retryable
func (c *OpenAIClient) isStatusRetryable(statusCode int) bool {
	switch statusCode {
	case 429: // Rate limit
		return true
	case 500, 502, 503, 504: // Server errors
		return true
	case 408: // Request timeout
		return true
	default:
		return false
	}
}

// HealthCheck validates OpenAI API connectivity
func (c *OpenAIClient) HealthCheck(ctx context.Context) error {
	// Create a minimal test request
	testReq := &OpenAIChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []OpenAIChatMessage{
			{Role: "user", Content: "test"},
		},
		MaxTokens: 1,
	}

	body, err := json.Marshal(testReq)
	if err != nil {
		return fmt.Errorf("failed to marshal health check request: %w", err)
	}

	apiReq := &APIRequest{
		Method:    "POST",
		Endpoint:  "chat/completions",
		Body:      body,
		RequestID: "health-check-" + uuid.New().String(),
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.makeHTTPRequest(ctx, apiReq, apiReq.RequestID, 0)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("health check returned HTTP %d", resp.StatusCode)
	}

	return nil
}

// GetProviderInfo returns information about the OpenAI provider
func (c *OpenAIClient) GetProviderInfo() *ProviderInfo {
	return &ProviderInfo{
		Name:            "OpenAI",
		Type:            models.ProviderOpenAI,
		BaseURL:         c.config.BaseURL,
		Version:         "v1",
		RateLimitPerMin: c.config.RateLimitPerMin,
		MaxRetries:      c.config.MaxRetries,
		Timeout:         c.config.Timeout,
		Features: []string{
			"chat-completions",
			"function-calling",
			"streaming",
			"token-counting",
			"cost-tracking",
		},
		Status:          c.circuitBreaker.getState().String(),
		LastHealthCheck: time.Now(),
	}
}

// Close gracefully shuts down the OpenAI client
func (c *OpenAIClient) Close() error {
	c.httpClient.CloseIdleConnections()
	log.Println("OpenAI client closed")
	return nil
}

// updateMetrics updates internal metrics
func (c *OpenAIClient) updateMetrics(success bool, tokens int, requestSize, responseSize int64, duration time.Duration) {
	c.metrics.mu.Lock()
	defer c.metrics.mu.Unlock()

	c.metrics.RequestCount++
	c.metrics.LastRequestTime = time.Now()

	if success {
		c.metrics.SuccessfulRequests++
		c.metrics.TotalTokensUsed += int64(tokens)

		// Update average latency
		totalRequests := float64(c.metrics.SuccessfulRequests)
		c.metrics.AverageLatency = (c.metrics.AverageLatency*(totalRequests-1) +
			float64(duration.Milliseconds())) / totalRequests
	} else {
		c.metrics.FailedRequests++
		if c.circuitBreaker.getState() == CircuitOpen {
			c.metrics.CircuitBreakerTrips++
		}
	}
}

// logAuditTrail logs API requests for audit purposes
func (c *OpenAIClient) logAuditTrail(req *APIRequest, resp *APIResponse, err error) {
	if c.auditSvc == nil || !c.config.EnableAuditLog {
		return
	}

	operation := "openai_api_request"
	if strings.Contains(req.Endpoint, "chat/completions") {
		operation = "openai_chat_completion"
	}

	details := map[string]interface{}{
		"endpoint":   req.Endpoint,
		"method":     req.Method,
		"request_id": req.RequestID,
		"user_id":    req.UserID,
		"session_id": req.SessionID,
	}

	if resp != nil {
		details["status_code"] = resp.StatusCode
		details["response_time_ms"] = resp.ResponseTime.Milliseconds()
		details["tokens_used"] = resp.TokensUsed
		details["cost"] = resp.Cost
		details["retries"] = resp.Retries
	}

	if err != nil {
		details["error"] = err.Error()
	}

	safetyLevel := models.SafetyLevelSafe
	if err != nil {
		safetyLevel = models.SafetyLevelWarning
	}

	uid := uuid.MustParse(req.UserID)
	auditReq := models.AuditLogRequest{
		UserID:           &uid,
		QueryText:        fmt.Sprintf("External API operation: %s", operation),
		GeneratedCommand: fmt.Sprintf("Execute %s request to external API", operation),
		SafetyLevel:      string(safetyLevel),
		ExecutionResult:  details,
		ExecutionStatus: func() string {
			if err == nil {
				return models.ExecutionStatusSuccess
			}
			return models.ExecutionStatusFailed
		}(),
	}

	if logErr := c.auditSvc.LogUserAction(context.Background(), auditReq); logErr != nil {
		log.Printf("Failed to log audit entry: %v", logErr)
	}
}

// APIError represents an API-specific error
type APIError struct {
	StatusCode int
	Message    string
	Type       string
	Retryable  bool
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// Circuit breaker implementation

func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			cb.halfOpenCalls = 0
			return true
		}
		return false
	case CircuitHalfOpen:
		return cb.halfOpenCalls < cb.maxHalfOpen
	default:
		return false
	}
}

func (cb *CircuitBreaker) onSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitHalfOpen:
		cb.halfOpenCalls++
		if cb.halfOpenCalls >= cb.maxHalfOpen {
			cb.state = CircuitClosed
			cb.failures = 0
		}
	case CircuitClosed:
		cb.failures = 0
	}
}

func (cb *CircuitBreaker) onFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.state = CircuitOpen
	}
}

func (cb *CircuitBreaker) getState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}
