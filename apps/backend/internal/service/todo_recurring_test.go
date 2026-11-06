package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/faizan1250/taskify/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Recurrence is server-scheduled state derived from the rule, never
// client-supplied beyond the rule itself. These tests pin the service
// contract: the rule passes through to the repository untouched, and
// updates carry rule changes the same way.
func TestTodoService_RecurrencePassthrough(t *testing.T) {
	daily := todo.RecurrenceDaily
	none := todo.RecurrenceNone

	t.Run("create forwards recurrence rule to repository", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		var got *todo.CreateTodoPayload
		todoRepo.CreateTodoFn = func(_ context.Context, _ string, p *todo.CreateTodoPayload) (*todo.Todo, error) {
			got = p
			return &todo.Todo{Title: p.Title, Recurrence: *p.Recurrence}, nil
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)

		due := time.Now().Add(24 * time.Hour)
		created, err := svc.CreateTodo(newEchoCtx(), "user-1", &todo.CreateTodoPayload{
			Title:      "daily standup",
			DueDate:    &due,
			Recurrence: &daily,
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.Recurrence)
		assert.Equal(t, todo.RecurrenceDaily, *got.Recurrence)
		assert.Equal(t, todo.RecurrenceDaily, created.Recurrence)
	})

	t.Run("create without rule leaves recurrence nil for repo defaulting", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		var got *todo.CreateTodoPayload
		todoRepo.CreateTodoFn = func(_ context.Context, _ string, p *todo.CreateTodoPayload) (*todo.Todo, error) {
			got = p
			return &todo.Todo{Title: p.Title}, nil
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)

		_, err := svc.CreateTodo(newEchoCtx(), "user-1", &todo.CreateTodoPayload{Title: "one-off"})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Nil(t, got.Recurrence)
	})

	t.Run("update forwards recurrence change to repository", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		var got *todo.UpdateTodoPayload
		todoRepo.UpdateTodoFn = func(_ context.Context, _ string, p *todo.UpdateTodoPayload) (*todo.Todo, error) {
			got = p
			return &todo.Todo{Recurrence: *p.Recurrence}, nil
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)

		id := uuid.New()
		updated, err := svc.UpdateTodo(newEchoCtx(), "user-1", &todo.UpdateTodoPayload{ID: id, Recurrence: &none})
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.Recurrence)
		assert.Equal(t, todo.RecurrenceNone, *got.Recurrence)
		assert.Equal(t, todo.RecurrenceNone, updated.Recurrence)
	})

	t.Run("recurring list filter validates", func(t *testing.T) {
		q := &todo.GetTodosQuery{Recurrence: &daily}
		require.NoError(t, q.Validate())
		assert.Equal(t, todo.RecurrenceDaily, *q.Recurrence)
	})
}
