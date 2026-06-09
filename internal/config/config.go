package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	gotenv "github.com/subosito/gotenv"
)

// Config represents the application configuration.
type Config struct {
	Templates TemplatesConfig `mapstructure:"templates"`
	Diff      DiffConfig      `mapstructure:"diff"`
	GitHub    GitHubConfig    `mapstructure:"github"`
	Jules     JulesConfig     `mapstructure:"jules"`
}

// JulesConfig contains Jules API configuration.
type JulesConfig struct {
	APIKey        string        `mapstructure:"api_key"`
	BaseURL       string        `mapstructure:"base_url"`
	Timeout       time.Duration `mapstructure:"timeout"`
	RetryAttempts int           `mapstructure:"retry_attempts"`
	DebugLog      bool          `mapstructure:"debug_log"`
}

// GitHubConfig contains GitHub API configuration.
type GitHubConfig struct {
	Token      string                `mapstructure:"token"`
	DefaultOrg string                `mapstructure:"default_org"`
	PR         GitHubPRConfig        `mapstructure:"pr"`
	Discovery  GitHubDiscoveryConfig `mapstructure:"discovery"`
}

// GitHubPRConfig contains GitHub PR settings.
type GitHubPRConfig struct {
	DefaultMergeMethod string `mapstructure:"default_merge_method"`
	AutoDeleteBranch   bool   `mapstructure:"auto_delete_branch"`
}

// GitHubDiscoveryConfig contains GitHub repository discovery settings.
type GitHubDiscoveryConfig struct {
	Enabled      bool          `mapstructure:"enabled"`
	UseGitRemote bool          `mapstructure:"use_git_remote"`
	CacheTTL     time.Duration `mapstructure:"cache_ttl"`
}

// TemplatesConfig contains template settings.
type TemplatesConfig struct {
	BuiltinPath  string `mapstructure:"builtin_path"`
	CustomPath   string `mapstructure:"custom_path"`
	EnableCustom bool   `mapstructure:"enable_custom"`
}

// DiffConfig contains diff viewing settings.
type DiffConfig struct {
	Tool        string `mapstructure:"tool"`
	ForceNative bool   `mapstructure:"force_native"`
}

// Load loads configuration from file and environment variables.
func Load() (*Config, error) {
	return load(true, true)
}

// LoadOptional loads configuration without requiring a Jules API key.
// Commands that need Jules API access should still validate credentials before use.
func LoadOptional() (*Config, error) {
	return load(false, true)
}

// LoadForValidation loads configuration without semantic validation so the
// config validate command can report all findings itself.
func LoadForValidation() (*Config, error) {
	return load(false, false)
}

func load(requireJulesAPIKey bool, validateConfig bool) (*Config, error) {
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
	applyCredentialFallbacks(&config)

	// Validate configuration
	if validateConfig {
		if err := validate(&config, requireJulesAPIKey); err != nil {
			return nil, fmt.Errorf("configuration validation failed: %w", err)
		}
	}

	return &config, nil
}

func applyCredentialFallbacks(config *Config) {
	if config.Jules.APIKey == "" {
		config.Jules.APIKey = os.Getenv("JULES_API_KEY")
	}
	if config.GitHub.Token == "" {
		config.GitHub.Token = os.Getenv("GITHUB_TOKEN")
	}
}

// loadEnvFiles loads .env files from multiple possible locations.
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

// setDefaults sets default configuration values.
func setDefaults() {
	viper.SetDefault("jules.api_key", "")
	viper.SetDefault("jules.base_url", "https://jules.googleapis.com/v1alpha")
	viper.SetDefault("jules.timeout", "30s")
	viper.SetDefault("jules.retry_attempts", 3)
	viper.SetDefault("jules.debug_log", false)

	viper.SetDefault("github.token", "")
	viper.SetDefault("github.default_org", "")
	viper.SetDefault("github.pr.default_merge_method", "squash")
	viper.SetDefault("github.pr.auto_delete_branch", true)
	viper.SetDefault("github.discovery.enabled", true)
	viper.SetDefault("github.discovery.use_git_remote", true)
	viper.SetDefault("github.discovery.cache_ttl", "5m")

	viper.SetDefault("templates.builtin_path", "./templates/builtin")
	viper.SetDefault("templates.custom_path", "${JULES_TEMPLATES_CUSTOM_PATH:-./templates/custom}")
	viper.SetDefault("templates.enable_custom", true)

	viper.SetDefault("diff.tool", "")
	viper.SetDefault("diff.force_native", false)
}

// validate validates the configuration.
func validate(config *Config, requireJulesAPIKey bool) error {
	if config.Jules.APIKey == "" && requireJulesAPIKey {
		return fmt.Errorf("Jules API key is required - set it in juleson.yaml or JULES_API_KEY environment variable") //nolint:staticcheck
	}

	return nil
}

// Save saves the configuration back to the config file.
func (c *Config) Save() error {
	// Set the values in viper
	viper.Set("jules.api_key", c.Jules.APIKey)
	viper.Set("jules.base_url", c.Jules.BaseURL)
	viper.Set("jules.timeout", c.Jules.Timeout.String())
	viper.Set("jules.retry_attempts", c.Jules.RetryAttempts)
	viper.Set("jules.debug_log", c.Jules.DebugLog)

	viper.Set("github.token", c.GitHub.Token)
	viper.Set("github.default_org", c.GitHub.DefaultOrg)
	viper.Set("github.pr.default_merge_method", c.GitHub.PR.DefaultMergeMethod)
	viper.Set("github.pr.auto_delete_branch", c.GitHub.PR.AutoDeleteBranch)
	viper.Set("github.discovery.enabled", c.GitHub.Discovery.Enabled)
	viper.Set("github.discovery.use_git_remote", c.GitHub.Discovery.UseGitRemote)
	viper.Set("github.discovery.cache_ttl", c.GitHub.Discovery.CacheTTL.String())

	viper.Set("templates.builtin_path", c.Templates.BuiltinPath)
	viper.Set("templates.custom_path", c.Templates.CustomPath)
	viper.Set("templates.enable_custom", c.Templates.EnableCustom)

	viper.Set("diff.tool", c.Diff.Tool)
	viper.Set("diff.force_native", c.Diff.ForceNative)

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
