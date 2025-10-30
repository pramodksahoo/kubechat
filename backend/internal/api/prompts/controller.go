package prompts

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/pramodksahoo/kubechat/backend/internal/api/repository"
	"github.com/pramodksahoo/kubechat/backend/internal/plan"
	"github.com/pramodksahoo/kubechat/backend/internal/telemetry"
	"github.com/labstack/echo/v4"
)

type PromptRequest struct {
	Prompt        string            `json:"prompt"`
	ClusterHint   string            `json:"clusterHint,omitempty"`
	NamespaceHint string            `json:"namespaceHint,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type ResponseMetrics struct {
	GenerationDurationMs int64     `json:"generationDurationMs"`
	CapturedAt           time.Time `json:"capturedAt"`
}

type PromptResponse struct {
	Plan      plan.PlanDraft  `json:"plan"`
	Metrics   ResponseMetrics `json:"metrics"`
	StoredAt  time.Time       `json:"storedAt,omitempty"`
	ExpiresAt time.Time       `json:"expiresAt,omitempty"`
}

type PromptController struct {
	builder plan.Builder
	metrics telemetry.PlanMetricsRecorder
	store   PlanStore
	logger  *log.Logger
	timeout time.Duration
	clock   func() time.Time
}

type PlanStore interface {
	Save(ctx context.Context, draft plan.PlanDraft) (repository.PlanRecord, error)
}

func NewPromptController(builder plan.Builder, metrics telemetry.PlanMetricsRecorder, store PlanStore, logger *log.Logger) *PromptController {
	if logger == nil {
		logger = log.Default()
	}
	return &PromptController{
		builder: builder,
		metrics: metrics,
		store:   store,
		logger:  logger,
		timeout: 5 * time.Second,
		clock:   time.Now,
	}
}

func (c *PromptController) Handle(ctx echo.Context) error {
	var req PromptRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
	}

	req.Prompt = strings.TrimSpace(req.Prompt)
	if req.Prompt == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "prompt is required"})
	}

	scopeSignals := map[string]string{}
	for k, v := range req.Metadata {
		scopeSignals[k] = v
	}
	if remote := strings.TrimSpace(ctx.RealIP()); remote != "" {
		scopeSignals["remote_addr"] = remote
	}
	if ua := strings.TrimSpace(ctx.Request().UserAgent()); ua != "" {
		scopeSignals["user_agent"] = ua
	}

	requestID := ctx.Response().Header().Get(echo.HeaderXRequestID)
	if requestID == "" {
		if candidate, ok := ctx.Get(echo.HeaderXRequestID).(string); ok {
			requestID = candidate
		}
	}
	if requestID != "" {
		scopeSignals["request_id"] = requestID
	}

	planInput := plan.BuildInput{
		Prompt:        req.Prompt,
		ClusterHint:   req.ClusterHint,
		NamespaceHint: req.NamespaceHint,
		ScopeSignals:  scopeSignals,
	}

	parentCtx := ctx.Request().Context()
	childCtx, cancel := context.WithTimeout(parentCtx, c.timeout)
	defer cancel()

	start := c.clock()
	draft, err := c.builder.BuildPlan(childCtx, planInput)
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to generate plan"
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(childCtx.Err(), context.DeadlineExceeded) {
			status = http.StatusGatewayTimeout
			message = "plan generation exceeded 5s SLA"
		}
		c.logger.Error("failed to generate plan", "error", err, "request_id", requestID)
		return ctx.JSON(status, map[string]string{"error": message})
	}

	duration := c.clock().Sub(start)
	draft.GenerationLatency = duration

	var record repository.PlanRecord
	if c.store != nil {
		saved, err := c.store.Save(parentCtx, draft)
		if err != nil {
			c.logger.Error("failed to persist plan", "error", err, "plan_id", draft.ID, "request_id", requestID)
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to persist plan"})
		}
		record = saved
	}

	c.metrics.ObservePlanGeneration(duration, draft.TargetCluster, draft.TargetNamespace)

	c.logger.Info("plan generated", "request_id", requestID, "plan_id", draft.ID, "cluster", draft.TargetCluster, "namespace", draft.TargetNamespace, "duration_ms", duration.Milliseconds(), "risk_level", draft.RiskSummary.Level)

	resp := PromptResponse{
		Plan: draft,
		Metrics: ResponseMetrics{
			GenerationDurationMs: duration.Milliseconds(),
			CapturedAt:           c.clock(),
		},
	}

	if !record.StoredAt.IsZero() {
		resp.StoredAt = record.StoredAt
	}
	if !record.ExpiresAt.IsZero() {
		resp.ExpiresAt = record.ExpiresAt
	}

	return ctx.JSON(http.StatusCreated, resp)
}
