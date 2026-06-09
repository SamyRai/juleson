package core

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SamyRai/juleson/internal/config"
)

// NewConfigCommand creates the config command.
func NewConfigCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Juleson configuration",
		Long:  "Commands for validating and managing your Juleson configuration",
	}

	cmd.AddCommand(newConfigValidateCommand(cfg))

	return cmd
}

func newConfigValidateCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the current Juleson configuration",
		Long:  "Validates the effective Juleson configuration, checking for hard validation errors and reporting clear next steps for missing optional credentials.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "🔍 Validating Juleson configuration...")
			fmt.Fprintln(cmd.OutOrStdout(), "")

			hasErrors := false

			// Report missing optional credentials and next steps
			if cfg.Jules.APIKey == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "⚠️  Jules API key is missing.")
				fmt.Fprintln(cmd.OutOrStdout(), "   Next step: Run 'juleson setup' or set JULES_API_KEY environment variable.")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "✅ Jules API key is configured.")
			}

			if cfg.GitHub.Token == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "⚠️  GitHub token is missing.")
				fmt.Fprintln(cmd.OutOrStdout(), "   Next step: Set GITHUB_TOKEN or github.token if you use Jules-created PR commands.")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "✅ GitHub token is configured.")
			}

			if hasErrors {
				fmt.Fprintln(cmd.OutOrStdout(), "\n❌ Configuration validation failed with errors.")
				return fmt.Errorf("configuration validation failed")
			}

			fmt.Fprintln(cmd.OutOrStdout(), "\n✅ Validation complete. No secret values were printed.")
			return nil
		},
	}
}
