package sessions

import (
	"github.com/SamyRai/juleson/internal/config"
	"github.com/spf13/cobra"
)

// NewSessionsCommand creates the sessions command.
func NewSessionsCommand(cfg *config.Config) *cobra.Command {
	sessionsCmd := &cobra.Command{
		Use:   "sessions",
		Short: "Manage Jules sessions",
		Long:  "List, monitor, and manage Jules AI coding sessions",
	}

	handler := NewCommandHandler(cfg)

	// Add subcommands via the handler
	sessionsCmd.AddCommand(handler.ListCmd())
	sessionsCmd.AddCommand(handler.CreateCmd())
	sessionsCmd.AddCommand(handler.WatchCmd())
	sessionsCmd.AddCommand(handler.ApproveCmd())
	sessionsCmd.AddCommand(handler.StatusCmd())
	sessionsCmd.AddCommand(handler.GetCmd())
	sessionsCmd.AddCommand(handler.PlansCmd())
	sessionsCmd.AddCommand(handler.ReviewCmd())
	sessionsCmd.AddCommand(handler.MessageCmd())
	sessionsCmd.AddCommand(handler.DeleteCmd())
	sessionsCmd.AddCommand(handler.ApplyCmd())
	sessionsCmd.AddCommand(handler.BatchCmd())
	sessionsCmd.AddCommand(handler.ArtifactsCmd())
	sessionsCmd.AddCommand(handler.OutputsCmd())
	sessionsCmd.AddCommand(handler.DownloadCmd())
	sessionsCmd.AddCommand(handler.DownloadActivityCmd())
	sessionsCmd.AddCommand(handler.PreviewCmd())
	sessionsCmd.AddCommand(handler.PreviewActivityCmd())
	sessionsCmd.AddCommand(handler.AutocleanCmd())

	return sessionsCmd
}
