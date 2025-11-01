package mcp

import (
	"testing"
	"time"

	"github.com/SamyRai/juleson/internal/config"

	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		Jules: config.JulesConfig{
			APIKey:        "test-key",
			BaseURL:       "https://test.googleapis.com/v1alpha",
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
		},
		Templates: config.TemplatesConfig{
			BuiltinPath:  "./templates/builtin",
			CustomPath:   "./templates/custom",
			EnableCustom: false,
		},
	}

	server := NewServer(cfg)
	require.NotNil(t, server)
}

func TestServerInitialization(t *testing.T) {
	cfg := &config.Config{
		Jules: config.JulesConfig{
			APIKey:        "test-key",
			BaseURL:       "https://test.googleapis.com/v1alpha",
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
		},
		Templates: config.TemplatesConfig{
			BuiltinPath:  "./templates/builtin",
			CustomPath:   "./templates/custom",
			EnableCustom: false,
		},
	}

	server := NewServer(cfg)
	require.NotNil(t, server)

	// Check that managers are initialized
	_, err := server.container.TemplateManager()
	require.NoError(t, err)

	_, err = server.container.AutomationEngine()
	require.NoError(t, err)
}
