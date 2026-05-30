package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyCredentialFallbacks(t *testing.T) {
	cases := []struct {
		name          string
		initial       Config
		envVars       map[string]string
		expectedJules string
		expectedGH    string
		expectedGem   string
	}{
		{
			name:    "empty config with env vars",
			initial: Config{},
			envVars: map[string]string{
				"JULES_API_KEY":  "env-jules",
				"GITHUB_TOKEN":   "env-gh",
				"GEMINI_API_KEY": "env-gemini",
			},
			expectedJules: "env-jules",
			expectedGH:    "env-gh",
			expectedGem:   "env-gemini",
		},
		{
			name: "config values take precedence over env vars",
			initial: Config{
				Jules:  JulesConfig{APIKey: "config-jules"},
				GitHub: GitHubConfig{Token: "config-gh"},
				Gemini: GeminiConfig{APIKey: "config-gemini"},
			},
			envVars: map[string]string{
				"JULES_API_KEY":  "env-jules",
				"GITHUB_TOKEN":   "env-gh",
				"GEMINI_API_KEY": "env-gemini",
			},
			expectedJules: "config-jules",
			expectedGH:    "config-gh",
			expectedGem:   "config-gemini",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup env vars
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}

			cfg := tc.initial
			applyCredentialFallbacks(&cfg)

			assert.Equal(t, tc.expectedJules, cfg.Jules.APIKey)
			assert.Equal(t, tc.expectedGH, cfg.GitHub.Token)
			assert.Equal(t, tc.expectedGem, cfg.Gemini.APIKey)
		})
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name               string
		config             Config
		requireJulesAPIKey bool
		expectError        bool
		errorContains      string
	}{
		{
			name: "valid config",
			config: Config{
				Jules: JulesConfig{APIKey: "valid-key"},
				MCP:   MCPConfig{Server: MCPServerConfig{Port: 8080}},
				Automation: AutomationConfig{
					MaxConcurrentTasks: 5,
				},
			},
			requireJulesAPIKey: true,
			expectError:        false,
		},
		{
			name: "missing jules api key when required",
			config: Config{
				Jules: JulesConfig{APIKey: ""},
				MCP:   MCPConfig{Server: MCPServerConfig{Port: 8080}},
				Automation: AutomationConfig{
					MaxConcurrentTasks: 5,
				},
			},
			requireJulesAPIKey: true,
			expectError:        true,
			errorContains:      "Jules API key is required",
		},
		{
			name: "missing jules api key when not required",
			config: Config{
				Jules: JulesConfig{APIKey: ""},
				MCP:   MCPConfig{Server: MCPServerConfig{Port: 8080}},
				Automation: AutomationConfig{
					MaxConcurrentTasks: 5,
				},
			},
			requireJulesAPIKey: false,
			expectError:        false,
		},
		{
			name: "invalid mcp port zero",
			config: Config{
				Jules: JulesConfig{APIKey: "valid-key"},
				MCP:   MCPConfig{Server: MCPServerConfig{Port: 0}},
				Automation: AutomationConfig{
					MaxConcurrentTasks: 5,
				},
			},
			requireJulesAPIKey: true,
			expectError:        true,
			errorContains:      "invalid MCP server port",
		},
		{
			name: "invalid mcp port too large",
			config: Config{
				Jules: JulesConfig{APIKey: "valid-key"},
				MCP:   MCPConfig{Server: MCPServerConfig{Port: 70000}},
				Automation: AutomationConfig{
					MaxConcurrentTasks: 5,
				},
			},
			requireJulesAPIKey: true,
			expectError:        true,
			errorContains:      "invalid MCP server port",
		},
		{
			name: "invalid max concurrent tasks zero",
			config: Config{
				Jules: JulesConfig{APIKey: "valid-key"},
				MCP:   MCPConfig{Server: MCPServerConfig{Port: 8080}},
				Automation: AutomationConfig{
					MaxConcurrentTasks: 0,
				},
			},
			requireJulesAPIKey: true,
			expectError:        true,
			errorContains:      "max concurrent tasks must be greater than 0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate(&tc.config, tc.requireJulesAPIKey)
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSetDefaults(t *testing.T) {
	// Not testing full viper instance, but ensuring we can call LoadForValidation
	// which resets and uses setDefaults implicitly.
	// Since viper is a global in the implementation, this might affect other tests
	// but within a single test it's fine to just test that the defaults get applied.

	// Create a backup of env var to ensure clean run
	t.Setenv("JULES_API_KEY", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GEMINI_API_KEY", "")

	// Temporarily override HOME to avoid loading actual user configs
	t.Setenv("HOME", "/tmp/non-existent-home-dir-for-tests")

	cfg, err := LoadForValidation()

	// We expect no error because LoadForValidation calls load(false, false)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check some default values
	assert.Equal(t, "https://jules.googleapis.com/v1alpha", cfg.Jules.BaseURL)
	assert.Equal(t, 3, cfg.Jules.RetryAttempts)
	assert.Equal(t, 30*time.Second, cfg.Jules.Timeout)

	assert.Equal(t, "squash", cfg.GitHub.PR.DefaultMergeMethod)
	assert.True(t, cfg.GitHub.PR.AutoDeleteBranch)

	assert.Equal(t, "gemini-api", cfg.Gemini.Backend)
	assert.Equal(t, "us-central1", cfg.Gemini.Location)
	assert.Equal(t, "gemini-2.0-flash", cfg.Gemini.Model)
	assert.Equal(t, 8192, cfg.Gemini.MaxTokens)

	assert.Equal(t, "gemini", cfg.ActiveBackend)
	assert.Equal(t, 8080, cfg.MCP.Server.Port)
	assert.Equal(t, 5, cfg.Automation.MaxConcurrentTasks)
	assert.Equal(t, "./projects", cfg.Projects.DefaultPath)
}
