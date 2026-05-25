package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/SamyRai/juleson/internal/config"
)

func TestConfigValidateCommand(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *config.Config
		wantOutput    []string
		wantErr       bool
		notWantOutput []string
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
				Gemini: config.GeminiConfig{
					APIKey:  "secret-gemini-key",
					Backend: "gemini-api",
				},
				MCP: config.MCPConfig{
					Server: config.MCPServerConfig{
						Port: 8080,
					},
				},
				Automation: config.AutomationConfig{
					MaxConcurrentTasks: 5,
				},
			},
			wantOutput: []string{
				"✅ Jules API key is configured.",
				"✅ GitHub token is configured.",
				"✅ Gemini integration is configured.",
				"✅ Validation complete.",
			},
			wantErr: false,
			notWantOutput: []string{
				"secret-jules-key",
				"secret-github-token",
				"secret-gemini-key",
			},
		},
		{
			name: "missing optional credentials",
			cfg: &config.Config{
				MCP: config.MCPConfig{
					Server: config.MCPServerConfig{
						Port: 8080,
					},
				},
				Automation: config.AutomationConfig{
					MaxConcurrentTasks: 5,
				},
				Gemini: config.GeminiConfig{
					Backend: "gemini-api",
				},
			},
			wantOutput: []string{
				"⚠️  Jules API key is missing.",
				"⚠️  GitHub token is missing.",
				"⚠️  Gemini API key is missing",
			},
			wantErr: false,
		},
		{
			name: "vertex-ai missing project",
			cfg: &config.Config{
				Jules: config.JulesConfig{
					APIKey: "key",
				},
				GitHub: config.GitHubConfig{
					Token: "token",
				},
				MCP: config.MCPConfig{
					Server: config.MCPServerConfig{
						Port: 8080,
					},
				},
				Automation: config.AutomationConfig{
					MaxConcurrentTasks: 5,
				},
				Gemini: config.GeminiConfig{
					Backend: "vertex-ai",
					Project: "",
				},
			},
			wantOutput: []string{
				"⚠️  Gemini Vertex AI project is missing.",
			},
			wantErr: false,
		},
		{
			name: "vertex-ai complete",
			cfg: &config.Config{
				Jules: config.JulesConfig{
					APIKey: "key",
				},
				GitHub: config.GitHubConfig{
					Token: "token",
				},
				MCP: config.MCPConfig{
					Server: config.MCPServerConfig{
						Port: 8080,
					},
				},
				Automation: config.AutomationConfig{
					MaxConcurrentTasks: 5,
				},
				Gemini: config.GeminiConfig{
					Backend: "vertex-ai",
					Project: "my-project",
				},
			},
			wantOutput: []string{
				"✅ Gemini integration is configured.",
			},
			wantErr: false,
		},
		{
			name: "invalid MCP port",
			cfg: &config.Config{
				MCP: config.MCPConfig{
					Server: config.MCPServerConfig{
						Port: 0,
					},
				},
				Automation: config.AutomationConfig{
					MaxConcurrentTasks: 5,
				},
			},
			wantOutput: []string{
				"❌ Error: Invalid MCP server port: 0",
				"❌ Configuration validation failed with errors.",
			},
			wantErr: true,
		},
		{
			name: "invalid max concurrent tasks",
			cfg: &config.Config{
				MCP: config.MCPConfig{
					Server: config.MCPServerConfig{
						Port: 8080,
					},
				},
				Automation: config.AutomationConfig{
					MaxConcurrentTasks: -1,
				},
			},
			wantOutput: []string{
				"❌ Error: Invalid automation max concurrent tasks: -1",
				"❌ Configuration validation failed with errors.",
			},
			wantErr: true,
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
