package prompts

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/r3labs/sse/v2"

	"github.com/pramodksahoo/kubechat/backend/handlers/helpers"
)

func PlanStreamHandler(server *sse.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		if server == nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "streaming unavailable"})
		}
		planID := strings.TrimSpace(c.Param("id"))
		if planID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing plan id"})
		}

		streamID := planStreamID(planID)
		server.CreateStream(streamID)
		helpers.ServeStream(c, server, streamID)

		return nil
	}
}
