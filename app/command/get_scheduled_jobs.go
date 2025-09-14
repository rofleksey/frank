package command

import (
	"context"
	"encoding/json"
	"fmt"
	"frank/app/dto"
	"log/slog"
	"strings"
)

type GetScheduledJobsCommand struct {
	scheduler Scheduler
}

func NewListScheduleCommand(scheduler Scheduler) *GetScheduledJobsCommand {
	return &GetScheduledJobsCommand{
		scheduler: scheduler,
	}
}

func (c *GetScheduledJobsCommand) Execute(_ context.Context, prompt dto.Prompt) (string, error) {
	slog.Info("Executing get_scheduled_jobs command",
		slog.String("text", prompt.Text),
	)

	jobs, err := c.scheduler.ListJobs()
	if err != nil {
		return "", fmt.Errorf("list scheduled jobs: %w", err)
	}

	result, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal scheduled jobs: %w", err)
	}

	return string(result), nil
}

func (c *GetScheduledJobsCommand) Name() string {
	return "get_scheduled_jobs"
}

func (c *GetScheduledJobsCommand) Description() string {
	return strings.TrimSpace(`
    type: object
    required:
      - command
    properties:
      command:
        type: string
        enum: 
          - get_scheduled_jobs
    description: returns a list of scheduled jobs. Returns result as a JSON string. This command will not display anything to the user, for this you MUST also use 'attach' and 'reply' commands.
  `)
}
