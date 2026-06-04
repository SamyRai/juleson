package core

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/logger"
	"github.com/SamyRai/juleson/internal/presentation/views/theme"
	"github.com/SamyRai/juleson/pkg/builder"
	"github.com/spf13/cobra"
	"log/slog"
)

var (
	setupNonInteractive bool
	setupSkipCompletion bool
	setupSkipGithub     bool
	setupSkipJules      bool
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard for Juleson",
	Long: `Run the setup wizard to configure Juleson for first-time use.

This command will guide you through:
• Installing shell completion
• Configuring Jules API credentials
• Setting up GitHub integration
• Validating your configuration

You can run this command anytime to reconfigure or verify your setup.`,
	RunE: runSetup,
}

func init() {
	setupCmd.Flags().BoolVar(&setupNonInteractive, "non-interactive", false, "Run setup without prompts (use existing config)")
	setupCmd.Flags().BoolVar(&setupSkipCompletion, "skip-completion", false, "Skip shell completion installation")
	setupCmd.Flags().BoolVar(&setupSkipGithub, "skip-github", false, "Skip GitHub configuration")
	setupCmd.Flags().BoolVar(&setupSkipJules, "skip-jules", false, "Skip Jules API configuration")
}

func runSetup(cmd *cobra.Command, args []string) error {
	slog.Info("Juleson Setup Wizard")

	// Load existing config or create new one
	cfg, err := config.Load()
	if err != nil {
		slog.Warn("No existing configuration found. Creating new configuration...")
		cfg = &config.Config{}
	} else {
		logger.Success(slog.Default(), "Found existing configuration")
	}

	// Step 1: Shell Completion
	if !setupSkipCompletion {
		if err := setupShellCompletion(cmd); err != nil {
			slog.Warn(fmt.Sprintf("Shell completion setup failed: %v", err))
		}
	}

	// Step 2: Jules API Configuration
	if !setupSkipJules {
		if err := setupJulesAPI(cfg); err != nil {
			slog.Error(fmt.Sprintf("Jules API setup failed: %v", err))
			return err
		}
	}

	// Step 3: GitHub Configuration
	if !setupSkipGithub {
		if err := setupGitHub(cfg); err != nil {
			slog.Warn(fmt.Sprintf("GitHub setup failed: %v", err))
			// Don't fail setup if GitHub config fails - it's optional
		}
	}

	// Step 4: Save Configuration
	if err := saveConfiguration(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Step 5: Validate Setup
	slog.Info("Validating configuration...")
	if err := validateSetup(cfg); err != nil {
		slog.Warn(fmt.Sprintf("Configuration validation found issues: %v", err))
		slog.Debug("You can fix these later by running:\n  • juleson setup  (to rerun setup)\n  • set GITHUB_TOKEN for Jules-created PR context")
	} else {
		logger.Success(slog.Default(), "Configuration is valid!")
	}

	// Success message
	slog.Info("Setup Complete!")
	slog.Debug("Next steps:\n  • Run 'juleson sources list' to inspect connected sources\n  • Run 'juleson sessions create \"your prompt\"' to start a session\n  • Run 'juleson mcp serve --version' to smoke-test MCP\n  • Run 'juleson help' to see all available commands")

	return nil
}

func setupShellCompletion(cmd *cobra.Command) error {
	slog.Info("Shell Completion Setup")

	shell := detectShell()
	slog.Info(fmt.Sprintf("Detected shell: %s", shell))

	if setupNonInteractive {
		slog.Warn("Non-interactive mode: Skipping completion installation")
		slog.Debug(fmt.Sprintf("To install manually, run: juleson completion %s", shell))
		return nil
	}

	install, _ := theme.Confirm("Install shell completion?", true)

	if !install {
		slog.Debug("Skipped shell completion installation")
		slog.Debug(fmt.Sprintf("To install later, run: juleson completion %s", shell))
		return nil
	}

	// Install completion based on shell
	switch shell {
	case "zsh":
		return installZshCompletion(cmd)
	case "bash":
		return installBashCompletion(cmd)
	case "fish":
		return installFishCompletion(cmd)
	default:
		slog.Warn(fmt.Sprintf("Unsupported shell: %s", shell))
		slog.Debug(fmt.Sprintf("To install manually, run: juleson completion %s", shell))
		return nil
	}
}

func setupJulesAPI(cfg *config.Config) error {
	slog.Info("Jules API Configuration")

	// Check if API key already exists
	if cfg.Jules.APIKey != "" {
		logger.Success(slog.Default(), "Jules API key is already configured")
		if !setupNonInteractive {
			update, _ := theme.Confirm("Update API key?", false)
			if !update {
				return nil
			}
		} else {
			return nil
		}
	}

	if setupNonInteractive {
		// Check environment variable
		if apiKey := os.Getenv("JULES_API_KEY"); apiKey != "" {
			cfg.Jules.APIKey = apiKey
			logger.Success(slog.Default(), "Using JULES_API_KEY from environment")
			return nil
		}
		return fmt.Errorf("Jules API key not found. Set JULES_API_KEY environment variable")
	}

	// Interactive prompt
	slog.Info("To use Jules, you need an API key.\nGet your API key from: https://jules.ai/settings/api")

	apiKey, _ := theme.InputSecret("Enter your Jules API key")
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	cfg.Jules.APIKey = apiKey
	cfg.Jules.BaseURL = "https://jules.googleapis.com/v1alpha"
	cfg.Jules.Timeout = 30000000000 // 30s in nanoseconds
	cfg.Jules.RetryAttempts = 3

	logger.Success(slog.Default(), "Jules API key configured")
	return nil
}

func setupGitHub(cfg *config.Config) error {
	slog.Info("GitHub Integration Setup")

	// Check if token already exists
	if cfg.GitHub.Token != "" {
		logger.Success(slog.Default(), "GitHub token is already configured")
		if !setupNonInteractive {
			update, _ := theme.Confirm("Update GitHub token?", false)
			if !update {
				return nil
			}
		} else {
			return nil
		}
	}

	if setupNonInteractive {
		// Check environment variable
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			cfg.GitHub.Token = token
			logger.Success(slog.Default(), "Using GITHUB_TOKEN from environment")
			return nil
		}
		slog.Info("GitHub token not found. Skipping Jules-created PR integration.\nYou can configure it later by setting GITHUB_TOKEN or github.token.")
		return nil
	}

	// Interactive prompt
	slog.Info("GitHub integration is optional and used for Jules-created pull request context.")

	install, _ := theme.Confirm("Configure GitHub integration now?", true)
	if !install {
		slog.Debug("Skipped GitHub integration. To configure later, set GITHUB_TOKEN or github.token.")
		return nil
	}

	slog.Info("To inspect or merge Jules-created pull requests, you need a Personal Access Token.\nCreate one at: https://github.com/settings/tokens\n\nRequired scopes:\n  • repo")
	token, _ := theme.InputSecret("Enter your GitHub Personal Access Token")
	token = strings.TrimSpace(token)

	if token == "" {
		slog.Debug("Skipped GitHub integration (empty token)")
		return nil
	}

	cfg.GitHub.Token = token
	cfg.GitHub.DefaultOrg = ""
	cfg.GitHub.PR.DefaultMergeMethod = "squash"
	cfg.GitHub.PR.AutoDeleteBranch = true
	cfg.GitHub.Discovery.Enabled = true
	cfg.GitHub.Discovery.UseGitRemote = true
	cfg.GitHub.Discovery.CacheTTL = 300000000000 // 5m in nanoseconds

	logger.Success(slog.Default(), "GitHub integration configured")
	return nil
}

func saveConfiguration(cfg *config.Config) error {
	slog.Info("Saving Configuration")

	// Ensure config directory exists
	configDir := filepath.Join(".", "configs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save configuration
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	logger.Success(slog.Default(), "Configuration saved to configs/juleson.yaml")
	return nil
}

func validateSetup(cfg *config.Config) error {
	// Basic validation
	if cfg.Jules.APIKey == "" {
		return fmt.Errorf("Jules API key is not configured")
	}

	// Optional: Try to validate API key by making a test request
	// This would require importing the jules client

	return nil
}

func detectShell() string {
	// Check SHELL environment variable
	shell := os.Getenv("SHELL")
	if shell != "" {
		// Extract shell name from path
		return filepath.Base(shell)
	}

	// Fallback based on OS
	switch runtime.GOOS {
	case "windows":
		return "powershell"
	default:
		return "bash"
	}
}

func installZshCompletion(cmd *cobra.Command) error {
	// Try to find zsh completion directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Check if .zfunc directory exists, create if not
	zfuncDir := filepath.Join(homeDir, ".zfunc")
	if _, err := os.Stat(zfuncDir); os.IsNotExist(err) {
		if err := os.MkdirAll(zfuncDir, 0755); err != nil {
			return fmt.Errorf("failed to create .zfunc directory: %w", err)
		}
	}

	completionFile := filepath.Join(zfuncDir, "_juleson")
	file, err := os.Create(completionFile)
	if err != nil {
		return fmt.Errorf("failed to create completion file: %w", err)
	}
	defer file.Close()

	if err := cmd.Root().GenZshCompletion(file); err != nil {
		return fmt.Errorf("failed to generate zsh completion: %w", err)
	}

	fmt.Printf("✅ Zsh completion installed to: %s\n", completionFile)
	fmt.Println()
	fmt.Println("To enable completion, add this to your ~/.zshrc:")
	fmt.Println("  fpath=(~/.zfunc $fpath)")
	fmt.Println("  autoload -Uz compinit && compinit")
	fmt.Println()
	fmt.Println("Then restart your shell or run: source ~/.zshrc")

	return nil
}

func installBashCompletion(cmd *cobra.Command) error {
	// Check if bash-completion is installed
	if !builder.CommandAvailable("bash-completion") {
		fmt.Println("⚠️  bash-completion is not installed")
		fmt.Println("Install it with your package manager, then run:")
		fmt.Println("  juleson completion bash > /etc/bash_completion.d/juleson")
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Create .bash_completion.d directory
	completionDir := filepath.Join(homeDir, ".bash_completion.d")
	if _, err := os.Stat(completionDir); os.IsNotExist(err) {
		if err := os.MkdirAll(completionDir, 0755); err != nil {
			return fmt.Errorf("failed to create completion directory: %w", err)
		}
	}

	completionFile := filepath.Join(completionDir, "juleson")
	file, err := os.Create(completionFile)
	if err != nil {
		return fmt.Errorf("failed to create completion file: %w", err)
	}
	defer file.Close()

	if err := cmd.Root().GenBashCompletion(file); err != nil {
		return fmt.Errorf("failed to generate bash completion: %w", err)
	}

	fmt.Printf("✅ Bash completion installed to: %s\n", completionFile)
	fmt.Println()
	fmt.Println("To enable completion, add this to your ~/.bashrc:")
	fmt.Printf("  source %s\n", completionFile)
	fmt.Println()
	fmt.Println("Then restart your shell or run: source ~/.bashrc")

	return nil
}

func installFishCompletion(cmd *cobra.Command) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Create fish completions directory
	completionDir := filepath.Join(homeDir, ".config", "fish", "completions")
	if _, err := os.Stat(completionDir); os.IsNotExist(err) {
		if err := os.MkdirAll(completionDir, 0755); err != nil {
			return fmt.Errorf("failed to create completion directory: %w", err)
		}
	}

	completionFile := filepath.Join(completionDir, "juleson.fish")
	file, err := os.Create(completionFile)
	if err != nil {
		return fmt.Errorf("failed to create completion file: %w", err)
	}
	defer file.Close()

	if err := cmd.Root().GenFishCompletion(file, true); err != nil {
		return fmt.Errorf("failed to generate fish completion: %w", err)
	}

	fmt.Printf("✅ Fish completion installed to: %s\n", completionFile)
	fmt.Println()
	fmt.Println("Completion is automatically loaded by fish shell")
	fmt.Println("Restart your shell to enable completion")

	return nil
}

// NewSetupCommand creates the setup command
func NewSetupCommand() *cobra.Command {
	return setupCmd
}
