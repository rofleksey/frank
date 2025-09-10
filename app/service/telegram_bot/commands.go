package telegram_bot

import (
	"context"
	"strings"
)

func (s *Service) handleCancel(ctx context.Context) {
	_ = s.tgSenderService.SendMessage(ctx, s.cfg.Telegram.ChatID, "ОК")
}

func (s *Service) handleUnknownMessage(ctx context.Context, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	s.reasonService.HandlePrompt(text)
}
