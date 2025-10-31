package prompts

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/pramodksahoo/kubechat/backend/internal/api/repository"
	"github.com/pramodksahoo/kubechat/backend/internal/plan"
	"github.com/r3labs/sse/v2"
)

type stubPlanUpdater struct {
	record  repository.PlanRecord
	changed bool
	err     error
}

func (s *stubPlanUpdater) Update(ctx context.Context, id string, update repository.PlanUpdate) (repository.PlanRecord, bool, error) {
	return s.record, s.changed, s.err
}

type stubBroadcaster struct {
	streams   []string
	published []*sse.Event
}

func (s *stubBroadcaster) CreateStream(id string) *sse.Stream {
	s.streams = append(s.streams, id)
	return nil
}

func (s *stubBroadcaster) Publish(id string, event *sse.Event) {
	s.published = append(s.published, event)
}

func TestPlanUpdateControllerSuccess(t *testing.T) {
	record := repository.PlanRecord{
		Plan: plan.PlanDraft{
			ID:              "plan-1",
			TargetCluster:   "prod",
			TargetNamespace: "payments",
			Parameters: plan.Parameters{
				Namespace: "payments",
			},
		},
		StoredAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	updater := &stubPlanUpdater{record: record, changed: true}
	broadcaster := &stubBroadcaster{}
	controller := NewPlanUpdateController(updater, broadcaster, log.NewWithOptions(io.Discard, log.Options{}))

	e := echo.New()
	payload := PlanUpdateRequest{
		TargetNamespace:  ptr("payments"),
		ReplicaOverrides: map[string]int{"deployments/foo": 4},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/plans/plan-1", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("plan-1")

	if err := controller.Handle(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if len(broadcaster.published) != 1 {
		t.Fatalf("expected SSE event published, got %d", len(broadcaster.published))
	}
	if string(broadcaster.published[0].Event) != "plan_update" {
		t.Fatalf("expected plan_update event, got %s", string(broadcaster.published[0].Event))
	}
}

func TestPlanUpdateControllerHandlesNotFound(t *testing.T) {
	updater := &stubPlanUpdater{err: repository.ErrPlanNotFound{ID: "missing"}}
	controller := NewPlanUpdateController(updater, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/plans/missing", bytes.NewReader([]byte(`{}`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("missing")

	if err := controller.Handle(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestPlanUpdateControllerValidatesInput(t *testing.T) {
	controller := NewPlanUpdateController(&stubPlanUpdater{}, nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/plans/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	if err := controller.Handle(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestPlanUpdateControllerNoChangeSkipsSSE(t *testing.T) {
	record := repository.PlanRecord{
		Plan: plan.PlanDraft{ID: "plan-1"},
	}
	updater := &stubPlanUpdater{record: record, changed: false}
	broadcaster := &stubBroadcaster{}
	controller := NewPlanUpdateController(updater, broadcaster, nil)

	e := echo.New()
	body := []byte(`{}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/plans/plan-1", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("plan-1")

	if err := controller.Handle(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if len(broadcaster.published) != 0 {
		t.Fatalf("expected no SSE events, got %d", len(broadcaster.published))
	}
}

func TestPlanStreamHandlerValidatesServerAndParams(t *testing.T) {
	handler := PlanStreamHandler(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/plans/plan-1/stream", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("")

	if err := handler(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 when server missing, got %d", rec.Code)
	}

	server := sse.New()
	handler = PlanStreamHandler(server)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("")

	if err := handler(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing plan id, got %d", rec.Code)
	}
}

func ptr[T any](v T) *T {
	return &v
}
