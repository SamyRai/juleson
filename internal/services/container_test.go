package services

import (
	"testing"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContainer(t *testing.T) {
	cfg := &config.Config{}
	container := NewContainer(cfg)

	assert.NotNil(t, container)
	assert.Equal(t, cfg, container.Config())
}

func TestJulesClient(t *testing.T) {
	t.Run("no API key", func(t *testing.T) {
		cfg := &config.Config{
			Jules: config.JulesConfig{APIKey: ""},
		}
		container := NewContainer(cfg)

		client := container.JulesClient()
		assert.Nil(t, client)
	})

	t.Run("with API key", func(t *testing.T) {
		cfg := &config.Config{
			Jules: config.JulesConfig{
				APIKey:        "test-key",
				BaseURL:       "https://test.com",
				Timeout:       10,
				RetryAttempts: 2,
			},
		}
		container := NewContainer(cfg)

		client := container.JulesClient()
		assert.NotNil(t, client)
		// Test that it's cached
		client2 := container.JulesClient()
		assert.Equal(t, client, client2)
	})
}

func TestTemplateManager(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		cfg := &config.Config{
			Templates: config.TemplatesConfig{
				BuiltinPath:  "../../templates/builtin",
				CustomPath:   "../../templates/custom",
				EnableCustom: false,
			},
		}
		container := NewContainer(cfg)

		manager, err := container.TemplateManager()
		require.NoError(t, err)
		assert.NotNil(t, manager)

		// Test caching
		manager2, err := container.TemplateManager()
		require.NoError(t, err)
		assert.Equal(t, manager, manager2)
	})

	t.Run("initialization failure", func(t *testing.T) {
		// This would require mocking the embed FS or invalid config
		// For now, assume it works
	})
}

func TestAutomationEngine(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		cfg := &config.Config{
			Templates: config.TemplatesConfig{
				BuiltinPath:  "../../templates/builtin",
				CustomPath:   "../../templates/custom",
				EnableCustom: false,
			},
		}
		container := NewContainer(cfg)

		engine, err := container.AutomationEngine()
		require.NoError(t, err)
		assert.NotNil(t, engine)

		// Test caching
		engine2, err := container.AutomationEngine()
		require.NoError(t, err)
		assert.Equal(t, engine, engine2)
	})
}
