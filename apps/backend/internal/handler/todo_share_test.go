package handler_test

import (
	"net/http"
	"testing"

	"github.com/faizan1250/taskify/internal/errs"
	"github.com/faizan1250/taskify/internal/handler"
	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/faizan1250/taskify/internal/server"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodoHandler_ShareTodo(t *testing.T) {
	id := uuid.New()

	t.Run("success returns 201 with token", func(t *testing.T) {
		fake := &FakeTodoService{
			ShareTodoFn: func(ctx echo.Context, userID string, todoID uuid.UUID) (*todo.ShareTodoResult, error) {
				assert.Equal(t, testUserID, userID)
				assert.Equal(t, id, todoID)
				return &todo.ShareTodoResult{Token: "tok", TodoID: todoID}, nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodPost, "/todos/"+id.String()+"/share", nil, "", map[string]string{"id": id.String()}, testUserID)

		require.NoError(t, h.ShareTodo(c))
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "tok")
	})

	t.Run("service error passes through", func(t *testing.T) {
		fake := &FakeTodoService{
			ShareTodoFn: func(ctx echo.Context, userID string, todoID uuid.UUID) (*todo.ShareTodoResult, error) {
				return nil, errs.NewNotFoundError("todo not found", false, nil)
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodPost, "/todos/"+id.String()+"/share", nil, "", map[string]string{"id": id.String()}, testUserID)

		requireHTTPError(t, h.ShareTodo(c), http.StatusNotFound)
	})
}

func TestTodoHandler_GetSharedTodo(t *testing.T) {
	token, err := todo.GenerateShareToken()
	require.NoError(t, err)

	t.Run("success returns 200 without auth user", func(t *testing.T) {
		fake := &FakeTodoService{
			GetSharedTodoFn: func(ctx echo.Context, got string) (*todo.PopulatedTodo, error) {
				assert.Equal(t, token, got)
				return &todo.PopulatedTodo{Todo: todo.Todo{Title: "shared"}}, nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		// No userID wired: public route bypasses RequireAuth.
		c, rec := newHandlerContext(http.MethodGet, "/shared/"+token, nil, "", map[string]string{"token": token}, "")

		require.NoError(t, h.GetSharedTodo(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "shared")
	})

	t.Run("unknown token returns 404", func(t *testing.T) {
		unknown, err := todo.GenerateShareToken()
		require.NoError(t, err)
		fake := &FakeTodoService{
			GetSharedTodoFn: func(ctx echo.Context, got string) (*todo.PopulatedTodo, error) {
				code := "SHARE_NOT_FOUND"
				return nil, errs.NewNotFoundError("shared todo not found", false, &code)
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodGet, "/shared/"+unknown, nil, "", map[string]string{"token": unknown}, "")

		requireHTTPError(t, h.GetSharedTodo(c), http.StatusNotFound)
	})

	t.Run("short token fails validation before service", func(t *testing.T) {
		called := false
		fake := &FakeTodoService{
			GetSharedTodoFn: func(ctx echo.Context, got string) (*todo.PopulatedTodo, error) {
				called = true
				return nil, errFakeService
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodGet, "/shared/x", nil, "", map[string]string{"token": "x"}, "")

		requireHTTPError(t, h.GetSharedTodo(c), http.StatusBadRequest)
		assert.False(t, called, "service must not be called on validation failure")
	})
}

func TestTodoHandler_UnshareTodo(t *testing.T) {
	id := uuid.New()

	t.Run("success returns 204", func(t *testing.T) {
		fake := &FakeTodoService{
			UnshareTodoFn: func(ctx echo.Context, userID string, todoID uuid.UUID) error {
				assert.Equal(t, testUserID, userID)
				assert.Equal(t, id, todoID)
				return nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodDelete, "/todos/"+id.String()+"/share", nil, "", map[string]string{"id": id.String()}, testUserID)

		require.NoError(t, h.UnshareTodo(c))
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("no active share returns 404", func(t *testing.T) {
		fake := &FakeTodoService{
			UnshareTodoFn: func(ctx echo.Context, userID string, todoID uuid.UUID) error {
				code := "SHARE_NOT_FOUND"
				return errs.NewNotFoundError("no active share found for todo", false, &code)
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodDelete, "/todos/"+id.String()+"/share", nil, "", map[string]string{"id": id.String()}, testUserID)

		requireHTTPError(t, h.UnshareTodo(c), http.StatusNotFound)
	})
}
