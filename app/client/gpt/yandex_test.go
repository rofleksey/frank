package gpt

import (
	"context"
	"frank/pkg/config"
	"testing"

	"github.com/samber/do"
	"github.com/stretchr/testify/require"
)

func TestYandex(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	di := do.New()
	do.ProvideValue(di, cfg)

	g, _ := NewYandexGpt(di)

	result, err := g.Process(context.Background(), Prompt{
		SystemText: "выполни математическую операцию, ответь одним числом без лишних комментариев",
		Text:       "два + 2",
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, "4", result)
}
