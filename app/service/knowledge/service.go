package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"frank/app/client/bothub"
	"frank/app/dto"
	"frank/pkg/config"
	"frank/pkg/database"
	"log/slog"
	"strings"

	_ "embed"

	"github.com/elliotchance/pie/v2"
	"github.com/samber/do"
)

//go:embed SYSTEM_PROMPT
var systemPromptTemplate string

type Service struct {
	appCtx       context.Context
	cfg          *config.Config
	queries      *database.Queries
	bothubClient *bothub.Client
}

func New(di *do.Injector) (*Service, error) {
	return &Service{
		appCtx:       do.MustInvoke[context.Context](di),
		cfg:          do.MustInvoke[*config.Config](di),
		queries:      do.MustInvoke[*database.Queries](di),
		bothubClient: do.MustInvoke[*bothub.Client](di),
	}, nil
}

type ReasonResult struct {
	Result []string `json:"result"`
}

func (s *Service) GetRelevant(ctx context.Context, prompt dto.Prompt) ([]string, error) {
	systemPrompt := s.generateSystemPrompt()

	reasonOutput, err := s.bothubClient.Process(ctx, bothub.Prompt{
		SystemText: systemPrompt,
		UserText:   prompt.Text,
		Model:      bothub.ModelDeepseekChatV3,
	})
	if err != nil {
		return nil, fmt.Errorf("gptClient.Process: %w", err)
	}

	reasonOutput = strings.TrimSpace(reasonOutput)
	reasonOutput = strings.TrimPrefix(reasonOutput, "```json")
	reasonOutput = strings.Trim(reasonOutput, "`")

	slog.Info("Got a result from bothub client",
		slog.String("text", prompt.Text),
		slog.Any("output", reasonOutput),
	)

	var reasonResult ReasonResult

	if err = json.Unmarshal([]byte(reasonOutput), &reasonResult); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	result := make([]string, 0)

outer:
	for _, entry := range s.cfg.Knowledge {
		for _, entryTag := range entry.Tags {
			if pie.Contains(reasonResult.Result, entryTag) {
				result = append(result, entry.Content)
				continue outer
			}
		}
	}

	return result, nil
}

func (s *Service) generateSystemPrompt() string {
	tagMap := make(map[string]struct{})

	for _, entry := range s.cfg.Knowledge {
		for _, tag := range entry.Tags {
			tagMap[tag] = struct{}{}
		}
	}

	tags := pie.Keys(tagMap)

	result := systemPromptTemplate

	result = strings.ReplaceAll(result, "{tags}", strings.Join(tags, ", "))

	return result
}
