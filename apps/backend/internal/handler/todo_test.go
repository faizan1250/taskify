package handler_test

import (
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/faizan1250/taskify/internal/errs"
	"github.com/faizan1250/taskify/internal/handler"
	"github.com/faizan1250/taskify/internal/model"
	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/faizan1250/taskify/internal/server"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodoHandler_CreateTodo(t *testing.T) {
	t.Run("success returns 201", func(t *testing.T) {
		fake := &FakeTodoService{
			CreateTodoFn: func(ctx echo.Context, userID string, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
				assert.Equal(t, testUserID, userID)
				assert.Equal(t, "Buy milk", payload.Title)
				return &todo.Todo{Title: payload.Title, UserID: userID}, nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodPost, "/todos", strings.NewReader(`{"title":"Buy milk"}`), "", nil, testUserID)

		require.NoError(t, h.CreateTodo(c))
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "Buy milk")
	})

	t.Run("validation error returns 400", func(t *testing.T) {
		called := false
		fake := &FakeTodoService{
			CreateTodoFn: func(ctx echo.Context, userID string, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
				called = true
				return nil, errFakeService
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodPost, "/todos", strings.NewReader(`{}`), "", nil, testUserID)

		requireHTTPError(t, h.CreateTodo(c), http.StatusBadRequest)
		assert.False(t, called, "service must not be called on validation failure")
	})

	t.Run("service error passes through", func(t *testing.T) {
		fake := &FakeTodoService{
			CreateTodoFn: func(ctx echo.Context, userID string, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
				return nil, errs.NewNotFoundError("category not found", false, nil)
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodPost, "/todos", strings.NewReader(`{"title":"x"}`), "", nil, testUserID)

		httpErr := requireHTTPError(t, h.CreateTodo(c), http.StatusNotFound)
		assert.Contains(t, httpErr.Message, "category not found")
	})
}

func TestTodoHandler_GetTodoByID(t *testing.T) {
	id := uuid.New()

	t.Run("success returns 200", func(t *testing.T) {
		fake := &FakeTodoService{
			GetTodoByIDFn: func(ctx echo.Context, userID string, todoID uuid.UUID) (*todo.PopulatedTodo, error) {
				assert.Equal(t, id, todoID)
				return &todo.PopulatedTodo{Todo: todo.Todo{Title: "found"}}, nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodGet, "/todos/"+id.String(), nil, "", map[string]string{"id": id.String()}, testUserID)

		require.NoError(t, h.GetTodoByID(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "found")
	})

	t.Run("invalid uuid returns 400", func(t *testing.T) {
		h := handler.NewTodoHandler(&server.Server{}, &FakeTodoService{})
		c, _ := newHandlerContext(http.MethodGet, "/todos/not-a-uuid", nil, "", map[string]string{"id": "not-a-uuid"}, testUserID)

		requireHTTPError(t, h.GetTodoByID(c), http.StatusBadRequest)
	})

	t.Run("service not found passes through", func(t *testing.T) {
		fake := &FakeTodoService{
			GetTodoByIDFn: func(ctx echo.Context, userID string, todoID uuid.UUID) (*todo.PopulatedTodo, error) {
				return nil, errs.NewNotFoundError("todo not found", false, nil)
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodGet, "/todos/"+id.String(), nil, "", map[string]string{"id": id.String()}, testUserID)

		requireHTTPError(t, h.GetTodoByID(c), http.StatusNotFound)
	})
}

func TestTodoHandler_GetTodos(t *testing.T) {
	t.Run("success returns 200 with query string", func(t *testing.T) {
		fake := &FakeTodoService{
			GetTodosFn: func(ctx echo.Context, userID string, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error) {
				require.NotNil(t, query.Page)
				assert.Equal(t, 1, *query.Page)
				return &model.PaginatedResponse[todo.PopulatedTodo]{
					Data: []todo.PopulatedTodo{{Todo: todo.Todo{Title: "one"}}},
					Page: 1, Limit: 10, Total: 1, TotalPages: 1,
				}, nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodGet, "/todos?page=1&limit=10", nil, "", nil, testUserID)

		require.NoError(t, h.GetTodos(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "one")
	})

	t.Run("service error passes through 500", func(t *testing.T) {
		fake := &FakeTodoService{
			GetTodosFn: func(ctx echo.Context, userID string, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error) {
				return nil, errs.NewInternalServerError()
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodGet, "/todos", nil, "", nil, testUserID)

		requireHTTPError(t, h.GetTodos(c), http.StatusInternalServerError)
	})
}

func TestTodoHandler_UpdateTodo(t *testing.T) {
	id := uuid.New()

	t.Run("success returns 200", func(t *testing.T) {
		fake := &FakeTodoService{
			UpdateTodoFn: func(ctx echo.Context, userID string, payload *todo.UpdateTodoPayload) (*todo.Todo, error) {
				assert.Equal(t, id, payload.ID)
				return &todo.Todo{Title: "updated"}, nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodPatch, "/todos/"+id.String(), strings.NewReader(`{"title":"updated"}`), "", map[string]string{"id": id.String()}, testUserID)

		require.NoError(t, h.UpdateTodo(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "updated")
	})

	t.Run("validation error returns 400", func(t *testing.T) {
		h := handler.NewTodoHandler(&server.Server{}, &FakeTodoService{})
		c, _ := newHandlerContext(http.MethodPatch, "/todos/"+id.String(), strings.NewReader(`{"status":"bogus"}`), "", map[string]string{"id": id.String()}, testUserID)

		requireHTTPError(t, h.UpdateTodo(c), http.StatusBadRequest)
	})

	t.Run("service not found passes through", func(t *testing.T) {
		fake := &FakeTodoService{
			UpdateTodoFn: func(ctx echo.Context, userID string, payload *todo.UpdateTodoPayload) (*todo.Todo, error) {
				return nil, errs.NewNotFoundError("todo not found", false, nil)
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodPatch, "/todos/"+id.String(), strings.NewReader(`{"title":"x"}`), "", map[string]string{"id": id.String()}, testUserID)

		requireHTTPError(t, h.UpdateTodo(c), http.StatusNotFound)
	})
}

func TestTodoHandler_DeleteTodo(t *testing.T) {
	id := uuid.New()

	t.Run("success returns 204", func(t *testing.T) {
		fake := &FakeTodoService{
			DeleteTodoFn: func(ctx echo.Context, userID string, todoID uuid.UUID) error {
				assert.Equal(t, id, todoID)
				return nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodDelete, "/todos/"+id.String(), nil, "", map[string]string{"id": id.String()}, testUserID)

		require.NoError(t, h.DeleteTodo(c))
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("service not found passes through", func(t *testing.T) {
		fake := &FakeTodoService{
			DeleteTodoFn: func(ctx echo.Context, userID string, todoID uuid.UUID) error {
				return errs.NewNotFoundError("todo not found", false, nil)
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodDelete, "/todos/"+id.String(), nil, "", map[string]string{"id": id.String()}, testUserID)

		requireHTTPError(t, h.DeleteTodo(c), http.StatusNotFound)
	})
}

func TestTodoHandler_GetTodoStats(t *testing.T) {
	t.Run("success returns 200", func(t *testing.T) {
		fake := &FakeTodoService{
			GetTodoStatsFn: func(ctx echo.Context, userID string) (*todo.TodoStats, error) {
				assert.Equal(t, testUserID, userID)
				return &todo.TodoStats{Total: 3, Active: 2, Completed: 1}, nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodGet, "/todos/stats", nil, "", nil, testUserID)

		require.NoError(t, h.GetTodoStats(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"total":3`)
	})
}

func TestTodoHandler_UploadTodoAttachment(t *testing.T) {
	id := uuid.New()

	t.Run("success returns 201", func(t *testing.T) {
		fake := &FakeTodoService{
			UploadTodoAttachmentFn: func(ctx echo.Context, userID string, todoID uuid.UUID, file *multipart.FileHeader) (*todo.TodoAttachment, error) {
				assert.Equal(t, id, todoID)
				require.NotNil(t, file)
				assert.Equal(t, "note.txt", file.Filename)
				return &todo.TodoAttachment{Name: "note.txt"}, nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		req := newMultipartRequest(t, "/todos/"+id.String()+"/attachments", "file", map[string]string{"note.txt": "hello"})
		c, rec := newHandlerContext(http.MethodPost, req.URL.String(), req.Body, req.Header.Get(echo.HeaderContentType), map[string]string{"id": id.String()}, testUserID)

		require.NoError(t, h.UploadTodoAttachment(c))
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "note.txt")
	})

	t.Run("missing multipart form returns 400", func(t *testing.T) {
		h := handler.NewTodoHandler(&server.Server{}, &FakeTodoService{})
		c, _ := newHandlerContext(http.MethodPost, "/todos/"+id.String()+"/attachments", strings.NewReader(`{}`), "", map[string]string{"id": id.String()}, testUserID)

		requireHTTPError(t, h.UploadTodoAttachment(c), http.StatusBadRequest)
	})

	t.Run("no file returns 400", func(t *testing.T) {
		h := handler.NewTodoHandler(&server.Server{}, &FakeTodoService{})
		req := newMultipartRequest(t, "/todos/"+id.String()+"/attachments", "other", map[string]string{"note.txt": "hello"})
		c, _ := newHandlerContext(http.MethodPost, req.URL.String(), req.Body, req.Header.Get(echo.HeaderContentType), map[string]string{"id": id.String()}, testUserID)

		httpErr := requireHTTPError(t, h.UploadTodoAttachment(c), http.StatusBadRequest)
		assert.Contains(t, httpErr.Message, "no file found")
	})

	t.Run("multiple files returns 400", func(t *testing.T) {
		h := handler.NewTodoHandler(&server.Server{}, &FakeTodoService{})
		req := newMultipartRequest(t, "/todos/"+id.String()+"/attachments", "file", map[string]string{"a.txt": "a", "b.txt": "b"})
		c, _ := newHandlerContext(http.MethodPost, req.URL.String(), req.Body, req.Header.Get(echo.HeaderContentType), map[string]string{"id": id.String()}, testUserID)

		httpErr := requireHTTPError(t, h.UploadTodoAttachment(c), http.StatusBadRequest)
		assert.Contains(t, httpErr.Message, "only one file allowed")
	})
}

func TestTodoHandler_DeleteTodoAttachment(t *testing.T) {
	todoID := uuid.New()
	attachmentID := uuid.New()

	t.Run("success returns 204", func(t *testing.T) {
		fake := &FakeTodoService{
			DeleteTodoAttachmentFn: func(ctx echo.Context, userID string, tid, aid uuid.UUID) error {
				assert.Equal(t, todoID, tid)
				assert.Equal(t, attachmentID, aid)
				return nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodDelete, "/todos/x/attachments/y", nil, "", map[string]string{"id": todoID.String(), "attachmentId": attachmentID.String()}, testUserID)

		require.NoError(t, h.DeleteTodoAttachment(c))
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("service not found passes through", func(t *testing.T) {
		fake := &FakeTodoService{
			DeleteTodoAttachmentFn: func(ctx echo.Context, userID string, tid, aid uuid.UUID) error {
				return errs.NewNotFoundError("attachment not found", false, nil)
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodDelete, "/todos/x/attachments/y", nil, "", map[string]string{"id": todoID.String(), "attachmentId": attachmentID.String()}, testUserID)

		requireHTTPError(t, h.DeleteTodoAttachment(c), http.StatusNotFound)
	})
}

func TestTodoHandler_GetAttachmentPresignedURL(t *testing.T) {
	todoID := uuid.New()
	attachmentID := uuid.New()

	t.Run("success returns 200 with url", func(t *testing.T) {
		fake := &FakeTodoService{
			GetAttachmentPresignedURLFn: func(ctx echo.Context, userID string, tid, aid uuid.UUID) (string, error) {
				return "https://example.com/presigned", nil
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodGet, "/todos/x/attachments/y/url", nil, "", map[string]string{"id": todoID.String(), "attachmentId": attachmentID.String()}, testUserID)

		require.NoError(t, h.GetAttachmentPresignedURL(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "https://example.com/presigned")
	})

	t.Run("service not found passes through", func(t *testing.T) {
		fake := &FakeTodoService{
			GetAttachmentPresignedURLFn: func(ctx echo.Context, userID string, tid, aid uuid.UUID) (string, error) {
				return "", errs.NewNotFoundError("attachment not found", false, nil)
			},
		}
		h := handler.NewTodoHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodGet, "/todos/x/attachments/y/url", nil, "", map[string]string{"id": todoID.String(), "attachmentId": attachmentID.String()}, testUserID)

		requireHTTPError(t, h.GetAttachmentPresignedURL(c), http.StatusNotFound)
	})
}
