package command

import (
	"context"
	"time"
)

type MessageSender interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type Actor interface {
	HandleMessage(ctx context.Context, dataBytes []byte) error
}

type Scheduler interface {
	ScheduleOneTime(fireAt time.Time, data []byte) error
	ScheduleCron(cron string, data []byte) error
}
