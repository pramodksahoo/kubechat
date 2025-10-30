package prompts

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pramodksahoo/kubechat/backend/internal/api/repository"
	"github.com/pramodksahoo/kubechat/backend/internal/plan"
	"github.com/labstack/echo/v4"
)

type stubFetcher struct {
	record repository.PlanRecord
	err    error
}

func (s *stubFetcher) Get(ctx context.Context, id string) (repository.PlanRecord, error) {
	return s.record, s.err
}

func TestPlanQueryControllerSuccess(t *testing.T) {
	repo := &stubFetcher{record: repository.PlanRecord{
		Plan:     plan.PlanDraft{ID: "plan-123"},
		StoredAt: time.Now().UTC(),
	}}
	controller := NewPlanQueryController(repo, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/plans/plan-123", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("plan-123")

	if err := controller.Handle(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestPlanQueryControllerNotFound(t *testing.T) {
	repo := &stubFetcher{err: repository.ErrPlanNotFound{ID: "missing"}}
	controller := NewPlanQueryController(repo, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/plans/missing", nil)
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

func TestPlanQueryControllerHandlesErrors(t *testing.T) {
	repo := &stubFetcher{err: errors.New("boom")}
	controller := NewPlanQueryController(repo, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/plans/fail", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("fail")

	if err := controller.Handle(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}

func TestPlanQueryControllerValidatesInput(t *testing.T) {
	repo := &stubFetcher{}
	controller := NewPlanQueryController(repo, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/plans/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	if err := controller.Handle(c); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
