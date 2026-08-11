package sessions

import (
	"github.com/spf13/cobra"
)

// ArtifactsCmd returns the command for inspecting artifacts.
func (h *CommandHandler) ArtifactsCmd() *cobra.Command {
	artifactsCmd := &cobra.Command{
		Use:   "artifacts",
		Short: "Inspect session artifact manifests",
	}

	artifactsCmd.AddCommand(&cobra.Command{
		Use:   "list [session-id]",
		Short: "List session artifact manifests",
		Long:  "List documented artifacts with activity IDs, indexes, patch metadata, media MIME types, and bash exit codes.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listSessionArtifacts(h.cfg, args[0])
		},
	})

	return artifactsCmd
}
