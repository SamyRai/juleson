package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyCredentialFallbacks(t *testing.T) {
	cases := []struct {
		envVars       map[string]string
		name          string
		expectedJules string
		expectedGH    string
		expectedGem   string
		initial       Config
	}{
		{
			name:    "empty config with env vars",
			initial: Config{},
			envVars: map[string]string{
				"JULES_API_KEY": "env-jules",
				"GITHUB_TOKEN":  "env-gh",
			},
			expectedJules: "env-jules",
			expectedGH:    "env-gh",
		},
		{
			name: "config values take precedence over env vars",
			initial: Config{
				Jules:  JulesConfig{APIKey: "config-jules"},
				GitHub: GitHubConfig{Token: "config-gh"},
			},
			envVars: map[string]string{
				"JULES_API_KEY": "env-jules",
				"GITHUB_TOKEN":  "env-gh",
			},
			expectedJules: "config-jules",
			expectedGH:    "config-gh",
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
		})
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name               string
		errorContains      string
		config             Config
		requireJulesAPIKey bool
		expectError        bool
	}{
		{
			name: "valid config",
			config: Config{
				Jules: JulesConfig{APIKey: "valid-key"},
			},
			requireJulesAPIKey: true,
			expectError:        false,
		},
		{
			name: "missing jules api key when required",
			config: Config{
				Jules: JulesConfig{APIKey: ""},
			},
			requireJulesAPIKey: true,
			expectError:        true,
			errorContains:      "Jules API key is required",
		},
		{
			name: "missing jules api key when not required",
			config: Config{
				Jules: JulesConfig{APIKey: ""},
			},
			requireJulesAPIKey: false,
			expectError:        false,
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
}
