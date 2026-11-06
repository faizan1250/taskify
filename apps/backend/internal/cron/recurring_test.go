package cron_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faizan1250/taskify/internal/cron"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubMaterializer is a scripted recurringMaterializer. It records the
// pass-through arguments so tests pin the job's contract without a database.
type stubMaterializer struct {
	count    int
	err      error
	gotNow   time.Time
	gotLimit int
	calls    int
}

func (s *stubMaterializer) MaterializeRecurringTodos(_ context.Context, now time.Time, limit int) (int, error) {
	s.calls++
	s.gotNow = now
	s.gotLimit = limit
	return s.count, s.err
}

func TestRecurringTodosJob_Registered(t *testing.T) {
	registry := cron.NewJobRegistry()
	job, err := registry.Get("recurring-todos")
	require.NoError(t, err)
	assert.Equal(t, "recurring-todos", job.Name())
	assert.NotEmpty(t, job.Description())
	assert.Contains(t, registry.Help(), "recurring-todos")
}

func TestMaterializeDueRecurring_PassThrough(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	stub := &stubMaterializer{count: 3}

	count, err := cron.MaterializeDueRecurringForTest(ctx, stub, now, 100)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, 1, stub.calls)
	assert.True(t, stub.gotNow.Equal(now))
	assert.Equal(t, 100, stub.gotLimit)
}

func TestMaterializeDueRecurring_Empty(t *testing.T) {
	ctx := context.Background()
	stub := &stubMaterializer{count: 0}

	count, err := cron.MaterializeDueRecurringForTest(ctx, stub, time.Now(), 100)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestMaterializeDueRecurring_ErrorPropagates(t *testing.T) {
	ctx := context.Background()
	expected := errors.New("db unavailable")
	stub := &stubMaterializer{err: expected}

	_, err := cron.MaterializeDueRecurringForTest(ctx, stub, time.Now(), 100)
	assert.ErrorIs(t, err, expected)
}
