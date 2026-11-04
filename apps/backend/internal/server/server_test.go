//nolint:testpackage // white-box testing needed for unexported httpServer field
package server

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShutdown_NilFields(t *testing.T) {
	t.Parallel()

	s := &Server{}
	err := s.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestShutdown_ClosesRedis(t *testing.T) {
	t.Parallel()

	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:0",
	})

	s := &Server{
		Redis: client,
	}

	ctx := context.Background()
	err := s.Shutdown(ctx)
	require.NoError(t, err)

	pingErr := s.Redis.Ping(ctx).Err()
	require.Error(t, pingErr)
	assert.ErrorIs(t, pingErr, redis.ErrClosed)
}

func TestShutdown_ErrorAggregation_HTTPFailureStillClosesRedis(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	handlerStarted := make(chan struct{})
	finishHandler := make(chan struct{})
	defer close(finishHandler)

	httpSrv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(handlerStarted)
			<-finishHandler
		}),
	}
	go func() {
		_ = httpSrv.Serve(ln)
	}()

	client := &http.Client{}
	go func() {
		resp, _ := client.Get("http://" + ln.Addr().String())
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	select {
	case <-handlerStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for HTTP handler to start")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:0",
	})

	s := &Server{
		httpServer: httpSrv,
		Redis:      redisClient,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err = s.Shutdown(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to shutdown HTTP server")
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	// Verify Redis cleanup was still executed even after HTTP shutdown error
	pingErr := s.Redis.Ping(context.Background()).Err()
	require.Error(t, pingErr)
	assert.ErrorIs(t, pingErr, redis.ErrClosed)
}

func TestShutdown_ErrorAggregation_MultipleErrors(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	handlerStarted := make(chan struct{})
	finishHandler := make(chan struct{})
	defer close(finishHandler)

	httpSrv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(handlerStarted)
			<-finishHandler
		}),
	}
	go func() {
		_ = httpSrv.Serve(ln)
	}()

	client := &http.Client{}
	go func() {
		resp, _ := client.Get("http://" + ln.Addr().String())
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	select {
	case <-handlerStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for HTTP handler to start")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:0",
	})
	// Pre-closing Redis causes s.Redis.Close() during Shutdown to return redis.ErrClosed
	require.NoError(t, redisClient.Close())

	s := &Server{
		httpServer: httpSrv,
		Redis:      redisClient,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err = s.Shutdown(ctx)
	require.Error(t, err)

	// Both errors must be aggregated
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.ErrorIs(t, err, redis.ErrClosed)
	assert.Contains(t, err.Error(), "failed to shutdown HTTP server")
	assert.Contains(t, err.Error(), "failed to close redis connection")
}
