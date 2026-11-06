package todo_test

import (
	"testing"
	"time"

	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNextOccurrence(t *testing.T) {
	base := time.Date(2026, time.January, 15, 10, 0, 0, 0, time.UTC)

	t.Run("daily advances 24h", func(t *testing.T) {
		next, ok := todo.NextOccurrence(base, todo.RecurrenceDaily)
		require.True(t, ok)
		assert.Equal(t, base.Add(24*time.Hour), next)
	})

	t.Run("weekly advances 7d", func(t *testing.T) {
		next, ok := todo.NextOccurrence(base, todo.RecurrenceWeekly)
		require.True(t, ok)
		assert.Equal(t, base.Add(7*24*time.Hour), next)
	})

	t.Run("monthly advances one calendar month", func(t *testing.T) {
		next, ok := todo.NextOccurrence(base, todo.RecurrenceMonthly)
		require.True(t, ok)
		assert.Equal(t, time.Date(2026, time.February, 15, 10, 0, 0, 0, time.UTC), next)
	})

	t.Run("monthly clamps end-of-month instead of overflowing", func(t *testing.T) {
		jan31 := time.Date(2026, time.January, 31, 10, 0, 0, 0, time.UTC)
		next, ok := todo.NextOccurrence(jan31, todo.RecurrenceMonthly)
		require.True(t, ok)
		// Go AddDate normalizes: Jan 31 + 1 month = Feb 28 (2026 is not a leap year).
		assert.Equal(t, time.Date(2026, time.February, 28, 10, 0, 0, 0, time.UTC), next)
	})

	t.Run("none never schedules", func(t *testing.T) {
		_, ok := todo.NextOccurrence(base, todo.RecurrenceNone)
		assert.False(t, ok)
	})

	t.Run("unknown rule never schedules", func(t *testing.T) {
		_, ok := todo.NextOccurrence(base, todo.Recurrence("yearly"))
		assert.False(t, ok)
	})
}

func TestIsRecurring(t *testing.T) {
	assert.False(t, todo.RecurrenceNone.IsRecurring())
	assert.False(t, todo.Recurrence("").IsRecurring())
	assert.True(t, todo.RecurrenceDaily.IsRecurring())
	assert.True(t, todo.RecurrenceWeekly.IsRecurring())
	assert.True(t, todo.RecurrenceMonthly.IsRecurring())
}

func TestResolveNextRun(t *testing.T) {
	now := time.Date(2026, time.January, 15, 10, 0, 0, 0, time.UTC)

	t.Run("none resolves to nil", func(t *testing.T) {
		due := now.Add(2 * time.Hour)
		assert.Nil(t, todo.ResolveNextRun(todo.RecurrenceNone, &due, now))
	})

	t.Run("first occurrence fires at due date when set", func(t *testing.T) {
		due := now.Add(2 * time.Hour)
		next := todo.ResolveNextRun(todo.RecurrenceDaily, &due, now)
		require.NotNil(t, next)
		assert.True(t, next.Equal(due))
	})

	t.Run("past due date is kept so cron catches up on next pass", func(t *testing.T) {
		past := now.Add(-48 * time.Hour)
		next := todo.ResolveNextRun(todo.RecurrenceWeekly, &past, now)
		require.NotNil(t, next)
		assert.True(t, next.Equal(past))
	})

	t.Run("no due date starts one step from now", func(t *testing.T) {
		next := todo.ResolveNextRun(todo.RecurrenceDaily, nil, now)
		require.NotNil(t, next)
		assert.True(t, next.Equal(now.Add(24*time.Hour)))
	})
}

func TestRecurrencePayloadValidation(t *testing.T) {
	t.Run("create rejects unknown recurrence", func(t *testing.T) {
		p := &todo.CreateTodoPayload{Title: "x", Recurrence: recurrencePtr(todo.Recurrence("yearly"))}
		assert.Error(t, p.Validate())
	})

	t.Run("create accepts every known rule", func(t *testing.T) {
		for _, r := range []todo.Recurrence{todo.RecurrenceNone, todo.RecurrenceDaily, todo.RecurrenceWeekly, todo.RecurrenceMonthly} {
			p := &todo.CreateTodoPayload{Title: "x", Recurrence: recurrencePtr(r)}
			assert.NoError(t, p.Validate())
		}
	})

	t.Run("update rejects unknown recurrence", func(t *testing.T) {
		p := &todo.UpdateTodoPayload{Recurrence: recurrencePtr(todo.Recurrence("hourly"))}
		assert.Error(t, p.Validate())
	})

	t.Run("query rejects unknown recurrence filter", func(t *testing.T) {
		q := &todo.GetTodosQuery{Recurrence: recurrencePtr(todo.Recurrence("yearly"))}
		assert.Error(t, q.Validate())
	})
}

func recurrencePtr(r todo.Recurrence) *todo.Recurrence {
	return &r
}
