package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/spf13/cobra"
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
	fmt.Println("🚀 Juleson Setup Wizard")
	fmt.Println("========================")
	fmt.Println()

	// Load existing config or create new one
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("⚠️  No existing configuration found. Creating new configuration...")
		cfg = &config.Config{}
	} else {
		fmt.Println("✅ Found existing configuration")
	}
	fmt.Println()

	// Step 1: Shell Completion
	if !setupSkipCompletion {
		if err := setupShellCompletion(cmd); err != nil {
			fmt.Printf("⚠️  Shell completion setup failed: %v\n", err)
		}
		fmt.Println()
	}

	// Step 2: Jules API Configuration
	if !setupSkipJules {
		if err := setupJulesAPI(cfg); err != nil {
			fmt.Printf("❌ Jules API setup failed: %v\n", err)
			return err
		}
		fmt.Println()
	}

	// Step 3: GitHub Configuration
	if !setupSkipGithub {
		if err := setupGitHub(cfg); err != nil {
			fmt.Printf("⚠️  GitHub setup failed: %v\n", err)
			// Don't fail setup if GitHub config fails - it's optional
		}
		fmt.Println()
	}

	// Step 4: Save Configuration
	if err := saveConfiguration(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Step 5: Validate Setup
	fmt.Println("🔍 Validating configuration...")
	if err := validateSetup(cfg); err != nil {
		fmt.Printf("⚠️  Configuration validation found issues: %v\n", err)
		fmt.Println()
		fmt.Println("You can fix these later by running:")
		fmt.Println("  • juleson setup  (to rerun setup)")
		fmt.Println("  • juleson github login  (for GitHub issues)")
	} else {
		fmt.Println("✅ Configuration is valid!")
	}
	fmt.Println()

	// Success message
	fmt.Println("🎉 Setup Complete!")
	fmt.Println("==================")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  • Run 'juleson analyze' to analyze your current project")
	fmt.Println("  • Run 'juleson github repos' to list your repositories")
	fmt.Println("  • Run 'juleson sessions create \"your prompt\"' to start automation")
	fmt.Println("  • Run 'juleson help' to see all available commands")
	fmt.Println()

	return nil
}

func setupShellCompletion(cmd *cobra.Command) error {
	fmt.Println("📦 Shell Completion Setup")
	fmt.Println("─────────────────────────")

	shell := detectShell()
	fmt.Printf("Detected shell: %s\n", shell)

	if setupNonInteractive {
		fmt.Println("ℹ️  Non-interactive mode: Skipping completion installation")
		fmt.Printf("To install manually, run: juleson completion %s\n", shell)
		return nil
	}

	fmt.Print("Install shell completion? (Y/n): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "n" || response == "no" {
		fmt.Println("⏭️  Skipped shell completion installation")
		fmt.Printf("To install later, run: juleson completion %s\n", shell)
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
		fmt.Printf("⚠️  Unsupported shell: %s\n", shell)
		fmt.Printf("To install manually, run: juleson completion %s\n", shell)
		return nil
	}
}

func setupJulesAPI(cfg *config.Config) error {
	fmt.Println("🔑 Jules API Configuration")
	fmt.Println("──────────────────────────")

	// Check if API key already exists
	if cfg.Jules.APIKey != "" {
		fmt.Println("✅ Jules API key is already configured")
		if !setupNonInteractive {
			fmt.Print("Update API key? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
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
			fmt.Println("✅ Using JULES_API_KEY from environment")
			return nil
		}
		return fmt.Errorf("Jules API key not found. Set JULES_API_KEY environment variable")
	}

	// Interactive prompt
	fmt.Println()
	fmt.Println("To use Jules, you need an API key.")
	fmt.Println("Get your API key from: https://jules.ai/settings/api")
	fmt.Println()
	fmt.Print("Enter your Jules API key: ")

	reader := bufio.NewReader(os.Stdin)
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	cfg.Jules.APIKey = apiKey
	cfg.Jules.BaseURL = "https://jules.googleapis.com/v1alpha"
	cfg.Jules.Timeout = 30000000000 // 30s in nanoseconds
	cfg.Jules.RetryAttempts = 3

	fmt.Println("✅ Jules API key configured")
	return nil
}

func setupGitHub(cfg *config.Config) error {
	fmt.Println("🔗 GitHub Integration Setup")
	fmt.Println("───────────────────────────")

	// Check if token already exists
	if cfg.GitHub.Token != "" {
		fmt.Println("✅ GitHub token is already configured")
		if !setupNonInteractive {
			fmt.Print("Update GitHub token? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
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
			fmt.Println("✅ Using GITHUB_TOKEN from environment")
			return nil
		}
		fmt.Println("ℹ️  GitHub token not found. Skipping GitHub integration.")
		fmt.Println("   You can configure it later with: juleson github login")
		return nil
	}

	// Interactive prompt
	fmt.Println()
	fmt.Println("GitHub integration is optional but recommended for:")
	fmt.Println("  • Managing pull requests from Jules sessions")
	fmt.Println("  • Connecting repositories to Jules")
	fmt.Println("  • Viewing repository information")
	fmt.Println()
	fmt.Print("Configure GitHub integration now? (Y/n): ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "n" || response == "no" {
		fmt.Println("⏭️  Skipped GitHub integration")
		fmt.Println("   To configure later, run: juleson github login")
		return nil
	}

	fmt.Println()
	fmt.Println("To integrate with GitHub, you need a Personal Access Token.")
	fmt.Println("Create one at: https://github.com/settings/tokens")
	fmt.Println()
	fmt.Println("Required scopes:")
	fmt.Println("  ✅ repo          - Full control of private repositories")
	fmt.Println("  ✅ workflow      - Update GitHub Action workflows")
	fmt.Println("  ✅ read:org      - Read org and team membership (optional)")
	fmt.Println()
	fmt.Print("Enter your GitHub Personal Access Token: ")

	token, _ := reader.ReadString('\n')
	token = strings.TrimSpace(token)

	if token == "" {
		fmt.Println("⏭️  Skipped GitHub integration (empty token)")
		return nil
	}

	cfg.GitHub.Token = token
	cfg.GitHub.DefaultOrg = ""
	cfg.GitHub.PR.DefaultMergeMethod = "squash"
	cfg.GitHub.PR.AutoDeleteBranch = true
	cfg.GitHub.Discovery.Enabled = true
	cfg.GitHub.Discovery.UseGitRemote = true
	cfg.GitHub.Discovery.CacheTTL = 300000000000 // 5m in nanoseconds

	fmt.Println("✅ GitHub integration configured")
	return nil
}

func saveConfiguration(cfg *config.Config) error {
	fmt.Println("💾 Saving Configuration")
	fmt.Println("───────────────────────")

	// Ensure config directory exists
	configDir := filepath.Join(".", "configs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save configuration
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("✅ Configuration saved to configs/juleson.yaml")
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
	if _, err := exec.LookPath("bash-completion"); err != nil {
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
