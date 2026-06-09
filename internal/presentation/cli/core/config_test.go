package core

import (
	"bytes"
	"strings"
	"testing"

	"github.com/SamyRai/juleson/internal/config"
)

func TestConfigValidateCommand(t *testing.T) {
	tests := []struct {
		cfg           *config.Config
		name          string
		wantOutput    []string
		notWantOutput []string
		wantErr       bool
	}{
		{
			name: "valid complete configuration",
			cfg: &config.Config{
				Jules: config.JulesConfig{
					APIKey: "secret-jules-key",
				},
				GitHub: config.GitHubConfig{
					Token: "secret-github-token",
				},
			},
			wantOutput: []string{
				"✅ Jules API key is configured.",
				"✅ GitHub token is configured.",
				"✅ Validation complete.",
			},
			wantErr: false,
			notWantOutput: []string{
				"secret-jules-key",
				"secret-github-token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newConfigValidateCommand(tt.cfg)
			outBuf := new(bytes.Buffer)
			cmd.SetOut(outBuf)
			cmd.SetErr(outBuf)

			err := cmd.RunE(cmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("expected error: %v, got: %v", tt.wantErr, err)
			}

			output := outBuf.String()

			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("output missing expected string: %q\nGot:\n%s", want, output)
				}
			}

			for _, notWant := range tt.notWantOutput {
				if strings.Contains(output, notWant) {
					t.Errorf("output contains string that should be hidden: %q\nGot:\n%s", notWant, output)
				}
			}
		})
	}
}
