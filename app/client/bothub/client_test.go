package bothub

import (
	"context"
	"frank/pkg/config"
	"os"
	"testing"

	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Process_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load configuration
	cfg, err := config.Load()
	require.NoError(t, err)

	// Check if Bothub token is configured for integration tests
	if cfg.Bothub.Token == "" {
		t.Skip("Skipping integration test: Bothub token not configured")
	}

	// Setup dependency injection
	di := do.New()
	do.ProvideValue(di, cfg)

	// Create client
	client, err := NewClient(di)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Test case: Basic arithmetic prompt
	t.Run("should return 4 for addition prompt", func(t *testing.T) {
		prompt := Prompt{
			SystemText: "You are a helpful math assistant. Return only the numerical result without any additional text or explanations.",
			UserText:   "add two + 2 and return a single number without any other comments",
			Model:      "deepseek-chat-v3-0324",
		}

		result, err := client.Process(context.Background(), prompt)
		require.NoError(t, err)
		assert.Equal(t, "4", result, "Expected result to be exactly '4'")
	})
}

func TestClient_Process_ErrorCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg, err := config.Load()
	require.NoError(t, err)

	if cfg.Bothub.Token == "" {
		t.Skip("Skipping integration test: Bothub token not configured")
	}

	di := do.New()
	do.ProvideValue(di, cfg)

	client, err := NewClient(di)
	require.NoError(t, err)

	// Test case: Invalid model
	t.Run("should handle invalid model gracefully", func(t *testing.T) {
		prompt := Prompt{
			UserText: "Hello",
			Model:    "invalid-model-name-123",
		}

		result, err := client.Process(context.Background(), prompt)
		assert.Error(t, err, "Should return error for invalid model")
		assert.Empty(t, result, "Result should be empty on error")
		assert.Contains(t, err.Error(), "API request failed", "Error should indicate API failure")
	})

	// Test case: Empty prompt
	t.Run("should handle empty user text", func(t *testing.T) {
		prompt := Prompt{
			UserText: "",
			Model:    "deepseek-chat-v3-0324",
		}

		result, err := client.Process(context.Background(), prompt)
		assert.Error(t, err, "Should return error for empty user text")
		assert.Empty(t, result, "Result should be empty on error")
	})
}

// TestMain handles setup/teardown for integration tests
func TestMain(m *testing.M) {
	// Check if we should run integration tests
	cfg, err := config.Load()
	if err != nil {
		// If config loading fails, we might be in unit test mode
		// Just run the tests and let them handle skipping
		os.Exit(m.Run())
	}

	if cfg.Bothub.Token == "" {
		// No token configured, but we can still run unit tests
		os.Exit(m.Run())
	}

	// Run all tests
	os.Exit(m.Run())
}
