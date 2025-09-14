package telegram_bot

import (
	"context"
	"frank/app/dto"
	"strings"
)

func (s *Service) handleCancel(ctx context.Context) {
	s.replyService.Reply(ctx, "ОК")
}

func (s *Service) handleUnknownMessage(_ context.Context, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	s.reasonService.Handle(dto.NewPrompt(text))
}
