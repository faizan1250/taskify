package router

import (
	"github.com/faizan1250/taskify/internal/handler"
	"github.com/faizan1250/taskify/static"

	"github.com/labstack/echo/v4"
)

func registerSystemRoutes(r *echo.Echo, h *handler.Handlers) {
	r.GET("/status", h.Health.CheckHealth)

	r.StaticFS("/static", static.Files)

	r.GET("/docs", h.OpenAPI.ServeOpenAPIUI)
}
