package secret

import (
	"frank/pkg/config"
	"frank/pkg/database"
	"strings"

	"github.com/samber/do"
)

type Service struct {
	cfg     *config.Config
	queries *database.Queries
}

func New(di *do.Injector) (*Service, error) {
	return &Service{
		cfg:     do.MustInvoke[*config.Config](di),
		queries: do.MustInvoke[*database.Queries](di),
	}, nil
}

func (s *Service) Fill(text string) string {
	for name, content := range s.cfg.Secrets {
		pattern := "%frank(" + name + ")"
		text = strings.ReplaceAll(text, pattern, content)
	}

	return text
}
