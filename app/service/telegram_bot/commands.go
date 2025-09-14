package telegram_bot

import (
	"context"
	"frank/app/dto"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (s *Service) handleCancel(ctx context.Context) {
	s.replyService.Reply(ctx, "ОК")
}

func (s *Service) handleUnknownMessage(ctx context.Context, messageID int, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	_, _ = s.tgBot.SetMessageReaction(ctx, &bot.SetMessageReactionParams{
		ChatID:    s.cfg.Telegram.ChatID,
		MessageID: messageID,
		Reaction: []models.ReactionType{
			{
				Type: models.ReactionTypeTypeEmoji,
				ReactionTypeEmoji: &models.ReactionTypeEmoji{
					Type:  models.ReactionTypeTypeEmoji,
					Emoji: "🗿",
				},
			},
		},
		IsBig: nil,
	})

	s.reasonService.Handle(dto.NewPrompt(text))
}
