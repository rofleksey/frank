package reason

import (
	"context"
	"fmt"
	"frank/app/client/bothub"
	"frank/app/dto"
	"frank/app/service/knowledge"
	"frank/app/service/telegram_reply"
	"frank/pkg/config"
	"frank/pkg/database"
	"log/slog"
	"strings"
	"time"

	_ "embed"

	"github.com/elliotchance/pie/v2"
	"github.com/samber/do"
)

var maxPromptDepth = 30
var reasonTimeout = 30 * time.Minute

//go:embed SYSTEM_PROMPT
var systemPromptTemplate string

type Actor interface {
	Handle(ctx context.Context, prompt dto.Prompt) (string, error)
	CommandsDescription() string
}

type Service struct {
	appCtx           context.Context
	cfg              *config.Config
	queries          *database.Queries
	replierService   *telegram_reply.Service
	bothubClient     *bothub.Client
	knowledgeService *knowledge.Service

	actor Actor
}

func New(di *do.Injector) (*Service, error) {
	return &Service{
		appCtx:           do.MustInvoke[context.Context](di),
		cfg:              do.MustInvoke[*config.Config](di),
		queries:          do.MustInvoke[*database.Queries](di),
		replierService:   do.MustInvoke[*telegram_reply.Service](di),
		knowledgeService: do.MustInvoke[*knowledge.Service](di),
		bothubClient:     do.MustInvoke[*bothub.Client](di),
	}, nil
}

func (s *Service) SetActor(actor Actor) {
	s.actor = actor
}

func (s *Service) Handle(prompt dto.Prompt) {
	if prompt.Depth > maxPromptDepth {
		slog.Error("Max prompt depth reached",
			slog.String("text", prompt.Text),
		)

		s.replierService.Reply(s.appCtx, "Failed to handle prompt: max prompt depth reached")

		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(s.appCtx, reasonTimeout)
		defer cancel()

		slog.Info("Handling prompt",
			slog.String("text", prompt.Text),
		)

		err := s.handlePromptImpl(ctx, prompt)
		if err != nil {
			slog.Error("Failed to handle prompt",
				slog.String("text", prompt.Text),
				slog.Any("error", err),
			)

			s.replierService.Reply(ctx, "Failed to handle prompt: "+err.Error())
		} else {
			slog.Info("Prompt handle success",
				slog.String("text", prompt.Text),
			)
		}
	}()
}

func (s *Service) handlePromptImpl(ctx context.Context, prompt dto.Prompt) error {
	systemPrompt, err := s.generateSystemPrompt(ctx, &prompt)
	if err != nil {
		return fmt.Errorf("failed to generate system prompt: %w", err)
	}

	userPrompt := prompt.Text + "\n\n" + s.generateAttachmentsDescription(&prompt)

	reasonOutput, err := s.bothubClient.Process(ctx, bothub.Prompt{
		SystemText: systemPrompt,
		UserText:   userPrompt,
	})
	if err != nil {
		return fmt.Errorf("gptClient.Process: %w", err)
	}

	reasonOutput = strings.TrimSpace(reasonOutput)
	reasonOutput = strings.TrimPrefix(reasonOutput, "```json")
	reasonOutput = strings.Trim(reasonOutput, "`")

	slog.Info("Got a result from bothub client",
		slog.String("text", prompt.Text),
		slog.Any("output", reasonOutput),
	)

	if _, err = s.actor.Handle(ctx, prompt.BranchWithNewText(reasonOutput)); err != nil {
		return fmt.Errorf("actService.Handle on '%s': %w", reasonOutput, err)
	}

	return nil
}

func (s *Service) generateSystemPrompt(ctx context.Context, prompt *dto.Prompt) (string, error) {
	contextDescription, err := s.generateContextDescription(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("generateContextDescription: %w", err)
	}

	result := systemPromptTemplate

	result = strings.ReplaceAll(result, "{commands}", s.actor.CommandsDescription())
	result = strings.ReplaceAll(result, "{context}", contextDescription)

	return result, nil
}

func (s *Service) generateContextDescription(ctx context.Context, prompt *dto.Prompt) (string, error) {
	contextEntries, err := s.knowledgeService.GetRelevant(ctx, *prompt)
	if err != nil {
		return "", fmt.Errorf("knowledgeService.GetRelevant: %w", err)
	}

	var builder strings.Builder

	builder.WriteString("- Current time: ")
	builder.WriteString(time.Now().Format(time.RFC3339))
	builder.WriteString("\n")

	builder.WriteString("- Available secrets: ")
	builder.WriteString(strings.Join(pie.Keys(s.cfg.Secrets), ", "))
	builder.WriteString("\n")

	for _, entry := range contextEntries {
		builder.WriteString("- ")
		builder.WriteString(entry)
		builder.WriteString("\n")
	}

	return builder.String(), nil
}

func (s *Service) generateAttachmentsDescription(prompt *dto.Prompt) string {
	if len(prompt.Attachments) == 0 {
		return ""
	}

	var builder strings.Builder

	builder.WriteString("# ATTACHMENTS\n\n")

	for _, att := range prompt.Attachments {
		builder.WriteString("## ")
		builder.WriteString(att.Name)
		builder.WriteString("\n")
		builder.WriteString(att.Content)
		builder.WriteString("\n\n")
	}

	return builder.String()
}
