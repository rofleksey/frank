package reason

import (
	"context"
	"fmt"
	"frank/app/client/bothub"
	"frank/app/service/act"
	"frank/app/service/telegram_sender"
	"frank/pkg/config"
	"frank/pkg/database"
	"log/slog"
	"strings"
	"time"

	_ "embed"

	"github.com/samber/do"
)

var reasonTimeout = 5 * time.Minute

//go:embed SYSTEM_PROMPT
var systemPromptTemplate string

type Service struct {
	appCtx                context.Context
	cfg                   *config.Config
	queries               *database.Queries
	telegramSenderService *telegram_sender.Service
	actService            *act.Service
	bothubClient          *bothub.Client
}

func New(di *do.Injector) (*Service, error) {
	return &Service{
		appCtx:                do.MustInvoke[context.Context](di),
		cfg:                   do.MustInvoke[*config.Config](di),
		queries:               do.MustInvoke[*database.Queries](di),
		telegramSenderService: do.MustInvoke[*telegram_sender.Service](di),
		actService:            do.MustInvoke[*act.Service](di),
		bothubClient:          do.MustInvoke[*bothub.Client](di),
	}, nil
}

func (s *Service) HandlePrompt(text string) {
	go func() {
		ctx, cancel := context.WithTimeout(s.appCtx, reasonTimeout)
		defer cancel()

		slog.Info("Handling prompt",
			slog.String("text", text),
		)

		err := s.handlePromptImpl(ctx, text)
		if err != nil {
			slog.Error("Failed to handle prompt",
				slog.String("text", text),
				slog.Any("error", err),
			)

			if err := s.telegramSenderService.SendMessage(ctx, s.cfg.Telegram.ChatID, "Failed to handle prompt: "+err.Error()); err != nil {
				slog.Error("Failed to send message",
					slog.String("text", text),
					slog.Any("error", err),
				)
			}
		} else {
			slog.Info("Prompt handle success",
				slog.String("text", text),
			)
		}
	}()
}

func (s *Service) handlePromptImpl(ctx context.Context, text string) error {
	systemPrompt, err := s.generateSystemPrompt(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate system prompt: %w", err)
	}

	reasonOutput, err := s.bothubClient.Process(ctx, bothub.Prompt{
		SystemText: systemPrompt,
		UserText:   text,
	})
	if err != nil {
		return fmt.Errorf("gptClient.Process: %w", err)
	}

	reasonOutput = strings.TrimSpace(reasonOutput)
	reasonOutput = strings.TrimPrefix(reasonOutput, "```json")
	reasonOutput = strings.Trim(reasonOutput, "`")

	dataBytes := []byte(reasonOutput)
	if err = s.actService.Handle(ctx, dataBytes); err != nil {
		return fmt.Errorf("actService.Handle on %s: %w", reasonOutput, err)
	}

	return nil
}

func (s *Service) generateSystemPrompt(ctx context.Context) (string, error) {
	contextDescription, err := s.generateContextDescription(ctx)
	if err != nil {
		return "", fmt.Errorf("generateContextDescription: %w", err)
	}

	result := systemPromptTemplate

	result = strings.ReplaceAll(result, "{commands}", s.actService.CommandsDescription())
	result = strings.ReplaceAll(result, "{context}", contextDescription)

	return result, nil
}

func (s *Service) generateContextDescription(ctx context.Context) (string, error) {
	contextEntries, err := s.queries.ListContextEntries(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get context: %w", err)
	}

	var builder strings.Builder

	builder.WriteString("• Current time: ")
	builder.WriteString(time.Now().Format(time.RFC3339))
	builder.WriteString("\n")

	for _, entry := range contextEntries {
		builder.WriteString("• ")
		builder.WriteString(entry.Text)
		builder.WriteString("\n")
	}

	return builder.String(), nil
}
