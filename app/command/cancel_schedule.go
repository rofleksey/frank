package command

import (
	"context"
	"encoding/json"
	"fmt"
	"frank/app/dto"
	"log/slog"
	"strings"
)

type CancelScheduleCommand struct {
	replier   Replier
	scheduler Scheduler
}

func NewCancelScheduleCommand(replier Replier, scheduler Scheduler) *CancelScheduleCommand {
	return &CancelScheduleCommand{
		replier:   replier,
		scheduler: scheduler,
	}
}

func (c *CancelScheduleCommand) Execute(ctx context.Context, prompt dto.Prompt) (string, error) {
	slog.Info("Executing cancel_schedule command",
		slog.String("text", prompt.Text),
	)

	var data CancelScheduleCommandData

	if err := json.Unmarshal([]byte(prompt.Text), &data); err != nil {
		return "", fmt.Errorf("json unmarshal: %w", err)
	}

	if data.Name == "" {
		return "", fmt.Errorf("schedule command name is empty")
	}

	if err := c.scheduler.CancelJob(data.Name); err != nil {
		return "", fmt.Errorf("cancel job: %w", err)
	}

	c.replier.Reply(ctx, fmt.Sprintf("Job '%s' was cancelled", data.Name))

	return "", nil
}

func (c *CancelScheduleCommand) Name() string {
	return "cancel_schedule"
}

func (c *CancelScheduleCommand) Description() string {
	return strings.TrimSpace(`
    type: object
    required:
      - command
    properties:
      command:
        type: string
        enum: 
          - cancel_schedule
      name:
        type: string
        description: scheduled job name
    description: cancels a scheduled job by it's name
  `)
}
