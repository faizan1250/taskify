package todo_test

import (
	"testing"

	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func reorderIDs(n int) []uuid.UUID {
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, uuid.New())
	}
	return ids
}

func TestReorderTodosPayload_Validate(t *testing.T) {
	t.Run("valid order passes", func(t *testing.T) {
		p := &todo.ReorderTodosPayload{OrderedIDs: reorderIDs(3)}
		require.NoError(t, p.Validate())
	})

	t.Run("empty order rejected", func(t *testing.T) {
		p := &todo.ReorderTodosPayload{OrderedIDs: []uuid.UUID{}}
		require.Error(t, p.Validate())
	})

	t.Run("nil order rejected", func(t *testing.T) {
		p := &todo.ReorderTodosPayload{OrderedIDs: nil}
		require.Error(t, p.Validate())
	})

	t.Run("over-limit order rejected", func(t *testing.T) {
		p := &todo.ReorderTodosPayload{OrderedIDs: reorderIDs(todo.MaxReorderTodoIDs + 1)}
		require.Error(t, p.Validate())
	})

	t.Run("nil UUID entry rejected", func(t *testing.T) {
		p := &todo.ReorderTodosPayload{OrderedIDs: []uuid.UUID{uuid.New(), uuid.Nil}}
		err := p.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty UUID")
	})

	t.Run("duplicate IDs rejected", func(t *testing.T) {
		dup := uuid.New()
		p := &todo.ReorderTodosPayload{OrderedIDs: []uuid.UUID{dup, uuid.New(), dup}}
		err := p.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicates")
	})
}

func TestPositionsForOrder(t *testing.T) {
	t.Run("assigns zero-based float slots", func(t *testing.T) {
		ids := reorderIDs(3)
		got := todo.PositionsForOrder(ids)
		require.Len(t, got, 3)
		for i, id := range ids {
			assert.Equal(t, float64(i), got[id])
		}
	})

	t.Run("empty input yields empty map", func(t *testing.T) {
		assert.Empty(t, todo.PositionsForOrder(nil))
	})

	t.Run("single item sits at zero", func(t *testing.T) {
		id := uuid.New()
		got := todo.PositionsForOrder([]uuid.UUID{id})
		assert.Equal(t, map[uuid.UUID]float64{id: 0}, got)
	})
}

func TestGetTodosQuery_PositionSortAccepted(t *testing.T) {
	sort := "position"
	order := "asc"
	q := &todo.GetTodosQuery{Sort: &sort, Order: &order}
	require.NoError(t, q.Validate())
	assert.Equal(t, "position", *q.Sort)
}
