package yandex

import (
	"context"
	"frank/pkg/config"
	"testing"

	"github.com/samber/do"
	"github.com/stretchr/testify/require"
)

func TestClient_WebSearch_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load configuration
	cfg, err := config.Load()
	require.NoError(t, err)

	// Setup dependency injection
	di := do.New()
	do.ProvideValue(di, cfg)

	// Create client
	client, err := NewClient(di)
	require.NoError(t, err)
	require.NotNil(t, client)

	t.Run("should return google", func(t *testing.T) {
		result, err := client.WebSearch(context.Background(), "current weather saint petersburg russia")
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Contains(t, result, "google.com")
	})
}
