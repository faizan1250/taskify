package service_test

import (
	"context"
	"testing"

	"github.com/faizan1250/taskify/internal/model/category"
	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/faizan1250/taskify/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodoService_BulkUpdateTodos(t *testing.T) {
	status := todo.StatusCompleted
	idA, idB := uuid.New(), uuid.New()

	t.Run("happy path dedupes ids and passes patch through", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		var gotIDs []uuid.UUID
		todoRepo.BulkUpdateTodosFn = func(_ context.Context, userID string, ids []uuid.UUID, patch *todo.BulkTodoPatch) (*todo.BulkUpdateTodosResult, error) {
			assert.Equal(t, "user-1", userID)
			gotIDs = ids
			require.NotNil(t, patch.Status)
			assert.Equal(t, status, *patch.Status)
			return &todo.BulkUpdateTodosResult{
				Updated: []todo.Todo{{Title: "a"}, {Title: "b"}},
				Failed:  []todo.BulkTodoFailure{},
			}, nil
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
		got, err := svc.BulkUpdateTodos(newEchoCtx(), "user-1", &todo.BulkUpdateTodosPayload{
			IDs:   []uuid.UUID{idA, idA, idB},
			Patch: todo.BulkTodoPatch{Status: &status},
		})
		require.NoError(t, err)
		assert.Equal(t, []uuid.UUID{idA, idB}, gotIDs)
		assert.Len(t, got.Updated, 2)
		assert.Empty(t, got.Failed)
		assert.False(t, got.HasFailures())
	})

	t.Run("partial failures pass through", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		missing := uuid.New()
		todoRepo.BulkUpdateTodosFn = func(_ context.Context, _ string, _ []uuid.UUID, _ *todo.BulkTodoPatch) (*todo.BulkUpdateTodosResult, error) {
			return &todo.BulkUpdateTodosResult{
				Updated: []todo.Todo{{Title: "a"}},
				Failed:  []todo.BulkTodoFailure{{ID: missing, Error: "todo not found"}},
			}, nil
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
		got, err := svc.BulkUpdateTodos(newEchoCtx(), "user-1", &todo.BulkUpdateTodosPayload{
			IDs:   []uuid.UUID{idA, missing},
			Patch: todo.BulkTodoPatch{Status: &status},
		})
		require.NoError(t, err)
		require.True(t, got.HasFailures())
		assert.Len(t, got.Updated, 1)
		require.Len(t, got.Failed, 1)
		assert.Equal(t, missing, got.Failed[0].ID)
	})

	t.Run("category validated once up front", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		catRepo := &FakeCategoryRepository{}
		categoryID := uuid.New()
		var catChecks int
		catRepo.GetCategoryByIDFn = func(_ context.Context, userID string, id uuid.UUID) (*category.Category, error) {
			catChecks++
			assert.Equal(t, "user-1", userID)
			assert.Equal(t, categoryID, id)
			return &category.Category{}, nil
		}
		called := false
		todoRepo.BulkUpdateTodosFn = func(_ context.Context, _ string, ids []uuid.UUID, _ *todo.BulkTodoPatch) (*todo.BulkUpdateTodosResult, error) {
			called = true
			assert.Len(t, ids, 2)
			return &todo.BulkUpdateTodosResult{Updated: []todo.Todo{}, Failed: []todo.BulkTodoFailure{}}, nil
		}
		svc := service.NewTodoService(nil, todoRepo, catRepo, nil)
		_, err := svc.BulkUpdateTodos(newEchoCtx(), "user-1", &todo.BulkUpdateTodosPayload{
			IDs:   []uuid.UUID{idA, idB},
			Patch: todo.BulkTodoPatch{CategoryID: &categoryID},
		})
		require.NoError(t, err)
		assert.Equal(t, 1, catChecks, "category must be checked once for the batch, not per row")
		assert.True(t, called)
	})

	t.Run("foreign category aborts before repo call", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		catRepo := &FakeCategoryRepository{}
		categoryID := uuid.New()
		catRepo.GetCategoryByIDFn = func(_ context.Context, _ string, _ uuid.UUID) (*category.Category, error) {
			return nil, errRepoBoom
		}
		called := false
		todoRepo.BulkUpdateTodosFn = func(_ context.Context, _ string, _ []uuid.UUID, _ *todo.BulkTodoPatch) (*todo.BulkUpdateTodosResult, error) {
			called = true
			return nil, nil
		}
		svc := service.NewTodoService(nil, todoRepo, catRepo, nil)
		_, err := svc.BulkUpdateTodos(newEchoCtx(), "user-1", &todo.BulkUpdateTodosPayload{
			IDs:   []uuid.UUID{idA},
			Patch: todo.BulkTodoPatch{CategoryID: &categoryID},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "repo boom")
		assert.False(t, called, "repo must not run when category authz fails")
	})

	t.Run("repo error propagates", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.BulkUpdateTodosFn = func(_ context.Context, _ string, _ []uuid.UUID, _ *todo.BulkTodoPatch) (*todo.BulkUpdateTodosResult, error) {
			return nil, errRepoBoom
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
		_, err := svc.BulkUpdateTodos(newEchoCtx(), "user-1", &todo.BulkUpdateTodosPayload{
			IDs:   []uuid.UUID{idA},
			Patch: todo.BulkTodoPatch{Status: &status},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "repo boom")
	})
}

func TestTodoService_BulkDeleteTodos(t *testing.T) {
	idA, idB := uuid.New(), uuid.New()

	t.Run("happy path collects keys and deletes", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.GetAttachmentKeysForTodoAndSubtasksFn = func(_ context.Context, userID string, id uuid.UUID) ([]string, error) {
			assert.Equal(t, "user-1", userID)
			if id == idA {
				return []string{"k1"}, nil
			}
			return nil, nil
		}
		var gotIDs []uuid.UUID
		todoRepo.BulkDeleteTodosFn = func(_ context.Context, userID string, ids []uuid.UUID) (*todo.BulkDeleteTodosResult, error) {
			assert.Equal(t, "user-1", userID)
			gotIDs = ids
			return &todo.BulkDeleteTodosResult{Deleted: ids, Failed: []todo.BulkTodoFailure{}}, nil
		}
		// awsClient nil: CleanupS3AttachmentsAsync is a documented no-op.
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
		got, err := svc.BulkDeleteTodos(newEchoCtx(), "user-1", &todo.BulkDeleteTodosPayload{IDs: []uuid.UUID{idA, idB}})
		require.NoError(t, err)
		assert.Equal(t, []uuid.UUID{idA, idB}, gotIDs)
		assert.Equal(t, []uuid.UUID{idA, idB}, got.Deleted)
		assert.False(t, got.HasFailures())
	})

	t.Run("partial failures pass through", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.GetAttachmentKeysForTodoAndSubtasksFn = func(_ context.Context, _ string, _ uuid.UUID) ([]string, error) {
			return nil, nil
		}
		missing := uuid.New()
		todoRepo.BulkDeleteTodosFn = func(_ context.Context, _ string, _ []uuid.UUID) (*todo.BulkDeleteTodosResult, error) {
			return &todo.BulkDeleteTodosResult{
				Deleted: []uuid.UUID{idA},
				Failed:  []todo.BulkTodoFailure{{ID: missing, Error: "todo not found"}},
			}, nil
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
		got, err := svc.BulkDeleteTodos(newEchoCtx(), "user-1", &todo.BulkDeleteTodosPayload{IDs: []uuid.UUID{idA, missing}})
		require.NoError(t, err)
		require.True(t, got.HasFailures())
		assert.Equal(t, []uuid.UUID{idA}, got.Deleted)
		require.Len(t, got.Failed, 1)
		assert.Equal(t, missing, got.Failed[0].ID)
	})

	t.Run("attachment keys error aborts before delete", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.GetAttachmentKeysForTodoAndSubtasksFn = func(_ context.Context, _ string, _ uuid.UUID) ([]string, error) {
			return nil, errRepoBoom
		}
		called := false
		todoRepo.BulkDeleteTodosFn = func(_ context.Context, _ string, _ []uuid.UUID) (*todo.BulkDeleteTodosResult, error) {
			called = true
			return nil, nil
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
		_, err := svc.BulkDeleteTodos(newEchoCtx(), "user-1", &todo.BulkDeleteTodosPayload{IDs: []uuid.UUID{idA}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "repo boom")
		assert.False(t, called)
	})

	t.Run("repo error propagates", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.GetAttachmentKeysForTodoAndSubtasksFn = func(_ context.Context, _ string, _ uuid.UUID) ([]string, error) {
			return nil, nil
		}
		todoRepo.BulkDeleteTodosFn = func(_ context.Context, _ string, _ []uuid.UUID) (*todo.BulkDeleteTodosResult, error) {
			return nil, errRepoBoom
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
		_, err := svc.BulkDeleteTodos(newEchoCtx(), "user-1", &todo.BulkDeleteTodosPayload{IDs: []uuid.UUID{idA}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "repo boom")
	})

	t.Run("duplicate ids deleted once", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.GetAttachmentKeysForTodoAndSubtasksFn = func(_ context.Context, _ string, _ uuid.UUID) ([]string, error) {
			return nil, nil
		}
		var gotIDs []uuid.UUID
		todoRepo.BulkDeleteTodosFn = func(_ context.Context, _ string, ids []uuid.UUID) (*todo.BulkDeleteTodosResult, error) {
			gotIDs = ids
			return &todo.BulkDeleteTodosResult{Deleted: ids, Failed: []todo.BulkTodoFailure{}}, nil
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
		_, err := svc.BulkDeleteTodos(newEchoCtx(), "user-1", &todo.BulkDeleteTodosPayload{IDs: []uuid.UUID{idA, idA}})
		require.NoError(t, err)
		assert.Equal(t, []uuid.UUID{idA}, gotIDs)
	})
}
