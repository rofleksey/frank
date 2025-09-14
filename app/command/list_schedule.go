package command

import (
	"context"
	"encoding/json"
	"fmt"
	"frank/app/dto"
	"log/slog"
	"strings"
)

type ListScheduleCommand struct {
	scheduler Scheduler
}

func NewListScheduleCommand(scheduler Scheduler) *ListScheduleCommand {
	return &ListScheduleCommand{
		scheduler: scheduler,
	}
}

type CancelScheduleCommandData struct {
	Name string `json:"name"`
}

func (c *ListScheduleCommand) Execute(_ context.Context, prompt dto.Prompt) (string, error) {
	slog.Info("Executing list_schedule command",
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

func (c *ListScheduleCommand) Name() string {
	return "list_schedule"
}

func (c *ListScheduleCommand) Description() string {
	return strings.TrimSpace(`
    type: object
    required:
      - command
    properties:
      command:
        type: string
        enum: 
          - list_schedule
    description: returns a list of scheduled jobs. Returns result as a JSON string, will NOT display it to the user.
  `)
}
