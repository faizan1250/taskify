package job

import (
	"context"

	"github.com/faizan1250/taskify/internal/config"
	"github.com/faizan1250/taskify/internal/database"
	"github.com/faizan1250/taskify/internal/lib/email"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
)

type JobService struct {
	Client      *asynq.Client
	server      *asynq.Server
	logger      *zerolog.Logger
	authService AuthServiceInterface
	emailClient *email.Client
}

type AuthServiceInterface interface {
	GetUserEmail(ctx context.Context, userID string) (string, error)
}

func NewJobService(logger *zerolog.Logger, cfg *config.Config) *JobService {
	_, asynqOpt, err := database.ParseRedisConfig(cfg.Redis.Address, cfg.Redis.Password)
	if err != nil {
		logger.Error().Err(err).Msg("failed to parse redis configuration for jobs")
		asynqOpt = asynq.RedisClientOpt{Addr: cfg.Redis.Address, Password: cfg.Redis.Password}
	}

	client := asynq.NewClient(asynqOpt)

	server := asynq.NewServer(
		asynqOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6, // Higher priority queue for important emails
				"default":  3, // Default priority for most emails
				"low":      1, // Lower priority for non-urgent emails
			},
		},
	)

	return &JobService{
		Client: client,
		server: server,
		logger: logger,
	}
}

func (j *JobService) SetAuthService(authService AuthServiceInterface) {
	j.authService = authService
}

func (j *JobService) Start() error {
	// Register task handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskWelcome, j.handleWelcomeEmailTask)
	mux.HandleFunc(TaskReminderEmail, j.handleReminderEmailTask)
	mux.HandleFunc(TaskWeeklyReportEmail, j.handleWeeklyReportEmailTask)

	j.logger.Info().Msg("Starting background job server")
	if err := j.server.Start(mux); err != nil {
		return err
	}

	return nil
}

func (j *JobService) Stop() {
	j.logger.Info().Msg("Stopping background job server")
	j.server.Shutdown()
	j.Client.Close()
}
