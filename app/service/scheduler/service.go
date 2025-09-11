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
	Handle(ctx context.Context, prompt dto.Prompt) (string, error)
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

type scheduleOptions struct {
	destructOnCreateFail bool
	destructOnFinish     bool
}

func (s *Service) scheduleInternal(name string, jobDef gocron.JobDefinition, prompt dto.Prompt, opts scheduleOptions) error {
	destructor := func() {
		if err := s.queries.DeleteScheduledJob(s.appCtx, name); err != nil {
			slog.Error("Failed to delete scheduled job",
				slog.String("name", name),
				slog.Any("error", err),
			)
		}
	}

	_, err := s.scheduler.NewJob(
		jobDef,
		gocron.NewTask(
			func(ctx context.Context) {
				if _, err := s.actor.Handle(ctx, prompt); err != nil {
					slog.ErrorContext(ctx, "Failed to handle deferred command",
						slog.Any("prompt", prompt),
						slog.Any("error", err),
					)
				}
			},
		),
		gocron.WithName(name),
		gocron.WithEventListeners(
			gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
				slog.Error("Job success",
					slog.String("name", name),
				)

				if opts.destructOnFinish {
					destructor()
				}
			}),
			gocron.AfterJobRunsWithError(func(jobID uuid.UUID, jobName string, err error) {
				slog.Error("Job finished with error",
					slog.String("name", name),
					slog.Any("prompt", prompt),
					slog.Any("error", err),
				)

				if opts.destructOnFinish {
					destructor()
				}
			}),
			gocron.AfterJobRunsWithPanic(func(jobID uuid.UUID, jobName string, recoverData any) {
				slog.Error("Job panicked",
					slog.String("name", name),
					slog.Any("prompt", prompt),
					slog.Any("recoverData", recoverData),
				)

				if opts.destructOnFinish {
					destructor()
				}
			}),
		),
	)
	if err != nil {
		if opts.destructOnCreateFail {
			destructor()
		}

		return fmt.Errorf("scheduler.NewJob: %w", err)
	}

	return nil
}

func (s *Service) ScheduleOneTime(name string, fireAt time.Time, prompt dto.Prompt, opts ...dto.ScheduleOptions) error {
	var options dto.ScheduleOptions

	if len(opts) > 0 {
		options = opts[0]
	}

	if fireAt.Before(time.Now()) {
		fireAt = time.Now().Add(10 * time.Second)
	}

	if !options.SkipDBEntry {
		err := s.queries.CreateScheduledJob(s.appCtx, database.CreateScheduledJobParams{
			Name:    name,
			Created: time.Now(),
			Data: dto.ScheduledJobData{
				Type:   dto.OneTimeJobType,
				FireAt: fireAt,
				Cron:   "",
				Prompt: prompt,
			},
		})
		if err != nil {
			return fmt.Errorf("CreateScheduledJob: %w", err)
		}
	}

	if err := s.scheduleInternal(name, gocron.OneTimeJob(
		gocron.OneTimeJobStartDateTime(
			fireAt,
		),
	), prompt, scheduleOptions{
		destructOnCreateFail: true,
		destructOnFinish:     true,
	}); err != nil {
		return fmt.Errorf("scheduleInternal: %w", err)
	}

	return nil
}

func (s *Service) ScheduleCron(name, cron string, prompt dto.Prompt, opts ...dto.ScheduleOptions) error {
	var options dto.ScheduleOptions

	if len(opts) > 0 {
		options = opts[0]
	}

	if !options.SkipDBEntry {
		err := s.queries.CreateScheduledJob(s.appCtx, database.CreateScheduledJobParams{
			Name:    name,
			Created: time.Now(),
			Data: dto.ScheduledJobData{
				Type:   dto.CronJobType,
				FireAt: time.Time{},
				Cron:   cron,
				Prompt: prompt,
			},
		})
		if err != nil {
			return fmt.Errorf("CreateScheduledJob: %w", err)
		}
	}

	if err := s.scheduleInternal(name, gocron.CronJob(cron, false), prompt, scheduleOptions{
		destructOnCreateFail: true,
		destructOnFinish:     false,
	}); err != nil {
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
			if err = s.ScheduleOneTime(job.Name, job.Data.FireAt, job.Data.Prompt, dto.ScheduleOptions{SkipDBEntry: true}); err != nil {
				return fmt.Errorf("ScheduleOneTime: %w", err)
			}
		case dto.CronJobType:
			if err = s.ScheduleCron(job.Name, job.Data.Cron, job.Data.Prompt, dto.ScheduleOptions{SkipDBEntry: true}); err != nil {
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
