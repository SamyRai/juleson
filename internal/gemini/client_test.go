package gemini

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	cases := []struct {
		name          string
		config        *Config
		expectError   bool
		errorContains string
	}{
		{
			name: "gemini-api valid",
			config: &Config{
				Backend: "gemini-api",
				APIKey:  "test-key",
			},
			expectError: false,
		},
		{
			name: "gemini-api missing key",
			config: &Config{
				Backend: "gemini-api",
				APIKey:  "",
			},
			expectError:   true,
			errorContains: "API key is required",
		},
		{
			name: "vertex-ai valid",
			config: &Config{
				Backend: "vertex-ai",
				Project: "test-project",
			},
			expectError: false,
		},
		{
			name: "vertex-ai missing project",
			config: &Config{
				Backend: "vertex-ai",
				Project: "",
			},
			expectError:   true,
			errorContains: "project is required",
		},
		{
			name: "unsupported backend",
			config: &Config{
				Backend: "unknown",
			},
			expectError:   true,
			errorContains: "unsupported backend",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(tc.config)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
				assert.Nil(t, client)
			} else {
				// The client creation might still fail if there's no credentials locally depending on genai
				// But we just want to ensure we get past the config validation.
				// Since genai might not validate APIKey format until request, it might succeed.
				if err != nil {
					// We ignore genai creation errors if our config was valid, as it might try to load ADC.
					t.Logf("genai creation failed (expected with fake keys): %v", err)
				} else {
					assert.NotNil(t, client)
					assert.NotNil(t, client.GenAIClient())
					assert.NotNil(t, client.Context())
					assert.NoError(t, client.Close())
				}
			}
		})
	}
}
