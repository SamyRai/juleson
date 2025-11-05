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
	GitHub     GitHubConfig     `mapstructure:"github"`
	Gemini     GeminiConfig     `mapstructure:"gemini"`
	MCP        MCPConfig        `mapstructure:"mcp"`
	Automation AutomationConfig `mapstructure:"automation"`
	Projects   ProjectsConfig   `mapstructure:"projects"`
	Templates  TemplatesConfig  `mapstructure:"templates"`
}

// JulesConfig contains Jules API configuration
type JulesConfig struct {
	APIKey        string        `mapstructure:"api_key"`
	BaseURL       string        `mapstructure:"base_url"`
	Timeout       time.Duration `mapstructure:"timeout"`
	RetryAttempts int           `mapstructure:"retry_attempts"`
}

// GitHubConfig contains GitHub API configuration
type GitHubConfig struct {
	Token      string                `mapstructure:"token"`
	DefaultOrg string                `mapstructure:"default_org"`
	PR         GitHubPRConfig        `mapstructure:"pr"`
	Discovery  GitHubDiscoveryConfig `mapstructure:"discovery"`
}

// GitHubPRConfig contains GitHub PR settings
type GitHubPRConfig struct {
	DefaultMergeMethod string `mapstructure:"default_merge_method"`
	AutoDeleteBranch   bool   `mapstructure:"auto_delete_branch"`
}

// GitHubDiscoveryConfig contains GitHub repository discovery settings
type GitHubDiscoveryConfig struct {
	Enabled      bool          `mapstructure:"enabled"`
	UseGitRemote bool          `mapstructure:"use_git_remote"`
	CacheTTL     time.Duration `mapstructure:"cache_ttl"`
}

// GeminiConfig contains Google Gemini AI configuration
type GeminiConfig struct {
	APIKey    string        `mapstructure:"api_key"`
	Backend   string        `mapstructure:"backend"`  // "gemini-api" or "vertex-ai"
	Project   string        `mapstructure:"project"`  // GCP project for Vertex AI
	Location  string        `mapstructure:"location"` // GCP location for Vertex AI
	Model     string        `mapstructure:"model"`    // Default model (e.g., "gemini-2.0-flash")
	Timeout   time.Duration `mapstructure:"timeout"`
	MaxTokens int           `mapstructure:"max_tokens"`
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

// TemplatesConfig contains template settings
type TemplatesConfig struct {
	BuiltinPath  string `mapstructure:"builtin_path"`
	CustomPath   string `mapstructure:"custom_path"`
	EnableCustom bool   `mapstructure:"enable_custom"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	// Load .env file from multiple possible locations
	loadEnvFiles()

	viper.SetConfigName("juleson")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")
	viper.AddConfigPath(os.Getenv("HOME")) // User's home directory
	viper.AddConfigPath("/etc/juleson")    // System-wide config

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

	// Expand environment variables in paths
	config.Templates.CustomPath = os.ExpandEnv(config.Templates.CustomPath)
	config.Templates.BuiltinPath = os.ExpandEnv(config.Templates.BuiltinPath)

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// loadEnvFiles loads .env files from multiple possible locations
func loadEnvFiles() {
	// Possible locations for .env files (in order of priority)
	envPaths := []string{
		".env",                              // Current working directory
		os.Getenv("HOME") + "/.env",         // User's home directory
		os.Getenv("HOME") + "/.juleson.env", // Juleson-specific config
		"/etc/juleson/.env",                 // System-wide config
	}

	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			// File exists, load it
			gotenv.Load(path)
		}
	}
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("jules.api_key", "")
	viper.SetDefault("jules.base_url", "https://jules.googleapis.com/v1alpha")
	viper.SetDefault("jules.timeout", "30s")
	viper.SetDefault("jules.retry_attempts", 3)

	viper.SetDefault("github.token", "")
	viper.SetDefault("github.default_org", "")
	viper.SetDefault("github.pr.default_merge_method", "squash")
	viper.SetDefault("github.pr.auto_delete_branch", true)
	viper.SetDefault("github.discovery.enabled", true)
	viper.SetDefault("github.discovery.use_git_remote", true)
	viper.SetDefault("github.discovery.cache_ttl", "5m")

	viper.SetDefault("gemini.api_key", "")
	viper.SetDefault("gemini.backend", "gemini-api")
	viper.SetDefault("gemini.project", "")
	viper.SetDefault("gemini.location", "us-central1")
	viper.SetDefault("gemini.model", "gemini-2.0-flash")
	viper.SetDefault("gemini.timeout", "30s")
	viper.SetDefault("gemini.max_tokens", 8192)

	viper.SetDefault("mcp.server.port", 8080)
	viper.SetDefault("mcp.server.host", "localhost")
	viper.SetDefault("mcp.client.timeout", "10s")

	viper.SetDefault("automation.strategies", []string{"modular", "layered", "microservices"})
	viper.SetDefault("automation.max_concurrent_tasks", 5)
	viper.SetDefault("automation.task_timeout", "300s")

	viper.SetDefault("projects.default_path", "./projects")
	viper.SetDefault("projects.backup_enabled", true)
	viper.SetDefault("projects.git_integration", true)

	viper.SetDefault("templates.builtin_path", "./templates/builtin")
	viper.SetDefault("templates.custom_path", "${JULES_TEMPLATES_CUSTOM_PATH:-./templates/custom}")
	viper.SetDefault("templates.enable_custom", true)
}

// validate validates the configuration
func validate(config *Config) error {
	if config.Jules.APIKey == "" {
		// Try to get from environment variable as fallback
		if apiKey := os.Getenv("JULES_API_KEY"); apiKey != "" {
			config.Jules.APIKey = apiKey
		} else {
			return fmt.Errorf("Jules API key is required - set it in juleson.yaml or JULES_API_KEY environment variable")
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

// Save saves the configuration back to the config file
func (c *Config) Save() error {
	// Set the values in viper
	viper.Set("jules.api_key", c.Jules.APIKey)
	viper.Set("jules.base_url", c.Jules.BaseURL)
	viper.Set("jules.timeout", c.Jules.Timeout.String())
	viper.Set("jules.retry_attempts", c.Jules.RetryAttempts)

	viper.Set("github.token", c.GitHub.Token)
	viper.Set("github.default_org", c.GitHub.DefaultOrg)
	viper.Set("github.pr.default_merge_method", c.GitHub.PR.DefaultMergeMethod)
	viper.Set("github.pr.auto_delete_branch", c.GitHub.PR.AutoDeleteBranch)
	viper.Set("github.discovery.enabled", c.GitHub.Discovery.Enabled)
	viper.Set("github.discovery.use_git_remote", c.GitHub.Discovery.UseGitRemote)
	viper.Set("github.discovery.cache_ttl", c.GitHub.Discovery.CacheTTL.String())

	viper.Set("gemini.api_key", c.Gemini.APIKey)
	viper.Set("gemini.backend", c.Gemini.Backend)
	viper.Set("gemini.project", c.Gemini.Project)
	viper.Set("gemini.location", c.Gemini.Location)
	viper.Set("gemini.model", c.Gemini.Model)
	viper.Set("gemini.timeout", c.Gemini.Timeout.String())
	viper.Set("gemini.max_tokens", c.Gemini.MaxTokens)

	viper.Set("mcp.server.port", c.MCP.Server.Port)
	viper.Set("mcp.server.host", c.MCP.Server.Host)
	viper.Set("mcp.client.timeout", c.MCP.Client.Timeout.String())

	viper.Set("automation.strategies", c.Automation.Strategies)
	viper.Set("automation.max_concurrent_tasks", c.Automation.MaxConcurrentTasks)
	viper.Set("automation.task_timeout", c.Automation.TaskTimeout.String())

	viper.Set("projects.default_path", c.Projects.DefaultPath)
	viper.Set("projects.backup_enabled", c.Projects.BackupEnabled)
	viper.Set("projects.git_integration", c.Projects.GitIntegration)

	viper.Set("templates.builtin_path", c.Templates.BuiltinPath)
	viper.Set("templates.custom_path", c.Templates.CustomPath)
	viper.Set("templates.enable_custom", c.Templates.EnableCustom)

	// Try to write to the config file
	if err := viper.WriteConfig(); err != nil {
		// If the config file doesn't exist, create it
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return viper.SafeWriteConfig()
		}
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}
