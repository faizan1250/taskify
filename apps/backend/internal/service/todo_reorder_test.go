package service_test

import (
	"context"
	"testing"

	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/faizan1250/taskify/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodoService_ReorderTodos(t *testing.T) {
	idA, idB, idC := uuid.New(), uuid.New(), uuid.New()

	t.Run("happy path passes order through user-scoped", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		var gotIDs []uuid.UUID
		todoRepo.ReorderTodosFn = func(_ context.Context, userID string, orderedIDs []uuid.UUID) ([]todo.Todo, error) {
			assert.Equal(t, "user-1", userID)
			gotIDs = orderedIDs
			return []todo.Todo{
				{Title: "c", Position: 0},
				{Title: "a", Position: 1},
				{Title: "b", Position: 2},
			}, nil
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
		got, err := svc.ReorderTodos(newEchoCtx(), "user-1", &todo.ReorderTodosPayload{
			OrderedIDs: []uuid.UUID{idC, idA, idB},
		})
		require.NoError(t, err)
		assert.Equal(t, []uuid.UUID{idC, idA, idB}, gotIDs)
		require.Len(t, got, 3)
		assert.Equal(t, "c", got[0].Title)
	})

	t.Run("foreign or missing id surfaces repo 404", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.ReorderTodosFn = func(_ context.Context, _ string, _ []uuid.UUID) ([]todo.Todo, error) {
			return nil, errRepoBoom
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
		_, err := svc.ReorderTodos(newEchoCtx(), "user-1", &todo.ReorderTodosPayload{
			OrderedIDs: []uuid.UUID{idA},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "repo boom")
	})
}
