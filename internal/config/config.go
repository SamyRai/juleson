package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	gotenv "github.com/subosito/gotenv"
)

// Config represents the application configuration
type Config struct {
	Jules      JulesConfig      `mapstructure:"jules"`
	MCP        MCPConfig        `mapstructure:"mcp"`
	Automation AutomationConfig `mapstructure:"automation"`
	Projects   ProjectsConfig   `mapstructure:"projects"`
}

// JulesConfig contains Jules API configuration
type JulesConfig struct {
	APIKey        string        `mapstructure:"api_key"`
	BaseURL       string        `mapstructure:"base_url"`
	Timeout       time.Duration `mapstructure:"timeout"`
	RetryAttempts int           `mapstructure:"retry_attempts"`
}

// MCPConfig contains MCP server configuration
type MCPConfig struct {
	Server MCPServerConfig `mapstructure:"server"`
	Client MCPClientConfig `mapstructure:"client"`
}

// MCPServerConfig contains MCP server settings
type MCPServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// MCPClientConfig contains MCP client settings
type MCPClientConfig struct {
	Timeout time.Duration `mapstructure:"timeout"`
}

// AutomationConfig contains automation settings
type AutomationConfig struct {
	Strategies         []string      `mapstructure:"strategies"`
	MaxConcurrentTasks int           `mapstructure:"max_concurrent_tasks"`
	TaskTimeout        time.Duration `mapstructure:"task_timeout"`
}

// ProjectsConfig contains project management settings
type ProjectsConfig struct {
	DefaultPath    string `mapstructure:"default_path"`
	BackupEnabled  bool   `mapstructure:"backup_enabled"`
	GitIntegration bool   `mapstructure:"git_integration"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	gotenv.Load(".env")

	viper.SetConfigName("jules-automation")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set default values
	setDefaults()

	// Read environment variables
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, use defaults and env vars
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("jules.api_key", "")
	viper.SetDefault("jules.base_url", "https://jules.googleapis.com/v1alpha")
	viper.SetDefault("jules.timeout", "30s")
	viper.SetDefault("jules.retry_attempts", 3)

	viper.SetDefault("mcp.server.port", 8080)
	viper.SetDefault("mcp.server.host", "localhost")
	viper.SetDefault("mcp.client.timeout", "10s")

	viper.SetDefault("automation.strategies", []string{"modular", "layered", "microservices"})
	viper.SetDefault("automation.max_concurrent_tasks", 5)
	viper.SetDefault("automation.task_timeout", "300s")

	viper.SetDefault("projects.default_path", "./projects")
	viper.SetDefault("projects.backup_enabled", true)
	viper.SetDefault("projects.git_integration", true)
}

// validate validates the configuration
func validate(config *Config) error {
	if config.Jules.APIKey == "" {
		// Try to get from environment
		if apiKey := os.Getenv("JULES_API_KEY"); apiKey != "" {
			config.Jules.APIKey = apiKey
		} else {
			return fmt.Errorf("Jules API key is required")
		}
	}

	if config.MCP.Server.Port <= 0 || config.MCP.Server.Port > 65535 {
		return fmt.Errorf("invalid MCP server port: %d", config.MCP.Server.Port)
	}

	if config.Automation.MaxConcurrentTasks <= 0 {
		return fmt.Errorf("max concurrent tasks must be greater than 0")
	}

	return nil
}
