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
	"gopkg.in/yaml.v3"
)

//go:embed SYSTEM_PROMPT
var systemPromptTemplate string

//go:embed base_knowledge.yaml
var baseKnowledgeString string

type Service struct {
	appCtx        context.Context
	cfg           *config.Config
	queries       *database.Queries
	bothubClient  *bothub.Client
	knowledgeBase map[string]string
}

func New(di *do.Injector) (*Service, error) {
	cfg := do.MustInvoke[*config.Config](di)

	knowledgeBase := make(map[string]string)

	if err := yaml.Unmarshal([]byte(baseKnowledgeString), &knowledgeBase); err != nil {
		return nil, fmt.Errorf("yaml unmarshal: %w", err)
	}

	for name, content := range cfg.Knowledge {
		knowledgeBase[name] = content
	}

	return &Service{
		appCtx:        do.MustInvoke[context.Context](di),
		cfg:           cfg,
		queries:       do.MustInvoke[*database.Queries](di),
		bothubClient:  do.MustInvoke[*bothub.Client](di),
		knowledgeBase: knowledgeBase,
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

	for name, content := range s.knowledgeBase {
		if pie.Contains(reasonResult.Result, name) {
			result = append(result, content)
		}
	}

	return result, nil
}

func (s *Service) generateSystemPrompt() string {
	names := make([]string, 0)

	for name := range s.knowledgeBase {
		names = append(names, name)
	}

	result := systemPromptTemplate

	result = strings.ReplaceAll(result, "{names}", strings.Join(names, ", "))

	return result
}
