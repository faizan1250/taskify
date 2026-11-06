package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/faizan1250/taskify/internal/server"
	"github.com/rs/zerolog"
)

func NewMockJobRunner(job Job, srv *server.Server) *JobRunner {
	nopLogger := zerolog.Nop()
	if srv == nil {
		srv = &server.Server{
			Logger: &nopLogger,
		}
	} else if srv.Logger == nil {
		srv.Logger = &nopLogger
	}

	return &JobRunner{
		job: job,
		ctx: &JobContext{
			Server: srv,
		},
	}
}

func LockKeyForJob(jobName string) string {
	return fmt.Sprintf("cron:lock:%s", jobName)
}

// MaterializeDueRecurringForTest exposes the recurring-todos orchestration
// pass to package cron_test. Callers pass any stub implementing
// MaterializeRecurringTodos; the unexported interface never needs naming.
func MaterializeDueRecurringForTest(ctx context.Context, store recurringMaterializer, now time.Time, limit int) (int, error) {
	return materializeDueRecurring(ctx, store, now, limit)
}
