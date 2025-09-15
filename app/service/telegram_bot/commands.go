package telegram_bot

import (
	"context"
	"strings"
)

func (s *Service) handleCancel(_ context.Context) {
	s.promptManager.CancelAll()
}

func (s *Service) handleUnknownMessage(_ context.Context, messageID int, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	newPrompt := s.promptManager.CreatePrompt(messageID, text)
	s.reasonService.Handle(newPrompt)
}
