package job_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/faizan1250/taskify/internal/errs"
	"github.com/faizan1250/taskify/internal/lib/job"
	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateReminderTaskID_DueDateReminder(t *testing.T) {
	todoID := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
	dueDate := time.Date(2026, 11, 10, 14, 30, 0, 0, time.UTC)

	task1 := &job.ReminderEmailTask{
		TodoID:   todoID,
		DueDate:  dueDate,
		TaskType: "due_date_reminder",
	}

	taskID1 := job.GenerateReminderTaskID(task1)
	expected := "due_date_reminder:12345678-1234-1234-1234-123456789abc:2026-11-10T14:30:00Z"
	assert.Equal(t, expected, taskID1)

	// Same todo and same date but different time of day must produce a distinct TaskID
	dueDateLater := time.Date(2026, 11, 10, 23, 59, 59, 0, time.UTC)
	task2 := &job.ReminderEmailTask{
		TodoID:   todoID,
		DueDate:  dueDateLater,
		TaskType: "due_date_reminder",
	}
	assert.NotEqual(t, taskID1, job.GenerateReminderTaskID(task2))

	// Different date produces different TaskID
	dueDateDifferent := time.Date(2026, 11, 11, 14, 30, 0, 0, time.UTC)
	taskDifferentDate := &job.ReminderEmailTask{
		TodoID:   todoID,
		DueDate:  dueDateDifferent,
		TaskType: "due_date_reminder",
	}
	assert.NotEqual(t, taskID1, job.GenerateReminderTaskID(taskDifferentDate))

	// Different TodoID produces different TaskID
	otherTodoID := uuid.MustParse("87654321-4321-4321-4321-cba987654321")
	taskDifferentTodo := &job.ReminderEmailTask{
		TodoID:   otherTodoID,
		DueDate:  dueDate,
		TaskType: "due_date_reminder",
	}
	assert.NotEqual(t, taskID1, job.GenerateReminderTaskID(taskDifferentTodo))
}

func TestGenerateReminderTaskID_OverdueNotification(t *testing.T) {
	todoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	dueDate := time.Date(2026, 11, 10, 14, 30, 0, 0, time.UTC)

	task := &job.ReminderEmailTask{
		TodoID:   todoID,
		DueDate:  dueDate,
		TaskType: "overdue_notification",
	}

	taskID := job.GenerateReminderTaskID(task)
	expected := "overdue_notification:aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee:2026-11-10T14:30:00Z"
	assert.Equal(t, expected, taskID)
}

func TestGenerateReminderTaskID_CustomType(t *testing.T) {
	todoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")

	task := &job.ReminderEmailTask{
		TodoID:   todoID,
		TaskType: "custom_type",
	}

	taskID := job.GenerateReminderTaskID(task)
	prefix := "custom_type:aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee:"
	require.True(t, len(taskID) > len(prefix), "task ID %q missing timestamp suffix", taskID)
	assert.Equal(t, prefix, taskID[:len(prefix)])
	_, err := time.Parse(time.RFC3339, taskID[len(prefix):])
	assert.NoError(t, err, "timestamp suffix must be full RFC3339 granularity")
}

type mockTodoRepo struct {
	status string
	err    error

	markReminderErr error
	markOverdueErr  error
	reminderMarked  []uuid.UUID
	overdueMarked   []uuid.UUID
}

func (m *mockTodoRepo) GetTodoStatus(ctx context.Context, id uuid.UUID) (string, error) {
	return m.status, m.err
}

func (m *mockTodoRepo) MarkDueDateReminderSent(ctx context.Context, todoID uuid.UUID) error {
	if m.markReminderErr != nil {
		return m.markReminderErr
	}
	m.reminderMarked = append(m.reminderMarked, todoID)
	return nil
}

func (m *mockTodoRepo) MarkOverdueNotificationSent(ctx context.Context, todoID uuid.UUID) error {
	if m.markOverdueErr != nil {
		return m.markOverdueErr
	}
	m.overdueMarked = append(m.overdueMarked, todoID)
	return nil
}

type mockAuthService struct {
	email string
	err   error
}

func (m *mockAuthService) GetUserEmail(ctx context.Context, userID string) (string, error) {
	return m.email, m.err
}

func TestHandleReminderEmailTask_FreshnessGuard(t *testing.T) {
	logger := zerolog.Nop()
	todoID := uuid.New()

	tests := []struct {
		name       string
		status     string
		repoErr    error
		wantErr    bool
		errContain string
	}{
		{
			name:    "completed todo is skipped cleanly",
			status:  string(todo.StatusCompleted),
			wantErr: false,
		},
		{
			name:    "archived todo is skipped cleanly",
			status:  string(todo.StatusArchived),
			wantErr: false,
		},
		{
			name:    "deleted todo (404) is skipped cleanly",
			repoErr: errs.NewNotFoundError("todo not found", false, nil),
			wantErr: false,
		},
		{
			name:       "database error is propagated",
			repoErr:    errors.New("db connection failure"),
			wantErr:    true,
			errContain: "failed to check todo status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &job.JobService{}
			svc.SetLogger(&logger)
			repo := &mockTodoRepo{
				status: tt.status,
				err:    tt.repoErr,
			}
			svc.SetTodoRepository(repo)
			svc.SetAuthService(&mockAuthService{
				email: "user@example.com",
			})

			payload, err := json.Marshal(job.ReminderEmailTask{
				UserID:    "user-123",
				TodoID:    todoID,
				TodoTitle: "Test task",
				TaskType:  "due_date_reminder",
				DueDate:   time.Now().Add(24 * time.Hour),
			})
			require.NoError(t, err)

			asynqTask := asynq.NewTask(job.TaskReminderEmail, payload)

			err = svc.HandleReminderEmailTask(context.Background(), asynqTask)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContain != "" {
					assert.Contains(t, err.Error(), tt.errContain)
				}
			} else {
				assert.NoError(t, err)
			}
			// Nothing was delivered on these paths, so nothing may be marked sent.
			assert.Empty(t, repo.reminderMarked, "due-date reminder must not be marked sent without delivery")
			assert.Empty(t, repo.overdueMarked, "overdue notification must not be marked sent without delivery")
		})
	}
}

func TestHandleReminderEmailTask_NilDependencies(t *testing.T) {
	logger := zerolog.Nop()
	todoID := uuid.New()

	newReminderTask := func() *asynq.Task {
		payload, err := json.Marshal(job.ReminderEmailTask{
			UserID:    "user-123",
			TodoID:    todoID,
			TodoTitle: "Test task",
			TaskType:  "due_date_reminder",
			DueDate:   time.Now().Add(24 * time.Hour),
		})
		require.NoError(t, err)
		return asynq.NewTask(job.TaskReminderEmail, payload)
	}

	t.Run("missing auth service returns error", func(t *testing.T) {
		svc := &job.JobService{}
		svc.SetLogger(&logger)

		err := svc.HandleReminderEmailTask(context.Background(), newReminderTask())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "auth service is not configured")
	})

	t.Run("missing email client returns error", func(t *testing.T) {
		svc := &job.JobService{}
		svc.SetLogger(&logger)
		svc.SetAuthService(&mockAuthService{email: "user@example.com"})

		err := svc.HandleReminderEmailTask(context.Background(), newReminderTask())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "email client is not configured")
	})
}

func TestHandleWelcomeEmailTask_NilEmailClient(t *testing.T) {
	logger := zerolog.Nop()
	svc := &job.JobService{}
	svc.SetLogger(&logger)

	payload, err := json.Marshal(map[string]string{
		"to":         "user@example.com",
		"first_name": "Test",
	})
	require.NoError(t, err)

	err = svc.HandleWelcomeEmailTask(context.Background(), asynq.NewTask(job.TaskWelcome, payload))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email client is not configured")
}

func TestHandleWeeklyReportEmailTask_NilDependencies(t *testing.T) {
	logger := zerolog.Nop()

	newReportTask := func() *asynq.Task {
		payload, err := json.Marshal(job.WeeklyReportEmailTask{
			UserID:    "user-123",
			WeekStart: time.Now().Add(-7 * 24 * time.Hour),
			WeekEnd:   time.Now(),
		})
		require.NoError(t, err)
		return asynq.NewTask(job.TaskWeeklyReportEmail, payload)
	}

	t.Run("missing auth service returns error", func(t *testing.T) {
		svc := &job.JobService{}
		svc.SetLogger(&logger)

		err := svc.HandleWeeklyReportEmailTask(context.Background(), newReportTask())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "auth service is not configured")
	})

	t.Run("missing email client returns error", func(t *testing.T) {
		svc := &job.JobService{}
		svc.SetLogger(&logger)
		svc.SetAuthService(&mockAuthService{email: "user@example.com"})

		err := svc.HandleWeeklyReportEmailTask(context.Background(), newReportTask())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "email client is not configured")
	})
}
