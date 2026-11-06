package handler

import (
	"net/http"

	"github.com/faizan1250/taskify/internal/middleware"
	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/labstack/echo/v4"
)

// ShareTodo mints (or returns the existing) bearer token for a caller-owned
// todo. Idempotent: repeated calls return the same active token.
func (h *TodoHandler) ShareTodo(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *todo.ShareTodoPayload) (*todo.ShareTodoResult, error) {
			userID := middleware.GetUserID(c)
			return h.todoService.ShareTodo(c, userID, payload.ID)
		},
		http.StatusCreated,
		&todo.ShareTodoPayload{},
	)(c)
}

// GetSharedTodo is a PUBLIC read: no auth middleware. Token possession is the
// authorization; invalid or revoked tokens read as 404.
func (h *TodoHandler) GetSharedTodo(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *todo.GetSharedTodoPayload) (*todo.PopulatedTodo, error) {
			return h.todoService.GetSharedTodo(c, payload.Token)
		},
		http.StatusOK,
		&todo.GetSharedTodoPayload{},
	)(c)
}

// UnshareTodo revokes every active token for a caller-owned todo.
func (h *TodoHandler) UnshareTodo(c echo.Context) error {
	return HandleNoContent(
		h.Handler,
		func(c echo.Context, payload *todo.UnshareTodoPayload) error {
			userID := middleware.GetUserID(c)
			return h.todoService.UnshareTodo(c, userID, payload.ID)
		},
		http.StatusNoContent,
		&todo.UnshareTodoPayload{},
	)(c)
}
