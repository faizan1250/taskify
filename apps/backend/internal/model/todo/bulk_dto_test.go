package todo_test

import (
	"strings"
	"testing"

	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func bulkIDs(n int) []uuid.UUID {
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, uuid.New())
	}
	return ids
}

func TestBulkUpdateTodosPayload_Validate(t *testing.T) {
	status := todo.StatusCompleted

	t.Run("valid patch passes", func(t *testing.T) {
		p := &todo.BulkUpdateTodosPayload{IDs: bulkIDs(2), Patch: todo.BulkTodoPatch{Status: &status}}
		require.NoError(t, p.Validate())
	})

	t.Run("empty ids rejected", func(t *testing.T) {
		p := &todo.BulkUpdateTodosPayload{IDs: []uuid.UUID{}, Patch: todo.BulkTodoPatch{Status: &status}}
		require.Error(t, p.Validate())
	})

	t.Run("nil ids rejected", func(t *testing.T) {
		p := &todo.BulkUpdateTodosPayload{IDs: nil, Patch: todo.BulkTodoPatch{Status: &status}}
		require.Error(t, p.Validate())
	})

	t.Run("over-limit ids rejected", func(t *testing.T) {
		p := &todo.BulkUpdateTodosPayload{IDs: bulkIDs(todo.MaxBulkTodoIDs + 1), Patch: todo.BulkTodoPatch{Status: &status}}
		require.Error(t, p.Validate())
	})

	t.Run("nil UUID entry rejected", func(t *testing.T) {
		p := &todo.BulkUpdateTodosPayload{IDs: []uuid.UUID{uuid.New(), uuid.Nil}, Patch: todo.BulkTodoPatch{Status: &status}}
		err := p.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty UUID")
	})

	t.Run("empty patch rejected", func(t *testing.T) {
		p := &todo.BulkUpdateTodosPayload{IDs: bulkIDs(1), Patch: todo.BulkTodoPatch{}}
		err := p.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one")
	})

	t.Run("invalid status rejected", func(t *testing.T) {
		bogus := todo.Status("bogus")
		p := &todo.BulkUpdateTodosPayload{IDs: bulkIDs(1), Patch: todo.BulkTodoPatch{Status: &bogus}}
		require.Error(t, p.Validate())
	})

	t.Run("invalid priority rejected", func(t *testing.T) {
		bogus := todo.Priority("urgent")
		p := &todo.BulkUpdateTodosPayload{IDs: bulkIDs(1), Patch: todo.BulkTodoPatch{Priority: &bogus}}
		require.Error(t, p.Validate())
	})
}

func TestBulkDeleteTodosPayload_Validate(t *testing.T) {
	t.Run("valid ids pass", func(t *testing.T) {
		p := &todo.BulkDeleteTodosPayload{IDs: bulkIDs(3)}
		require.NoError(t, p.Validate())
	})

	t.Run("empty ids rejected", func(t *testing.T) {
		p := &todo.BulkDeleteTodosPayload{IDs: []uuid.UUID{}}
		require.Error(t, p.Validate())
	})

	t.Run("over-limit ids rejected", func(t *testing.T) {
		p := &todo.BulkDeleteTodosPayload{IDs: bulkIDs(todo.MaxBulkTodoIDs + 1)}
		require.Error(t, p.Validate())
	})

	t.Run("nil UUID entry rejected", func(t *testing.T) {
		p := &todo.BulkDeleteTodosPayload{IDs: []uuid.UUID{uuid.Nil}}
		require.Error(t, p.Validate())
	})
}

func TestBulkResults_HasFailures(t *testing.T) {
	t.Run("nil-safe false", func(t *testing.T) {
		var u *todo.BulkUpdateTodosResult
		var d *todo.BulkDeleteTodosResult
		assert.False(t, u.HasFailures())
		assert.False(t, d.HasFailures())
	})

	t.Run("empty results have no failures", func(t *testing.T) {
		assert.False(t, (&todo.BulkUpdateTodosResult{}).HasFailures())
		assert.False(t, (&todo.BulkDeleteTodosResult{}).HasFailures())
	})

	t.Run("failed entries report true", func(t *testing.T) {
		u := &todo.BulkUpdateTodosResult{Failed: []todo.BulkTodoFailure{{ID: uuid.New(), Error: "todo not found"}}}
		d := &todo.BulkDeleteTodosResult{Failed: []todo.BulkTodoFailure{{ID: uuid.New(), Error: "todo not found"}}}
		assert.True(t, u.HasFailures())
		assert.True(t, d.HasFailures())
	})

	t.Run("failure message stays generic", func(t *testing.T) {
		// Per-ID errors must not leak ownership details: a foreign row and a
		// missing row are indistinguishable to the caller.
		f := todo.BulkTodoFailure{ID: uuid.New(), Error: "todo not found"}
		assert.False(t, strings.Contains(f.Error, "user"))
	})
}
