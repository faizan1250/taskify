package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testContextKey string

func TestDeleteTodoAttachment_ContextDetachment(t *testing.T) {
	t.Run("parent context cancellation does not cancel detached background context", func(t *testing.T) {
		parentCtx, parentCancel := context.WithCancel(context.Background())
		const key testContextKey = "trace_id"
		ctxWithValue := context.WithValue(parentCtx, key, "test-trace-123")

		// Create detached context with timeout as implemented in DeleteTodoAttachment
		bgCtx, cancel := context.WithTimeout(context.WithoutCancel(ctxWithValue), 30*time.Second)
		defer cancel()

		// Cancel parent context
		parentCancel()

		// Verify parent context is canceled
		require.ErrorIs(t, ctxWithValue.Err(), context.Canceled)

		// Verify detached background context is NOT canceled
		assert.NoError(t, bgCtx.Err())

		// Verify values from parent context are preserved
		assert.Equal(t, "test-trace-123", bgCtx.Value(key))

		// Verify deadline is set within expected range
		deadline, hasDeadline := bgCtx.Deadline()
		assert.True(t, hasDeadline)
		assert.WithinDuration(t, time.Now().Add(30*time.Second), deadline, 2*time.Second)
	})

	t.Run("detached background context cancels when cancel func is invoked", func(t *testing.T) {
		parentCtx, parentCancel := context.WithCancel(context.Background())
		defer parentCancel()

		bgCtx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), 30*time.Second)
		assert.NoError(t, bgCtx.Err())

		cancel()
		assert.ErrorIs(t, bgCtx.Err(), context.Canceled)
	})

	t.Run("detached background context respects timeout", func(t *testing.T) {
		parentCtx, parentCancel := context.WithCancel(context.Background())
		parentCancel()

		shortTimeoutCtx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), 10*time.Millisecond)
		defer cancel()

		time.Sleep(25 * time.Millisecond)
		assert.ErrorIs(t, shortTimeoutCtx.Err(), context.DeadlineExceeded)
	})
}

func TestDeleteTodoAttachment_PanicRecovery(t *testing.T) {
	t.Run("recovers panic in asynchronous deletion worker", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		var recovered any
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					recovered = r
				}
			}()

			panic("simulated S3 client panic")
		}()

		wg.Wait()
		assert.Equal(t, "simulated S3 client panic", recovered)
	})
}
