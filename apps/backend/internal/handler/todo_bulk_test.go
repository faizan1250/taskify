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

func bulkUpdateBody(ids []uuid.UUID, patch string) string {
	strs := make([]string, 0, len(ids))
	for _, id := range ids {
		strs = append(strs, `"`+id.String()+`"`)
	}
	return fmt.Sprintf(`{"ids":[%s],"patch":{%s}}`, strings.Join(strs, ","), patch)
}

func bulkDeleteBody(ids []uuid.UUID) string {
	strs := make([]string, 0, len(ids))
	for _, id := range ids {
		strs = append(strs, `"`+id.String()+`"`)
	}
	return fmt.Sprintf(`{"ids":[%s]}`, strings.Join(strs, ","))
}

func TestTodoHandler_BulkUpdateTodos(t *testing.T) {
	idA, idB := uuid.New(), uuid.New()

	t.Run("full success returns 200", func(t *testing.T) {
		fake := &FakeTodoService{
			BulkUpdateTodosFn: func(ctx echo.Context, userID string, payload *todo.BulkUpdateTodosPayload) (*todo.BulkUpdateTodosResult, error) {
				assert.Equal(t, testUserID, userID)
				assert.Len(t, payload.IDs, 2)
				require.NotNil(t, payload.Patch.Status)
				return &todo.BulkUpdateTodosResult{
					Updated: []todo.Todo{{Title: "a"}, {Title: "b"}},
					Failed:  []todo.BulkTodoFailure{},
				}, nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodPatch, "/todos/bulk", strings.NewReader(bulkUpdateBody([]uuid.UUID{idA, idB}, `"status":"completed"`)), "", nil, testUserID)

		require.NoError(t, h.BulkUpdateTodos(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"updated"`)
		assert.Contains(t, rec.Body.String(), `"failed":[]`)
	})

	t.Run("partial failure returns 207", func(t *testing.T) {
		missing := uuid.New()
		fake := &FakeTodoService{
			BulkUpdateTodosFn: func(ctx echo.Context, userID string, payload *todo.BulkUpdateTodosPayload) (*todo.BulkUpdateTodosResult, error) {
				return &todo.BulkUpdateTodosResult{
					Updated: []todo.Todo{{Title: "a"}},
					Failed:  []todo.BulkTodoFailure{{ID: missing, Error: "todo not found"}},
				}, nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodPatch, "/todos/bulk", strings.NewReader(bulkUpdateBody([]uuid.UUID{idA, missing}, `"status":"active"`)), "", nil, testUserID)

		require.NoError(t, h.BulkUpdateTodos(c))
		assert.Equal(t, http.StatusMultiStatus, rec.Code)
		assert.Contains(t, rec.Body.String(), missing.String())
		assert.Contains(t, rec.Body.String(), "todo not found")
	})

	t.Run("empty ids returns 400 without calling service", func(t *testing.T) {
		called := false
		fake := &FakeTodoService{
			BulkUpdateTodosFn: func(ctx echo.Context, userID string, payload *todo.BulkUpdateTodosPayload) (*todo.BulkUpdateTodosResult, error) {
				called = true
				return nil, errFakeService
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodPatch, "/todos/bulk", strings.NewReader(`{"ids":[],"patch":{"status":"active"}}`), "", nil, testUserID)

		requireHTTPError(t, h.BulkUpdateTodos(c), http.StatusBadRequest)
		assert.False(t, called)
	})

	t.Run("empty patch returns 400", func(t *testing.T) {
		h := handler.NewTodoHandler(&server.Server{}, &FakeTodoService{})
		c, _ := newHandlerContext(http.MethodPatch, "/todos/bulk", strings.NewReader(bulkUpdateBody([]uuid.UUID{idA}, "")), "", nil, testUserID)

		requireHTTPError(t, h.BulkUpdateTodos(c), http.StatusBadRequest)
	})

	t.Run("invalid status returns 400", func(t *testing.T) {
		h := handler.NewTodoHandler(&server.Server{}, &FakeTodoService{})
		c, _ := newHandlerContext(http.MethodPatch, "/todos/bulk", strings.NewReader(bulkUpdateBody([]uuid.UUID{idA}, `"status":"bogus"`)), "", nil, testUserID)

		requireHTTPError(t, h.BulkUpdateTodos(c), http.StatusBadRequest)
	})

	t.Run("service error passes through", func(t *testing.T) {
		fake := &FakeTodoService{
			BulkUpdateTodosFn: func(ctx echo.Context, userID string, payload *todo.BulkUpdateTodosPayload) (*todo.BulkUpdateTodosResult, error) {
				return nil, errs.NewNotFoundError("category not found", false, nil)
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodPatch, "/todos/bulk", strings.NewReader(bulkUpdateBody([]uuid.UUID{idA}, `"status":"active"`)), "", nil, testUserID)

		requireHTTPError(t, h.BulkUpdateTodos(c), http.StatusNotFound)
	})
}

func TestTodoHandler_BulkDeleteTodos(t *testing.T) {
	idA, idB := uuid.New(), uuid.New()

	t.Run("full success returns 200", func(t *testing.T) {
		fake := &FakeTodoService{
			BulkDeleteTodosFn: func(ctx echo.Context, userID string, payload *todo.BulkDeleteTodosPayload) (*todo.BulkDeleteTodosResult, error) {
				assert.Equal(t, testUserID, userID)
				assert.Len(t, payload.IDs, 2)
				return &todo.BulkDeleteTodosResult{Deleted: payload.IDs, Failed: []todo.BulkTodoFailure{}}, nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodDelete, "/todos/bulk", strings.NewReader(bulkDeleteBody([]uuid.UUID{idA, idB})), "", nil, testUserID)

		require.NoError(t, h.BulkDeleteTodos(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"deleted"`)
		assert.Contains(t, rec.Body.String(), idA.String())
	})

	t.Run("partial failure returns 207", func(t *testing.T) {
		missing := uuid.New()
		fake := &FakeTodoService{
			BulkDeleteTodosFn: func(ctx echo.Context, userID string, payload *todo.BulkDeleteTodosPayload) (*todo.BulkDeleteTodosResult, error) {
				return &todo.BulkDeleteTodosResult{
					Deleted: []uuid.UUID{idA},
					Failed:  []todo.BulkTodoFailure{{ID: missing, Error: "todo not found"}},
				}, nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodDelete, "/todos/bulk", strings.NewReader(bulkDeleteBody([]uuid.UUID{idA, missing})), "", nil, testUserID)

		require.NoError(t, h.BulkDeleteTodos(c))
		assert.Equal(t, http.StatusMultiStatus, rec.Code)
		assert.Contains(t, rec.Body.String(), missing.String())
	})

	t.Run("empty ids returns 400 without calling service", func(t *testing.T) {
		called := false
		fake := &FakeTodoService{
			BulkDeleteTodosFn: func(ctx echo.Context, userID string, payload *todo.BulkDeleteTodosPayload) (*todo.BulkDeleteTodosResult, error) {
				called = true
				return nil, errFakeService
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodDelete, "/todos/bulk", strings.NewReader(`{"ids":[]}`), "", nil, testUserID)

		requireHTTPError(t, h.BulkDeleteTodos(c), http.StatusBadRequest)
		assert.False(t, called)
	})

	t.Run("missing body returns 400", func(t *testing.T) {
		h := handler.NewTodoHandler(&server.Server{}, &FakeTodoService{})
		c, _ := newHandlerContext(http.MethodDelete, "/todos/bulk", strings.NewReader(`{}`), "", nil, testUserID)

		requireHTTPError(t, h.BulkDeleteTodos(c), http.StatusBadRequest)
	})

	t.Run("service error passes through", func(t *testing.T) {
		fake := &FakeTodoService{
			BulkDeleteTodosFn: func(ctx echo.Context, userID string, payload *todo.BulkDeleteTodosPayload) (*todo.BulkDeleteTodosResult, error) {
				return nil, errs.NewInternalServerError()
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodDelete, "/todos/bulk", strings.NewReader(bulkDeleteBody([]uuid.UUID{idA})), "", nil, testUserID)

		requireHTTPError(t, h.BulkDeleteTodos(c), http.StatusInternalServerError)
	})
}
