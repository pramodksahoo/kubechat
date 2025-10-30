package prompts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/pramodksahoo/kubechat/backend/internal/api/repository"
	"github.com/pramodksahoo/kubechat/backend/internal/plan"
	"github.com/pramodksahoo/kubechat/backend/internal/telemetry"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

type fakeBuilder struct {
	plan    plan.PlanDraft
	err     error
	invoked bool
}

func (f *fakeBuilder) BuildPlan(ctx context.Context, input plan.BuildInput) (plan.PlanDraft, error) {
	f.invoked = true
	return f.plan, f.err
}

type fakeRepo struct {
	saved plan.PlanDraft
	err   error
}

func (f *fakeRepo) Save(ctx context.Context, draft plan.PlanDraft) (repository.PlanRecord, error) {
	f.saved = draft
	if f.err != nil {
		return repository.PlanRecord{}, f.err
	}
	stored := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	return repository.PlanRecord{Plan: draft, StoredAt: stored, ExpiresAt: stored.Add(24 * time.Hour)}, nil
}

func TestPromptControllerSuccess(t *testing.T) {
	builder := &fakeBuilder{plan: plan.PlanDraft{ID: "plan-123", TargetCluster: "prod", TargetNamespace: "payments"}}
	repo := &fakeRepo{}
	reg := prometheus.NewRegistry()
	metrics := telemetry.NewPlanMetrics(reg)
	logger := log.NewWithOptions(io.Discard, log.Options{})

	controller := NewPromptController(builder, metrics, repo, logger)

	e := echo.New()
	payload := PromptRequest{Prompt: "Inspect prod cluster", Metadata: map[string]string{"source": "test"}}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/prompts", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.Set(echo.HeaderXRequestID, "req-123")

	if err := controller.Handle(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	if !builder.invoked {
		t.Fatal("expected builder to be invoked")
	}
	if repo.saved.ID == "" {
		t.Fatal("expected plan to be persisted")
	}

	var response PromptResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Plan.TargetCluster != "prod" || response.Plan.TargetNamespace != "payments" {
		t.Fatalf("unexpected plan classification: %+v", response.Plan)
	}

	if response.StoredAt.IsZero() || response.StoredAt.Year() != 2025 {
		t.Fatalf("expected stored timestamp to be propagated, got %v", response.StoredAt)
	}
	if response.ExpiresAt.IsZero() || response.ExpiresAt.Sub(response.StoredAt) <= 0 {
		t.Fatalf("expected expiry timestamp to follow stored timestamp, got %v", response.ExpiresAt)
	}

	metricsFamilies, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
	found := false
	for _, mf := range metricsFamilies {
		if mf.GetName() == "plan_generation_duration_seconds" {
			found = true
			if len(mf.Metric) == 0 || mf.Metric[0].GetHistogram().GetSampleCount() != 1 {
				t.Fatalf("expected exactly one histogram sample, got %+v", mf.Metric)
			}
		}
	}
	if !found {
		t.Fatal("expected plan_generation_duration_seconds metric to be registered")
	}
}

func TestPromptControllerValidatesPrompt(t *testing.T) {
	builder := &fakeBuilder{}
	repo := &fakeRepo{}
	metrics := telemetry.NewPlanMetrics(prometheus.NewRegistry())
	logger := log.NewWithOptions(io.Discard, log.Options{})
	controller := NewPromptController(builder, metrics, repo, logger)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/prompts", bytes.NewReader([]byte(`{"prompt":"   "}`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	if err := controller.Handle(c); err != nil {
		t.Fatalf("expected handler to return no error, got %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if builder.invoked {
		t.Fatal("builder should not be invoked when validation fails")
	}
}

func TestPromptControllerHandlesPersistenceError(t *testing.T) {
	builder := &fakeBuilder{plan: plan.PlanDraft{ID: "plan-123", TargetCluster: "prod", TargetNamespace: "payments"}}
	repo := &fakeRepo{err: errors.New("db down")}
	metrics := telemetry.NewPlanMetrics(prometheus.NewRegistry())
	logger := log.NewWithOptions(io.Discard, log.Options{})
	controller := NewPromptController(builder, metrics, repo, logger)

	e := echo.New()
	payload := PromptRequest{Prompt: "Inspect prod cluster"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/prompts", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	if err := controller.Handle(c); err != nil {
		t.Fatalf("expected handler to return no error, got %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}
