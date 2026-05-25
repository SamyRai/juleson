package commands

import (
	"fmt"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/spf13/cobra"
)

// NewSessionsCommand creates the sessions command
func NewSessionsCommand(cfg *config.Config) *cobra.Command {
	sessionsCmd := &cobra.Command{
		Use:   "sessions",
		Short: "Manage Jules sessions",
		Long:  "List, monitor, and manage Jules AI coding sessions",
	}

	// List sessions
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all sessions",
		Long:  "List all Jules sessions with their current status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listSessions(cfg)
		},
	})

	// Create session
	var createNoSource bool
	createCmd := &cobra.Command{
		Use:   "create [source-id] [prompt]",
		Short: "Create a new session",
		Long:  "Create a new Jules session with a repository source, or pass --no-source for a repoless session",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if createNoSource {
				if len(args) != 1 {
					return fmt.Errorf("--no-source accepts exactly one prompt argument")
				}
				return createSession(cfg, "", args[0], true)
			}
			if len(args) != 2 {
				return fmt.Errorf("provide source ID and prompt, or pass --no-source with a prompt")
			}
			return createSession(cfg, args[0], args[1], false)
		},
	}
	createCmd.Flags().BoolVar(&createNoSource, "no-source", false, "Create a repoless session without sourceContext")
	sessionsCmd.AddCommand(createCmd)

	// Approve plan
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "approve [session-id]",
		Short: "Approve a plan in a session",
		Long:  "Approve a plan that is waiting for approval in the specified session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return approveSessionPlan(cfg, args[0])
		},
	})

	// Show session status
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show session status summary",
		Long:  "Show a summary of current session statuses",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showSessionStatus(cfg)
		},
	})

	// Get session details with activities
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "get [session-id]",
		Short: "Get session details and activities",
		Long:  "Get detailed information about a specific session including all activities",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return getSessionDetails(cfg, args[0])
		},
	})

	// Send message to session
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "message [session-id] [message]",
		Short: "Send a message to a session",
		Long:  "Send a message to Jules within a session to request changes or provide feedback",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return sendSessionMessage(cfg, args[0], args[1])
		},
	})

	var deleteForce bool
	deleteCmd := &cobra.Command{
		Use:   "delete [session-id]",
		Short: "Delete a session",
		Long:  "Delete a Jules session. Without --force, type the session ID to confirm.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteSession(cfg, args[0], deleteForce)
		},
	}
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Delete without interactive confirmation")
	sessionsCmd.AddCommand(deleteCmd)

	// Download all session artifacts
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "download [session-id] [output-dir]",
		Short: "Download all artifacts from a session",
		Long:  "Download all artifacts (patches, outputs, media) from all activities in a session",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputDir := "."
			if len(args) > 1 {
				outputDir = args[1]
			}
			return downloadSessionArtifacts(cfg, args[0], outputDir)
		},
	})

	// Download artifacts from specific activity
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "download-activity [session-id] [activity-id] [output-dir]",
		Short: "Download artifacts from a specific activity",
		Long:  "Download all artifacts from a specific activity within a session",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputDir := "."
			if len(args) > 2 {
				outputDir = args[2]
			}
			return downloadActivityArtifacts(cfg, args[0], args[1], outputDir)
		},
	})

	// Preview all session artifacts
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "preview [session-id]",
		Short: "Preview all artifacts from a session",
		Long:  "Display artifacts (diffs, outputs, media info) from all activities in a session without downloading",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return previewSessionArtifacts(cfg, args[0])
		},
	})

	// Preview artifacts from specific activity
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "preview-activity [session-id] [activity-id]",
		Short: "Preview artifacts from a specific activity",
		Long:  "Display artifacts from a specific activity within a session without downloading",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return previewActivityArtifacts(cfg, args[0], args[1])
		},
	})

	return sessionsCmd
}
