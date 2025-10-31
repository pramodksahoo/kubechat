package helpers

import (
	"github.com/labstack/echo/v4"
	"github.com/r3labs/sse/v2"
)

// ServeStream proxies the request to the SSE server with the provided stream id.
func ServeStream(ctx echo.Context, server *sse.Server, streamID string) {
	req := ctx.Request()

	cloned := req.Clone(req.Context())
	query := cloned.URL.Query()
	query.Set("stream", streamID)
	cloned.URL.RawQuery = query.Encode()
	cloned.RequestURI = cloned.URL.RequestURI()

	server.ServeHTTP(ctx.Response().Writer, cloned)
}
