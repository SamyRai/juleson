package cli

import (
	"bytes"
	"testing"
	"time"

	"github.com/SamyRai/juleson/internal/config"
)

func TestCoreCommandsDoNotRequireJulesAPIKeyForHelp(t *testing.T) {
	tests := [][]string{
		{"--help"},
		{"--version"},
		{"version"},
		{"sessions", "--help"},
		{"sources", "--help"},
		{"activities", "--help"},
		{"official", "--help"},
	}

	for _, args := range tests {
		t.Run(args[0], func(t *testing.T) {
			app := NewApp(minimalTestConfig())
			var output bytes.Buffer
			app.rootCmd.SetOut(&output)
			app.rootCmd.SetErr(&output)
			app.rootCmd.SetArgs(args)

			if err := app.Execute(); err != nil {
				t.Fatalf("Execute(%v): %v", args, err)
			}
		})
	}
}

func minimalTestConfig() *config.Config {
	return &config.Config{
		Jules: config.JulesConfig{
			BaseURL:       "https://jules.googleapis.com/v1alpha",
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
		},

		Templates: config.TemplatesConfig{
			BuiltinPath:  "../../templates/builtin",
			CustomPath:   "../../templates/custom",
			EnableCustom: true,
		},
	}
}
