package dto

import "time"

type ScheduledJobType string

var OneTimeJobType ScheduledJobType = "one-time"
var CronJobType ScheduledJobType = "cron"

type ScheduleOptions struct {
	SkipDBEntry bool
}

type ScheduledJobData struct {
	Type   ScheduledJobType `json:"type"`
	FireAt time.Time        `json:"fire_at"`
	Cron   string           `json:"cron"`
	Prompt Prompt           `json:"prompt"`
}
