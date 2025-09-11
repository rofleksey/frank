package command

import (
	"context"
	"encoding/json"
	"fmt"
	"frank/app/dto"
	"log/slog"
	"strings"
	"time"
)

type ScheduleCommand struct {
	chatID    int64
	sender    MessageSender
	scheduler Scheduler
}

func NewScheduleCommand(chatID int64, sender MessageSender, scheduler Scheduler) *ScheduleCommand {
	return &ScheduleCommand{
		chatID:    chatID,
		sender:    sender,
		scheduler: scheduler,
	}
}

type ScheduleCommandData struct {
	Name             string          `json:"name"`
	Type             string          `json:"type"`
	Time             string          `json:"time"`
	ScheduledCommand json.RawMessage `json:"scheduled_command"`
}

func (c *ScheduleCommand) Execute(ctx context.Context, prompt dto.Prompt) (string, error) {
	slog.Info("Executing schedule command",
		slog.Any("prompt", prompt),
	)

	var data ScheduleCommandData

	if err := json.Unmarshal([]byte(prompt.Text), &data); err != nil {
		return "", fmt.Errorf("json unmarshal: %w", err)
	}

	if data.Name == "" {
		return "", fmt.Errorf("schedule command name is empty")
	}

	switch data.Type {
	case "cron":
		if err := c.scheduler.ScheduleCron(data.Name, data.Time, prompt.BranchWithNewText(string(data.ScheduledCommand))); err != nil {
			return "", fmt.Errorf("ScheduleCron: %w", err)
		}

		if err := c.sender.SendMessage(ctx, c.chatID, "Scheduled a cron job: "+data.Time); err != nil {
			return "", fmt.Errorf("send message: %w", err)
		}
	case "one-time":
		actualTime, err := time.Parse(time.RFC3339, data.Time)
		if err != nil {
			return "", fmt.Errorf("parse time: %w", err)
		}

		if err := c.scheduler.ScheduleOneTime(data.Name, actualTime, prompt.BranchWithNewText(string(data.ScheduledCommand))); err != nil {
			return "", fmt.Errorf("ScheduleOneTime: %w", err)
		}

		if err := c.sender.SendMessage(ctx, c.chatID, "Scheduled a one time job at "+data.Time); err != nil {
			return "", fmt.Errorf("send message: %w", err)
		}
	default:
		return "", fmt.Errorf("unknown schedule type: %s", data.Type)
	}

	return "", nil
}

func (c *ScheduleCommand) Name() string {
	return "schedule"
}

func (c *ScheduleCommand) Description() string {
	return strings.TrimSpace(`
    type: object
    required:
      - command
      - name
      - type
      - time
      - scheduled_command
    properties:
      command:
        type: string
        enum: 
          - schedule
      name:
        type: string
        description: Short but meaningful name of the scheduled job, alphanumerical, snake-case
      type:
        type: string
        enum: [cron, one-time]
        description: Type of schedule - "cron" for recurring schedules or "one-time" for single execution
      time:
        type: string
        description: |
          Schedule time value. For "cron" type, this should be a valid cron expression (without seconds, DAYS START FROM ZERO (!!!) [0-6]).
          For "one-time" type, this should be an ISO 8601 formatted datetime string.
        example: "0 0 * * *"  # for cron type
        # example: "2023-12-25T10:30:00Z"  # for one-time type
      scheduled_command:
        <JSON of the command to schedule>
    description: schedule a recurring or one-time command 
  `)
}
