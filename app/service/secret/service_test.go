package secret

import (
	"frank/pkg/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_Fill(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.Config
		input    string
		expected string
	}{
		{
			name: "replace single secret",
			config: &config.Config{
				Secrets: map[string]string{
					"API_KEY": "secret123",
				},
			},
			input:    "The API key is %frank(API_KEY)",
			expected: "The API key is secret123",
		},
		{
			name: "replace multiple secrets",
			config: &config.Config{
				Secrets: map[string]string{
					"DB_PASSWORD": "dbpass456",
					"API_SECRET":  "apisecret789",
				},
			},
			input:    "DB: %frank(DB_PASSWORD), API: %frank(API_SECRET)",
			expected: "DB: dbpass456, API: apisecret789",
		},
		{
			name: "no secrets to replace",
			config: &config.Config{
				Secrets: map[string]string{
					"TOKEN": "token123",
				},
			},
			input:    "This text has no secrets",
			expected: "This text has no secrets",
		},
		{
			name: "secret pattern but no matching secret",
			config: &config.Config{
				Secrets: map[string]string{
					"EXISTING": "value",
				},
			},
			input:    "Missing: %frank(NON_EXISTENT)",
			expected: "Missing: %frank(NON_EXISTENT)",
		},
		{
			name: "empty secrets map",
			config: &config.Config{
				Secrets: map[string]string{},
			},
			input:    "Text with %frank(ANY) pattern",
			expected: "Text with %frank(ANY) pattern",
		},
		{
			name: "multiple occurrences of same secret",
			config: &config.Config{
				Secrets: map[string]string{
					"TOKEN": "same_token",
				},
			},
			input:    "%frank(TOKEN) and %frank(TOKEN) again",
			expected: "same_token and same_token again",
		},
		{
			name: "mixed replacements and non-replacements",
			config: &config.Config{
				Secrets: map[string]string{
					"REAL":    "replaced",
					"ANOTHER": "also_replaced",
				},
			},
			input:    "%frank(REAL) %frank(FAKE) %frank(ANOTHER)",
			expected: "replaced %frank(FAKE) also_replaced",
		},
		{
			name:     "nil secrets map",
			config:   &config.Config{},
			input:    "Text with %frank(ANY) pattern",
			expected: "Text with %frank(ANY) pattern",
		},
		{
			name: "empty secret value",
			config: &config.Config{
				Secrets: map[string]string{
					"EMPTY": "",
				},
			},
			input:    "Value: %frank(EMPTY)",
			expected: "Value: ",
		},
		{
			name: "case sensitive secret names",
			config: &config.Config{
				Secrets: map[string]string{
					"lowercase": "lower_value",
					"UPPERCASE": "upper_value",
					"MixedCase": "mixed_value",
				},
			},
			input:    "%frank(lowercase) %frank(UPPERCASE) %frank(MixedCase)",
			expected: "lower_value upper_value mixed_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{
				cfg: tt.config,
				// queries can be nil since we're not using it in Fill method
			}

			result := service.Fill(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
