package command

import (
	"context"
	"frank/app/dto"
	"time"
)

type MessageSender interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type Actor interface {
	Handle(ctx context.Context, prompt dto.Prompt) error
}

type Scheduler interface {
	ScheduleOneTime(name string, fireAt time.Time, prompt dto.Prompt, opts ...dto.ScheduleOptions) error
	ScheduleCron(name, cron string, prompt dto.Prompt, opts ...dto.ScheduleOptions) error
}
