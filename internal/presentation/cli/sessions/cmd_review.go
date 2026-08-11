package sessions

import (
	"github.com/spf13/cobra"
)

// ReviewCmd returns the command for reviewing session state and patch readiness.
func (h *CommandHandler) ReviewCmd() *cobra.Command {
	var (
		reviewActivityID    string
		reviewArtifactIndex int
		reviewJSON          bool
	)

	reviewCmd := &cobra.Command{
		Use:   "review [session-id] [project-path]",
		Short: "Review session state and patch readiness",
		Long:  "Read-only operator review combining session state, latest plan, outputs, artifact manifests, patch dry-run preview, blockers, and next actions.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return reviewSession(h.cfg, args[0], args[1], ReviewSessionOptions{
				ActivityID:       reviewActivityID,
				ArtifactIndex:    reviewArtifactIndex,
				HasArtifactIndex: cmd.Flags().Changed("artifact-index"),
				JSON:             reviewJSON,
			})
		},
	}

	reviewCmd.Flags().StringVar(&reviewActivityID, "activity-id", "", "Review patches only from this activity ID or resource name")
	reviewCmd.Flags().IntVar(&reviewArtifactIndex, "artifact-index", 0, "Review only this artifact index within the selected scope")
	reviewCmd.Flags().BoolVar(&reviewJSON, "json", false, "Print machine-readable JSON")

	return reviewCmd
}
