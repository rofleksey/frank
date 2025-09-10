package scheduler

import (
	"context"
	"fmt"
	"frank/app/dto"
	"frank/pkg/config"
	"frank/pkg/database"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/samber/do"
)

var _ do.Shutdownable = (*Service)(nil)

type Actor interface {
	Handle(ctx context.Context, data []byte) error
}

type Service struct {
	appCtx  context.Context
	cfg     *config.Config
	queries *database.Queries

	actor     Actor
	scheduler gocron.Scheduler
}

func New(di *do.Injector) (*Service, error) {
	appCtx := do.MustInvoke[context.Context](di)
	cfg := do.MustInvoke[*config.Config](di)

	scheduler, err := gocron.NewScheduler(
		gocron.WithGlobalJobOptions(
			gocron.WithContext(appCtx),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("gocron.NewScheduler: %w", err)
	}

	return &Service{
		appCtx:    appCtx,
		cfg:       cfg,
		queries:   do.MustInvoke[*database.Queries](di),
		scheduler: scheduler,
	}, nil
}

func (s *Service) SetActor(actor Actor) {
	s.actor = actor
}

func (s *Service) scheduleInternal(id int64, jobDef gocron.JobDefinition, data []byte) error {
	jobName := fmt.Sprintf("job-%d", id)

	destructor := func() {
		if err := s.queries.DeleteScheduledJob(s.appCtx, id); err != nil {
			slog.Error("Failed to delete scheduled job",
				slog.Int64("job_id", id),
				slog.Any("error", err),
			)
		}
	}

	_, err := s.scheduler.NewJob(
		jobDef,
		gocron.NewTask(
			func(ctx context.Context) {
				if err := s.actor.Handle(ctx, data); err != nil {
					slog.ErrorContext(ctx, "Failed to handle deferred command",
						slog.String("data", string(data)),
						slog.Any("error", err),
					)
				}
			},
		),
		gocron.WithName(jobName),
		gocron.WithEventListeners(
			gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
				slog.Error("Job success",
					slog.Int64("job_id", id),
				)

				destructor()
			}),
			gocron.AfterJobRunsWithError(func(jobID uuid.UUID, jobName string, err error) {
				slog.Error("Job finished with error",
					slog.Int64("job_id", id),
					slog.String("data", string(data)),
					slog.Any("error", err),
				)

				destructor()
			}),
			gocron.AfterJobRunsWithPanic(func(jobID uuid.UUID, jobName string, recoverData any) {
				slog.Error("Job panicked",
					slog.Int64("job_id", id),
					slog.String("data", string(data)),
					slog.Any("recoverData", recoverData),
				)

				destructor()
			}),
		),
	)
	if err != nil {
		destructor()

		return fmt.Errorf("scheduler.NewJob: %w", err)
	}

	return nil
}

func (s *Service) ScheduleOneTime(fireAt time.Time, data []byte) error {
	if fireAt.Before(time.Now()) {
		fireAt = time.Now().Add(10 * time.Second)
	}

	id, err := s.queries.CreateScheduledJob(s.appCtx, dto.ScheduledJobData{
		Type:   dto.OneTimeJobType,
		FireAt: fireAt,
		Cron:   "",
		Data:   data,
	})
	if err != nil {
		return fmt.Errorf("CreateScheduledJob: %w", err)
	}

	if err = s.scheduleInternal(id, gocron.OneTimeJob(
		gocron.OneTimeJobStartDateTime(
			fireAt,
		),
	), data); err != nil {
		return fmt.Errorf("scheduleInternal: %w", err)
	}

	return nil
}

func (s *Service) ScheduleCron(cron string, data []byte) error {
	id, err := s.queries.CreateScheduledJob(s.appCtx, dto.ScheduledJobData{
		Type:   dto.CronJobType,
		FireAt: time.Time{},
		Cron:   cron,
		Data:   data,
	})
	if err != nil {
		return fmt.Errorf("CreateScheduledJob: %w", err)
	}

	if err = s.scheduleInternal(id, gocron.CronJob(cron, false), data); err != nil {
		return fmt.Errorf("scheduleInternal: %w", err)
	}

	return nil
}

func (s *Service) Start() error {
	jobs, err := s.queries.ListScheduledJobs(s.appCtx)
	if err != nil {
		return fmt.Errorf("ListScheduledJobs: %w", err)
	}

	for _, job := range jobs {
		switch job.Data.Type {
		case dto.OneTimeJobType:
			if err = s.ScheduleOneTime(job.Data.FireAt, job.Data.Data); err != nil {
				return fmt.Errorf("ScheduleOneTime: %w", err)
			}
		case dto.CronJobType:
			if err = s.ScheduleCron(job.Data.Cron, job.Data.Data); err != nil {
				return fmt.Errorf("ScheduleCron: %w", err)
			}
		default:
			return fmt.Errorf("unknown job type: %s", job.Data.Type)
		}
	}

	s.scheduler.Start()

	return nil
}

func (s *Service) Shutdown() error {
	_ = s.scheduler.Shutdown()

	return nil
}
