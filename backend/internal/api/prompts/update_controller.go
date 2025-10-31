package prompts

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/pramodksahoo/kubechat/backend/internal/api/repository"
	"github.com/r3labs/sse/v2"
)

type PlanEventBroadcaster interface {
	CreateStream(id string) *sse.Stream
	Publish(id string, event *sse.Event)
}

type PlanUpdater interface {
	Update(ctx context.Context, id string, update repository.PlanUpdate) (repository.PlanRecord, bool, error)
}

type PlanUpdateController struct {
	repo   PlanUpdater
	stream PlanEventBroadcaster
	logger *log.Logger
}

type PlanUpdateRequest struct {
	TargetNamespace  *string           `json:"targetNamespace,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	ReplicaOverrides map[string]int    `json:"replicaOverrides,omitempty"`
	UpdatedBy        string            `json:"updatedBy,omitempty"`
}

func NewPlanUpdateController(repo PlanUpdater, broadcaster PlanEventBroadcaster, logger *log.Logger) *PlanUpdateController {
	if logger == nil {
		logger = log.Default()
	}
	return &PlanUpdateController{
		repo:   repo,
		stream: broadcaster,
		logger: logger,
	}
}

func (c *PlanUpdateController) Handle(ctx echo.Context) error {
	planID := strings.TrimSpace(ctx.Param("id"))
	if planID == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "missing plan id"})
	}

	var req PlanUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
	}

	update := repository.PlanUpdate{
		UpdatedBy: req.UpdatedBy,
	}

	if req.TargetNamespace != nil {
		update.TargetNamespace = req.TargetNamespace
	}
	if req.Labels != nil {
		update.Labels = req.Labels
	}
	if req.ReplicaOverrides != nil {
		update.ReplicaOverrides = req.ReplicaOverrides
	}

	record, changed, err := c.repo.Update(ctx.Request().Context(), planID, update)
	if err != nil {
		var notFound repository.ErrPlanNotFound
		if errors.As(err, &notFound) {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "plan not found"})
		}
		c.logger.Error("failed to update plan", "plan_id", planID, "error", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update plan"})
	}

	if changed && c.stream != nil {
		payload, err := json.Marshal(record)
		if err != nil {
			c.logger.Warn("failed to marshal plan record for SSE", "plan_id", planID, "error", err)
		} else {
			streamID := planStreamID(planID)
			c.stream.CreateStream(streamID)
			c.stream.Publish(streamID, &sse.Event{
				Event: []byte("plan_update"),
				Data:  payload,
			})
		}
	}

	return ctx.JSON(http.StatusOK, record)
}

func planStreamID(planID string) string {
	return "plan:" + planID
}
