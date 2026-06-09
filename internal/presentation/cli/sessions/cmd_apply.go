package sessions

import (
	"github.com/spf13/cobra"
)

// ApplyCmd returns the command for applying or previewing patches.
func (h *CommandHandler) ApplyCmd() *cobra.Command {
	var (
		applyConfirm           bool
		applyAllowDirty        bool
		applyActivityID        string
		applyArtifactIndex     int
		applyAllowBaseMismatch bool
	)

	applyCmd := &cobra.Command{
		Use:   "apply [session-id] [project-path]",
		Short: "Preview or apply session patches",
		Long:  "Preview session patches by default. Pass --confirm to apply after a clean-worktree check.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return applySessionChanges(h.cfg, args[0], args[1], ApplySessionOptions{
				Confirm:           applyConfirm,
				AllowDirty:        applyAllowDirty,
				ActivityID:        applyActivityID,
				ArtifactIndex:     applyArtifactIndex,
				HasArtifactIndex:  cmd.Flags().Changed("artifact-index"),
				AllowBaseMismatch: applyAllowBaseMismatch,
			})
		},
	}

	applyCmd.Flags().BoolVar(&applyConfirm, "confirm", false, "Actually apply patches instead of dry-running")
	applyCmd.Flags().BoolVar(&applyAllowDirty, "allow-dirty", false, "Allow applying patches when the target worktree has local changes")
	applyCmd.Flags().StringVar(&applyActivityID, "activity-id", "", "Apply only changes from this activity ID or resource name")
	applyCmd.Flags().IntVar(&applyArtifactIndex, "artifact-index", 0, "Apply only this artifact index within the selected scope")
	applyCmd.Flags().BoolVar(&applyAllowBaseMismatch, "allow-base-mismatch", false, "Allow applying when a patch baseCommitId differs from target HEAD")

	return applyCmd
}
