package websocket

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// handleExecute processes command execution requests
func (c *client) handleExecute(message *models.WebSocketMessage) error {
	if c.requireAuth() {
		c.sendError("AUTH_REQUIRED", "Authentication required", "")
		return nil
	}

	var payload models.ExecutePayload
	if err := message.UnmarshalPayload(&payload); err != nil {
		return fmt.Errorf("invalid execute payload: %w", err)
	}

	// Validate command
	if payload.Command == "" && payload.Query == "" {
		c.sendError("INVALID_COMMAND", "Command or query is required", "")
		return nil
	}

	// Check concurrent command limit
	c.commandsMutex.Lock()
	if len(c.activeCommands) >= c.hub.config.MaxConcurrentCommands {
		c.commandsMutex.Unlock()
		c.sendError("TOO_MANY_COMMANDS", "Maximum concurrent commands exceeded", "")
		return nil
	}
	c.commandsMutex.Unlock()

	// Generate command if query is provided
	var actualCommand string
	var safetyLevel string

	if payload.Query != "" && payload.Command == "" {
		// Use NLP service to generate command
		if c.hub.nlpService == nil {
			c.sendError("NLP_SERVICE_UNAVAILABLE", "NLP service not available", "")
			return nil
		}

		nlpRequest := &models.NLPRequest{
			ID:          uuid.New(),
			UserID:      c.client.UserID,
			SessionID:   c.client.SessionID,
			Query:       payload.Query,
			Context:     payload.Context,
			ClusterInfo: payload.ClusterInfo,
			CreatedAt:   time.Now(),
		}

		if payload.Provider != "" {
			nlpRequest.Provider = models.NLPProvider(payload.Provider)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		nlpResponse, err := c.hub.nlpService.ProcessQuery(ctx, nlpRequest)
		if err != nil {
			c.sendError("NLP_PROCESSING_ERROR", "Failed to process query", err.Error())
			return nil
		}

		actualCommand = nlpResponse.GeneratedCommand
		safetyLevel = string(nlpResponse.SafetyLevel)
	} else {
		actualCommand = payload.Command
		safetyLevel = c.classifyCommand(actualCommand)
	}

	// Create command execution
	commandID := uuid.New().String()
	execution := &models.CommandExecution{
		ID:           commandID,
		ClientID:     c.client.ID,
		UserID:       c.client.UserID,
		SessionID:    c.client.SessionID,
		Command:      actualCommand,
		Query:        payload.Query,
		SafetyLevel:  safetyLevel,
		Status:       "queued",
		StartedAt:    time.Now(),
		StreamOutput: payload.StreamOutput,
	}

	// Add to active commands
	c.commandsMutex.Lock()
	c.activeCommands[commandID] = execution
	c.commandsMutex.Unlock()

	// Update metrics
	c.hub.metrics.ActiveCommands++

	// Send execution started notification
	c.sendExecutingNotification(execution)

	// Execute command asynchronously
	go c.executeCommand(execution)

	return nil
}

// executeCommand executes a kubectl command
func (c *client) executeCommand(execution *models.CommandExecution) {
	defer func() {
		// Remove from active commands
		c.commandsMutex.Lock()
		delete(c.activeCommands, execution.ID)
		c.commandsMutex.Unlock()

		// Update metrics
		c.hub.metrics.ActiveCommands--
		if execution.Status == "completed" {
			c.hub.metrics.CompletedCommands++
		} else {
			c.hub.metrics.FailedCommands++
		}
	}()

	execution.Status = "running"

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.ctx, c.hub.config.CommandTimeout)
	defer cancel()

	// Parse command
	parts := strings.Fields(execution.Command)
	if len(parts) == 0 {
		c.sendCommandError(execution.ID, "EMPTY_COMMAND", "Command is empty")
		execution.Status = "failed"
		execution.Error = "Command is empty"
		now := time.Now()
		execution.CompletedAt = &now
		return
	}

	// Security check - only allow kubectl commands
	if parts[0] != "kubectl" {
		c.sendCommandError(execution.ID, "UNSAFE_COMMAND", "Only kubectl commands are allowed")
		execution.Status = "failed"
		execution.Error = "Only kubectl commands are allowed"
		now := time.Now()
		execution.CompletedAt = &now
		return
	}

	// Send progress update
	c.sendProgress(execution.ID, 10.0, "preparing", "Preparing command execution")

	// Execute command
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)

	if execution.StreamOutput {
		c.executeStreamingCommand(ctx, cmd, execution)
	} else {
		c.executeBufferedCommand(ctx, cmd, execution)
	}

	// Log audit event
	if c.hub.auditService != nil {
		go func() {
			auditCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			executionResult := map[string]interface{}{
				"command_id":    execution.ID,
				"command":       execution.Command,
				"safety_level":  execution.SafetyLevel,
				"duration":      execution.Duration,
				"stream_output": execution.StreamOutput,
			}

			if execution.ExitCode != nil {
				executionResult["exit_code"] = *execution.ExitCode
			}

			status := models.ExecutionStatusSuccess
			if execution.Status == "failed" {
				status = models.ExecutionStatusFailed
			}

			c.hub.auditService.LogKubectlExecution(
				auditCtx,
				&execution.UserID,
				&execution.SessionID,
				execution.Query,
				execution.Command,
				executionResult,
				status,
				nil,
			)
		}()
	}
}

// executeStreamingCommand executes command with streaming output
func (c *client) executeStreamingCommand(ctx context.Context, cmd *exec.Cmd, execution *models.CommandExecution) {
	// Set up pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.sendCommandError(execution.ID, "PIPE_ERROR", err.Error())
		execution.Status = "failed"
		execution.Error = err.Error()
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		c.sendCommandError(execution.ID, "PIPE_ERROR", err.Error())
		execution.Status = "failed"
		execution.Error = err.Error()
		return
	}

	// Start command
	if err := cmd.Start(); err != nil {
		c.sendCommandError(execution.ID, "EXECUTION_ERROR", err.Error())
		execution.Status = "failed"
		execution.Error = err.Error()
		return
	}

	c.sendProgress(execution.ID, 30.0, "executing", "Command started")

	// Stream output
	outputBuilder := strings.Builder{}
	errorBuilder := strings.Builder{}

	// Handle stdout
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stdout.Read(buffer)
			if n > 0 {
				content := string(buffer[:n])
				outputBuilder.WriteString(content)
				c.sendOutput(execution.ID, content, "stdout")
			}
			if err != nil {
				break
			}
		}
	}()

	// Handle stderr
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stderr.Read(buffer)
			if n > 0 {
				content := string(buffer[:n])
				errorBuilder.WriteString(content)
				c.sendOutput(execution.ID, content, "stderr")
			}
			if err != nil {
				break
			}
		}
	}()

	// Wait for command to complete
	err = cmd.Wait()
	now := time.Now()
	execution.CompletedAt = &now
	execution.Duration = execution.CompletedAt.Sub(execution.StartedAt).String()

	if exitError, ok := err.(*exec.ExitError); ok {
		exitCode := exitError.ExitCode()
		execution.ExitCode = &exitCode
	} else if err == nil {
		exitCode := 0
		execution.ExitCode = &exitCode
	}

	execution.Output = outputBuilder.String()
	execution.Error = errorBuilder.String()

	if err != nil && !strings.Contains(err.Error(), "exit status") {
		execution.Status = "failed"
		execution.Error = err.Error()
	} else {
		execution.Status = "completed"
	}

	c.sendProgress(execution.ID, 100.0, "completed", "Command execution finished")
	c.sendResult(execution)
}

// executeBufferedCommand executes command with buffered output
func (c *client) executeBufferedCommand(ctx context.Context, cmd *exec.Cmd, execution *models.CommandExecution) {
	c.sendProgress(execution.ID, 30.0, "executing", "Command started")

	output, err := cmd.CombinedOutput()

	now := time.Now()
	execution.CompletedAt = &now
	execution.Duration = execution.CompletedAt.Sub(execution.StartedAt).String()
	execution.Output = string(output)

	if exitError, ok := err.(*exec.ExitError); ok {
		exitCode := exitError.ExitCode()
		execution.ExitCode = &exitCode
		execution.Status = "failed"
		execution.Error = err.Error()
	} else if err != nil {
		execution.Status = "failed"
		execution.Error = err.Error()
	} else {
		exitCode := 0
		execution.ExitCode = &exitCode
		execution.Status = "completed"
	}

	c.sendProgress(execution.ID, 100.0, "completed", "Command execution finished")
	c.sendResult(execution)
}

// handleCancel processes command cancellation requests
func (c *client) handleCancel(message *models.WebSocketMessage) error {
	if c.requireAuth() {
		c.sendError("AUTH_REQUIRED", "Authentication required", "")
		return nil
	}

	var payload models.CancelPayload
	if err := message.UnmarshalPayload(&payload); err != nil {
		return fmt.Errorf("invalid cancel payload: %w", err)
	}

	c.commandsMutex.Lock()
	execution, exists := c.activeCommands[payload.CommandID]
	if exists {
		execution.Status = "cancelled"
		now := time.Now()
		execution.CompletedAt = &now
		execution.Duration = execution.CompletedAt.Sub(execution.StartedAt).String()
	}
	c.commandsMutex.Unlock()

	if !exists {
		c.sendError("COMMAND_NOT_FOUND", "Command not found", payload.CommandID)
		return nil
	}

	c.logger.Info("Command cancelled",
		"client_id", c.client.ID,
		"command_id", payload.CommandID,
		"reason", payload.Reason)

	// Send cancellation notification
	statusPayload := models.StatusPayload{
		Type:    "command",
		Status:  "cancelled",
		Message: fmt.Sprintf("Command %s cancelled", payload.CommandID),
		Metadata: map[string]interface{}{
			"command_id": payload.CommandID,
			"reason":     payload.Reason,
		},
		Timestamp: time.Now(),
	}

	statusMessage, _ := models.NewWebSocketMessage(models.WSMsgTypeStatus, statusPayload)
	select {
	case c.send <- statusMessage:
	default:
		c.logger.Warn("Failed to send cancellation notification", "client_id", c.client.ID)
	}

	return nil
}

// classifyCommand classifies the safety level of a command
func (c *client) classifyCommand(command string) string {
	if command == "" {
		return models.SafetyLevelDangerous
	}

	lowerCmd := strings.ToLower(command)

	// Dangerous operations
	dangerousPatterns := []string{
		"delete", "destroy", "rm", "--force", "--cascade=foreground",
		"drain", "cordon", "evict", "--grace-period=0",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerCmd, pattern) {
			return models.SafetyLevelDangerous
		}
	}

	// Warning-level operations
	warningPatterns := []string{
		"create", "apply", "patch", "replace", "scale", "restart",
		"edit", "label", "annotate", "expose", "rollout",
	}

	for _, pattern := range warningPatterns {
		if strings.Contains(lowerCmd, pattern) {
			return models.SafetyLevelWarning
		}
	}

	// Safe operations (read-only)
	return models.SafetyLevelSafe
}

// Helper methods for sending different types of messages

func (c *client) sendExecutingNotification(execution *models.CommandExecution) {
	payload := models.ExecutingPayload{
		CommandID:     execution.ID,
		Command:       execution.Command,
		SafetyLevel:   execution.SafetyLevel,
		EstimatedTime: "unknown",
		StartedAt:     execution.StartedAt,
	}

	message, _ := models.NewWebSocketMessage(models.WSMsgTypeExecuting, payload)
	select {
	case c.send <- message:
	default:
		c.logger.Warn("Failed to send executing notification", "client_id", c.client.ID)
	}
}

func (c *client) sendProgress(commandID string, percentage float64, stage, description string) {
	payload := models.ProgressPayload{
		CommandID:   commandID,
		Percentage:  percentage,
		Stage:       stage,
		Description: description,
		Timestamp:   time.Now(),
	}

	message, _ := models.NewWebSocketMessage(models.WSMsgTypeProgress, payload)
	select {
	case c.send <- message:
	default:
		c.logger.Warn("Failed to send progress update", "client_id", c.client.ID)
	}
}

func (c *client) sendOutput(commandID, content, stream string) {
	payload := models.OutputPayload{
		CommandID: commandID,
		Content:   content,
		Stream:    stream,
		Timestamp: time.Now(),
	}

	message, _ := models.NewWebSocketMessage(models.WSMsgTypeOutput, payload)
	select {
	case c.send <- message:
	default:
		c.logger.Warn("Failed to send output", "client_id", c.client.ID)
	}
}

func (c *client) sendResult(execution *models.CommandExecution) {
	executionResult := map[string]interface{}{
		"safety_level":  execution.SafetyLevel,
		"stream_output": execution.StreamOutput,
	}

	payload := models.ResultPayload{
		CommandID:       execution.ID,
		Success:         execution.Status == "completed",
		ExitCode:        *execution.ExitCode,
		Output:          execution.Output,
		Error:           execution.Error,
		Duration:        execution.Duration,
		ExecutionResult: executionResult,
		Timestamp:       *execution.CompletedAt,
	}

	message, _ := models.NewWebSocketMessage(models.WSMsgTypeResult, payload)
	select {
	case c.send <- message:
	default:
		c.logger.Warn("Failed to send result", "client_id", c.client.ID)
	}
}

func (c *client) sendCommandError(commandID, code, message string) {
	payload := models.ErrorPayload{
		CommandID: commandID,
		Code:      code,
		Message:   message,
	}

	errorMessage, _ := models.NewWebSocketMessage(models.WSMsgTypeError, payload)
	select {
	case c.send <- errorMessage:
		c.hub.metrics.ErrorCount++
	default:
		c.logger.Warn("Failed to send command error", "client_id", c.client.ID)
	}
}
