package prompts

import (
	"context"
	"errors"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/pramodksahoo/kubechat/backend/internal/api/repository"
	"github.com/labstack/echo/v4"
)

type PlanFetcher interface {
	Get(ctx context.Context, id string) (repository.PlanRecord, error)
}

type PlanQueryController struct {
	repo   PlanFetcher
	logger *log.Logger
}

func NewPlanQueryController(repo PlanFetcher, logger *log.Logger) *PlanQueryController {
	if logger == nil {
		logger = log.Default()
	}
	return &PlanQueryController{repo: repo, logger: logger}
}

func (c *PlanQueryController) Handle(ctx echo.Context) error {
	planID := ctx.Param("id")
	if planID == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "missing plan id"})
	}

	record, err := c.repo.Get(ctx.Request().Context(), planID)
	if err != nil {
		var notFound repository.ErrPlanNotFound
		if errors.As(err, &notFound) {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "plan not found"})
		}
		c.logger.Error("failed to load plan", "plan_id", planID, "error", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load plan"})
	}

	return ctx.JSON(http.StatusOK, record)
}
