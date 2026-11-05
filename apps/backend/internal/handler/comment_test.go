package handler_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/faizan1250/taskify/internal/errs"
	"github.com/faizan1250/taskify/internal/handler"
	"github.com/faizan1250/taskify/internal/model/comment"
	"github.com/faizan1250/taskify/internal/server"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommentHandler_AddComment(t *testing.T) {
	todoID := uuid.New()

	t.Run("success returns 201", func(t *testing.T) {
		fake := &FakeCommentService{
			AddCommentFn: func(ctx echo.Context, userID string, tid uuid.UUID, payload *comment.AddCommentPayload) (*comment.Comment, error) {
				assert.Equal(t, todoID, tid)
				assert.Equal(t, "nice work", payload.Content)
				return &comment.Comment{TodoID: tid, UserID: userID, Content: payload.Content}, nil
			},
		}
		h := handler.NewCommentHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodPost, "/todos/"+todoID.String()+"/comments", strings.NewReader(`{"content":"nice work"}`), "", map[string]string{"id": todoID.String()}, testUserID)

		require.NoError(t, h.AddComment(c))
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "nice work")
	})

	t.Run("validation error returns 400", func(t *testing.T) {
		called := false
		fake := &FakeCommentService{
			AddCommentFn: func(ctx echo.Context, userID string, tid uuid.UUID, payload *comment.AddCommentPayload) (*comment.Comment, error) {
				called = true
				return nil, errFakeService
			},
		}
		h := handler.NewCommentHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodPost, "/todos/"+todoID.String()+"/comments", strings.NewReader(`{"content":""}`), "", map[string]string{"id": todoID.String()}, testUserID)

		requireHTTPError(t, h.AddComment(c), http.StatusBadRequest)
		assert.False(t, called, "service must not be called on validation failure")
	})

	t.Run("service not found passes through", func(t *testing.T) {
		fake := &FakeCommentService{
			AddCommentFn: func(ctx echo.Context, userID string, tid uuid.UUID, payload *comment.AddCommentPayload) (*comment.Comment, error) {
				return nil, errs.NewNotFoundError("todo not found", false, nil)
			},
		}
		h := handler.NewCommentHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodPost, "/todos/"+todoID.String()+"/comments", strings.NewReader(`{"content":"hi"}`), "", map[string]string{"id": todoID.String()}, testUserID)

		requireHTTPError(t, h.AddComment(c), http.StatusNotFound)
	})
}

func TestCommentHandler_GetCommentsByTodoID(t *testing.T) {
	todoID := uuid.New()

	t.Run("success returns 200", func(t *testing.T) {
		fake := &FakeCommentService{
			GetCommentsByTodoIDFn: func(ctx echo.Context, userID string, tid uuid.UUID) ([]comment.Comment, error) {
				assert.Equal(t, todoID, tid)
				return []comment.Comment{{TodoID: tid, Content: "first"}}, nil
			},
		}
		h := handler.NewCommentHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodGet, "/todos/"+todoID.String()+"/comments", nil, "", map[string]string{"id": todoID.String()}, testUserID)

		require.NoError(t, h.GetCommentsByTodoID(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "first")
	})

	t.Run("service not found passes through", func(t *testing.T) {
		fake := &FakeCommentService{
			GetCommentsByTodoIDFn: func(ctx echo.Context, userID string, tid uuid.UUID) ([]comment.Comment, error) {
				return nil, errs.NewNotFoundError("todo not found", false, nil)
			},
		}
		h := handler.NewCommentHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodGet, "/todos/"+todoID.String()+"/comments", nil, "", map[string]string{"id": todoID.String()}, testUserID)

		requireHTTPError(t, h.GetCommentsByTodoID(c), http.StatusNotFound)
	})
}

func TestCommentHandler_UpdateComment(t *testing.T) {
	id := uuid.New()

	t.Run("success returns 200", func(t *testing.T) {
		fake := &FakeCommentService{
			UpdateCommentFn: func(ctx echo.Context, userID string, commentID uuid.UUID, content string) (*comment.Comment, error) {
				assert.Equal(t, id, commentID)
				assert.Equal(t, "edited", content)
				return &comment.Comment{Content: content}, nil
			},
		}
		h := handler.NewCommentHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodPatch, "/comments/"+id.String(), strings.NewReader(`{"content":"edited"}`), "", map[string]string{"id": id.String()}, testUserID)

		require.NoError(t, h.UpdateComment(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "edited")
	})

	t.Run("validation error returns 400", func(t *testing.T) {
		h := handler.NewCommentHandler(&server.Server{}, &FakeCommentService{})
		c, _ := newHandlerContext(http.MethodPatch, "/comments/"+id.String(), strings.NewReader(`{"content":""}`), "", map[string]string{"id": id.String()}, testUserID)

		requireHTTPError(t, h.UpdateComment(c), http.StatusBadRequest)
	})

	t.Run("service not found passes through", func(t *testing.T) {
		fake := &FakeCommentService{
			UpdateCommentFn: func(ctx echo.Context, userID string, commentID uuid.UUID, content string) (*comment.Comment, error) {
				return nil, errs.NewNotFoundError("comment not found", false, nil)
			},
		}
		h := handler.NewCommentHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodPatch, "/comments/"+id.String(), strings.NewReader(`{"content":"x"}`), "", map[string]string{"id": id.String()}, testUserID)

		requireHTTPError(t, h.UpdateComment(c), http.StatusNotFound)
	})
}

func TestCommentHandler_DeleteComment(t *testing.T) {
	id := uuid.New()

	t.Run("success returns 204", func(t *testing.T) {
		fake := &FakeCommentService{
			DeleteCommentFn: func(ctx echo.Context, userID string, commentID uuid.UUID) error {
				assert.Equal(t, id, commentID)
				return nil
			},
		}
		h := handler.NewCommentHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodDelete, "/comments/"+id.String(), nil, "", map[string]string{"id": id.String()}, testUserID)

		require.NoError(t, h.DeleteComment(c))
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("service not found passes through", func(t *testing.T) {
		fake := &FakeCommentService{
			DeleteCommentFn: func(ctx echo.Context, userID string, commentID uuid.UUID) error {
				return errs.NewNotFoundError("comment not found", false, nil)
			},
		}
		h := handler.NewCommentHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodDelete, "/comments/"+id.String(), nil, "", map[string]string{"id": id.String()}, testUserID)

		requireHTTPError(t, h.DeleteComment(c), http.StatusNotFound)
	})
}
