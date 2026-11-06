package v1

import (
	"github.com/faizan1250/taskify/internal/handler"
	"github.com/faizan1250/taskify/internal/middleware"
	"github.com/labstack/echo/v4"
)

func RegisterV1Routes(router *echo.Group, handlers *handler.Handlers, middleware *middleware.Middlewares) {
	// Public shared-todo read: token possession is the authorization, so this
	// route intentionally sits outside RequireAuth.
	router.GET("/shared/:token", handlers.Todo.GetSharedTodo)

	// Register todo routes
	registerTodoRoutes(router, handlers.Todo, handlers.Comment, middleware.Auth)

	// Register category routes
	registerCategoryRoutes(router, handlers.Category, middleware.Auth)

	// Register comment routes
	registerCommentRoutes(router, handlers.Comment, middleware.Auth)
}
