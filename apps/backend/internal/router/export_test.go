package router

import (
	"github.com/faizan1250/taskify/internal/handler"
	"github.com/labstack/echo/v4"
)

func ExportRegisterSystemRoutes(r *echo.Echo, h *handler.Handlers) {
	registerSystemRoutes(r, h)
}
