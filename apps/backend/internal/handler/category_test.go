package handler_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/faizan1250/taskify/internal/errs"
	"github.com/faizan1250/taskify/internal/handler"
	"github.com/faizan1250/taskify/internal/model"
	"github.com/faizan1250/taskify/internal/model/category"
	"github.com/faizan1250/taskify/internal/server"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoryHandler_CreateCategory(t *testing.T) {
	t.Run("success returns 201", func(t *testing.T) {
		fake := &FakeCategoryService{
			CreateCategoryFn: func(ctx echo.Context, userID string, payload *category.CreateCategoryPayload) (*category.Category, error) {
				assert.Equal(t, testUserID, userID)
				assert.Equal(t, "Work", payload.Name)
				return &category.Category{Name: payload.Name, Color: payload.Color}, nil
			},
		}
		h := handler.NewCategoryHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodPost, "/categories", strings.NewReader(`{"name":"Work","color":"#ff0000"}`), "", nil, testUserID)

		require.NoError(t, h.CreateCategory(c))
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "Work")
	})

	t.Run("validation error returns 400", func(t *testing.T) {
		called := false
		fake := &FakeCategoryService{
			CreateCategoryFn: func(ctx echo.Context, userID string, payload *category.CreateCategoryPayload) (*category.Category, error) {
				called = true
				return nil, errFakeService
			},
		}
		h := handler.NewCategoryHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodPost, "/categories", strings.NewReader(`{"name":"","color":"not-a-color"}`), "", nil, testUserID)

		requireHTTPError(t, h.CreateCategory(c), http.StatusBadRequest)
		assert.False(t, called, "service must not be called on validation failure")
	})

	t.Run("service error passes through 500", func(t *testing.T) {
		fake := &FakeCategoryService{
			CreateCategoryFn: func(ctx echo.Context, userID string, payload *category.CreateCategoryPayload) (*category.Category, error) {
				return nil, errs.NewInternalServerError()
			},
		}
		h := handler.NewCategoryHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodPost, "/categories", strings.NewReader(`{"name":"Work","color":"#ff0000"}`), "", nil, testUserID)

		requireHTTPError(t, h.CreateCategory(c), http.StatusInternalServerError)
	})
}

func TestCategoryHandler_GetCategories(t *testing.T) {
	t.Run("success returns 200", func(t *testing.T) {
		fake := &FakeCategoryService{
			GetCategoriesFn: func(ctx echo.Context, userID string, query *category.GetCategoriesQuery) (*model.PaginatedResponse[category.Category], error) {
				return &model.PaginatedResponse[category.Category]{
					Data: []category.Category{{Name: "Work", Color: "#ff0000"}},
					Page: 1, Limit: 50, Total: 1, TotalPages: 1,
				}, nil
			},
		}
		h := handler.NewCategoryHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodGet, "/categories?search=Wo", nil, "", nil, testUserID)

		require.NoError(t, h.GetCategories(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Work")
	})

	t.Run("service error passes through 500", func(t *testing.T) {
		fake := &FakeCategoryService{
			GetCategoriesFn: func(ctx echo.Context, userID string, query *category.GetCategoriesQuery) (*model.PaginatedResponse[category.Category], error) {
				return nil, errs.NewInternalServerError()
			},
		}
		h := handler.NewCategoryHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodGet, "/categories", nil, "", nil, testUserID)

		requireHTTPError(t, h.GetCategories(c), http.StatusInternalServerError)
	})
}

func TestCategoryHandler_UpdateCategory(t *testing.T) {
	id := uuid.New()

	t.Run("success returns 200", func(t *testing.T) {
		fake := &FakeCategoryService{
			UpdateCategoryFn: func(ctx echo.Context, userID string, categoryID uuid.UUID, payload *category.UpdateCategoryPayload) (*category.Category, error) {
				assert.Equal(t, id, categoryID)
				return &category.Category{Name: "Personal"}, nil
			},
		}
		h := handler.NewCategoryHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodPatch, "/categories/"+id.String(), strings.NewReader(`{"name":"Personal"}`), "", map[string]string{"id": id.String()}, testUserID)

		require.NoError(t, h.UpdateCategory(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Personal")
	})

	t.Run("validation error returns 400", func(t *testing.T) {
		h := handler.NewCategoryHandler(&server.Server{}, &FakeCategoryService{})
		c, _ := newHandlerContext(http.MethodPatch, "/categories/"+id.String(), strings.NewReader(`{"color":"nope"}`), "", map[string]string{"id": id.String()}, testUserID)

		requireHTTPError(t, h.UpdateCategory(c), http.StatusBadRequest)
	})

	t.Run("service not found passes through", func(t *testing.T) {
		fake := &FakeCategoryService{
			UpdateCategoryFn: func(ctx echo.Context, userID string, categoryID uuid.UUID, payload *category.UpdateCategoryPayload) (*category.Category, error) {
				return nil, errs.NewNotFoundError("category not found", false, nil)
			},
		}
		h := handler.NewCategoryHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodPatch, "/categories/"+id.String(), strings.NewReader(`{"name":"x"}`), "", map[string]string{"id": id.String()}, testUserID)

		requireHTTPError(t, h.UpdateCategory(c), http.StatusNotFound)
	})
}

func TestCategoryHandler_DeleteCategory(t *testing.T) {
	id := uuid.New()

	t.Run("success returns 204", func(t *testing.T) {
		fake := &FakeCategoryService{
			DeleteCategoryFn: func(ctx echo.Context, userID string, categoryID uuid.UUID) error {
				assert.Equal(t, id, categoryID)
				return nil
			},
		}
		h := handler.NewCategoryHandler(&server.Server{}, fake)
		c, rec := newHandlerContext(http.MethodDelete, "/categories/"+id.String(), nil, "", map[string]string{"id": id.String()}, testUserID)

		require.NoError(t, h.DeleteCategory(c))
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("service not found passes through", func(t *testing.T) {
		fake := &FakeCategoryService{
			DeleteCategoryFn: func(ctx echo.Context, userID string, categoryID uuid.UUID) error {
				return errs.NewNotFoundError("category not found", false, nil)
			},
		}
		h := handler.NewCategoryHandler(&server.Server{}, fake)
		c, _ := newHandlerContext(http.MethodDelete, "/categories/"+id.String(), nil, "", map[string]string{"id": id.String()}, testUserID)

		requireHTTPError(t, h.DeleteCategory(c), http.StatusNotFound)
	})
}
