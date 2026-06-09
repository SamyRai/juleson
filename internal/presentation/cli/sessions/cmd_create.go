package sessions

import (
	"fmt"

	"github.com/spf13/cobra"
)

// CreateCmd returns the command for creating a session.
func (h *CommandHandler) CreateCmd() *cobra.Command {
	var createNoSource bool
	createOptions := CreateSessionOptions{}

	createCmd := &cobra.Command{
		Use:   "create [source-id] [prompt]",
		Short: "Create a new session",
		Long:  "Create a new Jules session with a repository source, or pass --no-source for a repoless session",
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := createOptions
			options.NoSource = createNoSource

			if createNoSource {
				if options.PromptFile != "" {
					if len(args) != 0 {
						return fmt.Errorf("--no-source with --prompt-file does not accept positional arguments")
					}
					return createSession(h.cfg, "", "", options)
				}
				if len(args) != 1 {
					return fmt.Errorf("--no-source accepts exactly one prompt argument, or use --prompt-file")
				}
				return createSession(h.cfg, "", args[0], options)
			}

			if options.PromptFile != "" {
				if len(args) != 1 {
					return fmt.Errorf("--prompt-file requires exactly one source ID argument")
				}
				return createSession(h.cfg, args[0], "", options)
			}

			if len(args) != 2 {
				return fmt.Errorf("provide source ID and prompt, use --prompt-file, or pass --no-source with a prompt")
			}

			return createSession(h.cfg, args[0], args[1], options)
		},
	}

	createCmd.Flags().BoolVar(&createNoSource, "no-source", false, "Create a repoless session without sourceContext")
	createCmd.Flags().StringVar(&createOptions.PromptFile, "prompt-file", "", "Read the session prompt from a file")
	createCmd.Flags().StringVar(&createOptions.Title, "title", "", "Optional session title")
	createCmd.Flags().StringVar(&createOptions.StartingBranch, "starting-branch", "", "Starting branch for source-backed sessions")
	createCmd.Flags().BoolVar(&createOptions.RequirePlanApproval, "require-plan-approval", false, "Require explicit plan approval before Jules starts work")
	createCmd.Flags().StringVar(&createOptions.AutomationMode, "automation-mode", "", "Automation mode such as AUTO_CREATE_PR")
	createCmd.Flags().BoolVar(&createOptions.WithIntel, "with-intel", false, "Analyze and attach codebase complexity and dependency graph to the prompt")

	return createCmd
}
