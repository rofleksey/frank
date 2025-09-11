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
	Data   []byte           `json:"data"`
}

type Attachment struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type PromptData struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
	History     []PromptData `json:"history"`
}
