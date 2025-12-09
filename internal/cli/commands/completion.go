package commands

import (
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for juleson.

The completion script needs to be loaded into your shell to enable command completion.

Bash:
  # Linux:
  juleson completion bash > /etc/bash_completion.d/juleson

  # macOS:
  juleson completion bash > $(brew --prefix)/etc/bash_completion.d/juleson

Zsh:
  # Add to ~/.zshrc:
  source <(juleson completion zsh)

  # Or generate completion file:
  juleson completion zsh > "${fpath[1]}/_juleson"

Fish:
  juleson completion fish > ~/.config/fish/completions/juleson.fish

PowerShell:
  juleson completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, add to your PowerShell profile:
  juleson completion powershell >> $PROFILE
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:                  runCompletion,
}

func runCompletion(cmd *cobra.Command, args []string) error {
	switch args[0] {
	case "bash":
		return cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		return cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		return cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
	}
	return nil
}

// NewCompletionCommand creates the completion command
func NewCompletionCommand() *cobra.Command {
	return completionCmd
}
