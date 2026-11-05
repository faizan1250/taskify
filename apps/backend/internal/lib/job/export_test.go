package job

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
)

func (j *JobService) SetLogger(logger *zerolog.Logger) {
	j.logger = logger
}

func (j *JobService) HandleReminderEmailTask(ctx context.Context, t *asynq.Task) error {
	return j.handleReminderEmailTask(ctx, t)
}

func (j *JobService) HandleWelcomeEmailTask(ctx context.Context, t *asynq.Task) error {
	return j.handleWelcomeEmailTask(ctx, t)
}

func (j *JobService) HandleWeeklyReportEmailTask(ctx context.Context, t *asynq.Task) error {
	return j.handleWeeklyReportEmailTask(ctx, t)
}
