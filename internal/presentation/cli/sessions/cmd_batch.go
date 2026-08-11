package sessions

import (
	"github.com/spf13/cobra"
)

// BatchCmd returns the command for creating parallel sessions.
func (h *CommandHandler) BatchCmd() *cobra.Command {
	var (
		batchParallel       int
		batchTitle          string
		batchID             string
		batchGroupTitle     string
		batchStartingBranch string
		batchAutomationMode string
	)

	batchCmd := &cobra.Command{
		Use:   "batch [source-id] [task-file-or-prompt]",
		Short: "Create parallel sessions for one task",
		Long:  "Create 1-5 parallel Jules sessions for the same source and task. Batch sessions require plan approval by default.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return batchCreateSessions(h.cfg, args[0], args[1], BatchSessionOptions{
				Parallel:       batchParallel,
				Title:          batchTitle,
				BatchID:        batchID,
				GroupTitle:     batchGroupTitle,
				StartingBranch: batchStartingBranch,
				AutomationMode: batchAutomationMode,
			})
		},
	}

	batchCmd.Flags().IntVar(&batchParallel, "parallel", 2, "Number of parallel sessions to create (1-5)")
	batchCmd.Flags().StringVar(&batchTitle, "title", "", "Optional title prefix for created sessions")
	batchCmd.Flags().StringVar(&batchID, "batch-id", "", "Optional batch identifier to include in prompts and output")
	batchCmd.Flags().StringVar(&batchGroupTitle, "group-title", "", "Optional group title to include in prompts and output")
	batchCmd.Flags().StringVar(&batchStartingBranch, "starting-branch", "", "Starting branch for the source-backed sessions")
	batchCmd.Flags().StringVar(&batchAutomationMode, "automation-mode", "", "Automation mode such as AUTO_CREATE_PR")

	return batchCmd
}
