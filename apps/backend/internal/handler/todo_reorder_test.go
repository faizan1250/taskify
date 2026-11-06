package handler_test

import (
	"fmt"
	"net/http"
	"strings"
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

func reorderBody(ids []uuid.UUID) string {
	strs := make([]string, 0, len(ids))
	for _, id := range ids {
		strs = append(strs, `"`+id.String()+`"`)
	}
	return fmt.Sprintf(`{"orderedIds":[%s]}`, strings.Join(strs, ","))
}

func TestTodoHandler_ReorderTodos(t *testing.T) {
	idA, idB := uuid.New(), uuid.New()

	t.Run("full success returns 200 in requested order", func(t *testing.T) {
		fake := &FakeTodoService{
			ReorderTodosFn: func(ctx echo.Context, userID string, payload *todo.ReorderTodosPayload) ([]todo.Todo, error) {
				assert.Equal(t, testUserID, userID)
				assert.Equal(t, []uuid.UUID{idB, idA}, payload.OrderedIDs)
				return []todo.Todo{{Title: "b", Position: 0}, {Title: "a", Position: 1}}, nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodPatch, "/todos/reorder", strings.NewReader(reorderBody([]uuid.UUID{idB, idA})), "", nil, testUserID)

		require.NoError(t, h.ReorderTodos(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"position":0`)
		assert.Contains(t, rec.Body.String(), `"position":1`)
	})

	t.Run("empty order returns 400 without calling service", func(t *testing.T) {
		called := false
		fake := &FakeTodoService{
			ReorderTodosFn: func(ctx echo.Context, userID string, payload *todo.ReorderTodosPayload) ([]todo.Todo, error) {
				called = true
				return nil, errFakeService
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodPatch, "/todos/reorder", strings.NewReader(`{"orderedIds":[]}`), "", nil, testUserID)

		requireHTTPError(t, h.ReorderTodos(c), http.StatusBadRequest)
		assert.False(t, called)
	})

	t.Run("duplicate ids return 400", func(t *testing.T) {
		h := handler.NewTodoHandler(&server.Server{}, &FakeTodoService{})
		c, _ := newHandlerContext(http.MethodPatch, "/todos/reorder", strings.NewReader(reorderBody([]uuid.UUID{idA, idA})), "", nil, testUserID)

		requireHTTPError(t, h.ReorderTodos(c), http.StatusBadRequest)
	})

	t.Run("nil UUID returns 400", func(t *testing.T) {
		h := handler.NewTodoHandler(&server.Server{}, &FakeTodoService{})
		c, _ := newHandlerContext(http.MethodPatch, "/todos/reorder", strings.NewReader(`{"orderedIds":["00000000-0000-0000-0000-000000000000"]}`), "", nil, testUserID)

		requireHTTPError(t, h.ReorderTodos(c), http.StatusBadRequest)
	})

	t.Run("unknown id surfaces 404", func(t *testing.T) {
		fake := &FakeTodoService{
			ReorderTodosFn: func(ctx echo.Context, userID string, payload *todo.ReorderTodosPayload) ([]todo.Todo, error) {
				return nil, errs.NewNotFoundError("todo not found", false, nil)
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodPatch, "/todos/reorder", strings.NewReader(reorderBody([]uuid.UUID{idA})), "", nil, testUserID)

		requireHTTPError(t, h.ReorderTodos(c), http.StatusNotFound)
	})
}
