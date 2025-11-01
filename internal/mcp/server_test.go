package mcp_test

import (
	"testing"
	"time"

	"jules-automation/internal/config"
	"jules-automation/internal/mcp"

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
	}

	server := mcp.NewServer(cfg)
	require.NotNil(t, server)
}
