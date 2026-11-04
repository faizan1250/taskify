package cron

import (
	"fmt"

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
