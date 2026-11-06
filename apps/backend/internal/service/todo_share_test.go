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

func TestTodoService_ShareTodo(t *testing.T) {
	todoID := uuid.New()

	t.Run("happy path returns token", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.CheckTodoExistsFn = func(_ context.Context, userID string, id uuid.UUID) (*todo.Todo, error) {
			assert.Equal(t, "user-1", userID)
			assert.Equal(t, todoID, id)
			return &todo.Todo{}, nil
		}
		todoRepo.CreateShareTokenFn = func(_ context.Context, userID string, id uuid.UUID) (*todo.TodoShareToken, error) {
			assert.Equal(t, "user-1", userID)
			return &todo.TodoShareToken{Token: "tok", TodoID: id}, nil
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)

		got, err := svc.ShareTodo(newEchoCtx(), "user-1", todoID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "tok", got.Token)
		assert.Equal(t, todoID, got.TodoID)
	})

	t.Run("foreign todo propagates ownership error", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
			return nil, errRepoBoom
		}
		called := false
		todoRepo.CreateShareTokenFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.TodoShareToken, error) {
			called = true
			return nil, errRepoBoom
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)

		_, err := svc.ShareTodo(newEchoCtx(), "user-1", todoID)
		require.Error(t, err)
		assert.False(t, called, "token must not be minted when ownership check fails")
	})

	t.Run("repo error propagates", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
			return &todo.Todo{}, nil
		}
		todoRepo.CreateShareTokenFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.TodoShareToken, error) {
			return nil, errRepoBoom
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)

		_, err := svc.ShareTodo(newEchoCtx(), "user-1", todoID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "repo boom")
	})
}

func TestTodoService_GetSharedTodo(t *testing.T) {
	t.Run("happy path returns populated todo", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.GetSharedTodoFn = func(_ context.Context, token string) (*todo.PopulatedTodo, error) {
			assert.Equal(t, "tok", token)
			return &todo.PopulatedTodo{Todo: todo.Todo{Title: "shared"}}, nil
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)

		got, err := svc.GetSharedTodo(newEchoCtx(), "tok")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "shared", got.Title)
	})

	t.Run("unknown token propagates", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.GetSharedTodoFn = func(_ context.Context, _ string) (*todo.PopulatedTodo, error) {
			return nil, errRepoBoom
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)

		_, err := svc.GetSharedTodo(newEchoCtx(), "nope")
		require.Error(t, err)
	})
}

func TestTodoService_UnshareTodo(t *testing.T) {
	todoID := uuid.New()

	t.Run("happy path revokes", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
			return &todo.Todo{}, nil
		}
		revoked := false
		todoRepo.RevokeShareTokensForTodoFn = func(_ context.Context, userID string, id uuid.UUID) error {
			assert.Equal(t, "user-1", userID)
			assert.Equal(t, todoID, id)
			revoked = true
			return nil
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)

		require.NoError(t, svc.UnshareTodo(newEchoCtx(), "user-1", todoID))
		assert.True(t, revoked)
	})

	t.Run("foreign todo never reaches revoke", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
			return nil, errRepoBoom
		}
		called := false
		todoRepo.RevokeShareTokensForTodoFn = func(_ context.Context, _ string, _ uuid.UUID) error {
			called = true
			return nil
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)

		require.Error(t, svc.UnshareTodo(newEchoCtx(), "user-1", todoID))
		assert.False(t, called, "revoke must not run when ownership check fails")
	})
}
