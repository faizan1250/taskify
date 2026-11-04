package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/faizan1250/taskify/internal/config"
	"github.com/faizan1250/taskify/internal/database"
	"github.com/faizan1250/taskify/internal/logger"
	"github.com/faizan1250/taskify/internal/repository"
	"github.com/faizan1250/taskify/internal/server"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type JobContext struct {
	Config        *config.Config
	Server        *server.Server
	JobClient     *asynq.Client
	Repositories  *repository.Repositories
	LoggerService *logger.LoggerService
}

func NewJobContext() (*JobContext, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	loggerService := logger.NewLoggerService(cfg.Observability)
	loggerInstance := logger.NewLoggerWithService(cfg.Observability, loggerService)

	db, err := database.New(cfg, &loggerInstance, loggerService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	redisOptions, _, err := database.ParseRedisConfig(cfg.Redis.Address, cfg.Redis.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis configuration: %w", err)
	}
	redisClient := redis.NewClient(redisOptions)

	srv := &server.Server{
		Config:        cfg,
		Logger:        &loggerInstance,
		LoggerService: loggerService,
		DB:            db,
		Redis:         redisClient,
	}

	jobClient, err := initJobClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize job client: %w", err)
	}

	repositories := repository.NewRepositories(srv)

	return &JobContext{
		Config:        cfg,
		Server:        srv,
		JobClient:     jobClient,
		Repositories:  repositories,
		LoggerService: loggerService,
	}, nil
}

func (c *JobContext) Close() {
	if c.Server != nil && c.Server.DB != nil {
		c.Server.DB.Pool.Close()
	}
	if c.Server != nil && c.Server.Redis != nil {
		_ = c.Server.Redis.Close()
	}
	if c.JobClient != nil {
		_ = c.JobClient.Close()
	}
	if c.LoggerService != nil {
		c.LoggerService.Shutdown()
	}
}

func initJobClient(cfg *config.Config) (*asynq.Client, error) {
	_, asynqOpt, err := database.ParseRedisConfig(cfg.Redis.Address, cfg.Redis.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis configuration for jobs: %w", err)
	}

	client := asynq.NewClient(asynqOpt)
	return client, nil
}

type Job interface {
	Name() string
	Description() string
	Run(ctx context.Context, jobCtx *JobContext) error
}

type JobRunner struct {
	job Job
	ctx *JobContext
}

func NewJobRunner(job Job) (*JobRunner, error) {
	ctx, err := NewJobContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create job context: %w", err)
	}

	return &JobRunner{
		job: job,
		ctx: ctx,
	}, nil
}

func (r *JobRunner) Run() error {
	defer r.ctx.Close()

	ctx := context.Background()

	if r.ctx.Server.Redis != nil {
		lockKey := fmt.Sprintf("cron:lock:%s", r.job.Name())
		locked, err := r.ctx.Server.Redis.SetNX(ctx, lockKey, "locked", 5*time.Minute).Result()
		if err != nil {
			r.ctx.Server.Logger.Error().
				Err(err).
				Str("job", r.job.Name()).
				Msg("Failed to check/acquire distributed lock")
			return fmt.Errorf("failed to acquire distributed lock: %w", err)
		}
		if !locked {
			r.ctx.Server.Logger.Info().
				Str("job", r.job.Name()).
				Msg("Cron job is already running on another instance, skipping execution")
			return nil
		}
		defer func() {
			_ = r.ctx.Server.Redis.Del(context.Background(), lockKey).Err()
		}()
	}

	r.ctx.Server.Logger.Info().
		Str("job", r.job.Name()).
		Msg("Starting cron job")

	err := r.job.Run(ctx, r.ctx)
	if err != nil {
		r.ctx.Server.Logger.Error().
			Err(err).
			Str("job", r.job.Name()).
			Msg("Failed to run cron job")
		return err
	}

	r.ctx.Server.Logger.Info().
		Str("job", r.job.Name()).
		Msg("Cron job completed successfully")
	return nil
}
