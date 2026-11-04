package cron_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/faizan1250/taskify/internal/cron"
	"github.com/faizan1250/taskify/internal/server"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockJob struct {
	name        string
	description string
	ran         bool
	err         error
}

func (m *mockJob) Name() string {
	if m.name != "" {
		return m.name
	}
	return "test-job"
}

func (m *mockJob) Description() string {
	if m.description != "" {
		return m.description
	}
	return "test job description"
}

func (m *mockJob) Run(ctx context.Context, jobCtx *cron.JobContext) error {
	m.ran = true
	return m.err
}

func TestErrTaskIDConflict_Detection(t *testing.T) {
	// Verify that errors.Is properly recognizes Asynq ErrTaskIDConflict both directly and wrapped
	directErr := asynq.ErrTaskIDConflict
	assert.True(t, errors.Is(directErr, asynq.ErrTaskIDConflict))

	wrappedErr := fmt.Errorf("task already enqueued: %w", asynq.ErrTaskIDConflict)
	assert.True(t, errors.Is(wrappedErr, asynq.ErrTaskIDConflict))

	otherErr := errors.New("connection failed")
	assert.False(t, errors.Is(otherErr, asynq.ErrTaskIDConflict))
}

func TestJobRegistry(t *testing.T) {
	registry := cron.NewJobRegistry()
	jobs := registry.List()

	expectedJobs := []string{
		"due-date-reminders",
		"overdue-notifications",
		"weekly-reports",
		"auto-archive",
	}

	for _, expected := range expectedJobs {
		assert.Contains(t, jobs, expected)
		job, err := registry.Get(expected)
		require.NoError(t, err)
		assert.Equal(t, expected, job.Name())
		assert.NotEmpty(t, job.Description())
	}

	_, err := registry.Get("nonexistent-job")
	assert.Error(t, err)

	helpText := registry.Help()
	assert.Contains(t, helpText, "due-date-reminders")
	assert.Contains(t, helpText, "overdue-notifications")
}

func TestDistributedLockKeyNaming(t *testing.T) {
	tests := []struct {
		jobName     string
		expectedKey string
	}{
		{
			jobName:     "due-date-reminders",
			expectedKey: "cron:lock:due-date-reminders",
		},
		{
			jobName:     "overdue-notifications",
			expectedKey: "cron:lock:overdue-notifications",
		},
		{
			jobName:     "weekly-reports",
			expectedKey: "cron:lock:weekly-reports",
		},
		{
			jobName:     "auto-archive",
			expectedKey: "cron:lock:auto-archive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.jobName, func(t *testing.T) {
			assert.Equal(t, tt.expectedKey, cron.LockKeyForJob(tt.jobName))
		})
	}
}

func TestJobRunner_NilRedis_ExecutesJob(t *testing.T) {
	mock := &mockJob{name: "test-run"}
	runner := cron.NewMockJobRunner(mock, nil)

	err := runner.Run()
	assert.NoError(t, err)
	assert.True(t, mock.ran)
}

func TestJobRunner_JobError_Propagates(t *testing.T) {
	expectedErr := errors.New("job execution failed")
	mock := &mockJob{name: "test-err", err: expectedErr}
	runner := cron.NewMockJobRunner(mock, nil)

	err := runner.Run()
	assert.ErrorIs(t, err, expectedErr)
	assert.True(t, mock.ran)
}

func TestJobRunner_RedisUnreachable_ReturnsLockError(t *testing.T) {
	nopLogger := zerolog.Nop()
	// Point to unreachable port to verify lock acquisition failure handling
	deadRedis := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:54321",
	})
	srv := &server.Server{
		Logger: &nopLogger,
		Redis:  deadRedis,
	}

	mock := &mockJob{name: "test-redis-fail"}
	runner := cron.NewMockJobRunner(mock, srv)

	err := runner.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to acquire distributed lock")
	assert.False(t, mock.ran, "Job must not run if lock acquisition fails")
}
