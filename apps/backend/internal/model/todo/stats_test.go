package todo_test

import (
	"encoding/json"
	"testing"

	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTodoStatsPayload_IncludeCategoryCounts(t *testing.T) {
	t.Run("nil payload defaults to true", func(t *testing.T) {
		var p *todo.GetTodoStatsPayload
		assert.True(t, p.IncludeCategoryCounts())
	})

	t.Run("absent param defaults to true", func(t *testing.T) {
		p := &todo.GetTodoStatsPayload{}
		assert.True(t, p.IncludeCategoryCounts())
		require.NoError(t, p.Validate())
	})

	t.Run("explicit false opts out", func(t *testing.T) {
		f := false
		p := &todo.GetTodoStatsPayload{IncludeCounts: &f}
		assert.False(t, p.IncludeCategoryCounts())
	})

	t.Run("explicit true opts in", func(t *testing.T) {
		tr := true
		p := &todo.GetTodoStatsPayload{IncludeCounts: &tr}
		assert.True(t, p.IncludeCategoryCounts())
	})
}

func TestTodoStats_JSONShape(t *testing.T) {
	stats := todo.TodoStats{
		Total:       3,
		DueThisWeek: 1,
		ByCategory:  []todo.CategoryTodoCount{},
	}
	out, err := json.Marshal(stats)
	require.NoError(t, err)
	// byCategory must encode as [] (not null) so clients can iterate safely.
	assert.Contains(t, string(out), `"dueThisWeek":1`)
	assert.Contains(t, string(out), `"byCategory":[]`)
}
